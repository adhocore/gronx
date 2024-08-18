package tasker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/adhocore/gronx"
)

// Option is the config options for Tasker.
type Option struct {
	File    string
	Tz      string
	Shell   string
	Out     string
	Until   int64
	Verbose bool
}

// TaskFunc is the actual task handler.
type TaskFunc func(ctx context.Context) (int, error)

// Task wraps a cron expr and its' command.
type Task struct {
	Expr string
	Cmd  string
}

// Tasker is the task manager.
type Tasker struct {
	until     time.Time
	ctx       context.Context
	loc       *time.Location
	gron      *gronx.Gronx
	Log       *log.Logger
	exprs     map[string][]string
	tasks     map[string]TaskFunc
	mutex     map[string]uint32
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup
	verbose   bool
	running   bool
	timeout   bool
	abort     bool
}

type result struct {
	err  error
	ref  string
	code int
}

var exit = os.Exit

// New inits a task manager.
// It returns Tasker.
func New(opt Option) *Tasker {
	gron := gronx.New()
	tasks := make(map[string]TaskFunc)
	exprs := make(map[string][]string)

	if opt.Tz == "" {
		opt.Tz = "Local"
	}

	loc, err := time.LoadLocation(opt.Tz)
	if err != nil {
		log.Printf("invalid tz location: %s", opt.Tz)
		exit(1)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	if opt.Out != "" {
		if _, err := os.Stat(filepath.Dir(opt.Out)); err != nil {
			log.Printf("output dir does not exist: %s", filepath.Base(opt.Out))
			exit(1)
		}

		file, err := os.OpenFile(opt.Out, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			log.Printf("can't open output file: %s", opt.Out)
			exit(1)
		}

		logger = log.New(file, "", log.LstdFlags)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Tasker{
		Log:       logger,
		loc:       loc,
		gron:      gron,
		exprs:     exprs,
		tasks:     tasks,
		verbose:   opt.Verbose,
		ctx:       ctx,
		ctxCancel: cancel,
	}
}

// WithContext adds a parent context to the Tasker struct
// and begins the abort when Done is received
func (t *Tasker) WithContext(ctx context.Context) *Tasker {
	t.ctx, t.ctxCancel = context.WithCancel(ctx)
	return t
}

// Taskify creates TaskFunc out of plain command wrt given options.
func (t *Tasker) Taskify(cmd string, opt Option) TaskFunc {
	sh := Shell(opt.Shell)

	return func(ctx context.Context) (int, error) {
		buf := strings.Builder{}
		exc := exec.Command(sh[0], sh[1], cmd)
		exc.Stderr = &buf
		exc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if t.Log.Writer() != exc.Stderr {
			exc.Stdout = t.Log.Writer()
		}

		err := exc.Run()
		if err == nil {
			return 0, nil
		}

		for _, ln := range strings.Split(strings.TrimRight(buf.String(), "\r\n"), "\n") {
			log.Println(ln)
		}

		code := 1
		if exErr, ok := err.(*exec.ExitError); ok {
			code = exErr.ExitCode()
		}

		return code, err
	}
}

// Shell gives a pair of shell and arg.
// It returns array of string.
func Shell(shell ...string) []string {
	if os.PathSeparator == '\\' {
		shell = append(shell, "git-bash.exe -c", "bash.exe -c", "powershell.exe -Command")
	} else {
		shell = append(shell, "bash -c", "sh -c", "zsh -c")
	}

	for _, sh := range shell {
		arg := "-c"
		cmd := strings.Split(sh, " -")
		if len(cmd) > 1 {
			arg = "-" + cmd[1]
		}
		if exc, err := exec.LookPath(cmd[0]); err == nil {
			return []string{exc, arg}
		}
	}

	return []string{"/bin/sh", "-c"}
}

const taskIDFormat = "[%s][#%d]"

// Task appends new task handler for given cron expr.
// It returns Tasker (itself) for fluency and bails if expr is invalid.
func (t *Tasker) Task(expr string, task TaskFunc, concurrent ...bool) *Tasker {
	segs, err := gronx.Segments(expr)
	if err != nil {
		log.Fatalf("invalid cron expr: %+v", err)
	}

	concurrent = append(concurrent, true)
	old, expr := gronx.SpaceRe.ReplaceAllString(expr, " "), strings.Join(segs, " ")
	if _, ok := t.exprs[expr]; !ok {
		if !t.gron.IsValid(expr) {
			log.Fatalf("invalid cron expr: %+v", err)
		}

		t.exprs[expr] = []string{}
	}

	ref := fmt.Sprintf(taskIDFormat, old, len(t.exprs[expr])+1)

	t.exprs[expr] = append(t.exprs[expr], ref)
	t.tasks[ref] = task

	if !concurrent[0] {
		if len(t.mutex) == 0 {
			t.mutex = map[string]uint32{}
		}
		t.mutex[ref] = 0
	}

	return t
}

// Until sets the cutoff time until which the tasker runs.
// It returns itself for fluency.
func (t *Tasker) Until(until interface{}) *Tasker {
	switch until := until.(type) {
	case time.Duration:
		t.until = t.now().Add(until)
	case time.Time:
		t.until = until
	default:
		log.Printf("until must be time.Duration or time.Time, got: %v", reflect.TypeOf(until))
		exit(1)
	}

	return t
}

func (t *Tasker) now() time.Time {
	return time.Now().In(t.loc)
}

// Run runs the task manager.
func (t *Tasker) Run() {
	t.doSetup()
	t.running = true

	first := true
	for !t.abort && !t.timeout {
		ref, willTime := t.tickTimer(first)
		if t.timeout || t.abort {
			break
		}

		tasks := make(map[string]TaskFunc)
		t.gron.C.SetRef(ref)
		for expr, refs := range t.exprs {
			if due, _ := t.gron.SegmentsDue(strings.Split(expr, " ")); !due {
				continue
			}

			for _, ref := range refs {
				tasks[ref] = t.tasks[ref]
			}
		}

		if len(tasks) > 0 {
			t.runTasks(tasks)
		}

		first = false
		t.timeout = willTime
	}

	t.wait()
	t.running = false
}

// Running tells if tasker is up and running
func (t *Tasker) Running() bool {
	return t.running && !t.abort && !t.timeout
}

// Stop the task manager.
func (t *Tasker) Stop() {
	t.stop()
}

func (t *Tasker) stop() {
	t.ctxCancel()
	t.abort = true
}

var dateFormat = "2006/01/02 15:04:05"

func (t *Tasker) doSetup() {
	if len(t.tasks) == 0 {
		t.Log.Fatal("[tasker] no tasks available")
	}
	if !t.until.IsZero() && t.verbose {
		if t.until.Before(t.now()) {
			log.Fatalf("[tasker] timeout must be in future")
		}
		t.Log.Printf("[tasker] final tick on or before %s", t.until.Format(dateFormat))
	}

	// If we have seconds precision tickSec should be 1
	for expr := range t.exprs {
		if expr[0:2] != "0 " {
			tickSec = 1
			break
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sig:
		case <-t.ctx.Done():
			if t.verbose {
				t.Log.Printf("[tasker] received signal on context.Done, aborting")
			}
		}

		t.stop()
	}()
}

var tickSec = 60

func (t *Tasker) tickTimer(first bool) (time.Time, bool) {
	now, timed, willTime := t.now(), !t.until.IsZero(), false
	if t.timeout || t.abort {
		return now, willTime
	}

	wait := tickSec - now.Second()%tickSec
	if !first && wait == 0 {
		wait = tickSec
	}

	if wait < 1 || wait > tickSec {
		return now, willTime
	}

	next := now.Add(time.Duration(wait) * time.Second)
	willTime = timed && next.After(t.until)
	if t.verbose && !willTime {
		t.Log.Printf("[tasker] next tick on %s", next.Format(dateFormat))
	}

	if willTime {
		next = now.Add(time.Duration(tickSec) - now.Sub(t.until))
	}
	for !t.abort && !t.timeout && t.now().Before(next) {
		time.Sleep(100 * time.Millisecond)
	}

	t.timeout = timed && next.After(t.until)

	return next, willTime
}

func (t *Tasker) runTasks(tasks map[string]TaskFunc) {
	if t.verbose {
		if t.abort {
			t.Log.Println("[tasker] completing pending tasks")
		} else {
			t.Log.Printf("[tasker] running %d due tasks\n", len(tasks))
		}
	}

	ctx := context.Background()
	if t.ctx != nil {
		ctx = t.ctx
	}

	for ref, task := range tasks {
		if !t.canRun(ref) {
			continue
		}

		t.wg.Add(1)
		rc := make(chan result)

		go t.doRun(ctx, ref, task, rc)
		go t.doOut(rc)
	}
}

func (t *Tasker) canRun(ref string) bool {
	lock, ok := t.mutex[ref]
	if !ok {
		return true
	}
	if atomic.CompareAndSwapUint32(&lock, 0, 1) {
		t.mutex[ref] = 1
		return true
	}
	return false
}

func (t *Tasker) doRun(ctx context.Context, ref string, task TaskFunc, rc chan result) {
	defer t.wg.Done()
	if t.abort || t.timeout {
		return
	}

	if t.verbose {
		t.Log.Printf("[tasker] task %s running\n", ref)
	}

	code, err := task(ctx)
	if lock, ok := t.mutex[ref]; ok {
		atomic.StoreUint32(&lock, 0)
		t.mutex[ref] = 0
	}

	rc <- result{err, ref, code}
}

func (t *Tasker) doOut(rc chan result) {
	res := <-rc
	if res.err != nil {
		t.Log.Printf("[tasker] task %s errored %v", res.ref, res.err)
	}

	if t.verbose {
		if res.code == 0 {
			t.Log.Printf("[tasker] task %s ran successfully", res.ref)
		} else {
			t.Log.Printf("[tasker] task %s returned error code: %d", res.ref, res.code)
		}
	}
}

func (t *Tasker) wait() {
	if !t.abort {
		t.Log.Println("[tasker] timed out, waiting tasks to complete")
	} else {
		t.Log.Println("[tasker] interrupted, waiting tasks to complete")
	}

	t.wg.Wait()

	// Allow a leeway period
	time.Sleep(100 * time.Microsecond)
}

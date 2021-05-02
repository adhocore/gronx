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
	Log     *log.Logger
	loc     *time.Location
	gron    *gronx.Gronx
	wg      sync.WaitGroup
	until   time.Time
	exprs   map[string][]string
	tasks   map[string]TaskFunc
	abort   bool
	timeout bool
	verbose bool
}

type result struct {
	ref  string
	code int
	err  error
}

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
		log.Fatalf("invalid tz location: %s", opt.Tz)
	}

	logger := log.Default()
	if opt.Out != "" {
		if _, err := os.Stat(filepath.Dir(opt.Out)); err != nil {
			log.Fatalf("output dir does not exist: %s", filepath.Base(opt.Out))
		}

		file, err := os.OpenFile(opt.Out, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			log.Fatalf("can't open output file: %s", opt.Out)
		}

		logger = log.New(file, "", log.LstdFlags)
	}

	return &Tasker{Log: logger, loc: loc, gron: &gron, exprs: exprs, tasks: tasks, verbose: opt.Verbose}
}

// Taskify creates TaskFunc out of plain command wrt given options.
func Taskify(cmd string, opt Option) TaskFunc {
	sh := Shell(opt.Shell)

	return func(ctx context.Context) (int, error) {
		err := exec.Command(sh[0], sh[1], cmd).Run()
		if err == nil {
			return 0, nil
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
func (t *Tasker) Task(expr string, task TaskFunc) *Tasker {
	segs, err := gronx.Segments(expr)
	if err != nil {
		log.Fatalf("invalid cron expr: %+v", err)
	}

	old, expr := gronx.SpaceRe.ReplaceAllString(expr, " "), strings.Join(segs, " ")
	if _, ok := t.exprs[expr]; !ok {
		// Validate expr.
		if _, err := t.gron.SegmentsDue(segs); err != nil {
			log.Fatalf("invalid cron expr: %+v", err)
		}

		t.exprs[expr] = []string{}
	}

	ref := fmt.Sprintf(taskIDFormat, old, len(t.exprs[expr])+1)

	t.exprs[expr] = append(t.exprs[expr], ref)
	t.tasks[ref] = task

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
		log.Fatalf("until must be time.Duration or time.Time, got: %v", reflect.TypeOf(until))
	}

	return t
}

func (t *Tasker) now() time.Time {
	return time.Now().In(t.loc)
}

// Run runs the task manager.
func (t *Tasker) Run() {
	t.doSetup()

	first := true
	for !t.abort && !t.timeout {
		ref, willTime := t.tickTimer(first)

		t.gron.C.SetRef(ref)
		if t.timeout {
			break
		}

		tasks := make(map[string]TaskFunc)
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
}

func (t *Tasker) doSetup() {
	if len(t.tasks) == 0 {
		t.Log.Fatal("[tasker] no tasks available")
	}
	if !t.until.IsZero() && t.verbose {
		if t.until.Before(t.now()) {
			log.Fatalf("[tasker] timeout must be in future")
		}
		t.Log.Printf("[tasker] final tick on %s", t.until.Format("2006/01/02 15:04:00"))
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		t.abort = true
		t.wait()
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

	if wait < 1 && wait > tickSec {
		return now, willTime
	}

	dur := time.Duration(wait) * time.Second
	next := now.Add(dur)
	willTime = timed && next.After(t.until)
	if t.verbose && !willTime {
		t.Log.Printf("[tasker] next tick on %s", next.Format("2006/01/02 15:04:05"))
	}

	time.Sleep(dur)
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

	ctx := context.TODO()
	for ref, task := range tasks {
		t.wg.Add(1)
		rc := make(chan result)

		go t.doRun(ctx, ref, task, rc)
		go t.doOut(rc)
	}
}

func (t *Tasker) doRun(ctx context.Context, ref string, task TaskFunc, rc chan result) {
	defer t.wg.Done()
	if t.abort {
		return
	}

	if t.verbose {
		t.Log.Printf("[tasker] task %s running\n", ref)
	}
	code, err := task(ctx)

	rc <- result{ref, code, err}
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
}

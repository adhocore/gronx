package tasker

import (
	"context"
	"log"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTaskResultHandlerDrainsAfterStopOrCutoff(t *testing.T) {
	modes := []struct {
		name       string
		concurrent bool
	}{
		{name: "overlap", concurrent: true},
		{name: "single-instance", concurrent: false},
	}
	exits := []struct {
		name string
		set  func(*Tasker)
	}{
		{name: "stop", set: func(taskr *Tasker) { taskr.Stop() }},
		{name: "cutoff", set: func(taskr *Tasker) { taskr.timeout = true }},
	}

	for _, mode := range modes {
		mode := mode
		for _, exit := range exits {
			exit := exit
			t.Run(exit.name+"/"+mode.name, func(t *testing.T) {
				taskr := New(Option{})
				callbackCalled := make(chan struct{}, 1)
				taskr.Task("* * * * * *", func(context.Context) (int, error) {
					callbackCalled <- struct{}{}
					return 0, nil
				}, mode.concurrent)

				var ref string
				for ref = range taskr.tasks {
				}
				if ref == "" {
					t.Fatal("setup: task was not registered")
				}

				// Stop or cutoff is visible before the admitted worker starts.
				exit.set(taskr)
				taskr.runTasks(taskr.tasks)
				taskr.wg.Wait()

				if hasTaskResultHandler(40 * time.Millisecond) {
					t.Errorf("result handler remained blocked after task accounting drained")
				}
				select {
				case <-callbackCalled:
					t.Errorf("callback ran after Stop or cutoff was published")
				default:
				}
				if !mode.concurrent && !taskr.canRun(ref) {
					t.Errorf("non-overlap admission remained held after early exit")
				}
			})
		}
	}
}

func hasTaskResultHandler(timeout time.Duration) bool {
	stack := make([]byte, 1<<20)
	deadline := time.Now().Add(timeout)
	for {
		n := runtime.Stack(stack, true)
		if strings.Contains(string(stack[:n]), "(*Tasker).doOut(") {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
}

func TestTaskAdmissionReleasedBeforeResultLogging(t *testing.T) {
	taskr := New(Option{Verbose: true})
	taskr.Task("* * * * * *", func(context.Context) (int, error) { return 0, nil }, false)
	var ref string
	for ref = range taskr.tasks {
	}

	entered := make(chan struct{})
	release := make(chan struct{})
	taskr.Log = log.New(taskLogWriter(func(p []byte) (int, error) {
		if strings.Contains(string(p), "ran successfully") {
			close(entered)
			<-release
		}
		return len(p), nil
	}), "", 0)

	taskr.runTasks(taskr.tasks)
	<-entered
	admitted := taskr.canRun(ref)
	if !admitted {
		t.Error("finished callback's admission remained held during result logging")
	}
	close(release)
	taskr.wg.Wait()
	if admitted && taskr.canRun(ref) {
		t.Error("result reporting released a later admission")
	}
}

type taskLogWriter func([]byte) (int, error)

func (w taskLogWriter) Write(p []byte) (int, error) { return w(p) }

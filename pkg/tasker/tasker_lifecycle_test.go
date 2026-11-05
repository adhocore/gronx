package tasker

import (
	"context"
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

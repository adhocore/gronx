package tasker

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		tickSec = 2
		taskr := New(Option{Verbose: true, Out: "../../test/tasker.out"})

		called := 0
		taskr.Task("@always", func(ctx context.Context) (int, error) {
			taskr.Log.Println("task [@always][#1] sleeping 3s")
			time.Sleep(3 * time.Second)
			called++

			return 0, nil
		})

		time.Sleep(time.Second - time.Duration(time.Now().Nanosecond()))

		dur := 5 * time.Second
		now := time.Now()

		taskr.Until(dur).Run()

		if called != 2 {
			t.Errorf("task should run 2 times, ran %d times", called)
		}

		wait := tickSec - now.Second()%tickSec
		tickDur := time.Duration(wait) * time.Second
		start := now.Format(dateFormat)
		end := now.Add(dur).Format(dateFormat)
		next1 := now.Add(tickDur).Format(dateFormat)
		fin1 := now.Add(tickDur + 3*time.Second).Format(dateFormat)
		next2 := now.Add(tickDur + time.Duration(tickSec)*time.Second).Format(dateFormat)
		fin2 := now.Add(tickDur + time.Duration(tickSec+3)*time.Second).Format(dateFormat)

		buffers := []string{
			start + " [tasker] final tick on or before " + end,
			start + " [tasker] next tick on " + next1,

			next1 + " [tasker] running 1 due tasks",
			next1 + " [tasker] next tick on " + next2,
			next1 + " [tasker] task [@always][#1] running",
			next1 + " task [@always][#1] sleeping 3s",

			next2 + " [tasker] running 1 due tasks",
			next2 + " [tasker] task [@always][#1] running",
			next2 + " task [@always][#1] sleeping 3s",

			fin1 + " [tasker] task [@always][#1] ran successfully",
			end + " [tasker] timed out, waiting tasks to complete",
			fin2 + " [tasker] task [@always][#1] ran successfully",
		}

		buf, _ := ioutil.ReadFile("../../test/tasker.out")
		buffer := string(buf)
		fmt.Println(buffer)

		for _, expect := range buffers {
			if !strings.Contains(buffer, expect) {
				t.Errorf("buffer should contain %s", expect)
			}
		}
	})
}

func TestTaskify(t *testing.T) {
	t.Run("Taskify", func(t *testing.T) {
		ctx := context.TODO()
		taskr := New(Option{})
		code, err := taskr.Taskify("echo -n 'taskify' > ../../test/taskify.out; echo 'test' >> ../../test/taskify.out", Option{})(ctx)

		if code != 0 {
			t.Errorf("expected code 0, got %d", code)
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestWithContext(t *testing.T) {
	t.Run("WithContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		taskr := New(Option{Verbose: true, Out: "../../test/tasker-ctx.out"}).WithContext(ctx)

		called := 0
		taskr.Task("@always", func(ctx context.Context) (int, error) {
			taskr.Log.Println("task [@always][#1] waiting 3s")
			called++
			ct := 0
		M:
			for {
				time.Sleep(300 * time.Millisecond)
				select {
				case <-ctx.Done():
					taskr.Log.Printf("task [@always][#1] received Done signal after %d ms\n", ct*300)
					break M
				default:
					ct++
				}
			}
			return 0, nil
		})

		startCh := make(chan bool)

		go func() {
			<-startCh
			time.Sleep(4 * time.Second)
			cancel()
		}()

		startCh <- true
		taskr.Until(5 * time.Second).Run()

		if called != 2 {
			t.Errorf("task should run 2 times, ran %d times", called)
		}

		buf, _ := ioutil.ReadFile("../../test/tasker-ctx.out")
		buffer := string(buf)
		fmt.Println(buffer)
	})
}

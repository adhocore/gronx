package tasker

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	exit = func(code int) {}
	t.Run("New invalid Tz", func(t *testing.T) {
		New(Option{Tz: "Local/Xyz"})
	})
	t.Run("New invalid Out", func(t *testing.T) {
		New(Option{Out: "/a/b/c/d/e/f/out.log"})
	})
	t.Run("Invalid Until", func(t *testing.T) {
		var zero time.Time

		taskr := New(Option{})
		taskr.Until(time.Now().Add(time.Minute))

		taskr.Until(zero)
		taskr.Until(1)
		if !taskr.until.IsZero() {
			t.Error("tasker.until should be zero")
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		tickSec = 1
		taskr := New(Option{Verbose: true, Out: "../../test/tasker.out"})

		called := 0
		taskr.Task("* * * * * *", func(_ context.Context) (int, error) {
			taskr.Log.Println("task [* * * * * *][#1] sleeping 1s")
			time.Sleep(time.Second)
			called++

			return 0, nil
		})

		// dummy task that will never execute
		taskr.Task("* * * * * 2022", func(_ context.Context) (int, error) {
			return 0, nil
		})

		time.Sleep(time.Second - time.Duration(time.Now().Nanosecond()))

		dur := 2500 * time.Millisecond
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
		fin1 := now.Add(tickDur + 2*time.Second).Format(dateFormat)
		next2 := now.Add(tickDur + time.Duration(tickSec)*time.Second).Format(dateFormat)
		fin2 := now.Add(tickDur + time.Duration(tickSec)*time.Second).Format(dateFormat)

		buffers := []string{
			start + " [tasker] final tick on or before " + end,
			start + " [tasker] next tick on " + next1,

			next1 + " [tasker] running 1 due tasks",
			next1 + " [tasker] next tick on " + next2,
			next1 + " [tasker] task [* * * * * *][#1] running",
			next1 + " task [* * * * * *][#1] sleeping 1s",

			next2 + " [tasker] running 1 due tasks",
			next2 + " [tasker] task [* * * * * *][#1] running",
			next2 + " task [* * * * * *][#1] sleeping 1s",

			fin1 + " [tasker] task [* * * * * *][#1] ran successfully",
			end + " [tasker] timed out, waiting tasks to complete",
			fin2 + " [tasker] task [* * * * * *][#1] ran successfully",
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

		t.Run("Taskify err", func(t *testing.T) {
			ctx := context.TODO()
			taskr := New(Option{})
			code, err := taskr.Taskify("false", Option{})(ctx)
			if code != 1 {
				t.Errorf("expected code 127, got %d", code)
			}
			if err == nil {
				t.Error("expected error")
			}
		})
	})
}

func TestWithContext(t *testing.T) {
	// tickSec = 2
	t.Run("WithContext", func(t *testing.T) {
		os.Remove("../../test/tasker-ctx.out")
		ctx, cancel := context.WithCancel(context.Background())
		taskr := New(Option{Verbose: true, Out: "../../test/tasker-ctx.out"}).WithContext(ctx)

		called := 0
		taskr.Task("* * * * * *", func(ctx context.Context) (int, error) {
			called++
			ct := 0
		Over:
			for {
				time.Sleep(300 * time.Millisecond)
				select {
				case <-ctx.Done():
					break Over
				default:
					ct++
				}
			}
			return 0, nil
		})

		startCh := make(chan bool)

		go func() {
			<-startCh
			time.Sleep(2100 * time.Millisecond)
			cancel()
		}()

		startCh <- true
		taskr.Until(2200 * time.Millisecond).Run()

		if called != 2 {
			t.Errorf("task should run 2 times, ran %d times", called)
		}

		buf, _ := ioutil.ReadFile("../../test/tasker-ctx.out")
		fmt.Println(string(buf))
	})
}

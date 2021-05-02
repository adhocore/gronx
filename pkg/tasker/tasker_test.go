package tasker

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		tickSec = 2
		taskr := New(Option{Verbose: true, Out: "../../test/tasker.out"})

		called := 0
		taskr.Task("@always", func(ctx context.Context) (int, error) {
			taskr.Log.Println("task [@always] sleeping 3s")
			time.Sleep(3 * time.Second)
			taskr.Log.Println("task [@always] finished")
			called++

			return 0, nil
		})

		taskr.Until(5 * time.Second).Run()

		if called != 2 {
			t.Errorf("task should be run 2 times, ran %d times", called)
		}
	})
}

func TestTaskify(t *testing.T) {
	t.Run("Taskify", func(t *testing.T) {
		ctx := context.TODO()
		code, err := Taskify("echo -n 'taskify' > ../../test/taskify.out; echo 'test' >> ../../test/taskify.out", Option{})(ctx)

		if code != 0 {
			t.Errorf("expected code 0, got %d", code)
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

package tasker

import (
	"strings"
	"testing"
)

func TestMustParseTaskfile(t *testing.T) {
	exit = func(code int) {}
	t.Run("MustParseTaskfile", func(t *testing.T) {
		tasks := MustParseTaskfile(Option{File: "../../test/taskfile.txt"})
		if len(tasks) != 8 {
			t.Errorf("should have 8 tasks, got %d", len(tasks))
		}

		if tasks[0].Expr != "*/1 0/1 * * *" {
			t.Errorf("expected '*/1 0/1 * * *', got %s", tasks[0].Expr)
		}

		if tasks[2].Cmd != "echo '[task 3] @always' > test/task3.out" {
			t.Errorf("expected `echo '[task 3] @always' > test/task3.out`, got %s", tasks[2].Cmd)
		}

		t.Run("complex file - seconds precision", func(t *testing.T) {
			tasks := MustParseTaskfile(Option{File: "../../test/taskfile-complex.txt"})
			if len(tasks) != 13 {
				t.Errorf("should have 13 tasks, got %d", len(tasks))
			}
			for i, task := range tasks {
				if !strings.HasPrefix(task.Cmd, `echo "`) {
					t.Errorf("invalid cmd at %d [%s]: %s", i, task.Expr, task.Cmd)
				}
			}
		})

		t.Run("must parse - no file", func(t *testing.T) {
			tasks := MustParseTaskfile(Option{File: "../../test/taskfile.txtx"})
			if len(tasks) != 0 {
				t.Errorf("should have 0 tasks, got %d", len(tasks))
			}
		})
	})
}

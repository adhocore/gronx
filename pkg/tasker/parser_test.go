package tasker

import (
	"testing"
)

func TestMustParseTaskfile(t *testing.T) {
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
	})
}

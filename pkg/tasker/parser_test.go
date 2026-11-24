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

func TestLinesToTasksYearInCommand(t *testing.T) {
	tests := []struct {
		name string
		expr string
		cmd  string
	}{
		{"year suffix", "* * * * *", "backup2026 --daily"},
		{"year prefix", "* * * * *", "2026backup --daily"},
		{"year within name", "* * * * *", "backup2026tool --daily"},
		{"year before hyphen", "* * * * *", "backup2026-tool --daily"},
		{"plain command", "* * * * *", "backup --daily"},
		{"no arguments", "* * * * *", "backup2026"},
		{"seconds", "0 * * * * *", "backup --daily"},
		{"year", "* * * * * 2026", "backup --daily"},
		{"seconds and year", "0 * * * * * 2026", "backup --daily"},
		{"year range", "* * * * * 2026-2028", "backup --daily"},
		{"year list", "* * * * * 2026,2028", "backup --daily"},
		{"year step", "* * * * * 2026/2", "backup --daily"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := linesToTasks([]string{tt.expr + " " + tt.cmd})
			if len(tasks) != 1 {
				t.Fatalf("expected 1 task, got %d", len(tasks))
			}
			if got := strings.Join(strings.Fields(tasks[0].Expr), " "); got != tt.expr {
				t.Errorf("expected expression %q, got %q", tt.expr, got)
			}
			if tasks[0].Cmd != tt.cmd {
				t.Errorf("expected command %q, got %q", tt.cmd, tasks[0].Cmd)
			}
		})
	}
}

func TestLinesToTasksOutOfBoundsYear(t *testing.T) {
	if tasks := linesToTasks([]string{"* * * * * 10000 backup --daily"}); len(tasks) != 0 {
		t.Errorf("expected an out-of-bounds year to be rejected, got %v", tasks)
	}
}

package main

import (
	"os"
	"testing"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
)

func TestMustGetOption(t *testing.T) {
	old := os.Args
	exit = func (code int) {}
	t.Run("Main", func(t *testing.T) {
		expect := tasker.Option{File: "../../test/taskfile.txt", Out: "../../test/out.txt"}
		os.Args = append(old, "-verbose", "-file", expect.File, "-out", expect.Out)
		mustParseOption()
		if opt.File != expect.File {
			t.Errorf("file: expected %v, got %v", opt.File, expect.File)
		}
		if opt.Out != expect.Out {
			t.Errorf("out: expected %v, got %v", opt.Out, expect.Out)
		}

		t.Run("must parse option", func (t *testing.T) {
			os.Args = append(old, "-verbose", "-out", expect.Out)
			mustParseOption()
			if opt.File != "" {
				t.Error("opt.File must be empty "+opt.File)
			}

			os.Args = append(old, "-verbose", "-file", "invalid", "-out", expect.Out)
			mustParseOption()
			if opt.File != "invalid" {
				t.Error("opt.File must be invalid")
			}
		})

		t.Run("run", func (t *testing.T) {
			tick = time.Second
			os.Args = append(old, "-verbose", "-file", expect.File, "-out", expect.Out, "-until", "2")
			main()
		})

		os.Args = old
	})
}

package main

import (
	"os"
	"testing"

	"github.com/adhocore/gronx/pkg/tasker"
)

func TestMustGetOption(t *testing.T) {
	old := os.Args
	t.Run("Main", func(t *testing.T) {
		expect := tasker.Option{File: "../../test/taskfile.txt", Out: "../../test/out.txt"}
		os.Args = append(os.Args, "-verbose", "-file", expect.File, "-out", expect.Out)
		opt := mustGetOption()
		os.Args = old
		if opt.File != expect.File {
			t.Errorf("file: expected %v, got %v", opt.File, expect.File)
		}
		if opt.Out != expect.Out {
			t.Errorf("out: expected %v, got %v", opt.Out, expect.Out)
		}
	})
}

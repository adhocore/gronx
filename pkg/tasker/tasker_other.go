//go:build !windows

package tasker

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

// Taskify creates TaskFunc out of plain command wrt given options.
func (t *Tasker) Taskify(cmd string, opt Option) TaskFunc {
	sh := Shell(opt.Shell)

	return func(ctx context.Context) (int, error) {
		buf := strings.Builder{}
		exc := exec.Command(sh[0], sh[1], cmd)
		exc.Stderr = &buf
		exc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if t.Log.Writer() != exc.Stderr {
			exc.Stdout = t.Log.Writer()
		}

		err := exc.Run()
		if err == nil {
			return 0, nil
		}

		for _, ln := range strings.Split(strings.TrimRight(buf.String(), "\r\n"), "\n") {
			log.Println(ln)
		}

		code := 1
		if exErr, ok := err.(*exec.ExitError); ok {
			code = exErr.ExitCode()
		}

		return code, err
	}
}

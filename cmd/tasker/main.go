package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
)

func main() {
	opt := mustGetOption()
	taskr := tasker.New(opt)

	for _, task := range tasker.MustParseTaskfile(opt) {
		taskr.Task(task.Expr, taskr.Taskify(task.Cmd, opt))
	}

	if opt.Until > 0 {
		taskr.Until(time.Duration(opt.Until) * time.Minute)
	}

	taskr.Run()
}

func mustGetOption() tasker.Option {
	var opt tasker.Option

	flag.StringVar(&opt.File, "file", "", "The task file in crontab format")
	flag.StringVar(&opt.Tz, "tz", "Local", "The timezone to use for tasks")
	flag.StringVar(&opt.Shell, "shell", tasker.Shell()[0], "The shell to use for running tasks")
	flag.StringVar(&opt.Out, "out", "", "The fullpath to file where output from tasks are sent to")
	flag.BoolVar(&opt.Verbose, "verbose", false, "The verbose mode outputs as much as possible")
	flag.Int64Var(&opt.Until, "until", 0, "The timeout for task daemon in minutes")
	flag.Parse()

	if opt.File == "" {
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(opt.File); err != nil {
		log.Fatalf("can't read taskfile: %s", opt.File)
	}

	return opt
}

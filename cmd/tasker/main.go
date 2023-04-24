package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
)

var exit = os.Exit
var tick = time.Minute

var opt tasker.Option
var v bool

// Version of tasker, injected in build
var Version = "n/a"

func init() {
	flag.StringVar(&opt.File, "file", "", "The task file in crontab format (without user)")
	flag.StringVar(&opt.Tz, "tz", "Local", "The timezone to use for tasks")
	flag.StringVar(&opt.Shell, "shell", tasker.Shell()[0], "The shell to use for running tasks")
	flag.StringVar(&opt.Out, "out", "", "The fullpath to file where output from tasks are sent to")
	flag.BoolVar(&opt.Verbose, "verbose", false, "The verbose mode outputs as much as possible")
	flag.Int64Var(&opt.Until, "until", 0, "The timeout for task daemon in minutes")
	flag.BoolVar(&v, "v", false, "Show version")
}

func main() {
	mustParseOption()

	taskr := tasker.New(opt)
	for _, task := range tasker.MustParseTaskfile(opt) {
		taskr.Task(task.Expr, taskr.Taskify(task.Cmd, opt))
	}

	if opt.Until > 0 {
		taskr.Until(time.Duration(opt.Until) * tick)
	}

	taskr.Run()
}

func mustParseOption() {
	opt = tasker.Option{}
	flag.Parse()

	if v {
		fmt.Printf("v%s\n", Version)
		exit(0)
	}

	if opt.File == "" {
		flag.Usage()
		exit(1)
	}

	if _, err := os.Stat(opt.File); err != nil {
		log.Printf("can't read taskfile: %s", opt.File)
		exit(1)
	}
}

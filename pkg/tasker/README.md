# adhocore/gronx/pkg/tasker

[![Latest Version](https://img.shields.io/github/release/adhocore/gronx.svg?style=flat-square)](https://github.com/adhocore/gronx/releases)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/adhocore/gronx)](https://goreportcard.com/report/github.com/adhocore/gronx)
[![Test](https://github.com/adhocore/gronx/actions/workflows/test-action.yml/badge.svg)](https://github.com/adhocore/gronx/actions/workflows/test-action.yml)
[![Donate](https://img.shields.io/badge/donate-paypal-blue.svg?style=flat-square)](https://www.paypal.me/ji10/50usd)
[![Tweet](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/intent/tweet?text=Lightweight+fast+and+deps+free+cron+expression+parser+for+Golang&url=https://github.com/adhocore/gronx&hashtags=go,golang,parser,cron,cronexpr,cronparser)


`tasker` is cron expression based task scheduler and/or daemon for programamtic usage in Golang (tested on v1.13 and above) or independent standalone usage.

## Installation

```sh
go get -u github.com/adhocore/gronx/cmd/tasker
```
---
## Usage
### Go Tasker

Tasker is a task manager that can be programatically used in Golang applications.
It runs as a daemon and and invokes tasks scheduled with cron expression:

```go
package main

import (
	"context"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
)

func main() {
	taskr := tasker.New(tasker.Option{
		Verbose: true,
		// optional: defaults to local
		Tz:      "Asia/Bangkok",
		// optional: defaults to stderr log stream
		Out:     "/full/path/to/output-file",
	})

	// add task to run every minute
	taskr.Task("* * * * *", func(ctx context.Context) (int, error) {
		// do something ...

		// then return exit code and error, for eg: if everything okay
		return 0, nil
	}).Task("*/5 * * * *", func(ctx context.Context) (int, error) { // every 5 minutes
		// you can also log the output to Out file as configured in Option above:
		taskr.Log.Printf("done something in %d s", 2)

		return 0, nil
	})

	// run task without overlap, set concurrent flag to false:
	concurrent := false
	taskr.Task("* * * * * *", , tasker.Taskify("sleep 2", tasker.Option{}), concurrent)

	// every 10 minute with arbitrary command
	taskr.Task("@10minutes", taskr.Taskify("command --option val -- args", tasker.Option{Shell: "/bin/sh -c"}))

	// ... add more tasks

	// optionally if you want tasker to stop after 2 hour, pass the duration with Until():
	taskr.Until(2 * time.Hour)

	// finally run the tasker, it ticks sharply on every minute and runs all the tasks due on that time!
	// it exits gracefully when ctrl+c is received making sure pending tasks are completed.
	taskr.Run()
}
```

#### Concurrency

By default the tasks can run concurrently i.e if previous run is still not finished
but it is now due again, it will run again.
If you want to run only one instance of a task at a time, set concurrent flag to false:

```go
taskr := tasker.New(tasker.Option{})

concurrent := false
expr, task := "* * * * * *", tasker.Taskify("php -r 'sleep(2);'")
taskr.Task(expr, task, concurrent)
```

### Task Daemon
It can also be used as standalone task daemon instead of programmatic usage for Golang application.

First, just install tasker command:
```sh
go install github.com/adhocore/gronx/cmd/tasker@latest
```

Or you can also download latest prebuilt binary from [release](https://github.com/adhocore/gronx/releases/latest) for platform of your choice.

Then prepare a taskfile ([example](https://github.com/adhocore/gronx/blob/main/test/taskfile.txt)) in crontab format
(or can even point to existing crontab).
> `user` is not supported: it is just cron expr followed by the command.

Finally run the task daemon like so
```
tasker -file path/to/taskfile
```

#### Version

```sh
tasker -v
```

> You can pass more options to control the behavior of task daemon, see below.

####  Tasker command options:
```txt
-file string <required>
    The task file in crontab format
-out string
    The fullpath to file where output from tasks are sent to
-shell string
    The shell to use for running tasks (default "/usr/bin/bash")
-tz string
    The timezone to use for tasks (default "Local")
-until int
    The timeout for task daemon in minutes
-verbose
    The verbose mode outputs as much as possible
```

Examples:
```sh
tasker -verbose -file path/to/taskfile -until 120 # run until next 120min (i.e 2hour) with all feedbacks echoed back
tasker -verbose -file path/to/taskfile -out path/to/output # with all feedbacks echoed to the output file
tasker -tz America/New_York -file path/to/taskfile -shell zsh # run all tasks using zsh shell based on NY timezone
```

> File extension of taskfile for (`-file` option) does not matter: can be any or none.
> The directory for outfile (`-out` option) must exist, file is created by task daemon.

> Same timezone applies for all tasks currently and it might support overriding timezone per task in future release.

#### Notes on Windows
In Windows if it doesn't find `bash.exe` or `git-bash.exe` it will use `powershell`.
`powershell` may not be compatible with Unix flavored commands. Also to note:
you can't do chaining with `cmd1 && cmd2` but rather `cmd1 ; cmd2`.

---
## Understanding Cron Expression

Checkout [gronx](https://github.com/adhocore/gronx#cron-expression) docs on cron expression.

---
## License

> &copy; [MIT](https://github.com/adhocore/gronx/blob/main/LICENSE) | 2021-2099, Jitendra Adhikari

## Credits

This project is ported from [adhocore/cron-expr](https://github.com/adhocore/php-cron-expr) and
release managed by [please](https://github.com/adhocore/please).

---
### Other projects
My other golang projects you might find interesting and useful:

- [**urlsh**](https://github.com/adhocore/urlsh) - URL shortener and bookmarker service with UI, API, Cache, Hits Counter and forwarder using postgres and redis in backend, bulma in frontend; has [web](https://urlssh.xyz) and cli client
- [**fast**](https://github.com/adhocore/fast) - Check your internet speed with ease and comfort right from the terminal

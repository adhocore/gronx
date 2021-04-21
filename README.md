# adhocore/gronx

[![Latest Version](https://img.shields.io/github/release/adhocore/gronx.svg?style=flat-square)](https://github.com/adhocore/gronx/releases)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE)

`gronx` is Golang cron expression parser ported from [adhocore/cron-expr](https://github.com/adhocore/php-cron-expr).

- Zero dependency.
- Very **fast** because it bails early in case a segment doesn't match.

## Installation

```sh
go get -u github.com/adhocore/gronx
```

## Usage

```go
import (
	"time"

	"github.com/adhocore/gronx"
)

gron := gronx.New()
expr := "* * * * *"

// check if expr is due for current time
gron.IsDue(expr)

// check if expr is due for given time
gron.IsDue(expr, time.Date(2021, time.April, 1, 1, 1, 0, 0, time.UTC))
```

---
### Cron Expression

Cron expression normally consists of 5 segments viz:
```
<minute> <hour> <day> <month> <weekday>
```
and sometimes there can be 6th segment for `<year>` at the end.

### Real Abbreviations

You can use real abbreviations for month and week days. eg: `JAN`, `dec`, `fri`, `SUN`

### Tags

Following tags are available and they are converted to real cron expressions before parsing:

- *@yearly* or *@annually* - every year
- *@monthly* - every month
- *@daily* - every day
- *@weekly* - every week
- *@hourly* - every hour
- *@5minutes* - every 5 minutes
- *@10minutes* - every 10 minutes
- *@15minutes* - every 15 minutes
- *@30minutes* - every 30 minutes
- *@always* - every minute

```go
gron.IsDue("@5minutes")
```

### Modifiers

Following modifiers supported

- *Day of Month / 3rd segment:*
    - `L` stands for last day of month (eg: `L` could mean 29th for February in leap year)
    - `W` stands for closest week day (eg: `10W` is closest week days (MON-FRI) to 10th date)
- *Day of Week / 5th segment:*
    - `L` stands for last weekday of month (eg: `2L` is last monday)
    - `#` stands for nth day of week in the month (eg: `1#2` is second sunday)

---
## License

> &copy; [MIT](./LICENSE) | 2017-2019, Jitendra Adhikari

## Credits

This project is ported from [adhocore/cron-expr](https://github.com/adhocore/php-cron-expr) and
release managed by [please](https://github.com/adhocore/please).

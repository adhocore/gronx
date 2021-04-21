# adhocore/gronx

[![Latest Version](https://img.shields.io/github/release/adhocore/gronx.svg?style=flat-square)](https://github.com/adhocore/gronx/releases)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](LICENSE)
[![Test](https://github.com/adhocore/gronx/actions/workflows/test-action.yml/badge.svg)](https://github.com/adhocore/gronx/actions/workflows/test-action.yml)

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

// check if expr is even valid, returns bool
gron.IsValid(expr) // true

// check if expr is due for current time, returns bool and error
gron.IsDue(expr) // true|false, nil

// check if expr is due for given time
gron.IsDue(expr, time.Date(2021, time.April, 1, 1, 1, 0, 0, time.UTC)) // true|false, nil
```

---
### Cron Expression

Cron expression normally consists of 5 segments viz:
```
<minute> <hour> <day> <month> <weekday>
```
and sometimes there can be 6th segment for `<year>` at the end.

For each segments you can have multiple choices separated by comma:
> Eg: `0,30 * * * *` means either 0th or 30th minute.

To specify range of values you can use dash:
> Eg: `10-15 * * * *` means 10th, 11th, 12th, 13th, 14th and 15th minute.

To specify range of step you can combine a dash and slash:
> Eg: `10-15/2 * * * *` means every 2 minutes between 10 and 15 i.e 10th, 12th and 14th minute.

For the 3rd and 5th segment, there are additional [modifiers](#modifiers) (optional).

And if you want, you can mix them up:
> `5,12-20/4,55 * * * *` matches if any one of `5` or `12-20/4` or `55` matches the minute.

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

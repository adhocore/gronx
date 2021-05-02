package gronx

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func inStep(val int, s string) (bool, error) {
	parts := strings.Split(s, "/")
	step, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}
	if step == 0 {
		return false, errors.New("step can't be 0")
	}

	if strings.Index(s, "*/") == 0 || strings.Index(s, "0/") == 0 {
		return val%step == 0, nil
	}

	sub, end := strings.Split(parts[0], "-"), val
	start, err := strconv.Atoi(sub[0])
	if err != nil {
		return false, err
	}

	if len(sub) > 1 {
		end, err = strconv.Atoi(sub[1])
		if err != nil {
			return false, err
		}
	}

	return inStepRange(val, start, end, step), nil
}

func inRange(val int, s string) (bool, error) {
	parts := strings.Split(s, "-")
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, err
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}

	return start <= val && val <= end, nil
}

func inStepRange(val, start, end, step int) bool {
	for {
		if start == val {
			return true
		}
		if start > end {
			return false
		}

		start += step
	}
}

func isValidMonthDay(val string, last int, ref time.Time) (bool, error) {
	day, loc := ref.Day(), ref.Location()
	if val == "L" {
		return day == last, nil
	}

	pos := strings.Index(val, "W")
	if pos < 1 {
		return false, errors.New("invalid offset value: " + val)
	}

	nval, err := strconv.Atoi(val[0:pos])
	if err != nil {
		return false, err
	}

	for _, i := range []int{0, -1, 1, -2, 2} {
		incr := i + nval
		if incr > 0 && incr <= last {
			iref := time.Date(ref.Year(), ref.Month(), incr, ref.Hour(), ref.Minute(), ref.Second(), 0, loc)
			week := int(iref.Weekday())

			if week > 0 && week < 6 && iref.Month() == ref.Month() {
				return day == iref.Day(), nil
			}
		}
	}

	return false, nil
}

func isValidWeekDay(val string, last int, ref time.Time) (bool, error) {
	loc := ref.Location()
	if pos := strings.Index(strings.ReplaceAll(val, "7L", "0L"), "L"); pos > 0 {
		nval, err := strconv.Atoi(val[0:pos])
		if err != nil {
			return false, err
		}

		for i := 0; i < 7; i++ {
			decr := last - i
			dref := time.Date(ref.Year(), ref.Month(), decr, ref.Hour(), ref.Minute(), ref.Second(), ref.Nanosecond(), loc)

			if int(dref.Weekday()) == nval {
				return ref.Day() == decr, nil
			}
		}

		return false, nil
	}

	pos := strings.Index(val, "#")
	parts := strings.Split(strings.ReplaceAll(val, "7#", "0#"), "#")
	if pos < 1 || len(parts) < 2 {
		return false, errors.New("invalid offset value: " + val)
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, err
	}

	nth, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}

	if day < 0 || day > 7 || nth < 1 || nth > 5 || int(ref.Weekday()) != day {
		return false, nil
	}

	return ref.Day()/7 == nth-1, nil
}

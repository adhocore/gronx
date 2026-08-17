package gronx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func inStep(val int, s string, bounds []int, isWeekDay bool) (bool, error) {
	parts := strings.Split(s, "/")
	step, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}
	if step <= 0 {
		return false, errors.New("step can't be 0")
	}

	if strings.Index(s, "*/") == 0 {
		return (val-bounds[0])%step == 0, nil
	}
	if strings.Index(s, "0/") == 0 {
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

	if (len(sub) > 1 && end < start) || start < bounds[0] || end > bounds[1] {
		return false, fmt.Errorf("step '%s' out of bounds(%d, %d)", parts[0], bounds[0], bounds[1])
	}

	if isWeekDay {
		// 7 is an alias for Sunday (0), same as inRange. A leading 7 ("7/step")
		// starts the sequence at Sunday, so normalize it to 0.
		if start == 7 && end != 7 {
			start = 0
		}
		// time.Weekday() reports Sunday as 0, but a step whose range ends at 7
		// (e.g. "4-7/3") expresses it as 7. Sunday is due if either form lands
		// on the step, and checking both keeps "0-7/2" (0 already in range) working.
		if val == 0 {
			return inStepRange(0, start, end, step) || inStepRange(7, start, end, step), nil
		}
	}

	return inStepRange(val, start, end, step), nil
}

func inRange(val int, s string, bounds []int, isWeekDay bool) (bool, error) {
	parts := strings.Split(s, "-")
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, err
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}

	if start < bounds[0] || start > bounds[1] || end < bounds[0] || end > bounds[1] {
		return false, fmt.Errorf("range '%s' out of bounds(%d, %d)", s, bounds[0], bounds[1])
	}

	if isWeekDay {
		// 7 is an alias for Sunday (0). A leading 7 in a range ("7-y",
		// y != 7) means Sunday through y, so normalize start 7 -> 0.
		// A trailing 7 ("x-7") is preserved because "0-7" legitimately
		// represents the whole week, while "x-7" (x > 0) includes Sunday
		// after Saturday. Since val = ref.Weekday() is always 0..6, when
		// the range ends at 7 we normalize Sunday (val == 0) to 7 so that
		// the standard inclusive range check works for all cases.
		if start == 7 && end != 7 {
			start = 0
		}
		if end == 7 && val == 0 {
			val = 7
		}
	}

	if end < start {
		return false, fmt.Errorf("range '%s' out of bounds(%d, %d)", s, bounds[0], bounds[1])
	}

	return start <= val && val <= end, nil
}

func inStepRange(val, start, end, step int) bool {
	for i := start; i <= end && i <= val; i += step {
		if i == val {
			return true
		}
	}
	return false
}

func isValidMonthDay(val string, last int, ref time.Time) (valid bool, err error) {
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
				valid = day == iref.Day()
				break
			}
		}
	}

	return valid, nil
}

func isValidWeekDay(val string, last int, ref time.Time) (bool, error) {
	loc := ref.Location()

	if pos := strings.Index(val, "L"); pos > 0 {
		nval, err := strconv.Atoi(val[0:pos])
		if err != nil {
			return false, err
		}

		for i := 0; i < 7; i++ {
			day := last - i
			dref := time.Date(ref.Year(), ref.Month(), day, ref.Hour(), ref.Minute(), ref.Second(), 0, loc)
			if int(dref.Weekday()) == nval%7 {
				return ref.Day() == day, nil
			}
		}
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

	return (ref.Day()-1)/7 == nth-1, nil
}

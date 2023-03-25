package gronx

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Checker is interface for cron segment due check.
type Checker interface {
	GetRef() time.Time
	SetRef(ref time.Time)
	CheckDue(segment string, pos int) (bool, error)
}

// SegmentChecker is factory implementation of Checker.
type SegmentChecker struct {
	ref time.Time
}

// GetRef returns the current reference time
func (c *SegmentChecker) GetRef() time.Time {
	return c.ref
}

// SetRef sets the reference time for which to check if a cron expression is due.
func (c *SegmentChecker) SetRef(ref time.Time) {
	c.ref = ref
}

// CheckDue checks if the cron segment at given position is due.
// It returns bool or error if any.
func (c *SegmentChecker) CheckDue(segment string, pos int) (due bool, err error) {
	ref, last := c.GetRef(), -1
	val, loc := valueByPos(ref, pos), ref.Location()

	for _, offset := range strings.Split(segment, ",") {
		mod := (pos == 2 || pos == 4) && strings.ContainsAny(offset, "LW#")
		if due, err = c.isOffsetDue(offset, val, pos); due || (!mod && err != nil) {
			return
		}
		if !mod {
			continue
		}
		if last == -1 {
			last = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0).Add(-time.Nanosecond).Day()
		}
		if pos == 2 {
			due, err = isValidMonthDay(offset, last, ref)
		} else if pos == 4 {
			due, err = isValidWeekDay(offset, last, ref)
		}
		if due || err != nil {
			return due, err
		}
	}

	return false, nil
}

func (c *SegmentChecker) isOffsetDue(offset string, val, pos int) (bool, error) {
	if offset == "*" || offset == "?" {
		return true, nil
	}

	bounds := boundsByPos(pos)
	if strings.Contains(offset, "/") {
		return inStep(val, offset, bounds)
	}
	if strings.Contains(offset, "-") {
		if pos == 4 {
			offset = strings.Replace(offset, "7-", "0-", 1)
		}
		return inRange(val, offset, bounds)
	}

	if pos != 4 && (val == 0 || offset == "0") {
		return offset == "0" && val == 0, nil
	}

	nval, err := strconv.Atoi(offset)
	if err != nil {
		return false, err
	}
	if nval < bounds[0] || nval > bounds[1] {
		return false, fmt.Errorf("segment#%d: '%s' out of bounds(%d, %d)", pos, offset, bounds[0], bounds[1])
	}

	return nval == val || (pos == 4 && nval == 7 && val == 0), nil
}

func valueByPos(ref time.Time, pos int) int {
	switch pos {
	case 0:
		return ref.Minute()
	case 1:
		return ref.Hour()
	case 2:
		return ref.Day()
	case 3:
		return int(ref.Month())
	case 4:
		return int(ref.Weekday())
	case 5:
		return ref.Year()
	}

	return 0
}

func boundsByPos(pos int) []int {
	switch pos {
	case 0:
		return []int{0, 59}
	case 1:
		return []int{0, 23}
	case 2:
		return []int{1, 31}
	case 3:
		return []int{1, 12}
	case 4:
		return []int{0, 7}
	case 5:
		return []int{1, 9999}
	}

	return []int{0, 0}
}

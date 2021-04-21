package gronx

import (
	"strconv"
	"strings"
	"time"
)

type Checker interface {
	GetRef() time.Time
	SetRef(ref time.Time)
	CheckDue(segment string, pos int) (bool, error)
}

type SegmentChecker struct {
	ref time.Time
}

func (c *SegmentChecker) GetRef() time.Time {
	return c.ref
}

func (c *SegmentChecker) SetRef(ref time.Time) {
	c.ref = ref
}

func (c *SegmentChecker) CheckDue(segment string, pos int) (bool, error) {
	ref := c.GetRef()
	val, loc := valueByPos(ref, pos), ref.Location()

	for _, offset := range strings.Split(segment, ",") {
		mod := pos == 2 || pos == 4
		due, err := c.isOffsetDue(offset, val)

		if due || (!mod && err != nil) {
			return due, err
		}
		if mod && !strings.ContainsAny(offset, "LW#") {
			continue
		}

		last := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0).Add(-time.Nanosecond).Day()
		if pos == 2 {
			return isValidMonthDay(offset, last, ref)
		}
		if pos == 4 {
			return isValidWeekDay(offset, last, ref)
		}
	}

	return false, nil
}

func (c *SegmentChecker) isOffsetDue(offset string, val int) (bool, error) {
	if strings.Contains(offset, "/") && inStep(val, offset) {
		return true, nil
	}

	if strings.Contains(offset, "-") && inRange(val, offset) {
		return true, nil
	}

	if val == 0 || offset == "0" {
		return offset == "0" && val == 0, nil
	}

	nval, err := strconv.Atoi(offset)
	if err != nil {
		return false, err
	}

	return nval == val, nil
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

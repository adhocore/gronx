package gronx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronDateFormat is Y-m-d H:i (seconds are not significant)
const CronDateFormat = "2006-01-02 15:04"

// FullDateFormat is Y-m-d H:i:s (with seconds)
const FullDateFormat = "2006-01-02 15:04:05"

// NextTick gives next run time from now
func NextTick(expr string, inclRefTime bool) (time.Time, error) {
	return NextTickAfter(expr, time.Now(), inclRefTime)
}

// NextTickAfter gives next run time from the provided time.Time
func NextTickAfter(expr string, start time.Time, inclRefTime bool) (time.Time, error) {
	gron, next := New(), start.Truncate(time.Minute)
	due, err := gron.IsDue(expr, start)
	if err != nil || (due && inclRefTime) {
		return start, err
	}

	segments, _ := Segments(expr)
	if len(segments) > 5 && isPastYear(segments[5], next, inclRefTime) {
		return next, fmt.Errorf("unreachable year segment: %s", segments[5])
	}

	next, err = loop(gron, segments, next, inclRefTime)
	if due, _ = gron.IsDue(expr, next); due {
		err = nil
	}
	return next, err
}

func loop(gron Gronx, segments []string, start time.Time, incl bool) (next time.Time, err error) {
	iter, next, bumped := 1000, start, false
	for iter > 0 {
	over:
		iter--
		for pos, seg := range segments {
			if seg == "*" || seg == "?" {
				continue
			}
			if next, bumped, err = bumpUntilDue(gron.C, seg, pos, next); bumped {
				goto over
			}
		}
		if !incl && next.Format(CronDateFormat) == start.Format(CronDateFormat) {
			next, _, err = bumpUntilDue(gron.C, segments[0], 0, next.Add(time.Minute))
			continue
		}
		return next, err
	}
	return start, errors.New("tried so hard")
}

func isPastYear(year string, ref time.Time, incl bool) bool {
	min := ref.Year()
	if !incl {
		min += 1
	}
	for _, offset := range strings.Split(year, ",") {
		if strings.Contains(offset, "/") && strings.Index(offset, "*/") != 0 && strings.Index(offset, "0/") != 0 {
			return false
		}
		for _, part := range strings.Split(offset, "-") {
			val, err := strconv.Atoi(part)
			if err != nil || val > min {
				return false
			}
		}
	}
	return true
}

var limit = map[int]int{0: 60, 1: 24, 2: 31, 3: 12, 4: 366, 5: 100}

func bumpUntilDue(c Checker, segment string, pos int, ref time.Time) (time.Time, bool, error) {
	// <minute> <hour> <day> <month> <weekday>
	iter := limit[pos]
	for iter > 0 {
		c.SetRef(ref)
		if ok, _ := c.CheckDue(segment, pos); ok {
			return ref, iter != limit[pos], nil
		}
		ref = bump(ref, pos)
		iter--
	}
	return ref, false, errors.New("tried so hard")
}

func bump(ref time.Time, pos int) time.Time {
	inc := ref
	switch pos {
	case 0:
		inc = inc.Add(time.Minute)
	case 1:
		inc = inc.Add(time.Hour)
	case 2, 4:
		inc = inc.AddDate(0, 0, 1)
	case 3:
		inc = inc.AddDate(0, 1, 0)
	case 5:
		inc = inc.AddDate(1, 0, 0)
	}
	return inc
}

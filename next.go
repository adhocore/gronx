package gronx

import (
	"errors"
	"fmt"
	"regexp"
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
	gron, next := New(), start.Truncate(time.Second)
	due, err := gron.IsDue(expr, start)
	if err != nil || (due && inclRefTime) {
		return start, err
	}

	segments, _ := Segments(expr)
	if len(segments) > 6 && isUnreachableYear(segments[6], next, inclRefTime, false) {
		return next, fmt.Errorf("unreachable year segment: %s", segments[6])
	}

	next, err = loop(gron, segments, next, inclRefTime, false)
	// Ignore superfluous err
	if err != nil && gron.isDue(expr, next) {
		err = nil
	}
	return next, err
}

func loop(gron Gronx, segments []string, start time.Time, incl bool, reverse bool) (next time.Time, err error) {
	iter, next, bumped := 500, start, false
over:
	for iter > 0 {
		iter--
		for i := 0; i < len(segments); i++ {
			pos := len(segments) - 1 - i
			seg := segments[len(segments)-1-i]
			if seg == "*" || seg == "?" {
				continue
			}
			if next, bumped, err = bumpUntilDue(gron.C, seg, pos, next, reverse); bumped {
				goto over
			}
		}
		if !incl && next.Format(FullDateFormat) == start.Format(FullDateFormat) {
			delta := time.Second
			if reverse {
				delta = -time.Second
			}
			next = next.Add(delta)
			continue
		}
		return
	}
	return start, errors.New("tried so hard")
}

var dashRe = regexp.MustCompile(`/.*$`)

func isUnreachableYear(year string, ref time.Time, incl bool, reverse bool) bool {
	if year == "*" || year == "?" {
		return false
	}

	edge, inc := ref.Year(), 1
	if !incl {
		if reverse {
			inc = -1
		}
		edge += inc
	}
	for _, offset := range strings.Split(year, ",") {
		if strings.Index(offset, "*/") == 0 || strings.Index(offset, "0/") == 0 {
			return false
		}
		for _, part := range strings.Split(dashRe.ReplaceAllString(offset, ""), "-") {
			val, err := strconv.Atoi(part)
			if err != nil || (!reverse && val >= edge) || (reverse && val < edge) {
				return false
			}
		}
	}
	return true
}

var limit = map[int]int{0: 60, 1: 60, 2: 24, 3: 31, 4: 12, 5: 366, 6: 100}

func bumpUntilDue(c Checker, segment string, pos int, ref time.Time, reverse bool) (time.Time, bool, error) {
	// <second> <minute> <hour> <day> <month> <weekday> <year>
	iter := limit[pos]
	for iter > 0 {
		c.SetRef(ref)
		if ok, _ := c.CheckDue(segment, pos); ok {
			return ref, iter != limit[pos], nil
		}
		ref = bump(ref, pos, reverse)
		iter--
	}
	return ref, false, errors.New("tried so hard")
}

func bump(ref time.Time, pos int, reverse bool) time.Time {
	factor := 1
	if reverse {
		factor = -1
	}
	loc := ref.Location()

	switch pos {
	case 0:
		ref = ref.Add(time.Duration(factor) * time.Second)
	case 1:
		minTime := ref.Add(time.Duration(factor) * time.Minute)
		if reverse {
			ref = time.Date(minTime.Year(), minTime.Month(), minTime.Day(), minTime.Hour(), minTime.Minute(), 59, 0, loc)
		} else {
			ref = time.Date(minTime.Year(), minTime.Month(), minTime.Day(), minTime.Hour(), minTime.Minute(), 0, 0, loc)
		}
	case 2:
		hTime := ref.Add(time.Duration(factor) * time.Hour)
		if reverse {
			ref = time.Date(hTime.Year(), hTime.Month(), hTime.Day(), hTime.Hour(), 59, 59, 0, loc)
		} else {
			ref = time.Date(hTime.Year(), hTime.Month(), hTime.Day(), hTime.Hour(), 0, 0, 0, loc)
		}
	case 3, 5:
		dTime := ref.AddDate(0, 0, factor)
		if reverse {
			ref = time.Date(dTime.Year(), dTime.Month(), dTime.Day(), 23, 59, 59, 0, loc)
		} else {
			ref = time.Date(dTime.Year(), dTime.Month(), dTime.Day(), 0, 0, 0, 0, loc)
		}
	case 4:
		ref = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc)
		if reverse {
			ref = ref.Add(time.Duration(factor) * time.Second)
		} else {
			ref = ref.AddDate(0, factor, 0)
		}
	case 6:
		yTime := ref.AddDate(factor, 0, 0)
		if reverse {
			ref = time.Date(yTime.Year(), 12, 31, 23, 59, 59, 0, loc)
		} else {
			ref = time.Date(yTime.Year(), 1, 1, 0, 0, 0, 0, loc)
		}
	}
	return ref
}

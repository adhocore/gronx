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
	if len(segments) > 6 && isUnreachableYear(segments[6], next, false) {
		return next, fmt.Errorf("unreachable year segment: %s", segments[6])
	}

	next, err = loop(gron, segments, next, inclRefTime, false)
	// Ignore superfluous err
	if err != nil && gron.isDue(expr, next) {
		err = nil
	}
	return next, err
}

func loop(gron *Gronx, segments []string, start time.Time, incl bool, reverse bool) (next time.Time, err error) {
	iter, next, bumped := 500, start, false
over:
	for iter > 0 {
		iter--
		skipMonthDayForIter := false
		for i := 0; i < len(segments); i++ {
			pos := len(segments) - 1 - i
			seg := segments[pos]
			isMonthDay, isWeekday := pos == 3, pos == 5

			if seg == "*" || seg == "?" {
				continue
			}

			if !isWeekday {
				if isMonthDay && skipMonthDayForIter {
					continue
				}
				if next, bumped, err = bumpUntilDue(gron.C, seg, pos, next, reverse); bumped {
					goto over
				}
				continue
			}
			// From here we process the weekday segment in case it is neither * nor ?

			monthDaySeg := segments[3]
			intersect := strings.Index(seg, "*/") == 0 || strings.Index(monthDaySeg, "*") == 0 || monthDaySeg == "?"

			nextForWeekDay := next
			nextForWeekDay, bumped, err = bumpUntilDue(gron.C, seg, pos, nextForWeekDay, reverse)
			if !bumped {
				// Weekday seg is specific and next is already at right weekday, so no need to process month day if union case
				next = nextForWeekDay
				if !intersect {
					skipMonthDayForIter = true
				}
				continue
			}
			// Weekday was bumped, so we need to check for month day

			if intersect {
				// We need intersection so we keep bumped weekday and go over
				next = nextForWeekDay
				goto over
			}
			// Month day seg is specific and a number/list/range, so we need to check and keep the closest to next

			nextForMonthDay := next
			nextForMonthDay, bumped, err = bumpUntilDue(gron.C, monthDaySeg, 3, nextForMonthDay, reverse)

			monthDayIsClosestToNextThanWeekDay := reverse && nextForMonthDay.After(nextForWeekDay) ||
				!reverse && nextForMonthDay.Before(nextForWeekDay)

			if monthDayIsClosestToNextThanWeekDay {
				next = nextForMonthDay
				if !bumped {
					// Month day seg is specific and next is already at right month day, we can continue
					skipMonthDayForIter = true
					continue
				}
			} else {
				next = nextForWeekDay
			}
			goto over
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

func isUnreachableYear(year string, ref time.Time, reverse bool) bool {
	if year == "*" || year == "?" {
		return false
	}

	edge := ref.Year()
	for _, offset := range strings.Split(year, ",") {
		if strings.Index(offset, "*/") == 0 || strings.Index(offset, "0/") == 0 {
			return false
		}
		for _, part := range strings.Split(dashRe.ReplaceAllString(offset, ""), "-") {
			val, err := strconv.Atoi(part)
			if err != nil || (!reverse && val >= edge) || (reverse && val <= edge) {
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
		if reverse {
			ref = bumpReverse(ref, pos)
		} else {
			ref = bump(ref, pos)
		}
		iter--
	}
	return ref, false, errors.New("tried so hard")
}

func bump(ref time.Time, pos int) time.Time {
	loc := ref.Location()

	switch pos {
	case 0:
		ref = ref.Add(time.Second)
	case 1:
		minTime := ref.Add(time.Minute)
		ref = time.Date(minTime.Year(), minTime.Month(), minTime.Day(), minTime.Hour(), minTime.Minute(), 0, 0, loc)
	case 2:
		hTime := ref.Add(time.Hour)
		ref = time.Date(hTime.Year(), hTime.Month(), hTime.Day(), hTime.Hour(), 0, 0, 0, loc)
	case 3, 5:
		dTime := ref.AddDate(0, 0, 1)
		ref = time.Date(dTime.Year(), dTime.Month(), dTime.Day(), 0, 0, 0, 0, loc)
	case 4:
		ref = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc)
		ref = ref.AddDate(0, 1, 0)
	case 6:
		yTime := ref.AddDate(1, 0, 0)
		ref = time.Date(yTime.Year(), 1, 1, 0, 0, 0, 0, loc)
	}
	return ref
}

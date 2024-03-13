package gronx

import (
	"fmt"
	"time"
)

// PrevTick gives previous run time before now
func PrevTick(expr string, inclRefTime bool) (time.Time, error) {
	return PrevTickBefore(expr, time.Now(), inclRefTime)
}

// PrevTickBefore gives previous run time before given reference time
func PrevTickBefore(expr string, start time.Time, inclRefTime bool) (time.Time, error) {
	gron, prev := New(), start.Truncate(time.Second)
	due, err := gron.IsDue(expr, start)
	if err != nil || (due && inclRefTime) {
		return prev, err
	}

	segments, _ := Segments(expr)
	if len(segments) > 6 && isUnreachableYear(segments[6], prev, inclRefTime, true) {
		return prev, fmt.Errorf("unreachable year segment: %s", segments[6])
	}

	prev, err = loop(gron, segments, prev, inclRefTime, true)
	// Ignore superfluous err
	if err != nil && gron.isDue(expr, prev) {
		err = nil
	}
	return prev, err
}

func bumpReverse(ref time.Time, pos int) time.Time {
	loc := ref.Location()

	switch pos {
	case 0:
		ref = ref.Add(-time.Second)
	case 1:
		minTime := ref.Add(-time.Minute)
		ref = time.Date(minTime.Year(), minTime.Month(), minTime.Day(), minTime.Hour(), minTime.Minute(), 59, 0, loc)
	case 2:
		hTime := ref.Add(-time.Hour)
		ref = time.Date(hTime.Year(), hTime.Month(), hTime.Day(), hTime.Hour(), 59, 59, 0, loc)
	case 3, 5:
		dTime := ref.AddDate(0, 0, -1)
		ref = time.Date(dTime.Year(), dTime.Month(), dTime.Day(), 23, 59, 59, 0, loc)
	case 4:
		ref = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, loc)
		ref = ref.Add(-time.Second)
	case 6:
		yTime := ref.AddDate(-1, 0, 0)
		ref = time.Date(yTime.Year(), 12, 31, 23, 59, 59, 0, loc)
	}
	return ref
}

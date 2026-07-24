package gronx

import (
	"testing"
	"time"
)

// TestDOWRangeTrailingSeven regressions the bug where a DOW range ending in 7
// (e.g. "5-7") dropped Sunday while the semantically identical list form
// ("5,6,7") matched it. 7 is gronx's alias for Sunday (0), so "5-7" must equal
// "5,6,7" (Fri, Sat, Sun), "6-7" must equal "6,0" (Sat, Sun), "7-7" must equal
// "0-0" (Sunday only), and "0-7" must keep spanning the whole week.
func TestDOWRangeTrailingSeven(t *testing.T) {
	gron := New()

	// 2023-01-01 is Sunday, 2023-01-02 Mon, ... 2023-01-07 Sat.
	weekdayDates := func() map[time.Weekday]time.Time {
		m := map[time.Weekday]time.Time{}
		for d := 0; d < 7; d++ {
			ref := time.Date(2023, 1, 1+d, 0, 0, 0, 0, time.UTC)
			m[ref.Weekday()] = ref
		}
		return m
	}()

	dueAt := func(expr string, ref time.Time) bool {
		due, err := gron.IsDue(expr, ref)
		if err != nil {
			t.Fatalf("IsDue(%q, %s): %v", expr, ref.Format("2006-01-02"), err)
		}
		return due
	}

	t.Run("RED cases resolved", func(t *testing.T) {
		sun := weekdayDates[time.Sunday]
		mon := weekdayDates[time.Monday]

		// "5-7" must match Sunday (was dropped before fix).
		if !dueAt("0 0 * * 5-7", sun) {
			t.Errorf("IsDue('0 0 * * 5-7', Sun) = false, want true (Fri-Sun includes Sunday)")
		}
		// "6-7" must match Sunday (was dropped before fix).
		if !dueAt("0 0 * * 6-7", sun) {
			t.Errorf("IsDue('0 0 * * 6-7', Sun) = false, want true (Sat-Sun includes Sunday)")
		}
		// "7-7" must match Sunday only (matched every day before fix).
		if !dueAt("0 0 * * 7-7", sun) {
			t.Errorf("IsDue('0 0 * * 7-7', Sun) = false, want true (Sunday only)")
		}
		if dueAt("0 0 * * 7-7", mon) {
			t.Errorf("IsDue('0 0 * * 7-7', Mon) = true, want false (7-7 is Sunday only, not every day)")
		}
	})

	t.Run("controls preserved", func(t *testing.T) {
		fri := weekdayDates[time.Friday]
		sat := weekdayDates[time.Saturday]
		sun := weekdayDates[time.Sunday]
		thu := weekdayDates[time.Thursday]

		// Non-Sunday members of "5-7" still match.
		if !dueAt("0 0 * * 5-7", fri) {
			t.Errorf("IsDue('0 0 * * 5-7', Fri) = false, want true")
		}
		if !dueAt("0 0 * * 5-7", sat) {
			t.Errorf("IsDue('0 0 * * 5-7', Sat) = false, want true")
		}
		// "7-4" (Sun-Thu) is preserved: leading 7 normalizes to 0, no wrap.
		if !dueAt("0 0 * * 7-4", sun) {
			t.Errorf("IsDue('0 0 * * 7-4', Sun) = false, want true (Sun-Thu)")
		}
		if !dueAt("0 0 * * 7-4", thu) {
			t.Errorf("IsDue('0 0 * * 7-4', Thu) = false, want true (Sun-Thu)")
		}
		if dueAt("0 0 * * 7-4", fri) {
			t.Errorf("IsDue('0 0 * * 7-4', Fri) = true, want false (Fri not in Sun-Thu)")
		}
		// "0-7" keeps spanning the whole week (0..7, 7 ≡ 0 duplicate).
		for w := time.Sunday; w <= time.Saturday; w++ {
			if !dueAt("0 0 * * 0-7", weekdayDates[w]) {
				t.Errorf("IsDue('0 0 * * 0-7', %s) = false, want true (whole week)", w)
			}
		}
		// Single 7 still means Sunday.
		if !dueAt("0 0 * * 7", sun) {
			t.Errorf("IsDue('0 0 * * 7', Sun) = false, want true (7 = Sunday alias)")
		}
		// List form unchanged.
		if !dueAt("0 0 * * 5,6,7", sun) {
			t.Errorf("IsDue('0 0 * * 5,6,7', Sun) = false, want true")
		}
	})

	t.Run("range equals equivalent list across all weekdays", func(t *testing.T) {
		pairs := []struct {
			rng  string
			list string
		}{
			{"0 0 * * 5-7", "0 0 * * 5,6,7"},         // Fri, Sat, Sun
			{"0 0 * * 6-7", "0 0 * * 6,0"},           // Sat, Sun
			{"0 0 * * 7-7", "0 0 * * 0"},             // Sunday only
			{"0 0 * * 0-7", "0 0 * * 0,1,2,3,4,5,6"}, // whole week
		}
		for _, p := range pairs {
			for w := time.Sunday; w <= time.Saturday; w++ {
				ref := weekdayDates[w]
				r := dueAt(p.rng, ref)
				l := dueAt(p.list, ref)
				if r != l {
					t.Errorf("IsDue(%q, %s) = %v but IsDue(%q, %s) = %v; range and list must agree",
						p.rng, ref.Format("2006-01-02"), r, p.list, ref.Format("2006-01-02"), l)
				}
			}
		}
	})

	t.Run("NextTickAfter reaches Sunday for 5-7", func(t *testing.T) {
		// 2023-01-07 is a Saturday at 12:00. The next 00:00 whose DOW is in
		// {Fri, Sat, Sun} is 2023-01-08 (Sunday). Before the fix gronx
		// skipped Sunday and jumped to 2023-01-13 (Friday).
		start := time.Date(2023, 1, 7, 12, 0, 0, 0, time.UTC)
		next, err := NextTickAfter("0 0 * * 5-7", start, false)
		if err != nil {
			t.Fatalf("NextTickAfter err: %v", err)
		}
		want := time.Date(2023, 1, 8, 0, 0, 0, 0, time.UTC)
		if !next.Equal(want) {
			t.Errorf("NextTickAfter('0 0 * * 5-7', 2023-01-07 12:00) = %s, want %s (Sunday)",
				next.Format(FullDateFormat), want.Format(FullDateFormat))
		}
		// The range and list forms must produce the same next tick.
		nextList, err := NextTickAfter("0 0 * * 5,6,7", start, false)
		if err != nil {
			t.Fatalf("NextTickAfter list err: %v", err)
		}
		if !next.Equal(nextList) {
			t.Errorf("NextTickAfter range %s != list %s", next.Format(FullDateFormat), nextList.Format(FullDateFormat))
		}
	})
}

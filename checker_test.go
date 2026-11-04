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

// TestFullDayOfMonthRangeKeepsWeekdayRestriction regressions the bug where a
// day-of-month segment covering the whole field ("1-31", or the same days spelled
// out one by one) disabled the weekday segment. Such a segment matches every
// date, so it is equivalent to "*" and cannot make the union rule apply; only the
// weekday is left to restrict the match. Before the fix "0 0 1-31 * 1" ran on
// every day of the year instead of Mondays.
func TestFullDayOfMonthRangeKeepsWeekdayRestriction(t *testing.T) {
	gron := New()

	// 2011-06-13 is a Monday, 2011-06-15 a Wednesday.
	wed := time.Date(2011, 6, 15, 0, 0, 0, 0, time.UTC)
	mon := time.Date(2011, 6, 13, 0, 0, 0, 0, time.UTC)
	thu := time.Date(2011, 6, 16, 0, 0, 0, 0, time.UTC)

	t.Run("IsDue honours the weekday", func(t *testing.T) {
		for _, expr := range []string{"0 0 1-31 * 1", "0 0 1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31 * 1"} {
			if due, err := gron.IsDue(expr, thu); err != nil || due {
				t.Errorf("IsDue(%q, 2011-06-16 Thu) = %v, %v; want false, nil", expr, due, err)
			}
		}
		want := time.Date(2011, 6, 13, 0, 0, 0, 0, time.UTC) // Monday
		if due, err := gron.IsDue("0 0 1-31 * 1", want); err != nil || !due {
			t.Errorf("IsDue('0 0 1-31 * 1', 2011-06-13 Mon) = %v, %v; want true, nil", due, err)
		}
	})

	t.Run("full range equals star", func(t *testing.T) {
		for w := time.Sunday; w <= time.Saturday; w++ {
			ref := time.Date(2011, 6, 12+int(w), 0, 0, 0, 0, time.UTC) // Sun 12th .. Sat 18th
			rangeDue, err := gron.IsDue("0 0 1-31 * 1", ref)
			if err != nil {
				t.Fatalf("IsDue range: %v", err)
			}
			starDue, err := gron.IsDue("0 0 * * 1", ref)
			if err != nil {
				t.Fatalf("IsDue star: %v", err)
			}
			if rangeDue != starDue {
				t.Errorf("on %s IsDue('0 0 1-31 * 1') = %v but IsDue('0 0 * * 1') = %v; a full day-of-month range matches every date and must agree with '*'",
					ref.Format("2006-01-02"), rangeDue, starDue)
			}
		}
	})

	// A stepped day of month is still a restriction: "*/2" is the odd days, so the
	// weekday and the step intersect. Test case #73 of the next-run table pins this
	// for "3-59/15 6-12 */15 1 2-5" and it must not change with the range fix.
	t.Run("stepped day of month keeps intersecting with the weekday", func(t *testing.T) {
		if due, err := gron.IsDue("0 0 */2 * 1", mon); err != nil || !due {
			t.Errorf("IsDue('0 0 */2 * 1', 2011-06-13 Mon, odd day) = %v, %v; want true, nil", due, err)
		}
		evenMon := time.Date(2011, 6, 20, 0, 0, 0, 0, time.UTC) // Monday, even day
		if due, err := gron.IsDue("0 0 */2 * 1", evenMon); err != nil || due {
			t.Errorf("IsDue('0 0 */2 * 1', 2011-06-20 Mon, even day) = %v, %v; want false, nil", due, err)
		}
		next, err := NextTickAfter("0 0 */2 * 1", wed, false)
		if err != nil {
			t.Fatalf("NextTickAfter err: %v", err)
		}
		if want := time.Date(2011, 6, 27, 0, 0, 0, 0, time.UTC); !next.Equal(want) {
			t.Errorf("NextTickAfter('0 0 */2 * 1', 2011-06-15 Wed) = %s, want %s",
				next.Format(FullDateFormat), want.Format(FullDateFormat))
		}
	})

	t.Run("NextTickAfter picks the next Monday", func(t *testing.T) {
		next, err := NextTickAfter("0 0 1-31 * 1", wed, false)
		if err != nil {
			t.Fatalf("NextTickAfter err: %v", err)
		}
		if want := time.Date(2011, 6, 20, 0, 0, 0, 0, time.UTC); !next.Equal(want) {
			t.Errorf("NextTickAfter('0 0 1-31 * 1', 2011-06-15 Wed) = %s, want %s",
				next.Format(FullDateFormat), want.Format(FullDateFormat))
		}

		// Consecutive ticks must be a week apart, not a day.
		second, err := NextTickAfter("0 0 1-31 * 1", next, false)
		if err != nil {
			t.Fatalf("NextTickAfter err: %v", err)
		}
		if got := int(second.Sub(next).Hours() / 24); got != 7 {
			t.Errorf("consecutive ticks are %d days apart, want 7", got)
		}

		sunday, err := NextTickAfter("0 0 1-31 * 0", wed, false)
		if err != nil {
			t.Fatalf("NextTickAfter err: %v", err)
		}
		if want := time.Date(2011, 6, 19, 0, 0, 0, 0, time.UTC); !sunday.Equal(want) {
			t.Errorf("NextTickAfter('0 0 1-31 * 0', 2011-06-15 Wed) = %s, want %s",
				sunday.Format(FullDateFormat), want.Format(FullDateFormat))
		}
	})

	t.Run("PrevTickBefore picks the previous Monday", func(t *testing.T) {
		prev, err := PrevTickBefore("0 0 1-31 * 1", wed, false)
		if err != nil {
			t.Fatalf("PrevTickBefore err: %v", err)
		}
		if !prev.Equal(mon) {
			t.Errorf("PrevTickBefore('0 0 1-31 * 1', 2011-06-15 Wed) = %s, want %s",
				prev.Format(FullDateFormat), mon.Format(FullDateFormat))
		}
	})

	t.Run("BatchDue agrees with IsDue", func(t *testing.T) {
		for _, ref := range []time.Time{wed, thu, mon} {
			want, err := gron.IsDue("0 0 1-31 * 1", ref)
			if err != nil {
				t.Fatalf("IsDue: %v", err)
			}
			batch := gron.BatchDue([]string{"0 0 1-31 * 1"}, ref)[0]
			if batch.Err != nil {
				t.Fatalf("BatchDue: %v", batch.Err)
			}
			if batch.Due != want {
				t.Errorf("BatchDue = %v but IsDue = %v at %s", batch.Due, want, ref.Format("2006-01-02"))
			}
		}
	})

	t.Run("a range short of the whole field still restricts", func(t *testing.T) {
		// 1-29 leaves the 30th and 31st out, so it still constrains the match
		// and must not be treated as unrestricted. 2011-07-31 is a Sunday.
		last := time.Date(2011, 7, 31, 0, 0, 0, 0, time.UTC)
		if due, err := gron.IsDue("0 0 1-29 * 1", last); err != nil || due {
			t.Errorf("IsDue('0 0 1-29 * 1', 2011-07-31) = %v, %v; want false, nil", due, err)
		}
		if due, err := gron.IsDue("0 0 1-31 * 0", last); err != nil || !due {
			t.Errorf("IsDue('0 0 1-31 * 0', 2011-07-31) = %v, %v; want true, nil", due, err)
		}
		if due, err := gron.IsDue("0 0 */1 * 1", time.Date(2011, 7, 4, 0, 0, 0, 0, time.UTC)); err != nil || !due {
			t.Errorf("IsDue('0 0 */1 * 1', 2011-07-04 Mon) = %v, %v; want true, nil", due, err)
		}
	})
}

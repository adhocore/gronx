package gronx

import (
	"testing"
	"time"
)

func TestLeapDayTicks(t *testing.T) {
	for _, tc := range []struct {
		name, start, want  string
		reverse, inclusive bool
	}{
		{"next leap year", "2025-03-01 00:00:00", "2028-02-29 00:00:00", false, false},
		{"skip non-leap century", "2096-03-01 00:00:00", "2104-02-29 00:00:00", false, false},
		{"previous across century", "2104-02-28 00:00:00", "2096-02-29 00:00:00", true, false},
		{"include exact tick", "2028-02-29 00:00:00", "2028-02-29 00:00:00", false, true},
		{"exclude exact tick", "2028-02-29 00:00:00", "2032-02-29 00:00:00", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, err := time.Parse(FullDateFormat, tc.start)
			if err != nil {
				t.Fatal(err)
			}
			tick := NextTickAfter
			if tc.reverse {
				tick = PrevTickBefore
			}
			got, err := tick("0 0 0 29 2 * *", start, tc.inclusive)
			if err != nil {
				t.Fatal(err)
			}
			if got.Format(FullDateFormat) != tc.want {
				t.Fatalf("got %s, want %s", got.Format(FullDateFormat), tc.want)
			}
		})
	}
}

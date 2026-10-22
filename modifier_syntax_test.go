package gronx

import (
	"testing"
	"time"
)

func TestRejectsMalformedCronModifiers(t *testing.T) {
	cases := []struct {
		expr string
		ref  time.Time
	}{
		{"0 0 * * 1#2#3", time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)},
		{"0 0 * * 1LL", time.Date(2026, time.October, 26, 0, 0, 0, 0, time.UTC)},
		{"0 0 31WW * *", time.Date(2026, time.October, 30, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			if _, err := Segments(tc.expr); err == nil {
				t.Errorf("Segments(%q) returned no syntax error", tc.expr)
			}
			if IsValid(tc.expr) {
				t.Errorf("IsValid(%q) = true, want false", tc.expr)
			}
			if due, err := New().IsDue(tc.expr, tc.ref); err == nil {
				t.Errorf("IsDue(%q) = (%v, nil), want a syntax error", tc.expr, due)
			}
			if got := New().BatchDue([]string{tc.expr}, tc.ref)[0]; got.Err == nil {
				t.Errorf("BatchDue(%q) = (%v, nil), want a syntax error", tc.expr, got.Due)
			}
		})
	}
}

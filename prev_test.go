package gronx

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPrevTick(t *testing.T) {
	exp := "* * * * * *"
	t.Run("prev tick "+exp, func(t *testing.T) {
		ref, _ := time.Parse(FullDateFormat, "2020-02-02 02:02:02")
		prev, _ := PrevTickBefore(exp, ref, true)
		if prev.Format(FullDateFormat) != "2020-02-02 02:02:02" {
			t.Errorf("[incl] expected %v, got %v", ref, prev)
		}

		expect := time.Now().Add(-time.Second).Format(FullDateFormat)
		prev, _ = PrevTick(exp, false)
		if expect != prev.Format(FullDateFormat) {
			t.Errorf("expected %v, got %v", expect, prev)
		}
	})

	t.Run("prev tick excl "+exp, func(t *testing.T) {
		ref, _ := time.Parse(FullDateFormat, "2020-02-02 02:02:02")
		prev, _ := PrevTickBefore(exp, ref, false)
		if prev.Format(FullDateFormat) != "2020-02-02 02:02:01" {
			t.Errorf("[excl] expected %v, got %v", ref, prev)
		}
	})
}

func TestPrevTickBefore(t *testing.T) {
	t.Run("prev tick before", func(t *testing.T) {
		t.Run("seconds precision", func(t *testing.T) {
			ref, _ := time.Parse(FullDateFormat, "2020-02-02 02:02:02")
			next, _ := NextTickAfter("*/5 * * * * *", ref, false)
			prev, _ := PrevTickBefore("*/5 * * * * *", next, false)
			if prev.Format(FullDateFormat) != "2020-02-02 02:02:00" {
				t.Errorf("next > prev should be %s, got %s", "2020-02-02 02:02:00", prev)
			}
		})

		for i, test := range testcases() {
			t.Run(fmt.Sprintf("prev tick #%d: %s", i, test.Expr), func(t *testing.T) {
				ref, _ := time.Parse(FullDateFormat, test.Ref)
				next1, err := NextTickAfter(test.Expr, ref, false)
				if err != nil {
					return
				}

				prev1, err := PrevTickBefore(test.Expr, next1, true)
				if err != nil {
					if strings.HasPrefix(err.Error(), "unreachable year") {
						return
					}
					t.Errorf("%v", err)
				}

				if next1.Format(FullDateFormat) != prev1.Format(FullDateFormat) {
					t.Errorf("next->prev expect %s, got %s", next1, prev1)
				}

				next2, _ := NextTickAfter(test.Expr, next1, false)
				prev2, err := PrevTickBefore(test.Expr, next2, false)
				if err != nil {
					if strings.HasPrefix(err.Error(), "unreachable year") {
						return
					}
					t.Errorf("%s", err)
				}

				if next1.Format(FullDateFormat) != prev2.Format(FullDateFormat) {
					t.Errorf("next->next->prev expect %s, got %s", next1, prev2)
				}
			})
		}
	})
}

func TestPrevTickBeforeDST(t *testing.T) {
	tests := []struct {
		name string
		zone string
		expr string
		ref  string
		want string
	}{
		{
			name: "hour search keeps the second New York occurrence",
			zone: "America/New_York",
			expr: "20 1 * * *",
			ref:  "2024-11-03T02:10:00-05:00",
			want: "2024-11-03T01:20:00-05:00",
		},
		{
			name: "minute search keeps the second New York occurrence",
			zone: "America/New_York",
			expr: "20 * * * *",
			ref:  "2024-11-03T01:40:00-05:00",
			want: "2024-11-03T01:20:00-05:00",
		},
		{
			name: "hour search crosses the repeated Paris hour",
			zone: "Europe/Paris",
			expr: "0 0 * * *",
			ref:  "2024-10-27T03:10:00+01:00",
			want: "2024-10-27T00:00:00+02:00",
		},
		{
			name: "minute search keeps the first Paris occurrence",
			zone: "Europe/Paris",
			expr: "20 * * * *",
			ref:  "2024-10-27T02:40:00+02:00",
			want: "2024-10-27T02:20:00+02:00",
		},
		{
			name: "minute search crosses the New York rollback",
			zone: "America/New_York",
			expr: "50 * * * *",
			ref:  "2024-11-03T01:10:00-05:00",
			want: "2024-11-03T01:50:00-04:00",
		},
		{
			name: "hour search keeps the second Lord Howe occurrence",
			zone: "Australia/Lord_Howe",
			expr: "45 1 * * *",
			ref:  "2024-04-07T02:10:00+10:30",
			want: "2024-04-07T01:45:00+10:30",
		},
		{
			name: "minute search crosses the New York spring gap",
			zone: "America/New_York",
			expr: "30 * * * *",
			ref:  "2024-03-10T03:10:00-04:00",
			want: "2024-03-10T01:30:00-05:00",
		},
		{
			name: "hour search respects a fractional UTC offset",
			zone: "Asia/Kolkata",
			expr: "20 1 * * *",
			ref:  "2024-11-03T02:10:00+05:30",
			want: "2024-11-03T01:20:00+05:30",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			ref, err := time.Parse(time.RFC3339, tc.ref)
			if err != nil {
				t.Fatal(err)
			}
			want, err := time.Parse(time.RFC3339, tc.want)
			if err != nil {
				t.Fatal(err)
			}
			got, err := PrevTickBefore(tc.expr, ref.In(loc), false)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(want) {
				t.Errorf("expected %v, got %v", want, got)
			}
			if got.Location() != loc {
				t.Errorf("expected location %v, got %v", loc, got.Location())
			}
		})
	}
}

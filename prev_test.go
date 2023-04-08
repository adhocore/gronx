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

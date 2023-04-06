package gronx

import (
	"fmt"
	"testing"
	"time"
)

func TestNextTick(t *testing.T) {
	exp := "* * * * * *"
	t.Run("next tick incl "+exp, func(t *testing.T) {
		now := time.Now().Format(CronDateFormat)
		next, _ := NextTick(exp, true)
		tick := next.Format(CronDateFormat)
		if now != tick {
			t.Errorf("expected %v, got %v", now, tick)
		}
	})
	t.Run("next tick excl "+exp, func(t *testing.T) {
		expect := time.Now().Add(time.Minute).Format(CronDateFormat)
		next, _ := NextTick(exp, false)
		tick := next.Format(CronDateFormat)
		if expect != tick {
			t.Errorf("expected %v, got %v", expect, tick)
		}
	})
}

func TestNextTickAfter(t *testing.T) {
	t.Run("next run after", func(t *testing.T) {
		t.Run("seconds precision", func(t *testing.T) {
			ref, _ := time.Parse(FullDateFormat, "2020-02-02 02:02:02")
			next, _ := NextTickAfter("*/5 * * * * *", ref, false)
			if next.Format(FullDateFormat) != "2020-02-02 02:02:05" {
				t.Errorf("2020-02-02 02:02:02 next tick should be 2020-02-02 02:02:05")
			}
		})

		for i, test := range testcases() {
			t.Run(fmt.Sprintf("next run after incl #%d: %s", i, test.Expr), func(t *testing.T) {
				ref, _ := time.Parse(FullDateFormat, test.Ref)
				if next, err := NextTickAfter(test.Expr, ref, true); err == nil {
					actual := next.Format(FullDateFormat)
					if test.Expect != (test.Ref == actual) {
						t.Errorf("[incl] expected %v, got %v", test.Ref, actual)
					}
				}
			})
		}

		gron := New()
		for i, test := range testcases() {
			t.Run(fmt.Sprintf("next run after excl #%d: %s", i, test.Expr), func(t *testing.T) {
				ref, _ := time.Parse(FullDateFormat, test.Ref)
				next, err := NextTickAfter(test.Expr, ref, false)
				if err == nil {
					expect := test.Next
					if expect == "" {
						expect = test.Ref
					}
					actual := next.Format(CronDateFormat)
					if due, _ := gron.IsDue(test.Expr, next); !due {
						t.Errorf("[%s][%s] should be due on %v", test.Expr, test.Ref, next.Format(CronDateFormat))
					}
					if expect != actual {
						t.Errorf("[%s][%s] expected %v, got %v", test.Expr, test.Ref, expect, actual)
					}
				} else {
					fmt.Println(test.Expr+" failed", err)
				}
			})
		}
	})
}

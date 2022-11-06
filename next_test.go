package gronx

import (
	"fmt"
	"testing"
	"time"
)

func TestNextTick(t *testing.T) {
	exp := "* * * * *"
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
		for _, test := range testcases() {
			t.Run("next run after incl "+test.Expr, func(t *testing.T) {
				ref, _ := time.Parse("2006-01-02 15:04:05", test.Ref)
				if next, err := NextTickAfter(test.Expr, ref, true); err == nil {
					actual := next.Format("2006-01-02 15:04:05")
					if test.Expect != (test.Ref == actual) {
						t.Errorf("[incl] expected %v, got %v", test.Ref, actual)
					}
				}
			})
		}

		gron := New()
		for _, test := range testcases() {
			t.Run("next run after excl "+test.Expr, func(t *testing.T) {
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

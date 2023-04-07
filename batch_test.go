package gronx

import (
	"fmt"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	gron := New()

	t.Run("batch no error", func(t *testing.T) {
		ref := time.Now()
		exprs := []string{"@everysecond", "* * * * * *", "*  *  *  *  *  *"}
		exprs = append(exprs, fmt.Sprintf("* %d * * * * %d", ref.Minute(), ref.Year()))
		exprs = append(exprs, fmt.Sprintf("* * * * * * %d-%d", ref.Year()-1, ref.Year()+1))

		for _, expr := range gron.BatchDue(exprs) {
			if expr.Err != nil {
				t.Errorf("%s error: %#v", expr.Expr, expr.Err)
			}
			if !expr.Due {
				t.Errorf("%s must be due", expr.Expr)
			}
		}
	})

	t.Run("batch error", func(t *testing.T) {
		exprs := []string{"* * * *", "A B C D E F"}
		ref, _ := time.Parse(FullDateFormat, "2022-02-02 02:02:02")
		for _, expr := range gron.BatchDue(exprs, ref) {
			if expr.Err == nil {
				t.Errorf("%s expected error", expr.Expr)
			}
			if expr.Due {
				t.Errorf("%s must not be due when there is error", expr.Expr)
			}
		}
	})
}

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

func TestBatchDueMatchesIsDueWhenDomAndDowSet(t *testing.T) {
	g := New()
	expr := "0 0 1 * 1"
	refs := []time.Time{
		time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), // Monday, not 1st
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),  // 1st, Wednesday
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),  // 1st and Monday
	}
	for _, ref := range refs {
		due, err := g.IsDue(expr, ref)
		if err != nil {
			t.Fatalf("IsDue(%v): %v", ref, err)
		}
		b := g.BatchDue([]string{expr}, ref)[0]
		if b.Err != nil {
			t.Fatalf("BatchDue(%v): %v", ref, b.Err)
		}
		if b.Due != due {
			t.Fatalf("BatchDue/IsDue mismatch at %s: BatchDue=%v IsDue=%v",
				ref.Format("2006-01-02"), b.Due, due)
		}
	}
}

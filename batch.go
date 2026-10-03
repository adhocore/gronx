package gronx

import (
	"strings"
	"time"
)

// Expr represents an item in array for batch check
type Expr struct {
	Err  error
	Expr string
	Due  bool
}

// BatchDue checks if multiple expressions are due for given time (or now).
// It returns []Expr with filled in Due and Err values.
func (g *Gronx) BatchDue(exprs []string, ref ...time.Time) []Expr {
	ref = append(ref, time.Now())
	g.C.SetRef(ref[0])

	var segs []string

	cache, batch := map[string]Expr{}, make([]Expr, len(exprs))
	for i := range exprs {
		batch[i].Expr = exprs[i]
		segs, batch[i].Err = Segments(exprs[i])
		key := strings.Join(segs, " ")
		if batch[i].Err != nil {
			cache[key] = batch[i]
			continue
		}

		if c, ok := cache[key]; ok {
			batch[i] = c
			batch[i].Expr = exprs[i]
			continue
		}

		dueSegs := segs
		if isMinutePrecision(exprs[i]) {
			dueSegs = append([]string{}, segs...)
			dueSegs[0] = "*"
		}

		// Use SegmentsDue so day-of-month / weekday OR semantics match IsDue.
		batch[i].Due, batch[i].Err = g.SegmentsDue(dueSegs)
		cache[key] = batch[i]
	}
	return batch
}

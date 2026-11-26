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
		if batch[i].Err != nil {
			continue
		}

		if isMinutePrecision(exprs[i]) {
			segs[0] = "*"
		}

		// Key on the segments that are checked: "* * * * *" ignores seconds
		// but "0 * * * * *" does not, though both have the same segments.
		key := strings.Join(segs, " ")
		if c, ok := cache[key]; ok {
			batch[i] = c
			batch[i].Expr = exprs[i]
			continue
		}

		// Use SegmentsDue so day-of-month / weekday OR semantics match IsDue.
		batch[i].Due, batch[i].Err = g.SegmentsDue(segs)
		cache[key] = batch[i]
	}
	return batch
}

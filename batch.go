package gronx

import (
	"time"
)

// Expr represents an item in array for batch check
type Expr struct {
	Expr string
	Due  bool
	Err  error
	segs []string
}

// BatchDue checks if multiple expressions are due for given time (or now).
// It returns []Expr with filled in Due and Err values.
func (g *Gronx) BatchDue(exprs []string, ref ...time.Time) []Expr {
	if len(ref) > 0 {
		g.C.SetRef(ref[0])
	} else {
		g.C.SetRef(time.Now())
	}

	batch := make([]Expr, len(exprs))
	for i := range exprs {
		if batch[i].segs, batch[i].Err = Segments(exprs[i]); batch[i].Err != nil {
			continue
		}
		due := true
		for pos, seg := range batch[i].segs {
			if seg != "*" && seg != "?" {
				if due, batch[i].Err = g.C.CheckDue(seg, pos); !due || batch[i].Err != nil {
					break
				}
			}
		}
		batch[i].Due = due
	}
	return batch
}

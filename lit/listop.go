package lit

import "github.com/mb0/diff"

// ListOp represents a list operation used for list path edits. N > 0 means retain N elements and
// N < 0 means delete -N elements and N == 0 means insert V.
type ListOp struct {
	N int
	V []Val
}

func opsToList(ops []ListOp) *List {
	opv := make([]Val, 0, len(ops))
	for i, op := range ops {
		if op.N != 0 {
			if op.N > 0 && i == len(ops)-1 {
				break
			}
			opv = append(opv, Int(op.N))
		} else {
			opv = append(opv, &List{Vals: op.V})
		}
	}
	return &List{Vals: opv}
}

// diffCounts contains accumulated total count of elements and operations for each op kind.
type diffCounts struct {
	ret, retn int
	del, deln int
	ins, insn int
}

func (t *diffCounts) changed() bool  { return t.del > 0 || t.ins > 0 }
func (t *diffCounts) replaced() bool { return t.ret == 0 }

// diffToOps collects list operations and diff totals from a list of diff changes.
func diffToOps(chgs []diff.Change, a, b []Val) (ops []ListOp, t diffCounts) {
	ops = make([]ListOp, 0, len(chgs)*2)
	for _, c := range chgs {
		if cr := c.A - t.ret - t.del; cr > 0 {
			ops = append(ops, ListOp{N: cr})
			t.ret += cr
			t.retn++
		}
		if c.Del > 0 {
			ops = append(ops, ListOp{N: -c.Del})
			t.del += c.Del
			t.deln++
		}
		if c.Ins > 0 {
			ops = append(ops, ListOp{V: b[c.B : c.B+c.Ins]})
			t.ins += c.Ins
			t.insn++
		}
	}
	// we want the trailing retain op for analysis
	if cr := len(a) - t.ret + t.del; cr > 0 {
		ops = append(ops, ListOp{N: cr})
		t.ret += cr
		t.retn++
	}
	return
}

// toVals returns the idxr values of v or nil if v is not an idxr or has no values.
func toVals(v Val) ([]Val, bool) {
	if l, ok := v.(*List); ok {
		return l.Vals, true
	}
	if x, ok := v.(Idxr); ok {
		vs := make([]Val, 0, x.Len())
		x.IterIdx(func(idx int, el Val) error {
			vs = append(vs, el)
			return nil
		})
		return vs, true
	}
	return nil, false
}

type valsDiff struct{ a, b []Val }

func (d *valsDiff) Equal(i, j int) bool { return Equal(d.a[i], d.b[j]) }

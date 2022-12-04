package lit

import (
	"fmt"

	"github.com/mb0/diff"
)

type (
	// Ops is an abstract diff operation interface implemented by ListOps, StrOps and RawOps.
	Ops interface {
		Len() int
		Op(int) (int, Val)
	}

	// StrOp represents a string operation used for str diff edits. N > 0 means retain N runes
	// and N < 0 means delete -N runes and N == 0 means insert V.
	StrOp struct {
		N int
		V Str
	}
	StrOps []StrOp

	// RawOp represents a string operation used for raw diff edits. N > 0 means retain N bytes
	// and N < 0 means delete -N bytes and N == 0 means insert V.
	RawOp struct {
		N int
		V Raw
	}
	RawOps []RawOp

	// ListOp represents a list operation used for list diff edits. N > 0 means retain N
	// elements and N < 0 means delete -N elements and N == 0 means insert V.
	ListOp struct {
		N int
		V Vals
	}
	ListOps []ListOp
)

func (ops StrOps) Len() int  { return len(ops) }
func (ops RawOps) Len() int  { return len(ops) }
func (ops ListOps) Len() int { return len(ops) }

func (ops StrOps) Op(i int) (int, Val)  { return ops[i].N, Str(ops[i].V) }
func (ops RawOps) Op(i int) (int, Val)  { return ops[i].N, Raw(ops[i].V) }
func (ops ListOps) Op(i int) (int, Val) { return ops[i].N, &ops[i].V }

func opsToVals(ops Ops) *Vals {
	vs := make(Vals, ops.Len())
	for i := range vs {
		n, v := ops.Op(i)
		if n != 0 {
			if n > 0 && i == len(vs)-1 {
				vs = vs[:i]
				break
			}
			vs[i] = Int(n)
		} else {
			vs[i] = v
		}
	}
	return &vs
}

// diffStr diffs a and b and appends any str ops to d and returns the result or an error.
func diffStr(a, b Str, pre string, d Delta) (Delta, error) {
	ars := []rune(a)
	if chgs := diff.Runes(ars, []rune(b)); len(chgs) != 0 {
		ops := make(StrOps, 0, len(chgs)*2)
		t := diffToOps(chgs, len(ars), func(n, s, l int) {
			op := StrOp{N: n}
			if n == 0 {
				op.V = b[s : s+l]
			}
			ops = append(ops, op)
		})
		return t.diffRes(ops, b, pre, d, nil)
	}
	return d, nil
}

// diffRaw diffs a and b and appends any raw ops to d and returns the result or an error.
func diffRaw(a, b Raw, pre string, d Delta) (Delta, error) {
	if chgs := diff.Bytes(a, b); len(chgs) != 0 {
		ops := make(RawOps, 0, len(chgs)*2)
		t := diffToOps(chgs, len(a), func(n, s, l int) {
			op := RawOp{N: n}
			if n == 0 {
				op.V = b[s : s+l]
			}
			ops = append(ops, op)
		})
		return t.diffRes(ops, b, pre, d, nil)
	}
	return d, nil
}

// diffIdxr diffs a and b and appends any list ops to d and returns the result or an error.
func diffIdxr(a, b Idxr, pre string, d Delta) (Delta, error) {
	if chgs := diff.Diff(a.Len(), b.Len(), &idxrDiff{a, b}); len(chgs) != 0 {
		ops := make(ListOps, 0, len(chgs)*2)
		t := diffToOps(chgs, a.Len(), func(n, s, l int) {
			op := ListOp{N: n}
			if n == 0 {
				vs := make(Vals, 0, l)
				for i := 0; i < l; i++ {
					vs = append(vs, idx(b, s+i))
				}
				op.V = vs
			}
			ops = append(ops, op)
		})
		return t.diffRes(ops, b, pre, d, func(d Delta) (Delta, error) {
			// … we have three ops u,v,w. retn is 1. v is del or ins and either u or w
			// is the other or four ops u,v,w,x. retn is 2. v and w are del and ins
			u, v, w := ops[0], ops[1], ops[2]
			if v.N == -1 {
				if len(w.V) == 1 {
					return diffSub(idx(a, u.N), w.V[0], pre, u.N, d)
				}
				if t.retn == 1 && len(u.V) == 1 {
					return diffSub(idx(a, 0), u.V[0], pre, 0, d)
				}
			} else if len(v.V) == 1 {
				if w.N == -1 {
					return diffSub(idx(a, u.N), v.V[0], pre, u.N, d)
				}
				if t.retn == 1 && u.N == -1 {
					return diffSub(idx(a, 0), v.V[0], pre, 0, d)
				}
			}
			return d, addMarker
		})
	}
	return d, nil
}

func diffSub(a, b Val, pre string, idx int, d Delta) (Delta, error) {
	return diffVals(a, b, fmt.Sprintf("%s%d.", pre, idx), d)
}

// diffCounts contains accumulated total count of elements and operations for each op kind.
type diffCounts struct {
	ret, retn int
	del, deln int
	ins, insn int
}

func (t *diffCounts) changed() bool  { return t.del > 0 || t.ins > 0 }
func (t *diffCounts) replaced() bool { return t.ret == 0 }

func (t *diffCounts) diffRes(ops Ops, b Val, pre string, d Delta, hook idxHook) (Delta, error) {
	if !t.changed() {
		return d, nil
	} else if t.replaced() {
		d = append(d, KeyVal{stripTailDot(pre), b})
		return d, nil
	}
	// we have at least two ops and known at least one of them to be ret and one del or ins
	// ops of the same kind are merged and do not follow each other
	oLen := ops.Len()
	o0N, _ := ops.Op(0)
	o1N, o1V := ops.Op(1)

	// we want to detect append and use special syntax. append does only occur when we have
	// two ops u,v where u is ret and v is ins
	if oLen == 2 && o0N > 0 && o1N == 0 {
		// lets return the special append op
		d = append(d, KeyVal{stripTailDot(pre) + "+", o1V})
		return d, nil
	}
	// we also want to detect replacing a single element and use idx path notation. that does
	// only occur in two instances:
	// we have three ops u,v,w. retn is 1. v is del or ins and either u or w is the other
	// and we have four ops u,v,w,x. retn is 2. v and w are del and ins
	if hook != nil && (oLen == 3 && t.retn == 1 || oLen == 4 && t.retn == 2) {
		if res, err := hook(d); err != addMarker {
			return res, err
		}
	}
	// lets return the ops as list
	d = append(d, KeyVal{stripTailDot(pre) + "*", opsToVals(ops)})
	return d, nil
}

var addMarker = fmt.Errorf("add")

type idxHook func(Delta) (Delta, error)

func diffToOps(chgs []diff.Change, alen int, add func(n, s, l int)) (t diffCounts) {
	for _, c := range chgs {
		if cr := c.A - t.ret - t.del; cr > 0 {
			add(cr, 0, 0)
			t.ret += cr
			t.retn++
		}
		if c.Del > 0 {
			add(-c.Del, 0, 0)
			t.del += c.Del
			t.deln++
		}
		if c.Ins > 0 {
			add(0, c.B, c.Ins)
			t.ins += c.Ins
			t.insn++
		}
	}
	// we want the trailing retain op for analysis
	if cr := alen - t.ret + t.del; cr > 0 {
		add(cr, 0, 0)
		t.ret += cr
		t.retn++
	}
	return
}

type idxrDiff struct{ a, b Idxr }

func (d *idxrDiff) Equal(i, j int) bool {
	return Equal(idx(d.a, i), idx(d.b, j))
}

func idx(a Idxr, i int) Val { v, _ := a.Idx(i); return v }

package ext

import (
	"fmt"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Node is an interface for mutable obj literal values that have a corresponding form.
type Node interface {
	lit.Keyr
	lit.Idxr
}

// NewNode returns a node for val or an error. It either accepts a Node or creates a new proxy.
func NewNode(c lit.Reg, val interface{}) (Node, error) {
	n, ok := val.(Node)
	if !ok {
		p, err := lit.Proxy(c, val)
		if err != nil {
			return nil, fmt.Errorf("proxy %T: %w", val, err)
		}
		n, ok = p.(Node)
		if !ok {
			return nil, fmt.Errorf("want node got %T", p)
		}
	}
	return n, nil
}

type NodeSpec struct {
	exp.SpecBase
	Rules
	Node Node
}

func NodeSpecSig(reg lit.Reg, sig string, val interface{}, rs Rules) (*NodeSpec, error) {
	decl, err := typ.Parse(sig)
	if err != nil {
		return nil, err
	}
	node, err := NewNode(reg, val)
	if err != nil {
		return nil, err
	}
	return NewNodeSpec(decl, node, rs), nil
}

func NodeSpecName(reg lit.Reg, name string, val interface{}, rs Rules) (*NodeSpec, error) {
	node, err := NewNode(reg, val)
	if err != nil {
		return nil, err
	}
	ps := make([]typ.Param, 0, 4)
	if rs.IdxKeyer != nil {
		ps = append(ps, typ.P("", typ.Opt(typ.Tupl)))
	}
	ps = append(ps, typ.P("", typ.Opt(typ.ElemTupl(typ.Tag))))
	if rs.Tail.Setter != nil {
		ps = append(ps, typ.P("", typ.Opt(typ.ElemTupl(typ.Exp))))
	}
	ps = append(ps, typ.P("", node.Type()))
	return NewNodeSpec(typ.Form(name, ps...), node, rs), nil
}

func NewNodeSpec(decl typ.Type, node Node, rs Rules) *NodeSpec {
	res := exp.SigRes(decl)
	res.Type = node.Type()
	return &NodeSpec{exp.SpecBase{Decl: decl}, rs, node}
}

func (s *NodeSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	l, err := lit.Clone(s.Node)
	if err != nil {
		return nil, err
	}
	n := l.(Node)
	for i, sp := range exp.SigArgs(c.Sig) {
		switch a := c.Args[i].(type) {
		case nil:
			continue
		case *exp.Tupl:
			if len(a.Els) == 0 {
				continue
			}
			et, _ := typ.TuplEl(a.Type)
			ek := typ.Deopt(et).Kind
			switch {
			case ek == knd.All: // tupl -> idx rule
				for idx, d := range a.Els {
					key := s.IdxKeyer(n, idx)
					_, err := s.Rules.Eval(p, c.Env, n, key, d)
					if err != nil {
						return nil, err
					}
				}
			case ek == knd.Tag || sp.Name == "tags": // tupl|tag, tags:tupl -> tag rule
				for _, d := range a.Els {
					t, ok := d.(*exp.Tag)
					var err error
					if ok {
						_, err = s.Rules.Eval(p, c.Env, n, t.Tag, t.Exp)
					} else {
						_, err = s.Rules.Eval(p, c.Env, n, "", d)
					}
					if err != nil {
						return nil, err
					}
				}
			case ek == knd.Exp: // tupl|exp -> tail rule
				if s.Tail.Prepper == nil {
					return nil, fmt.Errorf("tail without prepper")
				}
				v, err := s.Tail.Prepper(p, c.Env, n, "", a)
				if err != nil {
					return nil, fmt.Errorf("tail prep: %v", err)
				}
				if v != nil {
					if s.Tail.Prepper == nil {
						return nil, fmt.Errorf("tail without setter")
					}
					err = s.Tail.Setter(p, n, "", v)
					if err != nil {
						return nil, fmt.Errorf("setkey tail: %v", err)
					}
				}
			default:
			}
		case *exp.Tag:
			_, err := s.Rules.Eval(p, c.Env, n, a.Tag, a.Exp)
			if err != nil {
				return nil, err
			}
		default:
			_, err := s.Rules.Eval(p, c.Env, n, sp.Name, a)
			if err != nil {
				return nil, err
			}
			continue
		}
	}
	return exp.LitVal(n), nil
}

func (rs Rules) Eval(p *exp.Prog, env exp.Env, n Node, key string, e exp.Exp) (lit.Val, error) {
	r := rs.Rule(key)
	v, err := r.Prepper(p, env, n, key, e)
	if err != nil {
		return nil, fmt.Errorf("prep key %s: %v", key, err)
	}
	if v != nil {
		err = r.Setter(p, n, key, v)
		if err != nil {
			return nil, fmt.Errorf("set key %s: %v", key, err)
		}
	}
	return v, nil
}

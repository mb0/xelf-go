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
	WithReg(*lit.Reg)
}

// NewNode returns a node for val or an error. It either accepts a Node or creates a new proxy.
func NewNode(reg *lit.Reg, val interface{}) (Node, error) {
	n, ok := val.(Node)
	if ok {
		n.WithReg(reg)
	} else {
		p, err := reg.Proxy(val)
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
	Env  bool
	Sub  func(string) exp.Spec
}

func NodeSpecSig(reg *lit.Reg, sig string, val interface{}, rs Rules) (*NodeSpec, error) {
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

func NodeSpecName(reg *lit.Reg, name string, val interface{}, rs Rules) (*NodeSpec, error) {
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
	return &NodeSpec{exp.SpecBase{Decl: decl}, rs, node, false, nil}
}

func (s *NodeSpec) Value() lit.Val { return s }
func copyNode(reg *lit.Reg, node Node) (Node, error) {
	mut, err := node.New()
	if err != nil {
		return nil, err
	}
	n := mut.(Node)
	n.WithReg(reg)
	err = n.Assign(node)
	return n, err
}
func (s *NodeSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	if s.Env {
		n, err := copyNode(p.Reg, s.Node)
		if err != nil {
			return nil, err
		}
		c.Env = &NodeEnv{Par: env, Node: n, Spec: s.Sub}
	}
	res, err := s.SpecBase.Resl(p, env, c, h)
	if err != nil {
		return nil, err
	}
	if s.ReslHook != nil {
		err = s.ReslHook(p, c)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s *NodeSpec) GetNode(p *exp.Prog, c *exp.Call) (Node, error) {
	if s.Env && c != nil {
		if env, ok := c.Env.(*NodeEnv); ok {
			return env.Node, nil
		}
		return nil, fmt.Errorf("expect node env in call got %T", c.Env)
	}
	return copyNode(p.Reg, s.Node)
}

func (s *NodeSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	n, err := s.GetNode(p, c)
	if err != nil {
		return nil, fmt.Errorf("get node: %v", err)
	}
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
					err := s.dokey(p, c, n, key, d)
					if err != nil {
						return nil, err
					}
				}
			case ek == knd.Tag || sp.Name == "tags": // tupl|tag, tags:tupl -> tag rule
				for _, d := range a.Els {
					t, ok := d.(*exp.Tag)
					var err error
					if ok {
						err = s.dokey(p, c, n, t.Tag, t.Exp)
					} else {
						err = s.dokey(p, c, n, "", d)
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
			err := s.dokey(p, c, n, a.Tag, a.Exp)
			if err != nil {
				return nil, err
			}
		default:
			err := s.dokey(p, c, n, sp.Name, a)
			if err != nil {
				return nil, err
			}
			continue
		}
	}
	return &exp.Lit{Res: n.Type(), Val: n}, nil
}
func (s *NodeSpec) dokey(p *exp.Prog, c *exp.Call, prx Node, key string, el exp.Exp) error {
	r := s.Rule(key)
	v, err := r.Prepper(p, c.Env, prx, key, el)
	if err != nil {
		return fmt.Errorf("prep key %s: %v", key, err)
	}
	if v != nil {
		err = r.Setter(p, prx, key, v)
		if err != nil {
			return fmt.Errorf("set key %s: %v", key, err)
		}
	}
	return nil
}

type NodeEnv struct {
	Par  exp.Env
	Node Node
	Spec func(string) exp.Spec
}

func (e *NodeEnv) Parent() exp.Env { return e.Par }
func (e *NodeEnv) Resl(p *exp.Prog, s *exp.Sym, k string) (exp.Exp, error) {
	if e.Spec != nil {
		if s := e.Spec(k); s != nil {
			return &exp.Lit{Res: s.Type(), Val: s}, nil
		}
	}
	return e.Par.Resl(p, s, k)
}
func (e *NodeEnv) Eval(p *exp.Prog, s *exp.Sym, k string) (*exp.Lit, error) {
	if e.Spec != nil {
		if s := e.Spec(k); s != nil {
			return &exp.Lit{Res: s.Type(), Val: s}, nil
		}
	}
	return e.Par.Eval(p, s, k)
}
func (e *NodeEnv) Dyn() exp.Spec { return e.Par.Dyn() }

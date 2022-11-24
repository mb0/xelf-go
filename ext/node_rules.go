package ext

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Rule is a configurable helper for assigning tags to nodes.
type Rule struct {
	Prepper KeyPrepper
	Setter  KeySetter
}

// KeyPrepper resolves el and returns a value to be assigned to key or an error.
type KeyPrepper = func(p *exp.Prog, env exp.Env, n Node, key string, el exp.Exp) (lit.Val, error)

// KeySetter sets the property with key on not to value v or returns an error.
type KeySetter = func(p *exp.Prog, n Node, key string, v lit.Val) error

// IdxKeyer returns a key for an unnamed argument at idx.
type IdxKeyer = func(n Node, idx int) string

// Default is an alias for the default rule.
type Default = Rule

// Rules is a collection of rules that assign tags to nodes.
type Rules struct {
	// Key holds optional per key rules
	Key map[string]Rule
	// IdxKeyer will map unnamed tags to a key, when null unnamed tags result in an error
	IdxKeyer
	// Default holds an optional default rule.
	// If neither specific nor default rule are found DynPrepper and PathSetter are used.
	Default
	// Tail holds optional rules for tail elements.
	Tail Rule
}

func (rs *Rules) Rule(tag string) Rule {
	r := rs.Key[tag]
	if r.Prepper == nil {
		r.Prepper = DynPrepper
		if rs.Prepper != nil {
			r.Prepper = rs.Prepper
		}
	}
	if r.Setter == nil {
		r.Setter = PathSetter
		if rs.Setter != nil {
			r.Setter = rs.Setter
		}
	}
	return r
}

// ZeroKeyer is an index keyer without offset.
var ZeroKeyer = OffsetKeyer(0)

// OffsetKeyer returns an index keyer that looks up a field at the index plus the offset.
func OffsetKeyer(offset int) IdxKeyer {
	return func(n Node, i int) string {
		pb := n.Type().Body.(*typ.ParamBody)
		p := pb.Params[i+offset]
		return p.Key
	}
}

// ListPrepper resolves args using p and env and returns a list or an error.
func ListPrepper(p *exp.Prog, env exp.Env, _ Node, _ string, arg exp.Exp) (lit.Val, error) {
	res := &lit.List{}
	switch a := arg.(type) {
	case *exp.Tupl:
		res.Vals = make([]lit.Val, 0, len(a.Els))
		for _, el := range a.Els {
			aa, err := p.Eval(env, el)
			if err != nil {
				return nil, err
			}
			if !aa.Val.Zero() {
				res.Vals = append(res.Vals, aa.Val)
			}
		}
	default:
		aa, err := p.Eval(env, a)
		if err != nil {
			return nil, err
		}
		res.Vals = []lit.Val{aa.Val}
	}
	return res, nil
}

// DynPrepper resolves args using p and env and returns a value or an error.
// Empty args result in an untyped null value. Multiple args are resolved as call.
func DynPrepper(p *exp.Prog, env exp.Env, _ Node, _ string, arg exp.Exp) (_ lit.Val, err error) {
	switch a := arg.(type) {
	case nil:
		return lit.Bool(true), nil
	case *exp.Tupl:
		if len(a.Els) == 0 {
			return lit.Null{}, nil
		}
		if len(a.Els) == 1 {
			arg = a.Els[0]
		} else {
			arg = &exp.Call{Args: a.Els, Src: a.Src}
		}
		arg, err = p.Resl(env, arg, typ.Void)
		if err != nil {
			return nil, err
		}
	}
	a, err := p.Eval(env, arg)
	if err != nil {
		return nil, err
	}
	return a.Val, nil
}

// BitsPrepper returns a key prepper that tries to resolve a bits constant.
func BitsPrepper(consts []typ.Const) KeyPrepper {
	return func(p *exp.Prog, env exp.Env, n Node, key string, arg exp.Exp) (lit.Val, error) {
		v, err := DynPrepper(p, env, n, key, arg)
		if err != nil {
			return v, err
		}
		if v.Zero() && !v.Nil() {
			return lit.Int(0), nil
		}
		for _, b := range consts {
			if key == cor.Keyed(b.Name) {
				return lit.Int(b.Val), nil
			}
		}
		return nil, fmt.Errorf("no constant named %q", key)
	}
}

// PathSetter sets el to n using key as path or returns an error.
func PathSetter(p *exp.Prog, n Node, key string, v lit.Val) error {
	path, err := cor.ParsePath(key)
	if err != nil {
		return fmt.Errorf("parse %s: %w", key, err)
	}
	err = lit.CreatePath(n, path, v)
	if err != nil {
		return err
	}
	return nil
}

// ExtraSetter sets el to n using key as path or returns an error.
func ExtraSetter(m string) KeySetter {
	return func(p *exp.Prog, n Node, key string, v lit.Val) error {
		path, err := cor.ParsePath(key)
		if err != nil {
			return fmt.Errorf("parse %s: %w", key, err)
		}
		v = v.Value()
		err = lit.CreatePath(n, path, v)
		if err != nil {
			path = append(cor.Path{{Key: m}}, path...)
			e := lit.CreatePath(n, path, v)
			if e == nil {
				return nil
			}
		}
		return err
	}
}

// BitsSetter returns a key setter that tries to add to a node bits field with key.
func BitsSetter(b string) KeySetter {
	return func(p *exp.Prog, n Node, _ string, v lit.Val) error {
		f, err := n.Key(b)
		if err != nil {
			return err
		}
		fi, err := lit.ToInt(f)
		if err != nil {
			return err
		}
		vi, err := lit.ToInt(v)
		if err != nil {
			return err
		}
		return n.SetKey(b, lit.Int(uint64(fi)|uint64(vi)))
	}
}

var norule Rule

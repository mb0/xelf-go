// Package ext contains helpers to extend a xelf language with go function specs and node forms.
package ext

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Func is func spec implementation calling an native go function using reflection.
type Func struct {
	exp.SpecBase
	val  reflect.Value
	rts  []reflect.Type
	vari bool
	err  bool
}

// NewFunc reflects function value val and returns a named func spec or an error.
// The variadic names parameter can be specified for parameter names that is elided from the
// reflect function type information.
func NewFunc(reg lit.Reg, name string, val interface{}, names ...string) (*Func, error) {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Func {
		return nil, fmt.Errorf("expect function argument got %T", val)
	}
	if reg == nil {
		reg = lit.GlobalRegs()
	}
	t := v.Type()
	n := t.NumIn()
	pb := &typ.ParamBody{}
	s := &Func{SpecBase: exp.SpecBase{Decl: typ.Type{Kind: knd.Func, Ref: name, Body: pb}},
		val: v, rts: make([]reflect.Type, 0, n), vari: t.IsVariadic(),
	}
	for i := 0; i < n; i++ {
		rt := t.In(i)
		lt, err := reg.Reflect(rt)
		if err != nil {
			return nil, err
		}
		var name string
		if i < len(names) {
			name = names[i]
		}
		s.rts = append(s.rts, rt)
		pb.Params = append(pb.Params, typ.P(name, lt))
	}
	n = t.NumOut()
	var res typ.Type
	for i := 0; i < n; i++ {
		rt := t.Out(i)
		if rt.ConvertibleTo(refErr) {
			s.err = true
			if i+1 < n {
				return nil, fmt.Errorf("error can only be last result")
			}
			break
		}
		if i > 0 {
			return nil, fmt.Errorf("expect at most one result and maybe an error")
		}
		lt, err := reg.Reflect(rt)
		if err != nil {
			return nil, err
		}
		res = lt
	}
	pb.Params = append(pb.Params, typ.P("", res))
	return s, nil
}

func (s *Func) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	// get reflect values from argument
	rvs := make([]reflect.Value, len(s.rts))
	for i, rt := range s.rts {
		arg := args[i]
		if arg == nil || arg.Zero() {
			ptr := reflect.New(rt)
			rvs[i] = ptr.Elem()
			// reflect already provides a zero value
			continue
		}
		val, err := lit.Conv(p.Reg, rt, arg)
		if err != nil {
			return nil, err
		}
		rvs[i] = val
	}
	// call reflect function with value
	var res []reflect.Value
	if s.vari {
		res = s.val.CallSlice(rvs)
	} else {
		res = s.val.Call(rvs)
	}
	if s.err { // check last result
		last := res[len(res)-1]
		if !last.IsNil() {
			return nil, last.Interface().(error)
		}
		res = res[:len(res)-1]
	}
	if len(res) == 0 { // nothing to return
		return nil, nil
	}
	rv := res[0]
	if rv.Type().Kind() != reflect.Ptr {
		rn := reflect.New(rv.Type())
		rn.Elem().Set(rv)
		rv = rn
	}
	// create a proxy from the result and return
	prx, err := p.Reg.ProxyValue(rv)
	if err != nil {
		return nil, err
	}
	return lit.Wrap(prx, exp.SigRes(c.Sig).Type), nil
}

var refErr = reflect.TypeOf((*error)(nil)).Elem()

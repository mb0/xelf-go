package lit

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Prx is the interface for all reflection based mutable values.
type Prx interface {
	Mut
	// Reflect returns the reflect value pointed to by this proxy.
	Reflect() reflect.Value
	// NewWith returns a new proxy instance with ptr as value.
	// This method is used internally for proxy caching and should only be called with pointer
	// values known to be compatible with this proxy implementation.
	NewWith(ptr reflect.Value) (Mut, error)
}

type proxy struct {
	Reg *Reg
	typ typ.Type
	val reflect.Value
}

func (x *proxy) Type() typ.Type               { return x.typ }
func (x *proxy) Ptr() interface{}             { return x.val.Interface() }
func (x *proxy) Reflect() reflect.Value       { return x.val.Elem() }
func (x *proxy) Nil() bool                    { return false }
func (x *proxy) new() reflect.Value           { return reflect.New(x.val.Type().Elem()) }
func (x *proxy) with(ptr reflect.Value) proxy { return proxy{x.Reg, x.typ, ptr} }
func (x *proxy) WithReg(reg *Reg)             { x.Reg = reg }
func (x *proxy) unmarshal(b []byte, mut Mut) error {
	return x.Reg.ReadIntoMut(bytes.NewReader(b), "", mut)
}

type IntPrx struct{ proxy }

func (x *IntPrx) NewWith(v reflect.Value) (Mut, error) { return &IntPrx{x.with(v)}, nil }
func (x *IntPrx) New() (Mut, error)                    { return x.NewWith(x.new()) }

func (x *IntPrx) Zero() bool { return x.value() == 0 }
func (x *IntPrx) Value() Val { return Int(x.value()) }
func (x *IntPrx) Parse(a ast.Ast) error {
	if a.Kind != knd.Num {
		return fmt.Errorf("expect num")
	}
	n, err := strconv.ParseInt(a.Raw, 10, 64)
	if err != nil {
		return err
	}
	x.Reflect().SetInt(n)
	return nil
}
func (x *IntPrx) Assign(v Val) error {
	n, err := ToInt(v)
	if err != nil {
		return err
	}
	switch e := x.Reflect(); e.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16:
		e.SetInt(int64(n))
	case reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16:
		e.SetUint(uint64(n))
	default:
		return fmt.Errorf("unexpected int proxy target %s", e.Type())
	}
	return nil
}
func (x *IntPrx) value() int64 {
	switch e := x.Reflect(); e.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int32:
		return e.Int()
	case reflect.Uint64, reflect.Uint, reflect.Uint32:
		return int64(e.Uint())
	default:
		panic(fmt.Errorf("unexpected int proxy target %s", e.Type()))
	}
}
func (x *IntPrx) String() string               { return fmt.Sprintf("%d", x.value()) }
func (x *IntPrx) MarshalJSON() ([]byte, error) { return []byte(x.String()), nil }
func (x *IntPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *IntPrx) Print(p *bfr.P) error         { return p.Fmt("%d", x.value()) }

type RealPrx struct{ proxy }

func (x *RealPrx) NewWith(v reflect.Value) (Mut, error) { return &RealPrx{x.with(v)}, nil }
func (x *RealPrx) New() (Mut, error)                    { return x.NewWith(x.new()) }

func (x *RealPrx) Zero() bool { return x.value() == 0 }
func (x *RealPrx) Value() Val { return Real(x.value()) }
func (x *RealPrx) Parse(a ast.Ast) error {
	if a.Kind != knd.Real && a.Kind != knd.Num {
		return fmt.Errorf("expect num")
	}
	n, err := strconv.ParseFloat(a.Raw, 64)
	if err != nil {
		return err
	}
	x.Reflect().SetFloat(n)
	return nil
}
func (x *RealPrx) Assign(v Val) error {
	n, err := ToReal(v)
	if err != nil {
		return err
	}
	switch e := x.Reflect(); e.Kind() {
	case reflect.Float64, reflect.Float32:
		e.SetFloat(float64(n))
	default:
		return fmt.Errorf("unexpected real proxy target %s", e.Type())
	}
	return nil
}
func (x *RealPrx) value() float64 {
	switch e := x.Reflect(); e.Kind() {
	case reflect.Float64, reflect.Float32:
		return e.Float()
	default:
		panic(fmt.Errorf("unexpected real proxy target %s", e.Type()))
	}
}
func (x *RealPrx) String() string               { return fmt.Sprintf("%g", x.value()) }
func (x *RealPrx) MarshalJSON() ([]byte, error) { return []byte(x.String()), nil }
func (x *RealPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *RealPrx) Print(p *bfr.P) error         { return p.Fmt("%g", x.value()) }

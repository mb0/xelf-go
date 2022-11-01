package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/typ"
)

// Conv converts val to type t and returns a reflect value or an error.
func Conv(reg typ.Reg, t reflect.Type, val Val) (reflect.Value, error) {
	ptr := reflect.New(t)
	mut, err := reg.ProxyValue(ptr)
	if err != nil {
		return reflect.Value{}, err
	}
	if !val.Nil() {
		err = mut.Assign(val)
		if err != nil {
			return reflect.Value{}, err
		}
	}
	return ptr.Elem(), nil
}

// MustProxy returns a proxy value for ptr or panics.
func MustProxy(reg typ.Reg, ptr interface{}) Mut {
	mut, err := reg.Proxy(ptr)
	if err != nil {
		panic(err)
	}
	return mut
}

// Proxy returns a proxy value for ptr or an error.
func (reg *Reg) Proxy(ptr interface{}) (Mut, error) {
	return reg.ProxyValue(reflect.ValueOf(ptr))
}

// Reflect returns the xelf type for the reflect type or an error.
func (reg *Reg) Reflect(t reflect.Type) (typ.Type, error) {
	reg.init()
	return reg.Cache.Reflect(t)
}

// ProxyValue returns a proxy value for the reflect value ptr or an error.
func (reg *Reg) ProxyValue(ptr reflect.Value) (mut Mut, err error) {
	if ptr.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("requires pointer got %s", ptr.Type())
	}
	reg.init()
	pt, org := ptr.Type(), ptr
	if isMut(pt) {
		return checkMut(pt, ptr, ptr)
	}
	var opt, null bool
	et, el := pt.Elem(), ptr.Elem()
	if opt = et.Kind() == reflect.Ptr; opt {
		pt, ptr, et, el = et, el, et.Elem(), el.Elem()
		if isMut(pt) {
			return checkMut(pt, ptr, org)
		}
		null = ptr.IsNil()
	}
	if prx, ok := reg.Cache.Proxy(pt); ok {
		return prx.NewWith(org)
	}
	switch et.Kind() {
	case reflect.Bool:
		if v, ok := toRef(ptrBool, ptr, org); ok {
			mut = v.Interface().(*Bool)
			if opt {
				mut = &OptMut{mut, &org, null}
			}
			return mut, nil
		}
	case reflect.Int64:
		if et != typInt64 && et.Implements(ptrSecs.Elem()) {
			if v, ok := toRef(ptrSpan, ptr, org); ok {
				mut = v.Interface().(*Span)
				if opt {
					mut = &OptMut{mut, &org, null}
				}
				return mut, nil
			}
		}
		if !opt {
			if v, ok := toRef(ptrInt, ptr, noval); ok {
				return v.Interface().(*Int), nil
			}
			if v, ok := toRef(ptrNum, ptr, noval); ok {
				return v.Interface().(*Num), nil
			}
		}
		fallthrough
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Uint64, reflect.Uint32, reflect.Uint16:
		mut = &IntPrx{newProxy(reg, typ.Int, org)}
	case reflect.Float64:
		if !opt {
			if v, ok := toRef(ptrReal, ptr, noval); ok {
				return v.Interface().(*Real), nil
			}
		}
		fallthrough
	case reflect.Float32:
		mut = &RealPrx{newProxy(reg, typ.Real, org)}
	case reflect.String:
		if v, ok := toRef(ptrStr, ptr, org); ok {
			mut = v.Interface().(*Str)
		}
		if v, ok := toRef(ptrChar, ptr, org); ok {
			mut = v.Interface().(*Char)
		}
	case reflect.Slice:
		if et.Elem().Kind() == reflect.Uint8 {
			if v, ok := toRef(ptrRaw, ptr, org); ok {
				mut = v.Interface().(*Raw)
				if opt {
					mut = &OptMut{mut, &org, null}
				}
				return mut, nil
			}
		}
		et, err := reg.Cache.Reflect(et.Elem())
		if err != nil {
			return nil, err
		}
		mut = &ListPrx{newProxy(reg, typ.ListOf(et), org)}
	case reflect.Array:
		if v, ok := toRef(ptrUUID, ptr, org); ok {
			mut = v.Interface().(*UUID)
			if opt {
				mut = &OptMut{mut, &org, null}
			}
			return mut, nil
		}
	case reflect.Struct:
		if v, ok := toRef(ptrTime, ptr, org); ok {
			mut = v.Interface().(*Time)
			if opt {
				mut = &OptMut{mut, &org, null}
			}
			return mut, nil
		}
		if v, ok := toRef(ptrType, ptr, org); ok {
			mut = v.Interface().(*typ.Type)
			if opt {
				mut = &OptMut{mut, &org, null}
			}
			return mut, nil
		}
		nfo, ok := reg.Cache.Param(et)
		if !ok {
			rt, pm, err := reg.Cache.ReflectStruct(et)
			if err != nil {
				return nil, err
			}
			// because we call into reflectStruct we need to register the type
			nfo = typInfo{rt, pm}
			reg.Cache.SetParam(et, nfo)
		}
		// we use reflect struct because the proxy also needs the param map
		mut = &ObjPrx{proxy: newProxy(reg, nfo.Type, org), params: nfo.params}
	case reflect.Map:
		if et.Key().Kind() != reflect.String {
			break
		}
		et, err := reg.Cache.Reflect(et.Elem())
		if err != nil {
			return nil, err
		}
		mut = &MapPrx{proxy: newProxy(reg, typ.DictOf(et), org)}
	case reflect.Interface:
		mut = &AnyPrx{proxy{reg, typ.Any, org}, anyVal(org)}
	}
	if mut == nil {
		return nil, fmt.Errorf("cannot proxy type %s", ptr.Type())
	}
	pp, _ := mut.New()
	if prx, ok := pp.(Prx); ok {
		rt := prx.Reflect().Type()
		if rt.Kind() != reflect.Ptr {
			pv := reflect.New(rt)
			pp, _ = prx.NewWith(pv)
			prx, rt = pp.(Prx), pv.Type()
		}
		reg.Cache.SetProxy(rt, prx)
	}
	return mut, nil
}

func toRef(ref reflect.Type, v, org reflect.Value) (reflect.Value, bool) {
	t := v.Type()
	if t == ref {
		return v, true
	}
	if t.ConvertibleTo(ref) {
		if org.IsValid() {
			if v.IsNil() {
				v, _ = newPrimMut(org, v)
			}
		}
		return v.Convert(ref), true
	}
	return v, false
}

func isMut(t reflect.Type) bool { return t.NumMethod() > 10 && t.Implements(ptrMut.Elem()) }
func isPrx(t reflect.Type) bool { return t.NumMethod() > 12 && t.Implements(ptrPrx.Elem()) }

func checkMut(t reflect.Type, ptr, org reflect.Value) (Mut, error) {
	if ptr.IsNil() {
		if isPrx(t) {
			return nil, fmt.Errorf("cannot use nil proxy %s", t)
		}
		ptr, _ = newPrimMut(org, ptr)
		return &OptMut{ptr.Interface().(Mut), &org, true}, nil
	}
	return ptr.Interface().(Mut), nil
}

func newPrimMut(org, el reflect.Value) (ptr, val reflect.Value) {
	if el.IsNil() {
		el = reflect.New(org.Type().Elem().Elem())
		org.Elem().Set(el)
	}
	return el, el.Elem()
}

var noval reflect.Value

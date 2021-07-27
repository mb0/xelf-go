package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/typ"
)

// Conv converts val to type t and returns a reflect value or an error.
func (reg *Reg) Conv(t reflect.Type, val Val) (reflect.Value, error) {
	ptr := reflect.New(t)
	mut, err := reg.ProxyValue(ptr)
	if err != nil {
		return reflect.Value{}, err
	}
	err = mut.Assign(val)
	if err != nil {
		return reflect.Value{}, err
	}
	return ptr.Elem(), nil
}

// Proxy returns a proxy value for ptr or an error.
func (reg *Reg) Proxy(ptr interface{}) (Mut, error) {
	return reg.ProxyValue(reflect.ValueOf(ptr))
}

// MustProxy returns a proxy value for ptr or panics.
func (reg *Reg) MustProxy(ptr interface{}) Mut {
	mut, err := reg.Proxy(ptr)
	if err != nil {
		panic(err)
	}
	return mut
}

// ProxyValue returns a proxy value for the reflect value ptr or an error.
func (reg *Reg) ProxyValue(ptr reflect.Value) (Mut, error) {
	if ptr.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("requires pointer")
	}
	org := ptr
	pt := ptr.Type()
	if pt.NumMethod() >= 10 && pt.Implements(ptrMut.Elem()) {
		if ptr.IsNil() {
			return &OptMut{ptr.Interface().(Mut), &org, true}, nil
		}
		return ptr.Interface().(Mut), nil
	}
	et := pt.Elem()
	var opt, null bool
	if opt = et.Kind() == reflect.Ptr; opt {
		ptr = ptr.Elem()
		pt, et = et, et.Elem()
		if null = ptr.IsNil(); null {
			ptr = reflect.New(et)
		}
		if pt.NumMethod() >= 10 && pt.Implements(ptrMut.Elem()) {
			if null || ptr.IsNil() {
				return &OptMut{ptr.Interface().(Mut), &org, true}, nil
			}
			return ptr.Interface().(Mut), nil
		}
	}
	prx, ok := reg.proxy[et]
	if ok {
		res, err := prx.NewWith(ptr)
		if err != nil {
			return nil, err
		}
		if opt {
			res = &OptMut{res, &org, null}
		}
		return res, nil
	}
	switch et.Kind() {
	case reflect.Bool:
		if v, ok := toRef(ptrBool, ptr); ok {
			prx = v.Interface().(*Bool)
		}
	case reflect.Int64:
		if et != typInt64 && et.Implements(ptrSecs.Elem()) {
			if v, ok := toRef(ptrSpan, ptr); ok {
				prx = v.Interface().(*Span)
			}
			break
		}
		if v, ok := toRef(ptrInt, ptr); ok {
			prx = v.Interface().(*Int)
			break
		}
		fallthrough
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Uint64, reflect.Uint32, reflect.Uint16:
		prx = &IntPrx{proxy{reg, typ.Int, ptr}}
	case reflect.Float64:
		if v, ok := toRef(ptrReal, ptr); ok {
			prx = v.Interface().(*Real)
			break
		}
		fallthrough
	case reflect.Float32:
		prx = &RealPrx{proxy{reg, typ.Real, ptr}}
	case reflect.String:
		if v, ok := toRef(ptrStr, ptr); ok {
			prx = v.Interface().(*Str)
		}
	case reflect.Slice:
		if et.Elem().Kind() == reflect.Uint8 {
			if v, ok := toRef(ptrRaw, ptr); ok {
				prx = v.Interface().(*Raw)
				break
			}
		}
		et, err := reg.Reflect(et.Elem())
		if err != nil {
			return nil, err
		}
		prx = &ListPrx{proxy{reg, typ.ListOf(et), ptr}}
	case reflect.Array:
		if v, ok := toRef(ptrUUID, ptr); ok {
			prx = v.Interface().(*UUID)
		}
	case reflect.Struct:
		if v, ok := toRef(ptrTime, ptr); ok {
			prx = v.Interface().(*Time)
			break
		}
		if v, ok := toRef(ptrType, ptr); ok {
			prx = v.Interface().(*typ.Type)
		}
		nfo, ok := reg.param[et]
		if !ok {
			rt, pm, err := reg.reflectStruct(et, new(tstack))
			if err != nil {
				return nil, err
			}
			// because we call into reflectStruct we need to register the type
			nfo = typInfo{rt, pm}
			reg.setParam(et, nfo)
		}
		// we use reflect struct beacue the proxy also needs the param map
		prx = &StrcPrx{proxy: proxy{reg, nfo.Type, ptr}, params: nfo.params}
	case reflect.Map:
		if et.Key().Kind() != reflect.String {
			break
		}
		et, err := reg.Reflect(et.Elem())
		if err != nil {
			return nil, err
		}
		prx = &MapPrx{proxy: proxy{reg, typ.DictOf(et), ptr}}
	case reflect.Interface:
		prx = &AnyPrx{proxy{reg, typ.Any, ptr}, anyVal(ptr)}
	}
	if prx == nil {
		return nil, fmt.Errorf("cannot proxy type %s", ptr.Type())
	}
	pp, _ := prx.New()
	reg.setProxy(et, pp.(Prx))
	if opt {
		return &OptMut{prx, &org, null}, nil
	}
	return prx, nil
}

func toRef(ref reflect.Type, v reflect.Value) (reflect.Value, bool) {
	t := v.Type()
	if t == ref {
		return v, true
	}
	if t.ConvertibleTo(ref) {
		return v.Convert(ref), true
	}
	return v, false
}

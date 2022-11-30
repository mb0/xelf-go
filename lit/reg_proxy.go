package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/typ"
)

// ProxyValue returns a proxy value for the reflect value ptr or an error.
func (c *PrxReg) ProxyValue(ptr reflect.Value) (mut Mut, err error) {
	if ptr.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("requires pointer got %s", ptr.Type())
	}
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
	c.RLock()
	prx, ok := c.proxy[pt]
	c.RUnlock()
	if ok {
		return prx.NewWith(org), nil
	}
	switch et.Kind() {
	case reflect.Bool:
		if v, ok := toRef(ptrBoolMut, ptr, org); ok {
			return optPrx(v.Interface().(*BoolMut), org, opt, null)
		}
	case reflect.Int64:
		if et != typInt64 && et.Implements(ptrSecs.Elem()) {
			if v, ok := toRef(ptrSpanMut, ptr, org); ok {
				return optPrx(v.Interface().(*SpanMut), org, opt, null)
			}
		}
		if v, ok := toRef(ptrIntMut, ptr, noval); ok {
			return optPrx(v.Interface().(*IntMut), org, opt, null)
		}
		fallthrough
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Uint64, reflect.Uint32, reflect.Uint16:
		mut = &IntPrx{newProxy(c, typ.Int, org)}
	case reflect.Float64:
		if pt == ptrNumMut {
			return optPrx(ptr.Interface().(*NumMut), org, opt, null)
		}
		if v, ok := toRef(ptrRealMut, ptr, noval); ok {
			return optPrx(v.Interface().(*RealMut), org, opt, null)
		}
		fallthrough
	case reflect.Float32:
		mut = &RealPrx{newProxy(c, typ.Real, org)}
	case reflect.String:
		if pt == ptrCharMut {
			return optPrx(ptr.Interface().(*CharMut), org, opt, null)
		}
		if v, ok := toRef(ptrStrMut, ptr, org); ok {
			return optPrx(v.Interface().(*StrMut), org, opt, null)
		}
	case reflect.Slice:
		if et.Elem().Kind() == reflect.Uint8 {
			if v, ok := toRef(ptrRawMut, ptr, org); ok {
				return optPrx(v.Interface().(*RawMut), org, opt, null)
			}
		}
		et, err := c.Reflect(et.Elem())
		if err != nil {
			return nil, err
		}
		mut = &ListPrx{newProxy(c, typ.ListOf(et), org)}
	case reflect.Array:
		if v, ok := toRef(ptrUUIDMut, ptr, org); ok {
			return optPrx(v.Interface().(*UUIDMut), org, opt, null)
		}
	case reflect.Struct:
		if v, ok := toRef(ptrTimeMut, ptr, org); ok {
			return optPrx(v.Interface().(*TimeMut), org, opt, null)
		}
		if v, ok := toRef(ptrType, ptr, org); ok {
			return optPrx(v.Interface().(*typ.Type), org, opt, null)
		}
		nfo, err := c.ReflectStruct(et)
		if err != nil {
			return nil, err
		}
		// we use reflect struct because the proxy also needs the param map
		mut = &ObjPrx{proxy: newProxy(c, nfo.Type, org), params: nfo.params}
	case reflect.Map:
		if et.Key().Kind() != reflect.String {
			break
		}
		et, err := c.Reflect(et.Elem())
		if err != nil {
			return nil, err
		}
		mut = &MapPrx{proxy: newProxy(c, typ.DictOf(et), org)}
	case reflect.Interface:
		mut = &AnyPrx{proxy{c, typ.Any, org}, anyVal(org)}
	}
	if mut == nil {
		return nil, fmt.Errorf("cannot proxy type %s", ptr.Type())
	}
	pp := mut.New()
	if prx, ok := pp.(Prx); ok {
		rt := prx.Reflect().Type()
		if rt.Kind() != reflect.Ptr {
			pv := reflect.New(rt)
			pp = prx.NewWith(pv)
			prx, rt = pp.(Prx), pv.Type()
		}
		c.Lock()
		c.setProxy(rt, prx)
		c.Unlock()
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
				v = newPrimMut(org, v)
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
		ptr = newPrimMut(org, ptr)
		return &OptPrx{ptr.Interface().(Mut), org, true}, nil
	}
	return ptr.Interface().(Mut), nil
}

func newPrimMut(org, el reflect.Value) (ptr reflect.Value) {
	if el.IsNil() {
		el = reflect.New(org.Type().Elem().Elem())
		org.Elem().Set(el)
	}
	return el
}

var noval reflect.Value

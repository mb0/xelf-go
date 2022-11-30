package lit

import (
	"fmt"
	"reflect"
	"strings"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Reflect returns the xelf type for the reflect type or an error.
func (pr *PrxReg) Reflect(t reflect.Type) (typ.Type, error) {
	ptr := t.Kind() == reflect.Ptr
	if ptr {
		t = t.Elem()
	}
	res := reflectSimple(t)
	if res == typ.Void {
		pr.RLock()
		if info, ok := pr.param[t]; ok {
			res = info.Type
		}
		pr.RUnlock()
	}
	if res != typ.Void {
		if ptr {
			res = typ.Opt(res)
		}
		return res, nil
	}
	pr.Lock()
	defer pr.Unlock()
	return pr.reflectTypeRest(t, ptr, new(tstack))
}
func (pr *PrxReg) ReflectStruct(t reflect.Type) (nfo typInfo, err error) {
	pr.RLock()
	nfo, ok := pr.param[t]
	pr.RUnlock()
	if !ok {
		pr.Lock()
		nfo.Type, nfo.params, err = pr.reflectStruct(t, new(tstack))
		if err == nil {
			pr.setParam(t, nfo)
		}
		pr.Unlock()
	}
	return nfo, err
}
func reflectSimple(t reflect.Type) typ.Type {
	// first switch on the primitives that do not need convertible to type check
	// to avoid the extra map lookup
	switch t.Kind() {
	case reflect.Bool:
		return typ.Bool
	case reflect.Int64:
		if t != typInt64 {
			break
		}
		fallthrough
	case reflect.Int, reflect.Int32:
		return typ.Int
	case reflect.Uint64:
		fallthrough
	case reflect.Uint, reflect.Uint32:
		return typ.Int
	case reflect.Float32, reflect.Float64:
		return typ.Real
	case reflect.String:
		return typ.Str
	}
	return typ.Void
}
func (pr *PrxReg) reflectType(t reflect.Type, s *tstack) (res typ.Type, err error) {
	ptr := t.Kind() == reflect.Ptr
	if ptr {
		t = t.Elem()
	}
	res = reflectSimple(t)
	if res != typ.Void {
		return res, nil
	}
	if info, ok := pr.param[t]; ok {
		if ptr {
			return typ.Opt(info.Type), err
		}
		return info.Type, nil
	}
	return pr.reflectTypeRest(t, ptr, s)
}
func (pr *PrxReg) reflectTypeRest(t reflect.Type, ptr bool, s *tstack) (res typ.Type, err error) {
	// first switch on the primitives that do not need convertible to type check
	// to avoid the extra map lookup
	// now lets cache and lookup all other types that require more involved type checks
	var pm *params
	switch t.Kind() {
	case reflect.Int64:
		if t != typInt64 && isRef(t, ptrSecs.Elem()) {
			res = typ.Span
			break
		}
		res = typ.Int
	case reflect.Struct:
		if isRef(t, ptrTimeMut.Elem()) {
			res = typ.Time
			break
		}
		if isRef(t, ptrType.Elem()) {
			res = typ.Typ
			break
		}
		if isRef(t, ptrDict.Elem()) {
			res = typ.Dict
			break
		}
		if isRef(t, ptrList.Elem()) {
			res = typ.List
			break
		}
		res, pm, err = pr.reflectStruct(t, s)
		if err != nil {
			return res, err
		}
	case reflect.Array:
		if isRef(t, ptrUUIDMut.Elem()) {
			res = typ.UUID
		}
	case reflect.Map:
		if !isRef(t.Key(), ptrStrMut.Elem()) {
			return typ.Void, fmt.Errorf("map key can only be a string type")
		}
		et, err := pr.reflectType(t.Elem(), s)
		if err != nil {
			return res, err
		}
		res = typ.DictOf(et)
	case reflect.Slice:
		if isRef(t, ptrRawMut.Elem()) {
			res = typ.Raw
			break
		}
		et, err := pr.reflectType(t.Elem(), s)
		if err != nil {
			return typ.Void, err
		}
		res = typ.ListOf(et)
	case reflect.Interface:
		res, ptr = typ.Any, false
	}
	if res.Zero() {
		return res, fmt.Errorf("cannot reflect type of %s", t)
	}
	pr.setParam(t, typInfo{res, pm})
	if ptr {
		res = typ.Opt(res)
	}
	return res, nil
}
func (pr *PrxReg) reflectStruct(t reflect.Type, s *tstack) (typ.Type, *params, error) {
	ref := s.add(t)
	if ref != "" {
		return typ.Sel(ref), nil, nil
	}
	pm, err := pr.reflectFields(t, s)
	s.drop()
	if err != nil {
		return typ.Void, nil, err
	}
	tn := t.Name()
	if tn != "" {
		if cor.IsCased(tn) {
			tn = t.String()
		} else {
			tn = ""
		}
	}
	return typ.Type{Kind: knd.Obj, Ref: tn, Body: &typ.ParamBody{Params: pm.ps}}, &pm, nil
}
func (pr *PrxReg) reflectFields(t reflect.Type, s *tstack) (pm params, _ error) {
	n := t.NumField()
	pm.ps = make([]typ.Param, 0, n)
	pm.idx = make([][]int, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		err := pr.addField(&pm, f, s, nil)
		if err != nil {
			return pm, err
		}
	}
	return pm, nil
}

func (pr *PrxReg) addField(pm *params, f reflect.StructField, s *tstack, idx []int) error {
	jtag := f.Tag.Get("json")
	if jtag != "" && jtag[0] == '-' {
		return nil
	}
	if len(idx) > 0 {
		idx = idx[:len(idx):len(idx)]
	}
	idx = append(idx, f.Index...)
	if jtag == "" && f.Anonymous {
		ok, err := pr.addEmbed(pm, f.Type, s, idx)
		if err != nil || ok {
			return err
		}
	}
	ft, err := pr.reflectType(f.Type, s)
	if err != nil {
		return err
	}
	key := cor.Keyed(f.Name)
	if jtag != "" {
		if idx := strings.IndexByte(jtag, ','); idx >= 0 {
			key = jtag[:idx]
			if strings.Contains(jtag[idx:], ",omitempty") {
				key += "?"
			}
		} else {
			key = jtag
		}
		if key == "-" {
			return nil
		}
	} else if f.Anonymous {
		//key = "_" + key
	}
	pm.ps = append(pm.ps, typ.P(key, ft))
	pm.idx = append(pm.idx, idx)
	return nil
}

func (pr *PrxReg) addEmbed(pm *params, t reflect.Type, s *tstack, idx []int) (bool, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			err := pr.addField(pm, f, s, idx)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

type tstack struct {
	stack []reflect.Type
}

func (ts *tstack) add(t reflect.Type) string {
	for i := len(ts.stack) - 1; i >= 0; i-- {
		if ts.stack[i] != t {
			continue
		}
		var b strings.Builder
		for n := i; n < len(ts.stack); n++ {
			b.WriteByte('.')
		}
		return b.String()
	}
	ts.stack = append(ts.stack, t)
	return ""
}
func (ts *tstack) drop() {
	ts.stack = ts.stack[:len(ts.stack)-1]
}
func isRef(t reflect.Type, ref reflect.Type) bool {
	return t == ref || t.ConvertibleTo(ref)
}

var (
	ptrVal   = reflect.TypeOf((*Val)(nil))
	ptrMut   = reflect.TypeOf((*Mut)(nil))
	ptrPrx   = reflect.TypeOf((*Prx)(nil))
	ptrType  = reflect.TypeOf((*typ.Type)(nil))
	ptrList  = reflect.TypeOf((*List)(nil))
	ptrDict  = reflect.TypeOf((*Dict)(nil))
	ptrSecs  = reflect.TypeOf((*interface{ Seconds() float64 })(nil))
	typInt64 = reflect.TypeOf(int64(0))
)

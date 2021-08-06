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
func (reg *Reg) Reflect(t reflect.Type) (typ.Type, error) {
	return reg.reflectType(t, new(tstack))
}

func (reg *Reg) reflectType(t reflect.Type, s *tstack) (res typ.Type, err error) {
	ptr := t.Kind() == reflect.Ptr
	if ptr {
		t = t.Elem()
	}
	// first switch on the primitives that do not need convertible to type check
	// to avoid the extra map lookup
	switch t.Kind() {
	case reflect.Bool:
		return typ.Bool, nil
	case reflect.Int64:
		if t != typInt64 {
			break
		}
		fallthrough
	case reflect.Int, reflect.Int32:
		return typ.Int, nil
	case reflect.Uint64:
		fallthrough
	case reflect.Uint, reflect.Uint32:
		return typ.Int, nil
	case reflect.Float32, reflect.Float64:
		return typ.Real, nil
	case reflect.String:
		return typ.Str, nil
	}
	// now lets cache all other types that require more involved type checks
	if info, ok := reg.param[t]; ok {
		if ptr {
			return typ.Opt(info.Type), err
		}
		return info.Type, nil
	}
	var pm *params
	switch t.Kind() {
	case reflect.Int64:
		if t != typInt64 && isRef(t, ptrSecs.Elem()) {
			res = typ.Span
			break
		}
		res = typ.Int
	case reflect.Struct:
		if isRef(t, ptrTime.Elem()) {
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
		res, pm, err = reg.reflectStruct(t, s)
		if err != nil {
			return res, err
		}
	case reflect.Array:
		if isRef(t, ptrUUID.Elem()) {
			res = typ.UUID
		}
	case reflect.Map:
		if !isRef(t.Key(), ptrStr.Elem()) {
			return typ.Void, fmt.Errorf("map key can only be a string type")
		}
		et, err := reg.reflectType(t.Elem(), s)
		if err != nil {
			return res, err
		}
		res = typ.DictOf(et)
	case reflect.Slice:
		if isRef(t, ptrRaw.Elem()) {
			res = typ.Raw
			break
		}
		et, err := reg.reflectType(t.Elem(), s)
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
	reg.setParam(t, typInfo{res, pm})
	if ptr {
		res = typ.Opt(res)
	}
	return res, nil
}
func (reg *Reg) reflectStruct(t reflect.Type, s *tstack) (typ.Type, *params, error) {
	ref := s.add(t)
	if ref != "" {
		return typ.Sel(ref), nil, nil
	}
	pm, err := reg.reflectFields(t, s)
	s.drop()
	if err != nil {
		return typ.Void, nil, err
	}
	k := knd.Rec
	tn := t.Name()
	if tn != "" {
		if cor.IsCased(tn) {
			tn = t.String()
			k = knd.Obj
		} else {
			tn = ""
		}
	}
	return typ.Type{Kind: k, Body: &typ.ParamBody{Name: tn, Params: pm.ps}}, &pm, nil
}
func (reg *Reg) reflectFields(t reflect.Type, s *tstack) (pm params, _ error) {
	n := t.NumField()
	pm.ps = make([]typ.Param, 0, n)
	pm.idx = make([][]int, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		err := reg.addField(&pm, f, s, nil)
		if err != nil {
			return pm, err
		}
	}
	return pm, nil
}

func (reg *Reg) addField(pm *params, f reflect.StructField, s *tstack, idx []int) error {
	jtag := f.Tag.Get("json")
	if len(idx) > 0 {
		idx = idx[:len(idx):len(idx)]
	}
	idx = append(idx, f.Index...)
	if jtag == "" && f.Anonymous {
		ok, err := reg.addEmbed(pm, f.Type, s, idx)
		if err != nil || ok {
			return err
		}
	}
	ft, err := reg.reflectType(f.Type, s)
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

func (reg *Reg) addEmbed(pm *params, t reflect.Type, s *tstack, idx []int) (bool, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			err := reg.addField(pm, f, s, idx)
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

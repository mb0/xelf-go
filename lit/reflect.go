package lit

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

var DefaultCache = &Cache{}

type typInfo struct {
	typ.Type
	*params
}
type params struct {
	ps  []typ.Param
	idx [][]int
}

// Cache holds process-shared reflection information
type Cache struct {
	sync.RWMutex
	proxy map[reflect.Type]Prx
	param map[reflect.Type]typInfo
}

func (c *Cache) Param(rt reflect.Type) (typInfo, bool) {
	c.RLock()
	defer c.RUnlock()
	nfo, ok := c.param[rt]
	return nfo, ok
}
func (c *Cache) SetParam(rt reflect.Type, nfo typInfo) {
	c.Lock()
	c.setParam(rt, nfo)
	c.Unlock()
}
func (c *Cache) setParam(rt reflect.Type, nfo typInfo) {
	if c.param == nil {
		c.param = make(map[reflect.Type]typInfo)
	}
	c.param[rt] = nfo
}
func (c *Cache) Proxy(rt reflect.Type) (Prx, bool) {
	c.RLock()
	defer c.RUnlock()
	p, ok := c.proxy[rt]
	return p, ok
}
func (c *Cache) SetProxy(rt reflect.Type, prx Prx) {
	c.Lock()
	c.setProxy(rt, prx)
	c.Unlock()
}
func (c *Cache) setProxy(rt reflect.Type, prx Prx) {
	if c.proxy == nil {
		c.proxy = make(map[reflect.Type]Prx)
	}
	c.proxy[rt] = prx
}

// Reflect returns the xelf type for the reflect type or an error.
func (c *Cache) Reflect(t reflect.Type) (typ.Type, error) {
	c.Lock()
	defer c.Unlock()
	return c.reflectType(t, new(tstack))
}
func (c *Cache) ReflectStruct(t reflect.Type) (typ.Type, *params, error) {
	c.Lock()
	defer c.Unlock()
	return c.reflectStruct(t, new(tstack))
}

func (c *Cache) reflectType(t reflect.Type, s *tstack) (res typ.Type, err error) {
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
	if info, ok := c.param[t]; ok {
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
		res, pm, err = c.reflectStruct(t, s)
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
		et, err := c.reflectType(t.Elem(), s)
		if err != nil {
			return res, err
		}
		res = typ.DictOf(et)
	case reflect.Slice:
		if isRef(t, ptrRaw.Elem()) {
			res = typ.Raw
			break
		}
		et, err := c.reflectType(t.Elem(), s)
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
	c.setParam(t, typInfo{res, pm})
	if ptr {
		res = typ.Opt(res)
	}
	return res, nil
}
func (c *Cache) reflectStruct(t reflect.Type, s *tstack) (typ.Type, *params, error) {
	ref := s.add(t)
	if ref != "" {
		return typ.Sel(ref), nil, nil
	}
	pm, err := c.reflectFields(t, s)
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
func (c *Cache) reflectFields(t reflect.Type, s *tstack) (pm params, _ error) {
	n := t.NumField()
	pm.ps = make([]typ.Param, 0, n)
	pm.idx = make([][]int, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		err := c.addField(&pm, f, s, nil)
		if err != nil {
			return pm, err
		}
	}
	return pm, nil
}

func (c *Cache) addField(pm *params, f reflect.StructField, s *tstack, idx []int) error {
	jtag := f.Tag.Get("json")
	if jtag != "" && jtag[0] == '-' {
		return nil
	}
	if len(idx) > 0 {
		idx = idx[:len(idx):len(idx)]
	}
	idx = append(idx, f.Index...)
	if jtag == "" && f.Anonymous {
		ok, err := c.addEmbed(pm, f.Type, s, idx)
		if err != nil || ok {
			return err
		}
	}
	ft, err := c.reflectType(f.Type, s)
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

func (c *Cache) addEmbed(pm *params, t reflect.Type, s *tstack, idx []int) (bool, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			err := c.addField(pm, f, s, idx)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

// AddFrom updates the cache with entries from o.
func (c *Cache) AddFrom(o *Cache) {
	o.RLock()
	c.Lock()
	if len(o.proxy) > 0 {
		if c.proxy == nil {
			c.proxy = make(map[reflect.Type]Prx)
		}
		for rt, prx := range o.proxy {
			if _, ok := c.proxy[rt]; !ok {
				c.proxy[rt] = prx
			}
		}
	}
	if len(o.param) > 0 {
		if c.param == nil {
			c.param = make(map[reflect.Type]typInfo)
		}
		for rt, nfo := range o.param {
			if _, ok := c.param[rt]; !ok {
				c.param[rt] = nfo
			}
		}
	}
	c.Unlock()
	o.RUnlock()
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

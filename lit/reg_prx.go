package lit

import (
	"reflect"
	"sync"

	"xelf.org/xelf/typ"
)

type Reg interface {
	Reflect(rt reflect.Type) (typ.Type, error)
	ProxyValue(ptr reflect.Value) (Mut, error)
}

// Conv converts val to type t and returns a reflect value or an error.
func Conv(reg Reg, t reflect.Type, val Val) (reflect.Value, error) {
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

// Proxy returns a proxy value for ptr or an error.
func Proxy(reg Reg, ptr interface{}) (Mut, error) {
	return reg.ProxyValue(reflect.ValueOf(ptr))
}

// MustProxy returns a proxy value for ptr or panics.
func MustProxy(reg Reg, ptr interface{}) Mut {
	mut, err := Proxy(reg, ptr)
	if err != nil {
		panic(err)
	}
	return mut
}

// PrxReg holds process-shared reflection and proxy information
type PrxReg struct {
	sync.RWMutex
	proxy map[reflect.Type]Prx
	param map[reflect.Type]typInfo
}

// AddFrom updates the cache with entries from o.
func (pr *PrxReg) AddFrom(o *PrxReg) {
	if o == nil || pr == o {
		return
	}
	o.RLock()
	pr.Lock()
	if len(o.proxy) > 0 {
		if pr.proxy == nil {
			pr.proxy = make(map[reflect.Type]Prx)
		}
		for rt, prx := range o.proxy {
			if _, ok := pr.proxy[rt]; !ok {
				pr.proxy[rt] = prx
			}
		}
	}
	if len(o.param) > 0 {
		if pr.param == nil {
			pr.param = make(map[reflect.Type]typInfo)
		}
		for rt, nfo := range o.param {
			if _, ok := pr.param[rt]; !ok {
				pr.param[rt] = nfo
			}
		}
	}
	pr.Unlock()
	o.RUnlock()
}

type typInfo struct {
	typ.Type
	*params
}
type params struct {
	ps  []typ.Param
	idx [][]int
}

func (pr *PrxReg) setParam(rt reflect.Type, nfo typInfo) {
	if pr.param == nil {
		pr.param = make(map[reflect.Type]typInfo)
	}
	pr.param[rt] = nfo
}
func (pr *PrxReg) setProxy(rt reflect.Type, prx Prx) {
	if pr.proxy == nil {
		pr.proxy = make(map[reflect.Type]Prx)
	}
	pr.proxy[rt] = prx
}

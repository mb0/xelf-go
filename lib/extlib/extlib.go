package extlib

import (
	"xelf.org/xelf/ext"
	"xelf.org/xelf/lib"
)

type FuncMap map[string]interface{}

func MustLib(fms ...FuncMap) lib.Specs {
	s, err := Lib(fms...)
	if err != nil {
		panic(err)
	}
	return s
}

func Lib(fms ...FuncMap) (res lib.Specs, err error) {
	res = make(lib.Specs)
	for _, fm := range fms {
		for name, val := range fm {
			spec, err := ext.NewFunc(nil, name, val)
			if err != nil {
				return nil, err
			}
			res[name] = spec
		}
	}
	return res, nil
}

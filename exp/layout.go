package exp

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

func LayoutSpec(sig typ.Type, els []Exp) ([]Exp, error) {
	switch sig.Kind & knd.Spec {
	case knd.Form:
		return LayoutForm(sig, els)
	case knd.Func:
		return LayoutFunc(sig, els)
	}
	return nil, fmt.Errorf("layout unresolved sig")
}

func LayoutForm(sig typ.Type, els []Exp) ([]Exp, error) {
	ps := SigArgs(sig)
	res := make([]Exp, len(ps))
	for i, p := range ps {
		pt := p.Type
		opt := p.IsOpt() || pt.Kind&knd.None != 0
		var n int
		var arg Exp
		if pt.Kind&knd.Exp == knd.Tupl {
			var tn int
			pt, tn = typ.TuplEl(pt)
			if tn > 1 {
				n = consume(els, false)
				n -= n % tn
				fst := els[0].Source()
				lst := els[n-1].Source()
				arg = &Tupl{Res: p.Type, Els: els[:n], Src: ast.Src{
					Doc: fst.Doc, Pos: fst.Pos, End: lst.End,
				}}
			} else if pt.Kind == knd.Exp {
				arg = &Tupl{Res: typ.ElemTupl(pt), Els: els}
				n = len(els)
			} else {
				n = consume(els, pt.Kind == knd.Tag)
				arg = &Tupl{Res: typ.ElemTupl(pt), Els: els[:n]}
			}
		} else if len(els) > 0 {
			isTag := pt.Kind&knd.Exp == knd.Tag
			if el := els[0]; isTag == (el != nil && el.Type().Kind&knd.Tag != 0) {
				n, arg = 1, el
			}
		}
		if !opt && n == 0 {
			return nil, fmt.Errorf("missing argument %d %s", i, pt)
		}
		res[i] = arg
		els = els[n:]
	}
	if len(els) > 0 {
		return nil, fmt.Errorf("unexpected argument %s", els[0])
	}
	return res, nil
}
func LayoutFunc(sig typ.Type, els []Exp) ([]Exp, error) {
	ps := SigArgs(sig)
	if len(ps) == 0 {
		if len(els) > 0 {
			return nil, fmt.Errorf("unexpected argument for %s", sig)
		}
		return nil, nil
	}
	last := ps[len(ps)-1]
	variadic := last.Kind&knd.List != 0
	var tags bool
	res := make([]Exp, len(ps))
	var vs []Exp
NextEl:
	for i, el := range els {
		if tag, ok := el.(*Tag); ok {
			tags = true
			key := cor.Keyed(tag.Tag)
			for i, p := range ps {
				if p.Key != key {
					continue
				}
				if res[i] != nil {
					return nil, fmt.Errorf("duplicate argument %s", tag.Tag)
				}
				res[i] = tag.Exp
				continue NextEl
			}
			return nil, fmt.Errorf("unmatched tag argument %s", tag)
		} else if tags {
			return nil, fmt.Errorf("unexpected plain argument %s", el)
		} else if i >= len(ps)-1 && variadic {
			// if i< len(ps) check type else append
			vs = append(vs, el)
		} else if i < len(ps) {
			if res[i] != nil {
				return nil, fmt.Errorf("duplicate argument %d", i)
			}
			// TODO check type
			res[i] = el
		} else {
			return nil, fmt.Errorf("no parameter found for %s", el)
		}
	}
	if len(vs) > 0 {
		res[len(res)-1] = &Tupl{Els: vs}
	}
	for i, p := range ps {
		if res[i] == nil && !p.IsOpt() {
			return nil, fmt.Errorf("missing required argument %d %s", i, p.Name)
		}
	}
	return res, nil
}

func SigArgs(sig typ.Type) []typ.Param {
	pb, ok := sig.Body.(*typ.ParamBody)
	if !ok || len(pb.Params) < 2 {
		return nil
	}
	return pb.Params[:len(pb.Params)-1]
}
func SigRes(sig typ.Type) *typ.Param {
	pb, ok := sig.Body.(*typ.ParamBody)
	if !ok || len(pb.Params) < 1 {
		return &typ.Param{}
	}
	return &pb.Params[len(pb.Params)-1]
}

func consume(els []Exp, tags bool) int {
	for i, el := range els {
		if (el.Type().Kind == knd.Tag) != tags {
			return i
		}
	}
	return len(els)
}

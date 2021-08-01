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
		var vari bool
		pt := p.Type
		opt := p.IsOpt() || pt.Kind&knd.Exp != 0 && pt.Kind&knd.None != 0
		var n int
		var arg Exp
		if pt.Kind&knd.Exp == knd.Tupl {
			vari = true
			pb, ok := pt.Body.(*typ.ParamBody)
			if !ok || len(pb.Params) == 0 {
				pt = typ.Type{Kind: knd.Exp}
			} else {
				switch tn := len(pb.Params); tn {
				case 1:
					pt = pb.Params[0].Type
				default:
					pt = typ.Type{Kind: knd.Rec, Body: pb}
					n = consume(els, false)
					n -= n % tn
					tupl := &Tupl{Type: typ.TuplList(pt), Els: make([]Exp, 0, n/tn)}
					for i := 0; i < n; i += tn {
						fst := els[i].Source()
						lst := els[i+tn-1].Source()
						tupl.Els = append(tupl.Els, &Tupl{Type: pt, Els: els[i : i+tn],
							Src: ast.Src{Doc: fst.Doc, Pos: fst.Pos, End: lst.End},
						})
					}
					arg = tupl
				}
			}
		}
		if vari && arg == nil {
			if pt == typ.Exp {
				arg = &Tupl{Type: typ.TuplList(pt), Els: els}
				n = len(els)
			} else {
				n = consume(els, pt.Kind == knd.Tag)
				// TODO type dict|T for tags else list|T
				var at typ.Type
				if pt.Kind&knd.Tag != 0 {
					at = typ.DictOf(typ.ResEl(pt))
				} else {
					at = typ.ListOf(pt)
				}
				arg = &Tupl{Type: typ.TuplList(at), Els: els[:n]}
			}
		} else if len(els) > 0 {
			n, arg = 1, els[0]
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
			return nil, fmt.Errorf("unexpected argument for %s", sig.Type())
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
			return nil, fmt.Errorf("no paramter found for %s", el)
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

func SigName(sig typ.Type) string {
	pb, ok := sig.Body.(*typ.ParamBody)
	if !ok {
		return ""
	}
	return pb.Name
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
		if (el.Kind() == knd.Tag) != tags {
			return i
		}
	}
	return len(els)
}

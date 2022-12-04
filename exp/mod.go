package exp

import (
	"fmt"
	"strings"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// File is a simple representation of any xelf input that store information about modules.
type File struct {
	// URL is the resource locator for this input, it should be empty or conform to net/url.
	// It usually is a simple file path, but could point into a zip file served on the web.
	URL string

	// Refs contains all declared and used modules for this file.
	Refs ModRefs
}

func (f *File) AddRefs(refs ...ModRef) error {
	for _, fr := range f.Refs { // we more often add few refs, so invert to reduce worst case
		if key := fr.Key(); ModRefs(refs).Find(key) != nil {
			return fmt.Errorf("the module name %q is already in use", key)
		}
	}
	f.Refs = append(f.Refs, refs...)
	return nil
}

// Mod is a simple representation of a module.
type Mod struct {
	// File points to the original source file of this module.
	File *File

	// Name is the declare module name.
	Name string

	// Decl holds the exported module declarations, that are copied for each program.
	Decl *lit.Obj
}

// ModRef represent a module reference with a possible alias the original import path.
type ModRef struct {
	Alias string
	Path  string
	Pub   bool
	*Mod
}

// Key returns the alias or module name.
func (ref ModRef) Key() string {
	if ref.Alias != "" {
		return ref.Alias
	}
	return ref.Name
}

type ModRefs []ModRef

func (ms ModRefs) Find(k string) *ModRef {
	for i, m := range ms {
		if m.Key() == k {
			return &ms[i]
		}
	}
	return nil
}

func LookupMod(p *Prog, qual, rest string) (lit.Val, error) {
	m := p.File.Refs.Find(qual)
	if m == nil {
		return nil, ErrSymNotFound
	}
	val, err := lit.Select(m.Decl, rest)
	if err != nil {
		return nil, err
	}
	if m.File.URL == p.File.URL || m.Alias == "" || m.Alias == m.Name {
		return val, nil
	}
	return lit.EditTypes(val, func(e *typ.Editor) (typ.Type, error) {
		if e.Ref != "" {
			q, sel := SplitQualifier(e.Ref)
			if q == m.Name && m.Alias != "" { // cover the mod name itself for now
				e.Ref = m.Alias + sel
			}
		}
		return e.Type, nil
	})
}

func SplitQualifier(k string) (q, _ string) {
	dot := strings.IndexByte(k, '.')
	if dot > 0 {
		if q = k[:dot]; cor.IsKey(q) {
			return q, k[dot:]
		}
	}
	return "", k
}

package exp

import (
	"fmt"
	"strings"

	"xelf.org/xelf/lit"
)

// File is a simple representation of any xelf input that store information about modules.
type File struct {
	// URL is the resource locator for this input, it should be empty or conform to net/url.
	// It usually is a simple file path, but could point into a zip file served on the web.
	URL string

	// Uses contains all modules used by this file.
	Uses ModRefs

	// Decls contains modules exported and declared by this file.
	Decls ModRefs
}

// Mod is a simple representation of a module.
type Mod struct {
	// File points to the original source file of this module.
	File *File

	// Name is the declare module name.
	Name string

	// Decl holds the exported module declarations, that are copied for each program.
	Decl *Lit

	// Optional setup hook for platform support.
	Setup func(p *Prog, m *Mod) error
}

// ModRef represent a module reference with a possible alias the original import path.
type ModRef struct {
	Alias string
	Path  string
	*Mod
}

type ModRefs []ModRef

func (ms ModRefs) Lookup(p *Prog, k string) (*Lit, error) {
	if len(ms) == 0 {
		return nil, ErrSymNotFound
	}
	dot := strings.IndexByte(k, '.')
	if dot <= 0 {
		return nil, ErrSymNotFound
	}
	qual := k[:dot]
	m := ms.find(qual)
	if m.Mod == nil || m.Decl == nil {
		return nil, fmt.Errorf("module %s unresolved", m.Path)
	}
	got, ok := p.cache[m.Mod]
	if !ok {
		t, err := p.Sys.Inst(LookupType(p), m.Decl.Res)
		if err != nil {
			return nil, err
		}
		n, err := lit.Copy(m.Decl.Val)
		if err != nil {
			return nil, err
		}
		got = &Lit{Res: t, Val: n}
		if p.cache == nil {
			p.cache = make(map[*Mod]*Lit)
		}
		p.cache[m.Mod] = got
	}
	return Select(got, k[dot+1:])
}

func (ms ModRefs) find(key string) (ref ModRef) {
	for _, m := range ms {
		if m.Alias == key || m.Name == key {
			return m
		}
	}
	return
}

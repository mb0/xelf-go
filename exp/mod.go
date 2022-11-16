package exp

import (
	"fmt"
	"strings"

	"xelf.org/xelf/cor"
)

// File is a simple representation of any xelf input that store information about modules.
type File struct {
	// URL is the resource locator for this input, it should be empty or conform to net/url.
	// It usually is a simple file path, but could point into a zip file served on the web.
	URL string

	// Refs contains all declared and used modules for this file.
	Refs ModRefs
}

// Mod is a simple representation of a module.
type Mod struct {
	// File points to the original source file of this module.
	File *File

	// Name is the declare module name.
	Name string

	// Decl holds the exported module declarations, that are copied for each program.
	Decl *Lit
}

// ModRef represent a module reference with a possible alias the original import path.
type ModRef struct {
	Alias string
	Path  string
	Pub   bool
	*Mod
}

type ModRefs []ModRef

func (ms ModRefs) Lookup(k string) (*Lit, error) {
	if len(ms) == 0 {
		return nil, ErrSymNotFound
	}
	dot := strings.IndexByte(k, '.')
	if dot <= 0 {
		return nil, ErrSymNotFound
	}
	qual := k[:dot]
	if !cor.IsKey(qual) {
		return nil, ErrSymNotFound
	}
	m := ms.find(qual)
	if m.Mod == nil || m.Decl == nil {
		return nil, fmt.Errorf("module %s unresolved", m.Path)
	}
	return Select(m.Decl, k[dot+1:])
}

func (ms ModRefs) find(key string) (ref ModRef) {
	for _, m := range ms {
		if m.Alias == key || m.Name == key {
			return m
		}
	}
	return
}

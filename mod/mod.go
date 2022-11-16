package mod

import (
	"errors"
	"sync"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
)

var ErrFileNotFound = errors.New("mod file not found")

// Registry provides a central global module registry for convenience.
var Registry = new(SysMods)

type (
	File   = exp.File
	Mod    = exp.Mod
	ModRef = exp.ModRef
)

// Src is the raw and program independent module source for a location.
// The input is represented either as an abstract syntax tree or as program specific setup hook.
type Src struct {
	Rel string
	Loc
	Raw   []ast.Ast
	Setup func(*exp.Prog, *Src) (*File, error)
}

// Loader caches and loads module sources.
type Loader interface {
	LoadSrc(path, base *Loc) (*Src, error)
}

// SysMods is a thread-safe module registry that implements the module loader interface.
type SysMods struct {
	sync.RWMutex
	srcs map[string]*Src
}

func (sm *SysMods) Register(src *Src) *Src {
	sm.Lock()
	defer sm.Unlock()
	if sm.srcs == nil {
		sm.srcs = make(map[string]*Src)
	}
	sm.srcs[src.Rel] = src
	return src
}

func (sm *SysMods) LoadSrc(raw, base *Loc) (*Src, error) {
	if proto := raw.Proto(); proto != "" && proto != "xelf" {
		return nil, ErrFileNotFound
	}
	p := raw.Path()
	sm.RLock()
	defer sm.RUnlock()
	src := sm.srcs[p]
	if src == nil {
		return nil, ErrFileNotFound
	}
	return src, nil
}

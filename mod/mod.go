package mod

import (
	"errors"
	"sync"

	"xelf.org/xelf/exp"
)

var ErrFileNotFound = errors.New("mod file not found")

// Registry provides a central global module registry for convinience.
var Registry = new(SysMods)

type (
	File   = exp.File
	Mod    = exp.Mod
	ModRef = exp.ModRef
)

// Loader is a simple api to lookup and load modules.
type Loader interface {
	LoadFile(p *exp.Prog, path string) (*File, error)
}

// SysMods is a threadsafe module registry that implements the module loader interface.
type SysMods struct {
	sync.RWMutex
	files map[string]*File
}

func (sm *SysMods) Register(f *File) {
	sm.Lock()
	defer sm.Unlock()
	if sm.files == nil {
		sm.files = make(map[string]*File)
	}
	sm.files[f.URL] = f
}

func (sm *SysMods) LoadFile(prog *exp.Prog, raw string) (*File, error) {
	sm.RLock()
	defer sm.RUnlock()
	if f := sm.files[raw]; f != nil {
		return f, nil
	}
	return nil, ErrFileNotFound
}

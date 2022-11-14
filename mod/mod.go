package mod

import (
	"errors"
	"net/url"
	"sync"

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

// Loader is a simple api to lookup and load modules.
type Loader interface {
	LoadFile(*exp.Prog, *url.URL) (*File, error)
}

// SysMods is a thread-safe module registry that implements the module loader interface.
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

func (sm *SysMods) LoadFile(prog *exp.Prog, raw *url.URL) (*File, error) {
	if raw.Scheme != "" && raw.Scheme != "xelf:" {
		return nil, ErrFileNotFound
	}
	sm.RLock()
	defer sm.RUnlock()
	if f := sm.files[raw.Path]; f != nil {
		return f, nil
	}
	return nil, ErrFileNotFound
}

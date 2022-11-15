package mod

import (
	"errors"
	"fmt"
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

func (sm *SysMods) Register(f *File) error {
	u, err := url.Parse(f.URL)
	if err != nil {
		return err
	}
	if u.Scheme != "" && u.Scheme != "xelf" {
		return fmt.Errorf("incorrect scheme %s", u.Scheme)
	}
	sm.Lock()
	defer sm.Unlock()
	if sm.files == nil {
		sm.files = make(map[string]*File)
	}
	sm.files[rawPath(u)] = f
	return nil
}

func (sm *SysMods) LoadFile(prog *exp.Prog, raw *url.URL) (*File, error) {
	if raw.Scheme != "" && raw.Scheme != "xelf" {
		return nil, ErrFileNotFound
	}
	p := rawPath(raw)
	sm.RLock()
	defer sm.RUnlock()
	if f := sm.files[p]; f != nil {
		return f, nil
	}
	return nil, ErrFileNotFound
}

func rawPath(u *url.URL) string {
	if u.Opaque != "" {
		return u.Opaque
	}
	if u.RawPath != "" {
		return u.RawPath
	}
	return u.Path
}

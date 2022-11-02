package mod

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/typ"
)

func FileMods(roots ...string) *FSMods {
	fm := &FSMods{
		Roots:  make([]PathFS, 0, len(roots)),
		Tries:  DefaultTries,
		cache:  make(map[string]*File),
		lookup: make(map[string]*File),
	}
	for _, root := range roots {
		fm.Roots = append(fm.Roots, PathFS{Path: root, FS: os.DirFS(root)})
	}
	return fm
}

type PathFS struct {
	Path string
	FS   fs.FS
}

type FSMods struct {
	Roots []PathFS
	Tries func(string) (string, []string)

	lookup map[string]*File // by use path
	cache  map[string]*File // by abs file path
}

func (fm *FSMods) LoadFile(prog *exp.Prog, raw string) (*File, error) {
	var rel string
	key, sym, roots := raw, raw, fm.Roots
	if strings.HasPrefix(raw, "./") {
		rel = path.Dir(prog.File.URL)
		if rel == "" {
			return nil, fmt.Errorf("relative mod path not allowed here")
		}
		sym = raw[2:]
		key = path.Join(rel, raw)
		roots = []PathFS{{rel, os.DirFS(rel)}}
	}
	if f := fm.lookup[key]; f != nil {
		return f, nil
	}
	_, tries := fm.Tries(sym)
	// try cache
	for _, r := range roots {
		for _, try := range tries {
			tp := path.Join(r.Path, try)
			if f := fm.cache[tp]; f != nil {
				return f, nil
			}
		}
	}
	var root PathFS
	var found string
Outer:
	for _, r := range roots {
		// try each root
		for _, try := range tries {
			_, err := fs.Stat(r.FS, try)
			if err != nil {
				continue
			}
			// found file
			root = r
			found = try
			break Outer
		}
	}
	if root.FS == nil {
		return nil, ErrFileNotFound
	}
	b, err := fs.ReadFile(root.FS, found)
	if err != nil {
		return nil, err
	}
	abs := path.Join(root.Path, found)
	f, err := fm.readFile(prog, b, abs)
	if err != nil {
		return nil, err
	}
	fm.cache[abs] = f
	fm.lookup[key] = f
	return f, nil
}
func (fm *FSMods) readFile(prog *exp.Prog, f []byte, url string) (*File, error) {
	e, err := exp.Read(prog.Reg, bytes.NewReader(f), url)
	if err != nil {
		return nil, err
	}
	// shallow copy the loader for every loaded file
	p := *prog
	p.File = File{URL: url}
	e, err = p.Resl(&p, e, typ.Void)
	if err != nil {
		return nil, err
	}
	_, err = p.Eval(&p, e)
	return &p.File, err
}

func DefaultTries(sym string) (name string, res []string) {
	var dir, fn string
	// inspect path
	idx := strings.LastIndexByte(sym, '/')
	if idx < 0 {
		fn = sym
	} else {
		dir, fn = sym[:idx], sym[idx+1:]
	}
	if strings.HasSuffix(sym, ".xelf") {
		return fn[:len(fn)-5], []string{sym}
	}
	name = fn
	fn += ".xelf"
	if dir != "" {
		return name, []string{path.Join(dir, fn)}
	}
	return name, []string{fn, path.Join(name, fn), path.Join(name, "mod.xelf")}
}

package mod

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"strings"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/typ"
)

func FileMods(roots ...string) *FSMods {
	fm := &FSMods{
		Roots: make([]*PathFS, 0, len(roots)),
		Ext:   []string{".xelf"},
		Index: []string{"mod.xelf"},
	}
	for _, root := range roots {
		fm.Roots = append(fm.Roots, &PathFS{Path: root, FS: os.DirFS(root)})
	}
	return fm
}

type PathFS struct {
	Path  string
	FS    fs.FS
	Rel   string
	cache map[string]*File
}

type FSMods struct {
	Roots []*PathFS
	Ext   []string
	Index []string

	log   func(root, path string)
	local map[string]*PathFS
}

func (fm *FSMods) LoadFile(prog *exp.Prog, raw *url.URL) (*File, error) {
	if raw.Scheme != "" && raw.Scheme != "file" {
		return nil, ErrFileNotFound
	}
	sym, roots := rawPath(raw), fm.Roots
	if strings.HasPrefix(sym, "./") {
		sym = sym[2:]
		r, err := fm.relRoot(prog, sym)
		if err != nil {
			return nil, err
		}
		return fm.try(prog, r, sym)
	}
	for _, r := range roots {
		f, err := fm.try(prog, r, sym)
		if err != nil {
			if err == ErrFileNotFound {
				continue
			}
		}
		return f, err
	}
	return nil, ErrFileNotFound
}
func (fm *FSMods) relRoot(prog *exp.Prog, sym string) (*PathFS, error) {
	u, err := url.Parse(prog.File.URL)
	if err != nil {
		return nil, err
	}
	rel := path.Dir(rawPath(u))
	if rel == "" {
		return nil, fmt.Errorf("relative mod path not allowed here")
	}
	if fm.local == nil {
		fm.local = make(map[string]*PathFS)
	} else if r := fm.local[rel]; r != nil {
		return r, nil
	}
	for _, r := range fm.Roots {
		rp := path.Clean(r.Path)
		if rel == rp {
			return r, nil
		}
		if strings.HasPrefix(rel, rp) {
			if r.cache == nil {
				r.cache = make(map[string]*File)
			}
			tmp := *r
			tmp.Rel = rel[len(rp)+1:]
			fm.local[rel] = &tmp
			return &tmp, nil
		}
	}
	r := &PathFS{Path: rel, FS: os.DirFS(rel)}
	fm.local[rel] = r
	return r, nil
}
func (fm *FSMods) readFile(prog *exp.Prog, f []byte, url string) (*File, error) {
	e, err := exp.Read(bytes.NewReader(f), url)
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
func (fm *FSMods) try(prog *exp.Prog, r *PathFS, sym string) (f *File, err error) {
	p := sym
	if r.Rel != "" {
		p = path.Join(r.Rel, sym)
	}
	if r.cache == nil {
		r.cache = make(map[string]*File)
	} else if f, ok := r.cache[p]; ok {
		if f == nil {
			return nil, ErrFileNotFound
		}
		return f, nil
	}
	// always try sym as is
	fi, err := fs.Stat(r.FS, p)
	var found string
	if err != nil {
		for _, ext := range fm.Ext {
			pp := p + ext
			_, err := fs.Stat(r.FS, pp)
			if err == nil {
				found = pp
				break
			}
		}
	} else if fi.IsDir() {
		for _, name := range fm.Index {
			pp := path.Join(p, name)
			_, err := fs.Stat(r.FS, pp)
			if err == nil {
				found = pp
				break
			}
		}
		for _, ext := range fm.Ext {
			pp := path.Join(p, fi.Name()+ext)
			_, err := fs.Stat(r.FS, pp)
			if err == nil {
				found = pp
				break
			}
		}
	} else {
		found = p
	}
	if found == "" {
		r.cache[p] = nil
		return nil, ErrFileNotFound
	}
	if p != found {
		f := r.cache[found]
		if f != nil {
			return f, nil
		}
	}
	if fm.log != nil {
		fm.log(r.Path, found)
	}
	b, err := fs.ReadFile(r.FS, found)
	if err != nil {
		return nil, err
	}
	f, err = fm.readFile(prog, b, path.Join(r.Path, found))
	r.cache[p] = f
	if p != found {
		r.cache[found] = f
	}
	return f, err
}

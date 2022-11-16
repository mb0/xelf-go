package mod

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"xelf.org/xelf/ast"
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
	cache map[string]*Src
}

type FSMods struct {
	Roots []*PathFS
	Ext   []string
	Index []string

	log   func(root, path string)
	local map[string]*PathFS
}

func (fm *FSMods) LoadSrc(raw, base *Loc) (*Src, error) {
	if proto := raw.Proto(); proto != "" && proto != "file" {
		return nil, ErrFileNotFound
	}
	p, roots := raw.Path(), fm.Roots
	if strings.HasPrefix(p, "./") {
		p = p[2:]
		r, err := fm.relRoot(p, base)
		if err != nil {
			return nil, err
		}
		return fm.try(r, p)
	}
	for _, r := range roots {
		src, err := fm.try(r, p)
		if err != nil {
			if err == ErrFileNotFound {
				continue
			}
		}
		return src, err
	}
	return nil, ErrFileNotFound
}
func (fm *FSMods) relRoot(p string, base *Loc) (*PathFS, error) {
	bp := base.Path()
	if pr := base.Proto(); base == nil || pr != "" && pr != "file" {
		return nil, fmt.Errorf("relative mod path not allowed here")
	}
	rel := path.Dir(bp)
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
				r.cache = make(map[string]*Src)
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

func (fm *FSMods) try(r *PathFS, part string) (*Src, error) {
	p := part
	if r.Rel != "" {
		p = path.Join(r.Rel, part)
	}
	if r.cache == nil {
		r.cache = make(map[string]*Src)
	} else if s, ok := r.cache[p]; ok {
		if s == nil {
			return nil, ErrFileNotFound
		}
		return s, nil
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
		s := r.cache[found]
		if s != nil {
			return s, nil
		}
	}
	if fm.log != nil {
		fm.log(r.Path, found)
	}
	ff, err := r.FS.Open(found)
	if err != nil {
		return nil, err
	}
	defer ff.Close()
	full := path.Join(r.Path, found)
	as, err := ast.ReadAll(ff, full)
	if err != nil {
		return nil, err
	}
	src := &Src{Rel: found, Loc: Loc{URL: "file:" + full}, Raw: as}
	r.cache[p] = src
	if p != found {
		r.cache[found] = src
	}
	return src, nil
}

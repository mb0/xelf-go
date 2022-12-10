// Package xps provides helper and conventions for working with the go plugin system
package xps

import (
	"bytes"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"xelf.org/xelf/lit"
)

// EnvRoots returns root paths from the $XELF_PLUGINS environment variable.
func EnvRoots() []string { return filepath.SplitList(os.Getenv("XELF_PLUGINS")) }

// FindAll walks roots and returns all plugin manifests that were found.
// It skips node_modules and testdata directories, and hidden files starting with a dot.
func FindAll(roots []string) (res []Manifest, err error) {
	for _, root := range roots {
		res, err = findAll(root, res)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

// Manifest provides the plugin path, name and a list of provided module paths.
type Manifest struct {
	Path string `json:"-"`
	Name string `json:"name"`
	// Caps holds all plugin capabilities, most prominently mods and cmds.
	// mods should be a list of provided modules
	// cmds a dict with command keys and help values
	Caps lit.Keyed `json:"caps,omitempty"`
}

// Read reads and returns the plugin manifest for path or an error.
func Read(path string) (m Manifest, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return m, err
	}
	err = lit.ReadInto(bytes.NewReader(b), path, lit.MustProxy(xpsReg, &m))
	m.Path = path
	return m, err
}

func (m Manifest) String() string { return m.Path }
func (m Manifest) PlugPath() string {
	if n := len(m.Path); n > 8 && m.Path[n-8:] == ".so.xelf" {
		return m.Path[:n-5]
	}
	return ""
}
func (m Manifest) Cmds() lit.Keyed {
	v, _ := m.Caps.Key("cmds")
	if cmds, ok := v.(*lit.Keyed); ok {
		return *cmds
	}
	return nil
}
func (m Manifest) Mods() []string {
	v, _ := m.Caps.Key("mods")
	if vs, ok := v.(*lit.Vals); ok {
		res := make([]string, 0, len(*vs))
		for _, el := range *vs {
			res = append(res, el.String())
		}
		return res
	}
	return nil
}

var xpsReg = lit.NewRegs()

func findAll(root string, res []Manifest) ([]Manifest, error) {
	dir := os.DirFS(root)
	err := fs.WalkDir(dir, ".", func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		n := e.Name()
		if e.IsDir() {
			if (n[0] == '.' && n != ".") || n == "node_modules" || n == "testdata" {
				return fs.SkipDir
			}
			return nil
		}
		if len(n) > 8 && n[0] != '.' && n[len(n)-8:] == ".so.xelf" {
			m, err := Read(filepath.Join(root, path))
			if err != nil {
				return err
			}
			res = append(res, m)
		}
		return nil
	})
	return res, err
}

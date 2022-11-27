package xps

import (
	"bytes"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"xelf.org/xelf/lit"
)

var xpsReg = lit.NewRegs()

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

// FindAll walks roots and returns all plugin manifests that were found.
// It skips node_modules and testdata directories, and hidden files starting with a dot.
func FindAll(roots []string) (res []Manifest) {
	for _, root := range roots {
		res = findAll(root, res)
	}
	return res
}

func findAll(root string, res []Manifest) []Manifest {
	dir := os.DirFS(root)
	fs.WalkDir(dir, ".", func(path string, e fs.DirEntry, err error) error {
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
	return res
}

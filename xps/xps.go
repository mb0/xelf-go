// Package xps provides helper and conventions for working with the go plugin system
package xps

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func EnvRoots() []string { return filepath.SplitList(os.Getenv("XELF_PLUGINS")) }

// Plugin wraps a go plugin with path and name information.
type Plug struct {
	*plugin.Plugin
	Path string
	Name string
}

func (p Plug) String() string { return p.Path }

func FindAll(roots []string) (res []string) {
	for _, root := range roots {
		res = findAll(root, res)
	}
	return res
}

func LoadAll(roots []string) (res []Plug, err error) {
	all := FindAll(roots)
	for _, found := range all {
		p := Plug{Path: found}
		p.Plugin, err = plugin.Open(found)
		if err != nil {
			return res, err
		}
		res = append(res, p)
	}
	return res, nil
}

func Load(roots []string, name string) (p Plug, err error) {
	for _, root := range roots {
		if p.Path = find(root, name); p.Path != "" {
			break
		}
	}
	if p.Path == "" {
		return p, fmt.Errorf("plugin %s not found", name)
	}
	p.Name = name
	p.Plugin, err = plugin.Open(p.Path)
	return p, err
}

func find(root, name string) string {
	if !isdir(root) {
		return ""
	}
	dir := filepath.Join(root, name)
	if isdir(dir) {
		if p := dir + "/xps/xps.so"; isfile(p) {
			return p
		}
		if p := dir + "/xps/" + name + "-xps.so"; isfile(p) {
			return p
		}
	}
	if p := dir + ".so"; isfile(p) {
		return p
	}
	if p := dir + "-xps.so"; isfile(p) {
		return p
	}
	return ""
}

func findAll(root string, res []string) []string {
	if !isdir(root) {
		return res
	}
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
		if len(n) > 3 && n[0] != '.' && n[len(n)-3:] == ".so" {
			if strings.Contains(path, "xps") {
				res = append(res, filepath.Join(root, path))
			}
		}
		return nil
	})
	return res
}

func isdir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
func isfile(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

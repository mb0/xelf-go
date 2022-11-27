// Package xps provides helper and conventions for working with the go plugin system
package xps

import (
	"os"
	"path/filepath"
	"plugin"
)

// EnvRoots returns root paths from the $XELF_PLUGINS environment variable.
func EnvRoots() []string { return filepath.SplitList(os.Getenv("XELF_PLUGINS")) }

// Manifest provides the plugin path, name and a list of provided module paths.
type Manifest struct {
	Path string   `json:"-"`
	Name string   `json:"name"`
	Mods []string `json:"mods,omitempty"`
}

func (m Manifest) String() string { return m.Path }
func (m Manifest) PlugPath() string {
	if n := len(m.Path); n > 8 && m.Path[n-8:] == ".so.xelf" {
		return m.Path[:n-5]
	}
	return ""
}

// Plug wraps a go plugin with a manifest. The manifest must use the '.so.xelf' file extension,
// thereby encoding the expected location of the plugin binary.
type Plug struct {
	Manifest
	*plugin.Plugin
}

// Load returns the plugin at the plugin manifest path p or an error.
func Load(path string) (*Plug, error) {
	m, err := Read(path)
	if err != nil {
		return nil, err
	}
	return loadPlug(m)
}

// LoadAll finds, opens and returns all plugins in roots or an error.
func LoadAll(roots []string) ([]*Plug, error) {
	all := FindAll(roots)
	res := make([]*Plug, 0, len(all))
	for _, m := range all {
		p, err := loadPlug(m)
		if err != nil {
			return res, err
		}
		res = append(res, p)
	}
	return res, nil
}

func loadPlug(m Manifest) (p *Plug, err error) {
	p = &Plug{Manifest: m}
	p.Plugin, err = plugin.Open(m.PlugPath())
	return p, err
}

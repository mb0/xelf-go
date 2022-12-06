// Package xps provides helper and conventions for working with the go plugin system
package xps

import (
	"os"
	"path/filepath"
	"plugin"

	"xelf.org/xelf/lit"
)

// EnvRoots returns root paths from the $XELF_PLUGINS environment variable.
func EnvRoots() []string { return filepath.SplitList(os.Getenv("XELF_PLUGINS")) }

// Cmd is the type signature and name we check for plugin subcommands.
// The dir is the assumed working dir and args start with the plugin name itself.
type Cmd = func(dir string, args []string) error

// Manifest provides the plugin path, name and a list of provided module paths.
type Manifest struct {
	Path string   `json:"-"`
	Name string   `json:"name"`
	Mods []string `json:"mods,omitempty"`
	// Cmds holds the subcommand names and descriptions or an empty key and group title.
	Cmds lit.Keyed `json:"cmds,omitempty"`
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
	Cmd Cmd
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

func loadPlug(m Manifest) (*Plug, error) {
	p, err := plugin.Open(m.PlugPath())
	if err != nil {
		return nil, err
	}
	var cmd Cmd
	if len(m.Cmds) > 0 {
		sym, err := p.Lookup("Cmd")
		if err != nil {
			return nil, err
		}
		cmd = sym.(Cmd)
	}
	return &Plug{Manifest: m, Plugin: p, Cmd: cmd}, nil
}

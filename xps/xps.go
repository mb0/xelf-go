// Package xps provides helper and conventions for working with the go plugin system
package xps

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/mod"
)

// EnvRoots returns root paths from the $XELF_PLUGINS environment variable.
func EnvRoots() []string { return filepath.SplitList(os.Getenv("XELF_PLUGINS")) }

// Cmd is the type signature and name we check for plugin subcommands.
// The dir is the assumed working dir and args start with the plugin name itself.
type Cmd = func(*CmdCtx) error

type CmdCtx struct {
	Dir  string
	Args []string
	Mani []Manifest
	Mods *Mods
	Fmod *mod.FSMods
	Wrap func(*CmdCtx, exp.Env) exp.Env
	Prog func(*CmdCtx) *exp.Prog
}

func (c *CmdCtx) Split() string {
	if a := c.Args; len(a) > 0 {
		c.Args = a[1:]
		return a[0]
	}
	return ""
}

func (c *CmdCtx) Manifests() []Manifest {
	if c.Mani == nil {
		c.Mani = FindAll(EnvRoots())
	}
	return c.Mani
}

// CmdRedir is capability extension for plugin commands wrapped as an error.
// Plugin commands can change the default program environment to be reused by a xelf commands.
type CmdRedir struct{ Cmd string }

func (d *CmdRedir) Error() string { return fmt.Sprintf("redirect to %s", d.Cmd) }

// Manifest provides the plugin path, name and a list of provided module paths.
type Manifest struct {
	Path string `json:"-"`
	Name string `json:"name"`
	// Caps holds all plugin capabilities, most prominently mods and cmds.
	// mods should be a list of provided modules
	// cmds a dict with command keys and help values
	Caps lit.Keyed `json:"caps,omitempty"`
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
	if len(m.Cmds()) > 0 {
		sym, err := p.Lookup("Cmd")
		if err != nil {
			return nil, err
		}
		cmd = sym.(Cmd)
	}
	return &Plug{Manifest: m, Plugin: p, Cmd: cmd}, nil
}

// PlugCmd loads and returns a subcommand for plug or nil.
// It only returns an error if a plugin was found but could not load a command and nil otherwise.
func PlugCmd(ctx *CmdCtx, plug string) (Cmd, error) {
	for _, m := range ctx.Manifests() {
		if m.Name != plug {
			continue
		}
		p, err := loadPlug(m)
		if err != nil {
			return nil, err
		}
		if p.Cmd == nil {
			return nil, fmt.Errorf("plugin %q has no command", plug)
		}
		return p.Cmd, nil
	}
	return nil, nil
}

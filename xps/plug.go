package xps

import (
	"fmt"
	"plugin"

	"xelf.org/xelf/mod"
)

// Plug wraps a go plugin with a manifest. The manifest must use the '.so.xelf' file extension,
// thereby encoding the expected location of the plugin binary.
type Plug struct {
	Manifest
	*plugin.Plugin
	Cmd Cmd
}

// Plugs caches and lazy-loads plugins.
type Plugs struct {
	Mani []Manifest
	All  map[string]*Plug
}

func (ps *Plugs) Init(ms []Manifest) {
	if ps.All == nil {
		ps.All = make(map[string]*Plug, len(ms))
	}
	for _, m := range ms {
		if f := ps.All[m.Path]; f != nil {
			continue // we already have that path
		}
		ps.Mani = append(ps.Mani, m)
		ps.All[m.Path] = &Plug{Manifest: m}
	}
}

func (ps *Plugs) LoadPlug(path string) (p *Plug, err error) {
	p = ps.All[path]
	return p, p.ensure()
}

// LoadCmd loads and returns a subcommand for the plugin name or nil.
// It only returns an error if a plugin was found but could not load a command and nil otherwise.
func (ps *Plugs) LoadCmd(name string) (Cmd, error) {
	for _, p := range ps.All {
		if p.Name != name {
			continue
		}
		if err := p.ensure(); err != nil {
			return nil, err
		}
		if p.Cmd == nil {
			return nil, fmt.Errorf("plugin %q has no command", name)
		}
		return p.Cmd, nil
	}
	return nil, nil
}

func (p *Plug) Load() (err error) {
	p.Plugin, err = plugin.Open(p.PlugPath())
	if err != nil {
		return err
	}
	if len(p.Cmds()) > 0 {
		sym, err := p.Lookup("Cmd")
		if err != nil {
			return err
		}
		p.Cmd = sym.(Cmd)
	}
	return nil
}

func (p *Plug) ensure() error {
	if p == nil {
		return fmt.Errorf("plugin not found")
	}
	if p.Plugin == nil {
		return p.Load()
	}
	return nil
}

// ModLoader wrapps a SysMods module source registry with a plugin list.
// It lazy-loads plugins that provide module source missing from the registry.
type ModLoader struct {
	Sys *mod.SysMods
	*Plugs
}

func (l *ModLoader) LoadSrc(raw, base *mod.Loc) (*mod.Src, error) {
	if proto := raw.Proto(); proto != "" && proto != "xelf" {
		return nil, mod.ErrFileNotFound
	}
	src, err := l.Sys.LoadSrc(raw, base)
	if err != nil {
		p := modPlug(l.All, raw.Path())
		if p != nil && p.Plugin == nil {
			if err = p.Load(); err != nil {
				return nil, err
			}
			src, err = l.Sys.LoadSrc(raw, base)
		}
	}
	return src, err
}

func modPlug(all map[string]*Plug, mpath string) *Plug {
	for _, p := range all {
		for _, path := range p.Mods() {
			if mpath == path {
				return p
			}
		}
	}
	return nil
}

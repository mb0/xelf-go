package xps

import (
	"fmt"
	"sort"

	"xelf.org/xelf/mod"
)

// Mods wrapps a SysMods module source registry with a list of plugin manifests.
// It lazy-loads a plugin that provides a module source missing from the registry.
type Mods struct {
	Sys   *mod.SysMods
	All   []Manifest
	Plugs []*Plug
}

func NewMods(sys *mod.SysMods, all []Manifest) *Mods {
	return &Mods{Sys: sys, All: all}
}

func (sm *Mods) LoadSrc(raw, base *mod.Loc) (*mod.Src, error) {
	if proto := raw.Proto(); proto != "" && proto != "xelf" {
		return nil, mod.ErrFileNotFound
	}
	src, err := sm.Sys.LoadSrc(raw, base)
	if err != nil {
		ms := manifestsFor(sm.All, raw.Path())
		if len(ms) > 0 {
			// sorted by mods count desc, we use the match that covers the most plugins
			_, err = sm.load(ms[0])
			if err != nil {
				return nil, err
			}
			src, err = sm.Sys.LoadSrc(raw, base)
		}
	}
	return src, err
}

func (sm *Mods) RunCmd(dir string, args []string) error {
	plug := args[0]
	for _, m := range sm.All {
		if m.Name != plug {
			continue
		}
		p, err := sm.load(m)
		if err != nil {
			return err
		}
		if p.Cmd == nil {
			return fmt.Errorf("plugin %q has no command", plug)
		}
		return p.Cmd(dir, args)
	}
	return fmt.Errorf("subcommand %q not found", plug)
}
func (sm *Mods) load(m Manifest) (*Plug, error) {
	for _, p := range sm.Plugs {
		if p.Path == m.Path {
			return p, nil
		}
	}
	p, err := loadPlug(m)
	if err != nil {
		return nil, err
	}
	sm.Plugs = append(sm.Plugs, p)
	return p, nil
}

func manifestsFor(all []Manifest, path string) (res []Manifest) {
	for _, m := range all {
		for _, mod := range m.Mods {
			if path == mod {
				res = append(res, m)
			}
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return len(res[i].Mods) > len(res[j].Mods)
	})
	return res
}

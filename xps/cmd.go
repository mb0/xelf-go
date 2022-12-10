package xps

import (
	"fmt"

	"xelf.org/xelf/exp"
)

// Cmd is the type signature and name we check for plugin subcommands.
// The dir is the assumed working dir and args start with the plugin name itself.
type Cmd = func(ctx *CmdCtx) error

// CmdRedir is capability extension for plugin commands wrapped as an error.
// Plugin commands can change the default program environment to be reused by a xelf commands.
type CmdRedir struct{ Cmd string }

func (d *CmdRedir) Error() string { return fmt.Sprintf("redirect to %s", d.Cmd) }

type CmdCtx struct {
	Plugs
	Dir  string
	Args []string

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
		mani, _ := FindAll(EnvRoots())
		c.Plugs.Init(mani)
	}
	return c.Mani
}

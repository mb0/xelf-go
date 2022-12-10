Xelf command
============

We want a basic xelf command to provide some help working with xelf files.

Problem
-------

We need to decide the scope of the xelf command. Do we only provide basic helpers or do we use it
as default command to run all xelf code? Should it be part of this module or a new one?

Most helpers are generic enough, but everything involving evaluation is problematic. Xelf code often
relies on specific runtimes (e.g. daql, layla) to evaluate. We can provide only simple eval which
makes the command less useful or must find a way to integrate other runtimes.

Implementation
--------------

We created the new xelf.org/cmd module that provides generic xelf command related helpers and the
actual command package at xelf.org/cmd/xelf.

We use the new xps package to load external runtime support from $XELF_PLUGINS into the process.
Plugin manifest files store plugin capabilities and enable us to defer loading of plugins until we
actually need them. This mechanism currently covers subcommand extensions, module loaders and query
backend providers and works really great.

The command provides a collection of helpers organised as subcommands. Most subcommands read from
stdin and print to stdout, so that we can easily compose commands:

	echo '{a:1 b:[2 3 4]}' | xelf mut '{a:7 b+:[[5]]}' | xelf json

We want to provide following subcommands:
 * evaluation commands:  'run', 'test' and 'repl'
 * development commands: 'fmt', 'fix' and 'list'
 * literal commands:     'sel', 'mut' and 'json'
 * other commands:       'version' and 'help'

Discussion
----------

We already have separate daql and layla commands that provide some eval support. And we also have a
module system to expand program capabilities.

 * We could simply use a `$XELF_EVAL` environment variable to configure which command evaluates
   code. That however does not allow us to mix runtimes.
 * We could use a new xelf-cmd project that imports all runtimes and supported backends. This would
   not allow others to use xelf in the same way.
 * Could we somehow provide runtime integration using an external process, like rpc plugin modules?
   This would be nice and would also help provide external non-go code generators.
 * We could compile a new runner every time, that imports a list of go runtime packages. This would
   involve creating a temporary module, fetching all dependencies and compiling it. This was easier
   to implement with GOPATH. And then we are still limited to runtimes written in go.
 * We decided to use go plugins, they fit our requirements well and only require building a special
   go plugin binary, but otherwise has none of the problems, work or overhead associated with the
   other options. We would need an option to easily build and locate plugins. Go plugins must be
   build with the same go runtime version and results in reasonably large plugin files.

Using the plugin system we can provide the same features as the daql repl.

We want to make it easy to use the xelf command during development, however plug-ins must be rebuilt
whenever its code or a dependency was changes. For now we provide a simple rebuild subcommand that
looks for plugin.go files in folders along xps manifest files and runs the go tool. Instead of doing
that manually whenever we get an error, we could use the runtime/debug and debug/buildinfo go
standard packages to read the module versions of host and plug-in and compare them for mismatches
and rebuild if necessary. We should however measure the impact, to see whether we want to enable the
check permanently or toggle it with a dev flag.

We want some way to document specs. And a doc subcommand to discover that documentation.

Links
-----

 * https://pkg.go.dev/plugin go plugin package docs
 * https://appliedgo.net/plugins/ gives a good overview over external plugins in go
 * https://github.com/hashicorp/go-plugin is an implementation using net/rpc or grpc

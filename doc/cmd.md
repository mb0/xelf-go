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
This works really great.

The command should provide a collection of helpers organised as subcommands. Most subcommands should
read stdin and print to stdout, so that we can easily compose commands.

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

We currently load all plugins that we can find whenever a program is prepared. This is wasteful if
all we do is evaluating simple expressions. We could use a plugin manifest file that declares all
the xelf module paths it provides. That way we could add a lazy module registry that only loads the
plugins it needs once. It is also the case that plugins overlap: dapgx for example provides all daql
modules through its package dependencies, so if we load the dapgx plugin we can skip loading the
daql plugin.

It would be great for daql qry backend providers to use modules in the same way. For now we can
simply provide an empty module that ensures the provider is registered.

If we use xelf plugins to provide runtime modules and other extensions, we might also want to
lazy load and provide subcommands from plugins. This would allow us to drop the daql and layla
command entirely and use subcommands like `xelf daql graph` instead. Users that prefer simpler
commands can use bash aliases. The positive aspects are:

 * Projects can easily provide specialized subcommands without adding any extra dependencies.
 * We do not pollute the system path namespace and it is easier to discover features.

Links
-----

 * https://pkg.go.dev/plugin go plugin package docs
 * https://appliedgo.net/plugins/ gives a good overview over external plugins in go
 * https://github.com/hashicorp/go-plugin is an implementation using net/rpc or grpc

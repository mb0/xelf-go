Modules
=======

We want to add modules to encapsulate and manage named types, specs, and configuration.

Problem
-------

Type selection in type references uses the dot for both schema qualified types and selections. We
want to avoid that ambiguity and reuse selection code where possible. This simplification reaches
down into tooling code providing code completions.

We want xelf to work by itself as reasonable format for simple configuration files. Allowing values
in modules lets us easily import and compose configuration files without extra platform support.

To support a generic xelf repl we need to allow to dynamically configure the program environment. We
also want to be able to mark module uses explicitly. We can then start a generic clean environment
(e.g. for a repl) that can use hooks to load modules on demand. Modules should ideally encapsulate
capabilities like qry or dom resolution.

Providing tools and hooks for a xelf program to resolve external modules uses is generally useful;
and specifically for daql project files and repl as well as the layla project. To encapsulate these
capabilities in modules is some work but would go a long way towards usability and extensibility.

Use case
--------

project conf: one often has a common project configuration as well as deployment target specific
configuration, modules let use different files that are easily composed.

daql repl: we want start a blank daql repl - if not a generic xelf repl, that we use to load
a plain xelf configuration file for the development db details, pull in the dom package to define
a new schema, pull in gen to experiment with code generation, pull in mig to migrate the db and
insert some fixtures, pull in qry to read results from the database, pull in a script with reporting
helpers, that use layla for pdf reports and then export the whole repl session as script file.


Implementation
--------------

Modules represent a qualifier to group types, specs and values and on the other hand optional an
mechanism to add language capability extensions. Modules are unique within a program and setup hooks
are only executed once. They cannot be selected, instantiated or used as a value type, but can be
used for symbol resolution. To enforce that we introduce a mod meta kind to mark the mod types and
values provided to the language machinery. The module tree must be acyclic: a module cannot have
itself as dependency.

Special modules that require platform support must be registered into a module registry and may have
a setup hook that is called once to add themselves to a program environment.

A loader environment adds module awareness to the [program environment](./prog_env.md) and provides
the foundational specs to interact with modules. The loader env has a user supplied list of loaders.
The loader implementations loads module file sources for specific protocols. The raw input is then
evaluated and cached by the loader. Source mods are evaluated in isolation of any loading program.
All module dependencies are then registered with the loading program.

The basic module and file data structures are part of exp package. Every program environment stores
a module file with reference to used and declared modules and resolves qualified module symbols.

The `mod` form creates and registers a simple module with a module name and tags of named values and
returns void. This form creates a mod scope, that resolves its definitions as unqualified names. The
declared module is available after its declaration in the parent file env.

The `use` form loads modules into the file env and returns void. Use takes constant strings as
module paths or tagged paths to alias a specific module. The used modules are available after the
call in the parent loader env.

The `export` form loads modules just like the use form but also re-exports used modules.

	# file: /lib/company.com/prod/mod.xelf
	(mod prod
		Prod:<obj ID:int Name:Str>
	)
	# file: /home/work/util.xelf
	(mod util
		Mode:'dev'
		DBName:'myproj-dev'
		Foo:(fn a:str (lower a))
	)
	# file: /home/work/main.xelf
	(use 'company.com/prod' './util')
	(prod.Prod id:1 name:(util.Foo 'test'))

Discussion
----------

I experienced that using the simple module name as qualifier like go does is very readable and want
to use this as default for external modules. We need to be able to locally alias a module name to
resolve potential naming conflicts. Defaulting to qualified symbols and whole module uses we usually
do not want to use a filter syntax in the use statement. When not filtering module uses we can
accept multiple packages with one use or export call.

Daql projects and schemas should extend the module system and allow schema specific extra data.

Currently the program environment must be explicitly prepared to use daql packages. We would add dom
and qry modules to encapsulate that setup. The dom package would add its specs and export the
project, schema and model registry, that is then used to build up the dom schema. We still plan to
import and register the module at compile time in these cases.

We want to support process external code generators for the daql command. Maybe we can add a plugin
module wrapper that support external processes using something like github.com/hashicorp/go-plugin.

Daql and layla use the corresponding file name extension to indicate the expected xelf environment
maybe we should converge on the xelf file with a use header. The idea is that we would make things
easier for composing different extensions, tooling, being more explicit about requirements.

We want to import two versions of the same module. This would be especially handy for migrations
where we may want to import both the old and new schema. The thing that complicates multiple type
versions is that we currently use a program scoped type registry that uses the declaration name for
proxies and module types. We describe the [task in more detail](./lit_reg_split.md).

If we restrict the module loading environment to be independent of the loading program or arguments
(ideally immutable and static), would allow us to cache and reuse the results for that environment.

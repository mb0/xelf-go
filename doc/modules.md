Modules
=======

We want to add modules to encapsulate, filter and manage named types, specs, and configuration.

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

Mods represent on one hand a qualifier for types, specs and values and on the other hand optional
escape hatch to add language capability extensions. They cannot be selected, instantiated or used as
a value type, but can be used for symbol resolution. To enforce that we introduce a mod meta kind
to mark the mod types and values provided to the language machinary.

The `mod` form creates and registers a simple module with common values. Special modules that
require platform support cannot be declared in xelf and must be registered with the runtime.

The `use` form loads modules into the environment. The form allows to alias modules. The
implementation calls hooks in the program environment to resolve the external modules. The module
resolution can be swapped with an user implementation and must be expected to involve io errors. Use
takes constant strings or symbols as package identifier or tags thereof to alias the modules.

Modules can use other modules, but must not have import cycles.

Module files should evaluate to a common literal value that can be created with the generic mod form
or maybe with custom forms e.g. the daql schema form.

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
	(use 'company.com/prod' './util.xelf')
	(prod.Prod id:1 name:(util.Foo 'test'))

Discussion
----------

I experienced that using the simple module name as qualifier like go does is very readable and want
to use this as default for external modules. We need to be able to alias a module name to resolve
potential naming conflicts. Defaulting to qualified symbols and whole module uses we usually do
not want to use a filter syntax in the use statement. When not filtering module uses we can accept
multiple packages with one use call.

Figure out if daql projects and schemas should also be modules and whether and how we should allow
nested modules. Daql uses project files to include external daql schemas.

Currently the program environment must be explicitly prepared to use daql packages. We would add dom
and qry modules to encapsulate that setup. The dom package would add its specs and export the
project, schema and model registry, that is then used to build up the dom schema.

Daql and layla use the corresponding file name extension to indicate the expected xelf environment
maybe we should converge on the xelf file with a use header. The idea is that we would make things
easier for composing different extensions, tooling, being more explicit about requirements.

We want to support process external code generators for the daql command. Maybe we can add a plugin
module wrapper to support external processes.

We may want a simple way to re-export included modules and allow unqualified module use for system
modules. To easily compose and load common environments.


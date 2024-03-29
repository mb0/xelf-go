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

We want to import two versions of a module, or more generally allow importing modules with the same
name. This would be especially handy for migrations, where we want to backup the project data with
the corresponding project schemas and then later import both the old and new schema to migrate the
backup data.

Use case
--------

project conf: one often has a common project configuration as well as deployment target specific
configuration, modules let us import different files that are easily composed.

daql repl: we want a generic xelf repl, that we use to load a plain xelf configuration file for the
development db details, pull in the dom package to define a new schema, pull in gen to experiment
with code generation, pull in mig to migrate the db and insert some fixtures, pull in qry to read
results from the database, pull in a script with reporting helpers, that use layla for pdf reports
and then export the whole repl session as script file.

Implementation
--------------

Modules are just object literals qualified by a simple key, but as concept a language extension
mechanism. Modules are unique per program and cannot have themselves as (indirect) dependency.

Loaders locate, load and cache raw module sources by url. Sources are program independent and
represented either as ast or as program specific setup hook.

A loader environment stores module loaders and provides the foundational specs to interact with
modules. The loader environment loads the module sources and evaluates them to a module file.
Files provide a url and a list of references to imported and exported modules. Module reference keys
must be unique with the file. Module reference aliases can be used to disambiguate otherwise
conflicting imports.

The module and file data structures are part of exp package. Every program environment contains a
root file, and a list of all loaded files. The program resolves qualified module symbols and
attempts to update all name types crossing a file boundary to ensure valid references.

The `module` form creates and registers a simple module with a module name and tags of named values
and returns null. This form creates a mod env, that resolves its definitions as unqualified names.
The declared module is available after its declaration in the parent program env.

The `import` form loads modules into the program env and returns null. Import takes constant strings
as module paths or tagged paths to alias a specific module. The imported modules are then available
in the program env. A path fragment or the alias itself can be used to pick specific modules from a
file with multiple modules.

The `export` form loads modules just like the import form but also re-exports used modules.

All module specs and source module declarations are evaluate full when resolved.

	# file: /lib/company.com/prod/mod.xelf
	(module prod
		Prod:<obj ID:int Name:Str>
	)
	# file: /home/work/util.xelf
	(module util
		Mode:'dev'
		DBName:'myproj-dev'
		Foo:(fn a:str (lower a))
	)
	# file: /home/work/main.xelf
	(import 'company.com/prod' './util')
	(prod.Prod id:1 name:(util.Foo 'test'))

Discussion
----------

I experienced that using the simple module name as qualifier like go does is very readable and want
to use this as default for external modules. The import spec can load whole files or pick single
modules out and use aliases to resolve naming conflicts with one call.

Daql packages dom and qry register module sources that prepare a program and provide extra data.
The packages must either be imported into the go binary to register the module sources or we can use
the xelf plugin system to load module plugins on demand. We can use the module plugin system to for
go-based code generators too.

Daql projects and schemas integrate with the new module system and export the node as dom property
and all model types by name.

Daql and layla used a corresponding file name extension to indicate the expected xelf environment.
While we can still use explicit extensions, we should not encourage it to simplify tooling. We can
use the module system to ensure the expected environment.

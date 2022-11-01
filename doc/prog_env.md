Program Environment
===================

We want to introduce a program scoped environment that marks the boundary between universal and
possibly shared environments and program bound environments.

Problem
-------

Working on modules we discovered problems with type reference resolution. We have a combined and
literal and type registry. But we want to lookup reference types from the scoped environment. We
need extend the env api or provide a wrapper to lookup type reference from the environment. In both
cases ideally want to reuse the normal env lookup chain. The environment resl method however takes a
program parameter which complicates the issue. 


Implementation
--------------

The Prog instance itself can be an environment. This would have following benefits:
 * access to the the program instance through the environment chain
 * while avoiding a double linked structure between prog and prog env
 * a handy api to lookup program scoped symbols
 * a better place for the type symbol resolution fall-back

We want to merge the ArgEnv with prog because it is a common use case that we can provide better api
for. Any program arguments should by definition should be in or underneath the program scope.

To avoid the name clash with resl we rename the main env api method to env.Lookup and drop the prog
parameter, that is now accessible from the env chain.

The new env can now be more easily wrapped and queried to provide type reference lookup. But we
have to change the type system to accept a lookup function.

We may want to pull some module code into the program and merge prog and module file env. The idea
is to only provide the module data structure and still use the mod loader env to provide a module
system implementation.
Program Environment
===================

We want to introduce a program scoped environment that marks the boundary between universal and
possibly shared environments and program bound environments.

Problem
-------

Working on modules we discovered problems with type reference resolution. We have a combined literal
and type registry. But we want to lookup reference types from the scoped environment. We need to
extend the env api or provide a wrapper to lookup type references from the environment. In both
cases ideally want to reuse the normal env lookup chain. The environment resl method however takes a
program parameter which complicates the issue. 


Implementation
--------------

The Prog instance itself can be an environment. This would have following benefits:
 * access to the program instance through the environment chain
 * while avoiding a double linked structure between prog and prog env
 * a handy api to lookup program scoped symbols
 * a better place for the type symbol resolution fall-back

We want to merge the ArgEnv with prog because it is a common use case that we can provide better api
for. Any program arguments should by definition be in or underneath the program scope. This allows
special system modules to safely extend the program environment by prepending to the program root
environment chain.

To avoid the name clash with resl we rename the main env api method to env.Lookup and drop the prog
parameter, that is now accessible from the env chain from inside the program. By convention
environments in the root chain should mostly be viewed as external and immutable with reasonable
exceptions.

The new env can now be more easily wrapped to provide type reference lookup function. The type
system was changed to accept such a lookup function.

The program also stores and resolves loaded modules using the new module system. This makes the
module loader program independent, and the program self contained.

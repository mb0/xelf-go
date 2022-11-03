Literal registry split
======================

We want to rethink responsibilities of the literal registry and arrive at a system with simpler api
and separated concerns.

Problem
-------

Carrying lit.Reg through the lit and exp apis makes implementing literals tricky and annoying. But
we need it because it provides foundational api to work with types and literals. We even embed it in
many value implementations to support external apis like UnmarshalJSON.

The registry originally had three roles:
 * a mapping from type ref to type definition
 * a mapping from type ref to value implementation
 * a type info and proxy cache for reflection types

The ref mappings are obviously per xelf program. However the mapping to type definition should
usually use the local environment scope, while the mapping from ref to value implementation must be
provided externally to the program scope.

The reflection type caches are strictly necessary to support recursive types, but are independent
of any particular xelf program and could be shared and process scoped (like reflection caches of the
json package for example).

The mapping from ref to value implementation at every step was done to have user provide proxy
literals at the runtime boundaries like the program results and in custom specs. This convenience is
payed for by wrapping lots of reflection code at every step of the way. The description already
suggests that we reduce use of proxies to api boundaries in the evaluation phase and otherwise only
convert from generic literals to proxies as last evaluation step.

The original intention was to always work with user provided proxies throughout the chain, as not to
double allocate large slices for query result sets, however any operation on proxies from the
language needs to again proxy all its element and itself allocates lots of wrappers too cheep to
pool but potentially numerous enough for memory pressure.

Implementation
--------------

The type system was already changed to use the program environment to resolve type references and is
discussed in the [field reference doc](./field_refs.md).

We factor out a reflection cache and use a process-shared global by default. We can still provide
isolated caches for tests. The new cache encapsulates the type reflection code for coarser grained
locking.

We keep the literal registry for mapping to user provided value implementations, and reduce its api
to zero and proxy methods and use it explicitly at api boundaries.

ParseVal and ParseMut are already independent from the registry because we use the new primitive
Vals and Keyed literals.

The proxy methods and values do inherently need the implementation cache to reduce the overhead of
wrapping elements in mutable proxy implementations. It makes sense for proxies to keep a reference
to the origin registry where all required types are already registered by definition. If we complete
the separation from the type reference lookup, we can then use shared proxy values without worrying
about managing registries, as long as we use concrete types everywhere.

We still need to investigate in what circumstances we want to use the registered proxies.

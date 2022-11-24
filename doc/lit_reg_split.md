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
discussed in the [field reference doc](./field_refs.md). Now only sys.Inst takes a lookup function
that usually wraps an exp.Env to resolve type reference.

ParseVal and ParseMut are already independent from the registry because we use the new primitive
Vals and Keyed literals. We introduce AnyMut similar to AnyPrx to provide an independent Zero
function.

We factored out a reflection cache as PrxReg and use a process-shared global by default. We can
still provide isolated caches for tests. The new cache encapsulates the type reflection and proxy
code for coarser grained locking.

We factored out a literal registry as MutReg for user provided mutable value implementations, and
reduce its api to the Zero and SetRef method and use it explicitly at api boundaries.

We keep a Regs type the embeds both PrxReg and MutReg to make it easier to pass around or initilized
both new registries in one go. We provide the Reg interface that provides only Reflect and
ProxyValue methods, but due to embedding is implemented by both Regs and PrxReg types.

The proxy methods and values do inherently need the implementation cache to reduce the overhead of
wrapping elements in mutable proxy implementations. It makes sense for proxies to keep a reference
to the origin registry where all required types are already registered by definition. We therefor
share proxies without updating the proxy registry.

Conclusion
----------

The original concept was fundamental and invasive yet unclear and with mixed responsibilities.
The new concept is clean, unobtrusive, convenient with good defaults. It took years to arrive at
the concepts for this literal system in totality and it is better than it was every before.

Because of all the changes we could remove the registry from a lot of code. The program constructor
was rewritten as an extra to be more elegant and convenient. Overall this effort took some lots of
work, but was very fruitful and is now completed.

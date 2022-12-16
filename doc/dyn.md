dynamic spec
============

We want to revisit the dyn spec and its role in the language.

Problem
-------

We want to allow late binding of call specs, and also some syntax sugar.

We want to allow users to change the program environment and evaluation, to potentially define a
whole other language within the limits of the ast, typ, lit and the basic program structure.

Implementation
--------------

We add a minimal dyn hook to the program that resolves call specs. The default dyn function returns
either resolved spec directly, the call spec for unresolved specs, the make spec for type values or
the mut spec for all other value types. The specs use the program scope for lookup and can be
replaced individually.

The call spec covers important corner cases by deferring the spec resolution to the evaluation
phase. However in contrast to the previous dyn spec we initially need to know that the value will
resolve to a spec. 

The mut spec covers all mutations by using delta edits and additionally provides some syntax sugar
for common but otherwise inconvenient expression for append and cat.

The make spec covers all type value construction and conversions and should mostly reflect the
behaviour of mut for a value of that type, with some additions:

	(typ .) return the resolved type of the argument
	(@T .) attempts to convert the argument to type @T

Discussion
----------

We did load a dyn spec on init from the program environment. The default dyn spec provided syntax
sugar for following literals:

  * spec: call    - implicitly part of the xelf definition
  * typ:  make    - very useful as core part of the language
  * str:  cat     - useful for formatting and templates, should cover all char
  * num:  add     - not really that useful, consequent if we allow char
  * list: append  - should probably be covered by mut as well
  * keyr: mut     - this is generally nice but could be great if this listop concept works

The omissions were because all options would be confusing and are better served by the make sugar.
This indicated the limited usefulness of extending the dyn sugar concept to custom types. Adding
custom sugar has a spec lookup problem as well, because much of the interesting types already
match the mut spec.

In practice the core sugar specs were only good if they were a consistent part of all dialects. I
decided that dyn (and syntax sugar in general) should either be a reliable or completely dropped.
The implicit behaviour change would be hard to keep in mind and the code would be too involved, to
justify the cost of that feature. That means the dyn behaviour should not change during program
resolution in any way.

The most essential feature dyn provided was late binding calls and the typ sugar, values of these
two types should not be mutated. All other values can however be mutated. So, provided we add an
explicit call spec for late binding and unify a mut spec, we can drop dyn and reduce the supported
sugar to only three specs: call, make and mut. We can instead provide a dyn hook to the program,
that can replace this behaviour.

We found a good syntax of mutate that covers all needs and use the same syntax for the make spec
`(@proposal.Rating tag:'Cool')`. We explore that question in [mutate spec doc](./mut.md).

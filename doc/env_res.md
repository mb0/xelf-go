Env Lookup Result
=================

We want to re-evaluate the result type of the env lookup api.

Problem
-------

Can we change the lookup return type to lit.Val?

The exp.Lit type is now mostly used to add input positions to values and to implement the exp.Exp
interface. In the evaluation phase we are not really interested in the source positions anymore,
we already use values for prog and spec eval. The same applies to symbol evaluation.

Originally the env interface mirrored both resolution phases and had a resl and eval method. The
idea was that an env can return either a literal value or the unresolved but updated symbol. Then we
unified it and renamed the api to lookup to avoid a name conflict in prog. Now we pass in the symbol
and always update it when we find a value. We returned the value as Lit to carry the resolved value
type. Now, that values can carry a specific type by themselves, we have no reason to use a Lit other
than matching the lookup api that expects an Exp.

Environments only ever return the passed in symbol or a resolve literal so far. I can think of no
reason to return anything other than the value or the same symbol. All the other expressions in make
no sense and would be quite confusing anyway.

Implementation
--------------

Lookup now returns a result value, an error or nothing, when eval is false and the symbol resolves,
but the value does not. We should always update the symbol when we found the key. We must return
resolved value or an error if eval is true. All the environment code can be simplified.

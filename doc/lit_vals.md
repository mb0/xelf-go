Literal values
==============

We want to reason about and clarify the concepts for literals, values, mutables and proxies.

Problem
-------

While working on modules the need arose to update all type reference of values when crossing a file
boundary. We can write a routine that covers common value implementations that use a type, or even
add a method to the val or mut interface. But we also have exp lit that should be used to specify a
type for unspecific values. We should look into better using literal types.

 * Can we only update the lit type and keep the value with original types?
 * Should we provide a better value type editing api and reconsider Lit.Res?
 * Should we use Spec as value directly?
 * Should we add a Val.Mutable() or Val.Identity() Val to the lit api to better distinguish values?

Redefining types on the literal level is mostly a problem of confusion. We have both Lit.Res and
Val.Type() that provide type information without a clear source of truth. In many places we only
have access to the value. Another idea for Lit was to resolve the type independent of any value.

 * If we select into a literal which type do we use for the results?

Discussion
----------

We have the simple values `Null, Bool, Num, Int, Real, Char, Str, Raw, UUID, Time, Span, typ.Type,
exp.Spec exp.SpecRef` that are mostly there for ergonomic reasons to allow simple values when
working in go. All simple values except Null, Spec and Spec implement mut when used as pointer.

All other wrapper and container values and all proxies already implement mut.

We added SpecRef explicitly to edit the declaration of a spec value we could make it a mutable too.
But we should clear up what a mutable spec val means, could we parse one?

One generally helpful addition would be the Val.Mutable api, because we can easily and safely return
or create a mut for any value and often want to do just that. That would let us redefine Val.Value
to return the most primitive value representation. It is however cumbersome to implement the api
for all specs. We want to return a mutable SpecRef that wraps the spec, but don't have parent spec
available in the embedded SpecBase. We could just return nil for plain specs and let the caller
handle it as special case? Or do we only use a SpecRef and don't even implement Val in Spec itself?
If we use SpecRef exclusively for spec values we can drop all the senseless spec value
implementations everywhere.

We may want to drop the resolved type of exp.Lit so we have only the value type. This would resolve
a lot of type confusion and would keep the type near to the backing value. We still need exp.Lit to
provide source info and implement exp.Exp.

Instead we could provide a Mut.EditType(typ.EditFunc) api that returns the same or a new mutable
with updated type. We should be careful that the new type is compatible.

We can add a value wrapper that provides a new type or even an ast value that uses raw input until
evaluated we can probably reuse and maybe unify with AnyMut and OptMut.

Implementation
--------------

SpecRef is now a mutable value, it supports null specs, new and assign but not parse.
It is the only spec value representation, Spec itself does not need to implement Val anymore.
It still itself implements Spec and should be used as such.

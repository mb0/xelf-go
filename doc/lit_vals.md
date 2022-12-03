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

One problem with the mutable implementation for simple values is that every call to an value method
dereferences the pointer. This was no problem so far, but with the addition of Mut and As we
actually preserver the pointer. This behaviour is very surprising and should be fixed. Moving the
api into mutable makes makes no sense for Mut and loses much of the convenience for As.

We may want to drop the resolved type of exp.Lit so we have only the value type. This would resolve
a lot of type confusion and would keep the type near to the backing value. We still need exp.Lit to
provide source info and implement exp.Exp.

Instead we could provide type editing api for values that returns the same or a new mutable with
updated type. We should be careful that the new type is compatible. We probably need to be thorough
and edit even element values, that means we also need a value editor with state to handle self
referential values.

What do we mean by compatible type when updating a value type?
 * The current value must certainly be assignable to values of that type.
 * Therefor the current type must be assignable to the new type. That covers `<int>` to `<num>` or
   about any type generalization, but would fail for `<num>` to `<int>` - type specialization.
 * We can use a convertible-to check, that covers both cases, but it would cover more than we want.
 * We need to check the actual value and determine whether the conversion is ok.

We can add a value wrapper that provides a new type or even an ast value that uses raw input until
evaluated we can probably reuse and maybe unify with AnyMut and OptMut.

We noticed that List, Dict and Map allocate a new element type body for every call to Type, this
is wasteful if we want to use the value type information to match values; and unfortunate if want
use list types with names or ids.

We noticed too that a type `<int>` used as value identifies as `<typ>` and not as `<typ|int>`. We do
that to avoid the allocation. The same is true for the expression types. All these types use element
bodies and not a flag kind, so we can differentiate between `<typ?|int>` and `<typ|int?>`. It would
be great if we can report the full type without allocation and without work to keep the types in
sync.

Almost all type bodies are basically dressed up pointers to either a type or a slice. But
specifically for element bodies we could use type pointer directly instead to the same effect.

We could eventually drop the SelBody and use an element body too, but would need to reuse the type
ref field, a selection should never be named or a reference.

If we add a broader value conversion api to values, type values need to support this too. One way is
to pull a value wrapper into the typ package that decorates types. This would effectively mean, that
we have two type values, this is however already the case for every other value that is wrapped.

But this gives rise to another challenge to use a wrapper in package typ we want the null value
there too. We can use the same wrapper to wrap null successfully, but then we can still not provide
the wrapper as default any value, because we need to have access to a literal parser.

So we can really only define the wrapper for none null mutable values to support parsing? We cannot
pull the literal parser into the typ package without all the default implementations with it.
This is all fine for providing wrappers for type values itself, but makes it impossible to support
null mutable or typed null wrapper… unless we also provide a typed raw ast value that we can use to
defer parsing. Or we add a hook for the typ package for providing parsing capabilities for wrapped
nulls.

Implementation
--------------

SpecRef is now a mutable value, it supports null specs, new and assign but not parse.
It is the only spec value representation, Spec itself does not need to implement Val anymore.
It still itself implements Spec and should be used as such.

We add a `lit.Val.Mut() (lit.Mut)` function to the api. It returns the same or a new mutable value.
We change the embedded opt mut field to LitMut to remove the name conflict.

We redefine `lit.Val.Value() (lit.Val)` to always convert to minimal set of types. Idxr and Keyr
values can return as-is to avoid excessive allocations, users can simply create an implementation
of choice and use assign. In places where we used it to unwrap wrapper types, we now correctly use
`lit.Unwrap(lit.Val)` function that unwraps all layers of wrapper values. We add a small Wrapper
interface to allow external wrapper implementations.

We changed List, Dict and Map to store the full type to avoid allocations when checking value types.

We use a type pointer as type body directly instead element bodies, this allows us to wrap element
types with minimal additional allocation as long as the element type is addressable, this also
keeps an element type and the wrapped element in sync if used carefully. We drop both Exp.Kind and
Exp.Resl in favor of Exp.Type, that unifies both and now return `<call|int>` for example. We rename
typ.ResEl to typ.Res to make the resulting code easier to read.

We add `typ.Wrap` as value wrapper that allows generic type redefinition to compatible types. It
also covers null handling for values that do not. It unifies and obsoletes OptMut and AnyMut.
We introduce `typ.AstVal` to defer value parsing and `lit.Any` as a simplified replacement for
AnyMut. Both provide orthogonal features that stand on their own, but they also provide different
strategies to work around the fact, that the wrapper has no access to literal parsers and
value implementations directly. Finally we add a WrapNull hook to the typ package that the lit
package assigns to on init, that returns a wrapper around any that has access to literal parsers.

We add `Val.As(typ.Type) (Val, error)` and `lit.Edit(Val, EditFunc) (Val, error)` to the value api.
Together with 'typ.Edit' we can provide a `lit.EditTypes(Val, typ.EditFunc) (Val, error)` function.

The new `As(Type)` method provides a general conversion mechanism to compatible types. The value
type must be convertible to the new type, and any actual value conversion must succeed as well.
The method can help with adding type checking to value mutations.

We introduced dedicated mutable value types for primitive values. That means `Int` is a simple Val
and `*IntMut` a Mut. `*Int` should not be used from now on. This works still great for simple values
input and representation and works well with the new `Val.Mut` and `Val.As` api without surprising
corner cases.

We removed the Lit.Res field and use the value type exclusively. If we have differing types we
should use Val.As that wraps the value with a new type.

We removed the use of Lit from the eval and env api, reducing lots of wrapper bloat at call sites,
and making the code more palatable overall.

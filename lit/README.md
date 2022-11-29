xelf/lit
========

Xelf needs a literal representation to parse into and work with.

Xelf has the `Lit` type representing a literal expression that contains the resolved type, the
source information and an abstract value. The omitted value source location and the independent
value type simplify the implementation and allow lazy conversions.

We use three interfaces that describe sets of value implementations:
 * `Val` for simple values
 * `Mut` for mutable values
 * `Prx` for mutable proxy values

We have three ways to parse an `Ast` to a value implementations:
 * `ParseVal`  returns generic values, using immutable values for primitive literals
 * `ParseMut`  returns generic mutable values
 * `Mut.Parse` parses into a mutable value provided by the user.
    The parse method can be optimize and allows user defined values.

There are other some helper methods:
 * `Read`,  `ReadInto`  to read from a named reader
 * `Parse`, `ParseInto` to read from a string

We have another set of interfaces to cover capabilities:
 * `Idxr`     for indexable values like list or obj
 * `Appender` for appendable values like idxr, list
 * `Keyr`     for keyable values like dict or obj
 * `Wrapper`  for value wrappers that would hide the other interfaces

`Reg` is a registry interface for a reflection and proxy cache. Caching is required, not only to
improve performance, but to support self referential types. `PrxReg` is the notable implementation
that can be shared as process global cache. The interface provides:

 * `Reflect(reflect.Typ) (typ.Type, error)`
 * `ProxyValue(reflect.Value) (Mut, error)`

`MutReg` is a second optional registry for named and user-provided value implementations. It
provides a custom Zero method as api, that can be used instead of the generic package function.

`Regs` is a simple struct that embeds both registries, to make it easier to pass around.

The registry must be used to creating new proxy values. Created proxies keep a reference to their
origin registry to create proxies for contained values. Proxy values may point to a pointer and
represent an optional type directly.

The primary value implementations can be used as value `Bool` or as mutable `*Bool`, because working
exclusively with pointer types is cumbersome for simple values: `(*lit.Bool)(cor.Bool(true))`. Both
options however only represent a value with the type `bool`, we provide an OptMut wrapper with a
null flag of type `<bool?>`. The Null type has only a value implementation.

Other implementations are always mutable variants, because we would gain nothing from using a value
type implicitly addressed and wrapped in an interface and instead would increase code complexity.

The generic `Map` implementation can explicitly be used instead of dicts by users provided types to
make working with dicts easier. `Dict` is a useful default because it preserves order which may be
important for some internal conversions and program resolution.

Another special implementation is the `AnyPrx` that proxies to interface values with resolved type
alternatives and manage a literal that represents that interface value.

The `MapPrx` uses a neat trick to provide mutable element values even though go map elements are not
addressable without using a pointer element type.

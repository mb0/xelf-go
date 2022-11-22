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
 * `Read`, `ReadInto`, `ReadIntoMut` to read from a named reader
 * `Parse`, `ParseInto`, `ParseIntoMut` to read from a string

We have another set of interfaces to cover capabilities:
 * `Idxr` for indexable values like list or strc
 * `Apdr` for appendable values like list
 * `Keyr` for keyable values like dict or strc

`Reg` is a registry for custom mutable values and provides api to work proxies in general. Reg uses
the global reflection cache by default, to support self referential types and improve efficiency.

The registry must be used to creating new values or proxies. These operations are used all over the
place and do happen deep in call stacks. Passing the registry through as argument looks bad and is
not even used by some value implementations. We obviously have the registry available wherever we
first initialize types so we just save the registry in the value implementation.

The primary values (bool, int, real, str…) and type have implementations that can be used as value
`Bool` or as mutable `*Bool`. We allow the values for primary types because working with pointers to
primitives is cumbersome `(*lit.Bool)(cor.Bool(true))`. Both options however only represent an bool,
therefor we use an OptMut internally that has a null flag (similar to sql.NullX types but works for
any mutable value). The Null type has only a value implementation.

Other implementations are always mutable variants, because we would gain nothing from using a value
type implicitly addressed and wrapped in an interface and instead would increase code complexity.

All proxy values can point to pointer and then represent null directly without using an OptMut.

The generic `Map` implementation can explicitly be used instead of dicts by users provided types to
make working with dicts easier. `Dict` is a useful default because they preserve order which may be
important for some internal conversions and program resolution.

Another special implementation is the `AnyPrx` that proxies to interface values values with resolved
type alternatives and manage a literal value to represent that interface value.

The `MapPrx` uses a neat trick to provide mutable element values even though go map elements are not
addressable without using a pointer element type.

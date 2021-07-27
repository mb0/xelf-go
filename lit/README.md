xelf/lit
========

Xelf needs a literal representation to parse into and work with.

Xelf has the Lit type representing a literal expression that contains the resolved type, the source
information and an abstract value. The omitted value source location and the independent value type
simplify the implementation and allow lazy conversions.

We use three interfaces that describe sets of value implementations:
 * Val for simple values
 * Mut for mutable values
 * Prx for mutable proxy values

We have three ways to parse these value implementations:
 * ParseVal returns generic values that may be immutable
 * ParseMut returns generic and registered mutalble values
 * ParseInto parses into a mutable proxy value provided by the user

We have another set of interfaces to cover capabilities:
 * Idxr for indexable values like list or strc
 * Apdr for appendable values like list
 * Keyr for keyable values like dict or strc

We need a regsitry where proxies are registered. The registry then provides implementation for xelf
types, when no proxy is found a generic implementation is provided. We use the same registry as
reflection cache to break self referencial type cycles.

The registry must be used to creating new values or proxies. These operations are used all over the
place and do happen deep in call stacks. Passing the registry through as argument looks bad and is
not even used by some value implementations. We obviously have the registry available wherever we
first initialize types so we just save the registry in the value implementation.

Mutable values implement a parse method to optimize and allow user defined parsing into proxies.

The primary values (bool, int, real, str…) and type have implementations that can be used as value
`Bool` or as mutable `*Bool`. We allow the values for primary types because working with pointers to
primitives is cumbersome `(*lit.Bool)(cor.Bool(true))`. Both options however only represent an bool,
therefor we use an OptMut internally that has a null flag (similar to sql.NullX types but works for
any mutable value). The Null type has only a value implementation.

Other implementations are always mutable variants, because we would gain nothing from using value
type and instead would increase code complexity.

Another special implementation is the AnyPrx that proxies to interface values values with resolved
type alternatives.

The generic Map implementation can explicitly used instead of dicts by users provided types to make
working with dicts easier. Dicts are a useful default because they preserve order which may be
important for the internal conversion and resolution process.

The MapPrx uses a neat trick to provide mutable element values even though go map elements are not
addressable.

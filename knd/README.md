xelf/knd
========

Kind is a bitset that describes language elements and types in a compact and practical way.

Then we have expressions and types as part of the language:
 * lit  for literals
 * typ  for types
 * sym  for symbols
 * tag  for key expression pairs
 * tupl for expression lists
 * call for resolved expressions

We want to support unmodified JSON literals:
 * none for null
 * bool for boolean
 * num  for number
 * char for string
 * idxr for array
 * keyr for object

We want specialized sub types:
 * num:  int, real, bits
 * char: str, raw, uuid, span, time, enum
 * idxr: list
 * keyr: dict, rec, obj

We have functions, type variables and references, and alternative types.
 * meta: alt, var, ref, sel
 * spec: func, form

And some super types:
 * exp:  lit, typ, sym, tag, tupl, call
 * schm: bits, enum, obj
 * prim: bool, num, char
 * cont: list, dict
 * strc: rec, obj
 * data: prim, cont, strc
 * any:  data, none

All individual bits signify concrete types. Abstract and base types use a mask of all the
allowed concrete type bits.

For compatibility reasons we use only 32 bits, because all bitwise operations trunc a numer to 32
bit in javascript.

The `none` kind describes the type of the `null` literal. It is often used in type alternatives and
has a special suffix notation using `?`. The corresponding `some` kind with the `!` suffix
explicitly flagging the type as not null.

    (if (ne <str?> <alt str none>) (fail))
    <form dflt val:@1 default?:@1! @1!>

Kinds can be marked either as `lit|T` or `typ|T`. Kinds without prefix are treated as `lit|T`.
One special rule for variables, selections or references is to transform the target type.
The `lit` prefix is usually omitted except for identifying expression kinds or to transform a type.

    <form con typ|@1 … lit|@1>
    <form typof lit|@1 typ|@1>

The `none` bit and `prim` bits can be combined to a primitive alternative type that does not need
the `alt` bit. The bit is explicitly used whenever an alternative has an id a type body or is an
unresolved type. In that case all primitive alternatives are still combined into the kind bitset.

The expression bits `sym`, `tag` and `call` can contain an element type. These bits are used as
expression kinds and may in a type context be otherwise used as hints for the automatic resolution
process for custom specs.

A conditional spec for example does not want the automatic resolver to evaluate branches but does
want the type interference to work:

    <form if ifs:<list|tupl cond:any act:exp|@1> els?:exp|@1 @1>


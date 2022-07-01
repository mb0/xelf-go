Field References
================

While building a draft for a reusable daql web ui, that provide some basic generic admin views
for dom models, it became clear that it would be great to include referenced models in both
reference and primary key fields.

Daql is meant to be used to only query required fields from a model. By narrowing the fields
we basically create a new record type with the narrowed field set, and therefor lose the model
information. Even if the user does not explicitly narrow the field set, the server should be able
to automatically narrow it to the users read permission for example.

We could introduce a derived flag so we can still use object type with the model name.
But that solution seems wrong very fast. If we use a free-form query input and we query a list
of product ids: `(*prod.Prod id;)` we now get a `<list|rec id:int>` with the derived flag maybe
`<list|obj prod.Prod* id:int>` but then we probably only want a list of ids not nested in objects
and query `(*prod.Prod _:id)` now we get `list|uuid` which leaves us guessing how to generically
look up labels for example.

We would not only need the result type but also parse and analyse the query itself to decide
what model the uuid references if anything.

It would be much better if each reference field would carry the reference in the type.
So we want to get `list|uuid@prod.Prod` in the example above or something similar.
We currently have not concept of reference fields in xelf.

We already use @ref for named type references and knd@1 for type variables. We want to generally be
able to print a type with its id and reference. Even for unresolved type references.

Type reference need fixing and are not yet well thought out. The original idea was that the at-sign
is a generic prefix for any symbol to return the type, but that does not work that great, because
we use different resolution strategies for expression and type symbols. We can even use the
default conv with sugar to write (typ sym) to get to the type of any expression.
For the cases where we now use type references, there is often no clear distinction between
`@Ref` and `(obj Ref)`, except that the type ref has less information. We should prefer to have
one way to refer to things and it should be the more explicit one.

In Daql we have a special rule that allows reference to object fields like @prod.Prod.ID,
if we want to use object names in primary keys, this would be self referential and still not specify
the field type and look very funny `<obj prod.Prod ID:@prod.Prod.ID>`, so that option does not pan
out.

So in conclusion we want to re-imagine type refs and introduce a unified or separate concept for
references to schema types. We need to keep in mind and specify which types can be usable as
primary keys and whether or not to use the same concept for enum and bits types.
We should even go so far as to re-evaluate the concept of schema types within and outside the
context of the daql project.

Schema types
------------

Right now schema types are an interface primarily implemented by daql for dom models, but also to
manage language interoperation like proxies. Daql uses special forms to define models where
type references can be used to define enum and bits fields as well as reference fields.
Creating a Proxy values usually does not have enough information to infer references, but we could
use struct field tags in go to supply that information.

Because xelf really does not use schema types in the way daql does, maybe we should decouple these
concepts. For xelf a schema type is a name for an underlying type with potentially additional
properties implemented in other projects. We have some kind of registry where we can give
these names to types. Other than normal symbol resolution this registry should be usable
outside program evaluation to resolve the named types. Enums and bits could potentially
make good use of language support, maybe we should pull these types closer into the xelf library
and either reduce schema types to the obj type or extend schema types to any named type.

If we accept that every type can be named, we can rethink the other cases where we use names like
spec types. So object types would be just named record types, the enum and bits kinds would
be unnamed xelf types with const body that can be given a name like any other type.
With names for any type we can be more explicit in our models and possibly attach additionally
functionality.
Forms and funcs would be anonymous by default too. But what syntax do we use and where do we put
the name? We don't want to bloat the Type struct because it is used as value type. On the other hand
we also don't want to add another type wrapping only for names. The compromise would be to put the
reference name in the type body as we do now and add an extra named body type for types that do not
need a body otherwise. If every type can have a name we must think about how name declaration works.

`<list@Assets|rec@Asset Name:str Data:raw>`
`@Asset.Name` -> `str@Asset.Name`
`@Assets/Name` -> `list|str@Asset.Name`

If we allow selections in references we should normalize to a definitive representation.
We could treat references similar to ids, in that if we specify a reference it always has the same
id and that we normally do show the reference instead of the id.

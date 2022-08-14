Field References
================


Problem
-------

While building a draft for a reusable daql web ui, that provide some basic generic admin views for
dom models, it became clear that it would be great to include referenced models in both reference
and primary key fields. Daql is meant to be used to only query required fields from a model. By
narrowing the fields we basically create a new obj type with the narrowed field set, and therefor
lose the model information. Even if the user does not explicitly narrow the field set, the server
should be able to automatically narrow it to the users read permission for example.

While touching the issue of type reference we can rethink schema types in xelf altogether.
Schema types were an interface primarily implemented by daql for dom models, but also to
manage language interoperation like proxies. Daql uses special forms to define models where
type references can be used to define enum and bits fields as well as reference fields.
Creating Proxy values usually does not have enough information to infer references, but we could
use struct field tags in go to supply that information.

Implementation
--------------

Field references are pulled into the xelf language and declare self references for primary keys.
(Instead of using and managing derived model definitions.) Then model identifying information is
kept through filter steps as long as the primary key is kept. The daql query can just return the
result type that provides all type information. So reuse the existing xelf reference notation:
`list|@prod.Prod.ID`.

The reference field is now part of type structure and is kept after resolution. The ref field also
replaces the type name of const, obj and spec types; this changes the type name syntax to
`<func@atoi str int>`. Reference resolution specializes the type, sets a type id and normalizes the
reference. The reference must always resolve to a assignable super-type.

Type declaration is only possible outside the type context for now. We first implement the schema
type registry, but must think about program scoped type references.

Printing resolved type refs omits the type body for non-spec types, but keeps the kind if available.
This allows some local decisions and potentially avoids a registry lookup completely for types
without body.


Discussion
----------

Custom syntax for referential fields. In daql we had special rules that allows refs to object fields
like `Prod:@prod.Prod.ID` or embeds `@Common;`. We should pull these concepts into xelf with general
field refs available and add a custom syntax for self referential primary keys: `<obj@prod.Prod
ID:uuid@@>` equals `<obj@prod.Prod ID:uuid@prod.Prod.ID>`.

Program scoped type declaration must be possible. The question is if we use the scoped environment
or have special forms to declare types directly into a program global type registry.

Registry management is required for schema references. There are three layers of named types:

 * Hard-coded schema types per process are part of the program code
 * Dynamic session scoped schema types, can be queried from any daql endpoint
 * Program scoped types, probably a form to declare globally scoped named types.

Tasks
-----

  * [x] Unify obj and rec kinds and drop unused rec, strc and schm kinds
  * [x] Add and use ref field instead of body names and drop name helpers
  * [ ] Update type ref resolution
  * [ ] Factor out filter routine for obj types that keeps references
  * [ ] Pull special reference notation for obj types into the xelf type syntax
  * [ ] Add pk flag to kind bitset and self referential syntax-sugar
  * [ ] Implement and use a layered type registry
  * [ ] Update typescript port

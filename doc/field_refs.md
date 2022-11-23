Field References
================

We want to unify schema types and references to improve utility of the type system.


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

Work on a [Module System](./modules.md) introduced a clean concept for qualified names.

Type ref lookup api was changed to allow references to the whole environment. This adds lots of
power and flexibility, but we need to revisit keeping which reference around. We want to ensure that
all publicly visible references are resolvable from the program scope.

Printing resolved type refs omits the type body for non-spec types, but keeps the kind if available.
This allows some local decisions and potentially avoids a registry lookup completely for types
without body.

Discussion
----------

Custom syntax for referential fields. In daql we had special rules that allows refs to object fields
like `Prod:@prod.Prod.ID` or embeds `@Common;`. We should pull these concepts into xelf with general
field refs available and add a custom syntax for self referential primary keys: `<obj@prod.Prod
ID:uuid@@>` equals `<obj@prod.Prod ID:uuid@prod.Prod.ID>`.

Module uses allow local aliases that effect the local reference name, but store enough information
to resolve the original reference. That however makes qualified type names file-scoped and not
globally scoped, we may want to decide that each input must itself be evaluated as a separate
program as it is now, and separate the environment even more and then instantiate the module results
in the main program. We add a feature to the mod spec to accept named types as declarations with a
module local name to avoid stutter in any case.

We need to better define type name declaration. It seems that we have two distinct contexts:
 * The type context where names are not exported, but can be referenced locally in that type literal
 * The program context where names must explicitly be registered (usually reusing, sometimes
   qualifying the type name).

Tasks
-----

  * [x] Unify obj and rec kinds and drop unused rec, strc and schm kinds
  * [x] Add and use ref field instead of body names and drop name helpers
  * [x] Update type ref resolution
  * [x] Implement and use a layered type registry
  * [ ] Pull special reference notation for obj types into the xelf type syntax
  * [ ] Add pk flag to kind bitset and self referential syntax-sugar
  * [ ] Update typescript port

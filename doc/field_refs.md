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

Field references are pulled into the xelf language and also declare self references for primary
keys. (Instead of using and managing derived model definitions.) Then model identifying information
is kept through filter steps as long as the primary key is kept. The daql query can just return the
result type that provides all type information. So reuse the existing xelf reference notation:
`list|int@prod.Prod.ID`.

The reference field is now part of type structure and can be used to name concrete types. The ref
field also replaces the type name of const, obj and spec types; this changes the type name syntax to
`<func@atoi str int>`. Reference resolution specializes the type, sets a type id and normalizes the
reference. The reference must always resolve to a assignable super-type.

Work on a [Module System](./modules.md) introduced a clean concept for qualified names.

Type ref lookup api was changed to allow references to the whole environment. The resolved type
names are determined by the lookup function. Overall this adds lots of power and flexibility.

Name declarations are used within a type literal to name types and mark field references. Names and
field references are checked against the file scope when resolved, and must be correctly qualified
or dropped. Both must be checked again when crossing a file boundary.

Resolved type names and field refs must be resolvable by the program env of the current file.

Printing resolved type refs omits the type body for non-spec types, but keeps the kind if available.
This allows some local decisions and potentially avoids a registry lookup completely for types, that
cannot have a body.

Discussion
----------

Custom syntax for referential fields in dom specs allows refs to other models and model fields like
`@prod.Prod.ID;` or embeds `@Common`. We could pull these concepts into xelf.
We want to think about a solution to mark primary keys: `<obj@Prod ID!:uuid>` might resolve to
`<obj@prod.Prod ID!:uuid@prod.Prod.ID>`. Using the name suffix '!' corresponds to the '?' for
optional field and clearly marks the pk field as non-optional in a way. This would keep the obj
field specific flag out of the general type data structure and its bitset.

We want that publicly visible references can be resolved by the program env. That usually includes
the types used by builtins, mods and arg value. That also means we only want to keep references that
point to or into named types or field refs. To restrict resolved type names we can check if the
resolving env is prog or beyond, and otherwise attempt a second resolution using the program env.

 * limit names to program env, lookup check resolved sym env for prog or beyond
 * enforce correct names, check ref against type name, replace alias or other parts
 * edit type to change or clear all names
 * refs must use file local aliases, type names should match that
 * replace type names at file boundary in prog env

Tasks
-----

  * [x] Unify obj and rec kinds and drop unused rec, strc and schm kinds
  * [x] Add and use ref field instead of body names and drop name helpers
  * [x] Update type ref resolution
  * [x] Implement and use a layered type registry
  * [ ] Pull special reference notation for obj types into the xelf type syntax
  * [ ] Add pk indicator to obj fields
  * [ ] Update typescript port

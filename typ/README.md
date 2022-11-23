xelf/typ
========

Xelf is a typed language. We use a kind bitset from package `knd` to describe many details.
All types can have a type id and a name, type variables and references use these internally for
resolution. Complex types need to store additional information in specialized type bodies.

Types are naturally enclosed in angular brackets. However the brackets can be omitted in contexts
where a type is expected and the type description fits in one symbol.

Type alternatives represent type choices and are otherwise used internally for resolution. Nullable
types are most common form of type alternative and have a special prefix notation using `?`:

      (eq <alt int none> int?)

Type variables, references and selections use the at-sign '@' to separate kind from identifier.

      `@`       type variable with auto id
      `@1`      type variable with id 1
      `num@1`   type variable with constraint num
      `@ref`    type reference
      `num@ref` named num type
      `.sel`    local type selection

Variable and reference types must be instantiated by the type system by calling Sys.Inst, that means
variables get fresh ids, and references and selections are resolved.

Types with element bodies are frequently used, we therefor provide a shorthand notation that joins
the type and element with the pipe '|' character.

      (eq list|int@1? <list <alt@1 int none>>)
      sym|typ|list|dict|int

Parameter types like obj can have optional fields where names may ends in a question mark '?'. This
means if the field has a zero value it can be omitted when serialized or sent to database. This is
orthogonal to whether the type is optional.

      (obj name?:str opt?:str? (<> explicit optional (pointer to) string))


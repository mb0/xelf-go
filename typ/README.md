xelf/typ
========

Xelf is a typed language. We use a kind bitset from package `knd` to describe many details.
All types can have a type id. Type variables use these ids internally for resolution.
Complex types need to store additional information in specialized type bodies.

Types are naturally enclosed in angular brackets. However the brackets can be omitted in contexts
where a type is expected and the type description fits in one symbol.

Type alternatives represent type choices and are otherwise used internally for resolution. Nullable
types are most common form of type alternative and have a special prefix notation using `?`:

      (eq <alt int none> int?)

Type variables, references and selections use the at-sign '@' to separate kind from reference.

      `@`      type variable with auto id
      `@1`     type variable with id 1
      `num@`   type variable with constraint num
      `@ref`   type reference
      `.sel`   local type selection

Types with element bodies are frequently used, we therefor provide a shorthand notation that joins
the type and element with the pipe '|' character.

      (eq list|int@1? <list <alt@1 int none>>)
      sym|typ|list|dict|int

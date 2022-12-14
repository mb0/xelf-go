Ideas
=====

Note half-baked and smaller ideas tracked here for comments and later review.

Problem
-------

Think about restricting the environment for nested expressions. We could use that in daql/qry.
 * We separated the built-ins in core and decl, but we need a mechanism to restrict built-ins.
 * Currently it is easy to declare things in the environment but harder to hide things.
 * Maybe hiding built-ins is not even the best approach.
 * If we want to add special unqualified modules we could use them for built-ins too. Then it would
   at least be easier to filter references by origin module. This would also leave the program root
   environments solely to language extensions.

We want to use the cat and other specs both with individual arguments or a compatible list. We need
a syntax that discerns between lists used as element or as fill-in for a variadic tuple argument.
 * We could add a use make to type a list as tupl|lit and a dict as tupl|tag|lit
 * `(sep '-' (tupl (range 12)))` looks alright to me

Dicts are more or less `<list|obj key:str val:any>`. If we introduce a named key val obj type into
the core type system, we could allow conversion between `dict ` and `list|@keyval`, and promote dict
not only to a real idxr but to an appender as well.

Implementation
--------------

We combined the two variable declaration specs `with` and `let` into a new `with` spec, that allows
both old variants and combined declarations. The new decl environment always split a dot on lookup,
even if the first argument is a tag, to avoid confusion. The new spec allows chaining of expressions
in a more declarative and traditional way:

	`(with input ast:(scan .) exp:(parse ast) (run exp))`

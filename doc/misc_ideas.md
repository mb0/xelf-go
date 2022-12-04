
Ideas
=====

Note half-backed ideas for later review.


Unify the variable declaration specs `with` and `let`.
 * with is the better name, because it better implies the limited scope from use in other languages
 * let might imply the variable is registered in the parent env
 * with must have the dot value or a tag as first value, may have additional tags and an action.
 * we can chain operations for readability `(with input ast:(scan .) exp:(parse ast) res:(run exp))`
 * do we always split a dot on lookup, even if the first argument is a tag?
   * it would be confusing when changing to a tag and changing nested dot symbol paths with it


Think about restricting the environment for nested expressions. We could use that in daql/qry.
 * We separated the built-ins in core and decl, but we need a mechanism to restrict built-ins.
 * Currently it is easy to declare things in the environment but harder to hide things.
 * Maybe hiding built-ins is not even the best approach.
 * If we want to add special unqualified modules we could use them for built-ins too. Then it would
   at least be easier to filter references by origin module. This would also leave the program root
   environments solely to language extensions.


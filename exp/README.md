xelf/exp
========

Xelf is a simple but typed expression language with LISP-like parenthesis enclosed prefix calls.
Expressions are one of `Lit`, `Sym`, `Tag`, `Tupl` or `Call` and implement the `Exp` interface.

 * literals values are JSON-compatible as defined by package lit and can also hold types and specs
 * symbols can contain alphanumeric identifiers and ASCII punctuations
 * tags allow str, sym and num keys (`'str key':val`, `k:v`) and a special short variant (`flag;`)
 * tuples are lists of expressions used for documents roots and form specs

A `Prog` is used to resolve and evaluate an expression using a root environment. Program resolution
has two or more phases. In the first phases we call `Resl` methods to resolve all types. In the last
phase we call `Eval` methods to evaluate an expression to a literal.

The program itself mostly manages type checking and otherwise calls out to environments and specs.

A `Env` configures the program evaluation by defining all symbols, importantly these symbols can
resolve to specs that can be called and themselves create scoped environments that define symbols.
Calls and symbols cache their environment.

A `Spec` is a func or form definition that resolves and evaluates calls. If the first element of a
call does not resolve to a spec literal the program calls a `dyn` spec to allow syntax sugar.
The `dyn` spec is evaluated from the root environment on init.

Specs have a declaration type. When we resolves a call spec we instantiate the spec declaration
and call `LayoutSpec` to match and group all arguments to spec parameters.

Func specs have a strict and familiar parameter matching with support for named parameters and
variadic last argument. Func parameters are evaluated before the call is evaluated, therefor funcs
can also not change environment.

Form specs can mix in tupl parameters to match sequences of expressions, tags and even repeating
groups of parameters. Forms can also use the exp type in their signature to signal the resolution
not to resolve the argument and leave it to the spec. Form argument resolution and evaluation is
responsibility of the spec.

The `SpecBase` type is a partial spec implementation that implements automatic argument resolution
based on the type signature. Final implementations mostly need to handle call evaluation, but more
involved specs can implement custom call resolution too, for example to set nested scopes.

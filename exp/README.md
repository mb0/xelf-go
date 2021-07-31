xelf/exp
========

Xelf is a simple but typed expression language with LISP-like parenthesis enclosed prefix calls.
Expressions are one of `Lit`, `Sym`, `Tag`, `Tupl` or `Call` and implement the `Exp` interface.

 * literals values are JSON-compatible as defined by package lit and can also hold types and specs
 * symbols can contain alphanumeric identifiers and ASCII punctuations
 * tags allow str and sym keys (`'str key':val`, `key:val`) and a special short variant (`key;`)
 * tuples are lists of expressions used for documents roots and form specs

A `Prog` is used to resolve and evaluate an expression using a root environment. Program resolution
has two or more phases. In the first phases we call `Resl` methods to resolve all types. In the last
phase we call `Eval` methods to evaluate an expression to a literal.

A `Env` configures the program evaluation by defining all symbols, importantly these symbols can
resolve to specs that can be called and themselves create scoped environment defining symbols.
Calls and symbols cache their environment during the first resolution phase.

A `Spec` is a func or form definition that resolves and evaluates calls. If the first element of a
call does not resolve to a spec literal the program calls a special dyn spec from environment to
allow syntactic sugar.

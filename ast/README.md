xelf/ast
========

Xelf needs a simple and minimal syntax that supports JSON based literals, types and expressions.

Accepting plain JSON as valid literals makes it very simple to interoperate with existing code,
databases, and user assumptions. We already predefined JSON tokens:

 * null, true, false as symbols
 * str  uses `""`
 * list uses `[]`
 * dict uses `{}`
 * tags uses `:`

Xelf literals use a superset of JSON with additional rules to make handwriting literals easier:
 * str can use single quotes. Single quotes str literals are the default xelf format,
   because xelf is often used inside strings in other languages
 * char can use backtick quoted multiline string literals without escape sequence. This is
   helpful for templates and other pre-formatted multiline strings.
 * the comma separator is optional, whitespaces work just as well and don't clutter the output
 * tag keys do not need to be quoted if they are simple names as defined by the cor package
 * short flag tags can omit the value `(eq (obj flag;) (obj flag:true))`

We need to fit in types and expressions and only have parenthesis and angle brackets left:
 * exp uses `()` because it is familiar to write, needs less escaping and looks like a lisp.
   exp borrows tag notation form dict literals `(let a:1 b:2 (add a b))`
 * typ uses `<>` because we don't need it much and can mostly inferred in problematic contexts
   `<list|obj id:int name:str score:real date:time>`

Composite literals can only contain other literals, but no symbols or calls. Forms can be used
to construct literals from expressions, instead of reusing the literal syntax. This makes it
visually more obvious whether something is a literal, type or expression.

We use a token lexer and as scanner to scan an abstract syntax tree that can be used by the lit, typ
and exp packages to parse language elements.

The Ast stores the source position and optionally a input source name. Line based positions are used
because it is more likely to stay correct after small changes and more meaningful to a human when
printed.

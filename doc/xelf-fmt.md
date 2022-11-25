Xelf Fmt
========

We want to provide an authoritative standard formatting for program files, similar to go fmt.

Formatting in this context primarily means printing expressions for better readability, by deciding
where to add or remove white-spaces, specifically line-breaks, tabs and spaces.

Problem
-------

We use xelf source files to define project schemas and document layouts. Automatic formatting would
help when editing these source files.

Xelf is an expression language built around lots of calls to named specs and each spec has different
formatting requirements.

Implementation
--------------

One thing is certain: indentation should always use tabs for various reasons.
Parenthesis and brackets should end on the same indentation level as the start.

Rules:
 * indentation: marked by '\n\t'
 * hard line break: always applies maybe marked by ',\n'
 * soft break: may apply based on source context marked by '\n'
 * explicit param names might help readability for more involved specs like swt

Discussion
----------

We should identify a set of formatting rules that cover our requirements.

We need to resolve spec types to decide how to format calls in any way. Using naive spec name maps
to formatting rules does not handle aliases, name shadowing and is also cumbersome to implement.

That means we need to resolve all spec types before we can print it, this would change some
expressions, that would be a problem for editor input. Maybe we can write a custom resolver?

We could try to somehow stuff the formatting rules into the type signature. It should not be part
of the actual type value, but if we could use the same source ast to configure formatting rules, we
would have complete formatting rule coverage for every spec. We could use the ignored white-spaces
and commas in the type source to hint at formatting rules.

We already have a printing system using `Print(*bfr.P)` for all expressions that handles differences
between json and xelf formatting. We probably want to have access to resolved type information
in the printer, so we need to use an abstract printer or provide a formatting hook into the
printer system.

Example
-------

`<form@swt @1
	<tupl case!:@1 then:exp|@2>
	else!:exp?|@2
@2>` and `(swt $num 1'one' 2'two' 'more')`

Formats the soft breaks based on source context to:

`(swt $num case:1 'one' case:2 'two' else:'more')`

When we trigger one soft break, we break all soft breaks for that call:

`(swt $num
	case:1 'one'
	case:2 'two'
	else:'more'
)`

`<form@project name:sym,
	tags:tupl?|exp,
@>` and `(project site auth.dom blog.dom)`

`(project site
	auth.dom
	blog.dom
)`

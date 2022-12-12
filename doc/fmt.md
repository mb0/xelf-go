Xelf Fmt
========

We want to provide an authoritative standard formatting for program files, similar to go fmt.

Formatting in this context primarily means printing expressions for better readability, by deciding
where to add or remove white-spaces, specifically line-breaks, tabs and spaces.

Problem
-------

We use xelf source files to define project schemas and document layouts. Automatic formatting would
help when editing these source files.

Xelf is an expression language built around lots of calls to named specs and each spec may have
different formatting requirements.

Implementation
--------------

One thing is certain: the indentation level should always use tabs for various reasons. In addition
to indentation we might want to use positioning using spaces. Parenthesis and brackets should end on
the same indentation level as they start, this may help detect the specific culprit in input with
unmatched pairs.

Even a simple generic formatter needs some state for the indentation level, max length and current
output position, it must be easily copied and passed to subsequent levels. A simple default
formatter can also provide basic configuration and some optional flags. Specialized formatters
usually each use their own context customized to their needs.

To know the current output writer source position, is not as easy as it sounds. We can either
wrap a writer and check everything written for linebreaks, this might have some overhead but
provides reliable positions and is easier to maintain. Or we track what we have written in the
formatter itself.

Discussion
----------

The default format for serializing expressions uses no breaks at all and only basic spacing.
(Not minimal spacing though, we could theoretically omit spaces between simple delimiting tokens
like so: `(test'a'1(b))`. But this requires checking the last input rune to decide and is not
very readable at all. If we want to compress xelf input, we could simply use gzip compression and
not care about the extra space.

We could first try how far we get with a generic formatter. The generic formatter should only use
ast info and tries to find a balance between abstract metrics like number of calls in one line and a
max line length. Even if the result is not great we can use it as a fall-back for more specific
formatters. It is good to check other generic
[lisp formatting rules](https://stackoverflow.com/questions/36079915/how-to-format-lisp-code).

For more nuanced custom formatting rules, we would need to resolve spec types to lookup the rules.
Using naive spec name maps to formatting rules does not handle aliases and name shadowing and is
also cumbersome to implement. That means we need to resolve all spec types before we can print
calls, this would change some expressions, that would be a problem for editor input. Maybe we can
write a custom resolver?

We already have a printing system using `Print(*bfr.P)` for all expressions that handles differences
between json and xelf formatting. We probably want to have access to resolved type information
in the printer, so we need to use an abstract printer or provide a formatting hook into the
printer system.

A xelf fmt command must always reproduce the input if encountering errors. A default mode used
when saving editor buffers should usually not rewrite any tokens and only edit white-spaces. However
we may want to repair unmatched or superfluous parenthesis and other pairs if possible.

We could try to somehow stuff the formatting rules into the type signature. It should not be part
of the actual type value, but if we could use the same source ast to configure formatting rules, we
would have complete formatting rule coverage for every spec. We could use the ignored white-spaces
and commas in the type source to hint at formatting rules.

We should identify a set of formatting rules that cover our requirements. Possible rules:

 * indentation: marked by '\n\t'
 * maybe positioning: marked by ' +'
 * hard breaks: always applies maybe marked by '\n'
 * soft breaks: may apply based on source context marked by ',+\n'
 * forced param names might help readability for more involved specs like swt

Examples
--------
```
<form@swt @1,
 <tupl case!:@1 then:exp|@2>,
 else!:exp?|@2,
@2>
(swt $num 1'one' 2'two' 'more')

(<> Formats the soft breaks based on source context to:)
(swt $num case:1 'one' case:2 'two' else:'more')

(<> When we trigger one soft break, we break all soft breaks for that call:)
(swt $num
 case:1 'one'
 case:2 'two'
 else:'more'
)

<form@project name:sym
 tags:tupl?|exp
@>
(project site auth.dom blog.dom)

(project site
 auth.dom
 blog.dom
)

(<> maybe we indent by a lvl+ws and omit forced params if in the same scope and
    lvl+1 when in nested scope. or based on whether the new line starts with a '('
)
(try (cache.get $id)
 else:(let res:(calc $) (if (not (eq err "not found")) res
  else:(try (cache.set $id res) res)
 ))
)
(try (cache.get $id) err:(let res:(calc $)
	(if (not (eq err "not found")) res
	 else:(try (cache.set $id res) res)
	)
))
```

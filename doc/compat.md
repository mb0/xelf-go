Compatibility
=============

Xelf as a format and language framework should be versatile enough to be implemented for different
target platforms. Our main consideration are the host language targets go and js and the database
backends postgres and sqlite.

Database backends do not need to implement the full language, but it would be helpful if we can
translate a limited set of expressions to be run as part of the query.

One major pain point, that has no obvious answer, is with strings. Go uses utf-8 strings and usually
0-based byte indices, while javascript uses UCS-2 (similar to utf16) strings and indices accordingly
and postgres uses 1-based utf-8 encoded strings. Sqlite can be set to use utf-8 as well, but
must be evaluated in detail.

I am not really comfortable with accepting these differences. It is probably best if we can
use what go uses as runes, which should be the same as postgres with utf-8 encoding selected.
Only seldom used runes have two UCS-2 codepoints in javascript (emoticons for example).

Or we just document the difference for the string type and provide the conversion to the raw type
that then reveals comparable byte offsets using an array buffer and the encoding api for js.

I also stumbled upon that postgres has no last index function. For function that do not easily map
to corresponding target backend functions we have two or three options:
 * Drop it from the std lib to be consistent across targets
 * Use a suboptimal and contorted hack if possible
 * Use custom function or extensions for postgres and sqlite

I should explore postgres functions and sqlite extensions for last string index and compare it
to the contorted hacky version that can be found on the net as the most simple test.
The results will inform me about the process and the overhead costs. I thing to remember is that
relying on custom functions or extensions is a cost in itself and reduces backend compatibility,
we could dynamically check and add pql functions for postgres, but for sqlite we at least need
to use a driver that supports extensions and hook into the driver to register them.

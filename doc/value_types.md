Value types
===========

We want to correct literal value types to be the most generic kind for packages ast and lit.

Problem
-------

We don't use the most generic kind for literal values and rely to heavily on an expression context
for type annotations. For example we currently use lit.Str as default to the char value the literal
expression type to hold the more general char type. So if we parse json obj with a time or uuid
encoded as strings and pass the value to a program, then it cannot know the original type without
explicit annotations.

We did end up with a combination of:
 * bad hacks to generalize types at runtime boundaries
 * the need to explicitly provide types for literals
 * double checking them when working with literal values

We should use the most generic value type so we only have to worry about one direction. When we lose
the narrowed down expression type we should never be left with an incorrect type.

Implementation
--------------

We change ast tokens to use num, char, idxr, keyr again.

We introduce lit.Num and lit.Char to mirror lit.Int, lit.Str. The reason for num to be represented
as int is, that anything that discerns floats from ints makes the unambiguous floats, while any int
could also be float.

We introduce lit.Vals and lit.KeyVals, which are just dressed up []Val and []KeyVal respectively,
that are idxr and keyr values on their own and can be partially reused by list and dict.

History
-------

I originally started with the concept of most generic literal value type in mind, but got rid of it.
I was too impressed how expression literal types hide the problem in an context with known types.

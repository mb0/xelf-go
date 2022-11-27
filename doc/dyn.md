dynamic spec
============

We use the dyn spec to resolve syntax sugar calls that have a literal as first element.

We provide syntax sugar for following literals:

  * spec: call
  * type: make
  * num: add
  * str: cat
  * list: append
  * keyr: mut

The omissions are because all options would be confusing and are better served by the make sugar.

We load the dyn spec from the program environment. In a recent change we added a special env method
to return the dyn without allocating a new literal in the quest to make dyn spec reassignable in
nested environments. However, I now come to the conclusion that the sugar should either be a
reliable part of any dialect or completely dropped. The implicit behaviour change would be hard to
keep in mind and the implementation would be too involved, to justify the cost of that feature.

So we are back to loading the dyn spec on init from the program environment.

We could implement per type sugar specs to extend default sugar for custom types, these would be
independent of other language features and easier to compose. In practice the core sugar specs
are nice to have but only if a consistent part of all dialects.

The more essential feature that dyn provides is the late binding call specs for unresolved symbols.
We could add a call spec to the core builtins for that.

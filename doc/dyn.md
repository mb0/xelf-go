dynamic spec
============

We use the dyn spec to resolve syntax sugar calls that have a literal as first element.

We provide syntax sugar for following literals:

  * spec: call
  * type: make
  * num: add
  * str: cat
  * raw: json
  * list: append
  * keyr: mut

(We could add bool: and to the list, or remember the reason why not)

We load the dyn spec from the program environment. In a recent change we added a special env method
to return the dyn without allocating a new literal in the quest to make dyn spec reassignable in
nested environments. However, I now come the conclusion that the sugar should either be a reliable
part of any dialect or completely dropped. The implicit behaviour change would be hard to keep in
mind and the implementation would be too involved, to justify the cost of that feature.

We could implement per type sugar specs to extend default sugar for custom type, these would be
independent of other language features and easier to compose. In practice the core sugar specs
are nice to have but only if a consistent and more or less a fixed part of all dialects.

So we are back to loading the dyn spec on init from the program environment.


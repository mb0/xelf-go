mutate spec
===========

We want to revisit the mut spec and its role in the language and look for a unified way to express
all value mutations.

Problem
-------

Can we find a good spec definition to cover all our mutation needs to unify the language syntax?
Can we find a good data structure to represent any mutation that we can use for generalized deltas?
Both problems are highly related. Any generic mut spec should be able to apply those deltas.

Our mutation needs boil down to four distinct variants:

 * We want to **assign** one compatible value to any value
 * We want to **merge**  with compatible values for num, char, raw, str, time, span, list, keyr, obj
 * We want to **append** compatible element values to list and dict values
 * We want to **modify** compound idxr and keyr values

We want to **diff** values A, B and get a **delta** D that, when we **apply** D to A, results in B.

Discussion
----------

A **delta** must be able to represent all mutations, by only using meta data and new element values.
We probably want to use a dict as delta type to allow multiple, precise edits in one delta. Each
element in the diff represents an edit. We have an edit key and edit value to work with.

The simplest edit is to **assign** the edit value. We already have the concept of paths to select
into values and should reuse it for assign edit keys to select the target. This already covers our
requirements for assign and merge edits and more:

	{a:1 b:2} is valid delta to arrive at the same keyr value
	{.:1} the root path can be used to assing the whole value
	{.:{a:1 b:2}} and {.a:1 .b:2} can be simplified to {a:1 b:2}
	{.:{a:1} b:2} would require ordered edits which could be defined
	{.:[1 2 3] .1:3 .-1:7} we can select into any idxr from both ends
	{docs/secret:null} or even set the same value to a list selection

Another simple modify edit is to **delete** a key, it must be different than assigning it with null
to allow simple string sets with `<dict|none>`

	{a;} is equal {a:null} it uses short tags missing from json and cannot be used as indicator 
	{a;} for assigning null and {a-;} or {"a-":null} for deletion
	{docs/secret-;} remove any indications of your secrets
	string sets can use {foo;} to add and {foo-;} to delete keys

There is a problem with the path selection for dict keys that contain any special edit path chars.
We need an **alt-path-key** notation to put long or difficult key segments into the edit value.
This also makes sense for a simple `<dict|none>` sets, where we only have the key as data.

	{$.queryCounts.$:['postgres://mydb' 'my special key with lots of code' 1]}
	{characterNames.$:['$' 'dollar']} if we actually want to use the dollar as a key
	{mySet.$:['long key']} and delete with {.mySet.$-:['long key']}

This feature is probably best covered with intrusive path segment variables, that are implemented
for all path selections. We likely also want to support it somehow in the language. Maybe by adding
a dedicated spec that takes a path symbol and a tuple of variables.

We need syntax for appending, inserting or deleting elements from lists. We already use my diff
package which uses a very simple change representation using either retain, delete or insert. These
three operations can be represented as positive and negative number for retain and delete and a list
of elements for inserts. A sequence of these operation can represent any list operation. Without
operation the rest of the input is automatically retained by default.

	[] is an empty list of operations, [0 []] is a list of noops, both do nothing
	[1] retain first element, does nothing but move a cursor
	[1 -1] keep the first and delete the second
	[2 -3] keep the first two and delete the next three
	[2 ['a' 'b'] -6] keep two, insert a and b and then delete the next six
	[2 -6 ['a' 'b']] same effect different order
	[['a'] 2 ['b']] prepend a, keep the next two and replace two elements with b
	Most simple edits are even human readable with a bit of practice.

We need to mark the **list-ops** edit somehow to separate it from other edits and the alt-path-key
notation. We already use the minus suffix for deletion, we should use the star for modifications.

	{names*:[-5]} delete the first five names
	{names*:[2 -5]} delete the third upto and including seventh name
	{names*:[['a' 'b']]} prepend two names
	{names*:[2 -1 ['foo']]} would replace the third name with foo

Append is a common list mutation that is not easy to express in these terms. We need the length to
know where to insert. We can use an another plus marker to apply the list ops from the back to get
**mirror-ops**. Conceptually simply applied to a input in reverse.

	{names+:[['a' 'b']]} append a and b
	{names+:[-5]} delete the last five names
	{names+:[-1 ['a']]} replace the last name with a
	{names+:[1 -1 ['a']]} replace the next-to-last name with a

We can apply the concept of list-ops to **str-ops** and **raw-ops** as well. Both do not insert a
list of elements but runes and bytes respectively. Together combined with the alt-path-key and all
bells an whistles we get:

	{$/+:['names' ['!!!']]} to append three exclaimation marks to each element in list names

We want to add a flag to disable str and raw ops. And maybe only use it for longer strings. We might
want to add precautions to bound the diff allocations for very large values. There is also a problem
when we apply deltas to unresolved char typed values, we do not know from the delta whether it
represents raw or str ops.

We can use the star marker for a **nested delta**, to reduce path lookups and improving readability.
The plus marker makes no sense for deltas, we might allow both or raise an error for plus.

	{docs.cv*:{hobbies-; attr+:['serious']}} drop hobbies from the cv and add an serious attrib
	
This is already a great improvement on previous concepts. The nested deltas are readable. The
symmetry of the mirror ops is pretty. But now it gets outright crazy, buckle yourself! What if each
element value in list ops can itself be a delta again? **delta-ops** Mind blown!

We add another star marker and instead of inserting a string, bytes or elements, we apply the delta
to the element at the cursor. We will apply it to one element by default, but we can specify the
number of elements. If we only had a spare --- yes we do: use the dollar marker without conflicts.

	{docs**:[2 {secret-;}]}   drop the secret in the third doc
	{docs+*:[{secret-;}]}     drop the secret in the last doc
	{docs+*:[{$:2 secret-;}]} drop the secret in the last two docs

And on top all delta-ops can themselves have nested deltas, list-ops and delta-ops. Hurray!

I am very pleased with the overall result and the latest progress. Every aspect is clean, reasonable
but highly expressive in combination and covers more cases than expected using only a simple dict.

The remaining question for deltas is whether the dict must preserve order or not.

 * We can combine edits easily enough to be completely fine with unique edit keys.
 * Using ordered edits makes implementing and reasoning about deltas much easier.
 * Many json parsers however read into unordered hash maps by default.
 * The syntax should already be flexible enough to work around unordered deltas.
 * We should also be able to simplify most ordered deltas to order agnostic variants with hard work.
 * The only corner cases involve list selections that we could still somehow transform to delta-ops.

Without a clear conclusion we should use an ordered dict until we can better describe the problem.
Then we try to write a transformer to simplify deltas and remove order ambiguity.

Now to the specs: we want one spec that ideally handles all mutation variants. The basic signature
must be `<form@mut v:data? tupl?|expr _>` it must have a fist argument, any number of tag or value
arguments and return a result with the type of first argument. It would fill part of the role the
dyn spec had and like dyn could delegate to other specs based on input to provide a better structure
and provide ways to handle corner cases.

 * We should use the delta syntax in tag arguments for all input values. We can use nested tag
   syntax to apply delta dicts to the root value `(v .*:delta)`.
 * Add a new 'apply' spec to apply a delta dict directly, not using tags (maybe a diff spec too?)
 * We want to support simple merge and append syntax for str and list values specifically,
   but have a conflict with merge and append with one element and simple assign mutation `(v 1)`.
 * If we add a simple 'set' spec to explicitly assign one value we could clear up that conflict.
   We can always fall back on the delta syntax and use (v .:1) instead of (set v 1).
 * We would have have tag arguments exclusively to **modify** mutations covered by the delta syntax
   and plain arguments for merge and append mutations.
 * We prefer append for container types and add an explicit merge spec, to updates dicts and
   concatenate lists specifically. Again, both append mutations can be easily written as delta tag.

We want to add a sel spec `(sel path.with/$.var 'my key val')` to use path variables. This requires
some changes how we pass paths through the environment. Currently we pass in the symbol string
itself, that we can then change for parent environments. So we need to way to pass along the path
variables or parse the path once, fill vars if available, and lookup on the path.

Error Values
============

We want to consider error values for xelf and record the discussion.

Problem
-------

After reading the go user-defined iteration proposal, an attempt was made to convert the literal
iter api to the proposed push functions returning bool instead of error. Unfortunately proxies may
encounter errors themselves preparing the values, we should not panic or ignore errors. The only
elegant way to report an error is as value and break the iteration.

That got me generally thinking about error handling in xelf again. So if nothing else, this time I
wrote down my thoughts for next time. If we add error values we need to decide which type they
return, what they print as and how they fit into the picture overall.

Discussion
----------

We may want to add a special err kind and type for the err literal.

Is err a meta, exp, char kind? Is err kind special?

  * We should make it a char kind if it is treated as ordinary value.
  * If we provide a try form, we also want to collect err values in a list.
  * We should leave it a meta kind as long as we are undecided.

Do we use errors as a basic language concept or in corner cases?

 * The initial need for error values in push functions was itself decided against.

 * If we want to consider using it instead of error parameter in the go literal api we could compose
   some of the functions. Candidates are Idx, Key, maybe Select, Conv unlikely SetIdx, SetKey or
   Append. We would need to check `if err := v.(Err); err != nil {…}` instead, which is surprisingly
   fine.

 * The error value could be like NaN, that taints other values it interacts with until it is
   handled. That however excludes us from using err values as ordinary values anywhere. We could
   probably mark err values in contrast to err errors somehow. But do I think NaN is a great idea?

 * We could treat it as possible result from every call to allow error handling and recovery from
   within the xelf script. Whenever fail is called, we bubble an error value up the call stack
   ending the program evaluation, unless a try call handles the error. We could even handle
   program evaluation errors by the script itself.

 * Error values could also help with writing a runtime in simple host languages without multiple
   returns like xelf itself.

How to serialize errors? Do we receive it as input?

 * The simple solution would be to treat them as any other char value.
 * We often encounter error strings in result types for api endpoints.
 * If we send and receive errors we want to provide a mutable `*Err` and wrap native errors
   in proxies.
 * Err values should convert from and to Char and should be comparable by error string.


As for the original motivation to introduce the push function api: the code is ugly in all cases
involving errors, as long as the proposal is not accepted. The cases using errors include printing,
encoding and mutating values. We definitely keep the current api, it would be unreasonable not to.

The error values only help reporting preparation errors to the yield function, if the yield function
produces an error it must use a closure and careful variable declarations.

The push function api is only really better for simple find or contains use-cases. It can be
provided by wrapping the current iterator api that uses errors. The other way around is difficult.
Overall it is not valuable enough without any language support.

For now the strongest case for errors is to make error handling available to the code itself.
This would allow xelf scripts to handle more general tasks involving io errors without specialized
runtime support.

Where lit.Err would be a simple error wrapper that implements an error and literal value. It has the
err type with the meta kind err. It does not handle source position or other details but should
instead wrap an error that does.

And the error handling xelf api would look something like this:

	<form@fail tupl? err>
	<form@try @ <alt? _ tag|_> _>
	(try (cache.get $id) err:(let res:(calc $)
		(if (not (eq err "not found")) res
			else:(try (cache.set $id res) res)
		)
	))

Where fail raises an error with the concatenated arguments and try takes one exp to try and another
to evaluate with an err if the first failed. The second argument may be a tag to specify the err
value name.

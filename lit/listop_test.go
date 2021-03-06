package lit

import "testing"

// TestListOpSyntax is an experiment to structure my exploration and discussion document that
// I normally edit to pieces and discard as informational test.
// This kind of information does not really fit into doc or commit comments.
func TestListOpSyntax(t *testing.T) {
	if testing.Short() || !testing.Verbose() { // this not ready running any code so leave it
		return
	}
	isItGoodYet := false

	// We already have selection paths '.a/b.-1.5' and we have set and create path that
	// use them to update values, but do not support list selection a/b yet.
	// We use dicts of paths to values. Keyer values are also valid edit paths.
	t.Logf("{a:1 b:2} is a keyr value and also a valid delta to arrive at same value")

	// Other values can use a root edit path to update the root itself.
	t.Logf("{'.':true} updates the target value to true")

	// We usually want to start edit paths with a dot to make it clear it is a path.
	// But for single key paths we drop the dot in favor of the simpler keyr notation.
	t.Logf("{'.a':1 '.b':2} is printed as {a:1 b:2}")

	// We already support idx path segments including negative ones.
	t.Logf("{'.-1.2':1} update the last lists third value to 1")

	// This is all well but misses some important edits we cannot currently express.
	//  * We need the options to explicitly delete keys from dicts, not only set them to null.
	//  * We need the option to delete and insert multiple elements from lists.

	// We use the prefix to detect paths and preferably want to keep using paths as they are.
	// We also obviously express everything in plain literals. That means we only really can
	// add a suffix to the path as syntax that we chop of before selecting the element.

	// Delete paths for keyrs simply add a minus sign to indicate that we want to clear the key.
	t.Logf("{'.a-':null} looks similar to {'.a':null} both can use short tags {'.a-;'}")

	// But updating lists is harder, we need syntax to insert and drop multiple ranges.
	// We can use my diff package to compare values and find a good edit path.
	// We use that to generate a list of operations built from plain literals and insert lists.
	t.Logf("[2 -1 ['a' 'b']] should replace the third element with 'a' and 'b'")

	// In the general case we have multiple edits that are order dependent, but we want to save
	// them to an unordered dict. That means we need to use a list to store edits in order.
	// Because we use '-' as deletions suffix we probably want '*' for general editing and keep
	// '+' in mind for special insert syntax.
	t.Logf("{'.a*':[2 -1 ['a' 'b']]} should update {a:[1 2 3 4]} to {a:[1 2 'a' 'b' 4]}")

	// Most simple edits are flexible and great with this syntax to be human readable.
	t.Logf("{'.a*':[[7 8]]} prepend: {a:[1 2 3]} {a:[7 8 1 2 3]}")
	t.Logf("{'.a*':[1 [7 8]]} insert at: {a:[1 2 3]} {a:[1 7 8 2 3]}")
	t.Logf("{'.a*':[-2]} delete head: {a:[1 2 3]} {a:[3]}")
	t.Logf("{'.a*':[1 -2]} delete at: {a:[1 2 3]} {a:[1]}")
	t.Logf("{'.a*':[1 -1 [7 8]]} replace at: {a:[1 2 3]} to {a:[1 7 8 3]}")

	// The most common case is append and it requires the specific length which is not great.
	t.Logf("{'.a*':[3 [7 8]]} should update {a:[1 2 3]} to {a:[1 2 3 7 8]}")

	// We use a negative idx in selections to refer to the last element. We cannot use negative
	// index in list opts because they already represent deletions. We could use '$' to indicate
	// the next to last or end index. That makes ops more complicated and has potential issues
	// if we decide to extend deltas to str and raw values.
	t.Logf("{'.a*':['$' [7 8]]} could update {a:[1 2 3]} to {a:[1 2 3 7 8]}")

	// I though about a whole set of special list op syntax that is surely to complex to support
	t.Logf("{'.a.$+':[???]} would also imply {'.a.0+':[???]} and {'.a.3+':[???]} should also work")
	t.Logf("{'.a.3*':[???]} and {'.a.3-':2} would fit into that system")
	t.Logf("{'.a.-0+':[???]} instead of {'.a.$+':[???]} lets not get crazy!")

	// So maybe the ops edits are good enough and we actually want only one special for append
	t.Logf("{'.a+':[7 8]} should update {a:[1 2 3]} to {a:[1 2 3 7 8]}")

	// As an overview over planned syntax:
	t.Logf("With {a:[1 2 3 4 5]}")
	t.Logf("{'.a':[7 8]} sets the field {a:[7 8]}")
	t.Logf("{'.a-';} deletes the field  {}")
	t.Logf("{'.a+':[7 8]} appends to a  {a:[1 2 3 4 5 7 8]}")
	t.Logf("{'.a*':[2 -1 [7 8]]} edits  {a:[1 2 7 8 4 5]}")

	isItGoodYet = true

	if !isItGoodYet {
		t.Errorf("no good syntax found for list edit paths")
	}

	doesItHandleNestedEdits := false
	// We want to allow nested edits. This would make it more convenient for a human to write
	// deltas that have multiple edit paths into the same keyr and also easier for the computer
	// to avoid duplicate field lookups.
	t.Logf("With {a:{b:[1 2] c:{x:3 y:4} d:5}}")

	// We can already set a path to an object with this syntax:
	t.Logf("{'.a.c':{z:5}} results in {a:{b:[3 2] c:{z:5} d:5}}")

	// We do not however want to assign but instead edit the object.
	// We would repeat the path prefix for complex edits:
	t.Logf("{'.a.b.0':3 '.a.d':4} results in {a:{b:[3 2] c:{x:3 y:4} d:4}}")

	// To indicate a sub edits we need a prefix to indicate that we want to edit with and not
	// set to the following value, just like with list ops. That said we should use the same
	// suffix as for list edits.
	t.Logf("{'.a*':{'.b.0':3 d:4}} results in {a:{b:[3 2] c:{x:3 y:4} d:4}}")

	// With sub edits and the fact that each keyer by itself is also valid edit to itself.
	// We can use that syntax to merge keyrs.
	t.Logf("{'.a.c*':{x:1 z:2}} {a:{b:[1 2] c:{x:1 y:4 z:2} d:4}}")

	// So we have the star suffix followed by a list as listop edit and followed by a keyr
	// as sub edit. We have the minus suffix to delete list ranges and to delete keyr fields.
	// But lists have the plus suffix for appends, that is not currently mapped to keyrs.
	// We could leave it at that, make it also indicate a sub edit or come up with a difference
	// variant of sub edits. List append is always followed by a list so we could easily discern
	// both syntax variants. But what would be similar to appending for a keyr?
	// It could mean do not change non nil fields, but this is unintuitive and hard to use.
	// Or maybe only make a distinction only for dicts to not replace existing value but rather
	// delete them and append key and new value at the end:
	t.Logf("{'.a.c+':{x:1 z:2}} {a:{b:[1 2] c:{y:4 x:1 z:2} d:4}}")

	// We also want to allow sub edits in list ops.
	t.Logf("With {a:[[1 2] [3 4] 5]}")

	// we can start by creating edits for the single replacements we already detect.
	t.Logf("{'.a.0.0':3} sets the field {a:[[3 2] [2 4] 5}")

	// If we have only matching inserts and deletions (replacements) we can form idx paths for
	// all individual elements and create a delta.
	t.Logf("{'.a.0.0':3 .a.1.1:7} sets the field {a:[[3 2] [2 7] 5}")

	// We could use a sub edit for the common prefix.
	t.Logf("{'.a.*':{.0.0':3 .1.1:7}} sets the field {a:[[3 2] [2 7] 5}")

	// we could also check all neighbouring inserts and deletions for overlaps but this
	// is soon going to spiral out of control.

	if !doesItHandleNestedEdits {
		// maybe we just leave it at that?
		// t.Errorf("no good nested edit syntax found for list edit paths")
	}
}

# Akha language v1 - spec

sources: `spec_def.ak`
- the syntax shown in `spec_def.ak` is not repeated here
- this file keeps the rules, limits and errors the example cannot show

## overview
# purpose, goals, non goals
 - purpose: a flow step language, each step produces a result or moves to the next step
   - steps have events, the script has event observers on the exec bus
     - ⚠️ unknown: event definitions, assumed to come with macros
 - goals
   - easy to read and write, few keywords
   - data transfer and move
   - max validation at parse and check level, before execution
   - simple AI calls with min code (`AI.get<funcs>`), later versions
   - loosely typed, containers hold mixed values
   - predictable: no implicit conversions, a clear error instead of a silent coercion
   - every value is a Go object, an operation works when the value implements the operation interface
 - non goals in v1: builtins (`len`, `str` ...), functions, classes, generics, type annotations, pattern matching

## lexical structure
# source, lines, statements
 - ascii only sources, checked by the lexer, an encoder can be added later
 - `\r\n` is normalized to `\n` by the lexer
 - newline is the only statement terminator
   - `;` is a lex error with the hint "one statement per line"
 - lex errors report line and column (runes), AST `Pos` has the line only
 - `}` must be on its own line, except `} else {` and `} else if cond {`

# identifiers, keywords, reserved words
 - identifier: `[A-Za-z_][A-Za-z0-9_]*`, case sensitive
 - keywords and reserved words are case insensitive, `True` is `true`
 - keywords: `let var if else for in range break continue return exit and or not true false null`
 - reserved for later versions: `fn try catch`
 - keywords, reserved words and module names cannot be used as a name

# literals
 - number: no leading zeros (`007`), no hex or octal, `-0` prints as `0`
 - string
   - double quotes only, `'` is a lex error
   - escapes: `\n \t \r \" \\ \$`, an unknown escape is a lex error
   - no raw newline inside a string, no multiline strings in v1
   - `${ }` is a block scope inside the current block, v1 allows value expressions only
 - object: key is an identifier or a string literal, computed keys are v2
   - a duplicate key in a literal is a static error

# operators
 - a single `&` or `|` is a lex error
 - `+= -= *= /= %=` are lex errors

## types
# kinds
 - number: float64, integral values print without `.0`, `3 == 3.0` is true
   - `NaN` and `inf` cannot be produced, those operations are runtime errors
 - string: immutable text, not a char array, no index access
 - array: zero based, no negative indexes, out of range read is a runtime error
 - object: string keys, insertion ordered, objects and maps are the same
   - `obj[k]` needs a string `k`
 - null: the absence of a value, not an object or array
 - each kind is a Go type behind `Value`, operations are interfaces (`Adder`, `Comparer`, `Iterable` ...)
   - every type implements `Value` and `Stringer`

# truthiness
 - falsy: `false`, `null`, `0`, `""`, `[]`, `{}`, everything else is truthy

# conversions
 - no implicit conversion in operators, `"a" + 1` and `"2" * 3` are runtime errors
 - the only implicit conversion is to string inside `${ }`
   - array and object print as compact JSON
 - explicit conversions: v2

# equality and ordering
 - `==` / `!=` never error and never convert, different kinds are not equal
   - arrays and objects compare by value, object key order is ignored
 - `< <= > >=` only number with number or string with string (code point order), else a runtime error
 - comparisons do not chain, `a < b < c` is a parse error

## variables and scopes
 - `let` is deep immutable: no reassignment, no index or member assignment
 - `var` reassignment can change the kind only from `null`
 - values are copied on assignment and when passed (copy on write in Go)
 - lookup goes from the current scope to the root
 - use before declaration, undeclared names and redeclare in the same block are static errors
 - loop variables live in the loop scope, immutable unless `for var`, `for let` is a parse error
 - assigning to a `let`, an immutable loop variable or `input` is a static error
 - assignment is a statement, array index assignment must be in range
 - one declaration per statement, no `let a, b = ...`

## expressions
 - `not a == b` is `not (a == b)`, `-a[0]` is `-(a[0])`
 - `in` and `not in` share one non associative level
 - division or modulo by zero is a runtime error
 - `+` on numbers, string + string, array + array, objects is a runtime error (merge is v2)
 - `and` / `or` short circuit and return a boolean
 - `in`: value in array, substring in string, key in object, other pairs are a runtime error
 - `arr[i]`: `i` must be integral
 - `.` is the object member operator only
   - on a number, string, boolean, null or array literal it is a parse error `member-kind`
   - on any other non object value it is a runtime error
 - slices `arr[1:3]` are parse errors in v1
 - only module functions are callable
   - a positional arg after a named arg is a parse error `kwarg-order`
   - a duplicate named arg is a parse error `duplicate-kwarg`

## statements
 - a bare non call expression is a static error "unused value"
 - `if`: any number of `else if`, at most one `else`, empty blocks are allowed
 - `for in` on a non array non object is a runtime error
 - the loop iterates a snapshot of the collection
 - `range` takes one loop variable, bounds must be integral, `start >= end` runs zero times
 - `break` / `continue` outside a loop is a static error, no labels in v1
 - `return` / `exit` leave every loop and end the script, `return` alone outputs `null`
 - statements after `return` in the same block are a static error "unreachable"
 - no error handling in v1, a runtime error fails the step

## modules
# definitions
 - a module is a named set of Go functions (`lang/module` interfaces)
 - positional params are required, the arg count must match
 - named params must have a default, registering one without fails
 - every param has a type, checked statically when the value is known and at runtime
 - a duplicate module or function name fails on register
 - module functions return `null` when there is nothing to return

# core module `ak`
 - `ak.import(name)`: string literal, root block only, twice is a static error `import-dup`
 - `ak.log(msg)` / `ak.debug(msg)` write to the worker logger at info / debug level
 - ⚠️ unknown: what `setup(log, debug)` configures, it is a no op returning `null`

# fs module
 - sandboxed root (`os.Root`), opened once on the first call
 - `/` and `\` are separators, a leading `/` is root relative, `.` and `..` are cleaned
 - leaving the root, empty paths, null bytes and volume names (`C:`) are runtime errors
 - `read`, `write` (append, creates parent dirs), `overwrite`, `exists`
 - `readJSON` / `readYAML` keep the file key order
 - `writeJSON` / `writeYAML` overwrite with a 2 space indent
 - yaml anchors and aliases are resolved, timestamps become strings
 - ⚠️ unknown: the default root is `.` until the worker `root` config is passed in
 - ⚠️ unknown: yaml merge keys `<<` are kept as a plain `<<` key

# loading and checks
 - parse: unknown function, arity, unknown named arg, known arg type
 - check: unknown module, not imported, import rules, known arg type through `let`
 - runtime: args bound to params, defaults filled, types validated before the call
 - a call on a declared name (`obj.fn()`) is a static error `no-callable`
 - ⚠️ unknown: must imports come before every other statement

## errors
 - lex, parse and check stop on the first error, eval does not start
 - static format: `file.ak:line:col: error[code]: message`, optional `hint:` line
 - runtime format: `step.ak:12: runtime error: cannot add string and number`
   - line only, 1 based, carries the kinds involved
 - runtime cases: kind mismatch, ordering across kinds, division by zero, index out of range,
   non integral index or bound, access on `null`, non iterable, limits exceeded

## execution model
 - lex the whole source, parse (recursive descent), check, eval
 - eval walks statements in order, blocks push and pop scopes
 - the context is checked before each statement and loop iteration, cancel stops the step
 - worker limits: max statements, max loop iterations, max value size
 - `input` is any JSON compatible kind, the output is the `return` value
 - values crossing steps are JSON compatible

# Akha language v1 - spec

sources: `task.ai_spec.md`, `__arch/v3/spec_ded.md`, `__arch/v3/spec_def.ak`

## overview
# purpose, goals, non goals
 - purpose: a flow step language, a script a flow os steps each step produces a result or moves to next step
 - goals
   - easy to read and write: 
     - one statement per line 
       - statement is a step that produces a result or moves to next step
     - braces for blocks, 
     - few keywords
   - loosely typed: variables have no declared type, containers hold mixed values
     - re-assignment is allowed for `var` variables, not for `let` variables
   - predictable: no implicit conversions between kinds
     - a clear error instead of a silent coercion that can lose data
   - every value is a Go object, an operation works when the value implements the operation interface
   - one `number` kind, no int/float split
     - all numbers are float64
       - printed as int when integral (no decimal part), as float when not
 - non goals for v1
   - user defined functions (v2, `fn` is reserved)
   - error handling inside the script (`try`/`catch` is reserved)
     - errors are thrown and break the execution of the step
   - modules and builtins (`Ak`, `FS`, `len`, `str` ...), later versions
   - classes, generics, type annotations, pattern matching

## lexical structure
# source encoding, lines, statement terminator
 - source is UTF-8, a BOM is ignored, invalid UTF-8 is a lex error
 - `\r\n` is normalized to `\n`
 - newline is the only statement terminator
   - `;` is not a token, using it is a lex error with the hint "one statement per line"
 - one statement per line, a blank line is ignored
 - a statement continues across lines only inside an open `(` `[` or an object literal `{`
   - newlines inside those brackets are ignored, a trailing `,` before the closing bracket is allowed
   - long calls, arrays and objects are split this way
 - indentation has no meaning, it is for readability only
   - it's not included in the ast, it's not important
   - lexer only cares about position 
     - lexical errors report line and column, columns count characters (runes)
     - AST types have a `Pos` field with line  only, column is not needed
 - a block `{` must be on the header line, `}` must be on its own line, except `} else {` and `} else if cond {`

# comments
 - `// ...` line comment, runs to the end of the line, allowed after a statement but cannot be terminated

# identifiers, keywords, reserved words
 - identifier: `[A-Za-z_][A-Za-z0-9_]*`, ASCII only
 - case-sensitive for identifiers and keywords, `True` is an identifier, not `true`
 - `_` alone is a valid name that discards the value (eg `let _ = expr`)
   - only executes the expression, does not bind the value to a name
   - for expressions that have side effects but are not needed
 - keywords
   - `let var if else for in range break continue return exit`
   - `and or not`
   - `true false null`
 - reserved for later versions: `fn try catch`
 - keywords and reserved words cannot be used as a name

# literals: number, string, boolean, null, array, object
 - number
   - `0`, `42`, `3.2`, `1.32332`, `1.0`, `0.5`
   - no leading zeros (`007` is a lex error), no hex or octal
   - negative numbers are unary minus on a literal, `-3`
     - all numbers can be negative, `-3.2`, `-0.5`, `-0.1`
       - negative zero is `-0`, prints as `0` and used as a zero
 - string
   - double quotes only `"..."`, `'` is a lex error
   - escapes: `\n \t \r \" \\ \$`, an unknown escape is a lex error
   - interpolation `"hello ${name}"`, any expression inside `${ }`, the value is formatted with the string conversion rules
     - `${ }` is treated as a block with scope the current block the string is in
       - follows the nested block rules
     - for v1 only allow direct value expressions inside `${ }`
   - a literal `${` is written as `\${`
   - no multiline strings declarations in v1
   - a `"..."` string cannot contain a raw newline, use `\n`
 - boolean: `true`, `false`
 - null: `null`
 - array: `[1, "a", true]`, `[]`
   - mixed kinds allowed
 - object: `{name: "a", "content-type": "json", count: 3}`, `{}`
   - key is an identifier or a string literal
     - computed keys are v2
   - a duplicate key in a literal is a static error
   - keys keep insertion order
   - objects and maps are the same

# operators and punctuation
 - arithmetic: `+ - * / %`
 - comparison: `== != < <= > >=`
 - logical: `and or not`, aliases `&& || !`
   - a single `&` or `|` is a lex error
 - membership: `in`, `not in`
 - assignment: `=` - no fancy assignment, `+= -= *= /= %=` are lex errors
 - punctuation: `,` `:` `.` `..` `( )` `[ ]` `{ }`

## types
# number, string, boolean, array
 - number: float64, integral values are exact up to 2^53
   - `3 == 3.0` is true, both print as `3`, `3.2` prints as `3.2` (shortest form)
   - `NaN` and `inf` cannot be produced, operations that would produce them are runtime errors
 - string: immutable UTF-8 text
   - strings are not character arrays, they are a single value
     - no string index accessing
 - boolean: `true` / `false`
 - array: ordered, mixed kinds, zero based
   - no negative indexes
   - out of range read is a runtime error
 - each kind is a Go type behind a `Value` interface
   - operations are interfaces (`Adder`, `Comparer`, `Iterable` ...)
   - all types must implement `Value` and `Stringer` interfaces

# missing: object/map, null
 - object and maps are the same
   - string keys only, insertion ordered, mixed value kinds
   - `obj["key"]` read, reading a missing key returns `null`
     - keys are strings so they are treated the same way while parsing
       - eg: `obj["${name}"]` is a valid key, 
       - also `obj[name]` is a valid key, the value of `name` must be a string
   - `obj["key"] = v` replaces the values at the key
 - null is a value that represents the absence of a value
   - `null` is a literal, `null` is a value, `null` is not an object or array
   - `null == null` is true, `null != null` is false

# truthiness
 - falsy: `false`, `null`, `0`, `""`, `[]`, `{}`
 - everything else is truthy
 - applies to `if` and `for`

# conversions and coercion rules
 - no implicit conversion between kinds in operators
   - `"a" + 1` is a runtime error
   - `"2" * 3` is a runtime error
 - the only implicit conversion is to string inside `${ }` interpolation
   - number: shortest form, integral values without `.0`
   - boolean: `true` / `false`, null: `null`
   - array and object: compact JSON
 - explicit conversions: v2 with funcs and modules

# equality and ordering rules
 - `==` / `!=` never error and never convert
   - different kinds are not equal, `"1" == 1` is false
   - numbers by value, strings by content
   - arrays by value, same length and equal elements in order
   - objects by value, same keys and equal values, key order ignored
     - depending on the Compare interface implementation
       - but compared by value, not by reference
   - `null == null` is true
 - `< <= > >=`
   - number with number, numeric order
   - string with string, code point order
   - any other pair is a runtime error, `"a" < 1` is an error
 - comparisons do not chain, `a < b < c` is a parse error

## variables and scopes
# `let` immutable, `var` mutable
 - `let name = expr` declares an immutable binding, an initializer is required
 - `var name = expr` declares a mutable binding, `var name` alone starts as `null`
   - only null's can be reassigned as a type
 - `let` is deep immutable: no reassignment, no index or member assignment on the value
   - this means that no new member can be added to
   - also for arrays
 - values have value semantics: arrays and objects are copied on assignment and when passed (copy on write in the Go impl)
   - `var b = a` then `b[0] = 1` does not change `a`
   - a `var` copy of a `let` array is mutable but not

# block scope, lookup from current to root
 - every `{ }` block opens a new scope, the script is the root scope
 - lookup goes from the current scope to parents until the root
 - a name is visible from its declaration to the end of its block
 - use before declaration is a static error
 - an undeclared name is a static error

# shadowing rules
 - "can be overridden in nested blocks" means shadowing: a `let` or `var` in a nested block creates a new binding, 
   - the parent binding is untouched after the block
 - redeclaring a name in the same block is a static error
 - loop variables live in the loop scope and shadow outer names
   - spec_def.ak line 19 `for var i in arr3` shadows `let i = 0` from line 4, the outer `i` is `0` again after the loop
 - reserved words cannot be shadowed, `let if = 1` is a static error

# assignment and compound assignment
 - `name = expr` assigns to the nearest visible `var`, also a parent block `var`, no redeclare needed
   - but a var with the same name must be declared
 - `arr[i] = expr`, `obj.key = expr`, `obj["key"] = expr` on a `var` value
   - nested targets allowed (`data.items[0].name = "x"`)
 - assigning to a `let`, a loop variable declared without `var`, or a missing `var/let` is a static error
 - array index assignment must be in range
   - in V2, array's will have a `push` method to add elements, but not in v1
 - compound `+= -= *= /= %=` operators are not needed 
 - assignment is a statement, not an expression

## expressions
# precedence and associativity
 - lowest to highest
   - `or ||` - left
   - `and &&` - left
   - `not !` - unary prefix
   - `in`, `not in` - non-associative, one level
   - `== != < <= > >=` - non-associative
   - `+ -` - left
   - `* / %` - left
   - `-` - unary prefix
   - `.name` `[ ]` `( )` - postfix, left
 - grouping with `( )`
 - `not a == b` is `not (a == b)`

# arithmetic, concatenation, comparison, logical
 - `+ - * /` on numbers, `/` is float division
   - division by zero is a runtime error
 - `+` concatenates string with string and array with array, a new value is returned
 - `+` on objects is a runtime error in v1 (merge is v2)
 - `and` / `or` short circuit and return a boolean from the truthiness of the operands
 - `not` returns a boolean
 - `in`
   - value in array: any element `==` value
   - string in string: substring
   - string in object: key exists
   - other pairs are a runtime error

# index and member access
 - `arr[i]`: `i` must be an integral number, negative or out of range is a runtime error
 - `s[i]`: doesn't work on strings
 - `obj["key"]`: missing key returns `null`
   - non string index on an object is a runtime error
 - `.` on array, string, number, boolean is a runtime error in v1 (methods come with builtins)
 - slices `arr[1:3]` are parse errors in v1, they are v2

# calls, positional args, kwargs
 - v1 has no callables

# array and object literals
 - `[e1, e2, ...]`, elements are any expression
 - `{key: expr, "key": expr}`, values are any expression
   - keys are strings, or identifiers, identifiers are converted to strings for access
 - both can span lines and take a trailing `,`
 - an object literal cannot start an `if` / `for` header expression, wrap it in `( )`
   -  a `{` after a header expression always opens the block

## statements
# expression / call statement
 - a call can stand alone as a statement
 - any other bare expression (`1 + 2`, `x`) is a static error "unused value"

# `let`, `var`, assignment
 - `let name = expr`
 - `var name = expr`, `var name`
 - one declaration per statement, no `let a, b = ...`

# `if` / `else if` / `else`
 - `if cond {` ... `}`, the condition uses truthiness
 - `} else if cond {` and `} else {` must be on the `}` line
 - any number of `else if`, at most one `else`, last
 - empty blocks are allowed

# `for ... in`, `for ... range`
 - `for x in expr {`
   - array: elements in order
   - object: keys in insertion order
   - other kinds are a runtime error
 - `for i, v in expr {`
   - array: index, element
   - object: key, value
 - `_` discards a loop variable, `for _, v in arr {`
 - loop variables are declared fresh per iteration in the loop scope
   - `for x` / `for i, v` loop variables are immutable
   - `for var x` / `for var i, v` makes them mutable, reassigning does not change the collection or the next iteration
   - `for let x` is a parse error (redundant)
 - the loop iterates a snapshot, changing the collection in the body does not change the iteration
 - `for i range [start..end] {` - from `start` up to `end`, `end` excluded, step `1`
   - range is declared with `range` and `[start..end]`
   - range bounds are any expression, evaluated once, must be integral numbers
   - `start >= end` runs zero iterations
   - range takes a single loop variable

# `break`, `continue`
 - apply to the innermost loop
 - outside a loop is a static error
 - no labels in v1

# missing: `return` / exit, error handling
 - `return` ends the script with `null` as the step output
 - `return expr` ends the script with `expr` as the step output
 - allowed at any depth, it exits the whole script (no functions in v1)
 - statements after a `return` in the same block are a static error "unreachable"
 - error handling: none in v1, a runtime error stops the step and the step fails
 - exit is alias to `return`

## grammar (EBNF)
# notation
 - ISO style EBNF: `=` define, `|` choice, `[ ]` optional, `{ }` repeat, `( )` group, `;` end of rule
 - `"x"` terminal, `? ... ?` prose, `(* *)` comment
 - `NL` is a newline token, `EOF` the end of the source

# lexical grammar
```ebnf
source        = [ BOM ] { token | space | comment | NL } EOF ;
space         = " " | "\t" | "\r" ;
NL            = "\n" ;                                  (* "\r\n" is normalized to "\n" *)
comment       = "//" { ? any rune except NL ? } ;

letter        = "A" | ... | "Z" | "a" | ... | "z" | "_" ;
digit         = "0" | ... | "9" ;
nonzero       = "1" | ... | "9" ;

ident         = letter { letter | digit } ;             (* not a keyword or reserved word *)
keyword       = "let" | "var" | "if" | "else" | "for" | "in" | "range"
              | "break" | "continue" | "return" | "exit"
              | "and" | "or" | "not" | "true" | "false" | "null" ;
reserved      = "fn" | "try" | "catch" ;

number        = ( "0" | nonzero { digit } ) [ "." digit { digit } ] ;
                                                        (* no leading zeros, no exponent, not followed by a letter *)
string        = '"' { char | escape | interp } '"' ;
char          = ? any rune except '"', "\", NL and "$" followed by "{" ? ;
escape        = "\" ( "n" | "t" | "r" | '"' | "\" | "$" ) ;
interp        = "${" value_ref "}" ;
plain_string  = ? a string without interp ? ;

operator      = "+" | "-" | "*" | "/" | "%"
              | "==" | "!=" | "<" | "<=" | ">" | ">="
              | "=" | "&&" | "||" | "!" ;
punct         = "(" | ")" | "[" | "]" | "{" | "}" | "," | ":" | "." | ".." ;
```

# syntax grammar
```ebnf
script        = { line } EOF ;
line          = [ stmt ] ( NL | EOF ) ;
block         = "{" NL { line } "}" ;                   (* "{" on the header line, "}" on its own line *)

stmt          = let_stmt | var_stmt | if_stmt | for_stmt
              | break_stmt | continue_stmt | return_stmt
              | assign_stmt | call_stmt ;

let_stmt      = "let" ( ident | "_" ) "=" expr ;
var_stmt      = "var" ident [ "=" expr ] ;
assign_stmt   = target "=" expr ;
target        = ident { "." ident | "[" expr "]" } ;    (* ident is not "_" *)
call_stmt     = postfix_expr ;                          (* must end with a call, else "unused value" *)

if_stmt       = "if" header block [ "else" ( if_stmt | block ) ] ;
                                                        (* "else" on the same line as "}" *)
for_stmt      = "for" [ "var" ] loop_var ( [ "," loop_var ] "in" header | "range" range ) block ;
loop_var      = ident | "_" ;
range         = "[" expr ".." expr "]" ;
header        = expr ;                                  (* must not start with an object literal *)

break_stmt    = "break" ;                               (* inside a loop only *)
continue_stmt = "continue" ;                            (* inside a loop only *)
return_stmt   = ( "return" | "exit" ) [ expr ] ;

expr          = or_expr ;
or_expr       = and_expr { ( "or" | "||" ) and_expr } ;
and_expr      = not_expr { ( "and" | "&&" ) not_expr } ;
not_expr      = ( "not" | "!" ) not_expr | in_expr ;
in_expr       = cmp_expr [ ( "in" | "not" "in" ) cmp_expr ] ;
cmp_expr      = add_expr [ cmp_op add_expr ] ;
cmp_op        = "==" | "!=" | "<" | "<=" | ">" | ">=" ;
add_expr      = mul_expr { ( "+" | "-" ) mul_expr } ;
mul_expr      = unary_expr { ( "*" | "/" | "%" ) unary_expr } ;
unary_expr    = "-" unary_expr | postfix_expr ;
postfix_expr  = primary { "." ident | "[" expr "]" | "(" [ list ] ")" } ;

primary       = number | string | "true" | "false" | "null" | ident
              | array | object | "(" expr ")" ;
array         = "[" [ list ] "]" ;
object        = "{" [ entry { "," entry } [ "," ] ] "}" ;
entry         = ( ident | plain_string ) ":" expr ;     (* keys are unique *)
list          = expr { "," expr } [ "," ] ;

value_ref     = ident { "." ident | "[" ( number | plain_string | value_ref ) "]" } ;
```

# layout rules not expressed in the grammar
 - inside `( )`, `[ ]` and an object `{ }` newlines are ignored
 - a block `{` is never an object literal, an object literal is only parsed in operand position
 - `a < b < c` and `a in b in c` are parse errors, not left associative chains
 - `a[1:2]` is a parse error "slices are not supported in v1"

## v1 decisions
# resolved from the notes above
 - `&& || !` are aliases of `and or not`
 - `%` is part of v1, same level as `* /`
 - `in` and `not in` share one non-associative level
 - unary `-` binds tighter than `* / %`, looser than postfix, `-a[0]` is `-(a[0])`
 - negative indexes are a runtime error
 - loop variables are immutable unless declared with `for var`
 - range end is excluded, no step in v1
 - objects iterate in insertion order
 - lex runs over the whole source before parsing, the first error is reported
 - reported by the parser: `break` / `continue` outside a loop, unused values, duplicate keys, keywords and reserved words as names, invalid assignment targets
 - reported by the checker (not in the lang module yet): undeclared names, use before declaration, redeclare, assignment to `let` / loop var / `input`, unreachable code

## modules and builtins - for later versions, not in this one
# modules and builtins
 - not in v1

## errors
# static check errors
 - produced by lex, parse and check, before any evaluation
   - script parse/lex/check exits on the first error
 - each error has a position, a code and a message
   - format: `file.ak:line:col: error[code]: message`
   - optional `hint:` line (eg `;` -> "one statement per line")
 - check covers:
   - undeclared names,
   - use before declaration,
   - redeclare in same block,
   - assignment to `let` / loop var / `input`,
   - `break` / `continue` outside loops,
   - unreachable code after `return`,
   - unused value statements,
   - duplicate object keys and kwargs,
   - unknown callees,
   - reserved words as names

# runtime errors, positions, messages
 - the first runtime error stops the script
 - carries line of the failing expression and the kinds involved
   - eg `step.ak:12: runtime error: cannot add string and number`
 - positions are 1 based
   - columns are only included in lex errors, not in runtime errors
     - AST types don't have column info, only line
 - cases: 
   - kind mismatch in operators
   - ordering across kinds
   - division or modulo by zero
   - index out of range
   - non-integral index or range bound
   - access on `null`
   - iterating a non-iterable
   - limits exceeded

## execution model
# lex, parse, check, eval
 - lex: source to tokens with positions - no need to store the column
   - only the line is needed for runtime errors
   - column is only needed for lex errors, not for runtime errors
 - parse: tokens to an AST, recursive descent, precedence from the table above
 - check: static checks on every execution, eval does not start when check fails
 - eval: walks statements in order, blocks push and pop scopes
 - the context is checked before each statement and each loop iteration, cancellation stops the step
 - limits set by the worker:
   - max evaluated statements
   - max loop iterations
   - max value size, exceeding one is a runtime error

# script inputs and outputs for a flow step
 - input: predeclared immutable `input`, the output of the previous step or the flow start payload, any kind, `null` when none
   - read like any value: `input.user.name`, `input.items[0]`
 - output: the value of `return expr`, `null` when the script ends without `return`
 - values crossing steps are JSON compatible: number, string, boolean, null, array, object
 - logging goes through modules in later versions

## examples - for later versions, not in this one
# examples
 - not in v1 spec, `spec_def.ak` is extended with examples per section in the next step

# akha v1 - EBNF explained

source of truth: `spec_v1.ai.md`, section "grammar (EBNF)"

## notation
- `a = b ;` defines rule `a`
- `|` choice, `[ x ]` optional, `{ x }` zero or more, `( x )` group
- `"x"` is literal text, `NL` a newline, `EOF` the end of the source

## lexical grammar
- turns source text into tokens
- spaces, tabs and `//` comments are skipped, `NL` is kept, it ends a statement
- `ident` is ASCII letters, digits and `_`, not starting with a digit
- keywords (`let`, `if`, `and` ...) and reserved words (`fn try catch`) are not names
- `number` has no leading zeros and no exponent: `0`, `42`, `3.2`
- `string` uses double quotes, escapes `\n \t \r \" \\ \$`
- `${ ... }` inside a string holds a `value_ref`: a name with `.member` or `[key]` access

## script and blocks
- `script` is a list of lines, each line is empty or one statement
- `block` is `{` + newline + lines + `}`
  - `{` stays on the header line, `}` goes on its own line

## statements
- `let x = expr` immutable, `let _ = expr` discards the value
- `var x` or `var x = expr` mutable
- `target = expr` assigns to a name, member or index chain
- a bare call `f(x)` is the only expression allowed as a statement
- `if cond { } else if cond { } else { }`, `else` sits on the `}` line
- `for [var] x in expr { }`, `for [var] i, v in expr { }`
- `for [var] i range [start..end] { }`
- `break`, `continue`, `return [expr]`, `exit [expr]`

## expressions
- one rule per precedence level, lowest first:
  - `or_expr` - `or ||`
  - `and_expr` - `and &&`
  - `not_expr` - `not !` prefix
  - `in_expr` - `in`, `not in`, no chaining
  - `cmp_expr` - `== != < <= > >=`, no chaining
  - `add_expr` - `+ -`
  - `mul_expr` - `* / %`
  - `unary_expr` - `-` prefix
  - `postfix_expr` - `.name`, `[index]`, `(args)`
- `{ }` rules repeat and associate left, `[ ]` rules allow one operator only, so they don't chain
- `primary` covers literals, names, arrays, objects and `( expr )`
- object keys are names or plain strings and must be unique

## rules outside the grammar
- newlines are ignored inside `( )`, `[ ]` and object `{ }`
- a `{` after a header expression always opens a block, wrap an object in `( )`
- `a[1:2]` slices are a parse error in v1

# plan - evaluator module v1

sources
- `lang/parser/testdata/spec_def.canonical.ak`
- `lang/lexer`, `lang/parser`, `lang/ast`
- `../__arch/spec_v1.ai.md`

scope
- values, operators, environment and scopes, statements, control flow
- no functions, no modules, no builtins (v2)
- `CallExpr` exists in the AST, v1 has no callables

## packages

- `lang/eval` - values, operators, env, evaluator, runtime errors
- `lang/check` - static checker, runs before eval (spec: "not in the lang module yet")
  - the evaluator trusts a checked script but keeps defensive runtime errors
- `lang/runner` - lex, parse, check, eval from a source string

## pipeline

the pipeline has an event buss:
  - every step has an event 
    - parser: `parse`, `parse-error`
    - checker: `check`, `check-error`
    - evaluator: `eval`, `eval-error`, `eval-exit`, `eval-return`
      - the eval  can also launch events with prefix  `eval-event-${}`
        - in later version used by modules and builtins, v1 has no events in the eval step
executes in 3 steps:
- `parser.New(file, src).Parse()` returns `*ast.Script`
- `check.New().Check(ctx, script)` returns the first static error
- `eval.New(opts).Run(ctx, script, input)` returns the step output or the first runtime error
- the output is `null` when the script ends without `return`

## concepts

### values - dynamic typing

- every value is a Go type behind `Value`
- kinds: `null`, `boolean`, `number`, `string`, `array`, `object`
- a variable has no declared type, the kind lives on the value
- operations are optional interfaces, an operator works when the operand implements it
  - missing interface → runtime error `cannot <op> <kind> and <kind>`
- numbers are float64
  - `NaN` and `inf` are never produced, every Operationmetic result is checked
  - `-0` is normalized to `0` on output and in equality
  - printed shortest form, integral values without `.0`
- strings are immutable, no indexing
- arrays and objects have value semantics
  - objects keep insertion order, string keys only

### value semantics - copy rules

- plan for v1: clone on store, copy on write is a later optimization
- reads (`Ident`, `MemberExpr`, `IndexExpr`) return the stored value without a copy
- a container is cloned when it is stored into
  - a `let` / `var` binding
  - an assignment target
  - an element of an array or object literal
  - a loop variable
  - the `return` value
- a loop iterates a snapshot (clone of the iterable before the first iteration)
- literals and operator results are fresh, no clone needed
  - optimization: skip the clone when the expression is not a reference

### truthiness

- falsy: `false`, `null`, `0`, `""`, `[]`, `{}`
- true: everything else
- used by `if`, `else if`, `and`, `or`, `not`
- also used by `for` to mimic `while` in `for true {` till `break` or `return`

### expressions

- evaluation order is left to right
- `and` / `or` short circuit, the result is a boolean
- `not` returns a boolean
- `==` / `!=` never error, deep compare, different kinds are not equal
- `< <= > >=` only number/number and string/string
- `+` number/number, string/string, array/array (new value), object/maps (new value)
  - `- * / %` number/number, `/` and `%` by zero is a runtime error
- unary `-` number only
- `in` / `not in`
  - value in array: any element `==`
  - string in string: substring
  - string in object: key exists
  - other pairs: runtime error
- `arr[i]` integral, in range, not negative
- `obj[k]` k must be a string, missing key -> `null`
- `obj.name` same as `obj["name"]`
- `.` / `[ ]` on `null` -> runtime error "access on null"
- `.` on array, string, number, boolean -> runtime error
- template: each part formatted with the string conversion rules
  - string raw, number shortest, bool, `null`, array and object compact JSON

### statements and state

- `LetStmt` evaluates the value, declares an immutable binding in the current scope
- `VarStmt` declares a mutable binding, `null` when no value
- `let _ = expr` evaluates and drops the value, no binding
- `AssignStmt`
  - `Ident` target: find the nearest visible binding, must be mutable, replace the value
  - `MemberExpr` / `IndexExpr` target: resolve the root ident, it must be mutable
    - walk the path, the last step sets the member or index
    - array index set must be in range, no append
    - object member set adds or replaces the key
    - intermediate `null` -> runtime error
  - value is evaluated before the target path (see questions)
- `ExprStmt` only for calls, v1 -> runtime error "no callables in v1" (checker should report first)
- `ReturnStmt` / exit: stop the script, value or `null`

### scopes

- the root scope holds the predeclared `input` (immutable)
- every `Block` pushes a child scope, popped on exit (also on break, continue, return, error)
- lookup goes from the current scope up to the root
- shadowing: a `let` / `var` in a nested block creates a new binding, the parent is untouched after the block
  - canonical line 85 `let total = "shadowed"` inside `if`
- redeclaring in the same scope is a static error (checker), defensive runtime error in eval
- no implicit declaration
  - assignment to an undeclared name is a static error
  - the only implicit bindings: `input`, loop variables, `_` (discard, never bound)

### loops

- `ForInStmt` 
  - iterable is evaluated once, snapshot taken
  - array: single var -> element, two vars -> index, element
  - object: single var -> key, two vars -> key, value
    - AST note: single var is stored in `Value`, `Key` is empty, the evaluator maps it to the key for objects
  - other kinds -> runtime error "cannot iterate <kind>"
  - each iteration: new loop scope with the loop vars, then the body block scope
  - `Mutable` decides if the loop vars are `var` or `let`
  - add static validation for block exit, every `for` with a boolean evresion must have a `break`, `return`, or `exit` in every path
    - for loops come in range, boolen and iterable
      - loops always exit
- `ForRangeStmt`
  - start and end evaluated once, both integral numbers
  - `start >= end` -> zero iterations
  - end excluded, step 1
- `break` / `continue` apply to the innermost loop
- context and limits checked before each statement and each iteration

### control flow signals

- statements return a `Flow` instead of panics
  - `FlowNext`, `FlowBreak`, `FlowContinue`, `FlowReturn`
- a block stops on any non-next flow and hands it up
  - all type of blocks must always return a flow, even if the last statement is `return` or `exit`
- a loop consumes break / continue, passes return up
- the script consumes return and takes the value

### limits and context

- `Options` set by the worker
  - max evaluated statements
  - max loop iterations (total or per loop, see questions)
    - max value size
- `ctx.Err()` checked before each statement and each iteration
- a limit hit is a runtime error with code `limit`

### runtime errors

- `ErrRuntime` sentinel, `*RuntimeError` wraps it
- format `file.ak:line: runtime error: message`
- line only, no column
- codes
  - `kind-mismatch`, `ordering`, `div-zero`, `index-range`, `index-kind`
  - `non-integral`, `null-access`, `member-access`, `not-iterable`
  - `number-range` (NaN / inf), `immutable`, `undeclared`, `redeclared`
  - `no-callable`, `limit`, `canceled`, `internal`

## interfaces

### values

```go
package eval

type Kind int

const (
	KindNull Kind = iota
	KindBool
	KindNumber
	KindString
	KindArray
	KindObject
)

type Value interface {
	fmt.Stringer
	Kind() Kind
	Truth() bool
	Equal(Value) bool
	Clone() Value
}

type Comparer interface {
	Value
	Compare(Value) (int, error)
}

type Adder interface {
	Value
	Add(Value) (Value, error)
}

type OperationSum interface {
  Value
  Sum(Value) (Value, error)
}

type OperationMinus interface {
  Value
  Minus(Value) (Value, error)
}

type  OperationMul interface {
  Value
  Mul(Value) (Value, error)
}

type OperationDiv interface {
  Value
  Div(Value) (Value, error)
}

type Container interface {
	Value
	Contains(Value) (bool, error)
}

type Indexer interface {
	Value
	Index(Value) (Value, error)
}

type IndexSetter interface {
	Indexer
	SetIndex(Value, Value) error
}

type Memberer interface {
	Value
	Member(string) (Value, error)
}

type MemberSetter interface {
	Memberer
	SetMember(string, Value) error
}

type Iterable interface {
	Value
	Len() int
	Items() []Pair
}

type Pair struct {
	Key   Value
	Value Value
}

type Sizer interface {
	Size() int
}

type Formatter interface {
	Format() string
	JSON() string
}
```

### concrete types

```go
type Null struct{}

type Bool bool

type Number float64

type String string

type Array struct {
	Elems []Value
}

type Object struct {
	Keys   []string
	Fields map[string]Value
}
```

- assertions
  - `Null`, `Bool` -> `Value`
  - `Number` -> `Comparer`, `Adder`, `Operation`
  - `String` -> `Comparer`, `Adder`, `Container`
  - `*Array` -> `Adder`, `Container`, `IndexSetter`, `Iterable`, `Sizer`
  - `*Object` -> `Container`, `IndexSetter`, `MemberSetter`, `Iterable`, `Sizer`

### operators

```go
type Operators interface {
	Binary(op lexer.Kind, l, r Value) (Value, error)
	Unary(op lexer.Kind, x Value) (Value, error)
}
```

- `and` / `or` are not in `Binary`, the evaluator short circuits them
- `NotIn` is `not Contains`

### environment

```go
type Binding struct {
	Value   Value
	Mutable bool
}

type Env interface {
	Parent() Env
	Child() Env
	Declare(name string, v Value, mutable bool) error
	Lookup(name string) (*Binding, bool)
	Assign(name string, v Value) error
}
```

- `Declare` fails on a name already in this scope, ignores `_`
- `Assign` walks up, fails on missing or immutable

### evaluator

```go
type Flow int

const (
	FlowNext Flow = iota
	FlowBreak
	FlowContinue
	FlowReturn
)

type Options struct {
	File          string
	MaxStmts      int
	MaxIterations int
	MaxValueSize  int
}

type Evaluator interface {
	Run(ctx context.Context, script *ast.Script, input Value) (Value, error)
}

type Interp interface {
	Exec(ctx context.Context, env Env, s ast.Stmt) (Flow, error)
	ExecBlock(ctx context.Context, env Env, b *ast.Block) (Flow, error)
	Eval(ctx context.Context, env Env, x ast.Expr) (Value, error)
}

type Target interface {
	Set(ctx context.Context, env Env, t ast.Expr, v Value) error
}
```

### errors

```go
var ErrRuntime = errors.New("runtime error")

type RuntimeError struct {
	File string
	Line int
	Code string
	Msg  string
}
```

### conversion in and out

```go
type Codec interface {
	FromGo(any) (Value, error)
	ToGo(Value) any
}
```

- `input` and output cross steps as JSON compatible values

### checker (dependency, separate package)

```go
package check

type Checker interface {
	Check(ctx context.Context, script *ast.Script) error
}
```

- undeclared, use before declaration, redeclare in same block
- assignment to `let`, loop var without `var`, `input`
- unreachable after `return` / `exit`
- unknown callees (every call in v1)

## steps

- 1 values: kinds, `String`, `Truth`, `Equal`, `Clone`, formatting
- 2 operators: `Binary`, `Unary`, number range checks
- 3 env: declare, lookup, assign, shadowing
- 4 expressions: literals, idents, template, member, index, unary, binary, short circuit
- 5 statements: let, var, assign (ident and path), return
- 6 control flow: if / else if / else, for in, for range, break, continue
- 7 limits, context, runtime error positions
- 8 codec for `input` and output
- 9 runner: parse -> check -> eval, the canonical file end to end

## test cases

### values and formatting
- `3 == 3.0` true, both print `3`
- `-0` prints `0`, `-0 == 0` true
- `0.1 + 0.2` prints shortest form
- large integral `9007199254740992` prints without exponent
- `1e21` range values, formatting of very large and very small numbers
- nested array / object to compact JSON, key order kept
- string inside array in a template is quoted `["a"]`, top level is raw

### truthiness
- each falsy value in `if`, `not`, `and`, `or`
- `[0]`, `{a: null}`, `" "`, `-0` truthiness

### Operationmetic
- `7 / 2` is `3.5`
- `10 % 3`, `-7 % 3`, `7.5 % 2`
- division and modulo by zero
- overflow to inf, eg repeated `x = x * x` in a loop
- `"a" + 1`, `"2" * 3`, `{} + {}`, `null + 1`, `true + true`
- `[1] + [2]` new array, operands untouched
- `- "a"`, `- null`, `- - 3`

### comparison and equality
- `"1" == 1` false, `null == null` true, `null == false` false
- deep array and object equality, object key order ignored
- `"a" < 1` error, `[1] < [2]` error, `null < 1` error
- string code point order `"B" < "a"`

### logic
- short circuit skips a runtime error on the right, `false and (1 / 0)`
- `and` / `or` return booleans, not operands
- `not i == 1` is `not (i == 1)`

### membership
- `2 in [1, 2]`, `[1] in [[1]]` deep
- `"ell" in "hello"`, `"" in "x"`
- `"x" not in obj`, `1 in obj` error, `1 in "1"` error, `x in null` error

### access
- `arr[1.5]`, `arr[-1]`, `arr[3]` on len 3, `arr["0"]`
- `obj[1]` error, `obj.missing` -> `null`, `obj["missing"]` -> `null`
- `null.x`, `null[0]`, `"abc"[0]`, `3.x`, `[].x`
- `input` null and `input.items` -> null access error

### variables and scopes
- `var x` is `null`
- shadow in `if`, outer value back after the block
- assign to outer `var` from a nested block
- `for var i in arr3` shadows `let i = 0`, outer `i` is `0` after the loop
- `let pair` inside a loop body, fresh per iteration
- multiple `let _ = expr` in one block
- `_` is never readable

### value semantics
- `var b = a` then `b[0] = 1`, `a` unchanged
- `let x = v` from a `var v`, then mutate `v`, `x` unchanged
- `var arr = [inner]`, mutate `inner`, `arr` unchanged
- `state.items[0].name = "b"` nested set
- mutate the collection inside `for in`, iteration unchanged
- `for var x in arr { x = 1 }` does not change `arr`
- self reference `a[0] = a`

### assignment targets
- index out of range set
- set on a missing path segment `state.nope.x = 1` -> null access
- set a new key on an object via `.name` and `["name"]`
- set by index on a string, number, null

### loops
- `for x in obj` gives keys in insertion order
- `for i, v in arr` index is a number
- `for _, v in arr2`
- `for x in "abc"`, `for x in 3`, `for x in null` errors
- range `[0..0]`, `[5..2]`, `[0.5..3]`, `[0..count]` with `count` changed in the body
- range bound evaluated once
- `break` / `continue` in nested loops apply to the innermost
- `break` inside `if` inside a loop
- `return` inside a loop inside `if`
- empty loop body, empty `if` blocks

### templates
- `${arr[0]}`, `${user["content-type"]}`, `${a[b]}`
- `\${price}` stays literal
- interpolation of null, bool, number, nested objects
- runtime error inside `${ }` reports the string line

### limits and context
- canceled ctx before the first statement
- max statements, max iterations in an endless-like range
- max value size via repeated `arr = arr + arr`

### errors
- runtime error line for multi line objects and calls
- first runtime error stops, later statements not run
- `exit` and `return` both end with a value

### parser / lexer candidates (not verified against the current tests)
- keywords case insensitive in every position: `LET`, `If`, `TRUE`, `Not In`
- `! x in y` parses as `not (x in y)`, `!in` is not `not in`
- `not not x`, `- - x`, `-a[0]` is `-(a[0])`
- keyword as member name `obj.if`, keyword as object key `{if: 1}`
- `for var _, v in arr`, `for _ range [0..3]`
- `var input`, `let input = 1` (shadow of the predeclared name)
- `return` at the last line without newline, `return` followed by `}`
- object literal spanning lines inside a template `${ }` (lexer rejects newline)
- template with a string index `"${obj["a"]}"` and escaped quotes inside `${ }`
- number `1.` and `1..2` inside range, `[0..-1]`
- statements after `return` in the same block (checker)

## questions

- ⚠️ "only null's can be reassigned as a type": does a `var` keep its kind after the first non null value
  - `var x = 1` then `x = "a"`: error or allowed
  - canonical line 67 `i = null` on a string / number loop var, is `null` always allowed
- ⚠️ is kind locking a static or a runtime check
- ⚠️ loop scope vs body scope: is `for x in arr { let x = 1 }` a redeclare or a shadow
- ⚠️ `obj.missing` returns `null` like `obj["missing"]` or is it an error
- ⚠️ assignment order: value first or target path first, matters for errors in both
- ⚠️ setting a key on a nested missing member `state.a.b = 1` when `state.a` is missing: error or auto create
- ⚠️ `%` semantics for negative and non integral numbers (Go `math.Mod` sign follows the dividend)
- ⚠️ `for x in obj` single var gives keys, `for k, v in arr` index as number: confirm
- ⚠️ `let _` repeated in the same block: allowed
- ⚠️ `var` loop vars `for var _, v`: allowed
- ⚠️ `input` can it be shadowed by `let input` in a nested block
- ⚠️ max loop iterations: total for the script or per loop
- ⚠️ max value size: bytes of JSON, element count, or nesting depth
- ⚠️ deep immutable `let`: `let a = arr` then `var b = a` - spec line "a `var` copy of a `let` array is mutable but not" is cut off
- ⚠️ is the checker part of this task or a separate one, the evaluator depends on it
- ⚠️ `CallExpr` in v1: parse ok and fail in check, or fail in parse
- ⚠️ templates: are strings inside arrays / objects quoted (JSON) while a top level string is raw
- ⚠️ equality of `-0` and `0` inside arrays and as object values
- ⚠️ `exit` vs `return` any difference in output or step status
- ⚠️ runtime error positions for multi line expressions: line of the operator or of the statement

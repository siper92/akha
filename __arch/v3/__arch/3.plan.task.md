# plan - interpreter / evaluator v1

sources
- `../module/lang/parser/testdata/spec_def.canonical.ak`
- `__arch/spec_v1.ai.md`
- `../../../lang/lexer`, `lang/parser`, `lang/ast`

scope
- static check (names, scopes, mutability, unreachable, callees)
- values, operators, env, evaluator for expressions, statements and control flow
- runner: lex, parse, check, eval
- out of scope: functions, modules, builtins, `Ak.Allow`, try/catch

## packages

- `lang/value` - `Value`, kinds, operation interfaces, formatting, equality, clone
- `../../../lang/check` - static checker over `*ast.Script`, scope resolver
- `../../../lang/eval` - `Env`, `Binary`, `Unary`, evaluator, runtime errors, limits
- `../../../lang/runner` - glue: source or file in, output value or error out

dependency order: `value` <- `check` (no dep on value) <- `eval` <- `runner`

## concepts

### execution pipeline
- lex + parse (exists) -> check -> eval
- check runs on every execution, eval never starts when check fails
- first error stops everything (static or runtime)

### dynamic typing
- values carry their kind, bindings have no declared type
- six kinds: null, bool, number, string, array, object
- no implicit conversions in operators, only in `${ }` formatting
- operations dispatch on interfaces implemented by the value, a missing interface is a runtime error "cannot <op> <kind> and <kind>"
- var kind rule "only null's can be reassigned as a type"
  - proposal: a `var` holding a non null value keeps its kind, assigning another non null kind is a runtime error
  - `null` can be assigned to any var, a `null` var accepts any kind
  - see questions

### value semantics
- arrays and objects behave as values, never as shared references
- v1 strategy: clone on bind
  - clone when a value is stored into a binding (`let`, `var`, assignment, loop variable)
  - clone when a container value is stored into another container through an assignment target
  - literals are fresh, no clone needed
  - in place mutation is then safe, every `var` owns its value exclusively
- `for ... in` iterates a snapshot taken once before the first iteration
- copy on write (shared flag / refcount) is a later optimization behind the same `Cloner` interface
- `let` deep immutability is enforced statically (assignment root must be a mutable binding), no runtime freeze needed

### numbers
- float64 only, `-0` is normalized to `0` on output and in equality
- every arithmetic result is checked, NaN or Inf is a runtime error
- `/` float division, `/ 0` and `% 0` are runtime errors
- integral check for index and range bounds: `v == math.Trunc(v)` and within +-2^53
- formatting: shortest form, integral without `.0`

### truthiness
- falsy: `false`, `null`, `0`, `-0`, `""`, `[]`, `{}`
- used by `if`, `and`, `or`, `not`
- `and` / `or` short circuit and always return a bool

### variables, redeclaration and scope
- the script is the root scope, every `Block` is a new scope
- the `for` statement opens a loop scope for loop variables, the body block is nested inside it
- a new loop scope (with fresh loop variables) per iteration
- binding kinds: `let`, `var`, loop var immutable, loop var mutable, `input`
- lookup: current scope -> parents -> root
- a name is visible from its declaration to the end of its block (Go like)
  - `let x = x + 1` in a nested block reads the outer `x`
- redeclare in the same scope is a static error, shadowing in a nested scope is allowed
- `_` is never bound, `let _ = expr` evaluates and drops, can repeat in one scope
- `_` as a loop variable discards that slot

### implicit variable declaration
- there is none for user names: `x = 1` without a visible `var x` is a static error
- the only implicit bindings
  - `input` in the root scope, immutable, `null` when no input is given
  - loop variables, declared by the `for` header in the loop scope

### statements and state
- `LetStmt` - eval value, clone, declare immutable
- `VarStmt` - eval value or `null`, clone, declare mutable
- `AssignStmt`
  - target `Ident` - kind rule, clone, replace binding value
  - target path (`a.b[0].c`) - resolve root binding, walk the path, set the last step in place
  - member set on object adds or replaces the key
  - index set on array must be in range, index set on object needs a string key
  - walking through `null` or a wrong kind is a runtime error
- `ExprStmt` - only calls, v1 has no callables so the checker rejects them
- `ReturnStmt` - eval value or `null`, stop the script, `exit` behaves the same

### control flow
- statement execution returns a flow signal, no panics for control flow
- signals: normal, break, continue, return (carries the value)
- `if` / `else if` / `else` - `Else` is `*ast.IfStmt` or `*ast.Block`
- `for k, v in x`
  - array: index, element
  - object: key, value in insertion order
  - single variable: element for arrays, key for objects
  - other kinds: runtime error "cannot iterate <kind>"
- `for i range [a..b]` - bounds evaluated once, integral, `a >= b` gives zero iterations
- `break` / `continue` stop at the innermost loop, `return` propagates to the script
- context and limits checked before each statement and each loop iteration

### templates
- `${ }` reads from the current scope
- string conversion: strings raw at the top level, numbers shortest, bool, `null`
- arrays and objects as compact JSON, keys in insertion order, strings quoted inside JSON

### errors
- static: new sentinel `ErrCheck`, reuse `lexer.Error` shape with a code
- runtime: sentinel `ErrRuntime`, line only
  - format `file.ak:12: runtime error: cannot add string and number`
- cancellation returns the context error wrapped with the line

## interfaces

### value

```go
package value

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
}

type Equaler interface {
	Equal(other Value) bool
}

type Comparer interface {
	Compare(other Value) (int, error)
}

type Adder interface {
	Add(other Value) (Value, error)
}

type Arith interface {
	Arith(op lexer.Kind, other Value) (Value, error)
}

type Negater interface {
	Neg() (Value, error)
}

type Container interface {
	Contains(x Value) (bool, error)
}

type Indexer interface {
	Index(i Value) (Value, error)
}

type IndexSetter interface {
	SetIndex(i Value, v Value) error
}

type Memberer interface {
	Member(name string) (Value, error)
}

type MemberSetter interface {
	SetMember(name string, v Value) error
}

type Iterable interface {
	Len() int
	Iter() Iterator
}

type Iterator interface {
	Next() (key Value, val Value, ok bool)
}

type Cloner interface {
	Clone() Value
}

type Sizer interface {
	Size() int
}

type Formatter interface {
	Format(w io.Writer, json bool) error
}

type Null struct{}
type Bool bool
type Number float64
type String string

type Array struct {
	Elems []Value
}

type Object struct {
	keys []string
	vals map[string]Value
}

func Equal(a, b Value) bool
func Truth(v Value) bool
func Format(v Value) string
func Clone(v Value) Value
```

- asserts per type, eg `var _ Adder = (*Array)(nil)`
- `Number`: Equaler, Comparer, Adder, Arith, Negater
- `String`: Equaler, Comparer, Adder, Container
- `Array`: Equaler, Adder, Container, Indexer, IndexSetter, Iterable, Cloner, Sizer
- `Object`: Equaler, Container, Indexer, IndexSetter, Memberer, MemberSetter, Iterable, Cloner, Sizer
- `Null` / `Bool`: Equaler only

### check

```go
package check

var ErrCheck = errors.New("check error")

type Checker interface {
	Check(ctx context.Context, script *ast.Script) error
}

type BindKind int

const (
	BindLet BindKind = iota
	BindVar
	BindLoop
	BindLoopVar
	BindInput
)

type Scope interface {
	Parent() Scope
	Declare(name string, kind BindKind, pos ast.Pos) error
	Resolve(name string) (BindKind, bool)
	Child() Scope
}

func New(file string, predeclared ...string) Checker
```

- codes: `undeclared`, `use-before-decl`, `redeclare`, `assign-immutable`, `assign-undeclared`, `unreachable`, `unknown-callee`, `discard-read`, `duplicate-loop-var`

### eval

```go
package eval

var (
	ErrRuntime = errors.New("runtime error")
	ErrLimit   = errors.New("limit exceeded")
)

type Binding struct {
	Value   value.Value
	Mutable bool
}

type Env interface {
	Parent() Env
	Child() Env
	Declare(name string, v value.Value, mutable bool) error
	Lookup(name string) (*Binding, bool)
}

type Flow int

const (
	FlowNormal Flow = iota
	FlowBreak
	FlowContinue
	FlowReturn
)

type Limits struct {
	MaxStatements int
	MaxIterations int
	MaxValueSize  int
}

type Evaluator interface {
	Run(ctx context.Context, script *ast.Script, input value.Value) (value.Value, error)
}

type Binary func(op lexer.Kind, x, y value.Value) (value.Value, error)

type Unary func(op lexer.Kind, x value.Value) (value.Value, error)

type Error struct {
	Kind error
	File string
	Line int
	Code string
	Msg  string
}

func NewEnv(parent Env) Env
func New(file string, limits Limits) Evaluator
```

- internal walk
  - `exec(ctx, ast.Stmt, Env) (Flow, value.Value, error)`
  - `block(ctx, *ast.Block, Env) (Flow, value.Value, error)` opens `env.Child()`
  - `eval(ctx, ast.Expr, Env) (value.Value, error)`
  - `assign(ctx, ast.Expr, value.Value, Env) error` walks the target path
- `and` / `or` are handled in `eval`, not in `Binary` (short circuit)
- `not in` is `not (x in y)` via `Container`
- runtime codes: `kind-mismatch`, `order-kinds`, `div-zero`, `index-range`, `index-kind`, `non-integral`, `null-access`, `member-kind`, `not-iterable`, `number-range`, `var-kind`, `limit`

### runner

```go
package runner

type Options struct {
	File   string
	Input  value.Value
	Limits eval.Limits
}

type Runner interface {
	Run(ctx context.Context, src string, opts Options) (value.Value, error)
	Check(ctx context.Context, src string, opts Options) error
}

func New() Runner
```

## steps

1. `value` - kinds, constructors, `Truth`, `Equal`, `Format`, `Clone`, table tests
2. `eval.Binary` / `eval.Unary` - every op x every kind pair, table tests
3. `check` - scopes and every static rule, table tests with small `.ak` snippets
4. `eval.Env` - declare, lookup, shadowing, child scopes
5. evaluator expressions - literals, idents, templates, member, index, unary, binary, short circuit
6. evaluator statements - let, var, assign (ident and paths), return
7. control flow - if chains, for in, for range, break, continue, return inside loops
8. context cancellation and limits
9. `runner` - golden run of `spec_def.canonical.ak` with a given input

## test cases not covered yet

### parser level (seen in the code)
- `for i, i in arr` - duplicate loop vars are not rejected, needs a check rule
- `let a = _`, `"${_}"`, `for x in _` - reading `_` parses fine, needs a check rule
- `return 1` followed by a statement in the same block - unreachable, check rule
- `break` followed by a statement in the loop body - unreachable or not
- keywords as member names `obj.in` or keys `{if: 1}` - parse error today
- number literal above 2^53 like `9007199254740993` - silently loses precision
- `f()` / `a.b()` - parse ok, must fail check as unknown callee
- mixed case keywords in headers `IF x {`, `For i IN arr {`

### check
- undeclared name in an expression, in a template, in a range bound, in an assignment target
- use before declaration in the same block vs outer binding with the same name
- `let x = x + 1` in a nested block reads the outer `x`
- redeclare in the same block, shadowing in a nested block, same name in sibling blocks
- loop var visible only inside the loop, outer name restored after
- assign to `let`, to an immutable loop var, to `input`, to a member of a `let`
- assign to a parent `var` from a nested block and from a loop body
- `for var x` allows `x = ...` and `x.a = ...`
- `let _` repeated in one scope is allowed

### eval - expressions
- number formatting: `3.0` -> `3`, `-0` -> `0`, `0.1 + 0.2`, `1e21` sized values
- overflow `1e308 * 10` is a runtime error, `1 / 0`, `1 % 0`
- `%` with negative and non integral operands
- `"a" + 1`, `[1] + "a"`, `{} + {}` runtime errors
- `"a" < 1`, `[1] < [2]`, `null < 1` runtime errors
- `==` across kinds is false, `[1, [2]] == [1, [2]]`, object equality ignoring key order, `0 == -0`
- `in` for array, string, object key, number in object error, `null in arr`
- `and` / `or` short circuit: right side with a runtime error is not evaluated
- `not` / `!` on every kind
- unary `-` on non numbers
- index: negative, out of range, non integral `arr[1.5]`, string index error, object with non string key
- member on array, string, number, bool, null
- missing object key returns `null`, then `.x` on it is a null access error
- templates with arrays, nested objects, strings inside JSON, escaped `\${`

### eval - statements and state
- `var b = a` then `b[0] = 1` leaves `a` unchanged
- `var b = a` where `a` is a `var`, then change `a`, `b` unchanged
- `let x = v` where `v` is a `var`, later change of `v` does not affect `x`
- nested path assignment `state.items[0].name = "b"`
- member set adds a new key at the end of insertion order
- index set out of range, on `null`, through a missing key
- var kind rule: number -> string error, number -> null ok, null -> string ok
- `var last` starts as `null`

### eval - control flow
- `if` / `else if` / `else` picks the first truthy branch only
- truthiness of `0`, `""`, `[]`, `{}`, `null`
- `for x in obj` gives keys, `for k, v in obj` key and value
- body mutates the iterated `var`, iteration uses the snapshot
- `for var x` reassigning `x` does not change the collection or the next item
- `for i range [3..3]`, `[5..1]`, `[0..1.5]` error, `["a"..3]` error
- bounds evaluated once even if the body changes the bound variable
- nested loops: `break` / `continue` affect only the inner loop
- `return` inside a nested loop inside an `if` exits the script with the value
- script without `return` outputs `null`, `exit` equals `return`

### limits and context
- cancelled context before the first statement and inside a long loop
- max statements, max iterations, max value size (array concat in a loop)

### golden
- `spec_def.canonical.ak` with input `{items: [1, 2]}` returns `{"total": 9, "greeting": "hello ann, first is 1"}`
- same script with `input` = `null` fails at line 43 with a null access error

## questions

- ⚠️ var kind rule: is "only null's can be reassigned as a type" a runtime check, a check only on `var` roots, or also on members and elements (`state.count = "x"`)
- ⚠️ canonical line 67 `i = null` - is assigning `null` to a typed var always allowed
- ⚠️ loop body scope: is `for x in arr { let x = 1 }` shadowing or a redeclare
- ⚠️ use before declaration: Go like (outer binding wins) or TDZ like (error when the name is declared later in the same block)
- ⚠️ can `input` be shadowed in a nested block, can it be redeclared at the root
- ⚠️ unreachable: only after `return` / `exit`, or also after `break` / `continue`
- ⚠️ `%` semantics: Go `math.Mod` (sign of the dividend) or floored, allowed on non integral numbers
- ⚠️ number output for large or tiny values: `1e21` style or full digits
- ⚠️ number literals above 2^53: silent rounding or a lex/parse error
- ⚠️ keywords as member names and object keys (`obj.in`, `{if: 1}`)
- ⚠️ `.name` on an object with a missing key returns `null` like `obj["name"]`
- ⚠️ `for x in "abc"` - error in v1 (strings are not iterable) confirmed
- ⚠️ `exit` vs `return`: pure alias, or should `exit` mark the step differently (eg stop the flow)
- ⚠️ limits: is max iterations per loop or total per script, how is max value size measured (elements, bytes of JSON)
- ⚠️ check error position: `ast.Pos` has line only, format `file:line: error[code]` or keep `line:0`
- ⚠️ `${ }` "treated as a block": can it ever declare names, or is it only a read of the current scope
- ⚠️ runtime error messages: use kind names (`string`, `number`) or also values

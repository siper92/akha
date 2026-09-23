---
name: ak-lang-feature
description: Implement a new .ak language feature end to end (token, lexer, ast, parser, check, eval, module, runner) with data definitions and table tests written first. Use when adding syntax, statements, expressions, operators, scopes or builtins to the akha language.
---

# ak-lang-feature

A language feature crosses every layer of `lang/`. Work data first, tests second, code third,
one layer at a time, smallest public API that does the job.

## pipeline

Source flows through these packages in this order. A feature touches them in the same order.

1. `lang/token` - `Kind` enum, `kindNames` table, `Pos`, `Token`, keyword lookup
2. `lang/lexer` - runes to tokens, `Lexer` interface with `Next` and `All`
3. `lang/ast` - `Node`, `Stmt`, `Expr`, concrete nodes, `expr()` and `stmt()` markers
4. `lang/parser` - recursive descent, `Parser.Parse() (*ast.Script, error)`, `Errors` list, `sync()` recovery
5. `lang/check` - static rules over the ast, `Checker.Check(ctx, script, reg) []Diagnostic`
6. `lang/eval` - `Value`, `Env`, ops, `Evaluator.Eval(ctx, script) error`
7. `lang/module/ak` and `lang/module/fs` - builtins as `eval.NewBuiltin(eval.Spec{...}, fn)`
8. `lang/runner` - lex, parse, check, eval from a source string, exit codes

`__arch/v1/lang.spec` is the language spec. It is the source of truth and is edited first.

## procedure

### 1. spec

- add the syntax to `__arch/v1/lang.spec`: an example in the top part, a rule under `// rules`,
  a production under `// grammar`
- if a rule is open, mark it `unknown` in the spec instead of guessing

### 2. data definition

Define every new shape before writing any behaviour.

- `lang/token`: new `Kind` constants at the end of the enum, a row in `kindNames`, keywords in
  the keyword table
- `lang/ast`: one struct per node, exported fields, `P token.Pos`, `Pos()` method, marker method,
  `var _ Stmt = (*X)(nil)` or `var _ Expr = (*X)(nil)` assertion
- `lang/eval`: sentinel errors `var ErrX = errors.New("...")`, new `Value` kinds only if the
  feature really needs one
- `lang/check`: diagnostics are `Diagnostic{Pos, SeverityError, Msg}`, no new types

Rules:

- an ast node holds data only, no behaviour beyond `Pos()`
- a node that is both a statement and an expression is wrapped, not double marked
  (`CallStmt{Call *Call}`)
- keep the union small: reuse `token.Kind` as the operator type in `Binary` and `Unary`
- do not add fields for later features

### 3. test cases

Write the tables before the implementation. Every touched package gets cases.

- table tests use the multi-line `tu.Case` literal, grouped under a `// ---` banner per behaviour
- each behaviour gets a happy case, an edge case and an error case
- lexer cases: exact token list with positions, `tok(kind, lit, line, col)` and `eof(line, col)`
- parser cases: exact ast built with the helpers in `helpers_test.go`, error cases list every
  `parser.Error` with position and message, recovery cases show the next line still parses
- check cases: count and message of diagnostics
- eval cases: a recording module `T` captures calls, expected output is the joined call list
- runner cases: exit code or sentinel error (`ErrParse`, `ErrCheck`, `ErrRuntime`)
- golden files live in `lang/lexer/testdata/<feature>.ak`, one per feature, and are asserted
  by both the lexer golden test and the parser golden test
- `tu.Run` skips the value check when `Case.Err` is set, slice error types use a result struct

Naming: `snake_case` names that say the behaviour, `Test<Layer><Behaviour>` functions.

### 4. implementation

One layer at a time, run the tests of the layer before moving on.

- token: `String()` falls back to `kind(n)`, add the keyword table `Lookup(ident string) Kind`
- lexer: longest match first (`==` before `=`, `...` before `.`), one `case` per rune
- parser: one method per production, `expect(kind, what)` for required tokens,
  `errorf(pos, ...)` then `sync()` on syntax errors, keep parsing on rule errors
- parser error text: `expected X, got Y` where Y is `describe(tok)`
- check: one method per statement kind, one `errorf` per rule, message names the offender
- eval: one method per node, `Env` is the only mutable state, control flow signals are
  values returned from `stmt`, never panics
- eval error text: `line:col: what: %w` with a sentinel
- module builtins: `Spec` first, then `CallFunc`, arguments read with `eval.ArgString`,
  `eval.ArgInt`, `eval.KwargString`, `eval.KwargInt`
- runner: only when the exit codes or the pipeline order change

### 5. verify

```
just lint
just test-lang
just test
```

All three must pass. Report failures with the output, never hide them.

## api sheet

The shapes below are the shared contract for features in this version. Extend, do not rename.

```go
// ast
type Stmt interface { Node; stmt() }
type Expr interface { Node; expr() }
type Script struct { Stmts []Stmt }
type Block struct { Stmts []Stmt; P token.Pos }
type CallStmt struct { Call *Call }
type Let struct { Name string; Value Expr; P token.Pos }
type Assign struct { Name string; Value Expr; P token.Pos }
type If struct { Cond Expr; Then *Block; Else Stmt; P token.Pos }
type For struct { Var string; Iter Expr; Body *Block; P token.Pos }
type While struct { Cond Expr; Body *Block; P token.Pos }
type Break struct { P token.Pos }
type Continue struct { P token.Pos }
type Call struct { Target Selector; Args []Expr; Kwargs []Kwarg; P token.Pos }
type Ident struct { Name string; P token.Pos }
type Literal struct { Kind token.Kind; Value string; P token.Pos }
type Binary struct { Op token.Kind; X, Y Expr; P token.Pos }
type Unary struct { Op token.Kind; X Expr; P token.Pos }

// eval
type Env interface {
    Lookup(name string) (Value, bool)
    Define(name string, v Value) error
    Assign(name string, v Value) error
    Child() Env
}
func NewEnv(parent Env) Env
func Binary(op token.Kind, x, y Value) (Value, error)
func Unary(op token.Kind, x Value) (Value, error)
```

## conventions

- interface first, `var _ I = (*Impl)(nil)` for every concrete type
- interfaces live in `<pkg>/<pkg>.go`, impls next to them
- one allocation site per shape (`lexer.New`, `parser.New`, `eval.New`, `eval.NewEnv`)
- sentinel errors plus `fmt.Errorf("%w")`
- `context.Context` first argument everywhere
- no comments in code
- no new dependencies

## don't

- parse and check in the same layer, the parser knows syntax, the checker knows rules
- put scope or type logic in the parser
- make a keyword out of an identifier the spec does not list
- widen `Spec` or `Builtin` for one feature
- leave old tests asserting the old behaviour, update the expectation and say why in the report
- skip `just lint`
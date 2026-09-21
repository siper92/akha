# akha flow language - parser + VM system design

Source task: ./task.md
Inputs read: ./__arch/v1/lang.spec, ./__arch/v1/examples/**/*.specx

## 1. Summary

Build a small flow language (file extension `.ak`) that is compiled to an
intermediate program and executed by a Go virtual machine. The language targets
n8n-like automation: a script calls host modules (`FS`, `Git`, `Code`, `AI`,
`Flow`) under an explicit permission list, with simple control flow.

"Solidity style" is read as: source -> compiler -> portable program ->
sandboxed VM, with host functions exposed to the program through a fixed ABI.
⚠️ confirm this reading in section 4.

Minimal end-to-end target for the base structure:

```
Akha.Log("Hello World")
```

Running `akha run hello.ak` appends the line to `./akha.log`.

## 2. Observations from the inputs

Divergences between `lang.spec` and the examples that need a decision:

- root namespace is `Akha` in the spec, `BZ` in all examples
- extension is `.ak` in the task, `.specx` in the examples
- spec says `Log` writes to db, task says it writes to `./akha.log`
- spec uses named args (`code=0`, `log="info.log"`), examples use positional only
- git status shows a `lang.specx` next to `lang.spec`; not listed in the task,
  so not read ⚠️ it may be the "more full" spec the task refers to

Language surface actually used by the examples:

- comments: `// ...`
- variables: `$name = expr`, reassignment allowed
- string interpolation: `"text ${var}"`
- literals: string, integer, list `["a", "b"]`
- calls: `Module.Func(args)`, nested calls as args
- method calls on values: `$x.IsEmpty()`, `$x.Len()`, `$x.EndsWith(".go")`,
  `$x.Contains("s")`, `$cs.MovePackage(a, b)`
- operators: `==`, `!=`, `&&`, `||`, `!`
- blocks by indentation (4 spaces), no braces, no `end`
- `If` / `Else If` / `Else`
- `ForEach $x in expr`, `ForEach $i, $x in expr`
- `While cond`, `Repeat n`
- `Break`, `Continue`
- `BZ.Allow(A, B)`, `BZ.Disallow(Git)` with bare identifiers as args
- `BZ.Exit(msg, code)`
- `Flow.Generate`, `Flow.Validate`, `Flow.Run` (flows as values)
- folder-as-flow: files run in filename order, permissions granted in file 01
  are still in force in files 02 and 03

Modules referenced: `BZ`/`Akha`, `FS`, `Git`, `Code`, `AI`, `Flow`.

## 3. Goals and non-goals for the base structure

Goals:

- lexer + parser producing a typed AST for the full surface above
- validator that rejects bad programs before anything runs
  (undefined vars, unknown modules, calls to modules not in `Allow`)
- compiler from AST to a VM program
- VM that runs the program with pluggable host modules
- `Akha` module with `Log`, `Debug`, `Exit`, `Allow`, `Disallow`, `LogSetup`
- CLI: `akha run`, `akha check`, `akha fmt` (fmt optional)

Non-goals for now (⚠️ assumed, confirm):

- real `AI`, `Git`, `Code` implementations (stub interfaces only)
- user-defined functions, imports, types
- durable/resumable execution
- visual editor

## 4. Clarifying questions

Language:

- Q1 root namespace: `Akha` or `BZ`? Alias both?
- Q2 file extension: `.ak` only, or also accept `.specx`?
- Q3 blocks: indentation-based (as examples) or braces? If indentation,
  tabs allowed? Fixed 4 spaces?
- Q4 named args: required (`code=0`) or optional? Mixed with positional?
- Q5 typing: dynamic only, or static checks for module signatures at
  `akha check` time? Recommendation: dynamic values, static call-shape checks.
- Q6 are these needed in v1: arithmetic (`+ - * /`), `<`, `>`, maps/objects,
  `null`, booleans as literals (`true/false`), `Return`?
- Q7 is `$attempts = 0` inside the While in example 03 a bug in the example
  or intentional? (it never increments, loop is unbounded unless Break)
- Q8 string interpolation: only `${var}` or full expressions `${a.Len()}`?
- Q9 `BZ.Log("found ${files} files")` interpolates a list; what does a list
  render as?

Execution model:

- Q10 what exactly does "solidity style" mean to you: bytecode + stack VM,
  gas/step budget, deterministic execution, or just "compile then run"?
- Q11 tree-walking interpreter vs bytecode VM for v1? Recommendation: define
  a `Program` IR now, ship a tree-walker first, swap to bytecode later behind
  the same interface.
- Q12 concurrency: sequential only, or parallel `ForEach` later? Does the VM
  need to be safe to run several flows at once in one process?
- Q13 step/time budgets: cap on loop iterations, wall clock, AI calls?
- Q14 what does `BZ.Exit(msg, 1)` do inside a subflow run via `Flow.Run`,
  kill the parent or return an error value?

Permissions:

- Q15 is `Allow` static (checked before run) or dynamic (checked at call)?
  Recommendation: both; validator fails early, VM enforces at call time.
- Q16 can `Allow` appear after use, or must it be at the top?
- Q17 folder-as-flow: do permissions and variables both carry across files,
  or only permissions? Examples suggest only permissions ⚠️.
- Q18 is `Akha`/`BZ` itself always allowed (spec says yes)?
- Q19 `Disallow` semantics when a parent flow allowed the module: does a
  child `Disallow(Git)` narrow only itself or also the parent?

Runtime and IO:

- Q20 `Log` target: `./akha.log` (task) vs db (spec) vs `LogSetup` path.
  Recommendation: `LogSetup` sets sinks, default sink is `./akha.log`.
- Q21 `Debug`: file, stderr, or both? Enabled by `--debug` flag?
- Q22 working directory: dir of the `.ak` file or the process cwd?
- Q23 exit codes: `Exit` code propagates to the process exit status?
- Q24 error handling: does a failing module call abort the flow, or should
  there be a `Try` construct / error values like `$err = Flow.Validate(...)`?

Modules and extensibility:

- Q25 are modules compiled into the binary (Go interfaces) or loaded as
  plugins (go plugin, gRPC, WASM)?
- Q26 should module signatures be declared in a definition file
  (protobuf under `_defs/proto/`) and Go stubs generated, per CLAUDE.md
  code-gen convention?
- Q27 `Flow.Generate` requires an AI backend; which one, and is it in scope?
- Q28 do value methods (`.Len()`, `.EndsWith()`) belong to a fixed builtin
  set on string/list types, or can modules return objects with methods
  (`$cs.MovePackage`)? Examples need both.

Tooling:

- Q29 is `akha check` (validate without running) required in v1?
- Q30 do you want an LSP/editor support later? It affects how much position
  info the AST keeps.
- Q31 tests: golden-file tests per example flow using `tu.Case`?

## 5. Requirements derived from the examples

Functional:

- R1 parse every file under `./__arch/v1/examples` without error
- R2 validate: reject use of a module not allowed, undefined `$var`,
  `Break`/`Continue` outside a loop, `Else` without `If`
- R3 execute control flow: If/Else If/Else, ForEach (with index), While,
  Repeat, Break, Continue
- R4 values: string, int, bool, list, opaque host object, null
- R5 builtin methods on string and list: `Len`, `IsEmpty`, `EndsWith`,
  `Contains` (extend later)
- R6 host object methods dispatched to the Go module that produced them
- R7 permission set per execution context, inherited by subflows
- R8 `Flow.Run` executes a flow value (string or compiled program) in a child
  context
- R9 folder-as-flow: run `*.ak` files in a folder in lexical order sharing
  one permission set
- R10 `Log`/`Debug` sinks configurable, default `./akha.log`
- R11 `Exit` stops the flow and sets the process exit code

Non-functional:

- N1 module boundary is a Go interface; every module asserted with
  `var _ Module = (*impl)(nil)`
- N2 VM never imports concrete modules; only a registry
- N3 every AST node and runtime error carries file:line:col
- N4 deterministic given the same host responses (needed for replay tests)
- N5 no goroutines inside the VM core in v1

## 6. System architecture

Pipeline:

```
source (.ak)
  -> lexer      : tokens with positions, INDENT/DEDENT synthesized
  -> parser     : AST
  -> resolver   : symbol table, scopes, module lookup, permission analysis
  -> compiler   : Program (IR)
  -> vm         : executes Program against a Context
  -> host       : Module registry, values, sinks
```

Components:

- `lexer`: hand-written, indentation stack like Python, string interpolation
  tokenized into parts
- `ast`: node types, `Pos` on every node, visitor
- `parser`: recursive descent, precedence climbing for `! && || == !=`
- `sema` (resolver/validator): scopes, unknown identifiers, module and
  function existence, static permission check, loop nesting rules, warnings
- `ir`/`program`: the compiled unit; opaque to callers, serializable later
- `compiler`: AST -> Program
- `vm`: `Run(ctx, program) error`; owns frames, variables, loop state
- `value`: runtime value interface, builtin methods for string/list
- `module`: `Module` interface, `Registry`, `Permissions`
- `modules/akha`, `modules/fs`, `modules/git`, `modules/code`, `modules/ai`,
  `modules/flow` (last four stubs in v1)
- `runtime`/`context`: cwd, permissions, log sinks, cancellation, budgets,
  parent link for subflows
- `cli`: cobra commands `run`, `check`, `fmt`; viper for config
  (`--debug`, `--log`, `--cwd`)

Key interfaces (sketch, names open):

```go
type Module interface {
    Name() string
    Call(ctx *Context, fn string, args []Value) (Value, error)
}

type Object interface {
    Value
    Method(ctx *Context, name string, args []Value) (Value, error)
}

type Program interface {
    Entry() Block
    Modules() []string
}

type VM interface {
    Run(ctx *Context, p Program) (Result, error)
}
```

Execution context:

- `Permissions`: set of allowed module names, `Allow` adds, `Disallow`
  removes, child contexts copy the parent set
- `Sinks`: log and debug writers set by `LogSetup`
- `Budget`: max steps, max loop iterations, deadline (⚠️ values unknown)
- `Parent`: for `Flow.Run` nesting and for `Exit` propagation

Permission enforcement happens twice:

- static: `sema` walks the file(s) in order and fails on a call to a module
  that has no preceding `Allow` in this flow or its folder predecessors
- dynamic: `vm` checks `ctx.Permissions` before every `Module.Call`

## 7. Steps to design the VM and the system around it

Step 1: freeze the v1 grammar

- write an EBNF for the surface in section 2
- decide Q1-Q9
- produce a `grammar.md` and 5 to 10 tiny fixture programs, including
  invalid ones with expected error messages

Step 2: lexer

- token set, INDENT/DEDENT, string interpolation parts, positions
- table tests: token stream per fixture

Step 3: AST + parser

- node types, one factory per node
- recursive descent, error recovery good enough to report one clear error
- table tests: parse every example, compare to golden AST dump

Step 4: semantic analysis

- scopes and variable resolution
- module registry lookup for `Module.Func` and arity
- permission analysis, loop-nesting checks, unreachable code after `Exit`
- this is what `akha check` runs

Step 5: value model

- `Value` interface, concrete string/int/bool/list/null/object
- builtin method table for string and list
- truthiness and equality rules written down

Step 6: Program IR + compiler

- start with a lowered tree (desugar `Else If`, `Repeat` -> counted loop,
  interpolation -> concat)
- keep it serializable so `Flow.Generate`/`Flow.Validate` can pass it around
- decide here whether bytecode comes in v2 (Q11)

Step 7: VM core

- frame with variable map, loop stack for Break/Continue, step counter
- host call path: permission check -> registry -> module -> value
- `Exit` as a typed error carrying the code
- table tests: run each example with fake modules and assert the recorded
  calls and log lines

Step 8: module system

- `Registry` with `Register(Module)` and `Lookup(name)`
- `akha` module first, `fs` second, others as recording stubs
- decide plugin story (Q25, Q26)

Step 9: flows as values and folder-as-flow

- `Flow` module: `Validate` runs steps 2-4, `Run` compiles and runs in a
  child context
- folder runner: sort files, one shared permission set, one shared sink

Step 10: CLI and observability

- `akha run <file|dir>`, `akha check`, `--debug`, `--log`
- structured log lines with flow name, file, line
- exit code mapping

Step 11: hardening (after v1)

- budgets, cancellation, parallel ForEach, bytecode VM, LSP, formatter

## 8. Proposed layout

```
cmd/akha/                cobra entry
internal/lang/lexer/
internal/lang/ast/
internal/lang/parser/
internal/lang/sema/
internal/lang/ir/
internal/lang/compiler/
internal/vm/
internal/vm/value/
internal/module/         Module, Registry, Permissions
internal/module/akha/
internal/module/fs/
internal/module/git/     stub
internal/module/code/    stub
internal/module/ai/      stub
internal/module/flow/
internal/runtime/        Context, sinks, budget
_defs/proto/module.proto if Q26 is yes
testdata/flows/          fixtures and golden files
```

## 9. Alternative technologies with a similar idea

Workflow / automation engines (open source):

- n8n: the reference; fair-code license, not OSI open source
- Node-RED: Apache-2.0, flow-based, JS runtime
- Windmill: AGPL, scripts in Python/TS/Go/Bash as workflow steps
- Activepieces: MIT core, n8n-like
- Automatisch: AGPL, Zapier-like
- Temporal: MIT, durable workflows with a first-class Go SDK; strongest
  option if resumable execution becomes a requirement
- Cadence and Netflix Conductor: same family as Temporal
- Argo Workflows and Tekton: Apache-2.0, DAGs on Kubernetes
- Kestra: Apache-2.0, YAML-declared flows with a plugin system
- Dagger: Apache-2.0, Go-native pipelines, good model for a typed module API
- Prefect and Airflow: Python, Apache-2.0

Embeddable languages and VMs in Go (build on instead of from scratch):

- starlark-go: Python-like, indentation-based, deterministic, sandboxed,
  host functions in Go; closest match to the examples' syntax
- goja: ES5/ES6 in pure Go
- gopher-lua: Lua VM in Go
- tengo: small scripting language with a bytecode VM in Go
- risor: Go-flavoured scripting language, bytecode VM
- expr and cel-go: expression-only evaluators; useful for `If` conditions
  if the flow structure stays declarative
- yaegi: Go interpreter in Go
- wazero: run WASM modules as sandboxed steps
- HCL and CUE: configuration languages with evaluation, useful as a
  declarative alternative to an imperative script

Solidity-like reference points:

- go-ethereum EVM: stack VM with gas metering and a host ABI; the gas
  model maps well to step budgets and AI call limits
- Rego (OPA): compiled to an IR and run by a VM in Go, good example of
  compiler + VM layering in a Go codebase

Parser tooling in Go:

- hand-written recursive descent (recommended for indentation grammar)
- participle: struct-tag based parser generator
- goyacc, ANTLR4 with the Go target
- pigeon (PEG)

## 10. Risks and open decisions

- indentation grammar plus string interpolation is the hardest part of the
  lexer; decide early (Q3, Q8)
- permission analysis across folder files needs an ordering rule (Q16, Q17)
- `Flow.Generate` pulls an AI dependency into the language core; keep it
  behind the `Module` interface so the core stays offline
- host objects with methods (`$cs.MovePackage`) blur the line between
  values and modules; the `Object` interface in section 6 is the proposal
- unbounded `While` in the examples needs a budget or it hangs

## 11. Suggested next step

Answer Q1-Q11 and Q15-Q20. With those fixed, step 1 (grammar + fixtures)
can be written and reviewed before any Go code.

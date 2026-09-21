# akha - flow language and VM design

Design response for `task.md`. Scope: the base structure of the language, the compiler frontend, the IR, and the flow execution engine. Implementation of the VM internals is deliberately out of scope.

Source spec read: `../../_research/v1/lang.spec`.

⚠️ unknown: `task.md` points at `./__arch/v1/lang.spec`, but `__arch/` is empty and the spec lives in `_research/v1/lang.spec`. Confirm which path is canonical before more files are added there.

## 1. design approach and architecture

### 1.1 the shape of these systems

Every system in this class (n8n, Temporal, Airflow, Node-RED, the EVM) is the same five layers. Naming them up front keeps the module boundaries honest.

- source layer: text the user writes (`*.ak`), or a visual graph that serializes to it
- frontend: lexer -> parser -> AST. Knows syntax only. Knows nothing about modules.
- middle: resolver -> validator -> lowering. Binds names to the module registry, checks capabilities, emits IR.
- backend: scheduler + VM. Executes IR, owns time, retries, checkpoints.
- host: modules/effects. The only code allowed to touch the outside world.

The rule that makes the whole thing work:

- the frontend must never know a module exists
- the VM must never perform I/O
- the host must never see an AST

A module call is just a generic qualified call expression `Ns.Fn(args)`. Adding `HTTP`, `Git`, or `Docker` is a registry entry, not a grammar change. This is the single most important constraint in the design and it satisfies the "easy to add new modules" key feature directly.

### 1.2 the solidity/EVM analogy, made concrete

The task frames this as a "solidity style machine". The useful parts of that analogy:

- deterministic core: same IR + same inputs + same journal = same execution, always
- host calls: the EVM has precompiles; akha has modules. The VM never does syscalls itself.
- metering: gas becomes a step/time/memory/call budget. Prevents runaway flows.
- receipts and logs: every run produces a signed, replayable record of what happened

The parts to drop:

- no consensus, no global state tree, no stack-depth 1024 style artifacts
- values are rich (maps, lists, handles), not 256-bit words
- effects are genuinely non-deterministic (HTTP, FS), so determinism is recovered by journaling results, not by forbidding them

### 1.3 the one decision that drives everything: durable execution

The key feature list asks for tracable flows, retries, and error handling. That pushes hard toward one architecture:

- compile to a linear IR with an explicit program counter and explicit frames
- the VM is a step function: `Step(state) -> (state', Effect|nil, error)`
- the VM never calls a module directly. It yields an `Effect` and suspends.
- the runtime executes the effect, appends request+result to a journal, and resumes the VM
- because all VM state is serializable (pc, frames, locals, stack), any suspend point is a checkpoint
- crash recovery = load IR, replay journal without re-executing effects, continue

This is the Temporal model applied to a small VM. It gives, for free:
- retries at effect granularity, not flow granularity
- resumable long-running flows (waits, human approval, schedules)
- deterministic replay for debugging: "show me exactly what this run did"
- trivial test doubles: swap the effect executor for a table of canned results

A tree-walking interpreter is easier to write but cannot checkpoint mid-expression without capturing the Go call stack. That closes off durability permanently. So: do not tree-walk. Recommendation is bytecode.

### 1.4 architecture diagram

```
 file.ak
   |
   v
[lexer] --tokens--> [parser] --AST--> [resolver] --typed AST--> [lowering]
                                          |                          |
                                  module registry             Flow IR (proto)
                                  capability set                     |
                                                                     +--> [graph planner] --> node/edge view (n8n-like UI)
                                                                     |
                                                                     v
                                                              [scheduler]
                                                                     |
                                                            +--------+--------+
                                                            |                 |
                                                        [VM core]  <--->  [effect host]
                                                        pure, det.        modules, I/O
                                                            |                 |
                                                            +--------+--------+
                                                                     |
                                                              [journal / WAL]
                                                              [trace / logs]
```

## 2. alternative technologies

Prior art worth reading before writing a line of code. ⚠️ unknown: licenses change, verify each before adopting.

### 2.1 flow/workflow engines (the product space)

- n8n: the stated reference. Node graph, huge integration catalog, fair-code license (not OSI open source) ⚠️ verify commercial terms
- Temporal: durable execution done properly. Workflow code is normal Go/TS, replayed from an event history. The closest thing to the execution model proposed here.
- Windmill: scripts (TS/Python/Go) composed into flows, with a visual editor. Rust backend. Very close to the akha idea.
- Kestra: declarative YAML flows, plugin ecosystem, JVM
- Node-RED: the original visual flow runtime, JS, message-passing nodes
- Argo Workflows: k8s-native DAG/step engine, containers as steps
- Prefect / Dagster / Airflow: Python-first data orchestration, strong on scheduling and observability
- Conductor (Orkes) and Cadence: JSON-defined workflow DSLs over a durable engine
- StackStorm: event-driven "if this then that" ops automation with a large pack ecosystem
- Activepieces / Huginn: open source Zapier-likes, useful for module/trigger UX ideas
- Dagger: build/CI pipelines as code with a typed API and content-addressed caching. Good source of ideas for step caching.

### 2.2 embeddable languages/VMs in Go (the build-vs-buy question)

- Starlark (`go.starlark.net`): Python-like, deterministic by design, hermetic, battle-tested in Bazel. Strongest off-the-shelf candidate.
- Risor: modern Go-embeddable scripting language, compiles to bytecode, good Go interop
- Tengo: small, fast bytecode-VM scripting language for Go
- Expr (expr-lang) and CEL (cel-go): expression-only evaluators, safe and fast. Ideal for conditions inside a flow, not for the whole flow.
- Yaegi: Go interpreter in Go. Powerful, but a very large attack surface for untrusted input.
- goja: ECMAScript in Go, good if JS familiarity matters for module authors
- gopher-lua: mature Lua VM in Go, coroutines map nicely onto suspendable flows
- wazero: zero-dependency, pure-Go WASM runtime. Not a language, but the best available sandbox for untrusted module code.

### 2.3 parser tooling

- hand-written recursive descent + Pratt: recommended, see 5.3
- participle: struct-tag driven parser generator. Excellent for prototyping the grammar in an afternoon.
- goyacc, ANTLR, pigeon (PEG): heavier, better once the grammar stabilizes
- tree-sitter: needed later anyway for editor highlighting and incremental reparse

### 2.4 config/data languages worth stealing ideas from

- CUE: types and values unified, strong constraint checking. Good model for module signature validation.
- HCL: the Terraform language. Block syntax and its function-call model are close to what akha needs.
- Pkl, Dhall, Jsonnet: typed/total configuration languages, useful for the "safe by construction" mindset

### 2.5 recommendation

Build the language, do not embed one. Rationale:

- the differentiator is the static flow graph plus durable replay, and embedding a general language makes both much harder (arbitrary control flow is hard to render as a graph; foreign call stacks are hard to checkpoint)
- the surface area in `lang.spec` is tiny; a hand-written frontend is a few days of work
- borrow tactically: Starlark's determinism rules, CEL for guard expressions if a full expression VM is not wanted early, wazero for third-party module isolation

Hedge: prototype the grammar with participle in week 1 to validate the syntax with real users, then replace it with the hand-written frontend once the grammar stops moving.

## 3. clarifying questions

Answers here change the design materially. Each has a recommended default so work is not blocked.

### 3.1 language semantics

1. File extension: `.ak`, `.akha`, or `.flow`? Recommend `.ak` for source, `.akc` for compiled IR.
2. Are statements newline-terminated with no semicolons? Recommend yes, with `\` continuation and implicit continuation inside open parens.
3. Do variables exist in v0? `lang.spec` shows none, but `FS.ReadFile` returning content is useless without binding. Recommend adding `x := expr` in v0.
4. Static or dynamic typing? Recommend dynamic values with statically-checked module signatures (arity, param names, coarse types). Full inference later.
5. Is `Akha.Allow(FS...)` a statement executed at runtime, or a directive hoisted at compile time? Recommend compile-time directive, must appear before any use, so the capability manifest is known before execution starts.
6. What is `FS...` syntactically? A namespace glob, or a varargs spread? Recommend a dedicated `NamespaceGlob` node, and `FS.ReadFile` as the narrow form.
7. Mixed positional and named args confirmed by `Akha.Exit("error", code=0)`. Rule: all positional args precede all named args. Confirm.
8. Are top-level statements sequential by default, with parallelism opt-in? Recommend yes; explicit `parallel { }` later.
9. Error propagation model: exceptions, Go-style multi-return, or a Result value with `?` propagation? Recommend Result plus `try`/`catch` block, since the engine must see the boundary to attach retry policy.
10. Is string interpolation needed in v0? Recommend `"hello ${name}"` from the start; retrofitting it into the lexer is painful.
11. Do flows take typed inputs and produce typed outputs (callable as sub-flows)? Recommend yes, declared by `flow main(x: string) -> string`, even if v0 only supports `main`.

### 3.2 execution model

12. Is a single run allowed to outlive the process (waits, timers, human approval)? This is the biggest question. Recommend yes, and it is why the durable design is proposed.
13. Where does run state live: memory, embedded (sqlite/badger/bolt), or postgres? Recommend an interface with an embedded default and postgres for the distributed mode.
14. Single binary now, distributed workers later? Recommend designing the effect host as an interface so it can become an RPC boundary without touching the VM.
15. Are flows expected to be deterministic on replay? If yes, `time.Now`, randomness, and iteration order must all be routed through the host and journaled.
16. Concurrency model inside a run: none in v0, then fan-out/fan-in, or full async/await? Recommend a scheduler with a runnable queue from day one, running exactly one frame in v0.
17. What are the default budgets: max steps, wall clock, memory, per-module call count?
18. Is step-level caching wanted (skip a pure step whose inputs are unchanged)? Recommend designing the IR with a content hash per call site so this stays possible.

### 3.3 modules

19. Are modules first-party Go only in v0, or does third-party loading ship early? Recommend first-party in-process for v0, plugin isolation in v2.
20. Are module functions sync only, or can they be long-running/async with a completion callback? Recommend the effect envelope supports async from the start even if v0 only issues sync effects.
21. Do modules hold state or handles across calls (open file, HTTP session, db transaction)? Recommend opaque handle values owned by the runtime, never raw pointers into script values.
22. Is module versioning needed in source (`use FS@1`)? Recommend recording the version in the IR and pinning it at compile time.
23. Where do module definitions live given the sqlc/protobuf/graphql conventions? Recommend the module manifest is a protobuf message so it can be served to the UI and validated by the compiler from one definition.
24. Do triggers (cron, webhook, file watch) belong in the language or in the platform around it? Recommend language-level `on` declarations that compile into platform trigger registrations.

### 3.4 security and sandbox

25. Is untrusted `.ak` source a real threat model (multi-tenant SaaS), or is the author always trusted? This decides in-process vs wasm and how much effort goes into metering.
26. Are capabilities coarse (`FS`) or fine (`FS.read("./data/**")`)? Recommend the grammar accepts fine-grained constraints from day one even if v0 only enforces the coarse level.
27. Is a deny-by-default filesystem root/jail required per run? Recommend yes, a run gets a working directory and cannot escape it.
28. Are secrets a first-class value type with redaction in logs and traces? Recommend yes; retrofitting redaction is very hard.
29. Does network egress need an allowlist per flow?

### 3.5 observability and ops

30. `Akha.Log` logs to "db" per the spec, `Akha.Debug` to file, and the example says `Akha.Log("Hello World")` logs to `akha.log` in the current dir. Which is authoritative? Recommend a sink abstraction: default file sink, db sink when configured.
31. Is OpenTelemetry the trace target? Recommend yes, one span per call site, trace id = run id.
32. Is the execution trace a user-facing product surface (a replayable timeline in a UI)? This determines how much detail the journal must keep.
33. What are the retry defaults, and can a call site override them (`retries=3, backoff="exp"`)?
34. Is a GraphQL API for flows/runs/logs in scope, given the stack listed in `../../CLAUDE.md`?

### 3.6 product scope

35. Is there a visual editor, and must the mapping be round-trip (graph -> `.ak` -> graph)? Round-trip fidelity is a hard constraint that shapes the AST (comments and formatting must be preserved).
36. Is the "distributed build system" in the project overview the first real consumer of this language? If so, the caching and content-addressing questions get promoted.

## 4. requirements

### 4.1 functional, v0 base structure

- parse `*.ak` into an AST with full position info and recoverable errors
- support qualified calls, string/number/bool/null literals, positional and named args
- support the `Akha.Allow(FS...)` capability directive
- resolve every call against a module registry at compile time, failing with a precise diagnostic on unknown namespace, unknown function, bad arity, bad param name
- lower to a serializable Flow IR
- execute the IR on a VM that yields effects and never performs I/O itself
- run `Akha`, `FS` modules matching `lang.spec`
- emit a structured run record: every call, its args, its result, its duration, its error
- `akha run flow.ak`, `akha check flow.ak`, `akha fmt flow.ak`, `akha graph flow.ak` via cobra

### 4.2 functional, planned (design must not block these)

- `:=` bindings, expressions and operators
- `if`/`else`, `for ... in`, `while`
- user-defined `fn` and sub-flow calls
- `try`/`catch`, typed errors, retry policy at the call site
- `parallel` blocks, fan-out/fan-in over collections
- triggers (`on cron`, `on webhook`), waits and timers
- imports across files, module version pinning
- string interpolation and template values

### 4.3 non-functional

- adding a module requires zero changes to lexer, parser, AST, or VM
- VM core has no imports of `os`, `net`, `time`, or anything else non-deterministic; enforced by a lint rule in CI
- all VM state is serializable to protobuf at any suspend point
- compile of a 1k-line file well under 100ms ⚠️ unknown: no real perf target has been stated
- every diagnostic carries file, line, col, and a fix hint
- deny by default: a flow can touch nothing it did not declare
- replay of a journal reproduces the exact same state transitions

## 5. language and frontend design

### 5.1 v0 grammar

```ebnf
Program    = { Stmt } EOF .
Stmt       = ( Directive | ExprStmt | Empty ) Terminator .
Terminator = NEWLINE | EOF .

Directive  = "Akha" "." "Allow" "(" GlobList ")" .
GlobList   = Glob { "," Glob } .
Glob       = IDENT "..." | Qualified .

ExprStmt   = Expr .
Expr       = Call | Qualified | Literal .
Call       = Qualified "(" [ ArgList ] ")" .
ArgList    = Arg { "," Arg } [ "," ] .
Arg        = NamedArg | Expr .
NamedArg   = IDENT "=" Expr .

Qualified  = IDENT { "." IDENT } .
Literal    = STRING | INT | FLOAT | BOOL | NULL .
```

Reserved now, parsed later: `if else for in while fn return let try catch throw use flow on parallel await true false null`.

### 5.2 lexer

Hand-written, rune-based, single pass, no regex.

- emits `Token{Kind, Lit, Pos{File, Offset, Line, Col}}`
- significant newlines: emits `NEWLINE`, suppressed inside unclosed `(`/`[`/`{` so calls can wrap
- comments: `//` to end of line, `/* */` block. Attached to the following node as leading trivia so `fmt` and round-tripping work.
- strings: double-quoted with escapes, raw backtick strings, interpolation scanned into `STRING_PART` + `INTERP_START`/`INTERP_END` so the parser builds a concat node
- numbers: int, float, underscores as separators, and unit suffixes (`5s`, `10m`, `2h`) lexed as `DURATION`
- `...` is a single `ELLIPSIS` token, scanned before `.`
- never panics: on a bad rune it emits `ILLEGAL`, records a diagnostic, and skips to resync

Interface, per the interface-first convention:

```go
type Lexer interface {
    Next() Token
    Peek() Token
    Diagnostics() []diag.Diagnostic
}

var _ Lexer = (*lexer)(nil)
```

### 5.3 parser

Recursive descent for statements, Pratt (precedence climbing) for expressions. Pratt is chosen because adding an operator is one table entry, which matters when the operator set grows in v1.

Binding powers, loosest to tightest:
- `||`
- `&&`
- `== != < <= > >=`
- `+ -`
- `* / %`
- unary `! -`
- `?` (error propagation, postfix)
- call `()`, index `[]`, member `.`

Error recovery: panic-mode. On an unexpected token, record a diagnostic and skip to the next `NEWLINE` or closing brace, then continue. One bad line must not produce fifty errors.

AST design notes:
- every node carries `Pos` and `End`
- `CallExpr` keeps `Args []Arg` where `Arg` has an optional `Name`, so named args survive to the resolver
- trivia (comments, blank lines) is attached to nodes, enabling `fmt` and graph round-tripping
- the AST is syntax only. There is no `FSReadFileNode`. Ever.

```go
type Node interface {
    Pos() token.Pos
    End() token.Pos
}

type Visitor interface {
    Visit(Node) Visitor
}
```

### 5.4 resolver and validator

The first phase that knows modules exist.

- collects `Akha.Allow(...)` directives into a capability set, erroring if one appears after the first call
- resolves each `Qualified` head against the module registry
- checks: namespace known, function known, arity, named params exist, no duplicate named arg, positional args precede named
- checks capability: calling `FS.ReadFile` without `FS` allowed is a compile error, not a runtime error
- coarse type check of literal args against the declared signature
- constant folds literal-only expressions
- assigns a stable `CallSiteID` (hash of file, path, and index) used by the journal, traces, and step caching

Output: typed AST plus `[]Diagnostic` plus a `Manifest` of required capabilities and module versions.

### 5.5 why this makes modules cheap to add

Adding an `HTTP` module is:

1. write the Go type implementing `module.Module`
2. declare its functions in a manifest (protobuf)
3. register it

No frontend change, no VM change, no new opcode. The `CALL` opcode is generic over the module table. This is the direct answer to the first key feature.

## 6. flow IR

Defined in protobuf so it is serializable, versioned, language-neutral, and usable by the UI. Fits the `_defs/proto` convention.

Contents:
- header: IR version, source hash, compiler version
- constant pool: all literals, deduplicated
- module table: namespace, name, version, resolved module id
- call site table: `CallSiteID`, source position, retry policy, capability required
- instruction list: flat, addressable by program counter
- graph metadata: nodes and edges derived from the instruction list for visualization

Instruction sketch (stack machine, chosen for simplicity of lowering; a register machine is faster but harder to checkpoint-debug):

```
CONST   k          push constant
LOAD    slot       push local
STORE   slot       pop into local
LIST    n          build list from n stack values
MAP     n          build map from 2n stack values
CALL    site argc  pop args, yield Effect, suspend
JMP     addr
JMPF    addr       pop, jump if false
ITER    slot       begin iteration
NEXT    addr       advance or jump when exhausted
FRAME   fn argc    call a user fn
RET
TRAP    code       abort with a fault
```

`CALL` is the single point where the outside world is reachable. That is what makes the sandbox auditable.

## 7. flow execution engine

### 7.1 components

- Engine: owns configuration, module registry, storage, clock. Entry point.
- Scheduler: holds runnable flow instances, picks the next one, enforces global concurrency limits
- Instance: one run. Holds VM state, journal handle, trace context, budget.
- VM core: pure step function. No I/O, no clock, no goroutines.
- Effect host: executes effects against modules, applies timeouts and retries, runs on a worker pool
- Journal: append-only record of effect requests and results. The source of truth for replay.
- Sinks: log sink, trace sink, metric sink

```go
type VM interface {
    Step(ctx context.Context, st *State) (Outcome, error)
}

type EffectHost interface {
    Execute(ctx context.Context, req *Effect) (*Result, error)
}

type Journal interface {
    Append(ctx context.Context, e Entry) error
    Replay(ctx context.Context, runID string) (iter.Seq[Entry], error)
}

type Module interface {
    Manifest() *pb.ModuleManifest
    Invoke(ctx context.Context, fn string, args Args) (Value, error)
}
```

### 7.2 the run loop

```
load IR
open or replay journal
loop:
    outcome = vm.Step(state)
    switch outcome:
      Continue: continue
      Yield(effect):
          if journal has a recorded result for this effect: use it (replay)
          else: result = effectHost.Execute(effect); journal.Append(request, result)
          state.Resume(result)
      Suspend(wakeAt): persist state; deschedule; return
      Done(value): persist terminal record; emit receipt
      Fault(code): unwind; run handlers; persist failure
    enforce budget (steps, deadline, memory, calls)
```

Two properties fall out of this loop:
- replay skips already-executed effects, so crash recovery never double-sends an email
- the VM can be single-threaded and deterministic while effects run concurrently on a pool

### 7.3 values

Dynamic, immutable, serializable. `Null, Bool, Int, Float, String, Bytes, Duration, Time, List, Map, Error, Secret, Handle`.

- `Secret`: prints as `***` in logs, traces, and error messages. Comparison is constant time.
- `Handle`: an opaque id into a runtime-owned resource table (open file, http session). Scripts can pass handles around but never dereference them. Handles are closed when the frame that created them exits.

### 7.4 errors, faults, retries

Two distinct kinds, and conflating them is the usual mistake:

- error: a value. A module says "file not found". Flows can catch it.
- fault: a VM trap. Budget exceeded, capability denied, type error, journal corrupted. Not catchable by scripts.

Retry policy resolution order, first match wins:
1. call site args (`retries=5, backoff="exp"`)
2. module function default in the manifest
3. flow-level default
4. engine default

Retries apply at the effect level, use exponential backoff with jitter, and honor an idempotency key derived from `CallSiteID` plus an arg hash. Modules declare whether a function is idempotent; non-idempotent functions are not retried automatically.

### 7.5 tracing and logging

- one OTel span per call site; `trace_id` equals `run_id`
- span attributes: module, function, call site id, attempt number, duration, error class
- structured log record per step, correlated by run id and call site id
- `Akha.Log` goes to the log sink, `Akha.Debug` to the debug sink, per `lang.spec`. ⚠️ unknown: the spec says `Akha.Log` writes to db, the task example says it writes `akha.log` in the current dir. Sink abstraction resolves this once the intended default is confirmed.
- the journal is the replayable timeline, and is what a future UI renders

### 7.6 sandbox

Layered, because Go has no native sandbox:

1. language: scripts cannot express I/O. Only module calls reach outside.
2. capability: deny by default. `Akha.Allow` is the only grant, checked at compile time and re-checked at each effect.
3. resource: per-run budgets for steps, wall clock, memory, effect count, and per-module quotas
4. filesystem jail: a run gets a root directory; path args are cleaned and verified to stay inside it
5. network: per-flow egress allowlist
6. code isolation for third-party modules:
   - in-process: fastest, zero isolation. First-party only.
   - subprocess over gRPC (`hashicorp/go-plugin` style): process isolation, OS-level limits, moderate cost
   - wasm via wazero: strongest isolation, pure Go, no cgo. Recommended for third-party modules.

⚠️ unknown: whether untrusted `.ak` authors are in the threat model. Question 25 gates how much of layer 6 is needed in v1.

## 8. worked example

Source:

```
Akha.Allow(FS...)
Akha.Log("Hello World")
```

Tokens:

```
IDENT(Akha) DOT IDENT(Allow) LPAREN IDENT(FS) ELLIPSIS RPAREN NEWLINE
IDENT(Akha) DOT IDENT(Log) LPAREN STRING("Hello World") RPAREN NEWLINE EOF
```

AST:

```
Program
  AllowDirective  globs=[NamespaceGlob(FS)]
  ExprStmt
    CallExpr  callee=Qualified(Akha.Log)
      args=[ Arg{Value: StringLit("Hello World")} ]
```

After resolve:
- capability set `{FS, Akha}`
- `Akha.Log` resolves to module `akha@1`, function `Log`, arity 1, param `message: string`, effectful, non-idempotent, capability `Akha`
- call site id `ab12cd34`

IR:

```
consts: [0]="Hello World"
sites:  [0]=akha@1.Log  pos=2:1  retry=none
code:
  0000  CONST 0
  0001  CALL  site=0 argc=1
  0002  RET
```

Execution:

```
step 0: CONST      -> stack ["Hello World"]
step 1: CALL       -> Yield Effect{site:0, module:akha@1, fn:Log, args:["Hello World"]}
        journal    <- request
        host       -> akha module writes to the configured log sink
        journal    <- result{ok, 0.4ms}
        resume     -> stack [null]
step 2: RET        -> Done
```

Run record: 1 effect, 0 retries, 0 faults, one span named `akha.Log`.

## 9. package layout

Follows the conventions in `../../CLAUDE.md`, including `_defs` for generated types.

```
/cmd/akha                 cobra cli: run, check, fmt, graph, modules
/lang
  /token                  token kinds, positions
  /lexer                  Lexer
  /ast                    nodes, visitor, trivia
  /parser                 recursive descent + Pratt
  /diag                   diagnostics, rendering
/sema
  /resolve                registry binding, capability collection
  /check                  arity, names, coarse types
/ir
  /lower                  typed AST -> Flow IR
  /graph                  IR -> node/edge view
/vm
  /value                  Value kinds
  /machine                Step, frames, stack
  /budget                 metering
/runtime
  /engine                 Engine, Instance
  /sched                  scheduler
  /effect                 EffectHost, retry, idempotency
  /journal                WAL, replay
  /sink                   log, debug, trace sinks
/module
  /registry               Registry, manifest validation
  /modules/akha           Log, Debug, Exit, Allow, LogSetup
  /modules/fs             ReadFile, WriteFile, UpdateFile, ListFiles
/_defs
  /proto                  ir.proto, module.proto, effect.proto, run.proto
  /db                     schema.sql, q.runs.sql, q.journal.sql
  /graphql                flows, runs, traces
/sdk
  /proto-sdk              generated
  /db-sdk                 generated
/api/graphql              server
```

Every concrete type asserts its interface at the package level per the project convention.

## 10. build steps

### phase 0 - decide

- answer section 3, at minimum questions 1, 3, 5, 12, 19, 25
- lock the v0 grammar
- write 20 example `.ak` files covering the intended v1 feature set, before writing the lexer. They become the golden test corpus.

### phase 1 - frontend

- token kinds and positions
- lexer plus table tests using the `tu.Case` pattern, one banner per behaviour
- AST nodes and visitor
- recursive descent parser for the v0 grammar, with Pratt scaffolding present but unused
- diagnostics with source snippets and carets
- `akha check` and `akha fmt`
- exit criteria: all 20 corpus files parse or fail with a precise, correct message

### phase 2 - modules and resolution

- `module.Module` interface, manifest in protobuf, registry
- implement `akha` and `fs` modules per `lang.spec`
- resolver, capability collection, validator
- `akha modules` lists everything registered with signatures
- exit criteria: a new module is added with zero changes outside `/module`

### phase 3 - IR and VM

- `ir.proto`, constant pool, call sites, instructions
- lowering for v0 (constants and calls only)
- VM step function, frames, stack, budget
- effect yielding, in-memory effect host, no journal yet
- exit criteria: `Akha.Log("Hello World")` runs end to end

### phase 4 - engine

- journal, replay, crash recovery
- retries, backoff, idempotency keys
- error vs fault separation, unwinding
- log/debug/trace sinks, OTel spans
- run records persisted via sqlc
- exit criteria: kill the process mid-run, restart, and the run completes without repeating a completed effect

### phase 5 - language v1

- `:=` bindings, expression evaluation, operators via the Pratt table
- `if`/`else`, `for ... in`
- `try`/`catch`, call-site retry args
- string interpolation
- exit criteria: a real flow, for example read a directory, filter, transform, write a report

### phase 6 - composition and scale

- user `fn` and sub-flows, frames and returns
- `parallel` blocks, fan-out/fan-in through the scheduler
- triggers, waits, timers, suspend and resume across processes
- distributed workers: the effect host becomes an RPC boundary
- exit criteria: a flow that sleeps an hour survives a deploy

### phase 7 - platform

- GraphQL API for flows, runs, journals
- graph view from IR metadata
- third-party modules under wazero
- tree-sitter grammar and LSP

## 11. open items

- ⚠️ unknown: canonical spec path, `__arch/v1/lang.spec` versus `../../_research/v1/lang.spec`
- ⚠️ unknown: whether `Akha.Log` targets a db or `./akha.log` by default
- ⚠️ unknown: threat model for untrusted flow source
- ⚠️ unknown: whether a visual editor and round-trip graph editing are in scope
- ⚠️ unknown: performance targets for compile and for steps per second
- ⚠️ unknown: relationship to the "distributed build systems" goal in the project overview, which may promote caching and content addressing to v1

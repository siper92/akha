# CLAUDE.md

## Project Overview

akha: a platform for AI workflows, similar to n8n and temporal, written in Go
- a worker runs `.ak` scripts with an own interpreter
- a backend issues JWT tokens to workers over gRPC
- a backend can start workers as goroutines with backend privileges

tech stack: go 1.27, cobra, viper, grpc, sqlite, log/slog
generate: sqlc, protobuf, graphql (deferred)

## structure

- `cmd/akha/` - cobra root, `backend` and `worker` sub commands
- `internal/config/` - viper loading, typed structs for `config.yaml` and `config.be.yaml`
- `internal/cache/` - `Cache` interface, file impl under `_env/.cache/*`
- `internal/tu/` - test utils, `tu.Case`, `tu.Run`
- `lang/token/` - token kinds and positions
- `lang/lexer/` - rune based lexer
- `lang/ast/` - nodes
- `lang/parser/` - recursive descent parser
- `lang/check/` - static checks (Ak.Allow, arity, kwargs)
- `lang/eval/` - values, registry, evaluator
- `lang/module/ak/` - `Ak` namespace
- `lang/module/fs/` - `FS` namespace, sandboxed root
- `lang/runner/` - lex, parse, check, eval from a source string
- `platform/backend` - token issuer, verifier, access token store, login log
- `platform/backend` - gRPC server, interceptors
- `platform/worker` - gRPC client, login, token cache
- `platform/worker` - worker service, embedded mode
- `_defs/` - proto, sqlc and graphql definitions
- `platform/sdk` - generated code only
- `docker/` - backend and worker images, compose file, container configs, sample scripts
- `.claude/skills/` - project skills, see `ak-lang-feature`
- `__arch/v1/` - spec, plan, questions and run log for the current version

## language

- see `__arch/v1/lang.spec` for the `.ak` spec, it is the source of truth
- statements: call, `let`, assignment, `if`/`else if`/`else`, `for x in list`, `while`, `break`, `continue`
- expressions: literals, variable references, calls, `+ - * / %`, comparisons, `and or not`, grouping
- blocks use braces, `{` on the header line and `}` on its own line
- `Ak.Allow(FS...)` must be the first statement and declares the used modules
- `Ak` is always allowed
- static check on every execution (allow rules, arity, kwargs, scopes, loop keywords), then evaluate

## skills

- `ak-lang-feature` (`.claude/skills/ak-lang-feature/SKILL.md`) - how to add a language feature
  - spec first, then data definitions (token, ast, eval), then table tests, then one layer at a time
  - carries the shared api sheet for ast nodes, `eval.Env`, `eval.Binary` and `eval.Unary`
  - use it for any change to `lang/`

## config

- `config.yaml` - worker: `backend`, `worker_access_token`, `cache_dir`, `root`
- `config.be.yaml` - backend: `addr`, `db`, `cache_dir`, `workers`, `access_tokens`, `jwt`
- JWT keys and ttl live only in the backend config

## commands

- `just build` - build `./bin/akha`
- `just test` - run all tests
- `just test-lang` - run language tests only
- `just lint` - vet and gofmt
- `just run-backend` - start the backend
- `just run-ak file.ak` - run a script on the worker
- `just check-ak file.ak` - static check a script
- `just gen` - clear dsk and regenerate it in the `platform/sdk` from `_defs/`
- `just img-build` - build the backend and worker images from `docker/`
- `just img-up` - start the backend container
- `just img-run-ak file.ak` - run a script from `docker/scripts/` in a worker container
- `just img-down` - stop the containers and drop their volumes

## Conventions

### major patterns
- interface first; every concrete type asserted via `var _ I = (*Impl)(nil)`
- interfaces live in `<pkg>/<pkg>.go`, impls next to them
- factories sit next to the type they build; one allocation site per shape (`lexer.New`, `parser.New`, `eval.NewRegistry`)
- sentinel errors plus `fmt.Errorf("%w")`, no error framework
- `context.Context` first argument everywhere
- don't write comments unless specified in the task
- define interfaces when possible

### major test patterns
- table tests use the multi-line `tu.Case` literal, grouped under a `// ---` section banner per behaviour:
  ```go
  cases := []tu.Case[Kind, int]{
      {
          Name:     "kind_scalar_is_zero",
          Input:    KindScalar,
          Expected: 0,
      },
  }
  tu.Run(tu.New(t), cases, fn, nil)
  ```
- `tu.Run` skips the value check when `Case.Err` is set; slice error types use a result struct instead
- lexer and parser use golden `.ak` files under `testdata/`
- eval tests use a fake `FS` root in `t.TempDir()`
- auth tests use in memory sqlite and `bufconn`

### Code Generation - types are defined and code is generated
1. Define types in protobuf, sqlc, graphql files
 - structure: `_defs/`
   - `db/` is for sqlc, `proto/` is for protobuf and `graphql/` is for GraphQL
 - generate all types with `just gen` command

**`Generated` code:**
- Protobuf output: `platform/sdk`
- SQLC output: `platform/sdk`
- GraphQL server: `api/graphql/*` (deferred)

## Examples
 - __local/starlark-go - a starlark interpreter written in Go, used for inspiration and examples

# do
- use simple formats for MD files
   - list, titles, sections and code blocks
   - no tables, no fancy styling
   - minimal emojis usage allow only if specified in the task
- log agent work in `__arch/v1/run/log.md` when a task asks for a run log

# don't
- ever use the '—' symbols use '-' for lists and '--' for flags
- assume - mark as "unknown" and use emoji's (warning sign)
- read files 2 times
- read files you have created in this session
- write long descriptions
- use bash for reading, searching, writing, or other file operations
   - use Read, Write, Search, Glob, Grep tools
- run bash to re-check signatures of files already read, act on what is known
- read examples unless specified in the task
- read contents of ./_env/_arch folder or ./_env/examples folder
- read __arch folder or examples folder unless the task points to it
- run commands - use the `just` command only when not specified in the task

!!! don't are valid unless specified in the task
!!! very important don't read ./_env/* forbidden folders

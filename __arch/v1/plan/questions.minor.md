# Akha - minor questions (AI implementation)

Concrete questions an AI agent needs answered (or defaulted) before writing code.
Each has a proposed default. If a major question overrides it, follow the major answer.

## conventions

- module path: `github.com/siper92/akha` - default: keep
- go version: 1.27 - default: keep, use `iter`, generics, `log/slog`, `context` everywhere
- CLI: cobra + viper - default: one binary `akha` with sub commands `backend`, `worker`, `run`, `check`
- interfaces: every package exposes interfaces in `<pkg>/<pkg>.go`, impls asserted with `var _ I = (*Impl)(nil)` - default: yes
- factories: `New...` next to the type - default: yes
- errors: sentinel errors + `fmt.Errorf("%w")` - default: yes, no custom error framework
- tests: `tu.Case` table tests - where is `tu` defined? default: create `internal/tu` if missing

## layout

- default package tree:
  - `cmd/akha/` - main + cobra root
  - `internal/config/` - viper loading, typed config struct
  - `lang/token`, `lang/lexer`, `lang/ast`, `lang/parser`, `lang/eval`
  - `lang/module/` - `Ak`, `FS`, later `AI`
  - `../../../platform/backend`, `../../../platform/backend`, `../../../platform/backend`, `../../../platform/backend`, `../../../platform/backend`
  - `../../../platform/worker`, `../../../platform/worker`, `../../../platform/worker`
- is `internal/` allowed or must all be public? default: `internal/` for non sdk code

## language

- lexer: hand written or generated? default: hand written, rune based, positions (line, col)
- tokens needed from lang.spec: IDENT, STRING, INT, `.`, `(`, `)`, `,`, `=`, `...`, comments `//`, newline
- trailing comma allowed in calls - yes (spec)
- keyword args (`code=0`, `log="info.log"`) - yes, after positional args only?  default: yes
- variadic spread (`FS...`) inside `Ak.Allow` only or general? default: only in `Ak.Allow`
- are there variables, if, loops, functions in v1? default: no, calls only, add later
- evaluation: tree walking interpreter over AST - default: yes
- module interface default:
  - `Module { Name() string; Func(name string) (Builtin, bool) }`
  - `Builtin func(ctx context.Context, args []Value, kwargs map[string]Value) (Value, error)`
- value types in v1: string, int, bool, list, none - default: yes
- calling a module not passed to `Ak.Allow` - default: runtime error before execution (static check pass)
- `Ak.Exit("error", code=0)` - does code 0 with a message mean failure? default: exit code decides, message logged
- `FS.UpdateFile` vs `FS.WriteFile` - difference? default: Update fails if file does not exist, Write creates
- `FS.ListFiles` recursive? default: no, non recursive, returns list of strings
- paths: relative to script dir or worker work dir? default: worker work dir, sandboxed

## worker

- sandbox: all FS paths resolved and checked inside a root dir - default: yes, reject `..` escape and symlinks out
- concurrency: how many workflows at once per worker? default: configurable, default 1
- work source: poll `backend.NextJob` via gRPC stream - default: server stream `Subscribe`
- `Ak.Setup(log=..., debug=...)` files - local only or also streamed to backend? default: both
- cache dir location for `Ak.Debug` - default: `$XDG_CACHE_HOME/akha/<run-id>`

## backend

- JWT lib: `github.com/golang-jwt/jwt/v5` - default: yes, EdDSA keys
- token flow default:
  - worker sends `worker_access_token` (config.yaml) -> backend returns short JWT (15m) + refresh token
  - gRPC unary + stream interceptors validate JWT
- existing `access_tokens` table: store refresh tokens or worker access tokens? default: worker access tokens, hashed (not plain text)
- sqlite driver: `modernc.org/sqlite` (no cgo) - default: yes
- migrations: goose or plain `schema.sql`? default: goose, keep `schema.sql` as sqlc input
- LLM interface default:
  - `Provider { Name() string; Complete(ctx, Request) (Response, error); Stream(ctx, Request) (iter.Seq2[Chunk, error]) }`
- first provider: Anthropic Go SDK, model from config
- templates: go `text/template` stored in db table `prompt_templates` - default: yes
- logs: table `run_logs` or files? default: table, with level, run_id, ts, message

## proto / gRPC

- package name `proto_sdk` - keep? default: keep
- services default:
  - `AuthService`: `Register`, `Login`, `Refresh`, `ValidateToken`
  - `JobService`: `Subscribe` (stream), `ReportStatus`, `PushLogs` (client stream)
  - `LLMService`: `Complete`, `Stream`
- proto files split per service in `_defs/proto/` - default: yes
- add `buf` instead of raw protoc? default: no, keep `just gen-proto`

## db (sqlc)

- tables to add default: `workers`, `workflows`, `runs`, `run_logs`, `prompt_templates`
- ids: integer autoincrement or uuid/ulid text? default: ulid text for runs, integer for rest
- timestamps: UTC, `TIMESTAMP` - default: yes
- query files: one `q.<table>.sql` per table - default: yes

## graphql

- needed in v1? default: no, defer; keep `_defs/graphql/` empty
- if yes, generator: gqlgen - default: gqlgen, output `api/graphql/`

## testing

- lexer / parser: golden tests with `.ak` fixtures in `testdata/` - default: yes
- backend: in memory sqlite per test - default: yes
- gRPC: `bufconn` for in process tests - default: yes
- e2e: backend + worker in one test process - default: yes, one happy path

## unknown

- ⚠️ what `lang.spec` means by "logs a debug message to - file/cache directory" - unknown, default cache dir
- ⚠️ exact semantics of `Ak.Allow(FS...)` spread syntax - unknown, treated as allow whole namespace
- ⚠️ purpose of `README.md` "right flow" and product naming - unknown

# Akha - implementation plan (generated)

Derived from `plan.md`, `questions.major.md` (answered) and `questions.minor.md` (defaults).
Major answers override minor defaults.

## v1 in one line

A worker runs a `.ak` script with an own interpreter. 
A backend issues JWT tokens to workers over gRPC. 
A backend can start workers as goroutines with backend privileges.
Nothing else.

## decisions locked by the major answers

- v1 is language + interpreter first, minimal features, ergonomics over scope
- own interpreter, starlark used only as design reference
- static check on every execution, then evaluate
- `Ak.Allow` declares which modules the script uses, all modules exist, undeclared use is an error
- interpreter has an interface and a runner that accepts a string source
- backend only issues and validates JWT, no jobs, no LLM, no GraphQL, no UI
  - the only job allowed is to execute ak scripts as a go rutin
- 2 tiers of tokens: backend tier and worker tier
- backend can start workers as goroutines with backend privileges
- login: static access token from config -> short lived JWT, no refresh, re-login on expiry
- keys and JWT lifetime live in `config.be.yaml`
- every login attempt is logged
- gRPC middleware validates JWT on every worker call
- tokens are files in a cache dir behind a `Cache` interface, used by worker and backend
- sqlite for both, large outputs under `.cache/*`
- CLI: `akha backend ...`, `akha worker ...`
- viper for config, `log/slog` for logs, no OpenTelemetry yet
- unit tests for lexer, parser, evaluator; integration later
- workers and backends are replaced, never upgraded in place

## out of scope for v1

only take into account the posable future, but do not implement yet:
- AI / LLM namespace and `LLMService`
- job queue, run history tables, log streaming
- GraphQL, web UI
- HTTP, Shell, Git, Env namespaces
- variables, conditions, loops in the language
- postgres

## package layout

```
cmd/akha/                 cobra root, backend and worker sub commands
internal/config/          viper loading for config.yaml, config.be.yaml
internal/cache/           Cache interface + file impl (.cache/*)
internal/tu/              test utils, tu.Case, tu.Run
lang/token/               token kinds, positions
lang/lexer/               rune based lexer
lang/ast/                 nodes
lang/parser/              recursive descent parser
lang/check/               static checks (syntax, Ak.Allow, arity, kwargs)
lang/eval/                tree walking evaluator, module registry, values
lang/module/ak/           Ak namespace
lang/module/fs/           FS namespace
lang/runner/              Runner interface, runs source string end to end
backend/auth/             TokenIssuer, TokenVerifier, login log
backend/server/           gRPC server, interceptors, AuthService impl
worker/client/            gRPC client, login, token cache
worker/                   worker service, embedded mode for backend goroutines
_defs/proto/auth.proto    Register, ValidateToken
_defs/db/                 schema + queries (backend: login_attempts, workers)
sdk/                      generated only
```

## interfaces (names and responsibility only)

### lang

- `token.Kind`, `token.Token`, `token.Pos`
- `lexer.Lexer` - `Next() token.Token`, `All() iter.Seq[token.Token]`
- `parser.Parser` - `Parse() (*ast.Script, error)`
- `ast.Node`, `ast.Script`, `ast.Call`, `ast.Arg`, `ast.Literal`, `ast.Selector`, `ast.Spread`
- `check.Checker` - `Check(*ast.Script, Registry) []Diagnostic`
- `eval.Value` - string, int, bool, list, none
- `eval.Module` - `Name() string`, `Func(name string) (Builtin, bool)`
- `eval.Builtin` - `Call(ctx, args []Value, kwargs map[string]Value) (Value, error)`
- `eval.Registry` - `Register(Module)`, `Lookup(name) (Module, bool)`
- `eval.Evaluator` - `Eval(ctx, *ast.Script) error`
- `runner.Runner` - `Run(ctx, src string, opts) (Result, error)`

### backend

- `auth.TokenIssuer` - `Issue(ctx, subject, tier) (jwt string, error)`
- `auth.TokenVerifier` - `Verify(ctx, jwt string) (Claims, error)`
- `auth.AccessTokenStore` - `Lookup(ctx, hash) (Worker, bool, error)`
- `auth.LoginLog` - `Record(ctx, attempt) error`
- `server.Server` - `Serve(ctx) error`, wraps gRPC

### shared

- `cache.Cache` - `Get(key) ([]byte, bool, error)`, `Put(key, []byte) error`, `Del(key) error`
- `config.Loader` - `Load(path) (Config, error)`

### worker

- `client.Backend` - `Login(ctx) (jwt, error)`, `Validate(ctx, jwt) (bool, error)`
- `worker.Worker` - `Start(ctx) error`, `Run(ctx, src) (Result, error)`

Every impl: `var _ Iface = (*Impl)(nil)`, one `New...` factory next to the type.

## phases

### P0 - skeleton and ergonomics

- `cmd/akha` root with `backend` and `worker` groups, `--config` flag
- `internal/config` typed structs for `config.yaml` (worker) and `config.be.yaml` (backend)
- `internal/tu` with `Case[I, E]` and `Run`
- Justfile: `test`, `lint`, `run-backend`, `run-worker`, `run-ak file`
- `.cache/` in `.gitignore`
- done: `go build ./...` and `just test` pass with an empty test

### P1 - tokens and lexer

- token kinds: IDENT, STRING, INT, DOT, LPAREN, RPAREN, COMMA, ASSIGN, ELLIPSIS, NEWLINE, COMMENT, EOF, ILLEGAL
- lexer over runes with line and col, `//` comments dropped or kept as tokens (default dropped)
- tests: `lang/lexer/testdata/*.ak` with expected scripts
- done: `lang.spec` sample lexes without ILLEGAL

### P2 - ast and parser

- nodes: `Script{Calls}`, `Call{Target Selector, Args, Kwargs}`, `Selector{Module, Name}`, `Arg{Value}`, `Kwarg{Name, Value}`, `Spread{Selector}`
- rules: positional before keyword args, trailing comma allowed, `X...` only as positional arg
- errors carry `token.Pos`, parser recovers to next NEWLINE
- tests for each rule, golden tests for full scripts
- done: `lang.spec` sample parses to expected AST

### P3 - static check

- `Ak.Allow` must be the first call, collects allowed modules, `Ak` is implicit
- every `Selector.Module` must be allowed and registered
  - can be validated even at parsing because `Ak.Allow` is the first call
- every `Selector.Name` must exist in the module
  - can be validated even at parsing because module is loaded and has exist check
- arity and kwarg names validated against a `Builtin` signature (`Spec()` on Builtin)
- diagnostics list with positions, never panics
- done: undeclared `FS` use produces one diagnostic with line and col


### P4 - evaluator and modules

key feature is the module definition interface and extension, 
- the `Ak` and `FS` modules are implemented in v1, but the design allows for more modules to be added later.

- values: string, int, bool, list, none with `String()` and type name
- registry with `Ak` and `FS` registered by the runner
- `Ak`: `Allow` (noop at eval), `Setup(log=, debug=)`, `Log`, `Debug`, `Exit(msg, code=0)`
- `Exit` returns a typed error carrying code, runner maps it to process exit
- `FS`: `ReadFile`, `WriteFile`, `UpdateFile` (fails if missing), `ListFiles` (non recursive)
- `FS` takes a root dir and rejects `..` escapes and symlinks out of root
- `Ak.Debug` writes to `.cache/akha/<run-id>/`
- done: lang.spec sample runs and writes the log files

### P5 - runner and cli

- `runner.New(registry, opts)` - lex, parse, check, eval from a string
- `Result{ExitCode, Diagnostics, LogPath, DebugPath}`
- `akha worker run file.ak` and `akha worker check file.ak`
- done: `akha worker run` on a sample prints logs and exits with the script code

### P6 - backend auth

- `auth.proto`: `Register(RegisterRequest{access_token}) -> RegisterResponse{jwt, expires_at}`, keep `ValidateToken`
- `just gen` regenerates `sdk/proto-sdk`
- db: `workers` (id, name, token_hash, tier, created_at), `login_attempts` (id, worker_id nullable, ok, reason, at)
- `AccessTokenStore` compares sha256 of the config access token
- `TokenIssuer` with `golang-jwt/jwt/v5`, EdDSA, key pair path and ttl from `config.be.yaml`
  - token must validate against the BE db and see if the access token is there
  - provide seed data for a worker with a known access token for testing
- unary interceptor: skip `Register`, verify JWT on everything else, claims into context
- every `Register` call recorded in `login_attempts`
- `akha backend serve`
- done: `bufconn` test - bad token rejected and logged, good token gets JWT, expired JWT rejected

### P7 - worker client

- `client.New(addr, accessToken, cache)`
- `Login` stores JWT in `cache` under `jwt/<backend-addr>`
- on expired or missing JWT call `Register` again, no refresh
- `akha worker login`, `akha worker whoami` (validates cached token)
- done: worker against a live backend logs in, caches, re-logs after expiry
  - if token is not found always login, worker requires token to run or parse scripts, no token means no work

### P8 - embedded workers

- backend config `workers: n` starts `n` workers as goroutines
- embedded worker gets a backend tier JWT issued in process, no network login
- shared `worker.Worker` type for both modes, only `client.Backend` impl differs
- done: `akha backend serve` with `workers: 1` runs a sample script from a flag

## test plan

- lexer and parser: golden files under `testdata/`, `tu.Case` tables for edge rules
- check: one case per diagnostic kind
- eval: fake `FS` root in `t.TempDir()`
- auth: in memory sqlite per test, `bufconn` gRPC
- no integration suite in v1, a single smoke test in P8 is enough

## open items

- ⚠️ `Ak.Debug` "file/cache directory" - default `.cache/akha/<run-id>/debug.log`
- ⚠️ `FS...` spread outside `Ak.Allow` - default parse error
- ⚠️ worker tier vs backend tier permissions beyond token claims - unknown, only a `tier` claim in v1
- ⚠️ `config.be.yaml` file name - taken from major answers, confirm

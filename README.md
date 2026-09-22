# Akha

Akha is a platform for AI workflows, similar to n8n and temporal, written in Go.

Workflows are `.ak` scripts. A worker runs them with an own interpreter.
A backend authenticates workers and issues JWT tokens over gRPC.

## parts

- backend
  - issues short lived JWT tokens to workers
  - validates tokens on every gRPC call
  - logs every login attempt
  - can start its own workers as goroutines
- worker
  - lexes, parses, checks and runs `.ak` scripts
  - works with files inside a sandboxed root
  - logs in to the backend with a static access token

## the .ak language

A script is a list of calls. The first call must declare the modules it uses.
`Ak` is always available.

```
Ak.Allow(FS...)
Ak.Setup(
    log="info.log",
    debug="debug.log",
)

Ak.Log("hello")
Ak.Debug("dbg", 1, "two")

FS.WriteFile("file.txt", "content")
FS.ReadFile("file.txt")
FS.UpdateFile("file.txt", "new content")
FS.ListFiles("path/")

Ak.Exit("done", code=0)
```

Modules in v1:

- `Ak` - `Allow`, `Setup`, `Log`, `Debug`, `Exit`
- `FS` - `ReadFile`, `WriteFile`, `UpdateFile`, `ListFiles`

Rules:

- positional arguments come before keyword arguments
- trailing commas are allowed
- `Module...` spread is allowed only inside `Ak.Allow`
- `//` starts a comment
- every script is statically checked before it runs

## requirements

- go 1.27
- just
- docker, only for code generation

## setup

Worker config `config.wk.yaml`:

```yaml
backend: "localhost:50051"
worker_access_token: "akha_420_99078"
cache_dir: ".cache/akha"
root: ".cache/akha/root"
```

Backend config `config.be.yaml`:

```yaml
addr: ":50051"
db: ".cache/akha/backend.db"
cache_dir: ".cache/akha"
workers: 0
access_tokens:
  - "akha_420_99078"
jwt:
  private_key: ".cache/akha/jwt.key"
  public_key: ".cache/akha/jwt.pub"
  ttl: "15m"
```

## usage

Build and test:

```
just build
just test
just lint
```

Run a script on the worker:

```
just check-ak script.ak
just run-ak script.ak
```

Start the backend and log in a worker:

```
just run-backend
just login
```

Run the integration flow, backend plus worker on `_env/examples/akha/hello.ak`:

```
just integration
```

The binary can be used directly:

```
./bin/akha backend serve --config config.be.yaml
./bin/akha worker run --config config.wk.yaml script.ak
./bin/akha worker check --config config.wk.yaml script.ak
./bin/akha worker login --config config.wk.yaml
./bin/akha worker whoami --config config.wk.yaml
```

Regenerate the sdk from `_defs/`:

```
just gen
```

## layout

- `cmd/akha/` - cli
- `lang/` - token, lexer, ast, parser, check, eval, modules, runner
- `backend/` - auth and gRPC server
- `worker/` - worker service and backend client
- `internal/` - config, cache, test utils
- `_defs/` - proto and sqlc definitions
- `sdk/` - generated code
- `__arch/v1/` - language spec and plan

Logs and large outputs go under `.cache/`.

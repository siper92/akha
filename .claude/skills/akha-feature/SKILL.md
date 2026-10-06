---
name: akha-feature
description: Add or change a feature of the akha language, Use for any change under lang/.
---

# akha-feature

## sources
- rules, source of truth: [spec_v1.ai.md](spec_v1.ai.md)
  - only rules the example cannot show: errors, limits, edge cases, unknowns
- syntax by example: [spec_def.ak](spec_def.ak)
  - one runnable flow step: imports, bindings, module calls, loops, operators, scopes, files, return
  - it runs in `lang/module/fs/spec_def_test.go` on the fs `testdata`, keep the expected output in sync
  - a new syntax rule goes here as one commented line, not into the spec
  - it is not the parser golden, `lang/parser/testdata/spec_def.ak` is updated on its own
- canonical golden: `lang/parser/testdata/spec_def.canonical.ak`

## execution plan
flow.ak > lexer(token) > parser(ast) > check > eval

## order
1. spec first
   - syntax change: a line in `spec_def.ak`
   - rule, error or limit: a short list item in `spec_v1.ai.md`
   - mark open points as unknown with a warning sign, do not assume
2. data definitions
   - token kind and name: `lang/lexer/token.go`
   - error code: `lang/lexer/error.go`, kebab case, `Code<Name>`
   - ast node: `lang/ast/ast.go`, embed `Pos`, add the `var _ Expr/Stmt` assert and the marker method
   - canonical print: `lang/ast/string.go`
3. table tests, few cases, one per behavior
4. one layer at a time: lexer, then parser, then ast string, then check, then eval

## file splitting
- `<pkg>.go` is the entry file of a package
  - name consts, sentinel errors, interfaces, `var _ I = (*impl)(nil)` asserts
  - the impl struct, its factories and shared helpers (eg `fs.do`)
- one file per concern next to it, named by the concern
  - lexer: `token.go`, `error.go`, `lexer.go`
  - check: `check.go` (interface), `checker.go`, `calls.go`, `values.go`, `error.go`
  - eval: `value.go`, `ops.go`, `env.go`, `codec.go`, `interp.go`, `error.go`
  - fs: `fs.go`, `path.go`, `text.go`, `json.go`, `yaml.go`
- shared module interfaces in `lang/module/module_ast.go`, their impls in `lang/module/module.go`
- errors of a package live in its `error.go`, or in `<pkg>.go` when there are few
- tests
  - one `_test.go` per feature, a new feature gets a new file
  - external `<pkg>_test` package, only tests, helpers go in `lang/tests_utils`
  - fixtures in the package `testdata/`

## module interface rules
- a module is a package under `lang/module/<name>`, registered in `lang/module/std.New`
- `<name>.go` defines
  - `Name` and one const per function name, the script name (`ReadJSON = "readJSON"`)
  - a Go interface listing every function as `Func(args ...module.IValue) (module.IValue, error)`
  - small interfaces for outside deps, asserted on the real type (`var _ Storage = (*os.Root)(nil)`)
  - an unexported impl struct embedding `*module.Module`, asserted on its interface and `module.IModule`
- factories
  - `New(...) (module.IModule, error)` returns the interface, validates its args
  - an extra `NewWith<Dep>` takes the dep for tests (`fs.NewWithStorage`)
  - both go through one private builder, the single allocation site
- functions are built once in the builder
  - a def table of name, params, impl, each through `module.NewFunc`, then `module.NewModule`
  - param types come from `lang/module/types`, named params need a `Default`
  - lazy deps use `sync.OnceValues`
- impls
  - receive args already bound and validated by `Func.Bind`
  - return `eval` values, `eval.Null{}` when nothing is returned
  - wrap sentinels with `%w`, prefix with the function name and the main arg
  - group impls by concern in their own file (`text.go`, `json.go`)
- ⚠️ unknown: `ak` builds its funcs on every `Func` call through a switch, not once in the factory

## parser rules
- recursive descent in `lang/parser/parser.go`
- fail with `p.fail(code, hint, format, args...)` or `p.failAt(pos, ...)`
- nesting goes through `p.enter()` / `p.leave()`
- `p.open()` / `p.close(kind)` for `( [ {` so newlines inside are skipped
- loop only rules use `p.loops`
- validate known values at parse level, max validation before eval

## tests
- `tests_utils.Case[I, E]` multi line literal, grouped by a `// ---` banner
- run cases in a loop with `t.Run(c.Name, ...)` and one `pipeline_test.Validate*` call
- parser
  - `validateParserOutput(t, c)`, `Expected` is the canonical output, 4 space indent, reparsed for stability
  - `Err` is matched with `strings.Contains`, write `error[code]: message`
  - error snippets can go to `lang/parser/testdata/errors.ak`
    - cases split by `---`, each has a `// ### error[code]: message` line
    - a 3 line header is prepended, so lines start at 4
  - a new behavior gets its own `TestParser<Feature>` function
- modules at parse level: `pipeline_test.ValidateParse`
- check: `pipeline_test.ValidateCheck(t, c)`, `Case[string, struct{}]`
  - `Err` is `line: error[code]: message`, must wrap `check.ErrCheck`
  - static rules in `lang/check/checker.go`, known value rules implement `ValueChecker` in `lang/check/values.go`
- eval: `pipeline_test.ValidateRun(t, c, input)`, `Case[string, string]`
  - `Expected` is the `String()` of the returned value
  - `Err` is `line: runtime error: message`
  - the runner checks first, a runtime error case must hide the value from the checker (`var`, not `let` or a literal)
- modules with files: `pipeline_test.ValidateRunFS(t, c, "testdata")`, copies testdata into a temp root
- error states are grouped in `TestCheckErrors` / `TestEvalErrors`, one case per error code

## current decisions
- `.` is the object member operator only, literals of other kinds are `member-kind` parse errors
- eval gets the registry via `eval.Options.Modules`
- call rules live in `check.CallChecker` (parse and check), import scope in `lang/check/checker.go`
- `fs`: root opened lazily once, paths cleaned in `path.go`, json and yaml decode through tokens / `yaml.Node` to keep key order
- `return` / `exit` always leave every loop, `break` / `continue` only the innermost

## don't
- write many tests, unless specified, no more than 4 test cases per feature

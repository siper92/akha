---
name: akha-feature
description: Add or change a feature of the akha language, Use for any change under lang/.
---

# akha-feature

## sources
- spec, source of truth: [spec_v1.ai.md](spec_v1.ai.md)
- spec examples: [spec_def.ak](spec_def.ak)
- arch folder with tasks and plans: [__arch](./__arch)
- eval plan: [plan.eval.md](plan.eval.md)
- canonical golden: `lang/parser/testdata/spec_def.canonical.ak`

## order
1. spec first
   - add the rule to `spec_v1.ai.md` in the matching section, short list items
   - mark open points as unknown with a warning sign, do not assume
2. data definitions
   - token kind and name: `lang/lexer/token.go`
   - error code: `lang/lexer/error.go`, kebab case, `Code<Name>`
   - ast node: `lang/ast/ast.go`, embed `Pos`, add the `var _ Expr/Stmt` assert and the marker method
   - canonical print: `lang/ast/string.go`
3. table tests, few cases, one per behavior
4. one layer at a time: lexer, then parser, then ast string

## parser rules
- recursive descent in `lang/parser/parser.go`
- fail with `p.fail(code, hint, format, args...)` or `p.failAt(pos, ...)`
- nesting goes through `p.enter()` / `p.leave()`
- `p.open()` / `p.close(kind)` for `( [ {` so newlines inside are skipped
- loop only rules use `p.loops`
- try to validate values at parse level if the value is known
  - the idea is max validation on parse level

## tests
- `tests_utils.Case[I, E]` multi line literal, grouped by a `// ---` banner
- parser cases run through `validateParserOutput(t, c)`
  - `Expected` is the canonical output, 4 space indent, reparsed for stability
  - `Err` is matched with `strings.Contains`, write `error[code]: message`
  - but defined in different funks
- error snippets can go to `lang/parser/testdata/errors.ak`
  - cases split by `---`, each has a `// ### error[code]: message` line
  - a 3 line header is prepended, so lines start at 4
- a new behavior gets its own `TestParser<Feature>` function
- keep it minimal: one happy case, one error case
- check cases run through `pipeline.ValidateCheck(t, c)`, `Case[string, struct{}]`
  - `Err` is `line: error[code]: message`, must wrap `check.ErrCheck`
  - static rules live in `lang/check/checker.go`, known value rules implement `ValueChecker` in `lang/check/values.go`
- eval cases run through `pipeline.ValidateRun(t, c, input)`, `Case[string, string]`
  - `Expected` is the `String()` of the returned value
  - `Err` is `line: runtime error: message`
  - the runner checks first, so a runtime error case must hide the value from the checker (use `var`, not `let` or a literal)
- error states are grouped in `TestCheckErrors` / `TestEvalErrors`, one case per error code

## current decisions
- `.` is the object member operator only, literals of other kinds are `member-kind` parse errors
- modules (`Module.member`) are not implemented in v1
- `return` / `exit` always leave every loop, `break` / `continue` only the innermost

## don't
- write many tests, unless specified, always define no more than 4 test cases per feature

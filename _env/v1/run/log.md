# akha v1 - run log

Date: 2026-09-22
Task: task.ai.md - implement __arch/v1/plan with 2 sub agents (impl, test)
Brief shared with agents: scratchpad/brief.md (7 steps)
Rule: no commands are run by agents, only `just gen` by the main agent at the end

## steps

- step 1 - token and lexer
- step 2 - parser
- step 3 - values, registry, static check
- step 4 - modules Ak and FS, evaluator
- step 5 - runner, cache, config, tu, cli
- step 6 - backend auth
- step 7 - worker client, worker, embedded workers

## schedule

- impl works on step N while test writes tests for step N-1
- feedback from test is relayed to impl and fixed in the next step

## log

### test - step 1 part A (tu.Run)

- files: internal/tu/run.go, internal/tu/run_test.go
- signature: `Run[I, E any](t *T, cases []Case[I, E], fn func(I) (E, error), eq func(want, got E) bool)`
- feedback: brief pinned the signature exactly, no ambiguity. `Run` skips the value check when `Case.Err` is set, so cases that need a partial value and an error (parser step 2) must use plain `t.Run`. Nothing compiled, checked by hand.

### impl - step 1 (token, lexer)

- files: lang/token/kind.go, lang/lexer/options.go, lang/lexer/lex.go, handoff/step-1.md
- api: `Kind.String()`, `lexer.New(src, ...Option)`, `lexer.WithComments()`
- decisions: punctuation tokens carry source text as Lit, unknown escapes kept literally, `\r` advances Col
- feedback: brief called lang.spec "a script" but it is prose, impl proposed a derived `spec.ak` in the handoff for all golden tests. Brief did not say whether punctuation tokens carry a Lit. Both should be in the brief next time.

### impl - step 2 (parser)

- files: lang/parser/parser.go (Error format), lang/parser/parse.go, handoff/step-2.md
- api: `parser.New(lx)`, Error `line:col: msg`, Errors joined by newline, nil error when clean
- decisions: syntax errors drop the call and sync to NEWLINE, rule errors keep the node in the AST; unclosed paren at EOF is a dedicated message; parser skips COMMENT tokens
- feedback: brief should carry an error message catalogue per step, the tester's work is dominated by matching messages. The paste-ready AST literal in the handoff was worth it, hand derived columns are the error prone part.
- main agent answers: keep quoting only ILLEGAL lits, keep "rule error keeps the node"

### test - step 1 part B (token, lexer)

- files: lang/token/kind_test.go, lang/lexer/testdata/spec.ak, lang/lexer/lex_test.go, handoff/step-1.feedback.md
- coverage: idents, ints, string escapes, unterminated strings, punctuation, comments dropped and kept, CRLF, illegal runes, EOF repeat, All termination, golden spec.ak (90 tokens, with and without comments)
- suspected bugs: none. Observations only: CRLF shifts NEWLINE Col by one, backslash before CRLF stays in the ILLEGAL Lit, unknown escapes are permissive
- feedback: the handoff with exact token stream and Lit conventions made the tests writable without guessing; keep the same precision on error text and positions in later steps


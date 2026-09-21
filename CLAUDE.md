# CLAUDE.md

## Project Overview

akha: is a project for using AI to build anything, with focus on automation and distributed build systems

tech stack: go, cobra, viper,
generate: sqlc, protobuf, graphql

## structure:

@/ - root folder

## Conventions

### major patterns
- interface first; every concrete type asserted via `var _ I = (*Impl)(nil)`
- factories sit next to the type they build; one allocation site per shape (`pipeline.New`, `fsout.NewWriter`, `handler.NewObject`)
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

### Code Generation - types are defined and code is generated
1. Define types in protobuf, sqlc, graphql files
 - structure: `@/_defs/` 
   - `db/` is for sqlc, `proto/` is for protobuf and `graphql/` is for GraphQL
 
**`Generated` code:**
- Protobuf output: `@/sdk/proto-sdk/`
- SQLC output: `@/sdk/db-sdk/`
- GraphQL server: `@/api/graphql/*`

## Examples
 - __local/starlark-go - a starlark interpreter written in Go, used for inspiration and examples

# do
- use simple formats form MD files
   - list, titles, sections and code blocks
   - no tables, no fancy styling
   - minimal emojis usage allow only if specified in the task

# don't
- ever use the '—' symbols use '-' for lists and '--' for flags
- assume - mark as "unknown" and use emoji's (warning sign)
- read files 2 times
- read files you have created in this session
- write long descriptions
- use bash for reading, searching, writing, or other file operations
   - use Read, Write, Search, Glob, Grep tools
- read examples unless specified in the task
- read contents of ./_env/_arch folder or ./_env/examples folder
- read __arch folder or examples folder

!!! don't are valid unless specified in the task
!!! very important don't read ./_env/* forbidden folders
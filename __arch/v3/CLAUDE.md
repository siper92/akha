# CLAUDE.md

## Project Overview

akha: a platform for AI workflows, similar to n8n and temporal, written in Go
- a worker runs `.ak` scripts with an own interpreter
- a backend issues JWT tokens to workers over gRPC
- a backend can start workers as goroutines with backend privileges

tech stack: go 1.27, cobra, viper, grpc, sqlite, log/slog
generate: sqlc, protobuf, graphql (deferred)

# skill
 - akha-language: use Claude to implement new language features for akha-language
   - prefer specifically invocation over assumed usage

# do
- use simple formats for MD files
   - list, titles, sections and code blocks
   - no tables, no fancy styling
   - minimal emojis usage allow only if specified in the task

# don't
- ever use the '—' symbols use '-' for lists and '--' for flags
- assume you know, unless asked to do so 
  - mark as "unknown" and use emoji's (warning sign)
- read files 2 times
- read files you have created in this session
- write long descriptions
- use bash for reading, searching, writing, or other file operations
   - use Read, Write, Search, Glob, Grep tools
- run bash to re-check signatures of files already read, act on what is known
- read examples unless specified in the task
- read contents of ./_env or ./__arch folder
- read __arch folder or examples folder unless the task points to it
- run commands - use the `just` command only when not specified in the task

!!! don't are valid unless specified in the task
!!! very important don't read ./_env/* forbidden folders

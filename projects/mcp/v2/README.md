# mcp v2

CLI example of Claude tool use with the Anthropic Go SDK only.
No MCP library, no server, no HTTP - tools are plain Go types behind a `ToolService`.

## files

- `mcp.go` - `Agent`, `Tool`, `ToolService` interfaces, errors, default model
- `service.go` - `Service`, a `ToolService` that routes calls by tool name
- `tool_file.go` - `createFile(name, content)` tool, writes inside one dir via `os.Root`
- `agent.go` - `ClaudeAgent`, the tool loop
- `cmd/main.go` - CLI entry

## flow

1. send prompt + tool params to `Beta.Messages.New`
2. on `tool_use` blocks call `ToolService.Call`, send back `tool_result` blocks
3. repeat until stop reason is not `tool_use`, return the final text

## run

```
ANTHROPIC_API_KEY=... go run ./projects/mcp/v2/cmd -dir ./out
```

## result

- `out/file1.txt` - part 1 of the story
- `out/file2.txt` - part 2 of the story
- `out/file3.txt` - part 3 of the story

## add a tool

- implement `Tool` (`Name`, `Param`, `Call`)
- pass it to `NewService`

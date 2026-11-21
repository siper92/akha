# mcp

MCP server with file tools, driven by Claude through the Anthropic Go SDK.

## flow

- `NewServer(root)` builds an MCP server with tools
  - `createFile(name, content)` - writes a file into `root`
  - `listFiles()` - lists files in `root`
- the agent connects to the server over an in memory transport
- MCP tools are listed and passed to Claude as tool definitions
- call API with prompt -> get tool_use blocks -> call MCP tools -> send results -> repeat until done

## setup

```
go get github.com/modelcontextprotocol/go-sdk
go mod tidy
export ANTHROPIC_API_KEY=...
```

## run

```
go run ./projects/mcp/cmd -root ./out
go run ./projects/mcp/cmd -prompt "write a 3 line poem into poem.txt"
go run ./projects/mcp/cmd -serve
```

- `-serve` runs only the MCP server over stdio, for external MCP clients

## result

- out/part1.txt: part 1 of the story
- out/part2.txt: part 2 of the story
- out/part3.txt: part 3 of the story

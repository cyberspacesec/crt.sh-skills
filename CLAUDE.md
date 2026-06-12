# CLAUDE.md — go-crt.sh Project Instructions

## Project Overview

This is a Go SDK and MCP server wrapping the [crt.sh](https://crt.sh/) Certificate Transparency search engine.

## Architecture

- `pkg/crtsh/` — Go SDK (Client, models, API calls)
- `cmd/mcp-server/` — MCP server entry point with stdio/SSE/HTTP transports
- `mcp-server` — Pre-compiled binary (Linux x86-64)
- `examples/` — Usage examples for the Go SDK

## MCP Server

The MCP server exposes two tools:
- `search_certificates` — Search CT logs by domain, hash, serial, CA, etc.
- `get_certificate` — Retrieve a specific certificate by crt.sh ID

### Starting the server

```bash
# stdio mode (default, for Claude Code integration)
./mcp-server --transport stdio

# Or via go run
go run ./cmd/mcp-server --transport stdio

# HTTP mode (for remote access)
./mcp-server --transport http --addr :8080

# SSE mode (for browser-based clients)
./mcp-server --transport sse --addr :8080 --base-url https://my-server.com
```

## Development Commands

```bash
# Run tests
go test ./pkg/crtsh/...

# Build MCP server binary
go build -o mcp-server ./cmd/mcp-server/

# Run MCP server locally
go run ./cmd/mcp-server --transport stdio
```

## Code Style

- Follow standard Go conventions
- Error messages use lowercase, no trailing punctuation
- Exported functions and types must have doc comments

# CLAUDE.md — go-crt.sh Project Instructions

## Project Overview

This is a Go SDK and MCP server wrapping the [crt.sh](https://crt.sh/) Certificate Transparency search engine. The goal is to be **大而全** — every crt.sh feature that can be wrapped, must be wrapped.

## Architecture

- `pkg/crtsh/` — Go SDK (Client, models, API calls, info pages)
- `cmd/mcp-server/` — MCP server entry point with stdio/SSE/HTTP transports
- `mcp-server` — Pre-compiled binary (Linux x86-64)
- `examples/` — Usage examples for the Go SDK

## MCP Server Tools

The MCP server exposes **3 tools**:

### search_certificates
Search CT logs by domain, hash, serial, CA, etc. Supports all crt.sh search types:
- `""` (default), `c` (certificate fingerprint), `id`, `ctid`, `serial`, `ski`
- `spkisha1`, `spkisha256`, `subjectsha1`, `sha1`, `sha256`
- `ca`, `CAID`, `CAName`, `Identity`, `CN`, `E`, `OU`, `O`
- `dNSName`, `rfc822Name`, `iPAddress`

Options: match mode, exclude expired, deduplicate, show SQL, linter, lint type, pagination.

### get_certificate
Retrieve a specific certificate by crt.sh ID.

### get_info_page
Retrieve crt.sh information pages:
- `cert-populations`, `revoked-intermediates`, `ca-issuers`, `ocsp-responders`
- `test-websites`, `monitored-logs`, `accepted-roots-missing`, `gen-add-chain`
- `mozilla-disclosures`, `mozilla-certvalidations`, `mozilla-onecrl`
- `apple-disclosures`, `chrome-disclosures`

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

## crt.sh URL Parameter Format

**IMPORTANT:** The crt.sh JS uses specific URL parameter formats that differ from HTML form field names:
- `exclude=expired` (NOT `excludeExpired=on`)
- `deduplicate=Y` (NOT `deduplicate=on`)
- `showSQL=Y` (NOT `showSQL=on`)
- Linter params use linter name as key: `zlint=issues` (NOT `linter=zlint&linttype=issues`)
- Search type `c`: URL is `?c=<fingerprint>` (NOT `?q=<fingerprint>`)
- Search type `ca`: URL is `?ca=<value>` (NOT `?q=<value>`)

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
- When adding new crt.sh features, ensure the SDK layer (pkg/crtsh/) supports it first, then expose via MCP tool

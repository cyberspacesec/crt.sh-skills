# go-crt.sh

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go SDK and MCP server wrapping the [crt.sh](https://crt.sh/) Certificate Transparency search engine. Use it as a Go library, CLI tool, or MCP server for AI applications.

## Features

- 🔍 **Certificate Search** — Search CT logs by domain, SHA-256 hash, serial number, CA name, and more
- 📜 **Certificate Details** — Retrieve specific certificates by crt.sh ID
- 🤖 **MCP Server** — Expose crt.sh as tools for AI applications (Claude Code, Cursor, etc.)
- 📦 **Go SDK** — Full-featured Go client with retries, pagination, and flexible query building
- 🛡️ **Claude Code Skill** — Pre-built skills for certificate search and subdomain enumeration

## Quick Start

### As a Go SDK

```go
package main

import (
    "context"
    "fmt"
    "log"

    crtsh "github.com/cyberspacesec/go-crt.sh/pkg/crtsh"
)

func main() {
    client := crtsh.NewClient()

    // Search certificates for a domain
    certs, pagination, err := client.SearchCertificates(context.Background(), crtsh.QueryParams{
        Q:              "example.com",
        Deduplicate:    true,
        ExcludeExpired: true,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, cert := range certs {
        fmt.Printf("ID: %d, Domains: %v\n", cert.ID, cert.Domains)
    }

    if pagination != nil && pagination.NextPage > 0 {
        fmt.Printf("More results available (next page: %d)\n", pagination.NextPage)
    }
}
```

### As an MCP Server

Build and run the MCP server:

```bash
# Build
go build -o mcp-server ./cmd/mcp-server/

# Run in stdio mode (default, for Claude Code)
./mcp-server --transport stdio

# Run in HTTP mode (for remote access)
./mcp-server --transport http --addr :8080

# Run in SSE mode (for browser-based clients)
./mcp-server --transport sse --addr :8080 --base-url https://my-server.com
```

### Integration with Claude Code

This project includes a `.mcp.json` file that automatically registers the MCP server with Claude Code when opened as a project. The included skills (`crtsh-search` and `crtsh-cert`) provide guidance on when and how to use the tools.

**Option 1: Project-level (automatic)**

Clone this repo and open it in Claude Code — the MCP server is auto-registered via `.mcp.json`.

**Option 2: Global registration**

Add to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "go-crt-sh": {
      "command": "/path/to/mcp-server",
      "args": ["--transport", "stdio"]
    }
  }
}
```

**Option 3: HTTP mode (remote)**

Start the server in HTTP mode, then register:

```json
{
  "mcpServers": {
    "go-crt-sh": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## MCP Tools

### search_certificates

Search certificate transparency logs via crt.sh.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search term (domain, hash, serial, etc.) |
| `search_type` | string | No | Type: `""`, `sha256`, `serial`, `CN`, `dNSName`, `iPAddress`, etc. |
| `match` | string | No | Match mode: `""`, `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `exclude_expired` | boolean | No | Exclude expired certificates |
| `deduplicate` | boolean | No | Deduplicate precertificate pairs |
| `page` | number | No | Page number (1-based) |
| `page_size` | number | No | Results per page |

### get_certificate

Retrieve a specific certificate by its crt.sh ID.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | number | Yes | The crt.sh certificate ID |

## Claude Code Skills

This project includes two skills:

| Skill | Trigger | Purpose |
|-------|---------|---------|
| `crtsh-search` | CT log search, subdomain enum, domain reconnaissance | Search certificates by domain, hash, CA, etc. |
| `crtsh-cert` | Certificate ID lookup, cert detail query | Get specific certificate by ID |

## Examples

See the `examples/` directory for Go SDK usage:

- `001-basic-search` — Basic domain search
- `002-advance-search` — Search by SHA-256 fingerprint
- `003-ca-search` — Search by Certificate Authority
- `004-get-cert-by-id` — Get a specific certificate
- `005-page` — Pagination example

## Development

```bash
# Run tests
go test ./pkg/crtsh/...

# Build MCP server
go build -o mcp-server ./cmd/mcp-server/

# Run with debug output
go run ./cmd/mcp-server --transport stdio
```

## License

MIT License — see [LICENSE](LICENSE) for details.

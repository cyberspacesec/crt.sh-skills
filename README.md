# go-crt.sh

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go SDK and MCP server wrapping the [crt.sh](https://crt.sh/) Certificate Transparency search engine. Use it as a Go library, CLI tool, or MCP server for AI applications. **Every crt.sh feature is wrapped — nothing is left out.**

## Features

- 🔍 **Certificate Search** — Search CT logs by domain, SHA-256 hash, serial number, CA name, and more (21 search types)
- 📜 **Certificate Details** — Retrieve specific certificates by crt.sh ID
- 🔧 **Certificate Linting** — Run cablint, x509lint, zlint, keylint, or all linters
- 📊 **Info Pages** — Access crt.sh information pages (CA disclosures, CT log status, OCSP responders, etc.)
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
        fmt.Printf("ID: %d, CN: %s, Issuer: %s, Domains: %v\n",
            cert.ID, cert.CommonName, cert.IssuerName, cert.Domains)
    }

    if pagination != nil && pagination.NextPage > 0 {
        fmt.Printf("More results available (next page: %d)\n", pagination.NextPage)
    }

    // Fetch an info page
    info, err := client.FetchInfoPage(context.Background(), "monitored-logs")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Page: %s — %s\n", info.Title, info.Description)
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

Search certificate transparency logs via crt.sh. Supports all 21 search types.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search term (domain, hash, serial, etc.) |
| `search_type` | string | No | Type: `""`, `c`, `id`, `ctid`, `serial`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `sha1`, `sha256`, `ca`, `CAID`, `CAName`, `Identity`, `CN`, `E`, `OU`, `O`, `dNSName`, `rfc822Name`, `iPAddress` |
| `match` | string | No | Match mode: `""`, `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `exclude_expired` | boolean | No | Exclude expired certificates |
| `deduplicate` | boolean | No | Deduplicate precertificate pairs |
| `linter` | string | No | Linter: `cablint`, `x509lint`, `zlint`, `keylint`, `lint` (all) |
| `lint_type` | string | No | Lint output: `1 week` (summary), `issues` |
| `page` | number | No | Page number (1-based) |
| `page_size` | number | No | Results per page |

### get_certificate

Retrieve a specific certificate by its crt.sh ID.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | number | Yes | The crt.sh certificate ID |

### get_info_page

Retrieve crt.sh information pages (CA disclosures, CT log data, revocation lists, etc.).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | string | Yes | Page name (see below) |

**Available pages:**

| Page | Description |
|------|-------------|
| `cert-populations` | Certificate population statistics across CT logs |
| `revoked-intermediates` | List of revoked intermediate CA certificates |
| `ca-issuers` | Certificate Authority issuer information |
| `ocsp-responders` | OCSP responder information for CAs |
| `test-websites` | Test websites for certificate validation |
| `monitored-logs` | CT logs monitored by crt.sh |
| `accepted-roots-missing` | Root certificates accepted but missing from database |
| `gen-add-chain` | Certificate submission assistant |
| `mozilla-disclosures` | Mozilla CA certificate disclosures |
| `mozilla-certvalidations` | Mozilla certificate validation requirements |
| `mozilla-onecrl` | Mozilla certificate revocation list (OneCRL) |
| `apple-disclosures` | Apple CA certificate disclosures |
| `chrome-disclosures` | Chrome CA certificate disclosures |

## Claude Code Skills

This project includes two skills:

| Skill | Trigger | Purpose |
|-------|---------|---------|
| `crtsh-search` | CT log search, subdomain enum, domain reconnaissance | Search certificates by domain, hash, CA, etc. |
| `crtsh-cert` | Certificate ID lookup, cert detail query | Get specific certificate by ID |

## Certificate Model

The `Certificate` struct includes all fields returned by crt.sh:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | int | crt.sh certificate ID |
| `IssuerCAID` | int | Certificate Authority ID |
| `IssuerName` | string | Full issuer distinguished name |
| `CommonName` | string | Certificate commonName |
| `NameValue` | []string | All domain names (parsed) |
| `RawNameValue` | string | Raw name_value field |
| `Domains` | []string | Deduplicated, wildcard-stripped domains |
| `EntryTimestamp` | time.Time | CT log entry timestamp |
| `NotBefore` | time.Time | Certificate validity start |
| `NotAfter` | time.Time | Certificate validity end |
| `SerialNumber` | string | Certificate serial number |
| `ResultCount` | int | Number of matching results |

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

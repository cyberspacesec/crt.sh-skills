# go-crt.sh

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go SDK, CLI tool, and MCP server wrapping the [crt.sh](https://crt.sh/) Certificate Transparency search engine. **Every crt.sh feature is wrapped — nothing is left out.**

## Three-Layer Architecture

```
┌─────────────────────────────────────────────┐
│  Skills (.claude/skills/)                    │  AI-readable docs
├─────────────────────────────────────────────┤
│  MCP Server (5 tools) + CLI Tool (8 cmds)   │  AI-callable + Human-callable
├─────────────────────────────────────────────┤
│  Go SDK (5 methods)                          │  Programmatic API
└─────────────────────────────────────────────┘
```

All three layers expose the **exact same capabilities** — no feature exists in one layer but not another.

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

    // Search certificates
    certs, _, err := client.SearchCertificates(context.Background(), crtsh.QueryParams{
        Q:              "example.com",
        Deduplicate:    true,
        ExcludeExpired: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, cert := range certs {
        fmt.Printf("ID: %d, CN: %s, Domains: %v\n", cert.ID, cert.CommonName, cert.Domains)
    }

    // Build a Censys URL
    url, _ := crtsh.BuildCensysURL("CN", "example.com")
    fmt.Println("Censys:", url)
}
```

### As a CLI Tool

```bash
# Build
go build -o crtsh-cli ./cmd/crtsh-cli/

# Search certificates
crtsh-cli search example.com --exclude-expired --deduplicate
crtsh-cli search ABCDEF1234 --type sha256
crtsh-cli search "Let's Encrypt" --type CAName

# Get certificate by ID
crtsh-cli get-cert 26786991824 --json

# Get info page
crtsh-cli info-page monitored-logs

# Get CA details
crtsh-cli get-ca 16418

# Build Censys URL
crtsh-cli censys "example.com" --type CN

# List available options
crtsh-cli list-types
crtsh-cli list-pages
```

### As an MCP Server

```bash
# Build
go build -o mcp-server ./cmd/mcp-server/

# Run in stdio mode (for Claude Code)
./mcp-server --transport stdio

# Run in HTTP mode (for remote access)
./mcp-server --transport http --addr :8080

# Run in SSE mode (for browser-based clients)
./mcp-server --transport sse --addr :8080 --base-url https://my-server.com
```

**Integration with Claude Code:**

This project includes a `.mcp.json` file for automatic registration. For global registration:

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

## MCP Tools (5 tools)

### search_certificates

Search certificate transparency logs via crt.sh. Supports all 21 search types.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search term |
| `search_type` | string | No | Type: `""`, `c`, `id`, `ctid`, `serial`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `sha1`, `sha256`, `ca`, `CAID`, `CAName`, `Identity`, `CN`, `E`, `OU`, `O`, `dNSName`, `rfc822Name`, `iPAddress` |
| `match` | string | No | Match mode: `""`, `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `exclude_expired` | boolean | No | Exclude expired certificates |
| `deduplicate` | boolean | No | Deduplicate precertificate pairs |
| `show_sql` | boolean | No | Show SQL query (debugging) |
| `linter` | string | No | Linter: `cablint`, `x509lint`, `zlint`, `keylint`, `lint` |
| `lint_type` | string | No | Lint output: `1 week`, `issues` |
| `page` | number | No | Page number (1-based) |
| `page_size` | number | No | Results per page |

### get_certificate

Retrieve a specific certificate by its crt.sh ID.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | number | Yes | The crt.sh certificate ID |

### get_info_page

Retrieve crt.sh information pages.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | string | Yes | Page name (see list below) |

**Available pages:** cert-populations, revoked-intermediates, ca-issuers, ocsp-responders, test-websites, monitored-logs, accepted-roots-missing, gen-add-chain, mozilla-disclosures, mozilla-certvalidations, mozilla-onecrl, apple-disclosures, chrome-disclosures

### get_ca

Retrieve CA certificate details by CA ID.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ca_id` | number | Yes | The crt.sh CA ID (from issuer_ca_id) |

### search_censys

Build a Censys.io certificate search URL. Not all search types are supported by Censys.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search term |
| `search_type` | string | Yes | Type: `c`, `serial`, `sha1`, `sha256`, `ca`, `CAName`, `Identity`, `CN`, `OU`, `O`, `dNSName`, `rfc822Name`, `iPAddress` |

## Claude Code Skills

| Skill | Trigger | Purpose |
|-------|---------|---------|
| `crtsh-search` | CT log search, subdomain enum, domain reconnaissance | Full search + all 5 tools |
| `crtsh-cert` | Certificate ID lookup, CA investigation | Certificate + CA details |

## Go SDK API

```go
// Search certificates
certs, pagination, err := client.SearchCertificates(ctx, crtsh.QueryParams{
    Q:              "example.com",
    SearchType:     "dNSName",
    ExcludeExpired: true,
    Deduplicate:    true,
    Page:           1,
    PageSize:       50,
})

// Get certificate by ID
cert, err := client.GetCertificateByID(ctx, 26786991824)

// Get info page
info, err := client.FetchInfoPage(ctx, "monitored-logs")

// Get CA details
ca, err := client.FetchCAByID(ctx, 16418)

// Build Censys URL
url, err := crtsh.BuildCensysURL("CN", "example.com")
```

## Certificate Model

| Field | Type | Description |
|-------|------|-------------|
| `ID` | int | crt.sh certificate ID |
| `IssuerCAID` | int | Certificate Authority ID |
| `IssuerName` | string | Full issuer distinguished name |
| `CommonName` | string | Certificate commonName |
| `NameValue` | []string | All domain names (parsed) |
| `Domains` | []string | Deduplicated, wildcard-stripped domains |
| `EntryTimestamp` | time.Time | CT log entry timestamp |
| `NotBefore` | time.Time | Certificate validity start |
| `NotAfter` | time.Time | Certificate validity end |
| `SerialNumber` | string | Certificate serial number |
| `ResultCount` | int | Number of matching results |

## Development

```bash
# Run tests
go test ./pkg/crtsh/...

# Build all binaries
go build -o mcp-server ./cmd/mcp-server/
go build -o crtsh-cli ./cmd/crtsh-cli/

# Run MCP server
go run ./cmd/mcp-server --transport stdio
```

## License

MIT License — see [LICENSE](LICENSE) for details.

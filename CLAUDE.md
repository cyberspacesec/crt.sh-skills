# CLAUDE.md — crt.sh-skills Project Instructions

## Project Overview

This is a Go SDK and MCP server wrapping the [crt.sh](https://crt.sh/) Certificate Transparency search engine. The goal is **大而全** — every crt.sh feature that can be wrapped, must be wrapped.

**Repository:** `github.com/cyberspacesec/crt.sh-skills`
**Go Module:** `github.com/cyberspacesec/crt.sh-skills`

## Three-Layer Architecture

```
┌─────────────────────────────────────────────┐
│  Skills (.claude/skills/)                    │  AI-readable docs describing tools + CLI
├─────────────────────────────────────────────┤
│  MCP Server (cmd/mcp-server/)               │  AI-callable tools (5 tools)
│  CLI Tool (cmd/crtsh-cli/)                  │  Human-callable commands (8 commands)
├─────────────────────────────────────────────┤
│  Go SDK (pkg/crtsh/)                        │  Programmatic API (5 methods)
└─────────────────────────────────────────────┘
```

All three layers expose the **exact same capabilities** — no feature exists in one layer but not another.

## SDK Methods (pkg/crtsh/)

| Method | Description |
|--------|-------------|
| `SearchCertificates(ctx, QueryParams)` | Search CT logs (21 search types, 7 match modes, linting, pagination) |
| `GetCertificateByID(ctx, id)` | Get certificate by crt.sh ID |
| `FetchInfoPage(ctx, pagePath)` | Get info page (13 pages) |
| `FetchCAByID(ctx, caID)` | Get CA certificate details |
| `BuildCensysURL(searchType, value)` | Build Censys.io search URL |

## MCP Tools (cmd/mcp-server/)

| Tool | Maps to SDK Method |
|------|-------------------|
| `search_certificates` | SearchCertificates |
| `get_certificate` | GetCertificateByID |
| `get_info_page` | FetchInfoPage |
| `get_ca` | FetchCAByID |
| `search_censys` | BuildCensysURL |

## CLI Commands (cmd/crtsh-cli/)

| Command | Maps to SDK Method |
|---------|-------------------|
| `crtsh-cli search [query]` | SearchCertificates |
| `crtsh-cli get-cert [id]` | GetCertificateByID |
| `crtsh-cli info-page [page]` | FetchInfoPage |
| `crtsh-cli get-ca [ca-id]` | FetchCAByID |
| `crtsh-cli censys [query]` | BuildCensysURL |
| `crtsh-cli list-types` | List search types |
| `crtsh-cli list-pages` | List info pages |

## Release Process

Tag-based release via GitHub Actions:

```bash
# Create a release tag
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions builds binaries for:
# linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64
# Both mcp-server and crtsh-cli are built and uploaded as release assets.
```

Binary naming: `crtsh-skills-mcp-server-{os}-{arch}.tar.gz` and `crtsh-skills-cli-{os}-{arch}.tar.gz`

## crt.sh URL Parameter Format

**IMPORTANT:** The crt.sh JS uses specific URL parameter formats:
- Search types use their name as the URL param key: `?CN=value`, `?sha256=value`, etc.
- `exclude=expired` (NOT `excludeExpired=on`)
- `deduplicate=Y` (NOT `deduplicate=on`)
- `showSQL=Y` (NOT `showSQL=on`)
- Linter params use linter name as key: `zlint=issues`
- No `searchtype` URL param — the search type IS the param name

## Development Commands

```bash
# Run tests
go test ./pkg/crtsh/...

# Build MCP server binary
go build -o mcp-server ./cmd/mcp-server/

# Build CLI binary
go build -o crtsh-cli ./cmd/crtsh-cli/

# Build with version
go build -ldflags "-X main.Version=v1.0.0" -o mcp-server ./cmd/mcp-server/
go build -ldflags "-X main.Version=v1.0.0" -o crtsh-cli ./cmd/crtsh-cli/

# Run MCP server
go run ./cmd/mcp-server --transport stdio

# Run CLI
go run ./cmd/crtsh-cli/ search example.com --exclude-expired --deduplicate
```

## Code Style

- Follow standard Go conventions
- Error messages use lowercase, no trailing punctuation
- Exported functions and types must have doc comments
- When adding new crt.sh features, ensure ALL THREE layers are updated: SDK → MCP tool + CLI command → Skills

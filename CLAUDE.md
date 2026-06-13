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
│  CLI Tool (cmd/crtsh-cli/)                  │  Human-callable commands (10 commands)
├─────────────────────────────────────────────┤
│  Go SDK (pkg/crtsh/)                        │  Programmatic API (6 methods + helpers)
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
| `IterateCertificates(ctx, QueryParams, fn)` | Auto-pagination helper |

### SDK Helper Functions

| Function | Description |
|----------|-------------|
| `SearchTypes()` | Returns all 22 valid search types with descriptions |
| `MatchModes()` | Returns all 7 match modes |
| `Linters()` | Returns all 5 certificate linters |
| `LintTypes()` | Returns all 2 lint output types |
| `ValidSearchTypes()` | Returns a map for quick lookup |
| `IsNotFoundError(err)` | Check if error is a not-found error |
| `IsRateLimitError(err)` | Check if error is a rate-limit error |
| `IsServerError(err)` | Check if error is a 5xx server error |

### Client Options

`NewClient(opts ...ClientOption)` accepts: `WithTimeout`, `WithRetryCount`, `WithDebug`, `WithUserAgent`, `WithBaseURL`

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
| `crtsh-cli list-types` | SearchTypes() |
| `crtsh-cli list-pages` | InfoPages map |
| `crtsh-cli list-linters` | Linters() |
| `crtsh-cli list-match-modes` | MatchModes() |

Root flags: `--timeout`, `--debug`, `--output/-o` (json|table|csv)

## Release Process

Tag-based release via GitHub Actions + GoReleaser:

```bash
# Create a release tag
git tag v1.1.0
git push origin v1.1.0

# GoReleaser builds binaries for:
# linux/amd64, linux/arm64, linux/386
# darwin/amd64, darwin/arm64
# windows/amd64, windows/arm64, windows/386
# freebsd/amd64
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
go test -v -race ./pkg/crtsh/...

# Build MCP server binary
go build -o mcp-server ./cmd/mcp-server/

# Build CLI binary
go build -o crtsh-cli ./cmd/crtsh-cli/

# Build with version
go build -ldflags "-X main.Version=v1.1.0" -o mcp-server ./cmd/mcp-server/
go build -ldflags "-X main.Version=v1.1.0" -o crtsh-cli ./cmd/crtsh-cli/

# Run MCP server
go run ./cmd/mcp-server --transport stdio

# Run CLI
go run ./cmd/crtsh-cli/ search example.com --exclude-expired --deduplicate

# Dry-run GoReleaser (no publish)
goreleaser release --snapshot --clean
```

## Code Style

- Follow standard Go conventions
- Error messages use lowercase, no trailing punctuation
- Exported functions and types must have doc comments
- When adding new crt.sh features, ensure ALL THREE layers are updated: SDK → MCP tool + CLI command → Skills

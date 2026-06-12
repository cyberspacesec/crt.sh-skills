# crt.sh Skill/Plugin 封装 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: `superpowers:subagent-driven-development`
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** 将 go-crt.sh 项目封装为 Claude Code Skill + Plugin，使任何 AI 应用通过 Claude Code 的 Skill 系统和 MCP 协议直接调用 crt.sh 证书透明度搜索能力。

**Architecture:** 用户/AI 应用 → Claude Code Skill（`crtsh-search`/`crtsh-cert`）→ MCP Server（stdio 传输）→ crt.sh API。Skill 提供使用指导和触发条件，MCP Server 提供实际工具调用能力。同时提供 `.mcp.json` 让用户可直接注册 MCP Server，无需手动配置。整个项目作为 Claude Code Plugin 发布，包含 `.claude-plugin/plugin.json` 清单。

**Tech Stack:** Go 1.23.4, mark3labs/mcp-go v0.34.0, Claude Code Skill/Plugin 体系

**Risks:**
- `.mcp.json` 中 command 路径需兼容不同安装位置 → 缓解：默认用 `go run` 命令，文档说明编译后二进制的配置方式
- `.gitignore` 中 `.mcp/` 规则不影响根目录的 `.mcp.json` → 已确认无冲突
- Skill description 必须精准触发，避免误匹配 → 缓解：使用明确的证书/CT/域名子域名枚举等关键词

---

### Task 1: 创建 CLAUDE.md 项目指令文件

**Depends on:** None
**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: 创建 CLAUDE.md — 为 Claude Code 提供项目级上下文和指令**

```markdown
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
```

- [ ] **Step 2: 验证 CLAUDE.md**
Run: `cat /home/cc11001100/github/cyberspacesec/go-crt.sh/CLAUDE.md | head -5`
Expected:
  - Exit code: 0
  - Output contains: "go-crt.sh Project Instructions"

- [ ] **Step 3: 提交**
Run: `git add CLAUDE.md && git commit -m "docs: add CLAUDE.md project instructions for Claude Code"`

---

### Task 2: 创建 MCP Server 注册配置

**Depends on:** None
**Files:**
- Create: `.mcp.json`

- [ ] **Step 1: 创建 .mcp.json — 注册 MCP Server 到 Claude Code**

```json
{
  "go-crt-sh": {
    "command": "go",
    "args": ["run", "./cmd/mcp-server", "--transport", "stdio"]
  }
}
```

说明：使用 `go run` 作为默认 command，确保在项目目录下即可使用。用户编译二进制后可替换为：

```json
{
  "go-crt-sh": {
    "command": "./mcp-server",
    "args": ["--transport", "stdio"]
  }
}
```

- [ ] **Step 2: 验证 .mcp.json 格式**
Run: `python3 -c "import json; d=json.load(open('/home/cc11001100/github/cyberspacesec/go-crt.sh/.mcp.json')); print('OK' if 'go-crt-sh' in d else 'FAIL')"`
Expected:
  - Exit code: 0
  - Output contains: "OK"

- [ ] **Step 3: 提交**
Run: `git add .mcp.json && git commit -m "feat: add .mcp.json to register MCP server with Claude Code"`

---

### Task 3: 创建 Skill 定义文件

**Depends on:** Task 1
**Files:**
- Create: `.claude/skills/crtsh-search/SKILL.md`
- Create: `.claude/skills/crtsh-cert/SKILL.md`

- [ ] **Step 1: 创建 crtsh-search Skill — 证书透明度搜索技能**

```markdown
---
name: crtsh-search
description: Use when searching certificate transparency logs for domains, subdomains, SSL/TLS certificates, or certificate transparency data via crt.sh. Triggers on mentions of CT logs, subdomain enumeration, certificate search, SSL fingerprint lookup, or domain reconnaissance.
allowed-tools: ["mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_certificate"]
---

# crt.sh Certificate Search

> Search certificate transparency logs to discover domains, subdomains, and certificate details.

## When to Use

- User asks about domains or subdomains under a target domain
- User wants to find SSL/TLS certificates for a domain
- User needs certificate transparency log data for reconnaissance
- User asks to enumerate subdomains via CT logs
- User wants to find certificates by hash, serial number, or CA

## When NOT to Use

- User asks about DNS records directly (use DNS tools instead)
- User asks about WHOIS information (use WHOIS tools instead)
- User asks about website content or HTTP responses (use web tools instead)

## Instructions

### Step 1: Determine the search type

Based on what the user is looking for, choose the appropriate `search_type`:

| User wants | search_type | Example query |
|-----------|------------|---------------|
| All certs for a domain | `""` (empty/default) | `example.com` |
| Subdomains via DNS SAN | `dNSName` | `example.com` |
| Cert by SHA-256 fingerprint | `sha256` | `ABCD1234...` |
| Cert by serial number | `serial` | `00:11:22:33` |
| Certs by Common Name | `CN` | `example.com` |
| Certs by CA name | `CAName` | `Let's Encrypt` |
| Certs by email | `E` | `admin@example.com` |
| Certs by organization | `O` | `Example Inc` |
| Certs by IP address SAN | `iPAddress` | `1.2.3.4` |

### Step 2: Call search_certificates

Call the `search_certificates` MCP tool with the determined parameters.

**For subdomain enumeration** (most common use case):
- Set `query` to the target domain (e.g., `example.com`)
- Set `search_type` to `""` (default) or `dNSName`
- Set `deduplicate` to `true` to remove duplicate precertificate pairs
- Set `exclude_expired` to `true` if only active certificates matter

**For pagination:**
- Start with `page=1` and `page_size=50`
- If `pagination.next_page` is set, there are more results
- Continue fetching until all pages are retrieved or user is satisfied

### Step 3: Parse and present results

The response contains a `certificates` array. Each certificate has:
- `id` — crt.sh certificate ID (use for get_certificate)
- `name_value` — domain names associated with the cert
- `domains` — deduplicated, wildcard-stripped domain list
- `entry_timestamp` — when the cert was logged
- `not_before` / `not_after` — certificate validity period
- `issuer_ca_id` — the CA that issued the cert
- `serial_number` — certificate serial number

**Present results as:**
1. A summary: "Found N certificates for domain X"
2. A deduplicated domain/subdomain list (extract from all `domains` fields)
3. Key certificate details if user asked specifically

### Step 4: Deep-dive with get_certificate (if needed)

If the user wants details on a specific certificate, call `get_certificate` with the `id` from the search results.

## Examples

### Example 1: Subdomain enumeration

User: "Find all subdomains of example.com via CT logs"

1. Call `search_certificates(query="example.com", deduplicate=true, exclude_expired=true)`
2. Extract all unique domains from the results
3. Present: "Found 15 unique subdomains for example.com: www.example.com, api.example.com, ..."

### Example 2: Certificate fingerprint lookup

User: "Look up this certificate by SHA-256: ABCD1234..."

1. Call `search_certificates(query="ABCD1234...", search_type="sha256")`
2. Present the matching certificate details

### Example 3: CA investigation

User: "What certificates has Let's Encrypt issued for example.com?"

1. Call `search_certificates(query="example.com", deduplicate=true)`
2. Filter results where `issuer_ca_id` matches Let's Encrypt
3. Present filtered results

## Notes

- crt.sh can be slow or return 5xx errors during peak load — the SDK retries automatically
- For large domains, results can span many pages — ask user if they want all pages
- Wildcard certificates (`*.example.com`) are stripped to the base domain in the `domains` field
- The `name_value` field may contain multiple domains separated by newlines
```

- [ ] **Step 2: 创建 crtsh-cert Skill — 单证书详情查询技能**

```markdown
---
name: crtsh-cert
description: Use when retrieving a specific certificate's full details from crt.sh by its ID. Triggers on mentions of certificate ID lookup, specific cert details, or when user provides a crt.sh numeric ID.
allowed-tools: ["mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__search_certificates"]
---

# crt.sh Certificate Detail Lookup

> Retrieve detailed information about a specific certificate by its crt.sh ID.

## When to Use

- User provides a numeric crt.sh certificate ID and wants full details
- User found a certificate in search results and wants to deep-dive
- User asks about a specific certificate's issuer, validity, or other details

## When NOT to Use

- User wants to search for certificates by domain → use `crtsh-search` skill instead
- User asks about general CT log information → explain conceptually, then offer search

## Instructions

### Step 1: Obtain the certificate ID

The crt.sh certificate ID is a numeric value. Sources:
- Direct from user input (e.g., "look up cert 12345")
- From previous `search_certificates` results (the `id` field)

If the user provides a domain name or hash instead of an ID, first use `search_certificates` to find the cert, then use `get_certificate` for details.

### Step 2: Call get_certificate

Call the `get_certificate` MCP tool with:
- `id` — the numeric crt.sh certificate ID

### Step 3: Present the certificate details

The response contains the full certificate data including:
- `id` — crt.sh ID
- `issuer_ca_id` — Certificate Authority ID
- `name_value` — All names/domains on the certificate
- `entry_timestamp` — CT log entry time
- `not_before` — Certificate validity start
- `not_after` — Certificate validity end
- `serial_number` — Certificate serial number

Present the information in a structured, readable format.

## Examples

### Example 1: Direct ID lookup

User: "Show me certificate 9999999"

1. Call `get_certificate(id=9999999)`
2. Present: "Certificate #9999999: Issued by CA #12345, valid from 2024-01-01 to 2025-01-01, for domains: example.com, www.example.com"

### Example 2: Domain to certificate details

User: "Show me the details of the most recent certificate for example.com"

1. Call `search_certificates(query="example.com", exclude_expired=true, page_size=1)`
2. Get the `id` from the first result
3. Call `get_certificate(id=<that_id>)`
4. Present full details
```

- [ ] **Step 3: 验证 Skill 文件**
Run: `ls -la /home/cc11001100/github/cyberspacesec/go-crt.sh/.claude/skills/crtsh-search/SKILL.md /home/cc11001100/github/cyberspacesec/go-crt.sh/.claude/skills/crtsh-cert/SKILL.md`
Expected:
  - Exit code: 0
  - Output contains: "SKILL.md" (appears twice)

- [ ] **Step 4: 提交**
Run: `git add .claude/ && git commit -m "feat: add crtsh-search and crtsh-cert skills for Claude Code"`

---

### Task 4: 创建 Plugin 清单和完善文档

**Depends on:** Task 2, Task 3
**Files:**
- Create: `.claude-plugin/plugin.json`
- Modify: `README.md` (全文重写)

- [ ] **Step 1: 创建 Plugin 清单 — 定义 go-crt.sh 作为 Claude Code Plugin 的元数据**

```json
{
  "name": "go-crt-sh",
  "description": "Certificate transparency search via crt.sh — discover domains, subdomains, and SSL/TLS certificate details from CT logs",
  "version": "1.0.0",
  "author": {
    "name": "CyberspaceSec",
    "url": "https://github.com/cyberspacesec"
  },
  "homepage": "https://github.com/cyberspacesec/go-crt.sh",
  "repository": "https://github.com/cyberspacesec/go-crt.sh.git",
  "license": "MIT",
  "keywords": ["certificate-transparency", "crt.sh", "subdomain-enum", "ssl", "tls", "reconnaissance", "security", "mcp"]
}
```

- [ ] **Step 2: 重写 README.md — 完整的项目文档，包含安装和使用说明**

```markdown
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
```

- [ ] **Step 3: 验证文件创建**
Run: `cat /home/cc11001100/github/cyberspacesec/go-crt.sh/.claude-plugin/plugin.json | python3 -c "import sys,json; d=json.load(sys.stdin); print('OK:', d['name'])" && head -3 /home/cc11001100/github/cyberspacesec/go-crt.sh/README.md`
Expected:
  - Exit code: 0
  - Output contains: "OK: go-crt-sh"
  - Output contains: "go-crt.sh"

- [ ] **Step 4: 提交**
Run: `git add .claude-plugin/ README.md && git commit -m "feat: add plugin manifest and comprehensive README documentation"`

---

### Task 5: 更新 .gitignore 以确保关键文件不被忽略

**Depends on:** None
**Files:**
- Modify: `.gitignore:57` (`.mcp/` 行)

- [ ] **Step 1: 检查 .gitignore 现有规则 — 确认 .mcp.json 不被忽略**

当前 `.gitignore:57` 包含 `.mcp/` 规则。这个规则只忽略 `.mcp/` 目录，不会影响根目录的 `.mcp.json` 文件。但为了安全起见，添加一条显式排除规则确保 `.mcp.json` 不被忽略。

文件: `.gitignore:57`

```text
# MCP related
.mcp/
!.mcp.json
```

- [ ] **Step 2: 验证 .gitignore 不忽略 .mcp.json**
Run: `cd /home/cc11001100/github/cyberspacesec/go-crt.sh && git check-ignore .mcp.json; echo "exit code: $?"`
Expected:
  - Exit code: 1 (file is NOT ignored)
  - Output does not contain: ".mcp.json"

- [ ] **Step 3: 提交**
Run: `git add .gitignore && git commit -m "fix: ensure .mcp.json is not ignored by .gitignore"`

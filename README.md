# crt.sh-skills

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Release](https://img.shields.io/github/v/release/cyberspacesec/crt.sh-skills?include_prereleases)](https://github.com/cyberspacesec/crt.sh-skills/releases/latest) [![CI](https://github.com/cyberspacesec/crt.sh-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/cyberspacesec/crt.sh-skills/actions/workflows/ci.yml)

A comprehensive wrapper for the [crt.sh](https://crt.sh/) Certificate Transparency search engine — accessible via **Skills**, MCP, CLI, or Go SDK. **Every crt.sh feature is wrapped — nothing is left out.**

[简体中文](#简体中文) | **English**

---

## 🚀 4 Ways to Connect

| # | Method | Best for | One-liner |
|---|--------|----------|-----------|
| 1 | **Skills** | AI agents (Claude Code, Cursor, etc.) | Zero install — just add config |
| 2 | **MCP Server** | Any MCP-compatible AI tool | `./mcp-server --transport stdio` |
| 3 | **CLI** | Humans, scripts, pipelines | `crtsh-cli search example.com` |
| 4 | **Go SDK** | Go programs, integrations | `import crtsh "github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"` |

All four layers expose the **exact same capabilities** — no feature exists in one layer but not another.

```
┌─────────────────────────────────────────────┐
│  Skills (.claude/skills/)                    │  AI-readable docs (trigger-based)
├─────────────────────────────────────────────┤
│  MCP Server (5 tools)                       │  AI-callable (stdio/HTTP/SSE)
├─────────────────────────────────────────────┤
│  CLI Tool (10 commands)                      │  Human-callable (cobra-based)
├─────────────────────────────────────────────┤
│  Go SDK (6 methods + helpers)                │  Programmatic API
└─────────────────────────────────────────────┘
```

---

## 1️⃣ Skills — AI Agent Integration

> **Zero install.** Add the config below and your AI agent can immediately search CT logs, retrieve certificates, investigate CAs, and more.

### One-Click Setup for Claude Code

Add this to your `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "crt-sh-skills": {
      "command": "npx",
      "args": ["-y", "crtsh-skills-mcp-server"]
    }
  }
}
```

Or if you prefer a pre-built binary:

```json
{
  "mcpServers": {
    "crt-sh-skills": {
      "command": "bash",
      "args": ["-c", "OS=$(uname -s | tr '[:upper:]' '[:lower:]'); ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/'); curl -sL https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-mcp-server-${OS}-${ARCH}.tar.gz | tar xz && ./mcp-server --transport stdio"]
    }
  }
}
```

Or for project-local auto-registration, add `.mcp.json` to your project root:

```json
{
  "crt-sh-skills": {
    "command": "go",
    "args": ["run", "github.com/cyberspacesec/crt.sh-skills/cmd/mcp-server@latest", "--transport", "stdio"]
  }
}
```

### Available Skills

| Skill | Trigger | What it does |
|-------|---------|-------------|
| `crtsh-search` | CT log search, subdomain enumeration, domain reconnaissance, certificate search | Search CT logs + all 5 tools |
| `crtsh-cert` | Certificate ID lookup, CA investigation, cert validity checking | Certificate & CA detail lookup |

### What Your AI Can Do

Once connected, your AI agent can:

- 🔍 **Search CT logs** by domain, hash, serial, CA name, IP address, and 16+ more types
- 📜 **Retrieve certificates** by crt.sh ID
- 🏛️ **Investigate CAs** — chain discovery, CA disclosures, revoked intermediates
- 📊 **Access 13 info pages** — monitored CT logs, Mozilla/Apple/Chrome root programs, OCSP responders
- 🔗 **Cross-reference on Censys.io** — build Censys search URLs
- ✅ **Lint certificates** — cablint, x509lint, zlint, keylint

---

## 2️⃣ MCP Server

### Install & Run

**Option A: Download pre-built binary**

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/i686/386/;s/i386/386/')
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-mcp-server-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x mcp-server
```

**Option B: Go install**

```bash
go install github.com/cyberspacesec/crt.sh-skills/cmd/mcp-server@latest
```

**Option C: Clone & build**

```bash
git clone https://github.com/cyberspacesec/crt.sh-skills.git
cd crt.sh-skills && go build -o mcp-server ./cmd/mcp-server/
```

### Transport Modes

```bash
# stdio (Claude Code, Cursor, Windsurf, etc.)
./mcp-server --transport stdio

# HTTP (remote AI agents)
./mcp-server --transport http --addr :8080

# SSE (browser-based clients)
./mcp-server --transport sse --addr :8080 --base-url https://my-server.com
```

### 5 MCP Tools

| Tool | Required Params | Description |
|------|----------------|-------------|
| `search_certificates` | `query` | Search CT logs (22 search types, 7 match modes, linting, pagination) |
| `get_certificate` | `id` | Get certificate by crt.sh ID |
| `get_info_page` | `page` | Access 13 crt.sh info pages |
| `get_ca` | `ca_id` | Get CA certificate details |
| `search_censys` | `query`, `search_type` | Build Censys.io search URL |

<details>
<summary><b>search_certificates — All parameters</b></summary>

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search term |
| `search_type` | string | No | `""`, `c`, `id`, `ctid`, `serial`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `sha1`, `sha256`, `ca`, `CAID`, `CAName`, `Identity`, `CN`, `E`, `OU`, `O`, `dNSName`, `rfc822Name`, `iPAddress` |
| `match` | string | No | `""`, `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `exclude_expired` | boolean | No | Exclude expired certificates |
| `deduplicate` | boolean | No | Deduplicate precertificate pairs |
| `show_sql` | boolean | No | Show SQL query (debugging) |
| `linter` | string | No | `cablint`, `x509lint`, `zlint`, `keylint`, `lint` |
| `lint_type` | string | No | `1 week`, `issues` |
| `page` | number | No | Page number (1-based) |
| `page_size` | number | No | Results per page |

</details>

<details>
<summary><b>get_info_page — 13 available pages</b></summary>

`cert-populations`, `revoked-intermediates`, `ca-issuers`, `ocsp-responders`, `test-websites`, `monitored-logs`, `accepted-roots-missing`, `gen-add-chain`, `mozilla-disclosures`, `mozilla-certvalidations`, `mozilla-onecrl`, `apple-disclosures`, `chrome-disclosures`

</details>

---

## 3️⃣ CLI

### Install

```bash
# Option A: Download pre-built binary
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/i686/386/;s/i386/386/')
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-cli-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x crtsh-cli

# Option B: Go install
go install github.com/cyberspacesec/crt.sh-skills/cmd/crtsh-cli@latest

# Option C: Clone & build
git clone https://github.com/cyberspacesec/crt.sh-skills.git
cd crt.sh-skills && go build -o crtsh-cli ./cmd/crtsh-cli/
```

### Usage

```bash
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
crtsh-cli list-types          # 22 search types
crtsh-cli list-pages          # 13 info pages
crtsh-cli list-linters        # 5 linters
crtsh-cli list-match-modes    # 7 match modes

# Output formats
crtsh-cli search example.com -o json    # JSON
crtsh-cli search example.com -o csv     # CSV
crtsh-cli search example.com -o table   # Table (default)

# Root-level flags
crtsh-cli --timeout 60s search example.com   # Custom timeout
crtsh-cli --debug search example.com          # Debug output
```

<details>
<summary><b>search command — all flags</b></summary>

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-t` | `""` | Search type (22 types) |
| `--match` | `-m` | `""` | Match mode (7 modes) |
| `--exclude-expired` | `-e` | false | Exclude expired certificates |
| `--deduplicate` | `-d` | false | Deduplicate precertificate pairs |
| `--show-sql` | | false | Show SQL query (debugging) |
| `--linter` | | `""` | Linter: cablint, x509lint, zlint, keylint, lint |
| `--lint-type` | | `""` | Lint output: `1 week`, `issues` |
| `--page` | `-p` | 0 | Page number (1-based) |
| `--page-size` | `-s` | 0 | Results per page |
| `--json` | `-j` | false | Shorthand for `--output json` |

</details>

---

## 4️⃣ Go SDK

### Install

```bash
go get github.com/cyberspacesec/crt.sh-skills/pkg/crtsh
```

### Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    crtsh "github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"
)

func main() {
    client := crtsh.NewClient(
        crtsh.WithTimeout(10 * time.Second),
        crtsh.WithRetryCount(5),
    )

    // Search certificates
    certs, _, err := client.SearchCertificates(context.Background(), crtsh.QueryParams{
        Q:              "example.com",
        Deduplicate:    true,
        ExcludeExpired: true,
    })
    if err != nil {
        if crtsh.IsServerError(err) {
            log.Fatal("crt.sh is having issues:", err)
        }
        log.Fatal(err)
    }
    for _, cert := range certs {
        fmt.Printf("ID: %d, CN: %s, Domains: %v\n", cert.ID, cert.CommonName, cert.Domains)
    }
}
```

### Full API

```go
// Create client with options
client := crtsh.NewClient(
    crtsh.WithTimeout(10 * time.Second),
    crtsh.WithRetryCount(5),
    crtsh.WithDebug(true),
    crtsh.WithUserAgent("my-app/1.0"),
)

// Search certificates (22 search types, 7 match modes, linting, pagination)
certs, pagination, err := client.SearchCertificates(ctx, crtsh.QueryParams{
    SearchType:     "dNSName",
    Q:              "example.com",
    ExcludeExpired: true,
    Deduplicate:    true,
    Page:           1,
    PageSize:       50,
})

// Get certificate by ID
cert, err := client.GetCertificateByID(ctx, 26786991824)

// Get info page (13 pages)
info, err := client.FetchInfoPage(ctx, "monitored-logs")

// Get CA details
ca, err := client.FetchCAByID(ctx, 16418)

// Build Censys URL
url, err := crtsh.BuildCensysURL("CN", "example.com")

// Auto-paginate through all results
err := client.IterateCertificates(ctx, params, func(certs []crtsh.Certificate, pag *crtsh.Pagination) bool {
    for _, cert := range certs {
        fmt.Println(cert.CommonName)
    }
    return true // return false to stop early
})

// Typed error handling
if crtsh.IsNotFoundError(err) { /* 404 */ }
if crtsh.IsRateLimitError(err) { /* 429 */ }
if crtsh.IsServerError(err) { /* 5xx */ }

// Registry functions
types := crtsh.SearchTypes()    // 22 search types
modes := crtsh.MatchModes()     // 7 match modes
linters := crtsh.Linters()      // 5 linters
lintTypes := crtsh.LintTypes()  // 2 lint output types
```

<details>
<summary><b>Certificate Model</b></summary>

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

</details>

---

## 📦 Pre-built Binaries

Download from [GitHub Releases](https://github.com/cyberspacesec/crt.sh-skills/releases/latest). Available for **9 platform combinations**:

| Platform | Architecture | MCP Server | CLI Tool |
|----------|-------------|------------|----------|
| Linux | amd64 | `crtsh-skills-mcp-server-linux-amd64.tar.gz` | `crtsh-skills-cli-linux-amd64.tar.gz` |
| Linux | arm64 | `crtsh-skills-mcp-server-linux-arm64.tar.gz` | `crtsh-skills-cli-linux-arm64.tar.gz` |
| Linux | 386 | `crtsh-skills-mcp-server-linux-386.tar.gz` | `crtsh-skills-cli-linux-386.tar.gz` |
| macOS | amd64 | `crtsh-skills-mcp-server-darwin-amd64.tar.gz` | `crtsh-skills-cli-darwin-amd64.tar.gz` |
| macOS | arm64 | `crtsh-skills-mcp-server-darwin-arm64.tar.gz` | `crtsh-skills-cli-darwin-arm64.tar.gz` |
| Windows | amd64 | `crtsh-skills-mcp-server-windows-amd64.zip` | `crtsh-skills-cli-windows-amd64.zip` |
| Windows | arm64 | `crtsh-skills-mcp-server-windows-arm64.zip` | `crtsh-skills-cli-windows-arm64.zip` |
| Windows | 386 | `crtsh-skills-mcp-server-windows-386.zip` | `crtsh-skills-cli-windows-386.zip` |
| FreeBSD | amd64 | `crtsh-skills-mcp-server-freebsd-amd64.tar.gz` | `crtsh-skills-cli-freebsd-amd64.tar.gz` |

**Verify checksum:**
```bash
curl -sL https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/checksums.txt -o checksums.txt
sha256sum -c --ignore-missing checksums.txt
```

---

## 🛠️ Development

```bash
# Run tests
go test -v -race ./pkg/crtsh/...

# Build all binaries
go build -ldflags "-X main.Version=$(git describe --tags --always)" -o mcp-server ./cmd/mcp-server/
go build -ldflags "-X main.Version=$(git describe --tags --always)" -o crtsh-cli ./cmd/crtsh-cli/

# Dry-run GoReleaser (no publish)
goreleaser release --snapshot --clean

# Create a release
git tag v1.2.0
git push origin v1.2.0
# GitHub Actions + GoReleaser will build and publish binaries automatically
```

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<a id="简体中文"></a>

## 简体中文

[crt.sh-skills](https://crt.sh/) 证书透明度搜索引擎的完整封装 —— 支持 **Skills**、MCP、CLI、Go SDK 四种方式接入。**crt.sh 的每一个功能都已封装，一个不漏。**

**English** | [简体中文](#简体中文)

---

## 🚀 四种接入方式

| # | 方式 | 适用场景 | 一句话说明 |
|---|------|---------|-----------|
| 1 | **Skills** | AI Agent（Claude Code、Cursor 等） | 零安装 — 只需添加配置 |
| 2 | **MCP Server** | 任何兼容 MCP 的 AI 工具 | `./mcp-server --transport stdio` |
| 3 | **CLI** | 人工使用、脚本、流水线 | `crtsh-cli search example.com` |
| 4 | **Go SDK** | Go 程序、集成开发 | `import crtsh "github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"` |

四种方式暴露**完全相同的能力** —— 任何一个功能都不会只在某一种方式中存在。

```
┌─────────────────────────────────────────────┐
│  Skills (.claude/skills/)                    │  AI 可读文档（触发式）
├─────────────────────────────────────────────┤
│  MCP Server (5 tools)                       │  AI 可调用（stdio/HTTP/SSE）
├─────────────────────────────────────────────┤
│  CLI Tool (10 命令)                          │  人类可调用（cobra 风格）
├─────────────────────────────────────────────┤
│  Go SDK (6 方法 + 辅助函数)                   │  编程式 API
└─────────────────────────────────────────────┘
```

---

## 1️⃣ Skills — AI Agent 接入

> **零安装。** 添加以下配置，你的 AI Agent 即可立即搜索 CT 日志、获取证书、调查 CA 等。

### 一键配置（Claude Code）

将以下内容添加到 `~/.claude/settings.json`：

```json
{
  "mcpServers": {
    "crt-sh-skills": {
      "command": "npx",
      "args": ["-y", "crtsh-skills-mcp-server"]
    }
  }
}
```

或者使用预编译二进制：

```json
{
  "mcpServers": {
    "crt-sh-skills": {
      "command": "bash",
      "args": ["-c", "OS=$(uname -s | tr '[:upper:]' '[:lower:]'); ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/'); curl -sL https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-mcp-server-${OS}-${ARCH}.tar.gz | tar xz && ./mcp-server --transport stdio"]
    }
  }
}
```

或者项目级自动注册，在项目根目录添加 `.mcp.json`：

```json
{
  "crt-sh-skills": {
    "command": "go",
    "args": ["run", "github.com/cyberspacesec/crt.sh-skills/cmd/mcp-server@latest", "--transport", "stdio"]
  }
}
```

### 可用 Skills

| Skill | 触发条件 | 功能 |
|-------|---------|------|
| `crtsh-search` | CT 日志搜索、子域名枚举、域名侦察、证书搜索 | 搜索 CT 日志 + 全部 5 个工具 |
| `crtsh-cert` | 证书 ID 查询、CA 调查、证书有效性检查 | 证书与 CA 详情查询 |

### AI Agent 能做什么

连接后，你的 AI Agent 可以：

- 🔍 **搜索 CT 日志** — 按域名、哈希、序列号、CA 名称、IP 地址等 22 种类型搜索
- 📜 **获取证书** — 通过 crt.sh ID 获取证书详情
- 🏛️ **调查 CA** — 证书链发现、CA 披露、已吊销中间证书
- 📊 **访问 13 个信息页** — 受监控的 CT 日志、Mozilla/Apple/Chrome 根证书计划、OCSP 响应器
- 🔗 **在 Censys.io 上交叉引用** — 构建 Censys 搜索 URL
- ✅ **证书合规检查** — cablint、x509lint、zlint、keylint

---

## 2️⃣ MCP Server

### 安装与运行

```bash
# 方式 A：下载预编译二进制
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/i686/386/;s/i386/386/')
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-mcp-server-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x mcp-server

# 方式 B：Go install
go install github.com/cyberspacesec/crt.sh-skills/cmd/mcp-server@latest

# 方式 C：克隆并编译
git clone https://github.com/cyberspacesec/crt.sh-skills.git
cd crt.sh-skills && go build -o mcp-server ./cmd/mcp-server/
```

### 传输模式

```bash
./mcp-server --transport stdio              # Claude Code、Cursor、Windsurf 等
./mcp-server --transport http --addr :8080  # 远程 AI Agent
./mcp-server --transport sse --addr :8080 --base-url https://my-server.com  # 浏览器客户端
```

### 5 个 MCP 工具

| 工具 | 必填参数 | 说明 |
|------|---------|------|
| `search_certificates` | `query` | 搜索 CT 日志（22 种搜索类型、7 种匹配模式、合规检查、分页） |
| `get_certificate` | `id` | 通过 crt.sh ID 获取证书 |
| `get_info_page` | `page` | 访问 13 个 crt.sh 信息页 |
| `get_ca` | `ca_id` | 获取 CA 证书详情 |
| `search_censys` | `query`, `search_type` | 构建 Censys.io 搜索 URL |

---

## 3️⃣ CLI

### 安装

```bash
# 方式 A：下载预编译二进制
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/i686/386/;s/i386/386/')
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-cli-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x crtsh-cli

# 方式 B：Go install
go install github.com/cyberspacesec/crt.sh-skills/cmd/crtsh-cli@latest

# 方式 C：克隆并编译
git clone https://github.com/cyberspacesec/crt.sh-skills.git
cd crt.sh-skills && go build -o crtsh-cli ./cmd/crtsh-cli/
```

### 使用示例

```bash
# 搜索证书
crtsh-cli search example.com --exclude-expired --deduplicate
crtsh-cli search ABCDEF1234 --type sha256
crtsh-cli search "Let's Encrypt" --type CAName

# 获取证书详情
crtsh-cli get-cert 26786991824 --json

# 获取信息页
crtsh-cli info-page monitored-logs

# 获取 CA 详情
crtsh-cli get-ca 16418

# 构建 Censys URL
crtsh-cli censys "example.com" --type CN

# 列出可用选项
crtsh-cli list-types          # 22 种搜索类型
crtsh-cli list-pages          # 13 个信息页
crtsh-cli list-linters        # 5 种合规检查工具
crtsh-cli list-match-modes    # 7 种匹配模式

# 输出格式
crtsh-cli search example.com -o json    # JSON
crtsh-cli search example.com -o csv     # CSV
crtsh-cli search example.com -o table   # 表格（默认）
```

---

## 4️⃣ Go SDK

### 安装

```bash
go get github.com/cyberspacesec/crt.sh-skills/pkg/crtsh
```

### 快速示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    crtsh "github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"
)

func main() {
    client := crtsh.NewClient(
        crtsh.WithTimeout(10 * time.Second),
        crtsh.WithRetryCount(5),
    )

    // 搜索证书
    certs, _, err := client.SearchCertificates(context.Background(), crtsh.QueryParams{
        Q:              "example.com",
        Deduplicate:    true,
        ExcludeExpired: true,
    })
    if err != nil {
        if crtsh.IsServerError(err) {
            log.Fatal("crt.sh 服务异常:", err)
        }
        log.Fatal(err)
    }
    for _, cert := range certs {
        fmt.Printf("ID: %d, CN: %s, Domains: %v\n", cert.ID, cert.CommonName, cert.Domains)
    }
}
```

### 完整 API

```go
// 创建客户端
client := crtsh.NewClient(
    crtsh.WithTimeout(10 * time.Second),
    crtsh.WithRetryCount(5),
    crtsh.WithDebug(true),
    crtsh.WithUserAgent("my-app/1.0"),
)

// 搜索证书（22 种搜索类型、7 种匹配模式、合规检查、分页）
certs, pagination, err := client.SearchCertificates(ctx, crtsh.QueryParams{
    SearchType:     "dNSName",
    Q:              "example.com",
    ExcludeExpired: true,
    Deduplicate:    true,
    Page:           1,
    PageSize:       50,
})

// 通过 ID 获取证书
cert, err := client.GetCertificateByID(ctx, 26786991824)

// 获取信息页（13 个页面）
info, err := client.FetchInfoPage(ctx, "monitored-logs")

// 获取 CA 详情
ca, err := client.FetchCAByID(ctx, 16418)

// 构建 Censys URL
url, err := crtsh.BuildCensysURL("CN", "example.com")

// 自动翻页遍历所有结果
err := client.IterateCertificates(ctx, params, func(certs []crtsh.Certificate, pag *crtsh.Pagination) bool {
    for _, cert := range certs {
        fmt.Println(cert.CommonName)
    }
    return true // 返回 false 提前停止
})

// 类型化错误处理
if crtsh.IsNotFoundError(err) { /* 404 */ }
if crtsh.IsRateLimitError(err) { /* 429 */ }
if crtsh.IsServerError(err) { /* 5xx */ }

// 注册表函数
types := crtsh.SearchTypes()    // 22 种搜索类型
modes := crtsh.MatchModes()     // 7 种匹配模式
linters := crtsh.Linters()      // 5 种合规检查工具
lintTypes := crtsh.LintTypes()  // 2 种合规输出类型
```

---

## 📦 预编译二进制

从 [GitHub Releases](https://github.com/cyberspacesec/crt.sh-skills/releases/latest) 下载，支持 **9 种平台组合**：

| 平台 | 架构 | MCP Server | CLI 工具 |
|------|------|------------|---------|
| Linux | amd64 | `crtsh-skills-mcp-server-linux-amd64.tar.gz` | `crtsh-skills-cli-linux-amd64.tar.gz` |
| Linux | arm64 | `crtsh-skills-mcp-server-linux-arm64.tar.gz` | `crtsh-skills-cli-linux-arm64.tar.gz` |
| Linux | 386 | `crtsh-skills-mcp-server-linux-386.tar.gz` | `crtsh-skills-cli-linux-386.tar.gz` |
| macOS | amd64 | `crtsh-skills-mcp-server-darwin-amd64.tar.gz` | `crtsh-skills-cli-darwin-amd64.tar.gz` |
| macOS | arm64 | `crtsh-skills-mcp-server-darwin-arm64.tar.gz` | `crtsh-skills-cli-darwin-arm64.tar.gz` |
| Windows | amd64 | `crtsh-skills-mcp-server-windows-amd64.zip` | `crtsh-skills-cli-windows-amd64.zip` |
| Windows | arm64 | `crtsh-skills-mcp-server-windows-arm64.zip` | `crtsh-skills-cli-windows-arm64.zip` |
| Windows | 386 | `crtsh-skills-mcp-server-windows-386.zip` | `crtsh-skills-cli-windows-386.zip` |
| FreeBSD | amd64 | `crtsh-skills-mcp-server-freebsd-amd64.tar.gz` | `crtsh-skills-cli-freebsd-amd64.tar.gz` |

**校验文件完整性：**
```bash
curl -sL https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/checksums.txt -o checksums.txt
sha256sum -c --ignore-missing checksums.txt
```

---

## 🛠️ 开发

```bash
# 运行测试
go test -v -race ./pkg/crtsh/...

# 构建所有二进制
go build -ldflags "-X main.Version=$(git describe --tags --always)" -o mcp-server ./cmd/mcp-server/
go build -ldflags "-X main.Version=$(git describe --tags --always)" -o crtsh-cli ./cmd/crtsh-cli/

# GoReleaser 试运行（不发布）
goreleaser release --snapshot --clean

# 创建发布版本
git tag v1.2.0
git push origin v1.2.0
# GitHub Actions + GoReleaser 会自动构建并发布二进制
```

## 许可证

MIT License — 详见 [LICENSE](LICENSE)。

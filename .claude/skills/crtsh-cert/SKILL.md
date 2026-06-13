---
name: crtsh-cert
description: Use when retrieving specific certificate or CA details from crt.sh. Triggers on certificate ID lookup, CA investigation, cert validity checking, or when user provides a crt.sh numeric ID or CA ID.
allowed-tools: ["mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__get_ca", "mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_info_page", "mcp__go-crt-sh__search_censys"]
---

# crt.sh — Certificate & CA Detail Lookup

> Retrieve detailed information about specific certificates and Certificate Authorities from crt.sh.

---

## 📦 Installation

### Option 1: Download Pre-built Binary (Recommended)

No Go SDK required. Download from [GitHub Releases](https://github.com/cyberspacesec/crt.sh-skills/releases/latest):

```bash
# Detect platform and download
OS=$(uname -s | tr A-Z a-z)
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-mcp-server-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x crtsh-skills-mcp-server-*

# Download CLI tool
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-cli-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x crtsh-skills-cli-*
```

### Option 2: Clone & Build from Source

Requires Go 1.23+:

```bash
git clone https://github.com/cyberspacesec/crt.sh-skills.git
cd crt.sh-skills
go build -o mcp-server ./cmd/mcp-server/
go build -o crtsh-cli ./cmd/crtsh-cli/
```

---

## ⚡ 30-Second Quick Start

**Get a certificate by ID:**

```
get_certificate(id=26786991824)
```

CLI: `crtsh-cli get-cert 26786991824`

**Get CA certificate details:**

```
get_ca(ca_id=16418)
```

CLI: `crtsh-cli get-ca 16418`

**Investigate a certificate's CA chain:**

```
search_certificates(query="example.com", exclude_expired=true)
→ get issuer_ca_id from results →
get_ca(ca_id=<issuer_ca_id>)
```

---

## 🔧 5 Available Tools

| Tool | One-liner | Required params |
|------|----------|----------------|
| `get_certificate` | Get a certificate by its crt.sh ID | `id` |
| `get_ca` | Get CA certificate details | `ca_id` |
| `search_certificates` | Find certs when you don't have the ID | `query` |
| `get_info_page` | Get CA-related info pages | `page` |
| `search_censys` | Cross-reference on Censys.io | `query`, `search_type` |

---

## 📖 Common Use Cases

### Direct Certificate Lookup

User provides a crt.sh certificate ID:

```
get_certificate(id=26786991824)
```

Returns: common_name, issuer_name, serial_number, domains, not_before, not_after, entry_timestamp.

CLI: `crtsh-cli get-cert 26786991824 --json`

### Domain → Certificate → CA Chain

User wants to know who issued a domain's certificate:

1. `search_certificates(query="example.com", exclude_expired=true, page_size=1)` → find cert
2. Get `issuer_ca_id` from result
3. `get_ca(ca_id=<issuer_ca_id>)` → get CA details

CLI:
```bash
crtsh-cli search example.com -e --page-size 1 --json
crtsh-cli get-ca <issuer_ca_id> --json
```

### CA Disclosure Investigation

User asks about a root program's CA disclosures:

```
get_info_page(page="mozilla-disclosures")
get_info_page(page="revoked-intermediates")
```

CLI: `crtsh-cli info-page mozilla-disclosures`

---

## 🎛️ Tool Parameters

### `get_certificate`

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | number | **Yes** | The crt.sh certificate ID |

CLI: `crtsh-cli get-cert [id]` — `--json` / `-j` for JSON output.

### `get_ca`

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `ca_id` | number | **Yes** | The crt.sh CA ID (from `issuer_ca_id` in search results) |

CLI: `crtsh-cli get-ca [ca-id]` — `--json` / `-j` for JSON output.

### `get_info_page`

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `page` | string | **Yes** | Info page name (see table below) |

CLI: `crtsh-cli info-page [page-name]` — `--json` / `-j` for JSON output.

<details>
<summary><b>CA-related info pages</b></summary>

| Page | Description |
|------|-------------|
| `ca-issuers` | CA issuer information |
| `revoked-intermediates` | Revoked intermediate CA certificates |
| `mozilla-disclosures` | Mozilla CA certificate disclosures |
| `mozilla-certvalidations` | Mozilla cert validation requirements |
| `mozilla-onecrl` | Mozilla certificate revocation list |
| `apple-disclosures` | Apple CA certificate disclosures |
| `chrome-disclosures` | Chrome CA certificate disclosures |

All 13 pages: `crtsh-cli list-pages`

</details>

### `search_certificates` (when you need to find a cert first)

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `query` | string | **Yes** | Search term |
| `search_type` | string | No | Search type (22 types — see crtsh-search skill) |
| `exclude_expired` | boolean | No | Exclude expired certificates |
| `deduplicate` | boolean | No | Deduplicate precertificate pairs |

<details>
<summary><b>All search_certificates parameters</b></summary>

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `query` | string | — | Search term (required) |
| `search_type` | string | `""` | 22 types (see crtsh-search skill) |
| `match` | string | `""` | Match mode: `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `exclude_expired` | boolean | false | Exclude expired |
| `deduplicate` | boolean | false | Deduplicate precerts |
| `show_sql` | boolean | false | Show SQL (debug) |
| `linter` | string | `""` | Linter: cablint, x509lint, zlint, keylint, lint |
| `lint_type` | string | `""` | `1 week` or `issues` |
| `page` | number | 0 | Page number |
| `page_size` | number | 0 | Results per page |

</details>

---

## 📋 Certificate Result Format

`get_certificate` returns:

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | crt.sh certificate ID |
| `issuer_ca_id` | int | CA ID → use with `get_ca` |
| `common_name` | string | Certificate commonName |
| `issuer_name` | string | Full issuer distinguished name |
| `domains` | []string | Deduplicated, wildcard-stripped domain list |
| `serial_number` | string | Certificate serial number |
| `not_before` | timestamp | Validity start |
| `not_after` | timestamp | Validity end |
| `entry_timestamp` | timestamp \| null | CT log entry time |

---

## 💻 CLI Quick Reference

```bash
crtsh-cli get-cert 26786991824              # Certificate details
crtsh-cli get-cert 26786991824 --json       # JSON output
crtsh-cli get-ca 16418                      # CA certificate details
crtsh-cli get-ca 16418 --json               # JSON output
crtsh-cli info-page mozilla-disclosures     # CA disclosure info
crtsh-cli search example.com -e --json      # Find certs first
```

---

## ⚠️ Important Notes

- **crt.sh can be slow** — returns 5xx during peak load. SDK retries automatically (3 retries, exponential backoff).
- **Null timestamps** — `entry_timestamp` can be null for some certificates (e.g. "Issuer Not Found").
- **All CLI commands support `--json`** for machine-readable output.
- **Don't have the ID?** Use `search_certificates` first, then `get_certificate` or `get_ca` for details.

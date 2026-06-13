---
name: crtsh-search
description: Use when searching certificate transparency logs for domains, subdomains, SSL/TLS certificates, or certificate transparency data via crt.sh. Triggers on mentions of CT logs, subdomain enumeration, certificate search, SSL fingerprint lookup, domain reconnaissance, or CA investigation.
allowed-tools: ["mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__get_info_page", "mcp__go-crt-sh__get_ca", "mcp__go-crt-sh__search_censys"]
---

# crt.sh — Certificate Transparency Search

> Search CT logs, retrieve certificates, investigate CAs, access info pages, and build Censys URLs.

---

## 📦 Installation

### Option 1: Download Pre-built Binary (Recommended)

No Go SDK required. Download the binary for your platform from [GitHub Releases](https://github.com/cyberspacesec/crt.sh-skills/releases/latest):

```bash
# Detect platform and download MCP server
OS=$(uname -s | tr A-Z a-z)
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-mcp-server-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x crtsh-skills-mcp-server-*

# Download CLI tool
curl -sL "https://github.com/cyberspacesec/crt.sh-skills/releases/latest/download/crtsh-skills-cli-${OS}-${ARCH}.tar.gz" | tar xz
chmod +x crtsh-skills-cli-*
```

**Connect to Claude Code** — add to `~/.claude/settings.json`:
```json
{
  "mcpServers": {
    "crt-sh-skills": {
      "command": "/path/to/crtsh-skills-mcp-server-linux-amd64",
      "args": ["--transport", "stdio"]
    }
  }
}
```

### Option 2: Clone & Build from Source

Requires Go 1.23+:

```bash
git clone https://github.com/cyberspacesec/crt.sh-skills.git
cd crt.sh-skills

# Build MCP server
go build -o mcp-server ./cmd/mcp-server/

# Build CLI tool
go build -o crtsh-cli ./cmd/crtsh-cli/

# Run tests
go test ./pkg/crtsh/...
```

### Option 3: Go Install

```bash
go install github.com/cyberspacesec/crt.sh-skills/cmd/mcp-server@latest
go install github.com/cyberspacesec/crt.sh-skills/cmd/crtsh-cli@latest
```

---

## ⚡ 30-Second Quick Start

**Most common task — subdomain enumeration:**

```
search_certificates(query="example.com", deduplicate=true, exclude_expired=true)
```

CLI equivalent:
```bash
crtsh-cli search example.com -ed
```

**Get a specific certificate by ID:**

```
get_certificate(id=26786991824)
```

CLI:
```bash
crtsh-cli get-cert 26786991824
```

---

## 🔧 5 Available Tools

| Tool | One-liner | Required params |
|------|----------|----------------|
| `search_certificates` | Search CT logs by domain, hash, serial, CA, etc. | `query` |
| `get_certificate` | Get a certificate by its crt.sh ID | `id` |
| `get_info_page` | Access crt.sh info pages (CA disclosures, CT logs, etc.) | `page` |
| `get_ca` | Get CA certificate details | `ca_id` |
| `search_censys` | Build a Censys.io search URL | `query`, `search_type` |

---

## 📖 Common Use Cases

### Subdomain Enumeration

Search CT logs for all domains under a target:

```
search_certificates(query="target.com", deduplicate=true, exclude_expired=true)
```

The `domains` field in each result gives deduplicated, wildcard-stripped domain names.

CLI: `crtsh-cli search target.com -ed`

### Certificate Fingerprint Lookup

Find a certificate by its SHA-256 or SHA-1 fingerprint:

```
search_certificates(query="ABCD1234EF56...", search_type="sha256")
```

Use `search_type="c"` to search both SHA-1 and SHA-256 simultaneously.

CLI: `crtsh-cli search ABCD1234EF56... --type sha256`

### CA Investigation

Search by CA name, then get CA details:

```
search_certificates(query="Let's Encrypt", search_type="CAName")
→ get issuer_ca_id from results →
get_ca(ca_id=<issuer_ca_id>)
```

CLI: `crtsh-cli search "Let's Encrypt" --type CAName` → `crtsh-cli get-ca <id>`

### Certificate Linting

Check certificates for compliance issues:

```
search_certificates(query="example.com", linter="zlint", lint_type="issues")
```

CLI: `crtsh-cli search example.com --linter zlint --lint-type issues`

### CT Log & CA Disclosure Info

Access crt.sh information pages:

```
get_info_page(page="monitored-logs")
get_info_page(page="mozilla-onecrl")
```

CLI: `crtsh-cli info-page monitored-logs`

---

## 🎛️ `search_certificates` — All Parameters

### Core Parameters (most used)

| Param | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `query` | string | **Yes** | — | Search term (domain, hash, serial, CA name, etc.) |
| `search_type` | string | No | `""` | Type of search (see **Search Types** below) |
| `exclude_expired` | boolean | No | false | Exclude expired certificates |
| `deduplicate` | boolean | No | false | Deduplicate precertificate pairs |

### Advanced Parameters (less common)

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `match` | string | `""` | Match mode: `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `linter` | string | `""` | Linter: `cablint`, `x509lint`, `zlint`, `keylint`, `lint` (all) |
| `lint_type` | string | `""` | Lint output: `1 week` (summary), `issues` (issues only) |
| `show_sql` | boolean | false | Show the SQL query crt.sh uses (debugging) |
| `page` | number | 0 | Page number for pagination (1-based) |
| `page_size` | number | 0 | Results per page |

<details>
<summary><b>Search Types — 22 options</b></summary>

| `search_type` | Description | Example query |
|---------------|-------------|---------------|
| *(empty, default)* | General search (domain name) | `example.com` |
| `c` | Certificate fingerprint (SHA-1 or SHA-256) | `ABCD1234...` |
| `id` | crt.sh certificate ID | `26786991824` |
| `ctid` | CT Entry ID | `12345` |
| `serial` | Serial number | `00:11:22:33` |
| `ski` | Subject Key Identifier | `ABC123...` |
| `spkisha1` | SHA-1(SubjectPublicKeyInfo) | `ABC123...` |
| `spkisha256` | SHA-256(SubjectPublicKeyInfo) | `ABC123...` |
| `subjectsha1` | SHA-1(Subject) | `ABC123...` |
| `sha1` | SHA-1(Certificate) | `ABC123...` |
| `sha256` | SHA-256(Certificate) | `ABC123...` |
| `ca` | CA (general) | `DigiCert` |
| `CAID` | CA ID | `16418` |
| `CAName` | CA Name | `Let's Encrypt` |
| `Identity` | Identity | `example.com` |
| `CN` | commonName (Subject) | `example.com` |
| `E` | emailAddress (Subject) | `admin@example.com` |
| `OU` | organizationalUnitName (Subject) | `Engineering` |
| `O` | organizationName (Subject) | `Example Inc` |
| `dNSName` | dNSName (SAN) | `example.com` |
| `rfc822Name` | rfc822Name (SAN) | `user@example.com` |
| `iPAddress` | iPAddress (SAN) | `1.2.3.4` |

CLI: `crtsh-cli list-types` to see this table interactively.

</details>

<details>
<summary><b>Match Modes — 7 options</b></summary>

| `match` | SQL equivalent | When to use |
|---------|---------------|-------------|
| *(empty)* | Auto | Let crt.sh pick the best mode |
| `=` | exact | Exact identity match |
| `ILIKE` | ILIKE | Case-insensitive pattern match |
| `LIKE` | LIKE | Case-sensitive pattern match |
| `single` | — | Match single identity value |
| `any` | — | Match any identity value |
| `FTS` | full text search | Full text search across all fields |

</details>

---

## 📄 `get_info_page` — All 13 Pages

| Page | What it shows |
|------|-------------|
| `monitored-logs` | CT logs monitored by crt.sh |
| `cert-populations` | Certificate population statistics |
| `ca-issuers` | CA issuer information |
| `revoked-intermediates` | Revoked intermediate CA certificates |
| `ocsp-responders` | OCSP responder details |
| `mozilla-onecrl` | Mozilla certificate revocation list |
| `mozilla-disclosures` | Mozilla CA certificate disclosures |
| `mozilla-certvalidations` | Mozilla cert validation requirements |
| `apple-disclosures` | Apple CA certificate disclosures |
| `chrome-disclosures` | Chrome CA certificate disclosures |
| `test-websites` | Test websites for cert validation |
| `accepted-roots-missing` | Roots accepted but missing from DB |
| `gen-add-chain` | Certificate submission assistant |

CLI: `crtsh-cli list-pages` to see this table interactively.

---

## 🔗 `search_censys` — Censys.io URL Builder

Builds a Censys.io search URL equivalent to crt.sh's searchCensys feature.

**Supported search types:** `c`, `serial`, `sha1`, `sha256`, `ca`, `CAName`, `Identity`, `CN`, `OU`, `O`, `dNSName`, `rfc822Name`, `iPAddress`

**NOT supported by Censys:** `id`, `ctid`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `E`, `CAID`

```
search_censys(query="example.com", search_type="CN")
→ Returns: https://search.censys.io/search?resource=certificates&q=...
```

CLI: `crtsh-cli censys "example.com" --type CN`

---

## 💻 CLI Reference — All Commands

<details>
<summary><b>crtsh-cli search [query] — flags</b></summary>

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `""` | Search type (22 types) |
| `--match` | `-m` | string | `""` | Match mode (7 modes) |
| `--exclude-expired` | `-e` | bool | false | Exclude expired certificates |
| `--deduplicate` | `-d` | bool | false | Deduplicate precertificate pairs |
| `--show-sql` | | bool | false | Show SQL query (debugging) |
| `--linter` | | string | `""` | Linter: cablint, x509lint, zlint, keylint, lint |
| `--lint-type` | | string | `""` | Lint output: `1 week`, `issues` |
| `--page` | `-p` | int | 0 | Page number (1-based) |
| `--page-size` | `-s` | int | 0 | Results per page |
| `--json` | `-j` | bool | false | JSON output |

</details>

<details>
<summary><b>crtsh-cli other commands</b></summary>

| Command | Flags | Description |
|---------|-------|-------------|
| `crtsh-cli get-cert [id]` | `--json` `-j` | Get certificate by ID |
| `crtsh-cli info-page [page]` | `--json` `-j` | Get info page |
| `crtsh-cli get-ca [ca-id]` | `--json` `-j` | Get CA certificate details |
| `crtsh-cli censys [query]` | `--type` `-t` | Build Censys.io URL |
| `crtsh-cli list-types` | — | List all 22 search types |
| `crtsh-cli list-pages` | — | List all 13 info pages |

</details>

---

## 📋 Result Format

Each certificate from `search_certificates` contains:

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | crt.sh certificate ID → use with `get_certificate` |
| `issuer_ca_id` | int | CA ID → use with `get_ca` |
| `common_name` | string | Certificate commonName |
| `issuer_name` | string | Full issuer distinguished name |
| `domains` | []string | Deduplicated, wildcard-stripped domain list |
| `serial_number` | string | Certificate serial number |
| `not_before` | timestamp | Validity start |
| `not_after` | timestamp | Validity end |
| `entry_timestamp` | timestamp \| null | CT log entry time (can be null) |

---

## ⚠️ Important Notes

- **crt.sh can be slow** — returns 5xx during peak load. SDK retries automatically (3 retries, exponential backoff).
- **Wildcard stripping** — `*.example.com` appears as `example.com` in `domains`.
- **Null timestamps** — `entry_timestamp` can be null for some certificates (e.g. "Issuer Not Found").
- **Pagination** — If `pagination.next_page` is set, more results are available. Use `page` + `page_size` to iterate.
- **All CLI commands support `--json`** for machine-readable output.

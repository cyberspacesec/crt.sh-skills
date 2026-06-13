---
name: crtsh-cert
description: Use when retrieving specific certificate or CA details from crt.sh. Triggers on certificate ID lookup, CA investigation, cert validity checking, or when user provides a crt.sh numeric ID or CA ID.
allowed-tools: ["mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__get_ca", "mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_info_page", "mcp__go-crt-sh__search_censys"]
---

# crt.sh — Certificate & CA Detail Lookup

> Retrieve detailed information about specific certificates and Certificate Authorities from crt.sh.

## Tool Quick Reference

| Tool | What it does | Key params |
|------|-------------|-----------|
| `get_certificate` | Get cert by ID | id |
| `get_ca` | Get CA cert details | ca_id |
| `search_certificates` | Find certs when you don't have the ID | query, search_type |
| `get_info_page` | Get CA-related info pages | page |
| `search_censys` | Cross-reference on Censys.io | query, search_type |

## CLI Reference — All Commands & Flags

### `crtsh-cli get-cert [id]` — Get certificate by crt.sh ID

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |

Examples:
```bash
crtsh-cli get-cert 26786991824          # Human-readable: CN, issuer, validity, domains
crtsh-cli get-cert 26786991824 --json   # JSON output for programmatic use
```

### `crtsh-cli get-ca [ca-id]` — Get CA certificate details

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |

Examples:
```bash
crtsh-cli get-ca 16418          # Human-readable CA certificate details
crtsh-cli get-ca 16418 --json   # JSON output
```

### `crtsh-cli info-page [page-name]` — Get crt.sh info page

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |

CA-related info pages:

| Page | Description |
|------|-------------|
| `ca-issuers` | CA issuer information |
| `revoked-intermediates` | Revoked intermediate CAs |
| `mozilla-disclosures` | Mozilla CA certificate disclosures |
| `mozilla-certvalidations` | Mozilla cert validation requirements |
| `mozilla-onecrl` | Mozilla certificate revocation list |
| `apple-disclosures` | Apple CA certificate disclosures |
| `chrome-disclosures` | Chrome CA certificate disclosures |

### `crtsh-cli search [query]` — Search (when you need to find a cert first)

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `""` | Search type (22 types, see crtsh-search skill) |
| `--exclude-expired` | `-e` | bool | false | Exclude expired certificates |
| `--deduplicate` | `-d` | bool | false | Deduplicate precertificate pairs |
| `--page` | `-p` | int | 0 | Page number (1-based) |
| `--page-size` | `-s` | int | 0 | Results per page |
| `--json` | `-j` | bool | false | Output as JSON |

### `crtsh-cli censys [query]` — Build Censys.io URL

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `CN` | Search type for Censys |

## When to Use

- User provides a numeric crt.sh certificate ID
- User wants to investigate a Certificate Authority by CA ID
- User needs full certificate details (issuer, validity, domains, serial)
- User wants to trace the CA chain from a certificate
- User asks about specific CA disclosures (Mozilla, Apple, Chrome)

## When NOT to Use

- User wants to search for certificates by domain → use `crtsh-search` skill
- User wants bulk subdomain enumeration → use `crtsh-search` skill

## Instructions

### Certificate Lookup

1. If the user provides a crt.sh ID directly, call `get_certificate(id=<id>)`
2. If the user provides a domain/hash, first use `search_certificates` to find the cert, then `get_certificate` for details
3. Present the full certificate details:
   - Common Name, Issuer Name, Serial Number
   - Validity period (not_before → not_after)
   - All domain names (from `domains` array)
   - Entry timestamp (when logged to CT)

### CA Investigation

1. Get the `issuer_ca_id` from a certificate's search results
2. Call `get_ca(ca_id=<issuer_ca_id>)` for CA certificate details
3. For CA disclosure info, use `get_info_page`:
   - `ca-issuers` — General CA issuer information
   - `mozilla-disclosures` — Mozilla root program disclosures
   - `apple-disclosures` — Apple root program disclosures
   - `chrome-disclosures` — Chrome root program disclosures
   - `revoked-intermediates` — Revoked intermediate CAs

### Certificate Chain Investigation

When the user wants to understand the full certificate chain:

1. `search_certificates(query="example.com")` → find the cert
2. `get_certificate(id=<cert_id>)` → get cert details + issuer_ca_id
3. `get_ca(ca_id=<issuer_ca_id>)` → get CA cert details

## Examples

### Example 1: Direct ID lookup

User: "Show me certificate 26786991824"
→ `get_certificate(id=26786991824)`
→ CLI: `crtsh-cli get-cert 26786991824 --json`

### Example 2: Domain → cert → CA chain

User: "What CA issued the certificate for example.com?"

1. `search_certificates(query="example.com", exclude_expired=true, page_size=1)`
2. Get `issuer_ca_id` from result
3. `get_ca(ca_id=<issuer_ca_id>)`

CLI equivalent:
```bash
crtsh-cli search example.com -e --page-size 1 --json
crtsh-cli get-ca <issuer_ca_id> --json
```

### Example 3: CA disclosure investigation

User: "What are Mozilla's CA disclosures?"
→ `get_info_page(page="mozilla-disclosures")`
→ CLI: `crtsh-cli info-page mozilla-disclosures --json`

### Example 4: Check for revoked intermediates

User: "Are there any revoked intermediate CAs?"
→ `get_info_page(page="revoked-intermediates")`
→ CLI: `crtsh-cli info-page revoked-intermediates`

## Notes

- crt.sh can return 5xx during peak load — SDK retries automatically (3 retries)
- `entry_timestamp` can be null for some certificates
- All CLI commands support `--json` for machine-readable output
- Use `crtsh-cli list-pages` to see all 13 available info pages

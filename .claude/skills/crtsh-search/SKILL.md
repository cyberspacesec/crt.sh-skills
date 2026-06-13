---
name: crtsh-search
description: Use when searching certificate transparency logs for domains, subdomains, SSL/TLS certificates, or certificate transparency data via crt.sh. Triggers on mentions of CT logs, subdomain enumeration, certificate search, SSL fingerprint lookup, domain reconnaissance, or CA investigation.
allowed-tools: ["mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__get_info_page", "mcp__go-crt-sh__get_ca", "mcp__go-crt-sh__search_censys"]
---

# crt.sh — Certificate Transparency Search & Intelligence

> Complete crt.sh wrapper: search CT logs, retrieve certificates, investigate CAs, access info pages, and build Censys URLs.

## Tool Quick Reference

| Tool | What it does | Key params |
|------|-------------|-----------|
| `search_certificates` | Search CT logs | query, search_type, exclude_expired, deduplicate, linter |
| `get_certificate` | Get cert by ID | id |
| `get_info_page` | Get crt.sh info page | page (13 pages available) |
| `get_ca` | Get CA cert details | ca_id |
| `search_censys` | Build Censys.io URL | query, search_type |

## CLI Quick Reference

```bash
crtsh-cli search example.com -ed                    # Search with exclude-expired + deduplicate
crtsh-cli search ABCDEF --type sha256               # Search by SHA-256 fingerprint
crtsh-cli search "Let's Encrypt" --type CAName      # Search by CA name
crtsh-cli get-cert 26786991824 --json               # Get certificate by ID
crtsh-cli info-page monitored-logs                  # Get CT log info
crtsh-cli get-ca 16418                              # Get CA certificate details
crtsh-cli censys "example.com" --type CN            # Build Censys.io URL
crtsh-cli list-types                                # List all 22 search types
crtsh-cli list-pages                                # List all 13 info pages
```

## When to Use

- Subdomain enumeration via CT logs
- SSL/TLS certificate search and investigation
- Certificate fingerprint lookup (SHA-1, SHA-256)
- CA (Certificate Authority) investigation
- Certificate linting (compliance checking)
- CT log monitoring and CA disclosure review
- Censys.io cross-reference search

## When NOT to Use

- DNS records → use DNS tools
- WHOIS info → use WHOIS tools
- Website content → use web tools

## Instructions

### Step 1: Choose the right search_type

| User wants | search_type | Example query |
|-----------|------------|---------------|
| All certs for a domain | `""` (default) | `example.com` |
| Cert by SHA-1/SHA-256 fingerprint | `c` | `ABCD1234...` |
| Cert by crt.sh ID | `id` | `26786991824` |
| Subdomains via DNS SAN | `dNSName` | `example.com` |
| Cert by SHA-256 fingerprint | `sha256` | `ABCD1234...` |
| Cert by serial number | `serial` | `00:11:22:33` |
| Certs by Common Name | `CN` | `example.com` |
| Certs by CA name | `CAName` | `Let's Encrypt` |
| Certs by CA (general) | `ca` | `DigiCert` |
| Certs by CA ID | `CAID` | `16418` |
| Certs by email | `E` | `admin@example.com` |
| Certs by organization | `O` | `Example Inc` |
| Certs by IP address | `iPAddress` | `1.2.3.4` |

### Step 2: Search with appropriate options

**Subdomain enumeration** (most common):
```
search_certificates(query="example.com", deduplicate=true, exclude_expired=true)
```

**Certificate fingerprint lookup**:
```
search_certificates(query="ABCD1234...", search_type="sha256")
```

**Certificate linting**:
```
search_certificates(query="example.com", linter="zlint", lint_type="issues")
```

### Step 3: Parse results

Each certificate contains:
- `id` → use with `get_certificate` for full details
- `issuer_ca_id` → use with `get_ca` for CA investigation
- `domains` → deduplicated, wildcard-stripped domain list
- `common_name`, `issuer_name`, `serial_number`
- `not_before`, `not_after` → validity period

### Step 4: Deep-dive (if needed)

- **Certificate details**: `get_certificate(id=<cert_id>)`
- **CA investigation**: `get_ca(ca_id=<issuer_ca_id>)`
- **Info pages**: `get_info_page(page="monitored-logs")`
- **Censys cross-reference**: `search_censys(query="example.com", search_type="CN")`

### Available Info Pages

| Page | What it shows |
|------|-------------|
| `monitored-logs` | CT logs monitored by crt.sh |
| `revoked-intermediates` | Revoked intermediate CAs |
| `ca-issuers` | CA issuer information |
| `ocsp-responders` | OCSP responder details |
| `cert-populations` | Certificate population stats |
| `test-websites` | Test websites for cert validation |
| `accepted-roots-missing` | Roots accepted but missing from DB |
| `gen-add-chain` | Certificate submission assistant |
| `mozilla-disclosures` | Mozilla CA certificate disclosures |
| `mozilla-certvalidations` | Mozilla cert validation requirements |
| `mozilla-onecrl` | Mozilla certificate revocation list |
| `apple-disclosures` | Apple CA certificate disclosures |
| `chrome-disclosures` | Chrome CA certificate disclosures |

## Notes

- crt.sh can be slow or return 5xx during peak load — SDK retries automatically
- `search_censys` does NOT support: `id`, `ctid`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `E`, `CAID`
- Wildcard certs (`*.example.com`) are stripped to base domain in `domains`
- `entry_timestamp` can be null for some certificates

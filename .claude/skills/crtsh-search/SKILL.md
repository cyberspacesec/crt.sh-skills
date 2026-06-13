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

## CLI Reference — All Commands & Flags

### `crtsh-cli search [query]` — Search certificate transparency logs

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `""` | Search type. 22 types available (see table below) |
| `--match` | `-m` | string | `""` | Match mode: `=`, `ILIKE`, `LIKE`, `single`, `any`, `FTS` |
| `--exclude-expired` | `-e` | bool | false | Exclude expired certificates from results |
| `--deduplicate` | `-d` | bool | false | Deduplicate precertificate pairs |
| `--show-sql` | | bool | false | Show the SQL query crt.sh uses (for debugging) |
| `--linter` | | string | `""` | Run certificate linter: `cablint`, `x509lint`, `zlint`, `keylint`, `lint` (all) |
| `--lint-type` | | string | `""` | Lint output type: `1 week` (summary), `issues` (issues only) |
| `--page` | `-p` | int | 0 | Page number for pagination (1-based) |
| `--page-size` | `-s` | int | 0 | Number of results per page |
| `--json` | `-j` | bool | false | Output results as JSON |

Examples:
```bash
crtsh-cli search example.com -ed                     # Default search, exclude expired + deduplicate
crtsh-cli search example.com -ed --json               # Same, but JSON output
crtsh-cli search ABCDEF1234 --type sha256             # Search by SHA-256 fingerprint
crtsh-cli search "Let's Encrypt" --type CAName        # Search by CA name
crtsh-cli search example.com --type dNSName           # Search by DNS SAN
crtsh-cli search example.com --linter zlint --lint-type issues  # Lint with zlint
crtsh-cli search example.com --show-sql               # Debug: see the SQL query
crtsh-cli search example.com -p 2 -s 50               # Page 2, 50 results per page
crtsh-cli search example.com --match ILIKE            # Case-insensitive pattern match
```

**All 22 search types:**

| `--type` value | Description |
|----------------|-------------|
| *(empty, default)* | General search (domain name) |
| `c` | Certificate fingerprint (SHA-1 or SHA-256) |
| `id` | crt.sh certificate ID |
| `ctid` | CT Entry ID |
| `serial` | Serial number |
| `ski` | Subject Key Identifier |
| `spkisha1` | SHA-1(SubjectPublicKeyInfo) |
| `spkisha256` | SHA-256(SubjectPublicKeyInfo) |
| `subjectsha1` | SHA-1(Subject) |
| `sha1` | SHA-1(Certificate) |
| `sha256` | SHA-256(Certificate) |
| `ca` | CA (general) |
| `CAID` | CA ID |
| `CAName` | CA Name |
| `Identity` | Identity |
| `CN` | commonName (Subject) |
| `E` | emailAddress (Subject) |
| `OU` | organizationalUnitName (Subject) |
| `O` | organizationName (Subject) |
| `dNSName` | dNSName (SAN) |
| `rfc822Name` | rfc822Name (SAN) |
| `iPAddress` | iPAddress (SAN) |

### `crtsh-cli get-cert [id]` — Get certificate by crt.sh ID

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |

Example:
```bash
crtsh-cli get-cert 26786991824          # Human-readable output
crtsh-cli get-cert 26786991824 --json   # JSON output
```

### `crtsh-cli info-page [page-name]` — Get crt.sh information page

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |

**All 13 info pages:**

| Page | Description |
|------|-------------|
| `cert-populations` | Certificate population statistics |
| `revoked-intermediates` | Revoked intermediate CA certificates |
| `ca-issuers` | CA issuer information |
| `ocsp-responders` | OCSP responder details |
| `test-websites` | Test websites for cert validation |
| `monitored-logs` | CT logs monitored by crt.sh |
| `accepted-roots-missing` | Roots accepted but missing from DB |
| `gen-add-chain` | Certificate submission assistant |
| `mozilla-disclosures` | Mozilla CA certificate disclosures |
| `mozilla-certvalidations` | Mozilla cert validation requirements |
| `mozilla-onecrl` | Mozilla certificate revocation list |
| `apple-disclosures` | Apple CA certificate disclosures |
| `chrome-disclosures` | Chrome CA certificate disclosures |

### `crtsh-cli get-ca [ca-id]` — Get CA certificate details

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |

Example:
```bash
crtsh-cli get-ca 16418          # Human-readable output
crtsh-cli get-ca 16418 --json   # JSON output
```

### `crtsh-cli censys [query]` — Build Censys.io search URL

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type` | `-t` | string | `CN` | Search type for Censys (see list below) |

**Censys-supported search types:** `c`, `serial`, `sha1`, `sha256`, `ca`, `CAName`, `Identity`, `CN`, `OU`, `O`, `dNSName`, `rfc822Name`, `iPAddress`

**NOT supported by Censys:** `id`, `ctid`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `E`, `CAID`

Example:
```bash
crtsh-cli censys "example.com" --type CN           # Search by common name
crtsh-cli censys "example.com" --type dNSName      # Search by DNS SAN
```

### `crtsh-cli list-types` — List all search types

No flags. Outputs a table of all 22 search types with descriptions.

### `crtsh-cli list-pages` — List all info pages

No flags. Outputs a table of all 13 info pages with titles and descriptions.

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

**Debug SQL query**:
```
search_certificates(query="example.com", show_sql=true)
```

**Pagination**:
```
search_certificates(query="example.com", page=1, page_size=50)
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

## Notes

- crt.sh can be slow or return 5xx during peak load — SDK retries automatically (3 retries with exponential backoff)
- `search_censys` does NOT support: `id`, `ctid`, `ski`, `spkisha1`, `spkisha256`, `subjectsha1`, `E`, `CAID`
- Wildcard certs (`*.example.com`) are stripped to base domain in `domains`
- `entry_timestamp` can be null for some certificates (e.g. "Issuer Not Found" entries)
- All CLI commands support `--json` for machine-readable output
- Use `crtsh-cli list-types` and `crtsh-cli list-pages` to discover available options

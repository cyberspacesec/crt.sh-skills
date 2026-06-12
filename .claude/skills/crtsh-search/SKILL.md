---
name: crtsh-search
description: Use when searching certificate transparency logs for domains, subdomains, SSL/TLS certificates, or certificate transparency data via crt.sh. Triggers on mentions of CT logs, subdomain enumeration, certificate search, SSL fingerprint lookup, or domain reconnaissance.
allowed-tools: ["mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__get_info_page"]
---

# crt.sh Certificate Search

> Search certificate transparency logs to discover domains, subdomains, and certificate details.

## When to Use

- User asks about domains or subdomains under a target domain
- User wants to find SSL/TLS certificates for a domain
- User needs certificate transparency log data for reconnaissance
- User asks to enumerate subdomains via CT logs
- User wants to find certificates by hash, serial number, or CA
- User wants to lint certificates (check for compliance issues)
- User asks about CT log monitoring, CA disclosures, or revocation lists

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
| Cert by SHA-1 or SHA-256 fingerprint | `c` | `ABCD1234...` |
| Subdomains via DNS SAN | `dNSName` | `example.com` |
| Cert by SHA-256 fingerprint | `sha256` | `ABCD1234...` |
| Cert by serial number | `serial` | `00:11:22:33` |
| Certs by Common Name | `CN` | `example.com` |
| Certs by CA name | `CAName` | `Let's Encrypt` |
| Certs by CA (general) | `ca` | `DigiCert` |
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

**For certificate linting:**
- Set `linter` to the desired linter: `cablint`, `x509lint`, `zlint`, `keylint`, or `lint` (all)
- Set `lint_type` to `"1 week"` for summary or `"issues"` for issues only

**For pagination:**
- Start with `page=1` and `page_size=50`
- If `pagination.next_page` is set, there are more results
- Continue fetching until all pages are retrieved or user is satisfied

### Step 3: Parse and present results

The response contains a `certificates` array. Each certificate has:
- `id` — crt.sh certificate ID (use for get_certificate)
- `common_name` — the certificate's commonName
- `issuer_name` — full issuer distinguished name
- `name_value` — domain names associated with the cert
- `domains` — deduplicated, wildcard-stripped domain list
- `entry_timestamp` — when the cert was logged
- `not_before` / `not_after` — certificate validity period
- `issuer_ca_id` — the CA that issued the cert
- `serial_number` — certificate serial number
- `result_count` — number of matching results

**Present results as:**
1. A summary: "Found N certificates for domain X"
2. A deduplicated domain/subdomain list (extract from all `domains` fields)
3. Key certificate details if user asked specifically

### Step 4: Deep-dive with get_certificate (if needed)

If the user wants details on a specific certificate, call `get_certificate` with the `id` from the search results.

### Step 5: Info pages (if needed)

If the user asks about CT log monitoring, CA disclosures, or revocation information, call `get_info_page` with the appropriate page name:
- `monitored-logs` — CT logs monitored by crt.sh
- `revoked-intermediates` — Revoked intermediate CAs
- `mozilla-onecrl` — Mozilla's certificate revocation list
- `mozilla-disclosures`, `apple-disclosures`, `chrome-disclosures` — CA disclosures per root program
- `ca-issuers` — CA issuer information
- `ocsp-responders` — OCSP responder details
- `cert-populations` — Certificate population statistics

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
2. Filter results where `issuer_name` contains "Let's Encrypt"
3. Present filtered results

### Example 4: Linting

User: "Check example.com certificates for linting issues"

1. Call `search_certificates(query="example.com", linter="zlint", lint_type="issues", deduplicate=true, exclude_expired=true)`
2. Present any linting issues found

### Example 5: CT log information

User: "What CT logs does crt.sh monitor?"

1. Call `get_info_page(page="monitored-logs")`
2. Present the monitored log information

## Notes

- crt.sh can be slow or return 5xx errors during peak load — the SDK retries automatically
- For large domains, results can span many pages — ask user if they want all pages
- Wildcard certificates (`*.example.com`) are stripped to the base domain in the `domains` field
- The `name_value` field may contain multiple domains separated by newlines
- Info pages return HTML content — parse accordingly

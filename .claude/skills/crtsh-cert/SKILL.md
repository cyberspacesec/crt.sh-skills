---
name: crtsh-cert
description: Use when retrieving a specific certificate's full details from crt.sh by its ID, or when investigating a Certificate Authority. Triggers on mentions of certificate ID lookup, specific cert details, CA investigation, or when user provides a crt.sh numeric ID or CA ID.
allowed-tools: ["mcp__go-crt-sh__get_certificate", "mcp__go-crt-sh__get_ca", "mcp__go-crt-sh__search_certificates", "mcp__go-crt-sh__get_info_page", "mcp__go-crt-sh__search_censys"]
---

# crt.sh Certificate & CA Detail Lookup

> Retrieve detailed information about a specific certificate or Certificate Authority from crt.sh.

## Available Tools

| Tool | Purpose |
|------|---------|
| `get_certificate` | Get specific certificate by crt.sh ID |
| `get_ca` | Get CA certificate details by issuer_ca_id |
| `search_certificates` | Search for certificates when you don't have the ID |
| `get_info_page` | Access crt.sh info pages |
| `search_censys` | Build Censys.io search URL |

## CLI Usage

```bash
# Get certificate by ID
crtsh-cli get-cert 26786991824
crtsh-cli get-cert 26786991824 --json

# Get CA details
crtsh-cli get-ca 16418

# Search first, then get details
crtsh-cli search example.com --exclude-expired
crtsh-cli get-cert <id-from-results>
```

## When to Use

- User provides a numeric crt.sh certificate ID and wants full details
- User wants to investigate a Certificate Authority by its CA ID
- User found a certificate in search results and wants to deep-dive
- User asks about a specific certificate's issuer, validity, or other details

## When NOT to Use

- User wants to search for certificates by domain → use `crtsh-search` skill instead
- User asks about general CT log information → explain conceptually, then offer search

## Instructions

### For Certificate Lookup

1. Obtain the certificate ID (from user input or previous search results)
2. If the user provides a domain name or hash instead of an ID, first use `search_certificates` to find the cert
3. Call `get_certificate` with the `id` parameter
4. Present the full certificate details including:
   - Common Name, Issuer, Serial Number
   - Validity period (not_before / not_after)
   - All domain names
   - Entry timestamp

### For CA Lookup

1. Obtain the CA ID (from `issuer_ca_id` in search results)
2. Call `get_ca` with the `ca_id` parameter
3. Present the CA certificate details

## Examples

### Example 1: Direct ID lookup

User: "Show me certificate 9999999"

1. Call `get_certificate(id=9999999)`
2. Present: "Certificate #9999999: CN=example.com, Issued by Cloudflare TLS Issuing RSA CA 3..."

### Example 2: Domain to certificate to CA chain

User: "What CA issued the most recent certificate for example.com?"

1. Call `search_certificates(query="example.com", exclude_expired=true, page_size=1)`
2. Get the `issuer_ca_id` from the first result
3. Call `get_ca(ca_id=<issuer_ca_id>)`
4. Present the full CA details

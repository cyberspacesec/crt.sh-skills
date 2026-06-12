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

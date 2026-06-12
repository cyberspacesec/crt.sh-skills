// tools.go — MCP tool definitions and handlers for crt.sh
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cyberspacesec/go-crt.sh/pkg/crtsh"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(s *server.MCPServer, client *crtsh.Client) {
	// Tool: search_certificates
	searchTool := mcp.NewTool("search_certificates",
		mcp.WithDescription("Search certificate transparency logs via crt.sh. "+
			"Returns certificates matching the query. Supports searching by domain name, "+
			"SHA-256 hash, serial number, CA name, and many other criteria."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search term: domain name, hash, serial number, or other identifier depending on search_type"),
		),
		mcp.WithString("search_type",
			mcp.Description("Type of search to perform. Default is general search (empty string). "+
				"Common values: empty=general search, c=certificate fingerprint (SHA-1/SHA-256), "+
				"sha256=SHA-256 fingerprint, serial=serial number, "+
				"CN=commonName, dNSName=DNS SAN, iPAddress=IP address SAN, ca=CA search"),
			mcp.Enum("", "c", "id", "ctid", "serial", "ski", "spkisha1", "spkisha256",
				"subjectsha1", "sha1", "sha256", "ca", "CAID", "CAName",
				"Identity", "CN", "E", "OU", "O", "dNSName", "rfc822Name", "iPAddress"),
		),
		mcp.WithString("match",
			mcp.Description("Identity matching mode"),
			mcp.Enum("", "=", "ILIKE", "LIKE", "single", "any", "FTS"),
		),
		mcp.WithBoolean("exclude_expired",
			mcp.Description("Exclude expired certificates from results"),
		),
		mcp.WithBoolean("deduplicate",
			mcp.Description("Deduplicate (pre)certificate pairs"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number for pagination (1-based)"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("Number of results per page"),
		),
		mcp.WithString("linter",
			mcp.Description("Run certificate linting with the specified linter. "+
				"Returns linting results alongside certificate search results."),
			mcp.Enum("", "cablint", "x509lint", "zlint", "keylint", "lint"),
		),
		mcp.WithString("lint_type",
			mcp.Description("Linting output type: '1 week' for 1-week summary, 'issues' for issues only"),
			mcp.Enum("", "1 week", "issues"),
		),
	)
	s.AddTool(searchTool, searchCertificatesHandler(client))

	// Tool: get_certificate
	getCertTool := mcp.NewTool("get_certificate",
		mcp.WithDescription("Retrieve a specific certificate from crt.sh by its ID. "+
			"Returns detailed certificate information including issuer, validity dates, and domains."),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("The crt.sh certificate ID (numeric)"),
		),
	)
	s.AddTool(getCertTool, getCertificateHandler(client))
}

func searchCertificatesHandler(client *crtsh.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'query': %v", err)), nil
		}

		searchType := req.GetString("search_type", "")
		match := req.GetString("match", "")
		excludeExpired := req.GetBool("exclude_expired", false)
		deduplicate := req.GetBool("deduplicate", false)
		page := int(req.GetFloat("page", 0))
		pageSize := int(req.GetFloat("page_size", 0))
		linter := req.GetString("linter", "")
		lintType := req.GetString("lint_type", "")

		params := crtsh.QueryParams{
			SearchType:     searchType,
			Match:          match,
			ExcludeExpired: excludeExpired,
			Deduplicate:    deduplicate,
			Linter:         linter,
			LintType:       lintType,
			Page:           page,
			PageSize:       pageSize,
		}

		// Route query to the correct field based on search_type
		switch searchType {
		case "c":
			// Certificate fingerprint search (SHA-1 or SHA-256) — crt.sh matches against both
			params.Q = query
		case "id":
			params.ID = query
		case "ctid":
			params.CTID = query
		case "serial":
			params.Serial = query
		case "ski":
			params.SKI = query
		case "spkisha1":
			params.SPKISHA1 = query
		case "spkisha256":
			params.SPKISHA256 = query
		case "subjectsha1":
			params.SubjectSHA1 = query
		case "sha1":
			params.SHA1 = query
		case "sha256":
			params.SHA256 = query
		case "CAID":
			params.CAID = query
		case "CAName":
			params.CAName = query
		case "ca":
			// CA search — searches by CA ID or name (crt.sh handles routing)
			params.Q = query
		case "Identity":
			params.Identity = query
		case "CN":
			params.CN = query
		case "E":
			params.E = query
		case "OU":
			params.OU = query
		case "O":
			params.O = query
		case "dNSName":
			params.DNSName = query
		case "rfc822Name":
			params.RFC822Name = query
		case "iPAddress":
			params.IPAddress = query
		default:
			params.Q = query
		}

		certs, pagination, err := client.SearchCertificates(ctx, params)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		result := map[string]interface{}{
			"certificates": certs,
		}
		if pagination != nil {
			result["pagination"] = pagination
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

func getCertificateHandler(client *crtsh.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idFloat, err := req.RequireFloat("id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'id': %v", err)), nil
		}
		id := int(idFloat)

		cert, err := client.GetCertificateByID(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get certificate failed: %v", err)), nil
		}

		data, err := json.MarshalIndent(cert, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

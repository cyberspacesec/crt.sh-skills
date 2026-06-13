// tools.go — MCP tool definitions and handlers for crt.sh
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"
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
		mcp.WithBoolean("show_sql",
			mcp.Description("Show the SQL query used by crt.sh for this search (for debugging)"),
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

	// Tool: get_info_page
	infoPageTool := mcp.NewTool("get_info_page",
		mcp.WithDescription("Retrieve information pages from crt.sh such as CT log status, "+
			"CA disclosures, revoked intermediates, OCSP responders, and Mozilla/Apple/Chrome root program data."),
		mcp.WithString("page",
			mcp.Required(),
			mcp.Description("The info page to retrieve. Available: cert-populations, revoked-intermediates, "+
				"ca-issuers, ocsp-responders, test-websites, monitored-logs, accepted-roots-missing, "+
				"gen-add-chain, mozilla-disclosures, mozilla-certvalidations, mozilla-onecrl, "+
				"apple-disclosures, chrome-disclosures"),
			mcp.Enum("cert-populations", "revoked-intermediates", "ca-issuers", "ocsp-responders",
				"test-websites", "monitored-logs", "accepted-roots-missing", "gen-add-chain",
				"mozilla-disclosures", "mozilla-certvalidations", "mozilla-onecrl",
				"apple-disclosures", "chrome-disclosures"),
		),
	)
	s.AddTool(infoPageTool, getInfoPageHandler(client))

	// Tool: get_ca
	caTool := mcp.NewTool("get_ca",
		mcp.WithDescription("Retrieve CA (Certificate Authority) certificate details from crt.sh by CA ID. "+
			"Returns the CA certificate page content including issuer chain and certificate details."),
		mcp.WithNumber("ca_id",
			mcp.Required(),
			mcp.Description("The crt.sh CA ID (numeric, from issuer_ca_id field in certificate results)"),
		),
	)
	s.AddTool(caTool, getCAHandler(client))

	// Tool: search_censys
	censysTool := mcp.NewTool("search_censys",
		mcp.WithDescription("Build a Censys.io certificate search URL equivalent to crt.sh's searchCensys feature. "+
			"Returns a URL you can open to search Censys.io for the same certificate data. "+
			"Not all crt.sh search types are supported by Censys (id, ctid, ski, spkisha1, spkisha256, subjectsha1, E are unsupported)."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search term (same as would be used in crt.sh)"),
		),
		mcp.WithString("search_type",
			mcp.Required(),
			mcp.Description("Search type (same values as search_certificates, but Censys does not support: id, ctid, ski, spkisha1, spkisha256, subjectsha1, E, CAID)"),
			mcp.Enum("c", "serial", "sha1", "sha256", "ca", "CAName", "Identity", "CN", "OU", "O", "dNSName", "rfc822Name", "iPAddress"),
		),
	)
	s.AddTool(censysTool, searchCensysHandler(client))
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
		showSQL := req.GetBool("show_sql", false)
		page := int(req.GetFloat("page", 0))
		pageSize := int(req.GetFloat("page_size", 0))
		linter := req.GetString("linter", "")
		lintType := req.GetString("lint_type", "")

		params := crtsh.QueryParams{
			SearchType:     searchType,
			Match:          match,
			ExcludeExpired: excludeExpired,
			Deduplicate:    deduplicate,
			ShowSQL:        showSQL,
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

func getInfoPageHandler(client *crtsh.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		page, err := req.RequireString("page")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'page': %v", err)), nil
		}

		info, err := client.FetchInfoPage(ctx, page)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch info page: %v", err)), nil
		}

		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

func getCAHandler(client *crtsh.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		caIDFloat, err := req.RequireFloat("ca_id")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'ca_id': %v", err)), nil
		}
		caID := int(caIDFloat)

		info, err := client.FetchCAByID(ctx, caID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch CA info: %v", err)), nil
		}

		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

func searchCensysHandler(client *crtsh.Client) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'query': %v", err)), nil
		}

		searchType, err := req.RequireString("search_type")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'search_type': %v", err)), nil
		}

		censysURL, err := crtsh.BuildCensysURL(searchType, query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("censys search error: %v", err)), nil
		}

		result := map[string]string{
			"search_type": searchType,
			"query":       query,
			"censys_url":  censysURL,
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

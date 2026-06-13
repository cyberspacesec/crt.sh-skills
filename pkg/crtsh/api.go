// api.go — Core SDK client for crt.sh Certificate Transparency search
package crtsh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout       = 30 * time.Second
	maxBodySize    int64 = 10 * 1024 * 1024 // 10MB
)

// Client is the crt.sh API client.
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	RetryCount  int
	Debug       bool
	UserAgent   string
	rateLimiter <-chan time.Time
}

// NewClient creates a new crt.sh API client with the given options.
// With no options, it returns a client with sensible defaults:
//   - BaseURL: https://crt.sh/
//   - Timeout: 30s
//   - RetryCount: 3
//   - UserAgent: crt.sh-Go-SDK/1.0
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		BaseURL:    "https://crt.sh/",
		HTTPClient: &http.Client{Timeout: defaultTimeout},
		RetryCount: 3,
		UserAgent:  "Mozilla/5.0 (compatible; crt.sh-Go-SDK/1.0)",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SearchCertificates searches certificate transparency logs via crt.sh.
// It supports 22 search types, 7 match modes, linting, and pagination.
func (c *Client) SearchCertificates(ctx context.Context, params QueryParams) ([]Certificate, *Pagination, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, nil, &Error{Type: ErrorTypeRequest, Message: "url parse error", Cause: err}
	}

	query := c.buildQuery(params)
	u.RawQuery = query.Encode()

	var (
		certs []Certificate
		resp  *http.Response
		body  []byte
	)

	for attempt := 0; attempt <= c.RetryCount; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			return nil, nil, &Error{Type: ErrorTypeRequest, Message: "request creation failed", Cause: err}
		}

		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json")

		resp, body, err = c.doRequest(req)
		if err != nil {
			if attempt == c.RetryCount {
				return nil, nil, &Error{Type: ErrorTypeSearch, Message: "maximum retries exceeded", Cause: err}
			}
			select {
			case <-time.After(exponentialBackoff(attempt)):
				continue
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
		}

		// Retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			if attempt == c.RetryCount {
				return nil, nil, c.parseAPIError(resp.StatusCode, body)
			}
			select {
			case <-time.After(exponentialBackoff(attempt)):
				continue
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
		}

		break
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, c.parseAPIError(resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, &certs); err != nil {
		return nil, nil, &Error{Type: ErrorTypeParse, Message: "json decode error", Cause: err}
	}

	pagination := c.parsePagination(resp, params.Page, params.PageSize)
	return certs, pagination, nil
}

// GetCertificateByID retrieves a specific certificate from crt.sh by its numeric ID.
// crt.sh does not support output=json for the ?id= endpoint, so this method
// uses SearchCertificates with search_type=id instead.
func (c *Client) GetCertificateByID(ctx context.Context, id int) (*Certificate, error) {
	certs, _, err := c.SearchCertificates(ctx, QueryParams{
		SearchType: "id",
		ID:         strconv.Itoa(id),
	})
	if err != nil {
		return nil, &Error{Type: ErrorTypeSearch, Message: "get certificate by id failed", Cause: err}
	}

	if len(certs) == 0 {
		return nil, &Error{Type: ErrorTypeNotFound, Message: fmt.Sprintf("certificate not found: id=%d", id)}
	}

	return &certs[0], nil
}

func (c *Client) buildQuery(params QueryParams) url.Values {
	query := url.Values{}
	query.Set("output", "json")

	// crt.sh JS builds URL as: ?<searchtype>=<value>
	// e.g. ?CN=example.com, NOT ?searchtype=CN&common_name=example.com
	// The searchtype is used AS the URL parameter key directly.
	switch params.SearchType {
	case "c":
		if params.Q != "" {
			query.Set("c", params.Q)
		}
	case "id":
		if params.ID != "" {
			query.Set("id", params.ID)
		}
	case "ctid":
		if params.CTID != "" {
			query.Set("ctid", params.CTID)
		}
	case "serial":
		if params.Serial != "" {
			query.Set("serial", params.Serial)
		}
	case "ski":
		if params.SKI != "" {
			query.Set("ski", params.SKI)
		}
	case "spkisha1":
		if params.SPKISHA1 != "" {
			query.Set("spkisha1", params.SPKISHA1)
		}
	case "spkisha256":
		if params.SPKISHA256 != "" {
			query.Set("spkisha256", params.SPKISHA256)
		}
	case "subjectsha1":
		if params.SubjectSHA1 != "" {
			query.Set("subjectsha1", params.SubjectSHA1)
		}
	case "sha1":
		if params.SHA1 != "" {
			query.Set("sha1", params.SHA1)
		}
	case "sha256":
		if params.SHA256 != "" {
			query.Set("sha256", params.SHA256)
		}
	case "ca":
		if params.Q != "" {
			query.Set("ca", params.Q)
		}
	case "CAID":
		if params.CAID != "" {
			query.Set("CAID", params.CAID)
		}
	case "CAName":
		if params.CAName != "" {
			query.Set("CAName", params.CAName)
		}
	case "Identity":
		if params.Identity != "" {
			query.Set("Identity", params.Identity)
		}
	case "CN":
		if params.CN != "" {
			query.Set("CN", params.CN)
		}
	case "E":
		if params.E != "" {
			query.Set("E", params.E)
		}
	case "OU":
		if params.OU != "" {
			query.Set("OU", params.OU)
		}
	case "O":
		if params.O != "" {
			query.Set("O", params.O)
		}
	case "dNSName":
		if params.DNSName != "" {
			query.Set("dNSName", params.DNSName)
		}
	case "rfc822Name":
		if params.RFC822Name != "" {
			query.Set("rfc822Name", params.RFC822Name)
		}
	case "iPAddress":
		if params.IPAddress != "" {
			query.Set("iPAddress", params.IPAddress)
		}
	default:
		// Default search uses ?q=<value>
		if params.Q != "" {
			query.Set("q", params.Q)
		}
	}

	if params.Match != "" {
		query.Set("match", params.Match)
	}

	// Set boolean flags (using crt.sh URL parameter format from JS)
	if params.ExcludeExpired {
		query.Set("exclude", "expired")
	}
	if params.Deduplicate {
		query.Set("deduplicate", "Y")
	}
	if params.ShowSQL {
		query.Set("showSQL", "Y")
	}

	// Set linting parameters (crt.sh uses linter name as URL param key)
	if params.Linter != "" {
		query.Set(params.Linter, params.LintType)
	}

	// Set pagination
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}
	if params.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(params.PageSize))
	}

	return query
}

// InfoPage represents a crt.sh information page
type InfoPage struct {
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// Available info pages on crt.sh
var InfoPages = map[string]InfoPage{
	"cert-populations": {
		Path:        "cert-populations",
		Title:       "Certificate Populations",
		Description: "Statistics about certificate populations across CT logs",
	},
	"revoked-intermediates": {
		Path:        "revoked-intermediates",
		Title:       "Revoked Intermediates",
		Description: "List of revoked intermediate CA certificates",
	},
	"ca-issuers": {
		Path:        "ca-issuers",
		Title:       "CA Issuers",
		Description: "Certificate Authority issuer information",
	},
	"ocsp-responders": {
		Path:        "ocsp-responders",
		Title:       "OCSP Responders",
		Description: "OCSP responder information for CAs",
	},
	"test-websites": {
		Path:        "test-websites",
		Title:       "Test Websites",
		Description: "Test websites for certificate validation",
	},
	"monitored-logs": {
		Path:        "monitored-logs",
		Title:       "Monitored Logs",
		Description: "CT logs monitored by crt.sh",
	},
	"accepted-roots-missing": {
		Path:        "accepted-roots-missing",
		Title:       "Accepted Roots Missing",
		Description: "Root certificates accepted but missing from crt.sh database",
	},
	"gen-add-chain": {
		Path:        "gen-add-chain",
		Title:       "Certificate Submission Assistant",
		Description: "Tool to help submit certificates to CT logs",
	},
	"mozilla-disclosures": {
		Path:        "mozilla-disclosures",
		Title:       "Mozilla CA Certificate Disclosures",
		Description: "CA certificate disclosures for Mozilla root program",
	},
	"mozilla-certvalidations": {
		Path:        "mozilla-certvalidations",
		Title:       "Mozilla Certificate Validations",
		Description: "Certificate validation requirements for Mozilla root program",
	},
	"mozilla-onecrl": {
		Path:        "mozilla-onecrl",
		Title:       "Mozilla OneCRL",
		Description: "Mozilla's certificate revocation list (OneCRL)",
	},
	"apple-disclosures": {
		Path:        "apple-disclosures",
		Title:       "Apple CA Certificate Disclosures",
		Description: "CA certificate disclosures for Apple root program",
	},
	"chrome-disclosures": {
		Path:        "chrome-disclosures",
		Title:       "Chrome CA Certificate Disclosures",
		Description: "CA certificate disclosures for Chrome root program",
	},
}

// FetchInfoPage retrieves an information page from crt.sh
func (c *Client) FetchInfoPage(ctx context.Context, pagePath string) (*InfoPage, error) {
	info, ok := InfoPages[pagePath]
	if !ok {
		return nil, &Error{
			Type:    ErrorTypeInvalid,
			Message: fmt.Sprintf("unknown info page: %s", pagePath),
		}
	}

	u, err := url.Parse(c.BaseURL + pagePath)
	if err != nil {
		return nil, &Error{Type: ErrorTypeRequest, Message: "url parse error", Cause: err}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, &Error{Type: ErrorTypeRequest, Message: "request creation failed", Cause: err}
	}

	req.Header.Set("User-Agent", c.UserAgent)

	resp, body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseAPIError(resp.StatusCode, body)
	}

	info.Content = string(body)
	return &info, nil
}

// FetchCAByID retrieves CA certificate details from crt.sh by CA ID
func (c *Client) FetchCAByID(ctx context.Context, caID int) (*InfoPage, error) {
	u, err := url.Parse(fmt.Sprintf("%sca?id=%d", c.BaseURL, caID))
	if err != nil {
		return nil, &Error{Type: ErrorTypeRequest, Message: "url parse error", Cause: err}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, &Error{Type: ErrorTypeRequest, Message: "request creation failed", Cause: err}
	}

	req.Header.Set("User-Agent", c.UserAgent)

	resp, body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseAPIError(resp.StatusCode, body)
	}

	return &InfoPage{
		Path:        fmt.Sprintf("ca?id=%d", caID),
		Title:       fmt.Sprintf("CA Certificate #%d", caID),
		Description: "CA certificate details from crt.sh",
		Content:     string(body),
	}, nil
}

// CensysUnsupportedTypes lists search types that Censys does not support.
// These match the types that crt.sh's JS alert()s as unsupported.
var CensysUnsupportedTypes = map[string]bool{
	"id": true, "ctid": true, "ski": true, "spkisha1": true,
	"spkisha256": true, "subjectsha1": true, "E": true,
}

// BuildCensysURL constructs a Censys.io search URL equivalent to crt.sh's searchCensys feature.
// This mirrors the JavaScript logic from crt.sh's advanced search page.
// Returns an error if the search type is not supported by Censys.
func BuildCensysURL(searchType, value string) (string, error) {
	if CensysUnsupportedTypes[searchType] {
		return "", &Error{
			Type:    ErrorTypeInvalid,
			Message: fmt.Sprintf("censys does not support search type: %s", searchType),
		}
	}

	baseURL := "https://search.censys.io/search?resource=certificates&q="

	if value == "%" {
		return baseURL, nil
	}

	var query string
	switch searchType {
	case "c":
		v := strings.ToLower(value)
		query = fmt.Sprintf("fingerprint_sha1:\"%s\" OR fingerprint_sha256:\"%s\"", v, v)
	case "serial":
		v := strings.ToLower(strings.ReplaceAll(value, ":", ""))
		query = fmt.Sprintf("parsed.serial_number_hex:\"%s\"", v)
	case "sha1":
		query = fmt.Sprintf("fingerprint_sha1:\"%s\"", strings.ToLower(value))
	case "sha256":
		query = fmt.Sprintf("fingerprint_sha256:\"%s\"", strings.ToLower(value))
	case "ca", "CAName":
		query = fmt.Sprintf("parsed.issuer_dn:\"%s\"", value)
	case "CAID":
		return "", &Error{
			Type:    ErrorTypeInvalid,
			Message: "censys does not support CAID search",
		}
	case "Identity":
		query = fmt.Sprintf("names:\"%s\"", value)
	case "CN":
		query = fmt.Sprintf("parsed.subject.common_name:\"%s\"", value)
	case "OU":
		query = fmt.Sprintf("parsed.subject.organizational_unit:\"%s\"", value)
	case "O":
		query = fmt.Sprintf("parsed.subject.organization:\"%s\"", value)
	case "dNSName":
		query = fmt.Sprintf("parsed.extensions.subject_alt_name.dns_names:\"%s\"", value)
	case "rfc822Name":
		query = fmt.Sprintf("parsed.extensions.subject_alt_name.email_addresses:\"%s\"", value)
	case "iPAddress":
		query = fmt.Sprintf("parsed.extensions.subject_alt_name.ip_addresses:\"%s\"", value)
	default:
		return "", &Error{
			Type:    ErrorTypeInvalid,
			Message: fmt.Sprintf("unsupported search type for censys: %s", searchType),
		}
	}

	return baseURL + url.QueryEscape(query), nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, []byte, error) {
	if c.rateLimiter != nil {
		<-c.rateLimiter
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, &Error{Type: ErrorTypeRequest, Message: "http request failed", Cause: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, nil, &Error{Type: ErrorTypeParse, Message: "body read error", Cause: err}
	}

	if c.Debug {
		fmt.Printf("[DEBUG] Request: %s %s\n", req.Method, req.URL)
		fmt.Printf("[DEBUG] Response: %s\n", string(body))
	}

	return resp, body, nil
}

func (c *Client) parseAPIError(statusCode int, body []byte) error {
	var apiErr struct {
		Error string `json:"error"`
	}

	// Determine error type based on status code
	var errType ErrorType
	switch {
	case statusCode == http.StatusNotFound:
		errType = ErrorTypeNotFound
	case statusCode == http.StatusTooManyRequests:
		errType = ErrorTypeRateLimit
	case statusCode >= 500:
		errType = ErrorTypeServer
	default:
		errType = ErrorTypeSearch
	}

	msg := fmt.Sprintf("status %d", statusCode)
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		msg = fmt.Sprintf("status %d: %s", statusCode, apiErr.Error)
	}

	return &Error{Type: errType, Message: msg}
}

func (c *Client) parsePagination(resp *http.Response, currentPage, pageSize int) *Pagination {
	p := &Pagination{
		CurrentPage: currentPage,
		PageSize:    pageSize,
	}

	if links := resp.Header.Get("Link"); links != "" {
		if nextPage := parseLinkHeader(links); nextPage > currentPage {
			p.NextPage = nextPage
		}
	}

	return p
}

func parseLinkHeader(linkHeader string) int {
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) < 2 {
			continue
		}

		if strings.TrimSpace(parts[1]) == `rel="next"` {
			urlPart := strings.Trim(parts[0], "<>")
			u, err := url.Parse(urlPart)
			if err != nil {
				return 0
			}
			page := u.Query().Get("page")
			if page == "" {
				return 0
			}
			nextPage, _ := strconv.Atoi(page)
			return nextPage
		}
	}
	return 0
}

func exponentialBackoff(attempt int) time.Duration {
	return time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
}

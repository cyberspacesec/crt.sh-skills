// api.go
package crtsh

import (
	"context"
	"encoding/json"
	"errors"
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

var (
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")
	ErrInvalidResponse    = errors.New("invalid server response")
)

type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	RetryCount  int
	Debug       bool
	UserAgent   string
	rateLimiter <-chan time.Time
}

func NewClient() *Client {
	return &Client{
		BaseURL:    "https://crt.sh/",
		HTTPClient: &http.Client{Timeout: defaultTimeout},
		RetryCount: 3,
		UserAgent:  "Mozilla/5.0 (compatible; crt.sh-Go-SDK/1.0)",
	}
}

func (c *Client) SearchCertificates(ctx context.Context, params QueryParams) ([]Certificate, *Pagination, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("url parse error: %w", err)
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
			return nil, nil, fmt.Errorf("request creation failed: %w", err)
		}

		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json")

		resp, body, err = c.doRequest(req)
		if err != nil {
			if attempt == c.RetryCount {
				return nil, nil, fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, err)
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
		return nil, nil, fmt.Errorf("json decode error: %w", err)
	}

	pagination := c.parsePagination(resp, params.Page, params.PageSize)
	return certs, pagination, nil
}

func (c *Client) GetCertificateByID(ctx context.Context, id int) (*Certificate, error) {
	u, err := url.Parse(fmt.Sprintf("%s?output=json&id=%d", c.BaseURL, id))
	if err != nil {
		return nil, fmt.Errorf("url parse error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseAPIError(resp.StatusCode, body)
	}

	var cert Certificate
	if err := json.Unmarshal(body, &cert); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	return &cert, nil
}

func (c *Client) buildQuery(params QueryParams) url.Values {
	query := url.Values{}
	query.Set("output", "json")

	// Set search type and corresponding parameters
	if params.SearchType != "" {
		query.Set("searchtype", params.SearchType)
	}

	switch params.SearchType {
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
	case "CAID":
		if params.CAID != "" {
			query.Set("ca_id", params.CAID)
		}
	case "CAName":
		if params.CAName != "" {
			query.Set("ca_name", params.CAName)
		}
	case "Identity":
		if params.Identity != "" {
			query.Set("identity", params.Identity)
		}
	case "CN":
		if params.CN != "" {
			query.Set("common_name", params.CN)
		}
	case "E":
		if params.E != "" {
			query.Set("email", params.E)
		}
	case "OU":
		if params.OU != "" {
			query.Set("organizational_unit", params.OU)
		}
	case "O":
		if params.O != "" {
			query.Set("organization", params.O)
		}
	case "dNSName":
		if params.DNSName != "" {
			query.Set("dns_name", params.DNSName)
		}
	case "rfc822Name":
		if params.RFC822Name != "" {
			query.Set("rfc822_name", params.RFC822Name)
		}
	case "iPAddress":
		if params.IPAddress != "" {
			query.Set("ip_address", params.IPAddress)
		}
	}

	// Set general search parameters
	if params.Q != "" {
		query.Set("q", params.Q)
	}
	if params.Match != "" {
		query.Set("match", params.Match)
	}

	// Set boolean flags
	if params.ExcludeExpired {
		query.Set("excludeExpired", "on")
	}
	if params.Deduplicate {
		query.Set("deduplicate", "on")
	}
	if params.ShowSQL {
		query.Set("showSQL", "on")
	}
	if params.SearchCensys {
		query.Set("searchCensys", "on")
	}

	// Set linting parameters
	if params.Linter != "" {
		query.Set("linter", params.Linter)
	}
	if params.LintType != "" {
		query.Set("linttype", params.LintType)
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

func (c *Client) doRequest(req *http.Request) (*http.Response, []byte, error) {
	if c.rateLimiter != nil {
		<-c.rateLimiter
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, nil, fmt.Errorf("body read error: %w", err)
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
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		return fmt.Errorf("api error (%d): %s", statusCode, apiErr.Error)
	}
	return fmt.Errorf("unexpected status code: %d", statusCode)
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

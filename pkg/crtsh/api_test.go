// api_test.go
package crtsh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_SearchCertificates(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("CN") != "example.com" {
			t.Errorf("Unexpected query parameters: %v", r.URL)
		}

		certs := []Certificate{{
			ID:           1,
			IssuerCAID:   123,
			RawNameValue: "example.com\n*.example.com",
			NotBefore:    time.Now().Add(-24 * time.Hour),
			NotAfter:     time.Now().Add(365 * time.Hour),
			SerialNumber: "ABCD1234",
		}}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certs)
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"))

	params := QueryParams{
		SearchType: "CN",
		CN:         "example.com",
		PageSize:   10,
	}
	result, _, err := client.SearchCertificates(context.Background(), params)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(result) != 1 || result[0].ID != 1 {
		t.Errorf("Unexpected result: %+v", result)
	}
	if len(result[0].Domains) != 1 || result[0].Domains[0] != "example.com" {
		t.Errorf("Domain parsing failed: %v", result[0].Domains)
	}
}

func TestClient_SearchCertificates_URLParams(t *testing.T) {
	tests := []struct {
		name           string
		params         QueryParams
		expectedParams map[string]string
	}{
		{
			name: "exclude expired",
			params: QueryParams{
				Q:              "example.com",
				ExcludeExpired: true,
			},
			expectedParams: map[string]string{
				"exclude": "expired",
			},
		},
		{
			name: "deduplicate",
			params: QueryParams{
				Q:           "example.com",
				Deduplicate: true,
			},
			expectedParams: map[string]string{
				"deduplicate": "Y",
			},
		},
		{
			name: "show SQL",
			params: QueryParams{
				Q:       "example.com",
				ShowSQL: true,
			},
			expectedParams: map[string]string{
				"showSQL": "Y",
			},
		},
		{
			name: "linter zlint with issues",
			params: QueryParams{
				Q:        "example.com",
				Linter:   "zlint",
				LintType: "issues",
			},
			expectedParams: map[string]string{
				"zlint": "issues",
			},
		},
		{
			name: "search type c (certificate fingerprint)",
			params: QueryParams{
				SearchType: "c",
				Q:          "ABCDEF123456",
			},
			expectedParams: map[string]string{
				"c": "ABCDEF123456",
			},
		},
		{
			name: "search type ca",
			params: QueryParams{
				SearchType: "ca",
				Q:          "Let's Encrypt",
			},
			expectedParams: map[string]string{
				"ca": "Let's Encrypt",
			},
		},
		{
			name: "default search uses q param",
			params: QueryParams{
				Q: "example.com",
			},
			expectedParams: map[string]string{
				"q": "example.com",
			},
		},
		{
			name: "search type CN uses CN as URL param",
			params: QueryParams{
				SearchType: "CN",
				CN:         "example.com",
			},
			expectedParams: map[string]string{
				"CN": "example.com",
			},
		},
		{
			name: "search type CAID uses CAID as URL param",
			params: QueryParams{
				SearchType: "CAID",
				CAID:       "16418",
			},
			expectedParams: map[string]string{
				"CAID": "16418",
			},
		},
		{
			name: "search type dNSName uses dNSName as URL param",
			params: QueryParams{
				SearchType: "dNSName",
				DNSName:    "example.com",
			},
			expectedParams: map[string]string{
				"dNSName": "example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for key, expected := range tt.expectedParams {
					got := r.URL.Query().Get(key)
					if got != expected {
						t.Errorf("param %q: got %q, want %q", key, got, expected)
					}
				}
				// For c and ca search types, q must NOT be set
				if tt.params.SearchType == "c" || tt.params.SearchType == "ca" {
					if q := r.URL.Query().Get("q"); q != "" {
						t.Errorf("param %q should not be set for search_type %q, got %q", "q", tt.params.SearchType, q)
					}
				}
				// searchtype must NEVER be set as a URL param (crt.sh uses <type>=<value> directly)
				if st := r.URL.Query().Get("searchtype"); st != "" {
					t.Errorf("param %q should never be set in URL, got %q", "searchtype", st)
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]Certificate{})
			}))
			defer testServer.Close()

			client := NewClient(WithBaseURL(testServer.URL + "/"), WithRetryCount(0))

			_, _, _ = client.SearchCertificates(context.Background(), tt.params)
		})
	}
}

func TestClient_HandleHTTPErrors(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"), WithRetryCount(0))

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsServerError(err) {
		t.Errorf("Expected server error, got: %v (type: %T)", err, err)
	}
}

func TestClient_HandleRateLimitError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"), WithRetryCount(0))

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsRateLimitError(err) {
		t.Errorf("Expected rate limit error, got: %v", err)
	}
}

func TestClient_HandleNotFoundError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"), WithRetryCount(0))

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFoundError(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestClient_RetryLogic(t *testing.T) {
	retryCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if retryCount < 2 {
			retryCount++
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Certificate{})
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"), WithRetryCount(3))

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err != nil {
		t.Errorf("Retry failed: %v", err)
	}
	if retryCount != 2 {
		t.Errorf("Expected 2 retries, got %d", retryCount)
	}
}

func TestGetCertificateByID(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "123" {
			t.Errorf("Unexpected request: %v", r.URL)
		}
		if r.URL.Query().Get("searchtype") != "" {
			t.Errorf("searchtype param should NOT be set: %v", r.URL)
		}
		certs := []Certificate{{ID: 123, SerialNumber: "TEST123"}}
		json.NewEncoder(w).Encode(certs)
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"))

	cert, err := client.GetCertificateByID(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	if cert.ID != 123 || cert.SerialNumber != "TEST123" {
		t.Errorf("Unexpected certificate data: %+v", cert)
	}
}

func TestGetCertificateByID_NotFound(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Certificate{})
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"))

	_, err := client.GetCertificateByID(context.Background(), 999999999)
	if err == nil {
		t.Error("Expected error for non-existent certificate")
	}
	if !IsNotFoundError(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestBuildCensysURL(t *testing.T) {
	tests := []struct {
		name        string
		searchType  string
		value       string
		wantErr     bool
		wantContain string
	}{
		{
			name:        "CN search",
			searchType:  "CN",
			value:       "example.com",
			wantErr:     false,
			wantContain: "parsed.subject.common_name",
		},
		{
			name:        "dNSName search",
			searchType:  "dNSName",
			value:       "example.com",
			wantErr:     false,
			wantContain: "subject_alt_name.dns_names",
		},
		{
			name:        "sha256 fingerprint",
			searchType:  "sha256",
			value:       "ABCD1234",
			wantErr:     false,
			wantContain: "fingerprint_sha256",
		},
		{
			name:        "unsupported type id",
			searchType:  "id",
			value:       "123",
			wantErr:     true,
			wantContain: "",
		},
		{
			name:        "unsupported type ski",
			searchType:  "ski",
			value:       "abc",
			wantErr:     true,
			wantContain: "",
		},
		{
			name:        "unsupported type CAID",
			searchType:  "CAID",
			value:       "16418",
			wantErr:     true,
			wantContain: "",
		},
		{
			name:        "c fingerprint (both sha1 and sha256)",
			searchType:  "c",
			value:       "ABCD1234",
			wantErr:     false,
			wantContain: "fingerprint_sha1",
		},
		{
			name:        "serial number",
			searchType:  "serial",
			value:       "00:11:22:33",
			wantErr:     false,
			wantContain: "serial_number_hex",
		},
		{
			name:        "iPAddress",
			searchType:  "iPAddress",
			value:       "1.2.3.4",
			wantErr:     false,
			wantContain: "subject_alt_name.ip_addresses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := BuildCensysURL(tt.searchType, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for search_type %q", tt.searchType)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !strings.Contains(url, "search.censys.io") {
				t.Errorf("URL should be on censys.io, got: %s", url)
			}
			if tt.wantContain != "" && !strings.Contains(url, tt.wantContain) {
				t.Errorf("URL should contain %q, got: %s", tt.wantContain, url)
			}
		})
	}
}

func TestClientOption_WithTimeout(t *testing.T) {
	client := NewClient(WithTimeout(5 * time.Second))
	if client.HTTPClient.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", client.HTTPClient.Timeout)
	}
}

func TestClientOption_WithRetryCount(t *testing.T) {
	client := NewClient(WithRetryCount(5))
	if client.RetryCount != 5 {
		t.Errorf("Expected retry count 5, got %d", client.RetryCount)
	}
}

func TestClientOption_WithDebug(t *testing.T) {
	client := NewClient(WithDebug(true))
	if !client.Debug {
		t.Error("Expected debug to be true")
	}
}

func TestClientOption_WithUserAgent(t *testing.T) {
	client := NewClient(WithUserAgent("test-agent/1.0"))
	if client.UserAgent != "test-agent/1.0" {
		t.Errorf("Expected user agent 'test-agent/1.0', got %q", client.UserAgent)
	}
}

func TestClientOption_WithBaseURL(t *testing.T) {
	client := NewClient(WithBaseURL("http://localhost:8080/"))
	if client.BaseURL != "http://localhost:8080/" {
		t.Errorf("Expected base URL 'http://localhost:8080/', got %q", client.BaseURL)
	}
}

func TestClientOption_Defaults(t *testing.T) {
	client := NewClient()
	if client.BaseURL != "https://crt.sh/" {
		t.Errorf("Expected default base URL, got %q", client.BaseURL)
	}
	if client.RetryCount != 3 {
		t.Errorf("Expected default retry count 3, got %d", client.RetryCount)
	}
	if client.Debug {
		t.Error("Expected default debug to be false")
	}
}

func TestSearchTypes(t *testing.T) {
	types := SearchTypes()
	if len(types) != 22 {
		t.Errorf("Expected 22 search types, got %d", len(types))
	}
	// Verify first entry is the default empty type
	if types[0].Type != "" {
		t.Errorf("Expected first type to be empty string, got %q", types[0].Type)
	}
	// Verify each type has a description
	for _, st := range types {
		if st.Description == "" {
			t.Errorf("Search type %q has empty description", st.Type)
		}
	}
}

func TestMatchModes(t *testing.T) {
	modes := MatchModes()
	if len(modes) != 7 {
		t.Errorf("Expected 7 match modes, got %d", len(modes))
	}
}

func TestLinters(t *testing.T) {
	linters := Linters()
	if len(linters) != 5 {
		t.Errorf("Expected 5 linters, got %d", len(linters))
	}
}

func TestLintTypes(t *testing.T) {
	types := LintTypes()
	if len(types) != 2 {
		t.Errorf("Expected 2 lint types, got %d", len(types))
	}
}

func TestValidSearchTypes(t *testing.T) {
	valid := ValidSearchTypes()
	if !valid["CN"] {
		t.Error("Expected CN to be a valid search type")
	}
	if !valid[""] {
		t.Error("Expected empty string to be a valid search type")
	}
	if valid["INVALID"] {
		t.Error("Expected INVALID to not be a valid search type")
	}
}

func TestIterateCertificates(t *testing.T) {
	page := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		certs := []Certificate{{ID: page, SerialNumber: "TEST"}}
		w.Header().Set("Link", fmt.Sprintf("<%s?page=%d>; rel=\"next\"", r.URL.Path, page+1))
		if page >= 3 {
			w.Header().Del("Link")
		}
		json.NewEncoder(w).Encode(certs)
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"))

	var collected [][]Certificate
	err := client.IterateCertificates(context.Background(), QueryParams{
		Q: "test",
	}, func(certs []Certificate, pagination *Pagination) bool {
		collected = append(collected, certs)
		return len(collected) < 3
	})
	if err != nil {
		t.Fatalf("IterateCertificates failed: %v", err)
	}
	if len(collected) != 3 {
		t.Errorf("Expected 3 pages, got %d", len(collected))
	}
}

func TestIterateCertificates_EarlyStop(t *testing.T) {
	callCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		certs := []Certificate{{ID: callCount, SerialNumber: "TEST"}}
		w.Header().Set("Link", fmt.Sprintf("<%s?page=%d>; rel=\"next\"", r.URL.Path, callCount+1))
		json.NewEncoder(w).Encode(certs)
	}))
	defer testServer.Close()

	client := NewClient(WithBaseURL(testServer.URL + "/"))

	var collected [][]Certificate
	err := client.IterateCertificates(context.Background(), QueryParams{
		Q: "test",
	}, func(certs []Certificate, pagination *Pagination) bool {
		collected = append(collected, certs)
		return false // stop immediately after first page
	})
	if err != nil {
		t.Fatalf("IterateCertificates failed: %v", err)
	}
	if len(collected) != 1 {
		t.Errorf("Expected 1 page (early stop), got %d", len(collected))
	}
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}
}

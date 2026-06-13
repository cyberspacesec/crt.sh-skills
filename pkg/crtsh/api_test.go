// api_test.go
package crtsh

import (
	"context"
	"encoding/json"
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

	client := NewClient()
	client.BaseURL = testServer.URL + "/"

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

			client := NewClient()
			client.BaseURL = testServer.URL + "/"
			client.RetryCount = 0

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

	client := NewClient()
	client.BaseURL = testServer.URL + "/"

	_, _, err := client.SearchCertificates(context.Background(), QueryParams{})
	if err == nil || err.Error() != "api error (500): server error" {
		t.Errorf("Expected server error, got: %v", err)
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

	client := NewClient()
	client.BaseURL = testServer.URL + "/"
	client.RetryCount = 3

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
		// GetCertificateByID now uses SearchCertificates with search_type=id
		// URL should be ?id=123&output=json (NOT ?searchtype=id&id=123)
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

	client := NewClient()
	client.BaseURL = testServer.URL + "/"

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
		// Return empty array for non-existent certificate
		json.NewEncoder(w).Encode([]Certificate{})
	}))
	defer testServer.Close()

	client := NewClient()
	client.BaseURL = testServer.URL + "/"

	_, err := client.GetCertificateByID(context.Background(), 999999999)
	if err == nil {
		t.Error("Expected error for non-existent certificate")
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

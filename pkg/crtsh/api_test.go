// api_test.go
package crtsh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_SearchCertificates(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("searchtype") != "CN" || r.URL.Query().Get("common_name") != "example.com" {
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
				"searchtype": "c",
				"c":          "ABCDEF123456",
			},
		},
		{
			name: "search type ca",
			params: QueryParams{
				SearchType: "ca",
				Q:          "Let's Encrypt",
			},
			expectedParams: map[string]string{
				"searchtype": "ca",
				"ca":         "Let's Encrypt",
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
		if r.URL.Query().Get("searchtype") != "id" || r.URL.Query().Get("id") != "123" {
			t.Errorf("Unexpected request: %v", r.URL)
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

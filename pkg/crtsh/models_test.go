package crtsh

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCertificate_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    Certificate
		expectError bool
	}{
		{
			name: "valid certificate with all new fields",
			input: `{
					"id": 1,
					"issuer_ca_id": 123,
					"issuer_name": "C=US, O=Test CA, CN=Test CA",
					"common_name": "example.com",
					"name_value": "example.com\n*.example.com",
					"entry_timestamp": "2025-02-26T12:34:56.789",
					"not_before": "2025-01-01",
					"not_after": "2026-01-01",
					"serial_number": "ABC123",
					"result_count": 5
				}`,
			expected: Certificate{
				ID:             1,
				IssuerCAID:     123,
				IssuerName:     "C=US, O=Test CA, CN=Test CA",
				CommonName:     "example.com",
				RawNameValue:   "example.com\n*.example.com",
				NameValue:      []string{"example.com", "*.example.com"},
				Domains:        []string{"example.com"},
				EntryTimestamp: time.Date(2025, 2, 26, 12, 34, 56, 789000000, time.UTC),
				NotBefore:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				NotAfter:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				SerialNumber:   "ABC123",
				ResultCount:    5,
			},
		},
		{
			name: "valid timestamp without timezone",
			input: `{
					"id": 1,
					"issuer_ca_id": 123,
					"name_value": "example.com\n*.example.com",
					"entry_timestamp": "2025-02-26T12:34:56.789",
					"not_before": "2025-01-01",
					"not_after": "2026-01-01",
					"serial_number": "ABC123"
				}`,
			expected: Certificate{
				ID:             1,
				IssuerCAID:     123,
				RawNameValue:   "example.com\n*.example.com",
				NameValue:      []string{"example.com", "*.example.com"},
				Domains:        []string{"example.com"},
				EntryTimestamp: time.Date(2025, 2, 26, 12, 34, 56, 789000000, time.UTC),
				NotBefore:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				NotAfter:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				SerialNumber:   "ABC123",
			},
		},
		{
			name: "real crt.sh response format",
			input: `{
					"issuer_ca_id": 413868,
					"issuer_name": "C=US, O=SSL Corporation, CN=Cloudflare TLS Issuing ECC CA 3",
					"common_name": "example.com",
					"name_value": "*.example.com\nexample.com",
					"id": 26786991824,
					"entry_timestamp": "2026-05-31T21:49:13.481",
					"not_before": "2026-05-31T21:39:12",
					"not_after": "2026-08-29T21:41:26",
					"serial_number": "1aa73fea257be3334b9a29552e6f878e",
					"result_count": 3
				}`,
			expected: Certificate{
				ID:             26786991824,
				IssuerCAID:     413868,
				IssuerName:     "C=US, O=SSL Corporation, CN=Cloudflare TLS Issuing ECC CA 3",
				CommonName:     "example.com",
				RawNameValue:   "*.example.com\nexample.com",
				NameValue:      []string{"*.example.com", "example.com"},
				Domains:        []string{"example.com"},
				EntryTimestamp: time.Date(2026, 5, 31, 21, 49, 13, 481000000, time.UTC),
				NotBefore:      time.Date(2026, 5, 31, 21, 39, 12, 0, time.UTC),
				NotAfter:       time.Date(2026, 8, 29, 21, 41, 26, 0, time.UTC),
				SerialNumber:   "1aa73fea257be3334b9a29552e6f878e",
				ResultCount:    3,
			},
		},
		{
			name: "invalid time format",
			input: `{
					"entry_timestamp": "not-a-date",
					"not_before": "not-a-date",
					"not_after": "not-a-date"
				}`,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cert Certificate
			err := json.Unmarshal([]byte(tc.input), &cert)
			if tc.expectError && err == nil {
				t.Error("Expected error for invalid time format")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tc.expectError {
				if !cert.EntryTimestamp.Equal(tc.expected.EntryTimestamp) {
					t.Errorf("EntryTimestamp mismatch: got %v, want %v",
						cert.EntryTimestamp, tc.expected.EntryTimestamp)
				}
				if len(cert.Domains) != len(tc.expected.Domains) {
					t.Errorf("Domains count mismatch: got %d, want %d",
						len(cert.Domains), len(tc.expected.Domains))
				}
				if cert.IssuerName != tc.expected.IssuerName {
					t.Errorf("IssuerName mismatch: got %q, want %q",
						cert.IssuerName, tc.expected.IssuerName)
				}
				if cert.CommonName != tc.expected.CommonName {
					t.Errorf("CommonName mismatch: got %q, want %q",
						cert.CommonName, tc.expected.CommonName)
				}
				if cert.ResultCount != tc.expected.ResultCount {
					t.Errorf("ResultCount mismatch: got %d, want %d",
						cert.ResultCount, tc.expected.ResultCount)
				}
			}
		})
	}
}

func TestQueryParamsValidation(t *testing.T) {
	params := QueryParams{
		SearchType:     "CN",
		CN:             "example.com",
		ExcludeExpired: true,
		Linter:         "zlint",
		LintType:       "issues",
		Page:           2,
		PageSize:       50,
	}

	if params.SearchType != "CN" || params.CN != "example.com" {
		t.Error("QueryParams field validation failed")
	}
	if params.Linter != "zlint" {
		t.Error("QueryParams Linter field validation failed")
	}
	if params.LintType != "issues" {
		t.Error("QueryParams LintType field validation failed")
	}
}

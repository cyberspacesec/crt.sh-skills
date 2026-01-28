// models_test.go
package crtsh

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCertificate_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Certificate
	}{
		{
			name: "valid timestamp format",
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
				Domains:        []string{"example.com", "example.com"},
				EntryTimestamp: time.Date(2025, 2, 26, 12, 34, 56, 789000000, time.UTC),
				NotBefore:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				NotAfter:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				SerialNumber:   "ABC123",
			},
		},
		{
			name: "invalid time format",
			input: `{
				"entry_timestamp": "invalid",
				"not_before": "invalid",
				"not_after": "invalid"
			}`,
			expected: Certificate{
				EntryTimestamp: time.Time{},
				NotBefore:      time.Time{},
				NotAfter:       time.Time{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cert Certificate
			err := json.Unmarshal([]byte(tc.input), &cert)
			if tc.expected.EntryTimestamp.IsZero() && err == nil {
				t.Error("Expected error for invalid time format")
			}

			if !cert.EntryTimestamp.Equal(tc.expected.EntryTimestamp) {
				t.Errorf("EntryTimestamp mismatch: got %v, want %v",
					cert.EntryTimestamp, tc.expected.EntryTimestamp)
			}

			if len(cert.Domains) != len(tc.expected.Domains) {
				t.Errorf("Domains count mismatch: got %d, want %d",
					len(cert.Domains), len(tc.expected.Domains))
			}
		})
	}
}

func TestQueryParamsValidation(t *testing.T) {
	params := QueryParams{
		SearchType:     "CN",
		CN:             "example.com",
		ExcludeExpired: true,
		Page:           2,
		PageSize:       50,
	}

	if params.SearchType != "CN" || params.CN != "example.com" {
		t.Error("QueryParams field validation failed")
	}
}

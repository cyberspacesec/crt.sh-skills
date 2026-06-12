package crtsh

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var timeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func parseTime(s string) (time.Time, error) {
	for _, f := range timeFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %q", s)
}

type Certificate struct {
	ID             int       `json:"id"`
	IssuerCAID     int       `json:"issuer_ca_id"`
	IssuerName     string    `json:"issuer_name"`
	CommonName     string    `json:"common_name"`
	NameValue      []string  `json:"-"`
	RawNameValue   string    `json:"name_value"`
	EntryTimestamp time.Time `json:"entry_timestamp"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	SerialNumber   string    `json:"serial_number"`
	ResultCount    int       `json:"result_count"`
	Domains        []string  `json:"-"`
}

type QueryParams struct {
	SearchType     string
	Q              string
	ID             string
	CTID           string
	Serial         string
	SKI            string
	SPKISHA1       string
	SPKISHA256     string
	SubjectSHA1    string
	SHA1           string
	SHA256         string
	CAID           string
	CAName         string
	Identity       string
	CN             string
	E              string
	OU             string
	O              string
	DNSName        string
	RFC822Name     string
	IPAddress      string
	Match          string
	ExcludeExpired bool
	Deduplicate    bool
	ShowSQL        bool
	SearchCensys   bool
	Linter         string
	LintType       string
	Page           int
	PageSize       int
}

type Pagination struct {
	Total       int
	CurrentPage int
	PageSize    int
	NextPage    int
}

func (c *Certificate) UnmarshalJSON(data []byte) error {
	type Alias Certificate
	aux := &struct {
		EntryTimestamp string `json:"entry_timestamp"`
		NotBefore      string `json:"not_before"`
		NotAfter       string `json:"not_after"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	c.EntryTimestamp, err = parseTime(aux.EntryTimestamp)
	if err != nil {
		return fmt.Errorf("parse entry_timestamp: %w", err)
	}
	c.NotBefore, err = parseTime(aux.NotBefore)
	if err != nil {
		return fmt.Errorf("parse not_before: %w", err)
	}
	c.NotAfter, err = parseTime(aux.NotAfter)
	if err != nil {
		return fmt.Errorf("parse not_after: %w", err)
	}

	c.NameValue = splitDomains(c.RawNameValue)
	c.Domains = uniqueDomains(c.NameValue)

	return nil
}

func splitDomains(input string) []string {
	return strings.FieldsFunc(input, func(r rune) bool {
		return r == '\n' || r == ',' || r == ' '
	})
}

func uniqueDomains(domains []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(domains))

	for _, d := range domains {
		d = strings.TrimSpace(d)
		d = strings.TrimPrefix(d, "*.")
		if d == "" {
			continue
		}
		if _, exists := seen[d]; !exists {
			seen[d] = struct{}{}
			result = append(result, d)
		}
	}
	return result
}

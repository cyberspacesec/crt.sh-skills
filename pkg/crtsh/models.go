package crtsh

import (
	"encoding/json"
	"strings"
	"time"
)

type Certificate struct {
	ID             int       `json:"id"`
	IssuerCAID     int       `json:"issuer_ca_id"`
	NameValue      []string  `json:"-"`
	RawNameValue   string    `json:"name_value"`
	EntryTimestamp time.Time `json:"entry_timestamp"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	SerialNumber   string    `json:"serial_number"`
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

	c.EntryTimestamp, _ = time.Parse(time.RFC3339, aux.EntryTimestamp)
	c.NotBefore, _ = time.Parse("2006-01-02", aux.NotBefore)
	c.NotAfter, _ = time.Parse("2006-01-02", aux.NotAfter)

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

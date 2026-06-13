// registry.go — Single source of truth for crt.sh search types, match modes, linters, and lint types
package crtsh

// SearchTypeDescriptor describes a crt.sh search type.
type SearchTypeDescriptor struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// MatchModeDescriptor describes a crt.sh identity matching mode.
type MatchModeDescriptor struct {
	Mode        string `json:"mode"`
	SQL         string `json:"sql_equivalent"`
	Description string `json:"description"`
}

// LinterDescriptor describes a crt.sh certificate linter.
type LinterDescriptor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LintTypeDescriptor describes a crt.sh lint output type.
type LintTypeDescriptor struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SearchTypes returns all valid crt.sh search types with descriptions.
// The empty string represents the default general search.
func SearchTypes() []SearchTypeDescriptor {
	return []SearchTypeDescriptor{
		{Type: "", Description: "General search (domain name)"},
		{Type: "c", Description: "Certificate fingerprint (SHA-1 or SHA-256)"},
		{Type: "id", Description: "crt.sh certificate ID"},
		{Type: "ctid", Description: "CT Entry ID"},
		{Type: "serial", Description: "Serial number"},
		{Type: "ski", Description: "Subject Key Identifier"},
		{Type: "spkisha1", Description: "SHA-1(SubjectPublicKeyInfo)"},
		{Type: "spkisha256", Description: "SHA-256(SubjectPublicKeyInfo)"},
		{Type: "subjectsha1", Description: "SHA-1(Subject)"},
		{Type: "sha1", Description: "SHA-1(Certificate)"},
		{Type: "sha256", Description: "SHA-256(Certificate)"},
		{Type: "ca", Description: "CA (general)"},
		{Type: "CAID", Description: "CA ID"},
		{Type: "CAName", Description: "CA Name"},
		{Type: "Identity", Description: "Identity"},
		{Type: "CN", Description: "commonName (Subject)"},
		{Type: "E", Description: "emailAddress (Subject)"},
		{Type: "OU", Description: "organizationalUnitName (Subject)"},
		{Type: "O", Description: "organizationName (Subject)"},
		{Type: "dNSName", Description: "dNSName (SAN)"},
		{Type: "rfc822Name", Description: "rfc822Name (SAN)"},
		{Type: "iPAddress", Description: "iPAddress (SAN)"},
	}
}

// MatchModes returns all valid crt.sh identity matching modes.
func MatchModes() []MatchModeDescriptor {
	return []MatchModeDescriptor{
		{Mode: "", SQL: "Auto", Description: "Let crt.sh pick the best mode"},
		{Mode: "=", SQL: "=", Description: "Exact identity match"},
		{Mode: "ILIKE", SQL: "ILIKE", Description: "Case-insensitive pattern match"},
		{Mode: "LIKE", SQL: "LIKE", Description: "Case-sensitive pattern match"},
		{Mode: "single", SQL: "—", Description: "Match single identity value"},
		{Mode: "any", SQL: "—", Description: "Match any identity value"},
		{Mode: "FTS", SQL: "Full Text Search", Description: "Full text search across all fields"},
	}
}

// Linters returns all valid crt.sh certificate linters.
func Linters() []LinterDescriptor {
	return []LinterDescriptor{
		{Name: "cablint", Description: "CAB Forum linting (cablint)"},
		{Name: "x509lint", Description: "X.509 structure linting (x509lint)"},
		{Name: "zlint", Description: "Zlint — comprehensive certificate linter"},
		{Name: "keylint", Description: "Key linting (keylint)"},
		{Name: "lint", Description: "Run all linters"},
	}
}

// LintTypes returns all valid crt.sh lint output types.
func LintTypes() []LintTypeDescriptor {
	return []LintTypeDescriptor{
		{Type: "1 week", Description: "1-week summary of linting results"},
		{Type: "issues", Description: "Show only issues found by the linter"},
	}
}

// ValidSearchTypes returns a map of all valid search type strings for quick lookup.
func ValidSearchTypes() map[string]bool {
	types := make(map[string]bool, len(searchTypeList))
	for _, t := range searchTypeList {
		types[t] = true
	}
	return types
}

// searchTypeList is the internal list used by ValidSearchTypes.
var searchTypeList = []string{
	"", "c", "id", "ctid", "serial", "ski", "spkisha1", "spkisha256",
	"subjectsha1", "sha1", "sha256", "ca", "CAID", "CAName",
	"Identity", "CN", "E", "OU", "O", "dNSName", "rfc822Name", "iPAddress",
}

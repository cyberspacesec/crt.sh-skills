// crtsh-cli — Command-line interface for crt.sh Certificate Transparency search
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	crtsh "github.com/cyberspacesec/crt.sh-skills/pkg/crtsh"
	"github.com/spf13/cobra"
)

var (
	searchType     string
	match          string
	excludeExpired bool
	deduplicate    bool
	showSQL        bool
	linter         string
	lintType       string
	page           int
	pageSize       int
	outputJSON     bool

	// Root-level persistent flags
	timeout   time.Duration
	debug     bool
	outputFmt string
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "crtsh-cli",
		Short: "CLI for crt.sh Certificate Transparency search engine",
		Long:  "Command-line interface wrapping the crt.sh-skills SDK. Search CT logs, retrieve certificates, and access crt.sh info pages.",
	}

	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format: json, table, csv")

	rootCmd.AddCommand(
		searchCmd(),
		getCertCmd(),
		getInfoPageCmd(),
		getCACmd(),
		searchCensysCmd(),
		listPagesCmd(),
		listSearchTypesCmd(),
		listLintersCmd(),
		listMatchModesCmd(),
	)

	rootCmd.Version = Version

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// newClientFromFlags creates a crt.sh client configured from root-level flags.
func newClientFromFlags() *crtsh.Client {
	return crtsh.NewClient(
		crtsh.WithTimeout(timeout),
		crtsh.WithDebug(debug),
	)
}

// resolveOutputFormat determines the effective output format from --output and --json flags.
// The --json/-j flag is a shortcut for --output json, for backward compatibility.
func resolveOutputFormat() string {
	if outputJSON {
		return "json"
	}
	return outputFmt
}

func searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search certificate transparency logs",
		Long:  "Search crt.sh for certificates by domain, hash, serial number, CA name, and more.\nUse --type to specify the search type (default: general domain search).\nUse 'crtsh-cli list-types' to see all available search types.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearch,
	}

	cmd.Flags().StringVarP(&searchType, "type", "t", "", "Search type (use 'crtsh-cli list-types' to see all)")
	cmd.Flags().StringVarP(&match, "match", "m", "", "Match mode (use 'crtsh-cli list-match-modes' to see all)")
	cmd.Flags().BoolVarP(&excludeExpired, "exclude-expired", "e", false, "Exclude expired certificates")
	cmd.Flags().BoolVarP(&deduplicate, "deduplicate", "d", false, "Deduplicate precertificate pairs")
	cmd.Flags().BoolVar(&showSQL, "show-sql", false, "Show SQL query used by crt.sh")
	cmd.Flags().StringVar(&linter, "linter", "", "Run linter (use 'crtsh-cli list-linters' to see all)")
	cmd.Flags().StringVar(&lintType, "lint-type", "", "Lint output: '1 week' or 'issues'")
	cmd.Flags().IntVarP(&page, "page", "p", 0, "Page number (1-based)")
	cmd.Flags().IntVarP(&pageSize, "page-size", "s", 0, "Results per page")
	cmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON (shorthand for --output json)")

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	client := newClientFromFlags()
	query := args[0]

	params := crtsh.QueryParams{
		SearchType:     searchType,
		Match:          match,
		ExcludeExpired: excludeExpired,
		Deduplicate:    deduplicate,
		ShowSQL:        showSQL,
		Linter:         linter,
		LintType:       lintType,
		Page:           page,
		PageSize:       pageSize,
	}

	// Route query to correct field based on search type
	switch searchType {
	case "c":
		params.Q = query
	case "id":
		params.ID = query
	case "ctid":
		params.CTID = query
	case "serial":
		params.Serial = query
	case "ski":
		params.SKI = query
	case "spkisha1":
		params.SPKISHA1 = query
	case "spkisha256":
		params.SPKISHA256 = query
	case "subjectsha1":
		params.SubjectSHA1 = query
	case "sha1":
		params.SHA1 = query
	case "sha256":
		params.SHA256 = query
	case "ca":
		params.Q = query
	case "CAID":
		params.CAID = query
	case "CAName":
		params.CAName = query
	case "Identity":
		params.Identity = query
	case "CN":
		params.CN = query
	case "E":
		params.E = query
	case "OU":
		params.OU = query
	case "O":
		params.O = query
	case "dNSName":
		params.DNSName = query
	case "rfc822Name":
		params.RFC822Name = query
	case "iPAddress":
		params.IPAddress = query
	default:
		params.Q = query
	}

	certs, pagination, err := client.SearchCertificates(context.Background(), params)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	switch resolveOutputFormat() {
	case "json":
		result := map[string]interface{}{"certificates": certs}
		if pagination != nil {
			result["pagination"] = pagination
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	case "csv":
		cw := csv.NewWriter(os.Stdout)
		_ = cw.Write([]string{"id", "issuer_ca_id", "common_name", "not_before", "not_after", "domains", "serial_number"})
		for _, cert := range certs {
			domains := strings.Join(cert.Domains, "; ")
			_ = cw.Write([]string{
				strconv.Itoa(cert.ID),
				strconv.Itoa(cert.IssuerCAID),
				cert.CommonName,
				cert.NotBefore.Format("2006-01-02"),
				cert.NotAfter.Format("2006-01-02"),
				domains,
				cert.SerialNumber,
			})
		}
		cw.Flush()
	default: // table
		if len(certs) == 0 {
			fmt.Println("No certificates found.")
			return nil
		}

		fmt.Printf("Found %d certificates\n", len(certs))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tISSUER CA ID\tCOMMON NAME\tNOT BEFORE\tNOT AFTER\tDOMAINS")
		fmt.Fprintln(w, "--\t------------\t-----------\t----------\t---------\t-------")
		for _, cert := range certs {
			domains := strings.Join(cert.Domains, ", ")
			if len(domains) > 60 {
				domains = domains[:57] + "..."
			}
			fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\t%s\n",
				cert.ID, cert.IssuerCAID, cert.CommonName,
				cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"),
				domains)
		}
		w.Flush()

		if pagination != nil && pagination.NextPage > 0 {
			fmt.Printf("\nMore results available (next page: %d)\n", pagination.NextPage)
		}
	}

	return nil
}

func getCertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-cert [id]",
		Short: "Get a certificate by its crt.sh ID",
		Long:  "Retrieve detailed information about a specific certificate from crt.sh using its numeric ID.\nUse 'crtsh-cli search' to find certificate IDs first.",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetCert,
	}
	cmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON (shorthand for --output json)")
	return cmd
}

func runGetCert(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid certificate ID: %s", args[0])
	}

	client := newClientFromFlags()
	cert, err := client.GetCertificateByID(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}

	switch resolveOutputFormat() {
	case "json":
		data, _ := json.MarshalIndent(cert, "", "  ")
		fmt.Println(string(data))
	default: // table or csv
		fmt.Printf("Certificate #%d\n", cert.ID)
		fmt.Printf("  Common Name:    %s\n", cert.CommonName)
		fmt.Printf("  Issuer:         %s (CA ID: %d)\n", cert.IssuerName, cert.IssuerCAID)
		fmt.Printf("  Serial Number:  %s\n", cert.SerialNumber)
		fmt.Printf("  Not Before:     %s\n", cert.NotBefore.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Not After:      %s\n", cert.NotAfter.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Entry Timestamp: %s\n", cert.EntryTimestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Result Count:   %d\n", cert.ResultCount)
		fmt.Printf("  Domains:\n")
		for _, d := range cert.Domains {
			fmt.Printf("    - %s\n", d)
		}
	}

	return nil
}

func getInfoPageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info-page [page-name]",
		Short: "Get a crt.sh information page",
		Long:  "Retrieve crt.sh info pages such as CT log status, CA disclosures, and revoked intermediates.\nUse 'crtsh-cli list-pages' to see available pages.",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetInfoPage,
	}
	cmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON (shorthand for --output json)")
	return cmd
}

func runGetInfoPage(cmd *cobra.Command, args []string) error {
	client := newClientFromFlags()
	info, err := client.FetchInfoPage(context.Background(), args[0])
	if err != nil {
		return fmt.Errorf("failed to fetch info page: %w", err)
	}

	switch resolveOutputFormat() {
	case "json":
		data, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(data))
	default:
		fmt.Printf("Page: %s\nTitle: %s\nDescription: %s\n\n%s\n",
			info.Path, info.Title, info.Description, info.Content)
	}
	return nil
}

func getCACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-ca [ca-id]",
		Short: "Get CA certificate details by CA ID",
		Long:  "Retrieve CA (Certificate Authority) certificate details from crt.sh by CA ID.\nThe CA ID can be found in the issuer_ca_id field of search results.",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetCA,
	}
	cmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON (shorthand for --output json)")
	return cmd
}

func runGetCA(cmd *cobra.Command, args []string) error {
	caID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid CA ID: %s", args[0])
	}

	client := newClientFromFlags()
	info, err := client.FetchCAByID(context.Background(), caID)
	if err != nil {
		return fmt.Errorf("failed to fetch CA info: %w", err)
	}

	switch resolveOutputFormat() {
	case "json":
		data, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(data))
	default:
		fmt.Printf("CA Certificate #%d\n\n%s\n", caID, info.Content)
	}
	return nil
}

func searchCensysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "censys [query]",
		Short: "Build a Censys.io search URL for certificate search",
		Long:  "Builds a Censys.io URL equivalent to crt.sh's searchCensys feature.\nNot all search types are supported by Censys. Use 'crtsh-cli list-types' to see all types.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearchCensys,
	}
	cmd.Flags().StringVarP(&searchType, "type", "t", "CN", "Search type for Censys")
	return cmd
}

func runSearchCensys(cmd *cobra.Command, args []string) error {
	url, err := crtsh.BuildCensysURL(searchType, args[0])
	if err != nil {
		return fmt.Errorf("censys search error: %w", err)
	}
	fmt.Println(url)
	return nil
}

func listPagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-pages",
		Short: "List available crt.sh info pages",
		Long:  "Display all available crt.sh information pages with their titles and descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PAGE\tTITLE\tDESCRIPTION")
			fmt.Fprintln(w, "----\t-----\t-----------")
			for path, info := range crtsh.InfoPages {
				fmt.Fprintf(w, "%s\t%s\t%s\n", path, info.Title, info.Description)
			}
			w.Flush()
			return nil
		},
	}
}

func listSearchTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-types",
		Short: "List available search types",
		Long:  "Display all 22 crt.sh search types with their descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			types := crtsh.SearchTypes()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TYPE\tDESCRIPTION")
			fmt.Fprintln(w, "----\t-----------")
			for _, t := range types {
				display := t.Type
				if display == "" {
					display = "(empty)"
				}
				fmt.Fprintf(w, "%s\t%s\n", display, t.Description)
			}
			w.Flush()
			return nil
		},
	}
}

func listLintersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-linters",
		Short: "List available certificate linters",
		Long:  "Display all available crt.sh certificate linters with their descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			linters := crtsh.Linters()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "LINTER\tDESCRIPTION")
			fmt.Fprintln(w, "------\t-----------")
			for _, l := range linters {
				fmt.Fprintf(w, "%s\t%s\n", l.Name, l.Description)
			}
			w.Flush()
			return nil
		},
	}
}

func listMatchModesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-match-modes",
		Short: "List available match modes",
		Long:  "Display all crt.sh identity matching modes with their SQL equivalents and descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			modes := crtsh.MatchModes()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "MODE\tSQL\tDESCRIPTION")
			fmt.Fprintln(w, "----\t---\t-----------")
			for _, m := range modes {
				display := m.Mode
				if display == "" {
					display = "(auto)"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", display, m.SQL, m.Description)
			}
			w.Flush()
			return nil
		},
	}
}

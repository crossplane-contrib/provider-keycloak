// Command configaudit reports reference wiring in the provider configuration
// that is inconsistent, missing or incomplete.
//
// It cross-checks the Terraform schema against the configuration the generator
// builds, so it needs neither a Keycloak instance nor network access.
//
// Usage:
//
//	configaudit                                       # human-readable report
//	configaudit --format=json                         # machine-readable report
//	configaudit --show-all                            # include satisfied findings
//	configaudit --fail-on=drift,missing-multitype     # exit 1 on actionable findings
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/crossplane-contrib/provider-keycloak/config"
	"github.com/crossplane-contrib/provider-keycloak/internal/configaudit"
)

var detectors = []string{
	configaudit.DetectorDrift,
	configaudit.DetectorMissingMultitype,
	configaudit.DetectorUnclassified,
}

func main() {
	format := flag.String("format", "table", "output format: table or json")
	showAll := flag.Bool("show-all", false, "include satisfied and protocol-specific findings in the table output")
	failOn := flag.String("fail-on", "", "comma-separated detectors that make the command exit non-zero when they report an actionable finding: "+strings.Join(detectors, ", "))
	flag.Parse()

	if err := run(*format, *showAll, *failOn); err != nil {
		fmt.Fprintf(os.Stderr, "config-audit: %s\n", err)
		os.Exit(2)
	}
}

func run(format string, showAll bool, failOn string) error {
	gates, err := parseFailOn(failOn)
	if err != nil {
		return err
	}

	p, err := config.GetProvider(true)
	if err != nil {
		return fmt.Errorf("cannot build the provider configuration: %w", err)
	}
	report := configaudit.Audit(p)

	switch format {
	case "json":
		err = configaudit.WriteJSON(os.Stdout, report)
	case "table":
		err = configaudit.WriteTable(os.Stdout, report, showAll)
	default:
		return fmt.Errorf("unknown format %q, want table or json", format)
	}
	if err != nil {
		return err
	}

	failed := false
	for _, detector := range gates {
		findings := report.Actionable(detector)
		if len(findings) == 0 {
			continue
		}
		failed = true
		fmt.Fprintf(os.Stderr, "config-audit: %d actionable %s finding(s)\n", len(findings), detector)
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", f.Key(), f.Detail)
		}
	}
	if failed {
		os.Exit(1)
	}
	return nil
}

func parseFailOn(failOn string) ([]string, error) {
	if strings.TrimSpace(failOn) == "" {
		return nil, nil
	}
	var out []string
	for _, name := range strings.Split(failOn, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !contains(detectors, name) {
			return nil, fmt.Errorf("unknown detector %q in --fail-on, want one of: %s", name, strings.Join(detectors, ", "))
		}
		out = append(out, name)
	}
	return out, nil
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

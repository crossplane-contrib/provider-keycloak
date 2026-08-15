package configaudit

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// WriteJSON writes the report as JSON, for consumption by automation such as
// scripts/schema_diff_issues.py.
func WriteJSON(w io.Writer, r Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// WriteTable writes a human-readable report. Findings that are satisfied or
// protocol-specific are summarised rather than listed, unless showAll is set.
func WriteTable(w io.Writer, r Report, showAll bool) error {
	if _, err := fmt.Fprintf(w, "config-audit: %d resources\n", r.Resources); err != nil {
		return err
	}

	for _, detector := range []string{DetectorDrift, DetectorMissingMultitype, DetectorUnclassified} {
		var shown, hidden []Finding
		for _, f := range r.Findings {
			if f.Detector != detector {
				continue
			}
			if !showAll && (f.Status == StatusSatisfied || f.ProtocolSpecific) {
				hidden = append(hidden, f)
				continue
			}
			shown = append(shown, f)
		}
		if err := writeSection(w, detector, shown, hidden); err != nil {
			return err
		}
	}
	return nil
}

func writeSection(w io.Writer, detector string, shown, hidden []Finding) error {
	if _, err := fmt.Fprintf(w, "\n== %s (%d) ==\n", detector, len(shown)); err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, f := range shown {
		note := ""
		switch {
		case f.Status == StatusSatisfied:
			note = " [satisfied]"
		case f.ProtocolSpecific:
			note = " [protocol-specific]"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s%s\n", f.Class, f.Key(), f.Detail, note); err != nil {
			return err
		}
		if err := writeDriftDetail(tw, f); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	if len(hidden) > 0 {
		if _, err := fmt.Fprintf(w, "  (%d not shown: satisfied or protocol-specific; use --show-all)\n", len(hidden)); err != nil {
			return err
		}
	}
	return nil
}

func writeDriftDetail(w io.Writer, f Finding) error {
	if f.Detector != DetectorDrift {
		return nil
	}
	for _, target := range sorted(keys(f.WiredTo)) {
		if _, err := fmt.Fprintf(w, "\t  wired to %s\t%s\n", target, strings.Join(f.WiredTo[target], ", ")); err != nil {
			return err
		}
	}
	if len(f.UnwiredOn) > 0 {
		if _, err := fmt.Fprintf(w, "\t  unwired on\t%s\n", strings.Join(f.UnwiredOn, ", ")); err != nil {
			return err
		}
	}
	return nil
}

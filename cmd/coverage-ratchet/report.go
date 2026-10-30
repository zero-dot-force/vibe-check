package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/zero-dot-force/vibe-check/internal/coverratchet"
)

// summaryStatusOrder fixes the order in which non-pass statuses appear in the
// summary line, keeping the table report deterministic.
var summaryStatusOrder = []coverratchet.Status{
	coverratchet.StatusRaised,
	coverratchet.StatusAdded,
	coverratchet.StatusUnchanged,
	coverratchet.StatusBelowFloor,
	coverratchet.StatusUnregistered,
	coverratchet.StatusStaleFloor,
	coverratchet.StatusRefused,
}

// writeReport writes the gate report to w: the table by default, or a stable
// JSON object with a provenance envelope when JSON output is requested. The
// table is byte-stable for identical inputs; the JSON object is byte-stable
// except for the provenance timestamp.
func writeReport(w io.Writer, opts RatchetOptions, result *RatchetResult) error {
	if opts.JSON {
		return writeJSONReport(w, result)
	}
	return writeTableReport(w, opts, result)
}

// writeTableReport writes one line per package (measured, floor, status),
// sorted by package path, followed by a summary line. Absent values render as
// "-": a stale floor has no measurement, and an unregistered or refused
// package has no floor. In update mode, raised and added rows show the floor
// the update writes, and a final line reports whether the baseline file was
// rewritten.
func writeTableReport(w io.Writer, opts RatchetOptions, result *RatchetResult) error {
	if _, err := fmt.Fprintf(w, "%-30s %9s %8s  %s\n", "package", "measured", "floor", "status"); err != nil {
		return err
	}
	for _, p := range result.Result.Packages {
		measured := "-"
		if p.Status != coverratchet.StatusStaleFloor {
			measured = fmt.Sprintf("%.1f%%", p.Measured)
		}
		floor := "-"
		if p.Status != coverratchet.StatusUnregistered && p.Status != coverratchet.StatusRefused {
			floor = fmt.Sprintf("%.1f%%", p.Floor)
		}
		if _, err := fmt.Fprintf(w, "%-30s %9s %8s  %s\n", p.Package, measured, floor, p.Status); err != nil {
			return err
		}
	}
	counts := make(map[coverratchet.Status]int)
	for _, p := range result.Result.Packages {
		counts[p.Status]++
	}
	summary := fmt.Sprintf("summary: %d packages, %d pass", len(result.Result.Packages), counts[coverratchet.StatusPass])
	for _, s := range summaryStatusOrder {
		if counts[s] > 0 {
			summary += fmt.Sprintf(", %d %s", counts[s], s)
		}
	}
	if _, err := fmt.Fprintln(w, summary); err != nil {
		return err
	}
	if opts.Update {
		if result.Updated {
			_, err := fmt.Fprintf(w, "baseline updated: %s\n", opts.Floors)
			return err
		}
		_, err := fmt.Fprintf(w, "baseline not modified: %s\n", opts.Floors)
		return err
	}
	return nil
}

// jsonPackage is one per-package entry in the JSON report. Measured is null
// for a stale floor (the package has no measurement) and Floor is null for an
// unregistered or refused package (no floor exists); the values are pointers
// so absent data marshals as JSON null rather than a misleading 0.
type jsonPackage struct {
	Package  string   `json:"package"`
	Measured *float64 `json:"measured"`
	Floor    *float64 `json:"floor"`
	Status   string   `json:"status"`
}

// jsonReport is the --json output object: per-package results, the overall
// verdict, whether the baseline was rewritten, and a provenance envelope.
// Excluding the provenance timestamp it is byte-stable for identical inputs.
type jsonReport struct {
	Packages   []jsonPackage       `json:"packages"`
	Verdict    string              `json:"verdict"`
	Updated    bool                `json:"updated"`
	Provenance *provenanceEnvelope `json:"provenance"`
}

// writeJSONReport writes the gate report as an indented JSON object. The
// verdict is "pass" when every package satisfies the policy and "fail"
// otherwise.
func writeJSONReport(w io.Writer, result *RatchetResult) error {
	verdict := "fail"
	if result.Result.OK() {
		verdict = "pass"
	}
	report := jsonReport{
		Packages:   make([]jsonPackage, 0, len(result.Result.Packages)),
		Verdict:    verdict,
		Updated:    result.Updated,
		Provenance: newProvenanceEnvelope(),
	}
	for _, p := range result.Result.Packages {
		entry := jsonPackage{Package: p.Package, Status: string(p.Status)}
		if p.Status != coverratchet.StatusStaleFloor {
			entry.Measured = floatPtr(p.Measured)
		}
		if p.Status != coverratchet.StatusUnregistered && p.Status != coverratchet.StatusRefused {
			entry.Floor = floatPtr(p.Floor)
		}
		report.Packages = append(report.Packages, entry)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if _, err := w.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

// floatPtr returns a pointer to v, for nullable JSON number fields.
func floatPtr(v float64) *float64 { return &v }

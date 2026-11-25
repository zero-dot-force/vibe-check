package ripwireadapter

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// archZoneUnknown is the zone value ripwire prints for a module holding a
// file in a language it draws almost no file edges for (its no_file_edges=
// set). Such a row's ca/ce are floors, and absent when no edge was
// observed (openspec/changes/archive/2026-11-25-bump-ripwire-pin design D2/D3).
const archZoneUnknown = "unknown"

// archModuleRow is one directory-grain row of ripwire's --arch metrics
// table, reduced to the attributes the adapter consumes. Instability,
// abstractness, distance, and zone are deliberately not carried: the
// adapter derives them with metrics.Compute* (design D3 of the
// ripwire-adapter change), so ripwire's own values never reach output.
//
// Path is the raw root-relative form ripwire prints ("fixturepy/cycle_a",
// or "./fixturepy/cycle_a" from older releases, "." for the crawl root);
// path normalization is the directory-module mapping's job, so the parser
// stays faithful to the wire format.
type archModuleRow struct {
	// Path is the row's root-relative module path, verbatim.
	Path string
	// Ca and Ce are the module's afferent and efferent coupling counts
	// (unitless, >= 0). When CouplingFloor is true they are lower
	// bounds, not exact counts.
	Ca, Ce int
	// Types and Abstract are the module's type-declaration count and its
	// abstract-type count (ripwire's interface share for TypeScript;
	// structurally zero for Python).
	Types, Abstract int
	// CouplingFloor reports that ripwire marked the row zone="unknown":
	// Ca and Ce are floors, read as 0 where ripwire omitted them.
	CouplingFloor bool
	// NoFileEdges lists the languages ripwire named in the row's
	// no_file_edges= attribute, sorted; empty unless CouplingFloor.
	NoFileEdges []string
}

// archTable is the parsed --arch output: the per-module rows plus
// ripwire's own disclosure of denylist-pruned packages.
type archTable struct {
	// Rows are the per-module rows, in document order.
	Rows []archModuleRow
	// PrunedPackages are the root-relative package paths ripwire lists
	// under <pruned_packages>, sorted. ripwire caps the listing, so this
	// may be shorter than PrunedPackageCount.
	PrunedPackages []string
	// PrunedPackageCount is ripwire's exact count= of pruned packages
	// (0 when the element is absent).
	PrunedPackageCount int
}

// archRowXML, archMetricsXML, archPrunedXML, and archRootXML are the wire
// shapes. Consumed numeric attributes are pointers so a missing attribute
// is distinguishable from a legitimate zero and fails closed instead of
// silently decoding as 0. Attributes the adapter does not consume (I, A, D,
// reachable, isolated, and every <metrics> aggregate) are not declared:
// encoding/xml ignores them, so their presence or absence never matters.
type archRowXML struct {
	Path        *string `xml:"path,attr"`
	Ca          *int    `xml:"ca,attr"`
	Ce          *int    `xml:"ce,attr"`
	Types       *int    `xml:"types,attr"`
	Abstract    *int    `xml:"abstract,attr"`
	Zone        string  `xml:"zone,attr"`
	NoFileEdges string  `xml:"no_file_edges,attr"`
}

type archMetricsXML struct {
	Rows []archRowXML `xml:"m"`
}

type archPrunedPathXML struct {
	Path *string `xml:"path,attr"`
}

type archPrunedXML struct {
	Count *int                `xml:"count,attr"`
	Paths []archPrunedPathXML `xml:"p"`
}

type archRootXML struct {
	XMLName xml.Name        `xml:"arch"`
	Metrics *archMetricsXML `xml:"metrics"`
	Pruned  *archPrunedXML  `xml:"pruned_packages"`
}

// parseArchOutput parses the --arch output of the pinned ripwire release.
// It reads ONLY the aggregated per-directory <metrics> rows — never
// per-file <inc> lists, which ripwire caps — so module coupling counts are
// always the uncapped table values.
//
// The contract is limited to consumed attributes (openspec/changes/
// bump-ripwire-pin design D2): every row MUST carry path, types, and
// abstract; ca and ce are required except on a zone="unknown" row, where
// an absent value is a floor of 0 and the row is flagged CouplingFloor.
// Unconsumed attributes and aggregates may be present or absent.
//
// A present-but-rowless <metrics> element parses to an empty table without
// error (an empty crawl is the pipeline's problem to classify, not the
// parser's); a missing <metrics> element, a wrong root element, malformed
// XML, a row missing a required attribute, duplicate module paths, or a
// malformed <pruned_packages> element are errors — deterministic pinned
// input must never produce silent zero-valued or last-wins rows.
func parseArchOutput(data []byte) (archTable, error) {
	var root archRootXML
	if err := xml.Unmarshal(data, &root); err != nil {
		return archTable{}, fmt.Errorf("parse arch output: %w", err)
	}
	if root.Metrics == nil {
		return archTable{}, fmt.Errorf("parse arch output: no <metrics> table in --arch output")
	}

	table := archTable{Rows: make([]archModuleRow, 0, len(root.Metrics.Rows))}
	seen := make(map[string]bool, len(root.Metrics.Rows))
	for _, row := range root.Metrics.Rows {
		parsed, err := row.toModuleRow()
		if err != nil {
			return archTable{}, err
		}
		if seen[parsed.Path] {
			return archTable{}, fmt.Errorf("parse arch output: duplicate module path %q", parsed.Path)
		}
		seen[parsed.Path] = true
		table.Rows = append(table.Rows, parsed)
	}

	if root.Pruned != nil {
		paths, count, err := root.Pruned.toPrunedPackages()
		if err != nil {
			return archTable{}, err
		}
		table.PrunedPackages = paths
		table.PrunedPackageCount = count
	}
	return table, nil
}

// toModuleRow converts one wire row, failing closed on missing consumed
// attributes.
func (r archRowXML) toModuleRow() (archModuleRow, error) {
	if r.Path == nil {
		return archModuleRow{}, fmt.Errorf("parse arch output: module row missing path attribute")
	}
	missing := func(attr string) error {
		return fmt.Errorf("parse arch output: module row %q missing %s attribute", *r.Path, attr)
	}
	if r.Types == nil {
		return archModuleRow{}, missing("types")
	}
	if r.Abstract == nil {
		return archModuleRow{}, missing("abstract")
	}

	unknown := r.Zone == archZoneUnknown
	row := archModuleRow{
		Path:          *r.Path,
		Types:         *r.Types,
		Abstract:      *r.Abstract,
		CouplingFloor: unknown,
	}
	if unknown {
		row.NoFileEdges = splitLanguages(r.NoFileEdges)
	}

	switch {
	case r.Ca != nil:
		row.Ca = *r.Ca
	case !unknown:
		return archModuleRow{}, missing("ca")
	}
	switch {
	case r.Ce != nil:
		row.Ce = *r.Ce
	case !unknown:
		return archModuleRow{}, missing("ce")
	}
	return row, nil
}

// toPrunedPackages converts the <pruned_packages> element: its exact
// count= and the listed paths, sorted and de-duplicated. A missing count=
// or a <p> row without path= is an error.
func (p archPrunedXML) toPrunedPackages() ([]string, int, error) {
	if p.Count == nil {
		return nil, 0, fmt.Errorf("parse arch output: <pruned_packages> missing count attribute")
	}
	seen := make(map[string]bool, len(p.Paths))
	paths := make([]string, 0, len(p.Paths))
	for _, entry := range p.Paths {
		if entry.Path == nil {
			return nil, 0, fmt.Errorf("parse arch output: <pruned_packages> row missing path attribute")
		}
		if !seen[*entry.Path] {
			seen[*entry.Path] = true
			paths = append(paths, *entry.Path)
		}
	}
	sort.Strings(paths)
	return paths, *p.Count, nil
}

// splitLanguages splits a comma-separated language list, dropping empty
// entries, and returns it sorted and de-duplicated.
func splitLanguages(list string) []string {
	seen := map[string]bool{}
	var langs []string
	for _, lang := range strings.Split(list, ",") {
		lang = strings.TrimSpace(lang)
		if lang == "" || seen[lang] {
			continue
		}
		seen[lang] = true
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

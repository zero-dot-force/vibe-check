package ripwireadapter

import (
	"encoding/xml"
	"fmt"
)

// archModuleRow is one directory-grain row of ripwire's --arch metrics
// table: the per-module Martin metrics ripwire computed over the crawled
// tree. Path is the raw root-relative form ripwire prints ("./"-prefixed
// for non-root modules, "." for the crawl root itself); path normalization
// is the directory-module mapping's job (task 4.3), so the parser stays
// faithful to the wire format.
type archModuleRow struct {
	// Path is the row's root-relative module path, verbatim.
	Path string
	// Ca and Ce are the module's afferent and efferent coupling counts.
	Ca, Ce int
	// Types and Abstract are the module's type-declaration count and its
	// abstract-type count (ripwire's interface share for TypeScript;
	// structurally zero for Python).
	Types, Abstract int
	// Instability, Abstractness, and Distance are ripwire's I, A, and D.
	Instability, Abstractness, Distance float64
	// Zone is ripwire's main-sequence verdict for the row (pain, useless,
	// ok, or n/a for untyped modules).
	Zone string
	// Reachable and Isolated mirror ripwire's reachability flags.
	Reachable, Isolated bool
}

// archTable is the parsed --arch <metrics> table: the aggregate counters
// ripwire reports alongside its per-module rows.
type archTable struct {
	// Modules is the total number of directory modules in the table.
	Modules int
	// TypedModules is the number of modules with a nonzero type count.
	TypedModules int
	// ZonePain, ZoneUseless, ZoneOk, and ZoneNa partition Modules by
	// ripwire's main-sequence zones (n/a covers untyped modules).
	ZonePain, ZoneUseless, ZoneOk, ZoneNa int
	// PropagationCost is ripwire's MacCormack propagation-cost density.
	PropagationCost float64
	// Rows are the per-module rows, in document order.
	Rows []archModuleRow
}

// archRowXML, archMetricsXML, and archRootXML are the wire shapes. Every
// required attribute uses a pointer so a missing attribute is
// distinguishable from a legitimate zero and fails closed instead of
// silently decoding as 0.
type archRowXML struct {
	Path         *string  `xml:"path,attr"`
	Ca           *int     `xml:"ca,attr"`
	Ce           *int     `xml:"ce,attr"`
	Types        *int     `xml:"types,attr"`
	Abstract     *int     `xml:"abstract,attr"`
	Instability  *float64 `xml:"I,attr"`
	Abstractness *float64 `xml:"A,attr"`
	Distance     *float64 `xml:"D,attr"`
	Zone         *string  `xml:"zone,attr"`
	Reachable    bool     `xml:"reachable,attr"`
	Isolated     bool     `xml:"isolated,attr"`
}

type archMetricsXML struct {
	Modules         *int         `xml:"modules,attr"`
	TypedModules    *int         `xml:"typed_modules,attr"`
	ZonePain        *int         `xml:"zone_pain,attr"`
	ZoneUseless     *int         `xml:"zone_useless,attr"`
	ZoneOk          *int         `xml:"zone_ok,attr"`
	ZoneNa          *int         `xml:"zone_na,attr"`
	PropagationCost *float64     `xml:"propagation_cost,attr"`
	Rows            []archRowXML `xml:"m"`
}

type archRootXML struct {
	XMLName xml.Name        `xml:"arch"`
	Metrics *archMetricsXML `xml:"metrics"`
}

// parseArchOutput parses the --arch metrics table from ripwire 0.6.9
// output. It reads ONLY the aggregated per-directory <metrics> rows —
// never per-file <inc> lists, which ripwire caps at 40 targets per file —
// so module coupling counts are always the uncapped table values (design
// D3: ripwire's table is the authoritative source for Ca/Ce).
//
// A present-but-rowless <metrics> element parses to an empty table without
// error (an empty crawl is the pipeline's problem to classify, not the
// parser's); a missing <metrics> element, a wrong root element, malformed
// XML, a row missing any required attribute, or duplicate module paths are
// errors — deterministic pinned input must never produce silent
// zero-valued or last-wins rows.
func parseArchOutput(data []byte) (archTable, error) {
	var root archRootXML
	if err := xml.Unmarshal(data, &root); err != nil {
		return archTable{}, fmt.Errorf("parse arch output: %w", err)
	}
	if root.Metrics == nil {
		return archTable{}, fmt.Errorf("parse arch output: no <metrics> table in --arch output")
	}
	m := root.Metrics
	aggregates := map[string]bool{
		"modules":          m.Modules != nil,
		"typed_modules":    m.TypedModules != nil,
		"zone_pain":        m.ZonePain != nil,
		"zone_useless":     m.ZoneUseless != nil,
		"zone_ok":          m.ZoneOk != nil,
		"zone_na":          m.ZoneNa != nil,
		"propagation_cost": m.PropagationCost != nil,
	}
	for label, present := range aggregates {
		if !present {
			return archTable{}, fmt.Errorf("parse arch output: <metrics> missing %s attribute", label)
		}
	}

	seen := make(map[string]bool, len(m.Rows))
	table := archTable{
		Modules:         *m.Modules,
		TypedModules:    *m.TypedModules,
		ZonePain:        *m.ZonePain,
		ZoneUseless:     *m.ZoneUseless,
		ZoneOk:          *m.ZoneOk,
		ZoneNa:          *m.ZoneNa,
		PropagationCost: *m.PropagationCost,
		Rows:            make([]archModuleRow, 0, len(m.Rows)),
	}
	for _, row := range m.Rows {
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
	return table, nil
}

// toModuleRow converts one wire row, failing closed on missing required
// attributes.
func (r archRowXML) toModuleRow() (archModuleRow, error) {
	missing := func(attr string) error {
		return fmt.Errorf("parse arch output: module row %q missing %s attribute", deref(r.Path), attr)
	}
	if r.Path == nil {
		return archModuleRow{}, fmt.Errorf("parse arch output: module row missing path attribute")
	}
	if r.Ca == nil {
		return archModuleRow{}, missing("ca")
	}
	if r.Ce == nil {
		return archModuleRow{}, missing("ce")
	}
	if r.Types == nil {
		return archModuleRow{}, missing("types")
	}
	if r.Abstract == nil {
		return archModuleRow{}, missing("abstract")
	}
	if r.Instability == nil {
		return archModuleRow{}, missing("I")
	}
	if r.Abstractness == nil {
		return archModuleRow{}, missing("A")
	}
	if r.Distance == nil {
		return archModuleRow{}, missing("D")
	}
	if r.Zone == nil {
		return archModuleRow{}, missing("zone")
	}
	return archModuleRow{
		Path:         *r.Path,
		Ca:           *r.Ca,
		Ce:           *r.Ce,
		Types:        *r.Types,
		Abstract:     *r.Abstract,
		Instability:  *r.Instability,
		Abstractness: *r.Abstractness,
		Distance:     *r.Distance,
		Zone:         *r.Zone,
		Reachable:    r.Reachable,
		Isolated:     r.Isolated,
	}, nil
}

// deref returns *s or "<missing>" for nil pointers in error messages.
func deref(s *string) string {
	if s == nil {
		return "<missing>"
	}
	return *s
}

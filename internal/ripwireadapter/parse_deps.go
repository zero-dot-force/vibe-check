package ripwireadapter

import (
	"encoding/xml"
	"fmt"
)

// depsGraph is the parsed --deps output: the file-to-file dependency view
// ripwire reports, verbatim. Target resolution (dotted Python module ids
// and relative TypeScript paths to directory modules) is the directory
// mapping's job (task 4.3); this type keeps <inc> targets exactly as
// ripwire printed them.
type depsGraph struct {
	// Window and paging attributes from the <deps> root.
	Files, Shown, Capped, Total        int
	HasMore, NextOffset, Offset, Limit int
	Root                               string

	// Health carries the <health> aggregates.
	Health depsHealth

	// Godfiles lists the highest-afferent files ripwire calls out.
	Godfiles []depsGodfile

	// Cycles lists ripwire's own include-cycle listing. The Python
	// recording has none (its file-level graph is acyclic); the
	// TypeScript recording has one size-2 cycle. An absent <cycles>
	// element parses to an empty slice.
	Cycles []depsCycle

	// Rows are the dependency-capable files with their include targets.
	Rows []depsFile
}

// depsHealth mirrors the <health> element.
type depsHealth struct {
	Files, DepFiles int
	CCD             int
	ACD, NCCD       float64
	Shape           string
	DepLangs        string
}

// depsGodfile is one <godfiles> entry.
type depsGodfile struct {
	Path     string
	Afferent int
}

// depsCycle is one <cycles><cycle> listing: ripwire's own (capped) cycle
// report, used by task 4.4 for corroboration of the adapter's SCC
// derivation.
type depsCycle struct {
	Size    int
	Cost    int
	Cut     string
	Cutrefs string
	Members []string
}

// depsFile is one <f> row: a dependency-capable file and its raw include
// targets.
type depsFile struct {
	Path        string
	Includes    int
	Afferent    int
	Instability float64
	Transitive  int
	IncTargets  []string
}

// depsRootXML, depsHealthXML, depsGodfilesXML, depsCycleXML, depsFileXML,
// and depsIncXML are the wire shapes. Required attributes use pointers so
// absence fails closed instead of silently decoding as zero.
type depsRootXML struct {
	XMLName    xml.Name        `xml:"deps"`
	Files      *int            `xml:"files,attr"`
	Shown      *int            `xml:"shown,attr"`
	Capped     *int            `xml:"capped,attr"`
	Total      *int            `xml:"total,attr"`
	HasMore    *int            `xml:"has_more,attr"`
	NextOffset *int            `xml:"next_offset,attr"`
	Offset     *int            `xml:"offset,attr"`
	Limit      *int            `xml:"limit,attr"`
	Root       *string         `xml:"root,attr"`
	Health     depsHealthXML   `xml:"health"`
	Godfiles   depsGodfilesXML `xml:"godfiles"`
	Cycles     depsCyclesXML   `xml:"cycles"`
	Rows       []depsFileXML   `xml:"f"`
}

type depsHealthXML struct {
	Files    *int     `xml:"files,attr"`
	DepFiles *int     `xml:"dep_files,attr"`
	CCD      *int     `xml:"ccd,attr"`
	ACD      *float64 `xml:"acd,attr"`
	NCCD     *float64 `xml:"nccd,attr"`
	Shape    *string  `xml:"shape,attr"`
	DepLangs *string  `xml:"dep_langs,attr"`
}

type depsGodfilesXML struct {
	Total  *int             `xml:"total,attr"`
	Shown  *int             `xml:"shown,attr"`
	Capped *int             `xml:"capped,attr"`
	Files  []depsSimpleFXML `xml:"f"`
}

type depsSimpleFXML struct {
	Path     *string `xml:"p,attr"`
	Afferent *int    `xml:"afferent,attr"`
}

type depsCyclesXML struct {
	Cycles []depsCycleXML `xml:"cycle"`
}

type depsCycleXML struct {
	Size    *int             `xml:"size,attr"`
	Cost    *int             `xml:"cost,attr"`
	Cut     *string          `xml:"cut,attr"`
	Cutrefs *string          `xml:"cutrefs,attr"`
	Members []depsSimpleFXML `xml:"f"`
}

type depsFileXML struct {
	Path        *string      `xml:"p,attr"`
	Includes    *int         `xml:"includes,attr"`
	Afferent    *int         `xml:"afferent,attr"`
	Instability *float64     `xml:"instab,attr"`
	Transitive  *int         `xml:"transitive,attr"`
	Incs        []depsIncXML `xml:"inc"`
}

type depsIncXML struct {
	Target *string `xml:"t,attr"`
}

// parseDepsOutput parses the --deps XML from ripwire 0.6.9. The root's
// paging attributes, <health>, <godfiles>, <cycles>, and every <f> row
// with its <inc> targets are captured verbatim. Malformed XML, a wrong
// root element, missing required attributes, or duplicate file paths are
// errors; a missing <cycles> element is a valid empty listing.
func parseDepsOutput(data []byte) (depsGraph, error) {
	var root depsRootXML
	if err := xml.Unmarshal(data, &root); err != nil {
		return depsGraph{}, fmt.Errorf("parse deps output: %w", err)
	}
	if err := requireInts(map[string]*int{
		"files": root.Files, "shown": root.Shown, "capped": root.Capped,
		"total": root.Total, "has_more": root.HasMore,
		"next_offset": root.NextOffset, "offset": root.Offset, "limit": root.Limit,
	}); err != nil {
		return depsGraph{}, fmt.Errorf("parse deps output: <deps> %w", err)
	}
	if root.Root == nil {
		return depsGraph{}, fmt.Errorf("parse deps output: <deps> missing root attribute")
	}
	h := root.Health
	if err := requireInts(map[string]*int{
		"health files": h.Files, "health dep_files": h.DepFiles, "health ccd": h.CCD,
	}); err != nil {
		return depsGraph{}, fmt.Errorf("parse deps output: %w", err)
	}
	if h.ACD == nil || h.NCCD == nil || h.Shape == nil || h.DepLangs == nil {
		return depsGraph{}, fmt.Errorf("parse deps output: <health> missing attribute")
	}

	graph := depsGraph{
		Files: *root.Files, Shown: *root.Shown, Capped: *root.Capped,
		Total: *root.Total, HasMore: *root.HasMore,
		NextOffset: *root.NextOffset, Offset: *root.Offset, Limit: *root.Limit,
		Root: *root.Root,
		Health: depsHealth{
			Files: *h.Files, DepFiles: *h.DepFiles, CCD: *h.CCD,
			ACD: *h.ACD, NCCD: *h.NCCD, Shape: *h.Shape, DepLangs: *h.DepLangs,
		},
		Godfiles: make([]depsGodfile, 0, len(root.Godfiles.Files)),
		Cycles:   make([]depsCycle, 0, len(root.Cycles.Cycles)),
		Rows:     make([]depsFile, 0, len(root.Rows)),
	}

	for _, g := range root.Godfiles.Files {
		if g.Path == nil || g.Afferent == nil {
			return depsGraph{}, fmt.Errorf("parse deps output: <godfiles> entry missing p or afferent")
		}
		graph.Godfiles = append(graph.Godfiles, depsGodfile{Path: *g.Path, Afferent: *g.Afferent})
	}

	for _, c := range root.Cycles.Cycles {
		if c.Size == nil || c.Cost == nil {
			return depsGraph{}, fmt.Errorf("parse deps output: <cycle> missing size or cost")
		}
		cycle := depsCycle{Size: *c.Size, Cost: *c.Cost, Members: make([]string, 0, len(c.Members))}
		if c.Cut != nil {
			cycle.Cut = *c.Cut
		}
		if c.Cutrefs != nil {
			cycle.Cutrefs = *c.Cutrefs
		}
		for _, m := range c.Members {
			if m.Path == nil {
				return depsGraph{}, fmt.Errorf("parse deps output: <cycle> member missing p")
			}
			cycle.Members = append(cycle.Members, *m.Path)
		}
		graph.Cycles = append(graph.Cycles, cycle)
	}

	seen := make(map[string]bool, len(root.Rows))
	for _, f := range root.Rows {
		if f.Path == nil {
			return depsGraph{}, fmt.Errorf("parse deps output: <f> row missing p attribute")
		}
		if err := requireInts(map[string]*int{
			"includes": f.Includes, "afferent": f.Afferent, "transitive": f.Transitive,
		}); err != nil {
			return depsGraph{}, fmt.Errorf("parse deps output: <f> %q %w", *f.Path, err)
		}
		if f.Instability == nil {
			return depsGraph{}, fmt.Errorf("parse deps output: <f> %q missing instab attribute", *f.Path)
		}
		if seen[*f.Path] {
			return depsGraph{}, fmt.Errorf("parse deps output: duplicate file path %q", *f.Path)
		}
		seen[*f.Path] = true

		file := depsFile{
			Path: *f.Path, Includes: *f.Includes, Afferent: *f.Afferent,
			Instability: *f.Instability, Transitive: *f.Transitive,
			IncTargets: make([]string, 0, len(f.Incs)),
		}
		for _, inc := range f.Incs {
			if inc.Target == nil {
				return depsGraph{}, fmt.Errorf("parse deps output: <f> %q has <inc> missing t", *f.Path)
			}
			file.IncTargets = append(file.IncTargets, *inc.Target)
		}
		graph.Rows = append(graph.Rows, file)
	}
	return graph, nil
}

// requireInts fails closed naming the first nil pointer among attrs.
func requireInts(attrs map[string]*int) error {
	for label, v := range attrs {
		if v == nil {
			return fmt.Errorf("missing %s attribute", label)
		}
	}
	return nil
}

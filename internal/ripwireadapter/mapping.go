package ripwireadapter

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Warning codes emitted by the directory-module mapping.
const (
	warningDirectoryAsModule = "directory-as-module"
	warningNameResolvedEdges = "name-resolved-edges"
	warningAmbiguityRate     = "ambiguity-rate"
)

// mappedModule is one directory-grain module after identity mapping and
// aggregation. Ca/Ce/Types/Abstract are copied verbatim from the
// authoritative --arch metrics row (design D3 — never recounted from
// per-file edges, which are capped); the Internal*/External* counters are
// derived from the deduplicated directory-pair edge set and feed cycle
// derivation (task 4.4) and the Ce-only external accounting.
type mappedModule struct {
	// Identity is the D4 module identity: dot-style for Python
	// (fixturepy.cycle_a), slash-style for TypeScript (src/cycle-a), "."
	// for a crawl-root row.
	Identity string
	// Name is the last identity segment.
	Name string
	// ArchPath is the raw ripwire --arch row path, kept for diagnostics.
	ArchPath string
	// Ca, Ce, Types, Abstract are the authoritative arch-row values.
	Ca, Ce, Types, Abstract int
	// InternalOut and InternalIn count deduplicated directory-pair edges.
	InternalOut, InternalIn int
	// ExternalOut counts include targets that resolve outside the module
	// universe (external dependencies): counted in Ce only, they add no
	// module and no afferent coupling anywhere.
	ExternalOut int
}

// moduleEdge is one deduplicated directory-grain dependency edge between
// module identities.
type moduleEdge struct {
	From, To string
}

// moduleMapping is the result of mapping ripwire's parsed outputs onto the
// universal directory-module model.
type moduleMapping struct {
	// Modules sorted by Identity.
	Modules []mappedModule
	// Edges deduplicated per (From, To) pair, sorted lexicographically.
	Edges []moduleEdge
	// Warnings carries directory-as-module (always), name-resolved-edges
	// (Python dotted-target resolution), and ambiguity-rate (targets that
	// land inside the crawl but match no module).
	Warnings []metrics.Warning
}

// mapToModules maps the parsed --arch table and --deps graph onto
// directory-grain modules with D4 identities. Language selects the
// identity style (languagePython → dot-style, languageTypeScript →
// slash-style). The mapping is deterministic: identical inputs yield
// deep-equal outputs.
//
// Errors: an empty arch table, two arch rows mapping to the same identity,
// or a dependency-capable file that sits under no module directory (fail
// closed — a pinned surface must not produce silently unmapped files).
func mapToModules(language string, table archTable, deps depsGraph) (moduleMapping, error) {
	if len(table.Rows) == 0 {
		return moduleMapping{}, fmt.Errorf("map to modules: empty arch table")
	}

	// Build the module universe keyed by directory path (slash form).
	moduleByDir := make(map[string]*mappedModule, len(table.Rows))
	identitySeen := make(map[string]bool, len(table.Rows))
	modules := make([]mappedModule, 0, len(table.Rows))
	for _, row := range table.Rows {
		identity, err := moduleIdentity(language, row.Path)
		if err != nil {
			return moduleMapping{}, err
		}
		if identitySeen[identity] {
			return moduleMapping{}, fmt.Errorf("map to modules: duplicate identity %q (from %q)", identity, row.Path)
		}
		identitySeen[identity] = true
		modules = append(modules, mappedModule{
			Identity: identity,
			Name:     moduleName(identity),
			ArchPath: row.Path,
			Ca:       row.Ca,
			Ce:       row.Ce,
			Types:    row.Types,
			Abstract: row.Abstract,
		})
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].Identity < modules[j].Identity })
	for i := range modules {
		dir := identityToDir(modules[i].Identity)
		moduleByDir[dir] = &modules[i]
	}

	mapping := moduleMapping{
		Modules:  modules,
		Edges:    []moduleEdge{},
		Warnings: []metrics.Warning{directoryAsModuleWarning()},
	}

	// Resolve every include target to a directory-grain edge, external,
	// or unresolved-in-crawl.
	edgeSet := make(map[moduleEdge]bool)
	ambiguous := false
	nameResolved := false

	for _, file := range deps.Rows {
		sourceDir, err := fileModuleDir(file.Path)
		if err != nil {
			return moduleMapping{}, fmt.Errorf("map to modules: %w", err)
		}
		sourceModule, ok := moduleByDir[sourceDir]
		if !ok {
			return moduleMapping{}, fmt.Errorf(
				"map to modules: dependency file %q sits under no module directory", file.Path)
		}
		for _, target := range file.IncTargets {
			resolved, isNameResolved, err := resolveTarget(language, sourceDir, target, moduleByDir)
			if err != nil {
				return moduleMapping{}, err
			}
			nameResolved = nameResolved || isNameResolved
			switch {
			case resolved.external:
				sourceModule.ExternalOut++
			case resolved.unresolved:
				ambiguous = true
			default:
				edge := moduleEdge{From: sourceModule.Identity, To: resolved.identity}
				if !edgeSet[edge] {
					edgeSet[edge] = true
					sourceModule.InternalOut++
					if target, ok := moduleByDir[identityToDir(resolved.identity)]; ok {
						target.InternalIn++
					}
				}
			}
		}
	}

	for edge := range edgeSet {
		mapping.Edges = append(mapping.Edges, edge)
	}
	sort.Slice(mapping.Edges, func(i, j int) bool {
		if mapping.Edges[i].From != mapping.Edges[j].From {
			return mapping.Edges[i].From < mapping.Edges[j].From
		}
		return mapping.Edges[i].To < mapping.Edges[j].To
	})

	if nameResolved {
		mapping.Warnings = append(mapping.Warnings, metrics.Warning{
			Code: warningNameResolvedEdges,
			Message: "edge targets were resolved from dotted module names to directories; " +
				"name-based resolution can misattribute when package names collide",
		})
	}
	if ambiguous {
		mapping.Warnings = append(mapping.Warnings, metrics.Warning{
			Code: warningAmbiguityRate,
			Message: "some include targets land inside the crawl but match no module directory; " +
				"their edges were dropped from the module graph",
		})
	}
	return mapping, nil
}

// targetResolution classifies a resolved include target.
type targetResolution struct {
	// identity is the resolved module identity when internal.
	identity string
	// external marks a target outside the module universe.
	external bool
	// unresolved marks a target inside the crawl but matching no module.
	unresolved bool
}

// resolveTarget resolves one include target from a source file directory
// to an internal module edge, an external dependency, or an unresolved
// in-crawl reference. Python targets are dotted module ids resolved by
// longest directory prefix (name-based); TypeScript targets are relative
// paths joined against the source directory (path-based).
func resolveTarget(language, sourceDir, target string, moduleByDir map[string]*mappedModule) (targetResolution, bool, error) {
	if language == languagePython {
		// Dotted module id: try the longest prefix of components.
		components := strings.Split(target, ".")
		for len(components) > 0 {
			dir := strings.Join(components, "/")
			if moduleByDir[dir] != nil {
				return targetResolution{identity: moduleByDir[dir].Identity}, true, nil
			}
			components = components[:len(components)-1]
		}
		// No prefix matched a module: an external dependency (stdlib or
		// third-party). Externals add no edge, so the name-based method
		// risked nothing — the warning only fires on resolved edges.
		return targetResolution{external: true}, false, nil
	}

	// Relative path target: join against the source file's directory.
	joined := path.Join(sourceDir, target)
	if moduleByDir[joined] != nil {
		return targetResolution{identity: moduleByDir[joined].Identity}, false, nil
	}
	if strings.HasPrefix(joined, "../") || joined == ".." {
		return targetResolution{external: true}, false, nil
	}
	// Inside the crawl tree but not a module directory.
	return targetResolution{unresolved: true}, false, nil
}

// moduleIdentity converts a ripwire --arch row path to a D4 module
// identity: strip the "./" prefix, then dot-style (Python) or slash-style
// (TypeScript).
func moduleIdentity(language, archPath string) (string, error) {
	trimmed := strings.TrimPrefix(archPath, "./")
	if trimmed == "" {
		return "", fmt.Errorf("map to modules: empty arch row path %q", archPath)
	}
	if trimmed == "." {
		return ".", nil
	}
	if language == languagePython {
		return strings.ReplaceAll(trimmed, "/", "."), nil
	}
	return trimmed, nil
}

// moduleName returns the last identity segment.
func moduleName(identity string) string {
	if identity == "." {
		return "."
	}
	if i := strings.LastIndex(identity, "."); i >= 0 {
		return identity[i+1:]
	}
	if i := strings.LastIndex(identity, "/"); i >= 0 {
		return identity[i+1:]
	}
	return identity
}

// identityToDir converts a module identity back to its slash-form
// directory path (both styles key the same directory space).
func identityToDir(identity string) string {
	if identity == "." {
		return "."
	}
	return strings.ReplaceAll(identity, ".", "/")
}

// fileModuleDir returns the directory of a root-relative file path.
func fileModuleDir(filePath string) (string, error) {
	return path.Dir(filePath), nil
}

// directoryAsModuleWarning builds the always-present granularity warning.
func directoryAsModuleWarning() metrics.Warning {
	return metrics.Warning{
		Code: warningDirectoryAsModule,
		Message: "modules are mapped at directory grain: coupling between files in the same " +
			"directory is intra-module and invisible to the graph",
	}
}

package ripwireadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// warningUnsupportedMetric is the warning code for metrics this adapter
// reports as zero-valued because ripwire offers no validated signal for
// them.
const warningUnsupportedMetric = "unsupported-metric"

// warningDepsWindowTruncated is the warning code for a --deps listing the
// adapter requested with a window too small for the corpus: file rows
// beyond the window are absent from the edge data.
const warningDepsWindowTruncated = "deps-window-truncated"

// depsWindowWarning returns the truncation warning for a --deps listing
// whose window did not cover every dependency-capable file, or nil when
// the window covered the corpus. The adapter requests --limit=1000 per
// crawl (the capture convention); a larger corpus needs the edges the
// listing is missing, so the graph is partial rather than silently
// incomplete.
func depsWindowWarning(deps depsGraph) []metrics.Warning {
	if deps.HasMore == 0 {
		return nil
	}
	return []metrics.Warning{{
		Code: warningDepsWindowTruncated,
		Message: fmt.Sprintf(
			"the --deps listing window showed %d of %d files: edges for files beyond the window are missing",
			deps.Shown, deps.Total,
		),
	}}
}

// minimalArchRules is the rules file content the adapter generates for
// its --arch invocation. The --arch mode requires a rules file, but the
// adapter consumes only the Martin metrics table, which layer and deny
// rules do not alter. This was validated against the pinned ripwire
// 0.6.9: a comment-only rules file yields a byte-identical <metrics>
// table (only the unused layers count attribute differs), so the adapter
// ships no fixture-specific rules and enforces nothing.
const minimalArchRules = `# vibe-check generated arch rules: intentionally empty.
# The adapter consumes only the Martin metrics table --arch emits; layer
# and deny rules do not affect it, and no rule can fire a violation.
`

// sourceToolName is the provenance source-tool identifier (design D11).
const sourceToolName = "ripwire"

// Analyze performs a complete analysis of the Python or TypeScript project
// at projectPath and returns a [metrics.ModuleGraph].
//
// The pipeline, in order:
//
//  1. Fail-closed version pin: the ripwire binary is resolved on PATH and
//     its version and commit validated against the pinned pair before any
//     analysis work (task 2.2). The validated pair also populates the
//     source-tool provenance fields.
//  2. Crawl-root resolution per language (design D9), then three ripwire
//     crawls from that root: the --arch metrics table (with a generated
//     minimal rules file), the --deps file graph, and the --export=cc.json
//     per-file metrics (the crawl census feeding the zero-coupling guard).
//  3. Parsing (tasks 4.1/4.2), directory-module mapping (task 4.3), and
//     SCC cycle derivation with corroboration (task 4.4).
//  4. The light language parse on the same crawl root: Python abstractness
//     (task 5.1) and LCOM4 for both languages (task 5.2).
//  5. Hybrid assembly: Ca/Ce from the authoritative arch table, TS
//     abstractness from the arch interface share, Python abstractness from
//     the light parse, LCOM from the light parse, and every derived value
//     (instability, abstractness, distance, zone) via metrics.Compute* —
//     never reimplemented. Duplication is unsupported: zero-valued fields
//     plus a warning.
//
// Warnings from every stage are assembled, sorted by code (stable, so
// same-code warnings keep their emission order), and any warning forces
// status partial. Modules are sorted by path; module, cycle, and warning
// slices are always non-nil. Provenance carries the analyzed path, the
// project's manifest identity (pyproject.toml or package.json name, with
// a directory-name fallback), and the validated source-tool identity.
func (a *Adapter) Analyze(ctx context.Context, projectPath string) (*metrics.ModuleGraph, error) {
	bin, info, err := validatedRipwire(ctx)
	if err != nil {
		return nil, err
	}

	crawlRoot, err := a.resolveCrawlRoot(projectPath)
	if err != nil {
		return nil, err
	}
	moduleIdentity, err := a.resolveModuleIdentity(projectPath, crawlRoot)
	if err != nil {
		return nil, err
	}

	rulesPath, err := writeMinimalArchRules()
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(rulesPath) }()

	runner := newRipwireRunner(bin)
	scope := newCrawlScope(crawlRoot)

	archOut, err := runner.run(ctx, crawlRoot, append(scope.ripwireArgs(), "--arch="+rulesPath, "--no-cache")...)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: arch crawl of %s: %w", projectPath, err)
	}
	depsOut, err := runner.run(ctx, crawlRoot, append(scope.ripwireArgs(), "--deps", "--limit=1000", "--no-cache")...)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: deps crawl of %s: %w", projectPath, err)
	}
	ccOut, err := runner.run(ctx, crawlRoot, append(scope.ripwireArgs(), "--export=cc.json", "--no-cache")...)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: cc export of %s: %w", projectPath, err)
	}

	table, err := parseArchOutput(archOut)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: arch output of %s: %w", projectPath, err)
	}
	deps, err := parseDepsOutput(depsOut)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: deps output of %s: %w", projectPath, err)
	}
	ccFiles, err := parseCCOutput(ccOut)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: cc output of %s: %w", projectPath, err)
	}

	mapping, err := mapToModules(a.language, table, deps)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: mapping %s: %w", projectPath, err)
	}
	cycles, cycleWarnings := deriveCycles(mapping, deps)

	// Light parses over the same crawl root the coupling data saw.
	var abstractnessScan map[string]lightParseResult
	var abstractnessWarnings []metrics.Warning
	if a.language == languagePython {
		abstractnessScan, abstractnessWarnings, err = pythonAbstractnessScan(crawlRoot)
		if err != nil {
			return nil, fmt.Errorf("ripwire adapter: abstractness scan of %s: %w", projectPath, err)
		}
	}
	lcomScan, lcomWarnings, err := lcom4Scan(crawlRoot, a.language)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: lcom scan of %s: %w", projectPath, err)
	}

	// Guard and denylist warnings.
	guardWarnings := zeroCouplingGuard(len(mapping.Edges), len(ccFiles))
	denylistWarnings, err := detectDenylistPruned(crawlRoot)
	if err != nil {
		return nil, fmt.Errorf("ripwire adapter: denylist pre-check of %s: %w", projectPath, err)
	}

	// Module assembly, sorted by identity.
	sorted := append([]mappedModule(nil), mapping.Modules...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Identity < sorted[j].Identity })
	modules := make([]metrics.ModuleResult, 0, len(sorted))
	for _, m := range sorted {
		abstractTypes, exportedTypes := m.Abstract, m.Types
		if a.language == languagePython {
			// D8: ripwire's Python abstractness is structurally zero;
			// the light parse is the only source.
			if lp, ok := abstractnessScan[identityToDir(m.Identity)]; ok {
				abstractTypes, exportedTypes = lp.AbstractTypes, lp.ExportedTypes
			} else {
				abstractTypes, exportedTypes = 0, 0
			}
		}
		result := metrics.ModuleResult{
			Module: metrics.Module{
				Path:          m.Identity,
				Name:          moduleName(m.Identity),
				Ca:            m.Ca,
				Ce:            m.Ce,
				ExportedTypes: exportedTypes,
				AbstractTypes: abstractTypes,
			},
			LCOM:         metrics.LCOM(lcomScan[identityToDir(m.Identity)]),
			Duplications: []metrics.Duplication{},
			TotalLines:   0,
		}
		result.Instability = metrics.ComputeInstability(m.Ca, m.Ce)
		result.Abstractness = metrics.ComputeAbstractness(abstractTypes, exportedTypes)
		result.Distance = metrics.ComputeDistance(result.Abstractness, result.Instability)
		result.Zone = metrics.ComputeZone(result.Abstractness, result.Instability, result.Distance)
		modules = append(modules, result)
	}

	// Warning assembly.
	warnings := make([]metrics.Warning, 0)
	warnings = append(warnings, denylistWarnings...)
	warnings = append(warnings, guardWarnings...)
	warnings = append(warnings, depsWindowWarning(deps)...)
	warnings = append(warnings, mapping.Warnings...)
	warnings = append(warnings, cycleWarnings...)
	warnings = append(warnings, abstractnessWarnings...)
	warnings = append(warnings, lcomWarnings...)
	warnings = append(warnings, errorBudgetWarning(a.language))
	if a.language == languageTypeScript {
		warnings = append(warnings, tsAbstractnessBiasWarning())
	}
	warnings = append(warnings, metrics.Warning{
		Code: warningUnsupportedMetric,
		Message: "duplication is not computed: ripwire offers no validated duplication " +
			"signal, so duplication fields are zero-valued",
	})
	sort.SliceStable(warnings, func(i, j int) bool { return warnings[i].Code < warnings[j].Code })

	return &metrics.ModuleGraph{
		SchemaVersion: metrics.SchemaVersionCurrent,
		Language:      a.language,
		Modules:       modules,
		Cycles:        cycles,
		Warnings:      warnings,
		Status:        graphStatusForWarnings(warnings),
		Provenance: &metrics.Provenance{
			Input: metrics.ProvenanceInput{
				Path:              filepath.Clean(projectPath),
				ModulePath:        moduleIdentity,
				SourceTool:        sourceToolName,
				SourceToolVersion: info.version,
				SourceToolCommit:  info.commit,
			},
		},
	}, nil
}

// resolveCrawlRoot resolves the crawl root for the adapter's language:
// the Python src-layout rules for Python, the TypeScript manifest rules
// for TypeScript (design D9, task 3.1).
func (a *Adapter) resolveCrawlRoot(projectPath string) (string, error) {
	if a.language == languagePython {
		return resolvePythonCrawlRoot(projectPath)
	}
	return resolveTypeScriptCrawlRoot(projectPath)
}

// resolveModuleIdentity returns the provenance ModulePath analog for the
// analyzed project: the distribution name from pyproject.toml for Python
// or the package name from package.json for TypeScript, falling back to
// the crawl root's base directory name when no manifest names the
// project. metrics.Validate requires a non-empty modulePath whenever
// provenance is present, so the fallback guarantees validity for
// manifest-less trees.
func (a *Adapter) resolveModuleIdentity(projectPath, crawlRoot string) (string, error) {
	var name string
	var err error
	if a.language == languagePython {
		name, err = pythonProjectName(projectPath)
	} else {
		name, err = typeScriptPackageName(projectPath)
	}
	if err != nil {
		return "", err
	}
	if name != "" {
		return name, nil
	}
	base := filepath.Base(filepath.Clean(crawlRoot))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return filepath.Base(filepath.Clean(projectPath)), nil
	}
	return base, nil
}

// pythonProjectName extracts the project name from pyproject.toml: the
// [project] table's name field, or [tool.poetry]'s for legacy layouts. A
// missing manifest or a nameless manifest returns "" (the caller falls
// back); an unreadable manifest fails closed because the provenance
// identity must never be silently wrong.
func pythonProjectName(projectPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectPath, "pyproject.toml"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("ripwire adapter: read pyproject.toml: %w", err)
	}
	section := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if section != "project" && section != "tool.poetry" {
			continue
		}
		if name, ok := strings.CutPrefix(line, "name ="); ok {
			name = strings.TrimSpace(name)
			name = strings.Trim(name, `"'`)
			if name != "" {
				return name, nil
			}
		}
	}
	return "", nil
}

// typeScriptPackageName extracts the package name from package.json. A
// missing manifest or a nameless manifest returns "" (the caller falls
// back); an unreadable or malformed manifest fails closed.
func typeScriptPackageName(projectPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("ripwire adapter: read package.json: %w", err)
	}
	var manifest struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", fmt.Errorf("ripwire adapter: parse package.json: %w", err)
	}
	return manifest.Name, nil
}

// writeMinimalArchRules writes the generated minimal --arch rules file to
// a private temporary directory and returns its path. Callers must remove
// the file when done; the directory is removed with it by the os package
// once empty.
func writeMinimalArchRules() (string, error) {
	dir, err := os.MkdirTemp("", "vibe-check-arch-rules-")
	if err != nil {
		return "", fmt.Errorf("ripwire adapter: arch rules temp dir: %w", err)
	}
	path := filepath.Join(dir, "arch-rules.txt")
	if err := os.WriteFile(path, []byte(minimalArchRules), 0o600); err != nil {
		return "", fmt.Errorf("ripwire adapter: arch rules temp file: %w", err)
	}
	return path, nil
}

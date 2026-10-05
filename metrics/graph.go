package metrics

// SchemaVersionCurrent is the current schema version for ModuleGraph output.
// Consumers should check this value before processing to detect incompatible changes.
// Version changes follow semantic versioning: minor versions are backward-compatible,
// major versions may contain breaking changes.
const SchemaVersionCurrent = "1.3"

// ModuleGraph represents the complete analysis result for a project.
// It contains all modules with their computed metrics, detected circular
// dependencies, and any warnings produced during analysis.
type ModuleGraph struct {
	// SchemaVersion is the version of the output schema (e.g., "1.3").
	// Consumers use this to detect breaking changes in the JSON structure.
	SchemaVersion string `json:"schemaVersion"`
	// Language is the lowercase language identifier (e.g., "go", "python").
	Language string `json:"language"`
	// Modules contains the analysis results for each module in the project.
	Modules []ModuleResult `json:"modules"`
	// Cycles contains detected circular dependencies between modules.
	Cycles []Cycle `json:"cycles"`
	// Warnings contains language-specific caveats about metric accuracy.
	// This slice is always non-nil (empty slice, not nil) when there are no warnings.
	Warnings []Warning `json:"warnings"`
	// Status indicates the overall analysis outcome.
	Status Status `json:"status"`
	// Provenance, when non-nil, records how this graph was produced: the
	// producer, version, generation timestamp, and the analyzed input. It is
	// omitted from JSON when nil (omitempty), so set it to nil for
	// byte-reproducible output. Provenance is metadata only — it never affects
	// metric computation or comparison via ComputeDelta.
	Provenance *Provenance `json:"provenance,omitempty"`
}

// Provenance records metadata about how a [ModuleGraph] was produced: the tool
// that generated it, the tool's version, the generation timestamp, and the
// input that was analyzed. It is metadata, not a metric — it never feeds
// [ComputeDelta] or [DecideVerdict], and it may be omitted entirely for
// byte-reproducible output.
type Provenance struct {
	// Producer is the name of the tool that produced the graph (e.g.,
	// "vibe-check"). Populated by the CLI.
	Producer string `json:"producer"`
	// Version is the semantic version of the producing tool (e.g., "0.1.0"),
	// without commit/date decoration. Populated by the CLI.
	Version string `json:"version"`
	// GeneratedAt is the RFC3339 UTC timestamp at which the graph was produced.
	// Populated by the CLI.
	GeneratedAt string `json:"generatedAt"`
	// Input identifies the source that was analyzed.
	Input ProvenanceInput `json:"input"`
}

// ProvenanceInput identifies the source that was analyzed to produce a graph.
// The language adapter populates it because it is the layer that resolves the
// input identity (the analyzed path and the resolved module path).
//
// CommitSHA and Branch are snapshot metadata populated by the CLI when the
// --store flag is set. They are empty when not in a git repository or when
// --store is not used.
type ProvenanceInput struct {
	// Path is the path to the analyzed project directory.
	Path string `json:"path"`
	// ModulePath is the resolved module import path (e.g., "github.com/foo/bar").
	ModulePath string `json:"modulePath"`
	// CommitSHA is the full git commit SHA (40 hex characters) at the time of
	// analysis. Empty when not in a git repository or when --store is not set.
	CommitSHA string `json:"commitSHA,omitempty"`
	// Branch is the git branch name at the time of analysis. Empty when not in
	// a git repository, on a detached HEAD, or when --store is not set.
	Branch string `json:"branch,omitempty"`
}

// ModuleResult combines Module identity data with computed metrics and zone
// classification. It embeds Module to provide raw data alongside derived values.
type ModuleResult struct {
	Module
	// Instability is the computed instability metric I = Ce / (Ca + Ce).
	Instability Instability `json:"instability"`
	// Abstractness is the computed abstractness metric A = abstractTypes / totalExported.
	Abstractness Abstractness `json:"abstractness"`
	// Distance is the computed distance from main sequence D = |A + I - 1|.
	Distance Distance `json:"distance"`
	// LCOM is the computed Lack of Cohesion of Methods (LCOM4 variant).
	LCOM LCOM `json:"lcom"`
	// Zone is the classification of the module's position relative to the main sequence.
	Zone Zone `json:"zone"`
	// Duplications contains detected structurally similar code blocks within
	// this module. Each Duplication groups two or more blocks that are
	// structurally identical after identifier and literal normalization.
	// An empty slice (not nil) indicates no duplications were detected.
	Duplications []Duplication `json:"duplications"`

	// TotalLines is the total number of lines across all source files in
	// this module. Used to compute duplication percentage for threshold
	// enforcement. Zero indicates the line count could not be determined.
	TotalLines int `json:"totalLines"`

	// Extensions contains language-specific metric extensions namespaced by language.
	// Keys use the format "language.metricName" (e.g., "go.interfaceWidth").
	// Extensions are not schema-enforced beyond being a valid JSON object.
	// Use language-specific typed accessor functions for safe extraction after
	// JSON round-trip (JSON unmarshaling converts int to float64, nested maps to
	// map[string]interface{}).
	Extensions map[string]any `json:"extensions,omitempty"`
}

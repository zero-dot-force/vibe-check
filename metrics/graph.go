package metrics

// ModuleGraph represents the complete analysis result for a project.
// It contains all modules with their computed metrics, detected circular
// dependencies, and any warnings produced during analysis.
type ModuleGraph struct {
	// SchemaVersion is the version of the output schema (e.g., "1.0").
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
}

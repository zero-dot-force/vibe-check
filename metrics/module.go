package metrics

// Module represents the universal unit of analysis across all languages.
// In Go this maps to a package, in Python to a module, in TS/JS to a file/module.
//
// Module contains raw metric input data that language-specific adapters populate.
// Computed metrics (Instability, Abstractness, Distance, LCOM) are derived from
// these fields and stored in [ModuleResult].
type Module struct {
	// Path is the unique identifier for this module (e.g., "github.com/foo/bar").
	Path string `json:"path"`

	// Name is the human-readable short name (e.g., "bar").
	Name string `json:"name"`

	// Ca is Afferent Coupling — the number of modules that depend on this module.
	// Ca is a non-negative integer; a module with no dependents has Ca = 0.
	Ca int `json:"ca"`

	// Ce is Efferent Coupling — the number of modules this module depends on.
	// Ce is a non-negative integer; a module with no dependencies has Ce = 0.
	Ce int `json:"ce"`

	// ExportedTypes is the total count of exported types in this module.
	ExportedTypes int `json:"exportedTypes"`

	// AbstractTypes is the count of abstract types (types that cannot be directly
	// instantiated, e.g., Go interfaces, Python ABCs, TypeScript abstract classes).
	// Each language adapter documents its mapping from language-specific constructs
	// to the abstract/concrete classification.
	AbstractTypes int `json:"abstractTypes"`
}

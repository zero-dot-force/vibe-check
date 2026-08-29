package metrics

// Warning represents a language-specific caveat that may affect metric accuracy.
// Warnings annotate analysis results with context without preventing metric computation.
type Warning struct {
	// Code is a machine-readable warning identifier (e.g., "dynamic-imports").
	Code string `json:"code"`
	// Message is a human-readable description of the warning.
	Message string `json:"message"`
	// Module is the path of the affected module (empty string if warning applies globally).
	Module string `json:"module,omitempty"`
}

package metrics

// Zone represents a module's position relative to the main sequence in the
// Abstractness-Instability graph.
type Zone string

const (
	// ZoneMainSequence indicates the module lies on or near the main sequence (D < 0.2).
	ZoneMainSequence Zone = "main-sequence"
	// ZoneOfPain indicates the module is concrete and stable (A < 0.2 and I < 0.2).
	// Modules in this zone are hard to change because they are depended upon but not abstract.
	ZoneOfPain Zone = "zone-of-pain"
	// ZoneOfUselessness indicates the module is abstract and unstable (A > 0.8 and I > 0.8).
	// Modules in this zone provide abstractions that have few dependents.
	ZoneOfUselessness Zone = "zone-of-uselessness"
	// ZoneNormal indicates the module does not fall into any special classification.
	ZoneNormal Zone = "normal"
)

// Status represents the overall analysis outcome.
type Status string

const (
	// StatusComplete indicates all metrics were computed successfully.
	StatusComplete Status = "complete"
	// StatusPartial indicates some metrics are unavailable due to adapter limitations.
	// When Status is Partial, the Warnings slice explains which metrics are unavailable.
	StatusPartial Status = "partial"
	// StatusError indicates analysis failed with partial or no results.
	StatusError Status = "error"
)

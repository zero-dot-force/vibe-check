package metrics

import "context"

// Adapter is the interface that all language-specific analyzers must implement.
// Adding a new language adapter requires only implementing this interface and
// registering with a [Registry].
type Adapter interface {
	// Analyze analyzes the project at projectPath and returns a complete ModuleGraph.
	Analyze(ctx context.Context, projectPath string) (*ModuleGraph, error)
	// Language returns the lowercase language identifier (e.g., "go", "python").
	Language() string
	// Capabilities returns the list of metrics this adapter can compute.
	Capabilities() []Capability
}

// Capability represents a metric that an adapter can compute.
type Capability string

const (
	// CapAfferentCoupling indicates the adapter can compute afferent coupling (Ca).
	CapAfferentCoupling Capability = "ca"
	// CapEfferentCoupling indicates the adapter can compute efferent coupling (Ce).
	CapEfferentCoupling Capability = "ce"
	// CapInstability indicates the adapter can compute instability (I).
	CapInstability Capability = "instability"
	// CapAbstractness indicates the adapter can compute abstractness (A).
	CapAbstractness Capability = "abstractness"
	// CapDistance indicates the adapter can compute distance from main sequence (D).
	CapDistance Capability = "distance"
	// CapLCOM indicates the adapter can compute Lack of Cohesion of Methods (LCOM4).
	CapLCOM Capability = "lcom"
	// CapCircularDeps indicates the adapter can detect circular dependencies.
	CapCircularDeps Capability = "circular"
)

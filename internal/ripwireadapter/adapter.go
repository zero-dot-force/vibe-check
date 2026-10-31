package ripwireadapter

import (
	"github.com/zero-dot-force/vibe-check/metrics"
)

// Language identifiers reported by the two constructors' adapters. They
// double as metrics.Registry keys, which are keyed by Language().
const (
	languagePython     = "python"
	languageTypeScript = "typescript"
)

// Compile-time interface compliance check, mirroring internal/goadapter.
var _ metrics.Adapter = (*Adapter)(nil)

// Adapter implements [metrics.Adapter] for Python and TypeScript codebases
// by hybridizing coupling data from an external ripwire binary (pinned to
// 0.6.9 at commit bedc0cd) with vibe-check's own light language parse. See
// the package documentation for the per-metric sourcing matrix.
//
// One struct type parameterized by language is used rather than two distinct
// types (openspec/changes/ripwire-adapter design.md D1): the languages share
// ~90% of the pipeline (discovery, version pin, invocation, parsing,
// mapping, cycle detection, warnings) and differ only in language identifier,
// crawl-root rules, abstractness source, and light-parse dialect —
// differences that hang off the language field as the later tasks of the
// ripwire-adapter change land them. Two distinct types would duplicate every
// method or force an embedding shim.
type Adapter struct {
	language string
}

// NewPython creates a ripwire-backed adapter for Python codebases. The
// returned adapter's Language method reports "python".
func NewPython() *Adapter {
	return &Adapter{language: languagePython}
}

// NewTypeScript creates a ripwire-backed adapter for TypeScript codebases.
// The returned adapter's Language method reports "typescript".
func NewTypeScript() *Adapter {
	return &Adapter{language: languageTypeScript}
}

// Language returns the lowercase language identifier of this adapter:
// "python" for adapters created by [NewPython], "typescript" for adapters
// created by [NewTypeScript].
func (a *Adapter) Language() string {
	return a.language
}

// Capabilities returns the universal metrics this adapter can compute:
// afferent coupling (Ca), efferent coupling (Ce), instability, abstractness,
// distance from main sequence, LCOM, and circular dependency detection.
//
// [metrics.CapDuplication] is deliberately not declared for either language:
// ripwire offers no validated duplication signal, so analysis reports
// zero-valued duplication fields plus an "unsupported-metric" warning rather
// than a fabricated metric (the ripwire-adapter spec's per-language
// capability gating requirement).
//
// The returned slice is freshly allocated on every call; callers may mutate
// or retain it without affecting later calls.
func (a *Adapter) Capabilities() []metrics.Capability {
	return []metrics.Capability{
		metrics.CapAfferentCoupling,
		metrics.CapEfferentCoupling,
		metrics.CapInstability,
		metrics.CapAbstractness,
		metrics.CapDistance,
		metrics.CapLCOM,
		metrics.CapCircularDeps,
	}
}

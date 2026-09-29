package main

import "time"

// producerName is the identity stamped into every provenance envelope emitted by
// the CLI. It is shared by analyze, diff, and init so downstream consumers can
// attribute any vibe-check output to this tool.
const producerName = "vibe-check"

// provenanceEnvelope is the machine-readable provenance metadata attached to the
// diff and init --json payloads. It carries the three fields every vibe-check
// artifact must expose per the constitution: who produced it, which version, and
// when. The envelope is intentionally a smaller shape than metrics.Provenance,
// which additionally carries an input sub-object that only analyze populates.
type provenanceEnvelope struct {
	// Producer identifies the tool that emitted the payload.
	Producer string `json:"producer"`
	// Version is the semantic version of the tool, sourced from the same
	// ldflags/build-info fallback as `vibe-check --version`.
	Version string `json:"version"`
	// GeneratedAt is the RFC 3339 UTC timestamp when the payload was produced.
	GeneratedAt string `json:"generatedAt"`
}

// newProvenanceEnvelope returns a fully populated provenance envelope. Producer
// is the shared producerName constant, Version is derived from semanticVersion,
// and GeneratedAt is the current UTC time formatted as RFC 3339.
func newProvenanceEnvelope() *provenanceEnvelope {
	return &provenanceEnvelope{
		Producer:    producerName,
		Version:     semanticVersion(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

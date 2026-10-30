package main

import "time"

// producerName is the identity stamped into every provenance envelope emitted
// by the CLI, so downstream consumers can attribute any coverage-ratchet
// output to this tool.
const producerName = "coverage-ratchet"

// provenanceEnvelope is the machine-readable provenance metadata attached to
// the --json payload. It carries the three fields every artifact must expose:
// who produced it, which version, and when.
type provenanceEnvelope struct {
	// Producer identifies the tool that emitted the payload.
	Producer string `json:"producer"`
	// Version is the semantic version of the tool: the ldflags-embedded
	// version when set, else semanticVersion()'s build-info fallback. That
	// fallback is narrower than `coverage-ratchet --version` — it does not
	// harvest vcs.revision/vcs.time — so a `go run` invocation reports
	// "dev" with no commit attribution.
	Version string `json:"version"`
	// GeneratedAt is the RFC 3339 UTC timestamp when the payload was produced.
	GeneratedAt string `json:"generatedAt"`
}

// newProvenanceEnvelope returns a fully populated provenance envelope.
// Producer is the producerName constant, Version is derived from
// semanticVersion, and GeneratedAt is the current UTC time formatted as RFC
// 3339.
func newProvenanceEnvelope() *provenanceEnvelope {
	return &provenanceEnvelope{
		Producer:    producerName,
		Version:     semanticVersion(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

---
tag: metric-fidelity
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:46:22Z
identity: metric-fidelity-20261031T234622-jay-flowers
tier: draft
---

When a review-feedback fix populates a previously-zeroed schema field from an external tool's census, the disclosure must propagate to every surface the field feeds — not just the adapter's GoDoc sourcing matrix. In vibe-check PR #75's feedback fixes, the TotalLines population (ripwire physical line counts, config files excluded) updated doc.go, the recorded goldens, and the testdata ledger, but README's 'Metric sources (Python / TypeScript)' table (the user-facing mirror of the sourcing matrix), the warnings stream, and the CHANGELOG Unreleased entry all stayed silent — even though metrics/report sums the field into the report's total-lines aggregate and sorts modules by it, and the denominator (physical lines) differs from the Go adapter's (significant lines). Rule: for any newly populated ModuleResult field, enumerate its consumers (report renderer, diff engine, threshold flags) and its disclosure surfaces (adapter doc.go matrix, README sourcing/caveat tables, warning codes, CHANGELOG entry) and update each; an internal testdata ledger is not a disclosure surface.

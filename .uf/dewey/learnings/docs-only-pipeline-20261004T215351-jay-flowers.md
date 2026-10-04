---
tag: docs-only-pipeline
author: jay-flowers
created_at: 2026-10-04T21:53:51Z
identity: docs-only-pipeline-20261004T215351-jay-flowers
tier: draft
---

For docs-only OpenSpec changes, the pipeline still requires full CI parity (go build, go vet, go test -race, golangci-lint) even though only .md files change. This is constitutional — agents must replicate exactly what CI runs, not what they think is sufficient. Additionally, the CHANGELOG gate applies to docs-only changes: a new reference document is user-facing and deserves a docs: entry in the unreleased section. Multiple Divisor reviewers flagged the missing CHANGELOG task independently. The spec should include a requirement traceable to a task for CHANGELOG updates.

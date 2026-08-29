---
tag: universal-coupling-model
author: jay-flowers
category: gotcha
created_at: 2026-08-29T00:53:36Z
identity: universal-coupling-model-20260829T005336-jay-flowers
tier: draft
---

For the vibe-check project's golangci-lint configuration, version 2 of golangci-lint treats gofmt and goimports as formatters not linters — they must go under formatters.enable, not linters.enable. Also, gosimple was merged into staticcheck in v2, so listing it separately causes an error. The .golangci.yml needs version: "2" at the top level. When implementing a hand-rolled JSON schema validator (metrics/validate.go) to avoid external dependencies, keep complexity low by extracting helper functions (validateModule, validateWarning) — Gaze flagged the main Validate function at complexity 16 which is the only Q4 Dangerous function in the codebase.

---
tag: release-automation
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:03Z
identity: release-automation-20261028T015603-jay-flowers
tier: draft
---

GoReleaser release archives place built binaries at the archive ROOT (named per builds.binary; no bin/ directory, no wrapping unless wrap_in_directory is configured), so extracting a vibe-check release archive yields ./vibe-check — NOT the bin/vibe-check path a from-source CI self-gate produces via `go build -o bin/projectname`. When authoring install or usage steps that switch from build-from-source to released artifacts, re-derive the binary path from .goreleaser.yaml rather than copying the from-source invocation. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — tasks.md 1.2 specced a `bin/vibe-check --version` smoke test while the shipped reusable_structural_gate.yml correctly invokes the archive-root `./vibe-check`; the spec-review advisory caught it before merge. A from-source path in an install script against a release archive fails as file-not-found; the inverse (archive-root path against a go build output) fails identically.

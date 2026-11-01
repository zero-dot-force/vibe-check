---
tag: ci-supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:36Z
identity: ci-supply-chain-20261101T002536-jay-flowers
tier: draft
---

A Go binary installed via `go install module@version` reports its version through debug.ReadBuildInfo but carries no vcs.revision (module-cache builds have no VCS info), so a self-versioning generator distributed via go install cannot derive its own release commit SHA. SHA-pinning an emitted CI `uses:` ref from such a generator therefore requires either ldflags injection (goreleaser/release binaries only) or a network tag→SHA resolution step (git ls-remote refs/tags/<tag>^{}, fail-closed offline). vibe-check's init-scaffolds-structural-gate-caller design D3 chose tag pinning partly for this reason — which puts generated callers in tension with CI-001 / Constitution Principle V and with the README's own consumer guidance (SHA for `uses:`, tag for `version:`), because the design's rationale applied the README's version-input tag language to the `uses:` ref. When a generator emits CI artifacts, reconcile its pinning policy with the convention pack and README explicitly rather than inheriting the tension silently.

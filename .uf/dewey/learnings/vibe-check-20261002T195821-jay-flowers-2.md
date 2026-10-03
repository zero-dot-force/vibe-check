---
tag: vibe-check
author: jay-flowers
category: gotcha
created_at: 2026-10-02T19:58:21Z
identity: vibe-check-20261002T195821-jay-flowers-2
tier: draft
---

`vibe-check analyze` is static analysis but, because it uses go/packages type-aware loading, it COMPILES the target module — which can execute build-time code (cgo, generated files, custom build steps) from the analyzed repository. `GOTOOLCHAIN=local` only prevents downloading a toolchain named by a target go.mod; it does NOT prevent build-time code execution. Consequence: `analyze` MUST NOT be wired into CI that analyzes untrusted fork pull requests (the divisor-entropy agent documents this constraint). The structural-gate CI workflow addresses this with a job-level fork guard: `if: github.event.pull_request.head.repo.full_name == github.repository`, which skips the gate entirely for forked PRs. This is the security justification for re-scoping issue #8 to an opt-in `diff --gate` flag plus a fork-guarded workflow.

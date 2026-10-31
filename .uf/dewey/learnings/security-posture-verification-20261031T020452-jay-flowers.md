---
tag: security-posture-verification
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:04:52Z
identity: security-posture-verification-20261031T020452-jay-flowers
tier: draft
---

Constitution-alignment table claims like "V. Security by Default: PASS — path validation" are only verifiable at code-review time by grepping for the control's call site in the new adapter package. Vibe-check's path-validation control (metrics.ValidateProjectPath) is applied per-adapter, not centrally: internal/goadapter/adapter.go:87, metrics/external.go:180, internal/scaffold/scaffold.go:55 all call it, but the ripwire adapter (2026-10-30) shipped only a Stat+IsDir checkProjectDir while its proposal.md:102 claimed path validation — zero ValidateProjectPath call sites in internal/ripwireadapter/. Review checklist for any new metrics.Adapter implementation claiming goadapter security parity: grep the new package for ValidateProjectPath, SanitizeEnvironment usage, and bounded output capture; a reimplemented weaker check (Stat+IsDir) does not satisfy a traversal-rejection + symlink-resolution claim.

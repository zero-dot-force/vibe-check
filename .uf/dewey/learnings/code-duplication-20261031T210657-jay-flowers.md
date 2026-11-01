---
tag: code-duplication
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:06:57Z
identity: code-duplication-20261031T210657-jay-flowers
tier: draft
---

Security-relevant output-bounding helpers duplicated within one package drift faster than CS-004's two-copy tolerance anticipates. In vibe-check's internal/ripwireadapter, previewVersionOutput (version.go) and previewBytes (runner.go) both implement the truncate-at-cap + "..." + "<empty>" + %q preview of hostile subprocess output, but drifted to caps 120 vs 512, TrimSpace in only one copy, and a hostile-binary GoDoc rationale in only one; both truncate on byte boundaries, so any hardening (rune-safe truncation, cap policy change) must be applied twice. The project's recorded two-copy tolerance (learning cli-architecture-20261030T173307 / issue #73) covers cross-binary operational scaffolding where extraction requires a new internal package; for in-package security controls, extract a single helper parameterized by limit at two copies so hardening propagates to every call site. Tests rarely pin preview constants or normalization, so the consolidation is behavior-safe.

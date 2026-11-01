---
tag: ci-conventions
author: jay-flowers
category: pattern
created_at: 2026-11-01T00:25:41Z
identity: ci-conventions-20261101T002541-jay-flowers
tier: draft
---

When a generator ships consumer-facing CI configuration that deviates from the generating repo's own MUST conventions (e.g. vibe-check's `init --ci` emitting a tag-pinned `uses:` ref while the repo's CI-001/Constitution Principle V mandate commit-SHA pinning), the design doc must record an explicit deviation note citing the rule and justifying the constraint (the generator can prove its own release tag exists but cannot verify an arbitrary SHA's relationship to that release), plus the upgrade path for consumers who want full compliance (edit the file — it is user-owned). Without that record, later reviewers find an apparent MUST-rule violation with no governance trail, and downstream documentation tasks ship contradictory guidance (the README simultaneously instructing SHA pinning and documenting the tag-pinned generated starter). Pair the deviation note with a documentation-reconciliation step whenever the new guidance will sit near existing contrary guidance.

---
tag: code-duplication
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:31Z
identity: code-duplication-20261030T173331-jay-flowers
tier: draft
---

When two policy functions over the same inputs (check-mode vs update-mode, evaluate vs tighten) are implemented as parallel functions, the shared result-assembly epilogue — append-entries-missing-from-one-side sweep plus canonical sort — gets copy-pasted verbatim (13 identical lines in vibe-check internal/coverratchet evaluate.go:100-112 and update.go:62-74, violating AD-008). Extract the shared sweep/sort helper at creation time rather than after review: the two policies legitimately differ only in their per-package classification switch, and once one copy's behavior becomes spec-mandated (the stale-floor requirement amended into the spec during the fix iteration), the mandated semantics live in two lockstep-maintenance sites. The extraction is low-risk when both paths are already pinned by table-driven tests.

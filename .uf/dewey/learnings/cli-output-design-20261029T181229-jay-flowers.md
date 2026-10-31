---
tag: cli-output-design
author: jay-flowers
category: gotcha
created_at: 2026-10-29T18:12:29Z
identity: cli-output-design-20261029T181229-jay-flowers
tier: draft
---

When adding a new output slice to an existing Result struct whose siblings carry an implicit output contract (stable-sorted, normalized to non-nil), the new slice inherits that contract's expectations: either extend the sorting/normalization to the new field or explicitly document its different ordering. In vibe-check's init UF-registration change (PR #65), scaffold.run() sorted Written/Skipped/Forced but left the new Notices slice in generation order, while writeInitSummary's doc comment still claimed the summary lists are "in stable sorted order" — the doc over-generalized onto a list that was neither sorted nor category-prefixed paths. Consumers scripting against the --json payload infer invariants from sibling keys (all other arrays are sorted and never null), so an unsorted additive array is a silent contract divergence even though output stays deterministic.

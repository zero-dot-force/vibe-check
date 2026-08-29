---
tag: universal-coupling-model
author: jay-flowers
category: pattern
created_at: 2026-08-29T00:53:44Z
identity: universal-coupling-model-20260829T005344-jay-flowers
tier: draft
---

The review council consistently flagged zone classification threshold contradictions across spec artifacts — the design doc said thresholds are "configuration, not model" while the metrics-schema spec hardcoded them. The resolution was to accept thresholds as part of the model and update the design non-goal. This is a common spec review pattern: when a non-goal contradicts a spec requirement, the non-goal usually needs refinement rather than the spec. Another recurring theme was LCOM variant ambiguity — the constitution requires citing the measurement model, and LCOM4 (Hitz and Montazeri, 1995) was chosen for its connected-component semantics. Always specify WHICH variant of a well-known metric you're implementing.

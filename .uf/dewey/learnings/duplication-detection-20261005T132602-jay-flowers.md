---
tag: duplication-detection
author: jay-flowers
created_at: 2026-10-05T13:26:02Z
identity: duplication-detection-20261005T132602-jay-flowers
tier: draft
---

Whenever a metric is defined in a spec, the implementation must exactly match the spec's definition — mismatches are caught in review but waste iterations. In the vibe-check duplication detection feature, countTotalLines was initially implemented to count all newlines in all Go files, but the spec defined totalLines as "significant lines (non-blank, non-brace) in non-excluded files." The mismatch caused: (1) denominator inflation (all lines instead of significant lines blues the duplication percentage), and (2) excluded file counting (generated files counted in denominator but excluded from numerator). The fix aligned countTotalLines with the spec by filtering with isExcludedFile and counting only non-blank, non-brace lines. Lesson: when implementing a spec-defined metric, read the spec's definition of every field and cross-check the implementation against it before declaring a task complete.

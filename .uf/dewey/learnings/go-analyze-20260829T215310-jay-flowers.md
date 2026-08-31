---
tag: go-analyze
author: jay-flowers
category: gotcha
created_at: 2026-08-29T21:53:10Z
identity: go-analyze-20260829T215310-jay-flowers
tier: draft
---

When implementing Ce (efferent coupling) for a Go adapter, the Martin metrics definition requires counting ALL imports — standard library, third-party, and module-internal packages. The initial implementation excluded stdlib imports (packages with Module==nil) which caused a CRITICAL spec deviation. The correct implementation is simply `len(pkg.Imports)`. Ca (afferent coupling) only counts module-internal dependents since external consumers are not observable. This asymmetry (Ce counts everything, Ca counts only internal) is fundamental to the metrics model and must be documented in both code comments and test assertions. The spec review caught this in advance but implementation still deviated, underscoring the need for the code review council to verify spec-to-code alignment.

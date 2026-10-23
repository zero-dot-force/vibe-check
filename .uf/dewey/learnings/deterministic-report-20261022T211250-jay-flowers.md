---
tag: deterministic-report
author: jay-flowers
category: decision
created_at: 2026-10-22T21:12:50Z
identity: deterministic-report-20261022T211250-jay-flowers
tier: draft
---

vibe-check deterministic-report task 1.2 (metrics/report.go): the five grade classifiers are ClassifyDistance(avg Distance) Grade, ClassifyLCOM4(maxLCOM4 LCOM) Grade, ClassifyCycles(count int) Grade, ClassifyDuplication(pct float64) Grade, ClassifySpread(count int) Grade. Key decisions for tasks 1.3/1.4 to consume: (1) distance and LCOM4 use the named metric types (Distance, LCOM) as params — aggregates in 1.3 should store AvgDistance as Distance and max LCOM4 as LCOM to call these directly, else explicit conversion is needed; (2) all cascades are descending worst-first switch with >= on the band-minimum constants, so each constant is used exactly once and boundary values land in the lower (worse) grade's band; (3) ClassifyDuplication's B+ case is `pct > 0` (A is the point band at exactly 0%; no DuplicationGradeBPlus constant exists); (4) duplication pct is plain float64 percent in [0,100] — no named type was created (avoided task-1.3 scope creep). Tests: 53 table-driven subtests across TestClassify{Distance,LCOM4,Cycles,Duplication,Spread}_Boundaries using ±0.0001 just-below/above float probes (safe: 1e-4 offset vs ~1e-16 float64 representation error) and exhaustive integer band tables. Errorf format used: "FuncName(%v): got %v, want %v" which satisfies both AGENTS.md "got %v, want %v" and call-context conventions.

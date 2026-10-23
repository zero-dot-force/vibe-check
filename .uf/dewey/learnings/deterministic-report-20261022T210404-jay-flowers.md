---
tag: deterministic-report
author: jay-flowers
category: decision
created_at: 2026-10-22T21:04:04Z
identity: deterministic-report-20261022T210404-jay-flowers
tier: draft
---

vibe-check deterministic-report task 1.1 (metrics/report.go): the five grade threshold tables use INCLUSIVE BAND-MINIMUM constants (e.g. DistanceGradeBPlus = 0.1 means avg >= 0.1 grades B+), which exactly reproduces the spec boundary scenarios (distance exactly 0.1 → B+, LCOM4 exactly 4 → C+, duplication exactly 5% → C+, spread 1 → B, any cycle → F). Two irregularities to carry into task 1.2 classification functions: (1) duplication's A band is a POINT band at exactly 0% with B+ open at zero (0 < dup < 3), so classification needs a `pct > 0 → B+` case and there is no DuplicationGradeBPlus constant; (2) the Grade scale omits D entirely (A, B+, B, C, C... correction: A, B+, B, C+, C, F). Grade.Emoji is a method returning "" for unknown grades. Grades are informational only — never gates; verdict thresholds in metrics/verdict.go remain the sole enforcement.

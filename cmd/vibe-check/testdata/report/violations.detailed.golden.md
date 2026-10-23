# 🏗️ Architectural Health Report

**Provenance**: vibe-check 1.2.3 | 2 packages | 1200 lines

📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md

Needs Attention | 2 packages | Avg Instability: 0.65 | Avg Distance: 0.25 | Max LCOM4: 5 | Cycles: 0

## Prioritized Recommendations

1. 🔴 Split `pkg/monolith` — LCOM4 of `5` suggests multiple unrelated responsibilities
2. 🟡 Decouple `pkg/hub` — `21` efferent couplings; consider interface extraction

## 📊 Package Metrics

| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce | Zone |
|---------|-------------|--------------|----------|-------|----|----|------|
| `pkg/hub` | 0.80 | 0.00 | 0.25 | 1 | 1 | 21 | 🟡 Balanced |
| `pkg/monolith` | 0.50 | 0.00 | 0.25 | 5 | 2 | 2 | 🟡 Balanced |

## 🔗 Coupling Summary

- **Max Ca**: `pkg/monolith` (2 dependents)
- **Max Ce**: `pkg/hub` (21 dependencies)
- **Leaves (Ce=0)**: none
- **Heavily depended-on (Ca > 10)**: none

## 🧩 Cohesion Summary

- **Best**: `pkg/hub` (LCOM4=1)
- **Worst**: `pkg/monolith` (LCOM4=5)
- **Packages needing review (LCOM4 ≥ 4)**: `pkg/monolith` (5)

## 🔄 Circular Dependencies

🟢 No circular dependencies detected.

## 📋 Code Duplication

No duplicated blocks detected.

## 🏥 Health Scorecard

| Dimension | Value | Grade | Status |
|-----------|-------|-------|--------|
| Avg Distance from Main Sequence | 0.25 | B | 🟡 |
| Max LCOM4 | 5 | C | 🔴 |
| Cycle Count | 0 | A | 🟢 |
| Duplication % | 0.0% | A | 🟢 |
| Instability Spread | 0 extremes | A | 🟢 |

**Overall**: 🔴 Needs Attention

## ⚠️ Warnings

No warnings.

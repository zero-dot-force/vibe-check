# 🏗️ Architectural Health Report

**Provenance**: vibe-check 1.2.3 | 2 packages | 300 lines

📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md

All Clear | 2 packages | Avg Instability: 0.35 | Avg Distance: 0.25 | Max LCOM4: 2 | Cycles: 0

## Prioritized Recommendations

1. 🟢 Architecture is healthy — no significant issues detected; keep monitoring

## 📊 Package Metrics

| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce | Zone |
|---------|-------------|--------------|----------|-------|----|----|------|
| `pkg/alpha` | 0.25 | 0.00 | 0.12 | 1 | 1 | 1 | 🟡 Balanced |
| `pkg/beta` | 0.45 | 0.00 | 0.38 | 2 | 2 | 2 | 🟡 Balanced |

## 🔗 Coupling Summary

- **Max Ca**: `pkg/beta` (2 dependents)
- **Max Ce**: `pkg/beta` (2 dependencies)
- **Leaves (Ce=0)**: none
- **Heavily depended-on (Ca > 10)**: none

## 🧩 Cohesion Summary

- **Best**: `pkg/alpha` (LCOM4=1)
- **Worst**: `pkg/beta` (LCOM4=2)
- **Packages needing review (LCOM4 ≥ 4)**: none

## 🔄 Circular Dependencies

🟢 No circular dependencies detected.

## 📋 Code Duplication

No duplicated blocks detected.

## 🏥 Health Scorecard

| Dimension | Value | Grade | Status |
|-----------|-------|-------|--------|
| Avg Distance from Main Sequence | 0.25 | B | 🟡 |
| Max LCOM4 | 2 | B+ | 🟢 |
| Cycle Count | 0 | A | 🟢 |
| Duplication % | 0.0% | A | 🟢 |
| Instability Spread | 0 extremes | A | 🟢 |

**Overall**: 🟡 Mixed

## ⚠️ Warnings

No warnings.

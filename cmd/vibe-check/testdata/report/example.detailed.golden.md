# 🏗️ Architectural Health Report

**Provenance**: vibe-check 1.2.3 | 2 packages | 800 lines

📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md

Needs Attention | 2 packages | Avg Instability: 0.50 | Avg Distance: 0.32 | Max LCOM4: 2 | Cycles: 1

## Prioritized Recommendations

1. 🔴 Break cycle between `pkg/alpha`, `pkg/omega` — circular dependencies prevent independent testing
2. 🔴 Break cycle between `pkg/alpha`, `pkg/omega` — circular dependencies prevent independent testing
3. 🟢 Watch duplication in `x/a.go`, `x/b.go` — `2` blocks at `90`% similarity

## 📊 Package Metrics

| Package | Instability | Abstractness | Distance | LCOM4 | Ca | Ce | Zone |
|---------|-------------|--------------|----------|-------|----|----|------|
| `pkg/alpha` | 0.06 | 0.50 | 0.44 | 1 | 12 | 0 | 🟢 Main Sequence |
| `pkg/omega` | 0.94 | 0.90 | 0.20 | 2 | 0 | 3 | 🔴 Zone of Uselessness |

## 🔗 Coupling Summary

- **Max Ca**: `pkg/alpha` (12 dependents)
- **Max Ce**: `pkg/omega` (3 dependencies)
- **Leaves (Ce=0)**: `pkg/alpha`
- **Heavily depended-on (Ca > 10)**: `pkg/alpha`

## 🧩 Cohesion Summary

- **Best**: `pkg/alpha` (LCOM4=1)
- **Worst**: `pkg/omega` (LCOM4=2)
- **Packages needing review (LCOM4 ≥ 4)**: none

## 🔄 Circular Dependencies

🔴 **1 cycle detected:**

```
pkg/alpha → pkg/omega → pkg/alpha
```

## 📋 Code Duplication

**2 duplicated blocks detected (overall 4.0%):**

| File | Lines | Similarity | Occurrences |
|------|-------|------------|-------------|
| `x/a.go:10-25` | 16 lines | 90% | 1 |
| `x/b.go:40-55` | 16 lines | 90% | 1 |

## 🏥 Health Scorecard

| Dimension | Value | Grade | Status |
|-----------|-------|-------|--------|
| Avg Distance from Main Sequence | 0.32 | C+ | 🟡 |
| Max LCOM4 | 2 | B+ | 🟢 |
| Cycle Count | 1 | F | 🔴 |
| Duplication % | 4.0% | B | 🟡 |
| Instability Spread | 0 extremes | A | 🟢 |

**Overall**: 🔴 Needs Attention

## ⚠️ Warnings

- `pkg/gone`: package pkg/gone: type information unavailable
- `pkg/omega` is in the Zone of Uselessness

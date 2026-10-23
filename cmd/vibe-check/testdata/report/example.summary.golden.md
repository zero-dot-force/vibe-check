# 🏗️ Architectural Health Report

**Provenance**: vibe-check 1.2.3 | 2 packages | 800 lines

📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md

Needs Attention | 2 packages | Avg Instability: 0.50 | Avg Distance: 0.32 | Max LCOM4: 2 | Cycles: 1

## Prioritized Recommendations

1. 🔴 Break cycle between `pkg/alpha`, `pkg/omega` — circular dependencies prevent independent testing
2. 🔴 Break cycle between `pkg/alpha`, `pkg/omega` — circular dependencies prevent independent testing
3. 🟢 Watch duplication in `x/a.go`, `x/b.go` — `2` blocks at `90`% similarity

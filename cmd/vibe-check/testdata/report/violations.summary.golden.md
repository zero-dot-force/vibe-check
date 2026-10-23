# 🏗️ Architectural Health Report

**Provenance**: vibe-check 1.2.3 | 2 packages | 1200 lines

📖 Instability = ratio of outgoing to total dependencies (0.0=max stable, 1.0=max unstable) | Abstractness = ratio of abstracts to total types (0.0=fully concrete, 1.0=pure interfaces) | Distance = how far from ideal balance (0.0=perfect, 1.0=worst) | LCOM4 = cohesion (1=best, ≥4 suggests split) | Ca = packages depending on this one | Ce = packages this one depends on | Full guide: https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md

Needs Attention | 2 packages | Avg Instability: 0.65 | Avg Distance: 0.25 | Max LCOM4: 5 | Cycles: 0

## Prioritized Recommendations

1. 🔴 Split `pkg/monolith` — LCOM4 of `5` suggests multiple unrelated responsibilities
2. 🟡 Decouple `pkg/hub` — `21` efferent couplings; consider interface extraction

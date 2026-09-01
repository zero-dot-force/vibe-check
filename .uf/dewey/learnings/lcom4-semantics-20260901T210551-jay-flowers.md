---
tag: lcom4-semantics
author: jay-flowers
category: gotcha
created_at: 2026-09-01T21:05:51Z
identity: lcom4-semantics-20260901T210551-jay-flowers
tier: draft
---

LCOM4 in vibe-check is an integer counting connected components in a package's method-field graph (Hitz and Montazeri 1995). LCOM4=1 means maximally cohesive; higher values mean the package can be split. The existing CLI flag is --max-lcom (upper bound, integer). Never propose a --min-cohesion flag with a float threshold (0.0-1.0) as this represents a completely different metric domain. When writing specs or convention pack rules about cohesion, always use --max-lcom=N with an integer threshold. The AD-009 rule initially proposed --min-cohesion=0.5 and was corrected to --max-lcom=3 during spec review. This semantic inversion was the highest-convergence finding across all three review agents.

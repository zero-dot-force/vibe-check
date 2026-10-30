---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T17:33:19Z
identity: coverage-ratchet-20261030T173319-jay-flowers
tier: draft
---

Go test prints per-package coverage rounded to nearest one decimal (e.g. 89.3%), but a statement-weighted gate that computes covered/total and truncates to one decimal derives 89.2% from the same coverage profile. Quoting go test's display values as floor targets in specs or tasks makes the gate's first at-floor run fail by 0.1pp. Derive committed floors from the gate tool's own truncated output (its --update mode), and when task text quotes display values, record the display-vs-truncated distinction in a deviation note — as vibe-check's ci-coverage-ratchet-floors change did in its tasks.md 3.1 note (89.3/87.9 display vs 89.2/87.8 truncated for internal/goadapter and internal/scaffold).

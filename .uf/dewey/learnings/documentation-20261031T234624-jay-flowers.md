---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:46:24Z
identity: documentation-20261031T234624-jay-flowers
tier: draft
---

A documentation-consistency commit that corrects a version or constant string in user-facing text (e.g. ModuleGraph schema 1.2→1.3 in CLI help) must sweep the whole repository for the stale literal before declaring the docs consistent. In vibe-check PR #75's fixes, commit 8675c79 corrected analyze.go/main.go to schema 1.3 (matching SchemaVersionCurrent in metrics/graph.go) and edited AGENTS.md — but AGENTS.md still carried three stale v1.2 references (project-structure listing, validate.go description, architecture section), leaving the user-facing help and the agent-facing project doc contradicting each other on the output contract. Rule: after any version/constant correction in prose, grep the repo (AGENTS.md, README, docs, spec directories) for the old literal — the stale references are usually in the very file just edited for a different reason, and the drift often predates the current change (here, since the v0.2.0 schema bump).

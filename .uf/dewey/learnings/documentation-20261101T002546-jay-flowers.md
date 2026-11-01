---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:25:46Z
identity: documentation-20261101T002546-jay-flowers
tier: draft
---

Documentation commitments embedded in design-doc prose outside the Decisions and What Changes surfaces — Non-Goals bullets, risk-mitigation notes, rejected-alternative remarks — do not self-execute into tasks.md, because the task list is derived from What Changes and the Decisions: a phrase like "noted as a known limitation in the README" in a Non-Goals bullet lands nowhere. When authoring or spec-reviewing, grep design.md for documentation-deliverable verbs ("noted in the README", "documented in", "the README documents") and verify each has a corresponding task; separately, derive the AGENTS.md task scope from the documentation gate's named surfaces (project structure, conventions, build commands) rather than only from a rename-reference list. Canonical case: vibe-check opsx/init-scaffolds-structural-gate-caller (2026-10-31): design.md Non-Goals promises the root-only manifest assumption is "noted as a known limitation in the README", but tasks 5.1/4.2 and the proposal's documentation enumeration omit it (the README's Known limitations section has no such entry), and task 4.2's AGENTS.md scope covers only the structure listing, leaving the Build & Test Commands init flag comment and the Layer-3 init description stale despite the init-registers-divisor-entropy precedent (its task 5.3 updated all three surfaces).

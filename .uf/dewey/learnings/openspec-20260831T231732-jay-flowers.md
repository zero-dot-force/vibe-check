---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-08-31T23:17:32Z
identity: openspec-20260831T231732-jay-flowers
tier: draft
---

OpenSpec validation gotcha (vibe-check / openspec CLI): `openspec validate <change> --strict` requires that each ADDED requirement's STATEMENT — the prose between the `### Requirement: <name>` header and its first `#### Scenario:` — LEAD with a normative RFC-2119 clause containing SHALL or MUST. A requirement whose statement opens with a rationale sentence (e.g. "Because vibe-check analyze executes the target's build tooling...") FAILS strict validation with an error like "requirement '<name>' must contain SHALL or MUST", even when MUST/SHALL keywords appear later in the same paragraph. FIX: reorder the paragraph so the first sentence is the normative statement (e.g. "The divisor-entropy agent SHALL only run on refs the CI context already trusts and MUST NOT be wired into CI that analyzes untrusted fork pull requests."), then follow with the rationale. Encountered while authoring the 'Trusted-Refs-Only Operating Constraint' requirement in specs/divisor-entropy-agent/spec.md for the add-divisor-entropy-agent change.

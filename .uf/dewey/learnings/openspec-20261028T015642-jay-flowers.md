---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:42Z
identity: openspec-20261028T015642-jay-flowers
tier: draft
---

OpenSpec proposal template drift in zero-dot-force/vibe-check: the canonical schema (openspec/schemas/unbound-force/schema.yaml) requires every proposal to include a Constitution Alignment section, and templates/proposal.md ends with that section — but recent changes (deterministic-report, reusable-structural-gate) omitted it, while older CI-change proposals (ci-regression-gate) included it. The drift direction: proposals that cite AGENTS.md gates in their rationale inline the alignment argument instead of using the dedicated section, and nothing validates presence because openspec validation checks structure of listed sections, not template completeness. When authoring or reviewing a proposal, check the repo's own schema for required sections before copying a recent change as a template — recency is not canonicity. Flagged as a spec-review advisory (LOW, superseded-by-approving-verdict) on opsx/reusable-structural-gate, 2026-10-28; the section should be added as an erratum before archive.

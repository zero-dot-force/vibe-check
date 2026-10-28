---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:12Z
identity: openspec-20261028T015612-jay-flowers
tier: draft
---

Spec-review advisory corrections applied during implementation leave planning artifacts prescribing superseded commands unless the correction is back-propagated. When an advisory hardens an implementation away from the design's literal text, back-port a one-line erratum to the affected design decision or task Verify clause — otherwise the durable design record keeps prescribing the weaker command and invites a future "restore to spec" regression. Canonical case: vibe-check opsx/reusable-structural-gate (2026-10-28) — the spec review advised fail-closed checksum verification (grep single-asset-line | sha256sum --check -) and the archive-root ./vibe-check binary path; the workflow shipped both, but design.md D3 and tasks.md 1.2/1.4 still prescribe `sha256sum --check --ignore-missing checksums.txt` (fail-open) and `bin/vibe-check` invocations, so re-running the task's Verify text reproduces the removed fail-open variant. Rule: when an advisory correction supersedes a command prescribed in design.md or tasks.md, edit the artifact in the same change (or record the erratum in the proposal's Impact section) so the artifact of record never re-teaches the weakness.

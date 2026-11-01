---
tag: structural-entropy
author: jay-flowers
category: reference
created_at: 2026-11-01T14:32:42Z
identity: structural-entropy-20261101T143242-jay-flowers
tier: draft
---

vibe-check diff emits per-module DELTAS in both output modes (--json and the default human table, "Per-module deltas (PR minus base)"); neither carries base→PR absolutes. The entropy divisor's report format needs base → PR (Δ): materialize both refs' graphs with analyze --output, then read the changed modules' absolute values from the two snapshot JSONs (grep the module path for the line offset, then read the ~20-line module object). The diff --json "unreliable" field is the authoritative reliability verdict — it is false only when both graphs are complete — so a separate status grep of the analyze snapshots is unnecessary. Verified on vibe-check PR #75 round 3 (base 814053c → head 57d02b8): one --json diff run plus two targeted snapshot reads produced the full auditable table, and the deltas matched round 2's measurement exactly, confirming doc/test-only commits move no Martin metric.

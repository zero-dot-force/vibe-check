---
tag: integration-design
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:57Z
identity: integration-design-20261101T172957-jay-flowers
tier: draft
---

When wrapping an external tool that exposes paginated listings, requesting one fixed-window page and warning on has_more converts every large corpus into silently incomplete data. In vibe-check PR #75, the ripwire adapter issues --deps once with --limit=1000; corpora beyond 1000 dependency-capable files lose the edges (and cycles) of every unlisted file, the parser already captures next_offset but never follows it, and the truncation warning cannot change the output's overall status because ripwire graphs are always partial for other documented reasons — so status-keyed CI gates cannot see the difference between a truncated corpus and baseline approximations. Rule: either paginate to exhaustion (bounded loop over next_offset) or expose the window as configuration; a truncation warning that never alters any observable gate signal is a disclosure, not a control.

---
tag: golden-fixtures
author: jay-flowers
category: pattern
created_at: 2026-10-31T02:04:56Z
identity: golden-fixtures-20261031T020456-jay-flowers
tier: draft
---

When pinning golden fixtures against an externally pinned tool version, commit a value-by-value computation ledger alongside the goldens (vibe-check's internal/ripwireadapter/testdata/recorded/golden-notes.md is the model): for every module and metric, name the source (recorded number, fixture fact, or formula), and record every convention decision where the design left the choice open — the golden's choice then becomes the pinned contract. When the pinned tool version moves, re-recording and re-deriving goldens becomes mechanical instead of forensic.

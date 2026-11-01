---
tag: documentation-accuracy
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:18:59Z
identity: documentation-accuracy-20261101T001859-jay-flowers
tier: draft
---

A cobra threshold flag's registered default value is inert when the command gates application on Flags().Changed(): the registered value only surfaces in --help text and never becomes the effective threshold. Documentation must state the effective default, not the registered one. In vibe-check PR #75, the README --max-duplication row landed with Default `5.0` (the value registered via Float64VarP in analyze.go) while every sibling threshold row documents `unset (no check)` — the effective default under Changed() gating — so readers conclude a duplication gate is active by default when no check runs unless the flag is passed. Rule: when documenting a flag behind Changed() gating, derive the Default column from what the command does with a zero-value Options struct, and keep every row in the table on that same semantic.

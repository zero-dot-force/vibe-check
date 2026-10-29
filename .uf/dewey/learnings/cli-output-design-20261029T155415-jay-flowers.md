---
tag: cli-output-design
author: jay-flowers
category: decision
created_at: 2026-10-29T15:54:15Z
identity: cli-output-design-20261029T155415-jay-flowers
tier: draft
---

Distinguish "skipped because present and left untouched" from "skipped because absent" in machine-readable command output: an absent file must never land in a `skipped` array whose contract promises "file existed and was left byte-identical", because scripting consumers misread absence as presence. Route absence and unrecognized-shape outcomes through a separate notices channel (human section printed only when non-empty; JSON array normalized to [] and never null) and keep existing keys' semantics stable when adding new outcome categories. Canonical case: vibe-check init UF policy registration (design D5) — absent .uf/ policy files produce notices, never skipped entries, and the additive `notices` JSON key is always an array.

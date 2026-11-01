---
tag: ripwire-adapter
author: jay-flowers
category: pattern
created_at: 2026-10-31T20:09:43Z
identity: ripwire-adapter-20261031T200943-jay-flowers
tier: draft
---

When a shipped wiring substitutes a control's documented input, assess both drift directions AND whether other warnings only incidentally mask the gap. vibe-check PR #75 (ripwire-adapter): the zero-coupling guard was documented to receive countSourceFiles (a filesystem source-file census) but assembly passes len(ccFiles) (the cc.json export census). Fixture-verified drift: the TypeScript corpus's census lists 5 File nodes for 3 source files (package.json/tsconfig.json are indexed as File nodes), and the census omits denylist-pruned subtrees (operations/build/ absent from the Python fixture's cc.json) that the documented census deliberately includes. The over-count direction is safe (fires the guard more often); the under-count direction could disengage the guard for a corpus whose sources all live under denylisted directories — harm currently bounded only because the denylist-pruned warning independently forces partial status, an incidental intersection of two warning codes rather than a designed guarantee. Rule: an input substitution is acceptable only when the GoDoc contract, the user-facing warning text, and the failure-mode coverage are all reconciled deliberately; masking inherited by accident is undocumented coupling between warnings and silently disappears if the other warning is reordered, dropped, or scoped away.

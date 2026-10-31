---
tag: zero-waste
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:05:01Z
identity: zero-waste-20261031T020501-jay-flowers
tier: draft
---

Task-scaffolding code and comments go stale when a later assembly task supersedes them. In vibe-check's ripwire adapter (openspec/changes/ripwire-adapter), the zero-coupling guard shipped with a dedicated source-counter (countSourceFiles) and a "Wiring point: nothing calls the guard yet" comment; the later assembly task wired the guard with a different input (len(ccFiles)) and left the helper orphaned, the comment false, and both GoDocs naming the wrong supplier. The same pattern left a dead --exclude loop in the crawl scope, an uncalled version-pin wrapper (checkRipwirePin), unreferenced cycle-cap constants whose values were hardcoded into the warning message, and test-only cc.json helpers in production files. Sweep at the final assembly task: grep every helper's callers before marking the last task complete, delete "not yet wired" comments when the wiring lands, and reconcile every GoDoc that names a supplier or caller. Extra signal: when the substituted input has different semantics (here, a post-prune census vs a deliberately unfiltered filesystem walk), the stale helper is not just dead code — the wiring changed behavior the helper's GoDoc promised.

---
tag: safety-net-design
author: jay-flowers
category: pattern
created_at: 2026-11-01T00:19:11Z
identity: safety-net-design-20261101T001911-jay-flowers
tier: draft
---

A safety-net check must not consume the output of the component it guards. In vibe-check's ripwire adapter (PR #75), the zero-coupling guard originally took its source-file count from len(ccFiles) — the crawl's own cc.json census — so a crawl that failed to index anything presented an empty corpus and silently disengaged the guard exactly when it was needed. The remediation rewired the guard to countSourceFiles: the adapter's own recursive filesystem walk over the crawl root, a count independent of the guarded tool's success. The walk deliberately skips the crawl denylist (sources under build/ or vendor/ still count) because over-counting can only make a warning-emitting safety net fire more often — the safe direction — and that over-count posture is documented in the helper's GoDoc. Pattern: when a guard's firing condition is "the tool produced nothing over a real corpus," derive the corpus evidence from a channel the tool does not control, and document the over-count direction as intentional.

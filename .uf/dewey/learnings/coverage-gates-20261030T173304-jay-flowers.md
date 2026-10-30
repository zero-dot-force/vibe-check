---
tag: coverage-gates
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:04Z
identity: coverage-gates-20261030T173304-jay-flowers
tier: draft
---

Fail-closed design for tighten-only coverage gates (vibe-check internal/coverratchet): three structural controls make the ratchet non-bypassable without relying on reviewer vigilance. (1) Floor domain validation rejects any floor outside (0, 100] or finer than one decimal — a floor of 0 would silently disable the gate for its package, so it is treated as malformed input (exit 2), not accepted config. (2) Stale-floor violations fail the gate when a baseline entry has no measurement, so an emptied or narrowed coverage profile cannot pass vacuously — the baseline must mirror the module's coverable packages exactly. (3) Update mode is all-or-nothing: any violation returns the input baseline unchanged and the caller only writes when the result is OK, so a failed policy state never mutates the gate's own configuration. Complementary property: floors and measurements are truncated to one decimal via the same exact-integer arithmetic (covered*1000/total, then /10), so measured == floor compares equal without epsilon and a recorded floor can never exceed what the producing run measured.

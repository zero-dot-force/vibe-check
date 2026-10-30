---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T17:33:28Z
identity: coverage-ratchet-20261030T173328-jay-flowers
tier: draft
---

A self-gating coverage ratchet opens silent floor headroom when a fix iteration adds tests after the baseline was generated: floors only fail downward, so the gate stays green while a recorded floor drifts below the tree's measured coverage — visible only as an inaccuracy in the durable record (canonical case: vibe-check opsx/ci-coverage-ratchet-floors, internal/coverratchet floor 96.0 vs 97.2 measured after round-1 error-path tests landed post-baseline, while the CHANGELOG claimed only metrics deviated from measured coverage). Reconcile every floor against the final tree's measurement before submission. Caution when fixing: a blanket `--update` raises ALL floors to measured, including deliberately documented accommodation floors (e.g., a variance headroom floor like metrics 82.0 for flaky subprocess coverage) — clobbering them reintroduces the intermittent failure the accommodation prevents, so a targeted reviewed baseline edit is the right tool for single-package refreshes.

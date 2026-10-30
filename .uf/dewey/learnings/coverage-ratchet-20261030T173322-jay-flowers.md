---
tag: coverage-ratchet
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:22Z
identity: coverage-ratchet-20261030T173322-jay-flowers
tier: draft
---

Tighten-only coverage-ratchet gates flap when a package's measured coverage is non-deterministic run-to-run (vibe-check's metrics package varies ~82.1–82.9% under -race because its external-adapter subprocess tests race). The working accommodation pattern from the ci-coverage-ratchet-floors change: (1) record the floor below the observed variance band via a direct, human-reviewed baseline edit (82.0, not the typical 82.7–82.9 measurement); (2) document the accommodation durably in CHANGELOG with the variance band and rationale, not just in the review thread; (3) track the test-stabilization work and the follow-up floor re-raise in an issue (#71); (4) keep the ratchet strictly tighten-only so the accommodation can only be closed by re-raising, never silently widened. A floor set at typical measured coverage would fail the gate intermittently and train the team to ignore red CI.

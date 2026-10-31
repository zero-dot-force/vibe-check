---
tag: test-doubles
author: jay-flowers
category: pattern
created_at: 2026-10-31T02:05:03Z
identity: test-doubles-20261031T020503-jay-flowers
tier: draft
---

To keep a hermetic subprocess test double working under environment sanitization, bake its control variables into wrapper scripts placed on PATH rather than passing them through the environment. vibe-check's fake-riprewire harness (internal/ripwireadapter/fake_ripwire_test.go) works this way: tests write executable wrapper scripts that set and export FAKE_RIPWIRE_RECORDINGS / FAKE_RIPWIRE_VARIANT themselves, because the production runner sanitizes the subprocess environment (PATH/HOME/LANG only) and cannot carry FAKE_* vars through; the fake binary reads its own control vars, replays recorded golden outputs selected by flag, and fails closed when FAKE_RIPWIRE_RECORDINGS is unset. Production code never names the fake, the sanitized-runner security property is asserted in tests (no FAKE_RIPWIRE_* passthrough, version_test.go), and CI runs the full subprocess path without the real binary.

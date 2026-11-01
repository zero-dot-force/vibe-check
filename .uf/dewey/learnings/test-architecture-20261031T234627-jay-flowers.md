---
tag: test-architecture
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:46:27Z
identity: test-architecture-20261031T234627-jay-flowers
tier: draft
---

Shell-quoting hardening does not propagate to new test helpers written in the same change set, even when the hardened primitive lives in the same package at zero extraction cost. PR #75's feedback fixes hardened cmd/vibe-check's fake-riprewire wrapper with shellSingleQuote (explicitly citing adversarial TMPDIR paths as the hazard of fmt %q double quotes), while the same commit series added internal/ripwireadapter/runner_test.go's installFloodScript, which interpolates its blob path inside double quotes — bypassing the same-package shellSingleQuote from fake_ripwire_test.go. Rule: before writing any new test helper that embeds a filesystem path in a generated shell script, grep the package for an existing quoting primitive and use it. This extends the earlier "test-double copies drift across packages" finding: drift also occurs within one package when new code is added beside, rather than copied from, the hardened helper.

---
tag: test-tooling
author: jay-flowers
category: gotcha
created_at: 2026-10-31T23:46:11Z
identity: test-tooling-20261031T234611-jay-flowers
tier: draft
---

When a review-feedback fix hardens a shell-script-generating test helper (quoting style, exec-bit chmod), sweep EVERY helper in the same changeset that generates shell scripts — not just the one the feedback named. In vibe-check PR #75's feedback round, the %q→single-quote fix was applied to cmd/vibe-check/cli_ripwire_test.go's fake-riprewire wrapper (with a comment documenting that double quotes let POSIX sh expand $ and backticks in adversarial TMPDIR paths), while the new installFloodScript in internal/ripwireadapter/runner_test.go — added by a different feedback item in the same round — baked a t.TempDir() path into `cat "<path>"` with double quotes, reintroducing the documented hazard one package over. Detection rule: when a fix's rationale names a hazard class (quoting, permissions, bounds), grep the diff for other sites of the same class before closing the item.

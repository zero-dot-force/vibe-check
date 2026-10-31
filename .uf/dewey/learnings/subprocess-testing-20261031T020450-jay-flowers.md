---
tag: subprocess-testing
author: jay-flowers
category: pattern
created_at: 2026-10-31T02:04:50Z
identity: subprocess-testing-20261031T020450-jay-flowers
tier: draft
---

When production code sanitizes subprocess environments (e.g. metrics.SanitizeEnvironment allowlists PATH/HOME/LANG only), test doubles controlled via environment variables break — the sanitized env strips FAKE_RIPWIRE_RECORDINGS-style control vars before the subprocess runs. Working pattern (vibe-check cli_ripwire_test.go, 2026-10-30): write a wrapper shell script into t.TempDir(), prepend that dir to PATH in the exec environment, and have the wrapper script itself export the control variables before exec-ing the actual test double. The wrapper inherits the sanitized allowlisted env (PATH/HOME/LANG), then adds its own vars post-sanitization inside the script body — no production sanitizer needs loosening.

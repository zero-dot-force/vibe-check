---
tag: test-doubles
author: jay-flowers
category: gotcha
created_at: 2026-10-31T20:51:20Z
identity: test-doubles-20261031T205120-jay-flowers
tier: draft
---

A PATH-installed poison/sentinel test double that loses its executable bit (os.WriteFile's mode masked by an unusual umask, e.g. 0177) is silently skipped by exec.LookPath, so the negative assertion it guards — "the code under test never invoked the double" — passes vacuously: the test stops proving anything without failing. Positive doubles fail loudly in the same situation (the code under test reports the binary absent, e.g. ErrRipwireAbsent). When installing sentinel executables in tests, pin the mode with an explicit os.Chmod(path, 0o755) after os.WriteFile so the sentinel cannot silently vanish from PATH. Example: cmd/vibe-check/cli_ripwire_test.go installPoisonRipwire in PR #75 (feedback item 5) — the in-package internal/ripwireadapter harness already Chmods and documents the umask rationale.

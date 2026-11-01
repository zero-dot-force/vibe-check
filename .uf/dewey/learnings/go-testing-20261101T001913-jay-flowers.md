---
tag: go-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T00:19:13Z
identity: go-testing-20261101T001913-jay-flowers
tier: draft
---

PATH-shim test helpers that write an executable script must Chmod explicitly after WriteFile: WriteFile's mode argument is masked by the process umask, so a restrictive umask can strip the exec bit, and exec.LookPath silently skips non-executable files — turning a poison-binary negative test into a vacuous pass (the "never invoked" assertion passes because the shim is invisible to LookPath, not because the code under test is correct). Found in vibe-check PR #75 round 2 (cmd/vibe-check/cli_ripwire_test.go): installPoisonRipwire and installFakeRipwireWrapper both add a post-write os.Chmod(script, 0o755) with a comment naming the vacuous-pass failure mode. Same rule applies to the script contents: build shell wrappers with POSIX single-quoting (never fmt %q, whose double quotes let sh expand $ and backticks in adversarial TMPDIR paths).

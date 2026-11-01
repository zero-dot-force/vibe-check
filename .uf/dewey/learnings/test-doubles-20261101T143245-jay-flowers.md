---
tag: test-doubles
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:45Z
identity: test-doubles-20261101T143245-jay-flowers
tier: draft
---

os.WriteFile's perm argument is masked by the process umask, so an executable test double written with 0o755 can silently lose its exec bit under a restrictive umask; exec.LookPath then skips the non-executable file, the intended double never runs, and PATH-based assertions (especially negative ones like "a poison ripwire is never invoked") pass vacuously — the test proves nothing. Fix pattern from vibe-check PR #75 round 3 (cmd/vibe-check/cli_ripwire_test.go installPoisonRipwire/installFakeRipwireWrapper, internal/ripwireadapter/runner_test.go installFloodScript): follow every WriteFile of an executable test double with an explicit os.Chmod(path, 0o755) so the exec bit survives any umask, and bake control values into wrapper scripts with a POSIX single-quote helper — never fmt %q, whose double quotes let sh expand $ and backticks in adversarial checkout or TMPDIR paths.

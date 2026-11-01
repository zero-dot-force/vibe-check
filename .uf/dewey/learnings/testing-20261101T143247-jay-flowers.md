---
tag: testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:47Z
identity: testing-20261101T143247-jay-flowers
tier: draft
---

os.WriteFile's perm argument is masked by the process umask: a test fixture written with 0o755 can silently land non-executable under a restrictive umask. For fake binaries installed on PATH, exec.LookPath then skips the file entirely and PATH-based negative assertions (poison binary never invoked, adapter-selection checks) pass vacuously — the test proves nothing while staying green. Follow os.WriteFile with an explicit os.Chmod(path, 0o755) for every executable test fixture, and document the vacuous-pass rationale in the comment (pattern applied in vibe-check PR #75 cli_ripwire_test.go and runner_test.go).

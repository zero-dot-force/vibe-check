---
tag: testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T14:32:34Z
identity: testing-20261101T143234-jay-flowers
tier: draft
---

os.WriteFile(path, data, 0o755) does not guarantee the executable bit — the mode is ANDed with the process umask, so a restrictive umask strips exec permission from a test-double script. exec.LookPath silently skips non-executable files, which turns a negative test ("the poison binary must never run") into a vacuous pass: the poison is never resolved, so it trivially never runs, and the guard-under-test is never exercised. Always follow WriteFile of an executable test double with an explicit os.Chmod(path, 0o755) so the assertion tests the guard rather than the ambient umask. Encoded in vibe-check PR #75 (installPoisonRipwire, installFakeRipwireWrapper, installFloodScript all gained post-write Chmod with rationale comments).

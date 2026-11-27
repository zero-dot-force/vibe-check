---
tag: subprocess-testing
author: jay-flowers
category: pattern
created_at: 2026-11-26T20:31:57Z
identity: subprocess-testing-20261126T203157-jay-flowers
tier: draft
as_of: "vibe-check@40a2e0eb4a5313ddb15330ed799f48435eca3ad7"
---

To regression-test a subprocess runner that must return in bounded time when a grandchild process keeps its output pipes open (vibe-check opsx/runner-pipe-hang, PR #113): (1) make the pipe holder a grandchild — a /bin/sh probe script that backgrounds `sleep 60 &`, writes the child's PID to a file for t.Cleanup reaping, then floods or exits — so killing the direct process never releases the pipes and the old code hangs deterministically instead of intermittently (the pre-fix flake depended on whether the writer fit its blob into the 64 KiB pipe buffer before the kill); size the flood blob well above the pipe buffer (1 MiB) so the writer can never finish on its own. (2) Bound every run() call with a helper that executes it in a goroutine and t.Fatalf's by name at a test-level deadline (20s), far under the package timeout, so a regression is reported as a named hang. (3) Shrink the production wait delay to ~200ms via a test-only field so each test pays milliseconds, and assert return time against a slack bound (5s) distinct from the hang deadline. (4) Pin error-classification precedence with co-occurrence cases (cancellation + oversize, oversize + abandoned pipes), and pin the zero-value constructor fallback with a lower-bound elapsed assertion (>= default delay) so a silently dropped fallback fails on timing, not just on hanging. Residual flake vector to design around: a precedence test between an async goroutine trigger (cancel at +300ms) and the runtime WaitDelay timer (2s window) holds only while the goroutine is scheduled inside the window — scheduler starvation past the window flips the observed error class — so size the trigger offset far inside the window and document the starvation assumption in the test comment.

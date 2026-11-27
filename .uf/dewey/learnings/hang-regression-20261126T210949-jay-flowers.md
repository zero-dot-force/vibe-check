---
tag: hang-regression
author: jay-flowers
category: pattern
created_at: 2026-11-26T21:09:49Z
identity: hang-regression-20261126T210949-jay-flowers
tier: draft
as_of: "vibe-check@40a2e0eb4a5313ddb15330ed799f48435eca3ad7"
---

To regression-test an intermittent subprocess hang deterministically: (1) make the hang guaranteed — the fixture script backgrounds a pipe-holding child (sleep 60 &; echo $! > pidfile; wait) so EOF can never arrive while the direct child remains killable; (2) bound the test — run the call in a goroutine and select the result against time.After(runDeadline) so a reintroduced hang fails the test by name in seconds instead of consuming the package timeout; (3) write the child's PID to a file before the failure mechanism fires (before the flood starts) so the oversize kill can never destroy the trigger, and kill the child in t.Cleanup so nothing outlives the test; (4) pin co-occurring failure precedence with a wide wait-delay window and a cancellation triggered a fixed interval after the PID appears — the shared classification check (context error inspected after Wait returns) then decides the outcome on every mechanism interleaving.

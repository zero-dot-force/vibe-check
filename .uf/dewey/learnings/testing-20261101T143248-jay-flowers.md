---
tag: testing
author: jay-flowers
category: pattern
created_at: 2026-11-01T14:32:48Z
identity: testing-20261101T143248-jay-flowers
tier: draft
---

When asserting per-item expected values through a map lookup whose missing-key result is the zero value, pair the assertion loop with (a) a count check (len(got) == len(want)) and (b) a consumed-key sweep asserting every expected key matched an item. Without the sweep, a typo'd or stale expected key silently reads as 0 and collides with a legitimate expected-0 entry — the assertion passes while pinning nothing. Applied in vibe-check PR #75 assembly_test.go wantLines (round-3 advisory D): the TS root module genuinely expects totalLines 0, so a key mismatch would otherwise be indistinguishable from the correct expectation.

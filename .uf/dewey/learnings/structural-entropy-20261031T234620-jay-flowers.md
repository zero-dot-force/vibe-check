---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-10-31T23:46:20Z
identity: structural-entropy-20261031T234620-jay-flowers
tier: draft
---

Entropy review of a fix round layered on an already-reviewed PR head: measure the structural delta against the PR's merge-base (the dispatch input_context base_sha, matching the council's three-dot git diff scope), not against the previously reviewed head — the Go verdict is cumulative for the whole PR, so a structurally neutral fix diff can still inherit a COMMENT verdict from earlier commits. Verified on vibe-check PR #75 round 2 (base 814053c → head 8675c79, fix diff aeca31c..8675c79): the feedback-fix commits (test/doc/testdata additions plus intra-package refactors with no new cross-package repo imports; _test.go files do not enter package coupling) contributed zero metric movement, while all non-zero deltas (metrics Ca 4→5, I −0.0395, D +0.0395; cmd/vibe-check Ce 22→23) trace to the original adapter commits. Because findings confinement applies to the fix diff's changed lines, inherited sub-threshold drift belongs in the prose advisory and verdict explanation, not in a new structured finding against pre-existing scope.

---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-01T00:19:03Z
identity: content-pipeline-20261101T001903-jay-flowers
tier: draft
---

An address-feedback round's claim that remediation issues "were filed" is untrusted caller data until viewed: the diff carries no evidence of tracker state. Verify each claimed filing with `gh issue view <n> --repo <owner>/<repo> --json number,title,state,body` and require state OPEN plus a substantive on-topic body; in label-less repos the title prefix (blog:/tutorial:/docs:) is the tracking convention, so an empty labels array is not a defect. Canonical case: vibe-check PR #75 round 2 (2026-10-31) — the round-1 curator REQUEST CHANGES stood on three missing tracking issues (website-sync HIGH, blog MEDIUM, tutorial MEDIUM); the round-2 context claimed unbound-force/website#353, vibe-check#80, and vibe-check#81 were filed. gh issue view confirmed all three OPEN with on-topic bodies (plus the pre-existing repo-local docs issue #70 satisfying the dual-gate convention), converting the findings into verified resolutions and unblocking the push gate. The tracker, not the review context, is the ground truth for issue-existence findings; approving on the claim alone would make the remediation gate decorative.

---
tag: learning-systems
author: jay-flowers
category: gotcha
created_at: 2026-10-27T17:03:24Z
identity: learning-systems-20261027T170324-jay-flowers
tier: draft
---

Correcting a repo-mirrored Dewey learning file (.uf/dewey/learnings/<identity>.md) does not update the Dewey learning-store copy that dewey_semantic_search serves — the store entry keeps its originally indexed content and the two copies of the same identity diverge. Observed in vibe-check (2026-10-27, council iteration 2 of fix/release-workflow-reusable-refs): the opsxgoreleaser-release-automation-20261004T141907-jay-flowers learning was corrected in the repo mirror to state underscore org-infra workflow filenames, but the store copy still asserted the wrong hyphenated claim and ranked #3 (0.83 similarity) in a semantic search on the topic. When remediating a wrong stored learning: (1) correct the repo mirror, (2) store a superseding or re-stored corrected learning in the store under the same topic so search surfaces the correction, and (3) verify with dewey_semantic_search that the stale claim no longer appears without supersession context.

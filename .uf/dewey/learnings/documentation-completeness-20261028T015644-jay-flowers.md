---
tag: documentation-completeness
author: jay-flowers
category: pattern
created_at: 2026-10-28T01:56:44Z
identity: documentation-completeness-20261028T015644-jay-flowers
tier: draft
---

Any consumer-facing documentation for a vibe-check gate surface must carry the advisory-until-required caveat: a REQUEST_CHANGES verdict fails the check job, but merges are blocked only after the consuming repository's admin enables branch protection and marks the job required. The self-gate README states this explicitly, and ci-regression-gate dedicated a documentation task to it; new gate surfaces must repeat it rather than assume readers transfer the semantics. Canonical case: vibe-check opsx/reusable-structural-gate README "CI integration" section (2026-10-28) — the intro paragraph states the advisory-until-required semantics before any usage snippet, so a consumer skimming only the minimal snippet still encounters the caveat. Pattern: put the caveat in the section intro, not a footnote, because GitHub-flavored README sections are often read snippet-first.

---
tag: security-posture
author: jay-flowers
category: gotcha
created_at: 2026-11-01T17:29:41Z
identity: security-posture-20261101T172941-jay-flowers
tier: draft
---

Constitution-alignment claims in proposal tables persist into merged code unverified unless a task pins the control's call site. Worked example (vibe-check ripwire-adapter, PR #75): the 2026-10-30 spec review flagged that the adapter claimed "V. Security by Default: PASS — ... path validation" while planning only a Stat+IsDir check; the spec delta was cleaned to drop the path-validation requirement, but the proposal table kept the claim and the implementation shipped checkProjectDir with zero metrics.ValidateProjectPath call sites. The gap survived two code-review cycles because (a) no tasks.md item required the call site, (b) code reviewers audit changed code against spec requirements — and the requirement had been removed — not against constitution tables, and (c) the claim lives in a document reviewers no longer read at code-review time. Pattern: for every security control named in a constitution-alignment table, require a named call-site task in tasks.md and a test asserting the control fires; otherwise strike the claim from the table before implementation begins.

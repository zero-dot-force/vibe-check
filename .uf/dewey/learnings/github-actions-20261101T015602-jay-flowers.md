---
tag: github-actions
author: jay-flowers
category: gotcha
created_at: 2026-11-01T01:56:02Z
identity: github-actions-20261101T015602-jay-flowers
tier: draft
---

When a design doc rejects making a generator-emitted CI template adaptive on the grounds that \"repo introspection requires network or gh at init time,\" check whether local git metadata already answers the question offline before accepting the premise: `git symbolic-ref refs/remotes/origin/HEAD` (set by ordinary `git clone`; absent in manually-wired remotes and non-git targets) resolves the default branch with zero network. This keeps the emitted template bytes deterministic while upgrading an unconditional documentation hint into a conditional, precise notice — the difference between \"edit branches: when your default branch is not main\" (which every user must read and self-diagnose) and a warning that fires only for the repos actually at risk of the silent-gate-never-triggers failure mode (worst case: check marked required in branch protection and PRs deadlock). Canonical case: vibe-check opsx/init-scaffolds-structural-gate-caller design D9 (2026-10-31) — its rejection rationale overstated the cost of detection, and the branch's own committed github-actions learning recommended the local probe the rationale ruled out as network-bound. Rule of thumb: enumerate local-git, filesystem, and environment probes as a separate introspection tier before concluding adaptivity conflicts with an offline-capable path; best-effort detection with a documented fallback keeps both determinism and precision.

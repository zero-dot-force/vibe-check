---
tag: coverage-ratchet
author: jay-flowers
category: gotcha
created_at: 2026-10-30T20:45:04Z
identity: coverage-ratchet-20261030T204504-jay-flowers
tier: draft
---

A coverage-ratchet baseline committed from a feature branch cut behind its review base records the branch tree's measurements, while pull_request CI evaluates the merge commit — packages main touched after the branch point measure differently (usually higher, from added code+tests) on the merged tree, so their committed floors pass with silent headroom. Pre-merge reconciliation is blocked by local CI parity (the branch tree measures the lower value, so raising the floor fails the local run), and a blanket --update is blocked by documented accommodation floors. The working pattern: reconcile every floor against the merge-commit measurement; where pre-merge reconciliation is impossible, disclose the headroom in the CHANGELOG's exceptions list and open a tracking issue for the post-merge targeted re-raise, mirroring the metrics/#71 accommodation pattern. Canonical near-miss: vibe-check opsx/ci-coverage-ratchet-floors committed internal/scaffold at 87.8 (branch measurement) while the merge tree measures 91.6, disclosed only in the PR body — the CHANGELOG's "two documented exceptions" claim omitted it and no re-raise issue existed at review time.

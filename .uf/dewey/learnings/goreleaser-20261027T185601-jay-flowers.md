---
tag: goreleaser
author: jay-flowers
category: gotcha
created_at: 2026-10-27T18:56:01Z
identity: goreleaser-20261027T185601-jay-flowers
tier: draft
---

GoReleaser v2 changelog-groups gotcha (three-part): (1) `changelog.use: github-native` silently ignores the entire `changelog.groups` block — GitHub generates notes natively, so an incomplete group set stays latent; (2) in default git mode, commits matching no group regexp are silently OMITTED from release notes unless a regexp-less catch-all group is declared (canonical: `- title: Others` with highest `order`, no regexp); (3) `goreleaser build --snapshot` does not exercise the changelog pipe, so group completeness is only proven by `goreleaser release --snapshot` or a real dispatch. Review heuristic: when a spec/checklist/config claims changelog groups "(A, B, C, Others)", diff the claimed titles against the actual YAML — titled groups without a trailing regexp-less catch-all is the tell; the fix is 2-3 lines. Seen in zero-dot-force/vibe-check goreleaser-release-automation: spec D6/tasks/checklist/CHANGELOG all promised four groups while .goreleaser.yaml shipped three, masked by `use: github-native`; removing it (correct per spec) made the incomplete filter the live v0.2.0 generator, which would have dropped most release-automation commits (chore/ci) from the notes. Caught by review council iteration 5 (architect HIGH + guard/sre MEDIUM, three independent reviewers converging).

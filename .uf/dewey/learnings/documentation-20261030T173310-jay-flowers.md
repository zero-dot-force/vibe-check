---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:10Z
identity: documentation-20261030T173310-jay-flowers
tier: draft
---

When a change adds a new cmd/ binary and asserts a CI-only / repo-internal documentation disposition (exempting it from the docs-issue and website-sync gates), verify the claim against the release pipeline before accepting it: read .goreleaser.yaml (or the release workflow's build matrix) and confirm the build list is pinned to the product binary rather than globbing cmd/*, and confirm the README documents only the shipped surface. A pinned single-build config keeps the new tool out of the user-facing installation surface, which is what makes the exemption sound; a glob or multi-build config would ship the binary and make the disposition factually wrong (new installation surface, so README and website-sync obligations fire). Canonical case: vibe-check opsx/ci-coverage-ratchet-floors (2026-10-30, fix iteration 2): task 4.3's "CI-only change, exempt from the user-facing documentation-issue gate" disposition for cmd/coverage-ratchet was verified sound because .goreleaser.yaml declares exactly one build (main: ./cmd/vibe-check, id: vibe-check), so the ratchet binary is never released and the README's product-CLI scope stays accurate with no update needed.

---
tag: yaml-splicing
author: jay-flowers
category: gotcha
created_at: 2026-10-29T18:12:27Z
identity: yaml-splicing-20261029T181227-jay-flowers
tier: draft
---

Line-based YAML splices that anchor an insertion "after the last collected item" must guard the line that follows the anchor: if it is indented deeper than the item lines, it is a continuation of that item (nested mapping, block scalar), and inserting a sibling item there reassigns the continuation to the inserted item — every byte preserved, YAML semantics silently changed, no notice emitted. Symmetric recognition gap: entry/item-marker matching ("- agent:", "- name") done indentation-blind reads the same lines inside block scalars as present entries, suppressing insertion (fail-closed downstream at dispatch, not at init). Hardening: treat a deeper-indented follower at the insertion anchor as an unrecognized shape (notice, no edit), and require markers to sit at the section's minimum item indent. Canonical case: vibe-check internal/scaffold spliceMatrix insertion anchor and entryBlocks marker scan (init-registers-divisor-entropy, PR #65 adversary review).

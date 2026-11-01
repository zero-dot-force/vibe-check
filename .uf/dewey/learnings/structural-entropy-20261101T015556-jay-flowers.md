---
tag: structural-entropy
author: jay-flowers
category: pattern
created_at: 2026-11-01T01:55:56Z
identity: structural-entropy-20261101T015556-jay-flowers
tier: draft
---

Verified on vibe-check opsx/init-scaffolds-structural-gate-caller (base 9ccb82d → head 153ecfc, vibe-check v0.1.1-0.20261005175013-a5093d797202): adding exactly one new external-module import to an existing package is the minimal non-zero structural delta a PR can produce. internal/scaffold gained golang.org/x/mod/semver (design D13 promotion) and only its Ce moved, 9→10, with Instability 0.9→0.9091 (Δ ≈ +1/(Ca+Ce+1) ≈ 0.009), Distance 0.1→0.0909 (closer to the main sequence), while Ca, abstractness, and LCOM4 held — yet vibe-check diff returns COMMENT (materiality band: "non-zero metric shifts below REQUEST_CHANGES thresholds"), not APPROVE. So a single spec-sanctioned dependency addition still shifts the verdict off APPROVE; expect COMMENT, not a gate failure, whenever exactly one new efferent edge lands. Companion facts from the same measurement: (1) test-only imports (go.yaml.in/yaml/v3 in ci_test.go) do not enter Ce — the adapter counts the non-test package variant; (2) stdlib imports added to one file of a package whose union already contains them (os/exec and strings in init.go; analyze.go already imported both) leave Ce unchanged; (3) workflow YAML renames, embedded templates, golden fixtures, docs, and spec artifacts produce zero metric movement. Extends learning/structural-entropy-20261029T155410-jay-flowers (same-package growth with no new imports → all-zero delta → APPROVE).

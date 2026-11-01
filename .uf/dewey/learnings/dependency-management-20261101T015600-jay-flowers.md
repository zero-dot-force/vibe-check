---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-01T01:56:00Z
identity: dependency-management-20261101T015600-jay-flowers
tier: draft
---

Promoting a graph-present module (go.sum carries only its `/go.mod` hash) to a direct require can add `h1:` zip-hash lines for its test-only transitive dependencies as well: promoting go.yaml.in/yaml/v3 v3.0.4 added both yaml/v3's `h1:` line and gopkg.in/check.v1's (yaml's test-only transitive, previously graph-present with only its `/go.mod` hash), though no check.v1 package is ever compiled by the main module. Dependency-facts sections and "expected go.sum delta" verification clauses must enumerate transitive test-dep hash lines, or the next go.sum audit finds unexplained additions — extending the /go.mod-vs-h1 distinction already recorded in the go-modules learnings. Verified on vibe-check opsx/init-scaffolds-structural-gate-caller (2026-10-31): design D13 predicted yaml/v3's `h1:` line; the committed go.sum gained two.

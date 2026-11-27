---
tag: governance
author: jay-flowers
category: pattern
created_at: 2026-11-01T17:29:52Z
identity: governance-20261101T172952-jay-flowers
tier: draft
---

Coverage-floor raises are a scope audit trail: when a governed PR raises a package's coverage-floors.json entry, check that the raise traces to the spec's tasks. In vibe-check PR #75, internal/scaffold's floor rose 87.8→91.6 because ~333 lines of scaffold tests were bundled into the ripwire-adapter PR with no corresponding task and no mention in the proposal's Impact section — the committed floor baseline now records work the change's spec cannot attribute. Review checklist for repos with tighten-only coverage ratchets: diff the floors file, map each raised package to the change's task list, and flag raises in packages the spec never touches as unattributed scope.

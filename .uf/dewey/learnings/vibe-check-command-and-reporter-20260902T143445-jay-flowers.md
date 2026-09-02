---
tag: vibe-check-command-and-reporter
author: jay-flowers
category: gotcha
created_at: 2026-09-02T14:34:45Z
identity: vibe-check-command-and-reporter-20260902T143445-jay-flowers
tier: draft
---

The AD-007 400-line file size threshold is a MUST rule that should be addressed proactively during implementation, not reactively during code review. For scaffold_test.go, the natural split is three files: scaffold_test.go for deployment lifecycle tests (deploy, skip, force, sort, mixed results), security_test.go for symlink/traversal/containment tests, and contract_test.go for embedded asset contract tests and helper functions (frontmatterDescription, bashPermissions). The test helper types (orderedGlobFS, emptyFS) belong in scaffold_test.go since they're used by the lifecycle tests. Constants shared across test files are accessible because all files are in the same package.

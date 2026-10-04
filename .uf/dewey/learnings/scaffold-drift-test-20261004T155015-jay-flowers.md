---
tag: scaffold-drift-test
author: jay-flowers
category: pattern
created_at: 2026-10-04T15:50:15Z
identity: scaffold-drift-test-20261004T155015-jay-flowers
tier: draft
---

TestEmbeddedAssetsMatchSource is an integration/contract test that reads the real .opencode/ directory from the project root. It depends on go.mod being present in an ancestor directory (found via findProjectRoot walking up from cwd) and on .opencode/ having been populated by vibe-check init. On a fresh clone, this test will fail with `.opencode/ directory not found` unless vibe-check init has been run first. The test is gated behind testing.Short() to allow CI to skip it with go test -short. The remediation hint in failure messages directs users to run `vibe-check init --force .` from the repo root. The test intentionally violates TC-004 (t.TempDir() isolation) because it must verify the real deployment, not a temporary copy — this is a documented, intentional departure justified in design D6.

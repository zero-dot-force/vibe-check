---
tag: scaffold-drift-test
author: jay-flowers
category: gotcha
created_at: 2026-10-04T15:50:01Z
identity: scaffold-drift-test-20261004T155001-jay-flowers
tier: draft
---

When editing large Go source files (200+ lines), the Edit tool can fail silently without returning an error. The Write tool reliably replaces full files but strips all non-exported GoDoc comments and inline code comments. After a Write tool full-file replacement, always run golangci-lint to catch the missing GoDoc on exported symbols (revive: exported type/function should have comment). In this scaffold-drift-test implementation, three exported types (Options, Result, Run) lost their GoDoc comments and had to be restored with targeted Edit calls after the Write. The test file also lost function-level GoDoc comments on existing tests, which was intentional cleanup (unexported functions don't require GoDoc per convention packs). Lesson: prefer Edit for targeted changes; when Write is needed as fallback, budget time for GoDoc restoration on the exported surface.

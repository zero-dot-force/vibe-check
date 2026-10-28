---
tag: task-authoring
author: jay-flowers
category: gotcha
created_at: 2026-10-28T01:56:33Z
identity: task-authoring-20261028T015633-jay-flowers
tier: draft
---

Task Verify clauses that dry-run CI command sequences pinned to a linux/amd64 runner are not executable on a darwin development machine: linux ELF binaries fail with exec-format errors and macOS lacks GNU sha256sum by default. Author such verifications with an explicit platform adaptation — run the exact sequence in a linux container, or substitute the platform-matched release archive for local smoke tests (a release checksums.txt verifies any single platform's archive line, so grep the exact asset line and pipe it to `shasum -a 256 --check` on macOS). Canonical case: vibe-check opsx/reusable-structural-gate task 1.2 (2026-10-28) — its Verify clause assumed linux_amd64; the local dry-run on darwin succeeded only by substituting the darwin archive and `shasum -a 256`, which still proved the URL resolution, checksum-line format, and fail-closed grep behavior the task cared about.

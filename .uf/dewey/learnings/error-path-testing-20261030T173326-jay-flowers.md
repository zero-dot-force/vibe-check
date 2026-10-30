---
tag: error-path-testing
author: jay-flowers
category: pattern
created_at: 2026-10-30T17:33:26Z
identity: error-path-testing-20261030T173326-jay-flowers
tier: draft
---

Covering both failure points of an atomic temp+rename write with two distinct directory injections: create a directory at the staged .tmp path so the staging os.WriteFile fails (EISDIR on open), and a directory at the final target path so the os.Rename commit fails. Each test then asserts the staged file leaves no .tmp residue and the error maps to the tool-failure exit path (exit 2). The two injections exercise different lines of the same write helper — one alone leaves the other failure branch untested. Demonstrated in internal/coverratchet/baseline_test.go (TestWriteBaseline_RenameFailureCleansStagedFile) and cmd/coverage-ratchet/ratchet_test.go (TestRunRatchet_UpdateBaselineWriteFailure).

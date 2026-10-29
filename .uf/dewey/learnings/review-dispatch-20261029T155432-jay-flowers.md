---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-10-29T15:54:32Z
identity: review-dispatch-20261029T155432-jay-flowers
tier: draft
---

Divisor dispatch children can fail without executing (usage null, no findings) and recover without re-planning: dispatch_review_runs reported 3 of 7 reviewers failed with null usage on the init-registers-divisor-entropy iteration-2 council; redispatch_failed_runs (DR-FR-002) re-ran architect and testing successfully on the first recovery, while divisor-guard timed out once (retryable error code "timeout") and succeeded on a second targeted redispatch with an explicit 900s timeout. Recovery protocol: dispatch_status to identify failures, redispatch_failed_runs for all failed, then a named-agent redispatch with a raised timeout for any retryable timeout; the superseded/ archive preserves the audit trail and consolidation stays clean.

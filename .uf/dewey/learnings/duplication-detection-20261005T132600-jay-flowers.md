---
tag: duplication-detection
author: jay-flowers
created_at: 2026-10-05T13:26:00Z
identity: duplication-detection-20261005T132600-jay-flowers
tier: draft
---

Multi-way duplicate detection using pairwise comparison produces O(n²) Duplication entries for n identical functions (e.g., functions A, B, C produce pairs A-B, A-C, B-C). When computing a duplication percentage, summing LineCount across all entries double-counts lines from functions that appear in multiple pairs. The fix is to deduplicate blocks by a unique key (file path + line range, e.g., "file.go:10-20") using a seen map before summing. This ensures each duplicated line is counted once regardless of how many pairs it appears in. This pattern applies to any metric that aggregates across pairwise comparisons.

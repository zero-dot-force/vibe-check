---
tag: dewey
author: jay-flowers
category: gotcha
created_at: 2026-10-27T17:03:21Z
identity: dewey-20261027T170321-jay-flowers
tier: draft
---

Correcting a Dewey learning by editing its materialized file under .uf/dewey/learnings/ fixes the committed record but does not update the learning-store index: dewey_semantic_search continues to serve the pre-edit content under the same learning identity. Verified 2026-10-27 on opsxgoreleaser-release-automation-20261004T141907-jay-flowers: after in-place correction of the hyphenated-filename claim, semantic search still returned the original wrong text (indexed_at unchanged at the original store date) while the corrected file and the same-day disk-source re-index of CHANGELOG.md confirmed the divergence. When remediating a defective learning in place, verify post-merge that retrieval no longer surfaces the stale claim; re-store or re-index the learning if the store supports refresh, and ensure explicitly-superseding learnings are indexed so they outrank the stale entry in the interim.

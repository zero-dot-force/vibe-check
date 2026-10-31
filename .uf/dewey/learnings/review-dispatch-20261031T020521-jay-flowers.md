---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:05:21Z
identity: review-dispatch-20261031T020521-jay-flowers
tier: draft
---

consolidate_dispatch with provenance can fail schema validation with "Unrecognized key: recommendation" when child Divisor reviewers submit findings carrying an extra recommendation field (the finalize schema allows only severity/category/description/root_cause/file/line). Recovery (verified on correlation 6a572fb3, 2026-10-31, opsx/ripwire-adapter): the dispatch store lives at /var/folders/.../T/opencode/dispatch-<correlation_id>/ (locate via `find $TMPDIR/opencode ~/.local/share/opencode -name "*<correlation-prefix>*"`); load each active run-divisor-*.json, delete the recommendation key from every finding, rewrite the file, then re-run consolidate_dispatch with the same provenance — it finalizes successfully and writes the .uf/artifacts/dispatch/<correlation_id>.json artifact. The recommendation text is not lost: consolidate_dispatch returns the proposals array (including lesson proposals) regardless, and superseded/ run files keep the original bytes.

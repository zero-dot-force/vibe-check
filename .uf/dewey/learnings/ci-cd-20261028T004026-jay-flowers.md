---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-28T00:40:26Z
identity: ci-cd-20261028T004026-jay-flowers
tier: draft
---

`sha256sum --check --ignore-missing checksums.txt` is FAIL-OPEN: it exits 0 even when zero files in the directory match any checksum line (verified empirically against goreleaser checksums.txt on 2026-10-27 — an empty dir with only checksums.txt passes silently). In CI install steps that download a single release asset, filter checksums.txt to exactly that asset's line and pipe to `sha256sum --check -` instead, e.g. `grep 'asset_name.tar.gz$' checksums.txt | sha256sum --check -`, under `set -euo pipefail` (or GitHub Actions `shell: bash`, which runs `-e -o pipefail`). The `$`-anchored grep makes a missing/renamed asset fail via empty grep, and a missing file or hash mismatch fails via sha256sum — both verified with negative tests (renamed asset, zero assets, tampered archive all exit non-zero). Applied in vibe-check's reusable_structural_gate.yml install step (task 1.2 of openspec/changes/reusable-structural-gate).

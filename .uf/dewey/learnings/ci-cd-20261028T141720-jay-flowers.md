---
tag: ci-cd
author: jay-flowers
category: gotcha
created_at: 2026-10-28T14:17:20Z
identity: ci-cd-20261028T141720-jay-flowers
tier: draft
---

CI install steps that download release assets directly with curl should pass retry flags (e.g. `curl -fsSLO --retry 3 --retry-delay 2`): a transient CDN/network error otherwise fails the job spuriously, and a fail-closed-but-flaky gate erodes the operational meaning of red. Separately, N sequential requests to a `releases/latest/download` endpoint are not guaranteed to resolve to the same release: a release published between the archive download and the checksums.txt download yields artifacts from two different releases, and the (correct) checksum verification fails spuriously — fail-closed with no integrity risk, but an avoidable red run. Fix pattern: retry flags on every download; under `latest`, either accept the rare race window or resolve the redirect once to a concrete tag and fetch all files from that tag's download directory. Canonical case: vibe-check reusable_structural_gate.yml install step (opsx/reusable-structural-gate, PR #59 SRE review, 2026-10-28).

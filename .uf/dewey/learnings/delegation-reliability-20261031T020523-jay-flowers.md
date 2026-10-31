---
tag: delegation-reliability
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:05:23Z
identity: delegation-reliability-20261031T020523-jay-flowers
tier: draft
---

uf_delegate_command on the GLM (enmaas/rits/zai-org/glm-5-3) provider intermittently kills parented child sessions after a long reasoning-only turn (~8 minutes burned, then death, no artifact produced). Observed repeatedly during the opsx/ripwire-adapter implementation (phases 1-5, 2026-10-30); every task delegated after the first death also died. Working mitigation: attempt at most ONE delegation per ultra-directive prompt (command_path + tight prompt + expected_artifact), and on death self-implement the task directly in the parent session instead of retrying. Cost: delegation is still worth one attempt per phase for verifiable artifacts; retrying the same delegation reliably re-dies.

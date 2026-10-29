---
tag: uf-unleash
author: jay-flowers
category: gotcha
created_at: 2026-10-29T15:54:26Z
identity: uf-unleash-20261029T155426-jay-flowers
tier: draft
---

uf.unleash mark_task_complete calls race when parallelized: two concurrent calls on the same tasks.md each read-modify-write the file and the later write clobbers the earlier checkbox flip, silently un-marking a completed task. Always invoke mark_task_complete strictly sequentially (one task per message, waiting for each result). Canonical case: tasks 5.3/5.4 were clobbered by parallel calls during the init-registers-divisor-entropy implement phase and had to be re-marked (commit f78780f).

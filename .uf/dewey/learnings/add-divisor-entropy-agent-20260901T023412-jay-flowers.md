---
tag: add-divisor-entropy-agent
author: jay-flowers
category: gotcha
created_at: 2026-09-01T02:34:12Z
identity: add-divisor-entropy-agent-20260901T023412-jay-flowers
tier: draft
---

On 2026-08-31, while implementing the divisor-entropy change on branch add-divisor-entropy-agent in the vibe-check repo, delegating work to subagents proved unreliable under an autonomous /uf.unleash run: the cobalt-crush-dev subagent returned an empty result and created NO file for an implementation task (internal/scaffold/scaffold.go did not exist after the Task reported 'completed'), and later a gaze-reporter review subagent was cancelled under context pressure. The durable lesson (category: gotcha): never assume a 'completed' delegation actually produced its artifact — immediately verify by reading the expected output file (or checking git status), and be ready to self-implement or self-review as a sanctioned fallback. This matters most during long autonomous pipelines where a silent no-op would otherwise cascade into a broken build gate several steps later.

---
tag: openspec-artifact-consistency
author: jay-flowers
created_at: 2026-10-04T21:53:55Z
identity: openspec-artifact-consistency-20261004T215355-jay-flowers
tier: draft
---

The most common OpenSpec artifact failure mode is spec/tasks/design inconsistency: tasks require content the spec doesn't define, and design claims a structure the tasks don't reflect. In doc-agent-design-pack, tasks.md had 5 sections but design.md claimed "two main sections." The spec had 5 requirements but tasks required 3 additional deliverables (PASS/FAIL examples, introduction, tracking issue links). The correct fix direction is spec-driven: add the missing requirements to spec.md, then realign design.md and tasks.md to match. Never delete tasks to make them match an incomplete spec — the tasks reflect what stakeholders actually need from the change.

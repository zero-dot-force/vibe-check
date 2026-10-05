---
tag: opsxarchitectural-drift-tracking
author: jay-flowers
category: pattern
created_at: 2026-10-05T16:54:42Z
identity: opsxarchitectural-drift-tracking-20261005T165442-jay-flowers
tier: draft
---

Architectural drift tracking implementation added a --store flag to vibe-check analyze that enriches provenance with git metadata (commitSHA, branch, modulePath). The key design insight was that snapshot storage happens via agent (mx-f-architecture-trend) calling dewey_store_learning, not via Go binary code, to avoid importing Dewey SDK dependencies into the metrics core. The binary only enriches output; the agent captures it in CI and stores it. This separation of concerns (Go binary = analysis, agent = storage/integration) keeps the Go codebase pure while leveraging agent-native MCP tools for the Dewey integration. When implementing similar features, prefer agent-side integration for MCP-exposed backends rather than adding HTTP/RPC clients to the core binary.

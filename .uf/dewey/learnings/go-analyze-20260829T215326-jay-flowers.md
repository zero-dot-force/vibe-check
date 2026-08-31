---
tag: go-analyze
author: jay-flowers
category: context
created_at: 2026-08-29T21:53:26Z
identity: go-analyze-20260829T215326-jay-flowers
tier: draft
---

For the vibe-check Go adapter, the go/packages load mode must include NeedTypesInfo in addition to NeedName|NeedImports|NeedTypes|NeedSyntax|NeedModule. The NeedTypesInfo flag is required for LCOM4 computation because the types.Info.Selections map is needed to resolve field accesses in method bodies — without it, the adapter cannot reliably determine which struct fields each method accesses, which is the foundation of the connected-component LCOM4 algorithm. The environment sanitization allowlist for go/packages should include GOPATH, GOROOT, GOMODCACHE, GOPROXY, GONOSUMCHECK, GOMOD but must NOT include GOFLAGS (which enables arbitrary flag injection into go list subprocesses, including -toolexec for command execution).

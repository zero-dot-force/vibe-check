---
tag: authority-statement-pattern
author: jay-flowers
created_at: 2026-10-04T21:53:53Z
identity: authority-statement-pattern-20261004T215353-jay-flowers
tier: draft
---

Any derivative documentation artifact must declare its authoritative source. In the doc-agent-design-pack change, design.md specified this as a mitigation strategy, but the spec.md initially had no requirement for it. Three reviewers independently flagged this gap: the authority statement existed as a design decision but wasn't traceable to a spec requirement or implementation task. The fix added a spec requirement ("The documentation SHALL state that .opencode/uf/packs/agent-design.md is the authoritative source") with a GIVEN/WHEN/THEN scenario, plus an explicit implementation task 4.2. This ensures the authority statement can't be accidentally omitted during implementation.

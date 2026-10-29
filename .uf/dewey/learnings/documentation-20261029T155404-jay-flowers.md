---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-10-29T15:54:04Z
identity: documentation-20261029T155404-jay-flowers
tier: draft
---

Paraphrasing a scanner's recognition contract for user-facing docs can invert the behavior. In vibe-check's UF policy splice (init-registers-divisor-entropy, reviewed 2026-10-29), the GoDoc "flow-collection entries are outside the presence check's contract" was accurate, but the README compression "triggers no registration" told users init does nothing — when non-recognition by the presence check actually means the agent is treated as absent and insertion PROCEEDS (producing a duplicate block-style entry next to the flow entry). When documenting a recognition limit in user-facing docs, always state the behavioral consequence of non-recognition explicitly (what the tool does when it fails to recognize), and keep the doc's use of domain terms ("registration" = the insertion act) consistent with the surrounding section.

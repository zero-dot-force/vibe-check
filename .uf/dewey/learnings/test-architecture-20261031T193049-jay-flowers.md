---
tag: test-architecture
author: jay-flowers
category: pattern
created_at: 2026-10-31T19:30:49Z
identity: test-architecture-20261031T193049-jay-flowers
tier: draft
---

When a test-double harness (fake-binary PATH-install wrapper) is needed in two packages, the copies drift fast: cmd/vibe-check/cli_ripwire_test.go and internal/ripwireadapter/fake_ripwire_test.go both define installFakeRipwireWrapper with different signatures and different shell-quoting strategies for the same injection-safety concern. Either extract the double into an internal test-support package (ordinary internal packages are importable from _test files), or at minimum share the shell-quoting primitive so security hardening propagates to every copy.

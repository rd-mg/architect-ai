---
description: Verify caveman register firewall integrity
agent: sdd-orchestrator
---

# arch:firewall-check

Verify caveman register firewall is present in all targets.

## Steps

1. Check firewall: `go run ./cmd/firewall check`
2. If missing, inject: `go run ./cmd/firewall inject`
3. Re-check: `go run ./cmd/firewall check`

## Targets

- sdd-apply protocol
- SKILL.md
- skill-registry.md

---
description: Check and inject adaptive reasoning gate v2
agent: sdd-orchestrator
---

# arch:gate

Check and inject adaptive reasoning gate v2 across all targets.

## Steps

1. Check current gate status: `go run ./cmd/gate check`
2. Inject into missing targets: `go run ./cmd/gate inject`
3. Purge L2 auto-scoring: `go run ./cmd/gate l2-purge`
4. Verify: `go run ./cmd/gate check`

## Targets

- sdd-orchestrator.md
- general-orchestrator.md
- template .tmpl files

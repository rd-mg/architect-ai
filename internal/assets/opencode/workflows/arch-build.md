---
description: Rebuild architect-ai deployed configs
agent: sdd-orchestrator
---

# arch:build

Rebuild all architect-ai deployed configs from internal/assets sources.

## Steps

1. Generate foundation: `go run ./cmd/foundation`
2. Build deployed configs: `go run ./cmd/build`
3. Validate: `go run ./cmd/check`

## Artifacts

- CLAUDE.md, GEMINI.md, .antigravity/agent.md
- .github/copilot-instructions.md
- .atl/_generated/foundation.md

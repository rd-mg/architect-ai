# Design: Cognitive Postures Expansion

## Architecture
This is a prompt-only change expanding the `cognitive-mode` skill and orchestrator injection tables. No binary changes are required.

## Posture Definitions
- **+++Economic**: Focused on value/cost ratio and explicit budget constraints.
- **+++Empirical**: Focused on measurement-first reasoning and data-driven proof.

## File Changes
| File | Action | Purpose |
|------|--------|---------|
| `internal/assets/skills/cognitive-mode/SKILL.md` | MODIFY | Add 2 posture definitions and update mapping table. |
| `docs/cognitive-modes.md` | MODIFY | Update documentation and orthogonality notes. |
| `internal/assets/*/sdd-orchestrator.md` | MODIFY | Sync posture-injection tables (9 agents). |
| `internal/assets/skills/sdd-tasks/SKILL.md` | MODIFY | Adopt Economic default. |
| `internal/assets/skills/sdd-verify/SKILL.md` | MODIFY | Adopt conditional Empirical. |
| `internal/assets/assets_test.go` | MODIFY | Add 8-posture presence test. |

## Selection Logic (Sub-agent heuristic)
Sub-agents (`sdd-design`, `sdd-verify`) will conditionally add `+++Empirical` if the task's acceptance criteria contain numeric thresholds (latency, throughput, etc.).

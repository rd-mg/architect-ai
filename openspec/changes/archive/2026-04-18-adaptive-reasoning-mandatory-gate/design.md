# Design: Adaptive Reasoning Mandatory Gate

## Technical Approach

We will move the `adaptive-reasoning` classifier from an optional skill to a structural mandatory gate. The gate text will be defined once in a shared asset and injected into all orchestrator templates via markers. Orchestrators will be updated to validate the sub-agent's reasoning mode declaration and re-prompt if missing.

## Architecture Decisions

### Decision: Single Source of Truth for Gate Text
**Choice**: `internal/assets/skills/_shared/adaptive-reasoning-gate.md`
**Alternatives considered**: Inline in every orchestrator, shared Go string constant.
**Rationale**: Keeps orchestrator files clean and avoids drift across 10 agent families. Go constant would require binary rebuilds for prompt tweaks; assets are more flexible.

### Decision: Validation and Re-prompt Logic
**Choice**: Prompt-side instruction in "Result Processing" section of orchestrators.
**Alternatives considered**: Go-side regex validation in `adapter.go`.
**Rationale**: Minimizes Go code changes. The orchestrator's system prompt already manages sub-agent flow; adding a result contract rule is consistent with existing patterns (like `engram-protocol` enforcement).

### Decision: Testing Injection via Byte-Identical Assertion
**Choice**: Go test that asserts the content between markers matches the source file exactly.
**Alternatives considered**: Presence check only.
**Rationale**: Prevents accidental manual edits to the gate text inside orchestrators, ensuring uniform behavior across all models.

## Data Flow

    Orchestrator ──[Injects Gate]──→ Sub-Agent Prompt
                                          │
    Sub-Agent Output ←──[Parses Mode]── Orchestrator
          │                               │
    [Invalid] ───[Re-prompt Once]───────┘
          │
    [Valid] ────[Execute Phase Protocol]

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/assets/skills/_shared/adaptive-reasoning-gate.md` | Create | Canonical gate text |
| `internal/assets/*/sdd-orchestrator.md` | Modify | Add markers and result contract fields |
| `internal/assets/*/sdd-phase-protocols/*.md` | Modify | Add reference note to top of files |
| `internal/assets/skills/sdd-*/SKILL.md` | Modify | Add reference note to top of files |
| `internal/assets/assets_test.go` | Modify | Add TestAdaptiveReasoningGateInjected |
| `docs/adaptive-reasoning-v1.md` | Modify | Document the gate and enforcement |

## Interfaces / Contracts

### Sub-Agent Result Contract
```markdown
Each phase returns: ..., `chosen_mode`, `mode_rationale`.
```

### Mode Declaration Pattern (Regex)
```
(?m)^Mode:\s*(1|2|3|deferred|sdd-first)\s*\.\s*Why:\s*(.+?)\s*$
```
Scan first 5 non-blank lines.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (Assets) | Gate Injection | Verify byte-identical match in all 10 orchestrators |
| Unit (Assets) | Protocol Reference | Verify every phase protocol contains the gate reference phrase |
| Manual | Sub-agent Flow | Run a dummy change and verify Mode declaration appears |

## Migration / Rollout

No migration required. Existing in-flight sessions will see the new gate instruction in their next turn. Orchestrators will re-prompt once if the sub-agent doesn't comply immediately.

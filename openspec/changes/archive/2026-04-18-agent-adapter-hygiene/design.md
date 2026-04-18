# Design: Agent Adapter Hygiene

## Technical Approach

Standardize the agent adapter contract by formalizing requirements in documentation and fixing inaccurate hook reporting in host-wrapped adapters (VSCode, Antigravity). We will also harden the Antigravity adapter's token extraction to support Google-native payloads and ensure nil-safety across all adapters.

## Architecture Decisions

### Decision: Formalize Adapter Contract in Markdown

**Choice**: Create `internal/agents/ADAPTER-CONTRACT.md`.
**Alternatives considered**: Inline comments in `interface.go`.
**Rationale**: Markdown allows for rich formatting, examples, and is more visible to contributors than interface comments alone. It serves as a machine-readable-ish checklist for compliance audits.

### Decision: Hook Reporting Strategy for Wrapped Agents

**Choice**: `SessionHookEnabled()` returns `false` for VSCode.
**Alternatives considered**: Return `true` but log warning.
**Rationale**: Explicitly reporting incapacity prevents core logic from expecting non-existent data, leading to cleaner telemetry and UI state.

### Decision: Antigravity Payload Support

**Choice**: Dual support for OpenAI-compatible and Google-native schemas in `ExtractUsage`.
**Alternatives considered**: Stick to one and force normalization elsewhere.
**Rationale**: Adapters should be resilient to their provider's native output to avoid unnecessary middle-layer transformations.

## Data Flow

    Provider Response ──→ Adapter.RecordResponse() ──→ metering.ExtractUsage()
                                    │                           │
                                    │                  (Google-native / OpenAI)
                                    ↓                           ↓
                           metering.Current() ──────→ hook.Record(delta)

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/agents/ADAPTER-CONTRACT.md` | Create | Specification of adapter requirements and nil-safety invariants. |
| `internal/agents/interface.go` | Modify | Link to `ADAPTER-CONTRACT.md` and clarify interface docs. |
| `internal/agents/vscode/adapter_metering.go` | Modify | Correctly report hook incapacity. |
| `internal/agents/antigravity/adapter_metering.go` | Modify | Harden token extraction and report hook status correctly. |
| `internal/agents/vscode/adapter_metering_test.go` | Create | Nil-safety and type-safety tests. |
| `internal/agents/antigravity/adapter_metering_test.go` | Create | Nil-safety and Google-native payload tests. |

## Interfaces / Contracts

```go
// internal/agents/interface.go

// MeteringCapable identifies adapters that can observe and report token usage.
type MeteringCapable interface {
    // SessionHookEnabled reports if the adapter can observe usage from its host.
    SessionHookEnabled() bool
    
    // RecordResponse extracts usage and feeds it to the global hook.
    // MUST be nil-safe and handle unexpected types gracefully.
    RecordResponse(raw []byte)
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Nil-safety | Call `RecordResponse(nil)` and `RecordResponse([]byte("garbage"))`. |
| Unit | Token Extraction | Provide mock JSON payloads (OpenAI and Google-native) to `ExtractUsage`. |
| Integration | Hook Propagation | Verify `metering.Current().Record()` is called with correct values in a controlled test. |

## Migration / Rollout

No migration required. Changes improve accuracy of future telemetry.

# Adapter Contract

Every agent adapter MUST expose:
- `SessionHookEnabled() bool` — Reports whether the adapter can observe token counts from its host environment.
- `RecordResponse(raw []byte)` — Extracts token usage and feeds it to the global session hook.

## Invariants

- **Nil-Safety**: `RecordResponse(nil)` MUST NOT panic.
- **Type-Safety**: `RecordResponse` MUST handle malformed or unexpected JSON payloads gracefully (no-op).
- **Pure Detection**: `SessionHookEnabled()` MUST NOT perform any I/O or network calls.

## Compliance Grid

| Agent | Host-Wrapped? | SessionHookEnabled | Notes |
|-------|---------------|-------------------|-------|
| Claude | No | `true` | Direct API access. |
| VSCode | Yes | `false` | Copilot wrapping prevents observation. |
| Antigravity | No | `true` | Direct access, supports Google-native schema. |
| Gemini | No | `true` | Direct access. |
| Cursor | Yes | `false` | IDE wrapped. |

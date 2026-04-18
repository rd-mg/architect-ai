# Implementation Progress: Agent Adapter Hygiene
**Mode**: Standard

### Completed Tasks
- [x] 3.3 Run `go test ./internal/agents/...` (All passed).
- [x] 4.1 Update docs/comments (interface.go references contract).
- [x] 4.2 Verify metering package correctly handles incapacity (Verified via VSCode test logic).
- [x] 3.1 Create `internal/agents/vscode/adapter_metering_test.go` with nil-safety tests.
- [x] 3.2 Create `internal/agents/antigravity/adapter_metering_test.go` with Google-native tests.
- [x] 2.1 Modify `internal/agents/vscode/adapter_metering.go` to set `SessionHookEnabled() == false` (Verified already false).
- [x] 2.2 Harden `internal/agents/antigravity/adapter_metering.go` with dual OpenAI/Google-native schema support.
- [x] 2.3 Verify `SessionHookEnabled()` return values across all other 9+ adapters (Fixed Cursor).
- [x] 1.1 Create `internal/agents/ADAPTER-CONTRACT.md` defining metering invariants and nil-safety.
- [x] 1.2 Update `internal/agents/interface.go` to link to `ADAPTER-CONTRACT.md` and clarify `MeteringCapable` expectations.

### Files Changed
| `internal/agents/vscode/adapter_metering_test.go` | Created | Nil-safety tests. |
| `internal/agents/antigravity/adapter_metering_test.go` | Created | Google-native and nil-safety tests. |
| `internal/agents/antigravity/adapter_metering.go` | Modified | Added Google-native support and nil-safety. |
| `internal/agents/cursor/adapter_metering.go` | Modified | Set hook enabled to false. |
| File | Action | What Was Done |
|------|--------|---------------|
| `internal/agents/ADAPTER-CONTRACT.md` | Created | Formalized adapter contract. |
| `internal/agents/interface.go` | Modified | Added reference to contract. |

### Status
11/11 tasks complete. Ready for verify.

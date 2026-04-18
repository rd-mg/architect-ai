# Tasks: Agent Adapter Hygiene

## Phase 1: Foundation / Documentation

- [x] 1.1 Create `internal/agents/ADAPTER-CONTRACT.md` defining metering invariants and nil-safety.
- [x] 1.2 Update `internal/agents/interface.go` to link to `ADAPTER-CONTRACT.md` and clarify `MeteringCapable` expectations.

## Phase 2: Core Implementation

- [x] 2.1 Modify `internal/agents/vscode/adapter_metering.go` to set `SessionHookEnabled() == false`.
- [x] 2.2 Harden `internal/agents/antigravity/adapter_metering.go` with dual OpenAI/Google-native schema support in `ExtractUsage`.
- [x] 2.3 Verify `SessionHookEnabled()` return values across all other 9+ adapters for hygiene.

## Phase 3: Testing / Verification

- [x] 3.1 Create `internal/agents/vscode/adapter_metering_test.go` with nil-safety and type-safety tests.
- [x] 3.2 Create `internal/agents/antigravity/adapter_metering_test.go` with Google-native payload extraction tests and nil-safety.
- [x] 3.3 Run `go vet ./internal/agents/...` and verify all tests pass.

## Phase 4: Integration Audit

- [x] 4.1 Run `architect-ai run --agent antigravity --trace` (if possible) to verify live response handling.
- [x] 4.2 Verify `metering` package correctly ignores agents with `SessionHookEnabled() == false`.

# Proposal: Agent Adapter Hygiene

## Intent

Standardize and harden the agent adapter contract for VSCode and Antigravity. Currently, these adapters claim to support metering hooks while wrapped in hosts that prevent direct token observation, leading to inaccurate reporting. We must also ensure nil-safety and robust token extraction for the Antigravity (Google-native) adapter.

## Scope

### In Scope
- Create `internal/agents/ADAPTER-CONTRACT.md` to define adapter requirements.
- Fix VSCode adapter to report `SessionHookEnabled() == false`.
- Verify and fix Antigravity adapter token extraction (Google-native vs OpenAI-compatible).
- Implement nil-safety and type-safety tests for both adapters.
- Audit 9+ existing adapters for `SessionHookEnabled` hygiene.

### Out of Scope
- Migrating existing metering data.
- Refactoring the core `metering` package.
- Adding new adapters.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- agents: Add metering hook contract requirements and nil-safety invariants to the agent interface.

## Approach

1. **Documentation**: Formalize the contract in a new Markdown file.
2. **Correction**: Modify `adapter_metering.go` for VSCode to correctly report its limitations.
3. **Verification**: Use `architect-ai run --agent antigravity --trace` to capture real response payloads and adjust the `Record` method if necessary.
4. **Testing**: Add a dedicated test file per adapter verifying `Record` behavior with nil and malformed inputs.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/agents/ADAPTER-CONTRACT.md` | New | Official contract documentation. |
| `internal/agents/vscode/adapter_metering.go` | Modified | Set hook enabled to false. |
| `internal/agents/antigravity/adapter_metering.go` | Modified | Verify/fix token extraction and safety. |
| `internal/agents/*/adapter_metering.go` | Modified | Audit and standardize return values. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking changes in response schema | Med | Use trace logs to verify real-world payloads. |
| Regression in metering reporting | Low | Standardized tests for nil-safety and common types. |

## Rollback Plan

Revert `internal/agents/` to the previous git commit. The changes are local to adapter implementations and do not affect database schemas or core state.

## Dependencies

- None

## Success Criteria

- [ ] `ADAPTER-CONTRACT.md` exists and is referenced in `internal/agents/interface.go`.
- [ ] VSCode adapter reports false for session hooks.
- [ ] Antigravity adapter correctly extracts tokens from Google-native responses.
- [ ] No panics on `Record(nil)` across all modified adapters.
- [ ] `go vet ./internal/agents/...` passes.

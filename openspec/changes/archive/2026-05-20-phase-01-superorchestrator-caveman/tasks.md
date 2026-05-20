# Tasks: Phase 1 - Super-Orchestrator v2.0: Inline Execution + Delegation Triggers + Caveman


## Status: completed
## Implementation Tasks

### Assets
- [x] `internal/assets/_shared/caveman-identity-block.md` — dual-mode caveman rules (LITE/ULTRA/NORMAL)
- [x] `internal/assets/_shared/architect-identity.md` — L0 3-Mode architect role, Execution Mode, Model Routing
- [x] `internal/assets/_shared/super-orchestrator-gate.md` — 6 Mandatory Delegation Triggers, Mode A (Inline) vs Mode B (SDD) vs Mode C (General)
- [x] `internal/assets/_shared/sdd-orchestrator-identity.md` — L1a identity
- [x] `internal/assets/_shared/general-orchestrator-identity.md` — L1b identity

### Go Implementation
- [x] `internal/install/adapter/injector.go` — platform detection + section injection
- [x] `internal/install/adapter/injector_test.go` — Detect, InjectSection, ValidateInstallation tests

### Platform Adapters (section injection)
- [x] OpenCode: `opencode.json` agents with L0/L1a/L1b entries
- [x] Claude Code: `CLAUDE.md` with `<!-- architect-ai:L0:start -->` sections
- [x] VSCode Copilot: `copilot-instructions.md` with logical L0/L1/L2 sections
- [x] Antigravity: `.antigravity/agent.md` sequential simulation protocol
- [x] Gemini CLI: `GEMINI.md` with run_subagent delegation

### Tests
- [x] `TestDetect_OpenCode`, `TestDetect_Claude`, `TestDetect_NoMatch`
- [x] `TestInjectSection_NewSection`, `TestInjectSection_ReplaceExisting`
- [x] `TestValidateInstallation_AllPlatforms`

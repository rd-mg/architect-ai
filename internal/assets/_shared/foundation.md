## Project Foundation Standards [AUTO-GENERATED — do not edit]
<!-- architect-ai:foundation:start -->

### 1. Search (ripgrep)
- ALWAYS use `rg` not `grep -r`. Never `find . -name`.
- Type filter: `rg "pattern" --type go` / `--type py` / `--type js`
- File list: `rg -l "pattern"` | Context: `rg -C 3 "pattern"` | Count: `rg -c "pattern"`
- Word boundary: `rg -w "exactWord"` | JSON output: `rg --json "pattern" | jq ...`
- NEVER grep without --type in projects > 100 files.

### 2. Shell Safety (bash + fish)
- **bash**: Start scripts with `set -euo pipefail` and `IFS=$'\n\t'`.
- **fish**: Use `or begin ... end` for error handling. No `set -euo pipefail` needed.
- **Both**: Quote all variables. Use `rg` not `grep`. Check command with `command -v`.
- **Never**: `rm -rf $VAR` unquoted. Never pipe untrusted input to `bash` or `sh`.

### 3. Architecture Guardrails
- Respect module boundaries. Never import child packages from parent.
- No circular imports. No business logic in presentation layer.
- Prefer explicit error returns over panics. Log errors at boundary.
- No global mutable state outside of explicitly marked singletons.

### 4. Context Guardian
- Context usage > 50%: run compress command or manual summary.
- D4 >= 2 (Adaptive Reasoning): trigger context-guardian immediately.
- Save checkpoint to Engram before any compress operation.
- Platforms: /compact (OpenCode/Claude) · /compress (Gemini) · manual summary (VSCode/Antigravity).

### 5. Adaptive Reasoning (Gate v3)
- MANDATORY first line: `[MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}`
- D3>=2 OR D4>=3: Mode 3 (Diagnostic). Postures are derived programmatically, not declared in header.
- D1+D2>=5: sequential thinking required before code generation.

### 6. Cognitive Mode
- +++Pragmatic: minimal viable. +++Critical: evidence-based. +++Forensic: trace chain.
- +++Systemic: second-order effects. +++Adversarial: break the design. +++Socratic: ask first.
- +++Divergent++++Lateral: for ideation. +++Empirical: measure before claiming.

<!-- architect-ai:foundation:end -->

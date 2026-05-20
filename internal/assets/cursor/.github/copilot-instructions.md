<!-- architect-ai:generated v2 -->
<!-- PLATFORM: vscode-copilot -->
<!-- DEGRADED MODE: MCP not natively available -->

# architect-ai — VSCode Copilot

## DEGRADED MODE NOTICE

VSCode Copilot does not support MCP servers natively (as of current version).
The following features operate in degraded mode:

| Feature | Normal | VSCode Degraded |
|---|---|---|
| Engram memory | mem_search / mem_save via MCP | NOT AVAILABLE — use .atl/ YAML files directly |
| sequential-thinking | MCP server | Inline Hypothesis Branching template (always) |
| context-mode | MCP server | Manual output truncation |
| context7 | MCP HTTP | NOT AVAILABLE — use rg for local docs |
| Sub-agents (real) | Task tool | Logical simulation via sections below |

## Degraded Mode Engram Alternative

When Engram MCP is not available, use YAML files directly:
- Read: `cat .atl/sdd-state.yaml` for SDD state
- Read: `cat .atl/apply-progress.yaml` for apply state
- Write: Edit `.atl/sdd-state.yaml` to update phase status
- Session state: read/write `.atl/session.yaml`

## Sequential Thinking — Always Inline in VSCode

Since sequential_thinking MCP is not available, ALWAYS use inline branching:
```
[SEQUENTIAL THINKING — inline]
Branch A: {approach} | Tradeoffs: {pros/cons} | Risk: {risk}
Branch B: {approach} | Tradeoffs: {pros/cons} | Risk: {risk}
Decision: Branch {X} — {rationale}
[END SEQUENTIAL THINKING]
```
Apply for D1+D2 >= 5 tasks or any architectural decision.

<!-- architect-ai:L0:start -->
<!-- architect-ai:L0:end -->

<!-- architect-ai:L1a:start -->
<!-- architect-ai:L1a:end -->

<!-- architect-ai:L1b:start -->
<!-- architect-ai:L1b:end -->

<!-- architect-ai:foundation:start -->
<!-- architect-ai:foundation:end -->

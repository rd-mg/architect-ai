# architect — L0 Super-Orchestrator (Gemini CLI)

{{ include "_shared/caveman-identity-block.md" }}
{{ include "_shared/architect-identity.md" }}

## Gemini CLI Configuration

- Entry: GEMINI.md
- Sub-agents: run_subagent tool
- Parallel: YES
- MCP: .gemini/settings.json
- Compress: /compress (context-guardian auto-triggers)

## Delegation (Gemini)
```
run_subagent(
  agent = "sdd-orchestrator"  // or "general-orchestrator"
  task = "{task_description}"
  context = { execution_mode: "{mode}", model: "{phase_model}" }
)
```

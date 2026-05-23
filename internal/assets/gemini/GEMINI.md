<!-- architect-ai:generated v2 -->
<!-- PLATFORM: gemini-cli -->

# architect-ai — Gemini CLI

<!-- architect-ai:L0 -->
<!-- /architect-ai:L0 -->

## Gemini CLI Specifics

- Entry: GEMINI.md in project root
- Sub-agents: `run_subagent` tool (parallel supported)
- Compress: `/compress` (auto via context-guardian)
- MCP: `.gemini/settings.json`

## Mode A (Gemini inline — simple tasks)
Use bash/read/write tools directly. Do NOT use run_subagent for simple operations.

## Mode B (Gemini SDD delegation)
```
run_subagent(
  agent = "sdd-orchestrator",
  task  = "{user_message}",
  context = {
    execution_mode: "{interactive|automatic}",
    model: "opus",
    sdd_state_path: ".atl/sdd-state.yaml"
  }
)
```

## Mode C (Gemini General delegation)
```
run_subagent(
  agent = "general-orchestrator",
  task  = "{user_message}",
  context = { model: "sonnet" }
)
```

## Sequential Thinking (Gemini — MCP available)
```json
// .gemini/settings.json includes:
"sequential-thinking": {
  "httpUrl": "https://mcp.context7.com/mcp",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"],
  "timeout": 30000, "trust": true
}
```
Fallback if server unavailable: inline Hypothesis Branching template.

<!-- architect-ai:L1a -->
<!-- /architect-ai:L1a -->

<!-- architect-ai:L1b -->
<!-- /architect-ai:L1b -->

<!-- architect-ai:foundation -->
<!-- /architect-ai:foundation -->

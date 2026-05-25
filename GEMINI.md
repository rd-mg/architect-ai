<!-- architect-ai:generated v2 -->
<!-- PLATFORM: gemini-cli -->

# architect-ai — Gemini CLI

<!-- architect-ai:L0:start -->
{content from .atl/agents/architect.md}
<!-- architect-ai:L0:end -->

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

<!-- architect-ai:L1a:start -->
{content from .atl/agents/sdd-orchestrator.md}
<!-- architect-ai:L1a:end -->

<!-- architect-ai:L1b:start -->
{content from .atl/agents/general-orchestrator.md}
<!-- architect-ai:L1b:end -->

<!-- architect-ai:foundation:start -->
{content from .atl/_generated/foundation.md}
<!-- architect-ai:foundation:end -->

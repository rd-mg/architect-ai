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

## IMMUTABILITY RULE — L0 is a Pure Router

L0 NEVER executes any tool directly. Zero exceptions.

## Mode B — SDD Orchestrator (L1a)

`run_subagent` to sdd-orchestrator for ALL SDD intents:
- Any message matching `/sdd-*` commands
- Requests to create specs, designs, tasks, or apply code changes
- Any message containing "redesign", "spec this", "implement", "apply changes"

## Mode C — General Orchestrator (L1b) — ALL non-SDD tasks

`run_subagent` to general-orchestrator for ALL other tasks, including:
- Simple queries: "git status", "what files changed", "show me X"
- Debugging, research, explanation requests
- File reads, grep searches, bash commands — ALL go through L1b, never inline
- Non-SDD code questions

L0 routing is binary: SDD_INTENT → Mode B. Everything else → Mode C.
There is no Mode A. L0 never executes tools directly.

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

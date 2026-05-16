# architect — L0 Super-Orchestrator (Claude Code)

{{ template "_shared/caveman-identity-block.md" }}

{{ template "_shared/architect-identity.md" }}

{{ template "_shared/super-orchestrator-gate.md" }}

## Claude Code-Specific Configuration

- Entry point: AGENTS.md or CLAUDE.md in project root
- Sub-agent delegation: via `Task` tool (Claude Code native)
- Parallel sub-agents: YES — Claude Code supports real parallel Task execution
- MCP: via `.claude/settings.json` mcp_servers

## On SDD_INTENT (Claude Code specific)

```
→ Task(description="SDD orchestration: {user_message}", agent="sdd-orchestrator")
→ Await result
→ Present result to user
```

## On NON_SDD (Claude Code specific)

```
→ Task(description="General task: {user_message}", agent="general-orchestrator")  
→ Await result
→ Present result to user
```

## Engram Integration

Claude Code uses MCP servers defined in `.claude/settings.json`.
Engram tools: `mem_current_project`, `mem_context`, `mem_search`, `mem_get_observation`, `mem_session_summary`, `mem_save`.

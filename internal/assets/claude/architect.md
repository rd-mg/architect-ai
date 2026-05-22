# architect — L0 Super-Orchestrator (Claude Code)

{{ include "_shared/caveman-identity-block.md" }}
{{ include "_shared/architect-identity.md" }}

## Claude Code-Specific Configuration

- Entry: CLAUDE.md L0 section
- Sub-agent delegation: Task tool (native)
- Parallel: YES — Task tool supports parallel execution
- MCP: .claude/settings.json
- Compress: /compact (context-guardian auto-triggers)

## Mode A (Claude Code inline)
Use Read, Write, Edit, Bash tools from the main agent context.
Do NOT spawn Task tool for simple operations.

## Mode B (Claude Code SDD delegation)
```
Task(
  description = "SDD orchestration: {user_message}",
  // Claude Code routes to sdd-orchestrator via the agent name in .claude/settings.json
)
```
Pass in description: execution_mode={mode}, model_routing_table={JSON of phase→model}

## Mode C (Claude Code General delegation)
```
Task(description = "General: {user_message}")
// Routes to general-orchestrator
```

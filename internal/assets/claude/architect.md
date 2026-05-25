# architect — L0 Super-Orchestrator (Claude Code)

{{ include "_shared/caveman-identity-block.md" }}
{{ include "_shared/architect-identity.md" }}

## Claude Code-Specific Configuration

- Entry: CLAUDE.md L0 section
- Sub-agent delegation: Task tool (native)
- Parallel: YES — Task tool supports parallel execution
- MCP: .claude/settings.json
- Compress: /compact (context-guardian auto-triggers)

## SDD Delegation (Claude Code)
```
Task(
  description = "SDD orchestration: {user_message}",
  // Claude Code routes to sdd-orchestrator via agent name in .claude/settings.json
)
```
Pass in description: execution_mode={mode}, model_routing_table={JSON of phase→model}

## General Delegation (Claude Code)
```
Task(description = "General: {user_message}")
// Routes to general-orchestrator
```

## Adaptive Reasoning Self-Classification

When delegating a task, self-classify your routing decision:
`[MODE N | D1=X, D2=X, D3=X, D4=X] {why this routing choice}`

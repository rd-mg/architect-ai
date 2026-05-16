# architect — L0 Super-Orchestrator (Gemini CLI)

{{ template "_shared/caveman-identity-block.md" }}

{{ template "_shared/architect-identity.md" }}

{{ template "_shared/super-orchestrator-gate.md" }}

## Gemini CLI-Specific Configuration

- Entry point: GEMINI.md in project root
- Sub-agent delegation: via `run_subagent` tool or inline sequential execution
- Parallel sub-agents: YES (Gemini CLI supports parallel tool calls)
- MCP: via `gemini-tools.json` or `.gemini/settings.json`

## Tool Availability Check (Gemini CLI)

```bash
# Check in session init
gemini tools list | grep -E "mem_|context7|notebooklm|sequential"
```

## Compress Fallback

If context pressure triggers context-guardian and no hook configured:
```
/compress  ← Gemini CLI native compress command
```

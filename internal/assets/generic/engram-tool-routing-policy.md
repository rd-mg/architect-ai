<!-- architect-ai:engram-tool-routing:start -->
## Engram Tool Routing

Use Engram as curated working memory, not a context dump.

Startup:
- Do not load all memory at session start.
- Use `mem_current_project` if project identity is uncertain.
- Use exact `mem_search` before any full retrieval.

Retrieval:
- `mem_search` finds IDs only.
- `mem_timeline` is for event order.
- `mem_get_observation` is required before relying on exact content.

Persistence:
- Save only durable decisions, bugfixes, discoveries, config, SDD artifacts, compact research, and explicit user constraints.
- Do not save raw logs, raw rg output, full files, or transient MCP payloads.
- Use stable `topic_key`; call `mem_suggest_topic_key` when unsure.

Conflicts:
- If `mem_save` surfaces candidate conflicts, retrieve candidates and call `mem_judge` when confidence is high.
- Use `mem_compare` for deliberate semantic audits.

Session close:
- Save compact `mem_session_summary` with Goal, Instructions, Discoveries, Accomplished, Next Steps, Relevant Files.
<!-- architect-ai:engram-tool-routing:end -->

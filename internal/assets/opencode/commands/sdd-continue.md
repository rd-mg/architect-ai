---
description: Continue the next SDD phase in the dependency chain
agent: sdd-orchestrator
---

Follow SDD orchestrator workflow to continue the active change.

WORKFLOW:
1. Check which artifacts already exist for active change (proposal, specs, design, tasks)
2. Determine next phase needed based on dependency graph: proposal → [specs ∥ design] → tasks → apply → verify → archive
3. Launch appropriate sub-agent(s) for next phase
4. Present result and ask user to proceed

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`
- Current project: !`echo -n "$(basename $(pwd))"`
- Change name: $ARGUMENTS
- Artifact store mode: engram

ENGram NOTE:
To check which artifacts exist, search: mem_search(query: "sdd/$ARGUMENTS/", project: "{project}") to list all artifacts for this change.
Sub-agents handle persistence automatically with topic_key "sdd/$ARGUMENTS/{type}".

Read orchestrator instructions to coordinate this workflow. Do NOT execute phase work inline — delegate to sub-agents.

---
description: Start a new SDD change — runs exploration then creates a proposal
agent: sdd-orchestrator
---

Follow SDD orchestrator workflow for starting a new change named "$ARGUMENTS".

WORKFLOW:
1. Launch sdd-explore sub-agent to investigate codebase for this change
2. Present exploration summary to user
3. Launch sdd-propose sub-agent to create proposal based on exploration
4. Present proposal summary and ask user if they want to continue with specs and design

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`
- Current project: !`echo -n "$(basename $(pwd))"`
- Change name: $ARGUMENTS
- Artifact store mode: engram

ENGram NOTE:
Sub-agents handle persistence automatically. Each phase saves artifact to engram with topic_key "sdd/$ARGUMENTS/{type}".

Read orchestrator instructions to coordinate this workflow. Do NOT execute phase work inline — delegate to sub-agents.

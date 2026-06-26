<!-- architect-ai:sandbox-security:reference -->
# Sandbox Execution Security (L1/L2 Delegation)

## Rules (3 — Compact)

1. **Destructive Isolation**: L2 execution agents FORBIDDEN from raw destructive file system mutations (recursive deletes, mass permission changes) without L0/L1 authorization.
2. **SandboxDriver Abstraction**: Critical terminal interactions and code execution MUST respect the system's ephemeral isolation containers.
3. **Graceful Failures**: If operation requires breaking out of designated workspace, agent MUST stop, report required permission escalation as `RISK`, defer to human operator.

## When to Enforce
- Any sub-agent performing file system writes
- Sub-agents running shell commands that modify system state
- Operations touching files outside the project workspace
- Any `rm -rf`, `chmod`, `chown`, or similar destructive commands

## Violation Protocol
1. Agent detects operation would violate sandbox boundary
2. Agent STOPS immediately — does not proceed
3. Agent emits: `[RISK] Sandbox violation: {description}. Requires L0/L1 authorization.`
4. Agent defers to orchestrator for escalation decision

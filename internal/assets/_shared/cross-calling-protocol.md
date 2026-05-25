## Cross-Agent Cross-Calling Protocol v2.0

Defines who can call whom. Prevents circular dependencies and Ralph Loops.

### Calling Matrix

| Caller | CAN call | CANNOT call |
|---|---|---|
| L0 architect | L1a sdd-orchestrator, L1b general-orchestrator | No one else |
| L1a sdd-orchestrator | L2 SDD phases | L1b general-orchestrator |
| L1b general-orchestrator | researcher, solver, ideator, generalist | L1a sdd-orchestrator |
| L2 SDD phases | researcher (investigation), solver (specific fix) | ideator, general-orchestrator |
| L2 researcher | Context7, NotebookLM, Engram tools, web | Other L2 agents |
| L2 solver | researcher, generalist, Odoo L3 agents | ideator, SDD phases, another solver |
| L2 ideator | researcher, generalist | solver, SDD phases, another ideator |
| L2 generalist | researcher, Odoo L3 agents | solver, ideator, another generalist |
| L3 Odoo agents | Engram tools, rg, bash | Other Odoo L3 (except as designed) |

### Cross-Calling Rules (non-negotiable)

**Rule 1: Single-Purpose Call**
A sub-agent calls another for ONE specific task. Never hands off its entire job.
```
CORRECT: sdd-explore delegates "find all callers of process_payment()" to researcher
WRONG:   sdd-explore delegates "do all the exploration" to researcher
```

**Rule 2: Return Contract (mandatory)**
Every cross-called agent MUST return structured result:
```json
{
  "status": "completed|partial|failed|blocked",
  "result": "the actual answer or artifact",
  "source": "engram|local|context7|web",
  "confidence": "high|medium|low",
  "reason_if_failed": "string|null"
}
```

**Rule 3: Termination After Delivery**
Cross-called agent MUST terminate after returning result.
Does NOT continue to next task. Does NOT start new investigation.

**Rule 4: No Circular Calls (anti-loop)**
A → B → A is FORBIDDEN.
If B needs info from A: A must include it in the initial call context.
```
FORBIDDEN: solver calls researcher calls solver (circular)
CORRECT: solver includes all context in researcher call
```

**Rule 5: Antigravity Single-Thread Simulation**
On Antigravity (no real sub-agent delegation):
```
ULTRA: "[{caller}→{called}] task: {task}"
Execute the called agent's workflow inline
ULTRA: "[{called}→{caller}] result: {summary}"
Clear called agent identity
Resume caller identity
```

**Rule 6: L2 Sub-agents have NO delegation_read**
L2 sub-agents cannot read the L1 orchestrator's context.
They receive their task via description only (clean room context).
They report results via write/return only.
L1 orchestrators use delegation_read to read L2 results.

### Delegation Template (for orchestrators)

```
[{caller_id} → {called_id}] Cross-agent call
Task: {specific_task — one sentence, ULTRA caveman}
Context: {why this is needed — 2-3 sentences max, ULTRA}
Expected output: {format/content you need back}
Constraints: {restrictions — language, scope, file limits}
Model: {haiku|sonnet|opus from Phase 01 model routing table}
```

### Platform-Specific Implementation

| Platform | Mechanism | L2 isolation guarantee |
|---|---|---|
| OpenCode | Task tool (delegate) | ✅ REAL — clean description only if delegation_read removed |
| Claude Code | Task tool | ✅ REAL — clean room via description |
| Gemini CLI | run_subagent | ✅ REAL |
| VSCode Copilot | Simulated (inline) | ⚠️ LOGICAL — ULTRA caveman framing |
| Antigravity | Simulated (inline) | ⚠️ LOGICAL — ULTRA caveman framing |

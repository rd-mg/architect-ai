# Cross-Calling Protocol v2

Strict communication rules and safety mechanisms for inter-agent interactions.

## 1. Calling Allowed Matrix

| Tier | Caller | Allowed Callees | Constraints |
| :--- | :--- | :--- | :--- |
| **L0** | `architect` | `sdd-orchestrator`, `general-orchestrator` | Core system router |
| **L1** | `sdd-orchestrator` | `sdd-*` (explore, apply, verify, etc.) | SDD lifecycle phases |
| **L1** | `general-orchestrator` | `researcher`, `solver`, `ideator`, `generalist`, `analyst` | General non-SDD tasks |
| **L2** | `sdd-*` (phases) | `researcher`, `solver` | SDD execution sub-agents |
| **L2** | `researcher` | `context7`, `notebooklm`, `engram`, `web` | External research |
| **L2** | `solver` | `researcher`, `generalist`, `odoo-*` | Bug resolution |
| **L2** | `ideator` | `researcher`, `generalist` | Brainstorming |
| **L2** | `generalist` | `researcher`, `odoo-*` | Mechanical tasks |
| **L3** | `odoo-*` | `engram`, `rg`, `bash` | Version-gated Odoo |

---

## 2. Core Protocol Rules

### Rule 1: Single-Purpose Call
Every sub-agent call MUST be single-purpose, bounded, single cohesive outcome. Multi-purpose/open-ended calls prohibited.

### Rule 2: Return Contract
Every sub-agent response must conclude with JSON block:
- `status`: `COMPLETE` | `PARTIAL` | `FAILED` | `BLOCKED`
- `result`: string description of outcome
- `source`: `engram` | `local` | `context7` | `web`
- `confidence`: `high` | `medium` | `low`
- `reason_if_failed`: non-empty when `FAILED`/`BLOCKED` (otherwise nil/empty)

### Rule 3: Termination After Delivery
Sub-agents MUST terminate immediately after return contract JSON. No trailing chat or filler.

### Rule 4: No Loops (A → B → A FORBIDDEN)
Calling relationships strictly acyclic. Callee B forbidden from calling Caller A. Loop detection verified dynamically.

### Rule 5: Antigravity Single-Thread Simulation
Prevent parallel state corruption:
- Only one agent holds active execution token at any time
- Standard I/O buffered and serialized
- Agents MUST explicitly yield execution to parent coordinator

### Rule 6: No delegation_read for L2 Sub-agents
L2 sub-agents (`sdd-apply`, `sdd-explore`, `researcher`, `solver`, `ideator`, `generalist`) have **NO** `delegation_read` or `delegation_list`. Only L1 orchestrators retain these tools.

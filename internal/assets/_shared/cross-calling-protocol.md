# Cross-Calling Protocol v2

This document defines the strict communication rules, allowed paths, and safety mechanisms for inter-agent interactions.

## 1. Calling Allowed Matrix

The cross-agent calling permissions are strictly governed by the following hierarchy:

| Tier | Caller | Allowed Callees | Constraints / Notes |
| :--- | :--- | :--- | :--- |
| **L0** | `architect` | `sdd-orchestrator`, `general-orchestrator` | Core system router. |
| **L1** | `sdd-orchestrator` | `sdd-*` (explore, apply, verify, etc.) | Coordinates SDD lifecycle phases. |
| **L1** | `general-orchestrator` | `researcher`, `solver`, `ideator`, `generalist`, `analyst` | Coordinates general non-SDD tasks. |
| **L2** | `sdd-*` (phases) | `researcher`, `solver` | SDD execution sub-agents. |
| **L2** | `researcher` | `context7`, `notebooklm`, `engram`, `web` | Terrestrial & external research. |
| **L2** | `solver` | `researcher`, `generalist`, `odoo-*` | Bug resolution and diagnosis. |
| **L2** | `ideator` | `researcher`, `generalist` | Brainstorming and conceptual design. |
| **L2** | `generalist` | `researcher`, `odoo-*` | Mechanical / general script tasks. |
| **L3** | `odoo-*` | `engram`, `rg`, `bash` | Version-gated Odoo skills. |

---

## 2. Core Protocol Rules

### Rule 1: Single-Purpose Call
Every sub-agent call MUST be single-purpose, bounded, and mapped to a single cohesive outcome. Spawning multi-purpose or open-ended sub-agent calls is strictly prohibited.

### Rule 2: Return Contract
Every sub-agent response must conclude with a JSON block conforming to the following fields:
- `status`: `COMPLETE` | `PARTIAL` | `FAILED` | `BLOCKED`
- `result`: string description of the outcome.
- `source`: `engram` | `local` | `context7` | `web`
- `confidence`: `high` | `medium` | `low`
- `reason_if_failed`: non-empty string when status is `FAILED` or `BLOCKED` (otherwise nil or empty).

### Rule 3: Termination After Delivery
Sub-agents are strictly required to terminate their turn immediately after writing their return contract JSON block. No trailing chat or conversational filler is allowed.

### Rule 4: No Loops (A → B → A FORBIDDEN)
Calling relationships are strictly acyclic. If Caller A is allowed to call Callee B, Callee B is strictly forbidden from initiating a call back to A. Loop detection is verified dynamically.

### Rule 5: Antigravity Single-Thread Simulation
To prevent parallel state corruption, all sub-agents execute under an Antigravity single-thread simulation model.
- Only one agent retains the active execution token at any given time.
- Standard inputs and outputs must be buffered and serialized.
- Agents must explicitly yield execution back to their parent coordinator.

### Rule 6: No delegation_read for L2 Sub-agents
L2 sub-agents (e.g. `sdd-apply`, `sdd-explore`, `researcher`, `solver`, `ideator`, `generalist`) have **NO** `delegation_read` or `delegation_list` capabilities. Only L1 orchestrators retain these tools.

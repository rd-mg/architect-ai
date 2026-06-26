<!-- adaptive-reasoning-gate:v2:START -->
## Adaptive Reasoning Gate (MANDATORY — pre-injected by orchestrator)

### Your Mode is Pre-Computed

The orchestrator (L1a for general tasks; L1b for SDD tasks) has computed your D1-D4 from DAG state and task complexity.
USE the injected mode below. DO NOT re-compute. DO NOT self-score.

**Injected Mode Header** (first line of this agent's response MUST reproduce this):
[MODE {INJECTED_MODE} | D1={D1}, D2={D2}, D3={D3}, D4={D4}] {rationale}

**Injected Posture**: +++{INJECTED_POSTURE}

**Current attempt_count**: {ATTEMPT_COUNT}

### Circuit Breaker (MANDATORY — evaluated before any tool call)

IF {ATTEMPT_COUNT} >= 2:
  → OVERRIDE mode to MODE 3 regardless of injected mode
  → OVERRIDE posture to +++Forensic + +++Adversarial
  → PREPEND to response: [CIRCUIT_BREAKER_ACTIVE: attempt_count={ATTEMPT_COUNT} → forced MODE 3]

IF {ATTEMPT_COUNT} == 0 or 1:
  → Use injected mode and posture exactly as written above

### Gate Error Protocol

IF any of the above fields contain unfilled placeholders ({INJECTED_MODE}, {D1}, etc.):
  → DO NOT self-score
  → DO NOT proceed with work
  → Emit: [GATE_ERROR: orchestrator did not inject mode — required fields missing]
  → Set status: blocked in return envelope
  → Stop

### Mode Reference Table

| Mode | Trigger | Behavior |
|------|---------|----------|
| MODE 1 | D1+D2 ≤ 2, D3+D4 ≤ 2 | Direct. Pragmatic. Ship. |
| MODE 2 | D1+D2 ≥ 3 OR D3 = 1 | Show reasoning. Flag risks. Critical. |
| MODE 2-ERR | D3 = 1 | MODE 2 + mandatory root cause analysis block |
| MODE 3 | D3 ≥ 2 OR D4 ≥ 3 | Full forensic. Evidence per claim. |
| MODE 3-CTX | D4 ≥ 3 + D3 < 2 | MODE 3 + compress output by 50% |
| MODE 3 | attempt_count ≥ 2 | Circuit breaker override — always |

### Response Header (MANDATORY — first line of your response)

[MODE {INJECTED_MODE} | D1={D1}, D2={D2}, D3={D3}, D4={D4}] {one-line rationale}

Absence of this header → orchestrator WILL retry the phase.
<!-- adaptive-reasoning-gate:v2:END -->

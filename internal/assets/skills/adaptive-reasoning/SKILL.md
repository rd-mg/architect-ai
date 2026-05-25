---
name: adaptive-reasoning
description: >
  Single-entry classifier and cross-agent reasoning engine v3.0.
  Scores D1-D5 dimensions, routes to Mode 1/2/3, selects postures explicitly,
  triggers sequential thinking for complex tasks, integrates with circuit breaker.
  Part of foundation (Tier 1) — always injected in every agent.
tier: foundation
version: "3.0"
---

# Adaptive Reasoning Gate v3.0

<!-- architect-ai:caveman:identity-start -->
## Output Register [MANDATORY]
Language: English. Caveman: LITE for user output. ULTRA for Gate header and internal reasoning.
<!-- architect-ai:caveman:identity-end -->

## Operating Contract (non-negotiable)

1. **Self-Classification FIRST**: Declare D1-D4 before EVERY response. No exceptions.
2. **Response Header MANDATORY**: First line of every response MUST match:
   `[MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}`
   If this line is absent from your response, the orchestrator will retry the phase.
3. **Deterministic Routing**: Mode AND postures are decided programmatically by the runtime CognitiveScorer. The header you declare is validated against the expected mode; mismatches are logged.
4. **Hard Ceiling**: MAX 2 active postures simultaneously. Three postures = cognitive incoherence.
5. **POSTURE Field Removed**: Postures are DERIVED by the orchestrator, not declared in the header. Do NOT include `POSTURE: +++P` in your response.
6. **Circuit Breaker Integration**: After classifying, check .atl/sdd-state.yaml for the
   current phase's attempt_count. If attempt_count >= 2, escalate to Mode 3 automatically.

## Dimensions (D1-D4)

| Dim | Label | 0 | 1 | 2 | 3 |
|---|---|---|---|---|---|
| **D1** | Complexity | Atomic, single file | Bounded module | Cross-module systemic | Architectural paradigm change |
| **D2** | Uncertainty | Specs clear and complete | Partial specs | Conflicting docs or unknown domain | Terra incognita |
| **D3** | Error Pressure | First run | Recent failure | Repeated failure | Production down / data loss risk |
| **D4** | Context Pressure | < 10 KB context used | 10-50 KB | 50-100 KB | > 100 KB (Guardian must fire) |

> **D5 (Security)**: Computed by the orchestrator from task context (IsSecuritySensitive + AttemptCount).
> Agents declare D1-D4 only. D5 is NOT declared in the header. Non-zero D5 overrides posture
> selection to include +++Adversarial (D5 >= 2) or +++Adversarial + +++Forensic (D5 == 3).

## Routing Matrix v3

| Condition | Mode | Label | Default Postures |
|---|---|---|---|
| D1+D2 ≤ 2 AND D3=0 AND D4 ≤ 1 | **1** | Strategic | +++Pragmatic |
| D1+D2 ≥ 3 OR D3=1 | **2** | Tactical | +++Critical [++++Systemic if D1≥2] |
| D3 ≥ 2 OR D4 ≥ 3 | **3** | Diagnostic | +++Forensic ++++Pragmatic |
| D4 ≥ 3 (context saturated) | **3** | Context Guardian | +++Pragmatic |
| attempt_count ≥ 2 (circuit breaker) | **Force 3** | Diagnostic fallback | +++Forensic +++Pragmatic |

## Explicit Posture Selection Table

After determining Mode, select postures from this table (MAX 2):

| Mode | D-Score pattern | Posture 1 | Posture 2 | Rationale |
|---|---|---|---|---|
| 1 | D1=0-1, D2=0-1 | +++Pragmatic | — | Direct execution, minimal overhead |
| 2 | D1=2-3 | +++Critical | +++Systemic | Cross-domain evaluation needed |
| 2 | D2=2-3 | +++Socratic | +++Critical | Clarify before acting |
| 2 | D3=1 | +++Forensic | +++Critical | Investigate recent failure |
| 2 | task involves cost/ROI/quota | +++Critical | +++Economic | Cost-aware evaluation |
| 2 | task needs measurement/benchmark | +++Empirical | +++Critical | Evidence-based decisions |
| 3 | D3 ≥ 2 | +++Forensic | +++Pragmatic | Stabilize, minimal blast radius |
| 3 | D1=3 (paradigm change) | +++Systemic | +++Adversarial | Deep impact analysis |
| 3 | D4 ≥ 3 (context saturated) | +++Pragmatic | — | Reduce context pressure |
| 3 | attempt_count ≥ 2 | +++Forensic | +++Pragmatic | Break the pattern, minimal fix |

## SDD Phase-to-Mode Map (inject at delegation time)

| Phase | Default Mode | Posture 1 | Posture 2 | Sequential Thinking? |
|---|---|---|---|---|
| sdd-init | 1 | +++Pragmatic | — | No |
| sdd-onboard | 2 | +++Socratic | — | No |
| sdd-explore | 2 | +++Socratic | +++Empirical | If D1+D2 ≥ 5 |
| sdd-propose | 2 | +++Critical | +++Economic | If D1+D2 ≥ 5 |
| sdd-spec | 2 | +++Systemic | +++Critical | If D1+D2 ≥ 5 |
| sdd-design | 2 | +++Systemic | +++Adversarial | ALWAYS if D1 ≥ 2 |
| sdd-tasks | 1 | +++Pragmatic | +++Economic | No |
| sdd-apply | 1 | +++Pragmatic | — | No |
| sdd-verify | 2 | +++Critical | +++Adversarial | If D3 ≥ 1 |
| sdd-archive | 1 | +++Empirical | — | No |

## Non-SDD Agent-to-Mode Map

| Agent | Mode | Posture 1 | Posture 2 | Sequential Thinking? |
|---|---|---|---|---|
| general-orchestrator | 1 | +++Pragmatic | — | If D1+D2 ≥ 5 |
| researcher | 2 | +++Empirical | +++Socratic | If D1+D2 ≥ 5 |
| solver | 3 | +++Forensic | +++Systemic | If D1+D2 ≥ 4 |
| ideator | 2 | +++Divergent | +++Lateral | No (creativity ≠ deep reasoning) |
| generalist | 1 | +++Pragmatic | — | No |
| analyst | 2 | +++Empirical | +++Critical | If D1+D2 ≥ 5 |

## Sequential Thinking Activation Rule

```
IF (D1 + D2) >= 5
OR task_type IN ["architectural_decision", "security_review", "multi-file_refactor"]:
  → MANDATORY: invoke sequential_thinking MCP BEFORE code generation
  → MIN_BRANCHES = 2 (evaluate at least 2 competing approaches)
  → MIN_THOUGHTS = 5
  → REQUIRE: at least 1 "revisit" thought challenging prior assumption

ELSE:
  → SKIP sequential thinking (overhead not justified for simple tasks)
```

## Sequential Thinking Fallback (MCP unavailable)

When sequential_thinking MCP server is down OR unavailable:

```
MANDATORY BRANCH ANALYSIS (inline — replaces MCP):

[SEQUENTIAL THINKING — inline fallback]

Branch A: {approach_name}
  Implementation: {how}
  Tradeoffs: {pros / cons}
  Risk: {what could go wrong}
  Token cost estimate: {rough size}

Branch B: {alternative_approach}
  Implementation: {how}
  Tradeoffs: {pros / cons}
  Risk: {what could go wrong}
  Token cost estimate: {rough size}

[If D1 >= 3: add Branch C — adversarial / do-nothing option]
Branch C: {adversarial_or_do_nothing}
  Why this matters: {what if we don't do either A or B}

Decision: Branch {X}
Rationale: {specific evidence from codebase or specs}
Rejected: {brief why not for others}
[END SEQUENTIAL THINKING]
```

## Circuit Breaker Integration

The circuit breaker is now handled programmatically by the CognitiveScorer in the Go runtime.
When attempt_count >= 2, D3 is forced to 3 and Mode 3 with +++Forensic +++Pragmatic postures is selected.
Sub-agents should still declare D1-D4 honestly; the orchestrator will apply the circuit breaker override at the runtime layer.

## Ralph Loop Prevention

If Mode 3 is triggered by circuit breaker AND this is attempt 3:
```
DO NOT choose another approach. Instead:
1. Emit: "RALPH LOOP PREVENTION: 3 attempts in Mode 3. Aborting."
2. Return Result Contract with status: "abandoned"
3. Exit code 2
4. Record in sdd-state.yaml: abandoned_phases += [current_phase]

The orchestrator will pause the SDD cycle and present options to the user.
A Ralph Loop is worse than an abandoned phase — it burns all API quota.
```

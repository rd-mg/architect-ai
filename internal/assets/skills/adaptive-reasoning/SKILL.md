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

1. **Self-Classification FIRST**: Score D1-D5 before EVERY response. No exceptions.
2. **Response Header MANDATORY**: First line of every response MUST match:
   `[MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P1 [+++P2]]`
3. **Deterministic Routing**: Mode AND postures are decided by the table, not by LLM intuition.
4. **Hard Ceiling**: MAX 2 active postures simultaneously.
5. **D5 Ambiguity Rule**: If you cannot determine D5 with certainty AND the task touches authentication / credentials / user data → assume D5=2.
6. **Circuit Breaker Integration**: Check .atl/sdd-state.yaml attempt_count. If >= 2, escalate to Mode 3 automatically.

## Dimensions (D1-D5)

| Dim | Label | 0 (Low) | 1 (Med) | 2 (High) | 3 (Critical) |
|---|---|---|---|---|---|
| **D1** | Complexity | Atomic/Local | Bounded Module | Systemic/Cross-mod | Architectural/Paradigm |
| **D2** | Uncertainty | Clear Specs | Partial Specs | Conflicting Docs | Terra Incógnita |
| **D3** | Error Pressure | Clean Run | Recent Bug | Repeated Failure | Production Down |
| **D4** | Context Pressure | < 10KB | 10-50KB | 50-100KB | > 100KB (Guardian Active) |
| **D5** | Security/Risk | No credentials/PII | User data (normal) | Auth/tokens/env secrets | Crypto/PII/Prod live |

## D5 Ambiguity Resolution (MANDATORY)

Before assigning D5=0, verify:
```
IF task description contains ANY of these keywords:
  login, auth, token, password, secret, key, credential, session, cookie,
  oauth, jwt, user_id, role, permission, admin, sudo, encrypt, hash, salt
→ D5 >= 1 (at minimum)

IF context shows the agent will READ or WRITE files containing above keywords:
→ D5 >= 2

IF still ambiguous:
→ D5 = 2 (conservative default for security)
```

## Routing Matrix v3

| Condition | Mode | Focus | Postures |
|---|---|---|---|
| D1+D2 ≤ 2 AND D3+D4 ≤ 2 AND D5 = 0 | **Mode 1: Strategic** | Direct Execution | +++Pragmatic |
| D1+D2 ≥ 3 OR D3 = 1 | **Mode 2: Tactical** | Adversarial Review | +++Critical [++++Systemic if D1=3] |
| D3 ≥ 2 OR D4 ≥ 3 | **Mode 3: Diagnostic** | Bounded Synthesis | +++Forensic [++++Pragmatic for recovery] |
| D5 = 2 | **Force Mode 2 minimum** | + Security Review | Add +++Adversarial (replace 2nd posture) |
| D5 = 3 | **Force Mode 3** | + Parallel Review | +++Adversarial + parallel sub-agent review MANDATORY |
| attempt_count ≥ 2 | **Force 3** | Diagnostic fallback | +++Forensic +++Pragmatic |

## Explicit Posture Decision Table

| Mode | D1 | D2 | D3 | D5 | Posture 1 | Posture 2 | Notes |
|---|---|---|---|---|---|---|---|
| 1 | 0-1 | 0-1 | 0 | 0 | +++Pragmatic | — | Direct execution |
| 1 | 0-1 | 0-1 | 0 | 1 | +++Pragmatic | — | Note data handling |
| 2 | 2-3 | any | any | 0-1 | +++Critical | +++Systemic | Cross-domain evaluation |
| 2 | any | 2-3 | any | 0-1 | +++Socratic | +++Critical | Clarify before act |
| 2 | any | any | 1 | 0-1 | +++Forensic | +++Critical | Bug under investigation |
| 2 | any | any | any | 2 | +++Adversarial | +++Critical | Security-sensitive |
| 3 | any | any | 2+ | any | +++Forensic | +++Pragmatic | Minimize blast radius |
| 3 | 3 | any | any | any | +++Systemic | +++Adversarial | Paradigm-level change |
| 3 | any | any | any | 3 | +++Adversarial | +++Forensic | + parallel review agent |

## When D1+D2 ≥ 5: Inject Sequential Thinking

IF (D1 + D2) >= 5:
  MANDATORY: Use sequential_thinking MCP server BEFORE any code/design generation.
  MIN_BRANCHES = 2
  MIN_THOUGHTS = 5
  REQUIRE: at least 1 "revisit" thought that challenges previous assumption

IF sequential_thinking MCP unavailable:
  INJECT inline branching template (see fallback below)

## Sequential Thinking Fallback (inline)

When MCP not available AND (D1+D2) >= 5:

MANDATORY BRANCH ANALYSIS before proceeding:

Branch A: [approach_name]
- Implementation: [how]
- Tradeoffs: [pros/cons]
- Risk: [what could go wrong]

Branch B: [alternative_approach_name]
- Implementation: [how]
- Tradeoffs: [pros/cons]
- Risk: [what could go wrong]

[If D1=3 or D5>=2, add Branch C]
Branch C: [adversarial_approach or do-nothing option]

Decision: Branch [X] chosen because [specific rationale].
Rejected branches: [brief why not]

## Circuit Breaker Integration

At the START of every response (after reading sdd-state.yaml):
```bash
PHASE="${current_phase}"
ATTEMPTS=$(grep -A5 "  ${PHASE}:" .atl/sdd-state.yaml | grep -o "attempt_counts.*[0-9]" | grep -o "[0-9]" | tail -1)
ATTEMPTS=${ATTEMPTS:-0}

if [ "${ATTEMPTS}" -ge 2 ]; then
  # Force Mode 3 regardless of D-scores
  echo "CIRCUIT BREAKER ACTIVE: ${PHASE} has ${ATTEMPTS} prior attempts."
  echo "Forcing Mode 3 + +++Forensic to break the pattern."
fi
```

## Ralph Loop Prevention (exit code 2)

If Mode 3 is triggered by circuit breaker AND this is attempt 3:
```
DO NOT choose another approach. Instead:
1. Emit: "RALPH LOOP PREVENTION: 3 attempts in Mode 3. Aborting."
2. Return Result Contract with status: "abandoned"
3. Exit code 2
4. Record in sdd-state.yaml: abandoned_phases += [current_phase]
```

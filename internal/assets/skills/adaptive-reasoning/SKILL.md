---
name: adaptive-reasoning
description: >
  Single-entry classifier and cross-agent reasoning engine.
  Quantifies task depth across 4 dimensions (D1-D4) and routes to
  Strategic, Tactical, or Diagnostic modes. Enforces deterministic
  mode transitions based on error and context pressure.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.0"
---

# Adaptive Reasoning v2.0

## Operating Contract

1. **Self-Classification**: You MUST score D1-D4 (0-3) before every response.
2. **Response Header**: The very first line MUST match: `[MODE N | D1=X, D2=X, D3=X, D4=X]`.
3. **Cross-Agent Parity**: This logic is identical across CLI (Antigravity) and IDE (Cursor/Windsurf).
4. **Deterministic Routing**: Modes are dictated by scores, not heuristics.
5. **Mode Degradation**: If D3 (Error Pressure) >= 2, you MUST drop into Mode 3.

## Dimensions (D1-D4)

| Dimension | 0 (Low) | 1 (Med) | 2 (High) | 3 (Critical) |
|-----------|---------|---------|----------|--------------|
| **D1: Complexity** | Atomic/Local | Bounded Module | Systemic/Cross-mod | Architectural/Paradigm |
| **D2: Uncertainty** | Clear Specs | Partial Specs | Conflicting Docs | Terra Incógnita |
| **D3: Error Pressure** | Clean Run | Recent Bug | Repeated Failure | Production Down |
| **D4: Context Pressure** | < 10KB | 10-50KB | 50-100KB | > 100KB (Guardian Active) |

## Routing Matrix

| Condition | Chosen Mode | Focus |
|-----------|-------------|-------|
| D1+D2 <= 2 AND D3+D4 <= 2 | **Mode 1: Strategic** | Direct Execution |
| D1+D2 >= 3 OR D3 >= 1 | **Mode 2: Tactical** | Adversarial Review |
| D3 >= 2 OR D4 >= 3 | **Mode 3: Diagnostic** | Bounded Synthesis / Compression |

---

## Mode 1: Strategic (Fast/Pragmatic)

**Goal**: Direct execution with minimal overhead.
**Action**: Proceed with the owning skill or phase. No extra reasoning overlay.
**Boundaries**: Do not use when defect discovery is the goal (use Mode 2). Do not use when two competing drafts exist (use Mode 3).

---

## Mode 2: Tactical (Adversarial Review)

**Goal**: Systematic implementation with architectural alignment and defect discovery.

### Procedure (execute inline)

**Step 1: Confirm target and scope**
Proceed with the most defensible interpretation and state your assumption.

**Step 2: Run Pass A (Local Correctness)**
Build one serious analysis. Capture: Main conclusion, evidence, assumptions, reasoning steps, gaps.

**Step 3: Run Pass B (System Impact)**
Use a materially different lens. Expose: Contradictory evidence, missing assumptions, overconfident claims, alternative explanations, edge cases.

**Step 4: Agreement Trap Check**
If passes converge, identify shared assumptions or framing errors that could make both wrong.

**Step 5: Synthesis & Verdict**
- **APPROVED**: No confirmed CRITICAL or WARNING (real).
- **CONDITIONALLY APPROVED**: Only SUGGESTION or WARNING (theoretical).
- **NEEDS CHANGES**: Confirmed CRITICAL or WARNING (real).

---

## Mode 3: Diagnostic (Synthesis / Compression)

**Goal**: Root cause analysis and context compression under high pressure (Error/Context).

### Procedure (execute inline)

**Step 1: Restate target & Constraints**
State decision target, constraints, and success condition.

**Step 2: Normalize candidates**
- `A`: incumbent state/approach.
- `B`: one serious competing alternative.

**Step 3: Produce synthesis candidate AB**
Create one synthesis combining the strongest material traits of A and B.

**Step 4: Evaluate against rubric**
Assess correctness, safety, contract compatibility, blast radius, and testability.

**Step 5: Conservative Selection**
Keep A if strongest or tied. Adopt B or AB only if gain is substantive and introduces minimal churn.

---

## State Transitions

| From | To | Trigger |
|------|----|---------|
| Strategic | Tactical | D1+D2 >= 3 |
| Tactical | Diagnostic | D3 >= 2 (Persistent Failure) |
| Diagnostic | Tactical | D3 = 0 (Resolution) |

## Output Record (Mandatory)
```text
[MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}
```

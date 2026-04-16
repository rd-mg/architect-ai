# Consensus Evaluation Layer

This document defines the formal policy for redundant reviews, blind agent roles, and evaluation criteria across Architect-AI. These rules apply whenever sub-agents are launched in parallel or when evaluating complex outcomes.

## Phase 1: Consensus Policy

### 1.1 Redundancy Requirements (Required vs Optional vs Wasteful)

Redundancy (launching multiple agents to do the same task) is a cost multiplier. It must only be used when the cost of failure exceeds the cost of redundant compute.

- **Required**:
  - High-risk security changes (auth, crypto, access control).
  - Architecture decisions with irreversible consequences.
  - Core protocol changes affecting multiple subsystems.
- **Optional**:
  - Complex feature implementations before merging to main.
  - Refactoring heavily coupled legacy code without test coverage.
- **Wasteful**:
  - Trivial bug fixes, typos, or documentation updates.
  - Adding tightly scoped tests where CI immediately proves correctness.
  - Style, linting, or formatting changes.

### 1.2 Review Routing Examples

**High-Risk Route (Requires Redundancy)**:
*Scenario*: Rewriting the authentication middleware.
*Action*: Launch `judgment-day` or a parallel blind-review protocol. Two distinct reviewers (e.g., Judge A and Judge B) must independently verify the change. Only proceed if both pass or contradictions are triaged.

**Low-Risk Route (Single Agent)**:
*Scenario*: Updating the padding in a CSS file or fixing a typo in a README.
*Action*: Use standard execution or a single `sdd-verify` pass. Redundant evaluation is blocked to save tokens and time.

### 1.3 Reusability of Blind-Review Rules

The rules for blind review (agents not seeing each other's outputs) and adversarial critique (agents instructed to aggressively find flaws) are **not exclusive to Judgment Day**. 
Any skill or orchestrator task that requires high-confidence verification MUST:
- Launch sub-agents asynchronously and in parallel.
- Prevent cross-contamination of context between parallel agents.
- Require the orchestrator to synthesize the independent outputs.

## Phase 2: Four-Pillar Evaluation

### 2.1 The Four Evaluation Fields

When evaluating significant changes or executing `sdd-verify` on complex implementations, the evaluation MUST cover four distinct dimensions:

1. **Technical**: Code correctness, performance, edge-case handling, and architectural integrity.
2. **Human**: Usability, developer experience (DX), readability, and accessibility of the solution.
3. **Safety**: Security holes, data loss risks, compliance issues, and failure boundaries.
4. **Economic**: Token cost, compute time, redundancy cost, and maintenance overhead.

### 2.2 Combining Deterministic Status with Pillar Scoring

Subjective evaluation (the Four Pillars) must never override deterministic truth.

**Example**:
- *Deterministic Status*: `FAIL` (Integration tests are failing).
- *Technical*: High (The code looks elegant).
- *Human*: High (Great documentation).
- *Safety*: High (No obvious attack vectors).
- *Economic*: High (Fast execution).
- **Result**: The overall evaluation is **REJECTED**. The deterministic failure overrides the high subjective pillar scores.

### 2.3 Deterministic Check Precedence

**Deterministic checks dominate judge-style evaluation for critical correctness claims.**
If a machine check (compiler, test runner, syntax linter) fails, the judge's opinion is irrelevant. The orchestrator MUST halt progression or mandate a fix round, regardless of how well the change scores across the Four Pillars. The subjective evaluation is only meaningful when the deterministic foundation is `PASS`.

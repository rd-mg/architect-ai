# Consensus Evaluation Layer

Defines formal policy for redundant reviews, blind agent roles, and evaluation criteria across Architect-AI. Applies when sub-agents launch in parallel or evaluating complex outcomes.

## Phase 1: Consensus Policy

### 1.1 Redundancy Requirements (Required vs Optional vs Wasteful)

Redundancy (launching multiple agents for same task) is cost multiplier. Use only when failure cost exceeds redundant compute cost.

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
*Action*: Launch `judgment-day` or parallel blind-review protocol. Two distinct reviewers (e.g., Judge A and Judge B) must independently verify. Only proceed if both pass or contradictions are triaged.

**Low-Risk Route (Single Agent)**:
*Scenario*: Updating a CSS padding or fixing a README typo.
*Action*: Use standard execution or single `sdd-verify` pass. Redundant evaluation is blocked to save tokens and time.

### 1.3 Reusability of Blind-Review Rules

Blind review (agents not seeing each other's outputs) and adversarial critique (agents instructed to aggressively find flaws) are **not exclusive to Judgment Day**.
Any skill or orchestrator task requiring high-confidence verification MUST:
- Launch sub-agents asynchronously and in parallel.
- Prevent cross-contamination of context between parallel agents.
- Require orchestrator to synthesize independent outputs.

## Phase 2: Four-Pillar Evaluation

### 2.1 The Four Evaluation Fields

When evaluating significant changes or executing `sdd-verify` on complex implementations, evaluation MUST cover four dimensions:

1. **Technical**: Code correctness, performance, edge-case handling, architectural integrity.
2. **Human**: Usability, developer experience (DX), readability, accessibility.
3. **Safety**: Security holes, data loss risks, compliance issues, failure boundaries.
4. **Economic**: Token cost, compute time, redundancy cost, maintenance overhead.

### 2.2 Combining Deterministic Status with Pillar Scoring

Subjective evaluation (Four Pillars) must never override deterministic truth.

**Example**:
- *Deterministic Status*: `FAIL` (Integration tests are failing).
- *Technical*: High (Code looks elegant).
- *Human*: High (Great documentation).
- *Safety*: High (No obvious attack vectors).
- *Economic*: High (Fast execution).
- **Result**: Overall evaluation is **REJECTED**. Deterministic failure overrides high subjective pillar scores.

### 2.3 Deterministic Check Precedence

**Deterministic checks dominate judge-style evaluation for critical correctness claims.**
If machine check (compiler, test runner, syntax linter) fails, judge's opinion is irrelevant. Orchestrator MUST halt progression or mandate fix round, regardless of Four Pillars scores. Subjective evaluation is only meaningful when deterministic foundation is `PASS`.

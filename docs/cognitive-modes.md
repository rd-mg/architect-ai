# Cognitive Modes

Architect-AI uses discrete cognitive postures to shape how agents approach different tasks. By shifting the "lens" of the agent, we achieve higher rigor in verification and better creativity in exploration.

## The 11 Postures

1. **Socratic** (+++Socratic): Reveal assumptions via questions.
2. **Critical** (+++Critical): Evidence-based evaluation of claims.
3. **Systemic** (+++Systemic): Analyze 2nd and 3rd order effects.
4. **Adversarial** (+++Adversarial): Actively try to break the artifact.
5. **Pragmatic** (+++Pragmatic): Ship the smallest correct change.
6. **Forensic** (+++Forensic): Trace evidence chains and provenance.
7. **Economic** (+++Economic): Optimize value/cost under budget.
8. **Empirical** (+++Empirical): Data-driven proof and measurement.
9. **Divergent** (+++Divergent): Generate options without judgment.
10. **Lateral** (+++Lateral): Escape fixed frames via provocation.
11. **Diamond** (+++Diamond): Structured two-phase diverge-then-converge cycle.

## Phase → Posture Mapping

| Phase | Default Posture(s) |
|-------|--------------------|
| sdd-init | (none) |
| sdd-onboard | +++Socratic |
| sdd-explore | +++Socratic |
| sdd-propose | +++Critical |
| sdd-spec | +++Systemic |
| sdd-design | +++Critical + +++Systemic |
| sdd-tasks | +++Pragmatic + +++Economic |
| sdd-apply | +++Pragmatic |
| sdd-verify | +++Adversarial |
| sdd-archive | (none) |

## Orthogonality Check (why exactly 8)

Every posture must be discrete and non-overlapping with the others.
Proposed postures must pass this test:

1. Is there a task where this posture is obviously correct AND no other posture is?
2. Does its core verb overlap with an existing posture's core verb?
3. If removed, would any phase lose a distinct discipline?

### Evolution of the Posture Catalog
Past V2 experiments showed that beyond 6 analytical postures, sub-agents began to drift (confusing Adversarial with Critical). V3 initially raised this to 8 after introducing **Economic** (budgeting) and **Empirical** (measuring). V3.1 introduced three orthogonal **Creative** postures (**Divergent**, **Lateral**, **Diamond**) to support non-SDD tasks like brainstorming and problem-solving, bringing the total catalog to 11.

**Max Postures per prompt: 2**.

--- ARCHIVED: 2026-04-18T05:56:52Z ---


--- ARCHIVED: 2026-04-18T05:57:02Z ---

---
domain: cognitive-mode
change_name: cognitive-postures-expansion
---

# Spec Delta: Cognitive Postures Expansion

## Requirement: Cognitive Repertoire Expansion
The system MUST support exactly 8 cognitive postures, adding Economic and Empirical to the existing 6.

### Behavioral Specification: +++Economic
**Instructional Block (to be injected into prompt)**:
```markdown
### Economic (+++Economic)
**Verb**: Budgeting.
**Context**: Tradeoff analysis under resource constraints (tokens, latency, cost).
**Action**: 
1. Quantify cost/value for all options.
2. Reject options exceeding budget even if technically superior.
3. Recommend the Pareto-optimal choice.
**Distinction**: Pragmatic = ship fast; Economic = ship under budget.
```

### Behavioral Specification: +++Empirical
**Instructional Block (to be injected into prompt)**:
```markdown
### Empirical (+++Empirical)
**Verb**: Measuring.
**Context**: Performance claims or numeric acceptance criteria.
**Action**:
1. No claim without a measurement plan (metric, method, threshold).
2. Mark numbers without plans as PROVISIONAL.
3. Propose the smallest experiment for validation.
**Distinction**: Adversarial = how it breaks; Empirical = how it measures.
```

---

## Requirement: Phase-to-Posture Governance
The orchestration layer MUST enforce the following pairings for ALL sdd-* phases.

| Phase | Mandatory Posture(s) |
|-------|----------------------|
| `sdd-init` | (none) |
| `sdd-onboard` | +++Socratic |
| `sdd-explore` | +++Socratic |
| `sdd-propose` | +++Critical |
| `sdd-spec` | +++Systemic |
| `sdd-design` | +++Critical + +++Systemic |
| `sdd-tasks` | +++Pragmatic + +++Economic |
| `sdd-apply` | +++Pragmatic |
| `sdd-verify` | +++Adversarial |
| `sdd-archive` | (none) |

### Conditional Logic
- **IF** any requirement contains a numeric SLA (e.g., "p99 < 100ms", "coverage > 90%").
- **THEN** `sdd-design` and `sdd-verify` MUST add **+++Empirical** to their active set (subject to the Max 2 invariant).

---

## Requirement: Structural Invariants
- **Max Postures**: Any sub-agent prompt MUST NOT exceed 2 active postures.
- **Verification**: Assets test MUST assert the existence of exactly 8 posture definitions in `cognitive-mode/SKILL.md`.

--- ARCHIVED: 2026-04-18T05:57:09Z ---


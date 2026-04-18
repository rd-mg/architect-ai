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

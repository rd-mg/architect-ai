---
name: cognitive-mode
description: >
  Defines eleven cognitive postures that can be injected as a prompt prefix to 
  shape how an agent approaches a task. Maps each SDD phase to its default posture.
  The orchestrator injects the matching posture block before delegating to a
  sub-agent. This is a REFERENCE skill — the injection logic lives in the
  orchestrator, not here.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "3.0"
---

# Cognitive Mode

## Purpose

MANDATORY: State Cognitive Mode: {n} as first line of response per gate instructions.

Different tasks require different thinking postures. Debugging needs forensic rigor. Design review needs systemic breadth. Exploration needs Socratic questioning. One posture across all tasks produces mediocre results.

This skill defines eleven discrete postures. Orchestrator selects appropriate posture per SDD phase (or explicitly for non-SDD task) and injects it as prefix to sub-agent prompt.

## The Eleven Postures

### 1. Socratic (+++Socratic)

**Use when**: Task starts from ambiguity. Reveal what has NOT been said. Default for `sdd-explore`, `sdd-onboard`.

**Behavior**:
- Formulate **3 clarifying questions** about unstated assumptions before acting
- Identify what user assumes but hasn't stated
- Present questions to orchestrator for resolution
- Do NOT answer own questions — wait

```
+++Socratic
Before producing artifacts, formulate 3 questions about unstated assumptions.
Reveal what has NOT been said. Assumption types: data source, user role, error
handling expectations, performance constraints, backward compatibility,
integration points. Present questions; do not assume answers.
```

---

### 2. Critical (+++Critical)

**Use when**: Rigorous evaluation of claims or feasibility. Default for `sdd-propose`, `sdd-verify`.

**Behavior**:
- For each claim: (1) What evidence supports? (2) What evidence contradicts? (3) What alternatives exist?
- Do NOT accept aesthetic preferences as evidence
- Identify biases (availability, authority, recency, confirmation)

```
+++Critical
Evaluate objectively based on evidence. For each claim:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists? No aesthetic preferences as evidence.
Flag assumptions lacking grounding.
```

---

### 3. Systemic (+++Systemic)

**Use when**: Decision has cross-domain effects or long-term consequences. Default for `sdd-spec`, contributor to `sdd-design`.

**Behavior**:
- Analyze 2nd and 3rd order effects
- Ask: "What breaks elsewhere?" "What new dependencies?" "What becomes harder to change later?"
- Prefer reversible decisions over optimal-but-irreversible ones

**Example prefix**:
```
+++Systemic
Analyze 2nd and 3rd order effects before deciding:
- What OTHER subsystems could break?
- What new dependencies are created?
- What becomes harder to change later?
- Is this decision reversible?
Prefer reversible decisions over optimal-but-irreversible ones.
```

---

### 4. Adversarial (+++Adversarial)

**Use when**: Find what's wrong. Security threat modeling, robust verification. Default for `sdd-verify`. Auto-triggered for D5>=2.

**Behavior**:
- Actively try to BREAK the artifact or compromise security
- Find failure mode or vulnerability author missed
- Assume nothing correct or secure until proven
- Construct counterexamples, edge cases, hostile inputs, privilege escalation scenarios

**Example prefix**:
```
+++Adversarial
Try to break the artifact under review. Identify security risks.
Find failure modes and vulnerabilities author missed. Assume nothing
correct or secure until proven. Construct:
- Counterexamples violating stated invariants
- Edge cases happy path ignores
- Hostile inputs exploiting assumptions or input validation
- Race conditions, environment manipulation, secrets leakage risks
- Upgrade paths corrupting existing data or bypassing access controls
```

---

### 5. Pragmatic (+++Pragmatic)

**Use when**: Mechanical execution. Default for `sdd-tasks`, `sdd-apply`.

**Behavior**:
- Minimum viable solution
- No gold-plating, no over-engineering
- Ship smallest correct change
- "Good enough now" beats "perfect later"
- Avoid scope creep — do exactly what was asked

**Example prefix**:
```
+++Pragmatic
Execute with minimum viable approach. No gold-plating. No over-engineering.
Ship smallest correct change satisfying spec. Do exactly what was asked —
no scope creep, no speculative additions. "Good enough now" beats "perfect later".
```

---

### 6. Forensic (+++Forensic)

**Use when**: Debugging, incident response, context reconstruction. Default for `context-guardian`, explicit debugging.

**Behavior**:
- Trace evidence chains — every claim needs provenance
- Never assume — verify
- Document confidence per finding (`valid`, `stale`, `unverified`)
- Distinguish observed facts from inferred conclusions

**Example prefix**:
```
+++Forensic
Trace evidence chains. For every claim:
- State source (file path, command output, memory ID)
- Mark validation state: [valid] | [stale] | [unverified]
- Distinguish observed facts from inferred conclusions
Never assume — verify. If source cannot be produced, mark claim
as [unverified] and note what evidence would resolve it.
```

---

### 7. Economic (+++Economic)

**Use when**: Tradeoff analysis under resource constraints — token budget, latency SLA, dollar cost, developer-hours. Default for `sdd-tasks`, `sdd-propose` under budget constraints.

**Behavior**:
- Evaluate options through cost lens (time, API tokens, long-term maintenance)
- Reject options technically superior but economically infeasible
- Prefer reuse over reimplementation when ROI favors it
- Surface hidden costs (N+1 queries, per-request API calls, egress costs)

**Example prefix**:
```
+++Economic
Evaluate options through cost lens. For each approach:
(1) Token cost estimate (2) Latency impact (3) Maintenance overhead
(4) Reversibility cost. Reject technically superior options if economic
cost disproportionate. Surface hidden costs (API quotas, N+1, egress).
```

---

### 8. Empirical (+++Empirical)

**Use when**: Measurement-first reasoning — benchmarks, prototypes, data-driven decisions. Default for `researcher`, `sdd-explore` (D2>=1), `sdd-archive`.

**Behavior**:
- No claim without measurement/experimentation plan (metric, method, threshold)
- Mark speculative numbers or claims without evidence as PROVISIONAL
- Propose smallest experiment for validation

**Example prefix**:
```
+++Empirical
For every design or behavior claim, state: (a) metric/evidence source,
(b) how to collect or verify it, (c) acceptance threshold. If verification
has not occurred, mark claim PROVISIONAL. Numbers/speculations without
measurement or proof plan are PROVISIONAL by default.
```

---

### 9. Divergent (+++Divergent)

**Use when**: Brainstorming or creative option generation. Default for first phase of `/brainstorm`.

**Behavior**:
- Generate ≥7 ideas before evaluating any
- Do NOT reject ideas during generation
- Mark each: `[conventional]`, `[stretch]`, `[moonshot]`
- Cluster ideas by theme after generation

**Example prefix**:
```
+++Divergent
Generate ideas without judgment:
- Quantity over quality — 7+ ideas minimum
- Wild ideas welcome — trigger useful associations
- Build on previous ideas (Yes, AND...)
- No evaluation during generation
- Defer judgment — flag concerns for later
- After generation, cluster by theme
- Mark each: [conventional], [stretch], [moonshot]
```

---

### 10. Lateral (+++Lateral)

**Use when**: Stuck or needs non-obvious solution. Default alternative for `/brainstorm`, complex `/solve`.

**Behavior**:
- Apply deliberate provocations to escape fixed frames
- Use REVERSAL, RANDOM ENTRY, CHALLENGE, ANALOGY, or ESCAPE
- Extract ONE actionable insight per technique

**Example prefix**:
```
+++Lateral
Apply deliberate provocations to escape fixed patterns:
1. REVERSAL: What if we did the exact opposite?
2. RANDOM ENTRY: Pick unrelated concept — force connection
3. CHALLENGE: Why must this be this way? Question every assumption
4. ANALOGY: What domain solved structurally similar problem?
5. ESCAPE: What constraint are we treating as fixed that isn't?
Extract ONE actionable insight per technique.
```

---

### 11. Diamond (+++Diamond)

**Use when**: Full creative cycle needed. Default for full `/brainstorm` workflow.

**Behavior**:
- Execute two explicit phases: Diverge, then Converge
- Diverge: generate freely without evaluation
- Converge: evaluate based on feasibility, desirability, viability
- Output: top 3 options ranked with rationale

**Example prefix**:
```
+++Diamond
Two-phase structured ideation:
PHASE 1 (Diverge): Apply +++Divergent. Generate options freely.
PHASE 2 (Converge): Apply +++Critical. Evaluate each against:
  - Feasibility (can we build with current resources?)
  - Desirability (does it solve actual problem?)
  - Viability (can we maintain long-term?)
Output: Top 3 options ranked with rationale.
```

---

## Phase → Posture Mapping v3

| SDD Phase | Default Posture(s) | Alternative (user override or conditional) |
|-----------|--------------------|---------------------------------------------|
| sdd-init | +++Pragmatic | — |
| sdd-onboard | +++Socratic | — |
| sdd-explore | +++Socratic + +++Empirical | Default +++Empirical when D2>=1 |
| sdd-propose | +++Critical + +++Economic | Default +++Economic when budget/quota mentioned |
| sdd-spec | +++Systemic + +++Critical | — |
| sdd-design | +++Systemic + +++Adversarial | ALWAYS adversarial for D1>=2 |
| sdd-tasks | +++Pragmatic + +++Economic | — |
| sdd-apply | +++Pragmatic | +++Forensic +++Pragmatic when CB triggered |
| sdd-verify | +++Critical + +++Adversarial | If D3>=1 |
| sdd-archive | +++Empirical | — |

## Selection Rule: Empirical

Add +++Empirical when task contains numeric acceptance criterion:
latency target, throughput target, memory budget, p99 threshold,
coverage percentage, error rate ceiling.

## Selection Rule: Adversarial

Add +++Adversarial ALWAYS when D5>=2 resolved for current task.

## Hard Ceiling

MAX 2 postures per phase. Three or more produce incoherent prompts.

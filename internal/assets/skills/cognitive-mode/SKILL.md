---
name: cognitive-mode
description: >
  Defines ten cognitive postures that can be injected as a prompt prefix to 
  shape how an agent approaches a task. Maps each SDD phase to its default posture.
  The orchestrator injects the matching posture block before delegating to a
  sub-agent. This is a REFERENCE skill — the injection logic lives in the
  orchestrator, not here.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.0"
---

# Cognitive Mode

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Different tasks require different thinking postures. A debug session needs
forensic rigor. A design review needs systemic breadth. An exploration needs
Socratic questioning. Forcing one posture across all tasks produces mediocre
results on each.

This skill defines eleven discrete postures. The orchestrator selects the
appropriate posture per SDD phase (or explicitly for a non-SDD task) and
injects it as a prefix to the sub-agent's prompt.

## The Eleven Postures

### 1. Socratic (+++Socratic)

**Use when**: Task starts from ambiguity. Goal is to reveal what has NOT been said. Default for `sdd-explore` and `sdd-onboard`.

**Behavior**:
- Before acting, formulate **3 clarifying questions** about unstated assumptions
- Identify what the user is assuming but hasn't explicitly stated
- Present questions to the orchestrator for resolution
- Do NOT answer your own questions — wait for resolution

**Example prefix**:
```
+++Socratic
Before producing artifacts, formulate 3 questions about unstated assumptions
in the request. Reveal what has NOT been said. Examples of assumption types:
data source, user role, error handling expectations, performance constraints,
backward compatibility, integration points. Present the questions; do not
assume answers.
```

---

### 2. Critical (+++Critical)

**Use when**: Task requires rigorous evaluation of claims or feasibility. Default for `sdd-propose` and `sdd-verify`.

**Behavior**:
- For each claim, ask:
  1. What evidence supports this?
  2. What evidence contradicts this?
  3. What alternative explanations exist?
- Do NOT accept aesthetic preferences as evidence
- Identify biases (availability, authority, recency, confirmation)

**Example prefix**:
```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists? Do not accept aesthetic preferences
as evidence. Flag assumptions that lack grounding.
```

---

### 3. Systemic (+++Systemic)

**Use when**: Decision has cross-domain effects or long-term consequences. Default for `sdd-spec` and contributor to `sdd-design`.

**Behavior**:
- Analyze 2nd and 3rd order effects
- Ask: "What breaks elsewhere if I do this?"
- Ask: "What new dependencies does this create?"
- Ask: "What becomes harder to change later?"
- Prefer reversible decisions over optimal-but-irreversible ones

**Example prefix**:
```
+++Systemic
Analyze 2nd and 3rd order effects before deciding. For the proposed approach:
- What OTHER subsystems could break?
- What new dependencies are created?
- What becomes harder to change later?
- Is this decision reversible?
Prefer reversible decisions over optimal-but-irreversible ones.
```

---

### 4. Adversarial (+++Adversarial)

**Use when**: Goal is to find what's wrong. Default for `sdd-verify`. Also used by `adaptive-reasoning` in adversarial-review mode.

**Behavior**:
- Actively try to BREAK the artifact
- Find the failure mode the author missed
- Assume nothing is correct until proven
- Construct counterexamples, edge cases, hostile inputs

**Example prefix**:
```
+++Adversarial
Try to break the artifact under review. Find the failure modes the author
missed. Assume nothing is correct until proven. Construct:
- Counterexamples that violate stated invariants
- Edge cases the happy path ignores
- Hostile inputs that exploit assumptions
- Race conditions in concurrent execution
- Upgrade paths that corrupt existing data
```

---

### 5. Pragmatic (+++Pragmatic)

**Use when**: Task is mechanical execution. Default for `sdd-tasks` and `sdd-apply`.

**Behavior**:
- Minimum viable solution
- No gold-plating, no over-engineering
- Ship the smallest correct change
- "Good enough now" beats "perfect later"
- Avoid scope creep — do exactly what was asked

**Example prefix**:
```
+++Pragmatic
Execute the task with the minimum viable approach. No gold-plating. No
over-engineering. Ship the smallest correct change that satisfies the spec.
Do exactly what was asked — no scope creep, no speculative additions.
"Good enough now" beats "perfect later".
```

---

### 6. Forensic (+++Forensic)

**Use when**: Debugging, incident response, or context reconstruction. Default for `context-guardian` and explicit debugging work.

**Behavior**:
- Trace evidence chains — every claim needs provenance
- Never assume — verify
- Document confidence per finding (`valid`, `stale`, `unverified`)
- Distinguish observed facts from inferred conclusions

**Example prefix**:
```
+++Forensic
Trace evidence chains. For every claim:
- State the source (file path, command output, memory ID)
- Mark validation state: [valid] | [stale] | [unverified]
- Distinguish observed facts from inferred conclusions
Never assume — verify. If a source cannot be produced, mark the claim
as [unverified] and note what evidence would resolve it.
```

---

---

### 7. Economic (+++Economic)

**Use when**: the task requires tradeoff analysis under resource
constraints — token budget, latency SLA, dollar cost, developer-hours.

**Behavior**:
- Quantify cost/value for all options
- Reject options whose cost exceeds the declared budget
- Recommend the Pareto-optimal choice

**Example prefix**:
```
+++Economic
Budget constraints: {tokens|time|cost — state the limit}. Enumerate 2–3
options. For each: estimate cost in the constrained dimension and value
delivered. Recommend the Pareto-optimal choice. Reject options that
exceed budget, stating which constraint they violate.
```

---

### 8. Empirical (+++Empirical)

**Use when**: the task requires measurement-first reasoning —
benchmarks, A/B prototypes, data-driven design decisions,
performance regression verification.

**Behavior**:
- No claim without a measurement plan (metric, method, threshold)
- Mark numbers without plans as PROVISIONAL
- Propose the smallest experiment for validation

**Example prefix**:
```
+++Empirical
For every design claim, state: (a) the metric, (b) how to collect it,
(c) the acceptance threshold. If measurement hasn't happened, mark
the claim PROVISIONAL and propose the smallest experiment to validate
it. Numbers without a measurement plan are PROVISIONAL by default.
```

---

### 9. Divergent (+++Divergent)

**Use when**: Task requires brainstorming or creative option generation. Default for the first phase of `/brainstorm`.

**Behavior**:
- Generate ≥7 ideas before evaluating any
- Do NOT reject ideas during generation phase
- Mark each idea as `[conventional]`, `[stretch]`, or `[moonshot]`
- Cluster ideas by theme after generation

**Example prefix**:
```
+++Divergent
Generate ideas without judgment. Rules:
- Quantity over quality — aim for 7+ ideas minimum
- Wild ideas welcome — they trigger useful associations
- Build on previous ideas (Yes, AND...)
- No evaluation during generation phase
- Defer judgment — flag concerns for later evaluation
- After generation, cluster ideas by theme
- Mark each: [conventional], [stretch], [moonshot]
```

---

### 10. Lateral (+++Lateral)

**Use when**: The task is stuck or needs a non-obvious solution. Default alternative for `/brainstorm` or complex `/solve`.

**Behavior**:
- Apply deliberate provocations to escape fixed frames
- Use REVERSAL, RANDOM ENTRY, CHALLENGE, ANALOGY, or ESCAPE
- Extract ONE actionable insight per technique applied

**Example prefix**:
```
+++Lateral
Apply deliberate provocations to escape fixed patterns:
1. REVERSAL: What if we did the exact opposite?
2. RANDOM ENTRY: Pick an unrelated concept — force a connection
3. CHALLENGE: Why does this have to be this way? Question every assumption
4. ANALOGY: What domain solved a structurally similar problem?
5. ESCAPE: What constraint are we treating as fixed that isn't?
After provocation, extract ONE actionable insight per technique.
```

---

### 11. Diamond (+++Diamond)

**Use when**: A full creative cycle is needed. Default for the full `/brainstorm` workflow.

**Behavior**:
- Execute two explicit phases: Diverge, then Converge
- Diverge phase: generate freely without evaluation
- Converge phase: evaluate based on feasibility, desirability, viability
- Output must be the top 3 options ranked with rationale

**Example prefix**:
```
+++Diamond
Two-phase structured ideation:
PHASE 1 (Diverge): Apply +++Divergent rules. Generate options freely.
PHASE 2 (Converge): Apply +++Critical rules. Evaluate each option against:
  - Feasibility (can we build it with current resources?)
  - Desirability (does it solve the actual problem?)
  - Viability (can we maintain it long-term?)
Output: Top 3 options ranked with rationale.
```

---

## Phase → Posture Mapping

| SDD Phase | Default Posture(s) | Alternative (user override or conditional) |
|-----------|--------------------|---------------------------------------------|
| sdd-init | (none) | — |
| sdd-onboard | +++Socratic | — |
| sdd-explore | +++Socratic | — |
| sdd-propose | +++Critical | — |
| sdd-spec | +++Systemic | — |
| sdd-design | +++Critical + +++Systemic | +++Critical + +++Empirical (numeric SLAs) |
| sdd-tasks | +++Pragmatic + +++Economic | — |
| sdd-apply | +++Pragmatic | — |
| sdd-verify | +++Adversarial | +++Adversarial + +++Empirical (numeric SLAs) |
| sdd-archive | (none) | — |

## Non-SDD Task → Posture Mapping

| Task Type | Required Postures |
|-----------|------------------|
| /brainstorm | +++Divergent, +++Diamond |
| /solve | +++Forensic, +++Systemic |
| /investigate | +++Socratic, +++Empirical |
| /debug | +++Forensic, +++Adversarial |
| /prototype | +++Pragmatic |

## Selection Rule for Empirical

Add +++Empirical when the task contains any numeric acceptance
criterion: latency target, throughput target, memory budget,
p99 threshold, coverage percentage, error rate ceiling.

## Maximum Postures Per Phase

Hard ceiling: **2**. Three or more produce incoherent prompts.

---
name: cognitive-mode
description: >
  Defines six cognitive postures (Socratic, Critical, Systemic, Adversarial,
  Pragmatic, Forensic) that can be injected as a prompt prefix to shape how
  an agent approaches a task. Maps each SDD phase to its default posture.
  The orchestrator injects the matching posture block before delegating to a
  sub-agent. This is a REFERENCE skill — the injection logic lives in the
  orchestrator, not here.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "1.0"
---

# Cognitive Mode

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.


Different tasks require different thinking postures. A debug session needs
forensic rigor. A design review needs systemic breadth. An exploration needs
Socratic questioning. Forcing one posture across all tasks produces mediocre
results on each.

This skill defines six discrete postures. The orchestrator selects the
appropriate posture per SDD phase (or explicitly for a non-SDD task) and
injects it as a prefix to the sub-agent's prompt.

## The Six Postures

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

## Phase → Posture Mapping

The orchestrator uses this table when delegating an SDD phase:

| SDD Phase | Primary Posture | Why |
|-----------|----------------|-----|
| sdd-explore | +++Socratic | Reveal assumptions, explore the problem space |
| sdd-propose | +++Critical | Evaluate feasibility with rigor |
| sdd-spec | +++Systemic | Detect cross-domain dependencies |
| sdd-design | +++Critical + +++Systemic | Architecture needs both rigor and system view |
| sdd-tasks | +++Pragmatic | Mechanical breakdown, no over-engineering |
| sdd-apply | +++Pragmatic | Execute the spec, don't freelance |
| sdd-verify | +++Adversarial | Find defects, assume nothing works |
| sdd-archive | (none) | Mechanical copy and close |
| sdd-init | (none) | Tool detection, mechanical |
| sdd-onboard | +++Socratic | First-time user needs question-driven flow |
| context-guardian | +++Forensic | Trace provenance, validate facts |

Non-SDD tasks (explicit postures):

| Task Type | Posture |
|-----------|---------|
| PR review / code audit | +++Adversarial + +++Forensic |
| Debugging / incident response | +++Forensic |
| Architecture decision with competing options | +++Critical + +++Systemic |
| Quick feature implementation | +++Pragmatic |
| First interaction with a new codebase | +++Socratic |

---

## Multi-Posture Injection

If a phase maps to multiple postures (e.g., sdd-design = Critical + Systemic),
inject BOTH blocks. Keep them distinct — do not merge them into one paragraph.

Example for sdd-design:

```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists?

+++Systemic
Analyze 2nd and 3rd order effects. What breaks elsewhere? What new
dependencies are created? What becomes harder to change later?
```

---

## Orchestrator Integration

The orchestrator performs these steps before each sub-agent launch:

1. Identify the phase being delegated (e.g., `sdd-propose`)
2. Look up the posture(s) in the Phase → Posture Mapping table above
3. Inject the matching prefix block(s) at the top of the sub-agent's prompt
4. Follow with the standard `## Project Standards (auto-resolved)` block
5. Follow with the phase-specific task instructions

See `internal/assets/{agent}/sdd-orchestrator.md` for the canonical injection template.

---

## Sub-Agent Integration

When a sub-agent receives a prompt with a `+++Posture` block at the top:

1. Read and internalize the posture before reading task instructions
2. Apply the behavior throughout the work
3. Reflect the posture in the return envelope — e.g., Socratic mode returns questions, not answers; Adversarial mode returns findings, not proposals

---

## User Override

A user can explicitly request a posture for any task:

```
User: "Review this PR with forensic rigor."
→ Orchestrator injects +++Forensic (plus +++Adversarial by phase default)

User: "Brainstorm options for X."
→ Orchestrator injects +++Socratic (divergent thinking)

User: "Just implement it — no over-engineering."
→ Orchestrator injects +++Pragmatic
```

---

## When NOT to Use Postures

- `sdd-archive` and `sdd-init` are mechanical phases; no posture is beneficial
- Trivial confirmations ("yes, ok, proceed") do not need posture injection
- Machine-to-machine handoffs between sub-agents reuse the original posture; they do not re-inject

---

## Anti-Patterns

- Mixing three or more postures in one prompt — dilutes the effect
- Overriding a phase posture casually — the defaults are calibrated
- Using +++Adversarial for synthesis tasks (use +++Critical instead)
- Using +++Socratic for well-defined execution tasks (use +++Pragmatic)
- Skipping posture injection because "the sub-agent should know what to do" — they don't; posture shapes the behavior

## Resources

- `internal/assets/skills/adaptive-reasoning/SKILL.md` — references +++Adversarial in Mode 2
- `internal/assets/skills/context-guardian/SKILL.md` — uses +++Forensic
- `internal/assets/{agent}/sdd-orchestrator.md` — contains the injection logic

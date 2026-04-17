# Cognitive Modes (6 Postures)

**Scope**: `internal/assets/skills/cognitive-mode/SKILL.md` defines the mechanism. This doc is the human-readable reference.

The orchestrator injects ONE (or more) posture block at the top of every sub-agent prompt, before `## Project Standards` and before the phase protocol. The posture is NOT a personality — it's a thinking discipline that shapes how the sub-agent approaches the task.

---

## Why six, not one

A single "be smart" prompt produces mediocre results across the board. Different SDD phases reward different disciplines: exploration rewards question-asking, verification rewards adversarial skepticism. Forcing the same posture on both gives you unfounded specs and missed bugs.

The six postures are **discrete and non-overlapping**. They're mapped 1:1 (sometimes 1:2) to SDD phases.

---

## The six postures

| # | Posture | Core verb | Default SDD phase |
|---|---------|-----------|-------------------|
| 1 | Socratic | *ask* | `sdd-explore`, `sdd-onboard` |
| 2 | Critical | *evaluate* | `sdd-propose`, `sdd-verify` |
| 3 | Systemic | *connect* | `sdd-spec`, `sdd-design` |
| 4 | Adversarial | *attack* | `sdd-verify` |
| 5 | Pragmatic | *ship* | `sdd-tasks`, `sdd-apply` |
| 6 | Forensic | *trace* | `context-guardian`, debugging |

### 1. Socratic (+++Socratic)

**Use when**: the task starts from ambiguity. Goal is to reveal what has NOT been said.

**Behavior**: formulate 3 questions about unstated assumptions; do NOT answer them; surface them to the orchestrator.

**Typical prefix**:
```
+++Socratic
Before producing artifacts, formulate 3 questions about unstated assumptions
in the request. Reveal what has NOT been said. Examples: data source, user
role, error handling expectations, performance constraints, backward
compatibility. Present the questions; do not assume answers.
```

### 2. Critical (+++Critical)

**Use when**: the task requires rigorous evaluation of claims, feasibility, or tradeoffs.

**Behavior**: for each claim — what evidence supports it, what contradicts it, what's a stronger version; reject unfounded statements.

### 3. Systemic (+++Systemic)

**Use when**: the task crosses modules, bounded contexts, or long-lived interfaces.

**Behavior**: draw the boundary lines; identify upstream and downstream effects; call out coupling; prefer explicit contracts over implicit ones.

### 4. Adversarial (+++Adversarial)

**Use when**: the task is validation — spec compliance, security, edge cases.

**Behavior**: try to break it. Enumerate failure modes. Assume the implementer was optimistic. Write attacks, not confirmations.

### 5. Pragmatic (+++Pragmatic)

**Use when**: the task is execution — break work into steps, ship code.

**Behavior**: smallest working change; follow existing patterns unless they're wrong; resist re-architecture mid-stream.

### 6. Forensic (+++Forensic)

**Use when**: you're assembling evidence chains, reconstructing state, or debugging.

**Behavior**: every claim needs provenance; mark validation state per fact; trace cause → effect explicitly.

---

## Combining postures

Some phases use two postures. Example — `sdd-design`:

```
+++Critical
[...critical block...]

+++Systemic
[...systemic block...]
```

The sub-agent is asked to evaluate AND connect. Don't combine more than two — three is an incoherent prompt.

---

## Phase → posture mapping

Canonical table (sourced from `skills/cognitive-mode/SKILL.md`):

| Phase | Posture(s) |
|-------|------------|
| sdd-init | (none) |
| sdd-onboard | +++Socratic |
| sdd-explore | +++Socratic |
| sdd-propose | +++Critical |
| sdd-spec | +++Systemic |
| sdd-design | +++Critical + +++Systemic |
| sdd-tasks | +++Pragmatic |
| sdd-apply | +++Pragmatic |
| sdd-verify | +++Adversarial |
| sdd-archive | (none) |

The orchestrator reads this table and injects the posture automatically — you do not pass it manually.

---

## Overriding per-task

If a task needs a posture other than the phase default, the user can prefix the command:

```
/sdd-apply --posture=Forensic
```

The orchestrator substitutes the posture block for that single delegation. This is rare and usually signals the wrong phase is being used.

---

## Why NOT five, seven, or nine

- **5**: Forensic is too different from Critical to merge; they operate on different time axes (Forensic = past evidence; Critical = present claims).
- **7+**: Every additional posture dilutes the signal. Experiments on V2 showed sub-agents confusing Adversarial with Critical when both were present; we dropped "Skeptical" for that reason.

---

## See also

- `internal/assets/skills/cognitive-mode/SKILL.md` — machine-readable definition
- `caveman-integration.md` — how caveman dual-mode interacts with posture injection
- `adaptive-reasoning-v1.md` — how the orchestrator chooses reasoning depth independently of posture

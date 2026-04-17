# Cognitive Modes — Architecture Guide

**Status**: Stable | **Introduced in**: V3 | **Related skill**: `internal/assets/skills/cognitive-mode/SKILL.md`

---

## Why Cognitive Modes Exist

Different tasks need different thinking postures. A debug session requires forensic rigor. A design review needs systemic breadth. An exploration phase needs Socratic questioning. Before V3, architect-ai had a single implicit "thinking mode" — analytical, evaluative — that was applied to every task. The results were mediocre for edge cases.

V3 introduces **six explicit cognitive postures**, each mapped to the SDD phase(s) where it produces the best outcomes. The orchestrator injects the appropriate posture block at the top of each sub-agent's prompt.

## The Six Postures

### 1. +++Socratic (Question-Driven)

**Principle**: Before producing artifacts, reveal what has NOT been said.

**Behavior**:
- Formulates 3 clarifying questions about unstated assumptions
- Does NOT answer its own questions
- Surfaces data sources, user roles, error expectations, performance constraints

**Default phases**: `sdd-explore`, `sdd-onboard`

**Example**: When exploring "add dark mode" without clarification, the Socratic posture asks:
- Does "dark mode" mean just colors, or also contrast/motion preferences?
- Is the toggle per-user or global?
- Should the preference survive a logout?

---

### 2. +++Critical (Evidence-Driven)

**Principle**: Evaluate claims against evidence. Don't accept aesthetic preferences.

**Behavior**:
- For each claim: What evidence supports it? What contradicts it? What's the alternative?
- Flags unproven assumptions
- Identifies biases (availability, authority, recency, confirmation)

**Default phases**: `sdd-propose`, `sdd-verify`

**Example**: When proposing "rewrite the auth service", the Critical posture asks:
- Evidence: what specific failures of the current auth exist?
- Counter-evidence: have we measured vs guessed?
- Alternative: would a targeted fix cover 80% of the benefit?

---

### 3. +++Systemic (Second-Order Effects)

**Principle**: What breaks elsewhere? What becomes harder to change later?

**Behavior**:
- Analyzes 2nd and 3rd order consequences
- Maps dependency ripples
- Prefers reversible decisions over optimal-but-irreversible

**Default phases**: `sdd-spec`, `sdd-design` (combined with Critical)

**Example**: When designing "move sessions to Redis", Systemic asks:
- What OTHER subsystems assume local session state?
- What new operational dependencies are created?
- Is this reversible? (Can we roll back without data loss?)

---

### 4. +++Adversarial (Break-It-On-Purpose)

**Principle**: Nothing is correct until proven. Find what the author missed.

**Behavior**:
- Constructs counterexamples that violate stated invariants
- Surfaces edge cases the happy path ignores
- Identifies hostile inputs, race conditions, corrupting upgrade paths

**Default phases**: `sdd-verify`

**Example**: When verifying "payment processing feature", Adversarial tries:
- Double-submission race: two clicks on the submit button within 100ms
- Partial failure: payment succeeds but order creation fails
- Hostile input: negative amount, zero amount, overflow amount
- Upgrade path: what if an existing payment is in a state this new code doesn't recognize?

---

### 5. +++Pragmatic (Minimum Viable)

**Principle**: Ship the smallest correct change. No gold-plating.

**Behavior**:
- Does exactly what was asked, no scope creep
- Prefers "good enough now" over "perfect later"
- Resists speculative additions

**Default phases**: `sdd-tasks`, `sdd-apply`

**Example**: When implementing "add a search filter", Pragmatic:
- Implements only the filter that was specced
- Does NOT refactor the search service to be "more flexible"
- Does NOT add a framework for "future filters"

---

### 6. +++Forensic (Evidence Chains)

**Principle**: Every claim needs provenance. Never assume — verify.

**Behavior**:
- States the source for every fact (file, command, memory ID)
- Marks validation state: `[valid]`, `[stale]`, `[unverified]`
- Distinguishes observed facts from inferred conclusions

**Default phases**: `context-guardian`, explicit debugging

**Example**: When debugging "login is broken for some users", Forensic:
- `[provenance: logs/2026-04-17 14:30] [valid]` 47 login failures in the last hour
- `[provenance: logs/2026-04-17 14:30] [valid]` all failures have `x-forwarded-for` set
- `[provenance: inferred]` reverse proxy may be stripping auth headers
- Next: verify by inspecting proxy config (not yet confirmed)

---

## Phase → Posture Mapping

| SDD Phase | Primary Posture(s) | Why |
|-----------|-------------------|-----|
| sdd-explore | +++Socratic | Reveal assumptions before acting |
| sdd-propose | +++Critical | Evaluate feasibility with rigor |
| sdd-spec | +++Systemic | Detect cross-domain dependencies |
| sdd-design | +++Critical + +++Systemic | Architecture needs rigor AND system view |
| sdd-tasks | +++Pragmatic | Mechanical breakdown, no over-engineering |
| sdd-apply | +++Pragmatic | Execute the spec, don't freelance |
| sdd-verify | +++Adversarial | Assume nothing works, find defects |
| sdd-archive | (none) | Mechanical close-out |
| sdd-init | (none) | Detection and configuration |
| sdd-onboard | +++Socratic | New user needs question-driven flow |
| context-guardian | +++Forensic | Trace provenance, validate facts |

## Multi-Posture Injection

When a phase maps to multiple postures (e.g., sdd-design = Critical + Systemic), the orchestrator injects BOTH blocks at the top of the sub-agent prompt:

```
+++Critical
Evaluate objectively based on evidence. For each claim made or implied:
(1) What evidence supports it? (2) What evidence contradicts it?
(3) What alternative explanation exists?

+++Systemic
Analyze 2nd and 3rd order effects. What breaks elsewhere? What new
dependencies are created? What becomes harder to change later?

## Project Standards (auto-resolved)
[compact rules]

## Task
[phase-specific instructions]
```

The sub-agent applies both simultaneously. It does not choose one.

## User Override

A user can explicitly request a posture for any task:

```
User: "Review this PR with forensic rigor."
→ Orchestrator injects +++Forensic (plus +++Adversarial by phase default)

User: "Brainstorm options for X."
→ Orchestrator injects +++Socratic

User: "Just implement it, no over-engineering."
→ Orchestrator injects +++Pragmatic (phase default holds)
```

## When NOT to Use Postures

- `sdd-archive` and `sdd-init` are mechanical phases; injection adds noise
- Trivial confirmations ("yes, proceed") need no posture
- Machine-to-machine handoffs between sub-agents reuse the caller's posture

## Implementation Detail

The injection happens in the orchestrator prompt before delegation:

1. Orchestrator identifies the phase (e.g., `sdd-propose`)
2. Looks up posture(s) from the Phase → Posture table
3. Reads the corresponding prefix block from `internal/assets/skills/cognitive-mode/SKILL.md`
4. Injects at the TOP of the sub-agent's prompt, BEFORE `## Project Standards`
5. Sub-agent's `sdd-phase-common.md` section A2 reads the posture and applies it

See:
- Orchestrator: `internal/assets/claude/sdd-orchestrator.md` section "Cognitive Posture Injection"
- Sub-agent receiver: `internal/assets/skills/_shared/sdd-phase-common.md` section A2
- Posture reference: `internal/assets/skills/cognitive-mode/SKILL.md`

## Anti-Patterns

- Mixing three or more postures in one prompt dilutes the effect
- Overriding a phase's default posture casually — the defaults are calibrated
- Using +++Adversarial for synthesis tasks (use +++Critical instead)
- Using +++Socratic for well-defined execution tasks (use +++Pragmatic)
- Skipping posture injection because "the sub-agent should know what to do" — it doesn't; posture shapes the behavior

## Related

- `docs/caveman-integration.md` — how output style interacts with postures
- `docs/adaptive-reasoning-v1.md` — how reasoning modes are routed
- `internal/assets/skills/cognitive-mode/SKILL.md` — the authoritative reference

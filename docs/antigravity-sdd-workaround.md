# Antigravity SDD Workaround — Single-Threaded Simulation

**Scope**: Antigravity's runtime is single-threaded with respect to sub-agent delegation. True parallel sub-agents are simulated. This affects how SDD phases compose and how the user should drive them.

---

## What's different about Antigravity

Claude Code, Cursor, Gemini CLI, and most other agents support **real parallel sub-agents**: the orchestrator launches N sub-agents, they run concurrently, results come back out of order but independently.

Antigravity does not. Its sub-agent primitive is **simulated**: the orchestrator pushes a "sub-agent" context frame onto its own stack, runs it in-line, pops the frame, then continues. From the model's perspective it *looks* like a sub-agent call, but under the hood it's one thread.

This has three consequences for SDD:

1. **Latency is additive**. Two "parallel" sub-agents take roughly 2× the time of one, not 1× as on other agents.
2. **Context contamination is possible**. If the orchestrator forgets to strip the sub-agent frame after the pop, leftover instructions from the sub-agent can leak into the next phase.
3. **Long cycles exhaust the context window**. Running `sdd-explore → sdd-propose → sdd-spec → sdd-design → sdd-tasks → sdd-apply → sdd-verify → sdd-archive` in one Antigravity session can hit 180K tokens, leaving no headroom for the orchestrator's own reasoning.

---

## The workaround — fresh-session per phase boundary

**Do NOT try to run the full SDD cycle in one Antigravity session.**

Instead, the orchestrator prints this reminder after each phase:

```
Phase sdd-{phase} complete. For Antigravity:
  Option A: continue here IF context usage < 50%
  Option B: start a fresh Antigravity session and run /sdd-continue {change-name}
```

Option B re-loads the state from Engram (the `sdd/{change-name}/state` topic-key), so nothing is lost. The fresh session starts clean.

This notice is NOT present in the Claude, Gemini, Cursor, etc. orchestrators — they can handle the full cycle in one session.

---

## What the user should do

### Short cycles (≤4 phases): stay in one session

If you're running `sdd-explore → sdd-propose → sdd-spec → sdd-tasks` and the context usage stays under 50%, finish in one session. The workaround is for when the cycle is long enough to hit window pressure.

### Long cycles (full SDD): session per phase boundary

Recommended boundaries:
- Session 1: `/sdd-init`, `/sdd-explore`
- Session 2: `/sdd-propose`, `/sdd-spec`, `/sdd-design`
- Session 3: `/sdd-tasks`, `/sdd-apply`
- Session 4: `/sdd-verify`, `/sdd-archive`

Each new session starts with:
```
/sdd-continue <change-name>
```

The orchestrator will:
1. Read `sdd/{change-name}/state` from Engram
2. Determine the next dependency-ready phase
3. Run it

### Strict TDD cycles — also session per batch

If you're in strict TDD with multiple apply batches, use a fresh session per batch. Each batch persists `sdd/{change-name}/apply-progress`, and `/sdd-continue` will merge new progress with old.

---

## When the workaround doesn't help

**Two real problems remain**:

1. **Real-time iteration feels slow**. Fresh sessions have a ~5-10s startup cost. For quick experimentation, Antigravity is the wrong tool — use Claude Code or Cursor.

2. **Simulated parallelism is still simulated**. If a phase protocol calls for two sub-agents in parallel (e.g., `sdd-verify` doing deterministic checks AND adversarial checks simultaneously), you're paying the full sequential cost on Antigravity. The orchestrator does NOT try to be clever about this — it just runs them in sequence.

---

## Detection

The orchestrator auto-detects Antigravity via its environment. The detection happens once per session at the top of the orchestrator. If detection is wrong, the user can force it:

```
I'm on Antigravity. Please apply the single-threaded workaround.
```

or

```
I'm not on Antigravity (detection wrong). Please run full cycle without the per-phase reminder.
```

---

## Engineering notes (for contributors)

- The orchestrator core is identical across agents EXCEPT for `internal/assets/antigravity/sdd-orchestrator.md`, which has a `## Single-Threaded Simulation` notice at the top.
- The notice is a **block**, not a flag in the orchestrator's own reasoning. It's visible to the user by design — they need to know the workaround exists.
- When Antigravity gains real parallel sub-agents, remove the notice block from `antigravity/sdd-orchestrator.md` and delete this doc.

---

## See also

- `internal/assets/antigravity/sdd-orchestrator.md` — the notice block
- `internal/assets/claude/sdd-orchestrator.md` — canonical orchestrator without the notice
- `plans/master-plan.md` section 3.2 — installation notes

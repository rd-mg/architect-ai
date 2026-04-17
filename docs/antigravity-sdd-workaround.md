# Antigravity — SDD Integration Workaround

**Status**: Known Limitation | **Applies to**: Antigravity IDE

---

## The Problem

Antigravity IDE is the newest agent surface supported by architect-ai. It provides:
- Mission Control (a planning panel for tasks)
- Agent-based code execution

However, Antigravity **does not yet support nested sub-agent delegation natively**. When architect-ai's SDD orchestrator launches a sub-agent for a phase (e.g., `sdd-propose`), Antigravity does not spawn a parallel context. Instead, the orchestrator itself switches context and executes the sub-agent's role within the same thread.

This is called **single-threaded simulation** — the orchestrator plays both the coordinator and the executor roles sequentially.

## Why This Matters

The SDD architecture assumes parallel context isolation between orchestrator and sub-agents:
- The orchestrator keeps a thin working context (just summaries and decisions)
- Sub-agents operate in fresh contexts per invocation (no accumulated noise)
- Results flow back via structured return envelopes

With single-threaded simulation, the orchestrator accumulates ALL sub-agent context inline. Over a 6-phase SDD cycle (explore → propose → spec → design → tasks → apply → verify → archive), the orchestrator's context grows by the sum of all phase artifacts.

For a moderate change (~4K tokens per phase), this means the orchestrator context can exceed 30K tokens by the end of the cycle. Antigravity's context window can handle this, but responsiveness degrades.

## The Workaround

In the Antigravity orchestrator prompt, we include an explicit limitation notice and guidance:

```markdown
## Antigravity Limitation: Single-Threaded Simulation

This orchestrator runs in Antigravity, which does not yet support native
nested sub-agent delegation. As a result:

- All SDD phases execute in this same context (no parallel sub-agents)
- Context accumulates across phases (no reset between delegations)
- For long SDD cycles, invoke `/sdd-init` then `/sdd-explore` then
  manually start fresh sessions for subsequent phases to preserve context

When this orchestrator "delegates" a phase, it switches posture:
1. Read the relevant phase protocol from `sdd-phase-protocols/`
2. Apply the posture and execute the phase as an inline role switch
3. Return the phase artifact per the phase's return envelope contract
4. Resume orchestrator mode for the next decision

If the context grows beyond comfort (e.g., > 50K tokens), invoke
context-guardian to compact and continue.
```

## User Experience

For Antigravity users, the workflow is:

1. **Short changes (1-3 phases)**: Run the full SDD cycle in one session
2. **Medium changes (4-6 phases)**: Run phases 1-3, then start a fresh Antigravity session for phases 4-6 (Engram memory preserves state across sessions)
3. **Long changes (full cycle + iteration)**: Fresh session per phase, relying on Engram persistence for context

The `sdd-orchestrator.md` for Antigravity includes a reminder after each phase:

```
Phase {phase} complete. For the next phase ({next}):
- Option A: continue in this session (simpler, may hit context limits)
- Option B: start a fresh Antigravity session and run /sdd-continue {change-name}
  (preferred for moderate+ changes)
```

## Context Guardian Integration

The Antigravity orchestrator auto-triggers `context-guardian` at 50% context usage (the standard threshold). This compresses the accumulated history into a Context Pack persisted to Engram.

When a fresh session resumes via `/sdd-continue`, the orchestrator reads the Context Pack first, establishing continuity without replaying the raw history.

## When Antigravity Adds Native Delegation

If/when Antigravity supports native nested sub-agents:

1. Remove the `Single-Threaded Simulation` notice from `internal/assets/antigravity/sdd-orchestrator.md`
2. Enable the standard delegation syntax (same as Claude, Cursor, Gemini)
3. No changes to phase protocols are needed — they're agent-agnostic
4. Update this doc to archive it as historical

Track Antigravity release notes for parallel context support.

## Comparison to Other Agents

| Agent | Native Delegation | Strategy |
|-------|------------------|----------|
| Claude Code | ✅ Yes (Task tool) | Standard multi-agent |
| Cursor | ✅ Yes (native subagents) | Standard multi-agent |
| Gemini CLI | ✅ Yes (native agents) | Standard multi-agent |
| OpenCode | ✅ Yes (multi-mode overlay) | Overlay JSON with mode profiles |
| Windsurf | ✅ Yes (Plan Mode / Code Mode) | Standard multi-agent |
| Kiro | ✅ Yes (native subagents + steering) | Standard multi-agent |
| Codex | ⚠️ Solo agent | Single-threaded simulation (similar to Antigravity) |
| **Antigravity** | ⚠️ Solo agent (for now) | **Single-threaded simulation + fresh session workaround** |

## User FAQ

**Q: Why does my context fill up so fast?**
A: Because Antigravity runs SDD phases in the same context. Use fresh sessions for each phase on longer changes.

**Q: Does the workaround lose state?**
A: No. Engram persists every artifact. `/sdd-continue` reads the state back.

**Q: Can I run the full SDD flow in one Antigravity session?**
A: For short changes, yes. For moderate+ changes, expect to restart sessions between phases.

**Q: Will this be fixed?**
A: When Antigravity ships native delegation, this workaround becomes unnecessary.

## References

- `internal/assets/antigravity/sdd-orchestrator.md` — orchestrator with limitation notice
- `internal/assets/skills/context-guardian/SKILL.md` — auto-compaction hook
- `docs/cognitive-modes.md` — posture management (applies to Antigravity same as other agents)

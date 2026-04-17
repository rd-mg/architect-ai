# Caveman Dual-Mode — Integration Guide

**Status**: Stable | **Introduced in**: V3 | **Inspiration**: [github.com/JuliusBrussee/caveman](https://github.com/JuliusBrussee/caveman)

---

## The Problem

Default agent output is verbose. Agents produce filler text ("Let me think about this...", "I'll now...", "This should work because..."), preamble that restates the question, and conversational softeners that inflate token counts without adding information. Over a long session, this waste adds up — both in cost and in context window pressure.

Simultaneously, truly compressing EVERY output would break things: code needs proper grammar, security warnings need clarity, irreversible action confirmations need explicit English.

## The Solution: Dual-Mode

V3 adopts a **two-level compression strategy**:

- **ULTRA mode** for internal artifacts (telegraphic, drops articles and filler)
- **LITE mode** for user-facing responses (no filler, grammar intact, concise)
- **Normal English** for code, commits, PRs, security warnings, irreversible actions

The mode is determined by the OUTPUT TARGET, not by user preference. Internal artifacts going to Engram or other sub-agents use ULTRA. Messages shown to the user use LITE. Code and security content use normal English.

## The Three Modes

### ULTRA (Internal Artifacts)

Used for: Engram content, context packs, artifacts between sub-agents, thought traces.

Rules:
- Drop articles (a, the)
- Drop filler (so, then, basically, actually)
- Use fragments where grammar permits
- Pattern: `[thing] [action] [reason]. [next step].`

Example:

```
# Before (normal English, 47 tokens)
The proposal I've drafted adds a dark mode toggle to the settings page.
This will affect the theme.js file and the user preferences store.
There's a potential risk around cache invalidation when the theme
switches at runtime.

# After (ULTRA, 22 tokens — 53% reduction)
Change: dark-mode toggle.
Files: settings.py, theme.js, user_prefs.json.
Risk: cache invalidation on runtime switch.
```

---

### LITE (User-Facing Responses)

Used for: chat responses, executive summaries in return envelopes, status updates.

Rules:
- Remove filler ("let me...", "so what we have is...", "basically")
- Grammar intact — no telegraphic fragments
- Lead with the answer or decision
- Close with next action or open question

Example:

```
# Before (normal English, 58 tokens)
I've finished drafting the proposal for the dark mode feature. Let me
walk you through what I've found. The change affects three files and
there's a main risk around theme state persistence across browser tabs.
I've included a rollback strategy using a feature flag. Would you like
to proceed to the spec phase now?

# After (LITE, 33 tokens — 43% reduction)
Proposal ready. Scope: dark mode toggle. 3 files affected. Main risk:
theme state persistence across tabs. Rollback: feature flag. Next:
sdd-spec or sdd-design?
```

---

### Normal English (Exceptions)

Used for: code output, commit messages, PR descriptions, security warnings, irreversible action confirmations, multi-step sequences where fragment order risks misread.

Rules:
- Full sentences, proper grammar
- Industry-standard formatting (markdown, code fences)
- Clarity over brevity

Example (security warning):

```
WARNING: This will delete the production database. This action is
irreversible. A backup was created at 2026-04-17T14:30:00Z. To proceed,
type the exact phrase: "DELETE PRODUCTION DB"
```

---

## When Each Mode Fires

| Output Target | Mode | Example |
|--------------|------|---------|
| Engram `mem_save` content | ULTRA | Proposal, spec, design artifacts |
| Context pack | ULTRA | Protected facts, active constraints |
| Sub-agent launch prompt | ULTRA | Task instructions passed down |
| Return envelope `executive_summary` | LITE | User-visible status |
| Chat response to user | LITE | All direct messages to the user |
| Code block content | Normal | `.py`, `.js`, `.sql`, etc. |
| Commit message | Normal | Git commit body |
| PR description | Normal | GitHub PR body |
| Security warning | Normal | Auth errors, data loss warnings |
| Irreversible confirm | Normal | "Are you sure?" for destructive actions |
| Multi-step sequence | Normal | Numbered instructions where order matters |

## Why No Full "Monkey Speak" Mode

The original caveman library offers ULTRA mode optionally everywhere. V3 deliberately restricts where ULTRA applies because:

1. **Code fragments are incorrect code** — breaking grammar breaks syntax
2. **Security warnings must be unambiguous** — a compressed warning may be misread
3. **Commit messages are audited** — team members reading them deserve readability
4. **PR descriptions explain business value** — fragments fail this purpose

ULTRA stays internal. LITE covers user-facing. Normal English covers everything else.

## User Override

The user can request mode changes:

- `"stop caveman"` or `"normal mode"` — disables LITE for user-facing responses (reverts to normal English conversation)
- `"verbose"` — same effect as above
- `"ok back to compressed"` — re-enables LITE
- Persona files will always preserve ULTRA for internal artifacts regardless

## How It's Implemented

The caveman dual-mode is injected into the agent's persona file:

```markdown
## Output Compression (Caveman Dual-Mode)

Sub-agent internal work (thinking, artifacts to Engram, context packs):
  ULTRA mode. Telegraphic. Drop articles, filler, pleasantries.
  Pattern: [thing] [action] [reason]. [next step].

User-facing responses (chat, executive summaries, status updates):
  LITE mode. No filler, grammar intact, professional concise.

Exceptions — use normal English for:
  Security warnings. Irreversible action confirmations.
  Code, commits, PRs. Multi-step sequences where fragment order risks misread.

This is ACTIVE EVERY RESPONSE. No revert after many turns.
Off only when user explicitly says "stop caveman" or "normal mode".
```

This block is inserted in:
- `internal/assets/claude/persona-architect.md`
- `internal/assets/{agent}/persona-architect.md` for each of the 8 agents
- All persona variants (persona-neutral.md, etc.) per the master plan Appendix A

## Sub-Agent Behavior

Sub-agents read the persona's caveman block before producing output. Their own artifact generation applies:

- Engram persistence calls: content in ULTRA
- Return envelope `executive_summary` field: LITE
- Return envelope `detailed_report`: mode depends on whether the orchestrator will forward to user (LITE) or use internally (ULTRA)

The `sdd-phase-common.md` section E makes this explicit:

```
## E. Caveman Output Mode

When producing artifacts, apply caveman compression per the persona file:

- Artifacts stored to Engram / OpenSpec: ULTRA mode
- `executive_summary` field in return envelope: LITE mode
- Code, commits, PRs: Normal English
- Security warnings, irreversible action confirmations: Normal English
```

## Measured Impact

In our internal testing:
- Typical SDD phase artifact: ~30% reduction in token count with ULTRA
- Typical user response: ~40% reduction with LITE
- Session totals: ~35% reduction in token usage over a 20-turn session
- Qualitative: users report improved readability — less noise, answer comes first

No measurable degradation in task success rates.

## Anti-Patterns

- Applying ULTRA to user-facing responses — breaks readability
- Applying LITE to code — breaks syntax
- Applying ULTRA to security warnings — dangerous ambiguity
- Applying ULTRA to PR descriptions — fails audit review purpose
- Reverting to verbose mode after a few turns — caveman is persistent until user stops it explicitly
- Dropping caveman block from persona to "simplify" — the persona file is the on-switch

## Related

- `docs/cognitive-modes.md` — how cognitive postures interact with output style
- `internal/assets/claude/persona-architect.md` — caveman injection reference
- `internal/assets/claude/output-style-architect.md` — dual-mode formatting rules
- `internal/assets/skills/_shared/sdd-phase-common.md` section E — sub-agent enforcement

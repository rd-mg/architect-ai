# Caveman Integration (Dual-Mode Output Compression)

**Scope**: How the orchestrator compresses its own output and sub-agent prompts without losing information, and when it switches back to normal English.

---

## What "Caveman" means

"Caveman" mode is a compression discipline — drop articles, filler, and grammar niceties that don't carry information. It is not a persona ("me Tarzan, me code"), it is a terse writing style:

**Normal English**:
> I'm going to start by reading the existing models to understand the relationships between them, then I'll propose a small refactor that separates the data access layer from the business logic.

**Caveman ULTRA**:
> Read models. Infer relations. Propose split: data access vs business logic.

**Caveman LITE**:
> Reading existing models, then proposing data-access / business-logic split.

---

## Two modes, three registers

The orchestrator runs in **dual mode** — different contexts get different registers:

| Register | When | Who sees it |
|----------|------|-------------|
| Normal English | Code, commits, PRs, security warnings, irreversible confirmations | User + VCS history |
| LITE caveman | User-facing summaries, status updates, phase transitions | User |
| ULTRA caveman | Internal artifacts (Engram, context packs, state), sub-agent prompts | Orchestrator + sub-agents |

The split matters because:
- **Normal English** where clarity trumps brevity (code, warnings)
- **LITE** where the user is reading and expects professional tone but hates filler
- **ULTRA** where only the model reads it; every token counts, grammar is a liability

---

## Switching

Caveman is ALWAYS active. It turns off ONLY on explicit user request:
- "stop caveman"
- "normal mode"
- "hablame normal" / "habla normal"

Once off, it stays off for the rest of the session. The user can turn it back on with "caveman mode" / "modo caveman".

Override for code blocks: **code is always normal English** regardless of caveman state. This includes:
- Code comments
- Commit messages
- PR descriptions
- SQL/DDL
- Shell commands (users need to copy-paste them correctly)

---

## What to drop, what to keep

### Drop (LITE and ULTRA)

- Articles: *a*, *an*, *the* — where unambiguous
- Filler: *let me*, *I'll now*, *first of all*, *in this case*
- Restatements: *"You asked me to X. X is..."*
- Conversational softeners: *please*, *kindly*, *just*, *maybe*
- Hedges that add no information: *I think*, *it seems*, *probably*

### Keep (LITE and ULTRA)

- Nouns, verbs, numbers
- Technical terms
- Negations (*not*, *no*) — meaning-carrying
- Connectors when they carry logic: *because*, *unless*, *if*
- Proper names

### ULTRA-only drops

- Subject pronouns: *I*, *you*, *we* — usually implicit
- Prepositions where context disambiguates: *to*, *with*, *for*
- Imperative markers: *please do* → just the verb

---

## Anti-patterns

**❌ Pidgin / "me Tarzan"**
> Me read file. Me fix bug. Me commit.

**❌ Over-compression that loses meaning**
> models split → ambiguous: "split the models" or "models were split"?

**❌ Caveman in code comments**
```python
# calc total. handle null. return int.
```
Keep code comments normal:
```python
# Computes the order total, returning 0 for an empty cart.
```

**❌ Caveman in error messages or security warnings**
```
bad input. aborted.
```
Normal:
```
Input failed schema validation. Aborting to prevent data corruption.
```

---

## Sizing expectations per SDD phase

| Phase | Budget (user-facing, LITE) | Sub-agent prompt (ULTRA) |
|-------|---------------------------:|-------------------------:|
| sdd-init | 150 words | 300 words |
| sdd-onboard | 500+ words (conversational) | 400 words |
| sdd-explore | 600 words | 800 words |
| sdd-propose | 450 words | 700 words |
| sdd-spec | 1000 words | 1200 words |
| sdd-design | 800 words | 1200 words |
| sdd-tasks | 530 words | 700 words |
| sdd-apply | (variable by task) | 900 words |
| sdd-verify | 700 words | 900 words |
| sdd-archive | 200 words | 400 words |

Budgets are soft targets — exceeding them by 20% is fine; doubling them is a smell.

---

## Interaction with postures

Postures dictate *what* the sub-agent thinks; caveman dictates *how much space* it takes. They compose:

```
+++Critical                          ← posture (from cognitive-mode)
[critical block]                     ← posture prefix

## Project Standards (auto-resolved)
[compact rules — ULTRA]              ← caveman ULTRA

## Task
[task description — ULTRA]           ← caveman ULTRA
```

The user sees (LITE):
> Running sdd-propose with Critical posture. Will evaluate feasibility against current tech stack and return 2-3 candidates with tradeoffs.

---

## Troubleshooting

**"Output feels robotic / Google-Translate-y"** → the sub-agent is in ULTRA where LITE was expected. Check that the phase protocol routes user-facing summary generation through the orchestrator (LITE), not the sub-agent (ULTRA).

**"User complained 'what?'"** → caveman dropped a needed article or pronoun. Fix: the ambiguity is load-bearing — rewrite that one sentence as normal English, leave the rest LITE.

**"Code has Tarzan comments"** → caveman leaked into code generation. Check the `persona-architect.md` for an override that forces normal English in code outputs.

---

## See also

- `internal/assets/claude/output-style-architect.md` — formatting rules (dual-mode block)
- `internal/assets/claude/persona-architect.md` — caveman activation block
- `cognitive-modes.md` — how postures compose with caveman

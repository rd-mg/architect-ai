# Research Routing Policy (Shared Fragment)

**Scope**: Any skill or phase protocol deciding between external research tools MUST follow this priority order.

---

## Priority (5-Step)

```
1. Engram               ← past decisions & previous findings (fastest)
2. Local ripgrep        ← repo itself (or ripgrep-odoo for Odoo projects)
3. NotebookLM           ← version-specific curated knowledge
4. Context7             ← library/framework official documentation
5. Web Search           ← LAST, only on EXPLICIT user request
```

No other order acceptable. Deviation requires explicit user approval.

---

## Decision tree

```
Question asked?
  │
  ├─ Step 1: Engram (mem_search)
  │    If hit and fresh (<168h) → use it, STOP.
  │
  ├─ Step 2: Local Code (ripgrep)
  │    Is it about THIS repo? → rg local
  │    Is it about Odoo? → ripgrep-odoo (~/gitproj/odoo/)
  │    If found → use it, STOP.
  │
  ├─ Step 3: NotebookLM (notebooklm_query)
  │    Is it version-specific or curated external? → Query notebook
  │    If hit → use it, STOP.
  │
  ├─ Step 4: Context7 (context7_resolve)
  │    Is it framework/library specific? → Query Context7
  │    If hit → use it, STOP.
  │
  └─ Step 5: Web Search
       ONLY on explicit user trigger ("search web", "google it").
```

---

## Step 1 — Engram (ALWAYS FIRST)

Call `mem_search` with most specific topic_key.
- Pattern found: USE IT. Skip all other steps.
- No relevant result: Proceed to Step 2.

---

## Step 2 — Local ripgrep / ripgrep-odoo

Use when question is about function signatures, types, or call sites.
- **Odoo projects**: Use `ripgrep-odoo` (base path: `~/gitproj/odoo/`) to see HOW Odoo implements something.
- **Other projects**: Use local `ripgrep` to walk repo tree.

---

## Step 3 — NotebookLM

Use when version-specific changes, migration guides, or external library behavior not found in local code.

**NotebookLM is query-only** (enforced in V3). Never create notebooks or write artifacts during research.

---

## Step 4 — Context7

Use when question is about framework, library, or language feature (NOT our code), and Steps 1-3 failed, OR question is obviously external (e.g., "how does React's `useTransition` behave under Suspense?").

---

## Step 5 — Web Search (LAST RESORT)

**Only when user explicitly asks.** Trigger phrases:
- "search the web", "look online", "check the internet", "google it"
- "busca en internet", "busca online"

Absent explicit trigger, do NOT call `web_search` / `web_fetch`.

---

## Rationale

Engram-first: don't repeat work. Local-second: stay grounded in actual codebase. NotebookLM: high-quality curated knowledge. Context7: official reference. Web search: safety valve for unindexed items.

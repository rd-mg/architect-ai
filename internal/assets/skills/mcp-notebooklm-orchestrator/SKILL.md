---
name: mcp-notebooklm-orchestrator
description: >
  OPTIONAL research source. Curated project knowledge base. Query-only — never
  creates notebooks or writes artifacts. FIRST choice for any question that
  might live in a curated notebook, BEFORE local ripgrep and BEFORE Context7.
  Internet never used unless user explicitly requests it.
license: Apache-2.0
bridge: always
applies-when: "any research question, especially project-specific knowledge"
metadata:
  author: rd-mg
  version: "2.2"
---

# NotebookLM Orchestrator v2.2

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as first line of response per gate instructions in prompt.

NotebookLM = **primary** external research source. Holds curated notebooks hand-selected by team — architecture decisions, onboarding guides, vendor playbooks, Odoo-upgrade notes.

V3.1 promoted from "consulted sometimes" to **primary research authority**. Research priority order:

1. **NotebookLM** (this skill) ← you are here
2. Local code + docs (ripgrep, find, cat)
3. Context7 (framework docs)
4. Internet (explicit user request only)

See `_shared/research-routing.md` for full routing table.

---

## What this skill does (query-only)

- **Queries** existing notebooks via MCP
- **Updates notebook tags, descriptions, instructions** (metadata only)
- **Persists findings** to Engram under `knowledge/{domain}/external/{topic}`

## What this skill does NOT do (enforced)

-  Create new notebooks
-  Upload sources
-  Generate audio overviews
-  Create video overviews, mind maps, or Studio content
-  Write to notebooks

**Any create-artifact attempt must redirect** — tell user to use NotebookLM web app. Read-only on content; metadata-only edits.

---

## Query procedure

### Step 1 — Check cached findings first (saves tokens)

```
mem_search(query: "knowledge/{domain}/external/{topic-guess}", project: "{project}")
  → if found, read observation
  → if fresh (< 7d living docs, < 30d frozen), use it
  → if stale or missing, proceed to step 2
```

### Step 2 — List notebooks to find right one

```
notebooklm_list_notebooks()
  → returns array of {id, title, description, tags}
  → pick best match by title/tags/description
```

If none match, fall through to local code + docs (Step 2 of research routing).

### Step 3 — Query notebook

```
notebooklm_query(
  notebook_id: "{id}",
  query: "{user question, verbatim or paraphrased}"
)
  → returns answer + source citations
```

### Step 4 — Persist finding (MANDATORY)

```
mem_save(
  title: "knowledge/{domain}/external/{topic}",
  topic_key: "knowledge/{domain}/external/{topic}",
  type: "external-research",
  project: "{project}",
  content: "Q: {question}\nA: {answer}\nSources: {citations}\nNotebook: {notebook-id}\nDate: {iso-date}"
)
```

Next session same query hits Engram cache instead of re-querying NotebookLM.

---

## AFTER QUERY MANDATORY HOOK

After EVERY `notebooklm_query`, execute `mem_save` immediately. Failure = violation of knowledge-persistence contract. Ensures external knowledge internalized, project context grows robust over time.

**Pattern**: `mem_save(topic_key: "knowledge/{domain}/external/{topic}", ...)`

---

## Odoo-specific instruction pattern

When active overlay is Odoo (detected via `.atl/overlays/odoo-*//`), add code-first constraint to every NotebookLM query:

```
Query prefix: "Answer with code-first examples from Odoo source. Quote model
names, field names, decorators verbatim. Cite file path (addons/x/models/y.py)
when possible."
```

Prevents NotebookLM from returning marketing-style prose when user wants code.

---

## When NotebookLM is NOT the right tool

- **This repo's actual code** → use `ripgrep`, not NotebookLM.
- **Framework public API** (React, Go stdlib, Django) → use Context7 directly.
- **Notebook empty or list returned nothing** → fall through to local code + docs.

Do NOT pad queries with filler hoping for hit. If topic doesn't match, move on.

---

## Caching contract

Every NotebookLM response persisted in Engram. Topic-key format:

```
knowledge/{domain}/external/{topic-slug}
```

`{domain}` = sanitized notebook name (e.g. `odoo-migration`, `architecture`).

Before calling NotebookLM, ALWAYS check Engram with `mem_search`. Orchestrator pays for NotebookLM calls; Engram is free.

Staleness rules:
- Living docs (team conventions, onboarding): re-query every 7 days
- Frozen docs (version-locked Odoo upgrade notes): re-query every 30 days
- Research queries (one-off findings): never re-query — Engram authoritative

---

## Return envelope (sub-agent result)

```
{
  "source": "notebooklm",
  "notebook_id": "...",
  "query": "...",
  "answer_summary": "2-3 sentences",
  "citation_count": N,
  "engram_key": "knowledge/{domain}/external/{topic}",
  "cached_hit": true|false
}
```

Orchestrator uses `cached_hit` to compute token-savings banner.

---

## Failure modes

- **NotebookLM MCP unavailable** → `{source: "notebooklm", status: "unavailable"}`, falls through to local research.
- **No matching notebook** → `{source: "notebooklm", status: "no-match"}`, falls through.
- **Rate limit** → `{source: "notebooklm", status: "rate-limited", retry_after_s: N}`, orchestrator waits or falls through.

---

## See also

- `_shared/research-routing.md` — 4-step priority order
- `mcp-context7-skill/SKILL.md` — tertiary research, defers to this skill
- `ripgrep/SKILL.md` — secondary (local) research

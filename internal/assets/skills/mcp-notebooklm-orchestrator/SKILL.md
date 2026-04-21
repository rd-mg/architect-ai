---
name: mcp-notebooklm-orchestrator
description: >
  PRIMARY research source. NotebookLM holds the project's curated knowledge
  base. Query-only — never creates notebooks or writes artifacts. For any
  question that might live in a curated notebook, NotebookLM is the FIRST
  choice, BEFORE local ripgrep and BEFORE Context7. Internet is never used
  unless the user explicitly requests it.
license: Apache-2.0
bridge: always
applies-when: "any research question, especially project-specific knowledge"
metadata:
  author: rd-mg
  version: "2.2"
---

# NotebookLM Orchestrator v2.2

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.


NotebookLM is the **primary** external research source for this project. It holds curated notebooks hand-selected by the team — architecture decisions, onboarding guides, vendor playbooks, Odoo-upgrade notes, and anything else the team decided deserved a durable home.

In V3.1, this skill was promoted from "consulted sometimes" to **primary research authority**. The research priority order is:

1. **NotebookLM** (this skill) ← you are here
2. Local code + docs (ripgrep, find, cat)
3. Context7 (framework docs)
4. Internet (only on explicit user request)

See `_shared/research-routing.md` for the full routing table.

---

## What this skill does (query-only)

- **Queries** existing notebooks via MCP
- **Updates notebook tags, descriptions, and instructions** (metadata only)
- **Persists findings** back to Engram under `knowledge/{domain}/external/{topic}`

## What this skill does NOT do (enforced)

- ❌ Create new notebooks
- ❌ Upload sources
- ❌ Generate audio overviews
- ❌ Create video overviews, mind maps, or Studio content
- ❌ Write to notebooks

**Any attempt to create artifacts must be redirected** — tell the user to go to the NotebookLM web app. This skill is read-only on content; it may only edit metadata.

---

## Query procedure

### Step 1 — Check cached findings first (saves tokens)

```
mem_search(query: "knowledge/{domain}/external/{topic-guess}", project: "{project}")
  → if found, read the observation
  → if the finding is fresh (< 7 days old for living docs, < 30 days for frozen), use it
  → if stale or missing, proceed to step 2
```

### Step 2 — List notebooks to find the right one

```
notebooklm_list_notebooks()
  → returns array of {id, title, description, tags}
  → pick the best match by title/tags/description
```

If none match the topic, fall through to Step 2 of research routing (local code + docs).

### Step 3 — Query the notebook

```
notebooklm_query(
  notebook_id: "{id}",
  query: "{user question, verbatim or paraphrased}"
)
  → returns answer + source citations
```

### Step 4 — Persist the finding (MANDATORY)

```
mem_save(
  title: "knowledge/{domain}/external/{topic}",
  topic_key: "knowledge/{domain}/external/{topic}",
  type: "external-research",
  project: "{project}",
  content: "Q: {question}\nA: {answer}\nSources: {citations}\nNotebook: {notebook-id}\nDate: {iso-date}"
)
```

Next session asking the same question hits the Engram cache instead of re-querying NotebookLM.

---

## AFTER QUERY MANDATORY HOOK

After EVERY call to `notebooklm_query`, you MUST execute `mem_save` immediately. Failure to do so is a violation of the knowledge-persistence contract. This ensures that external knowledge is internalized and the project context grows more robust over time.

**Pattern**:
`mem_save(topic_key: "knowledge/{domain}/external/{topic}", ...)`

---

## Odoo-specific instruction pattern

When the active overlay is Odoo (detected via `.atl/overlays/odoo-*//`), add a code-first constraint to every NotebookLM query:

```
Query prefix: "Answer with code-first examples from the Odoo source. Quote model
names, field names, and decorators verbatim. Cite the file path (addons/x/models/y.py)
when possible."
```

This prevents NotebookLM from returning marketing-style prose when the user wants code.

---

## When NotebookLM is NOT the right tool

- **The question is about THIS repo's actual code** — use `ripgrep`, not NotebookLM.
- **The question is about a framework's public API** (React, Go stdlib, Django) — use Context7 directly.
- **The notebook is empty or the notebooks list returned nothing** — fall through to local code + docs.

Do NOT pad NotebookLM queries with filler in hopes of getting a hit. If the topic doesn't match, move on.

---

## Caching contract

Every NotebookLM response gets persisted in Engram. The topic-key format is:

```
knowledge/{domain}/external/{topic-slug}
```

`{domain}` is the sanitized name of the notebook (e.g. `odoo-migration`, `architecture`).

Before calling NotebookLM, ALWAYS check Engram first with `mem_search`. The orchestrator pays for NotebookLM calls; Engram is free.

Staleness rules:
- Living docs (team conventions, onboarding): re-query every 7 days
- Frozen docs (version-locked Odoo upgrade notes): re-query every 30 days
- Research queries (one-off findings): never re-query — Engram is authoritative

---

## Return envelope (in sub-agent result)

```
{
  "source": "notebooklm",
  "notebook_id": "...",
  "query": "...",
  "answer_summary": "2-3 sentences",
  "citation_count": N,
  "engram_key": "knowledge/{domain}/external/{topic}",
  "cached_hit": true|false  // was it served from Engram?
}
```rue|false  // was it served from Engram?
}
```

The orchestrator uses `cached_hit` to compute the token-savings banner (see Phase E of V3.1 plan).

---

## Failure modes

- **NotebookLM MCP not available** → return `{source: "notebooklm", status: "unavailable"}`, orchestrator falls through to local research.
- **No matching notebook** → return `{source: "notebooklm", status: "no-match"}`, orchestrator falls through.
- **Rate limit** → return `{source: "notebooklm", status: "rate-limited", retry_after_s: N}`, orchestrator waits or falls through.

---

## See also

- `_shared/research-routing.md` — the 4-step priority order
- `mcp-context7-skill/SKILL.md` — tertiary research, defers to this skill
- `ripgrep/SKILL.md` — secondary (local) research

# Phase Protocol: sdd-explore

**Version**: 3.1 (adds explicit Research Routing Policy)

## Dependencies
- **Reads**: project context (`sdd-init`), existing exploration if any
- **Writes**: `sdd/{change-name}/explore` artifact

## Cognitive Posture
+++Socratic — reveal what has NOT been said; formulate 3 questions before producing artifacts.

## Model
sonnet — reads code structurally; heavier reasoning not required.

## Word Budget
600 words (LITE mode user-facing summary); sub-agent prompt budget 800 words.

## When Triggered

- User invokes `/sdd-explore <topic>` explicitly
- Orchestrator resolves natural-language intent "explore X" (including Spanish equivalent "investiga X") <!-- trigger-phrase-allowlist -->
- Orchestrator runs as first phase of `/sdd-new`

---

## Research Routing Policy (MANDATORY — enforced in this phase)

This phase does more research than any other. It MUST follow the priority order defined in `_shared/research-routing.md`:

```
1. NotebookLM           ← FIRST — curated project knowledge
2. Local code + docs    ← SECOND — ripgrep, find, cat, extract-text
3. Context7             ← THIRD — framework/library official docs
4. Internet             ← ONLY on EXPLICIT user request
```

### Decision tree for this phase

```
User asks about a topic. Is it project-specific?
  YES → Try NotebookLM first:
           1. mem_search(query: "notebooklm/") — any indexed notebooks?
           2. If yes, notebooklm_list_notebooks()
           3. Query the best-matching notebook
           4. If hit → use result, persist to engram, report source: "notebooklm"
           5. If miss → fall through
        Then local code:
           6. ripgrep for symbols/concepts
           7. If hit → use result, report source: "ripgrep"
           8. If miss → fall through

  NO (framework/library question) → Context7 directly:
           1. context7_resolve, context7_get_docs
           2. Persist under `context7/{fw}/{ver}/{topic}`
           3. Report source: "context7"
```

**Internet**: NOT used unless the user's message contains an explicit trigger phrase ("search the web", "look online", "google it", "busca en internet"). If the user did not explicitly ask, do NOT call `web_search` / `web_fetch` — return what you have and note that internet was not searched.

---

## Sub-Agent Launch Template

```
+++Socratic
Before producing artifacts, formulate 3 questions about unstated assumptions
in the exploration topic. Reveal what has NOT been said. Examples: scope
boundary, data ownership, backward compatibility, performance constraints.
Present the questions; do NOT assume answers.

## Project Standards (auto-resolved)
{mandatory: ripgrep, bash-expert, notebooklm, context-guardian}
{task-matched skills}

## Research Routing Policy
{content of _shared/research-routing.md}

## Available Tools
{verified tool list}

## Phase: sdd-explore

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Task: Explore the topic "{topic}" and return:
  1. Current state (3-5 bullet facts about how things work today)
  2. Unknowns (the 3 Socratic questions and their importance)
  3. References (files, symbols, notebooks, external docs consulted)
  4. Next-phase recommendation (propose? more exploration? something else?)

## Research Procedure
1. FIRST: mem_search for cached findings. If cache is fresh, use it.
2. SECOND: NotebookLM. Query + persist under notebooklm/{nb}/{topic}.
3. THIRD: Local ripgrep. Walk the repo. Persist key snippets.
4. FOURTH: Context7 ONLY if the question is framework-specific AND 2+3 gave nothing.
5. NEVER: Internet, unless user message contains an explicit trigger.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd/{change-name}/explore",
  topic_key: "sdd/{change-name}/explore",
  type: "exploration-artifact",
  project: "{project}",
  content: "Topic: {topic}\nCurrent state: ...\nUnknowns: ...\nReferences: ...\nSources consulted: [{ordered list — notebooklm, ripgrep, ...}]\nNext-phase recommendation: ..."
)

## Return Envelope per sdd-phase-common.md Section D
Include NEW field: research_sources_used: [<ordered list>]
Example: ["notebooklm", "ripgrep"]  (means notebooklm hit, then ripgrep consulted)
Example: ["context7"]                (means went straight to Context7)
```

---

## Result Processing

The orchestrator inspects `research_sources_used`:

- If `["internet"]` appears and the user did not explicitly request it → the sub-agent violated routing. Re-prompt with stricter instructions.
- If the ordered list starts with `["context7"]` for a project-specific question → suggest the user ask the team to add a NotebookLM notebook for that topic.
- Append the ordered list to the exploration artifact; `sdd-archive` uses it for the final report.

---

## Failure Handling

- If NotebookLM is unavailable → note in return envelope (`"notebooklm": "unavailable"`) and fall through to local.
- If the repo has no matching files AND Context7 has no matching library → return `status: "partial"` with a request to the user: either provide more context, authorize internet search, or narrow the question.
- If the sub-agent returns fewer than 2 sources consulted on a non-trivial topic → re-delegate with expanded scope.

---

## Interactive Mode Behavior

After the exploration result returns, in interactive mode:

1. Show concise summary (LITE caveman, ≤ 600 words)
2. List the 3 Socratic questions back to the user
3. Ask: "Answer any of these, refine the topic, or proceed to sdd-propose?"

User's answer feeds into the next phase's context.

---

## Automatic Mode Behavior

In automatic mode, `sdd-explore` writes the artifact and IMMEDIATELY passes control to `sdd-propose`. The 3 Socratic questions are included in the propose sub-agent's input so it can make them concrete proposals rather than leaving them open.

---

## See also

- `_shared/research-routing.md` — the priority order
- `mcp-notebooklm-orchestrator/SKILL.md` — how to query NotebookLM
- `ripgrep/SKILL.md` — local search patterns
- `mcp-context7-skill/SKILL.md` — framework docs
- `cognitive-mode/SKILL.md` — the +++Socratic posture

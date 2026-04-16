---
name: mcp-notebooklm-orchestrator
description: >
  Orchestrates research-first NotebookLM workflows through the active MCP tool
  surface while keeping context recovery minimal, notebook selection explicit,
  and artifact generation confirmation-first. Trigger: When the user asks to
  investigate, synthesize, or generate a NotebookLM artifact such as audio,
  video, report, quiz, flashcards, mind map, slide deck, infographic, or data
  table.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.0"
---

## Purpose
Orchestrate long-context NotebookLM work without smuggling repository-specific
defaults into the bundled skill. The skill should inspect the available tool
surface, recover only the context it actually needs, select the right notebook,
and recommend artifact creation only when it materially helps.

## When to Use
- User asks to investigate a topic before generating a NotebookLM artifact.
- User asks for a NotebookLM-generated audio overview, report, quiz,
  flashcards, mind map, slide deck, infographic, or data table.
- User asks to compare notebook content, import research, or choose the right
  notebook before synthesis.
- User asks to use NotebookLM MCP tooling or notebooklm-mcp-cli for
  long-context analysis.
- Do not use when: the user only needs a normal documentation lookup, a plain
  repository answer, or a wrapper-only integration workflow.

## Critical Patterns
- Start with tool discovery and routing. Use the actual NotebookLM MCP tool
  surface that exists in the current host instead of assuming tool names.
- Resolve the notebook target in this order:
  1. user-provided notebook ID or URL
  2. a project-specific default notebook recovered from local instructions or
     memory in the current environment
  3. notebook inspection across the available notebooks
- Do not hardcode a notebook URL, notebook ID, or repository-specific default
  notebook into this skill.
- Query or inspect first when context is weak. Prefer notebook discovery,
  notebook description, source inspection, or research workflows before
  artifact generation.
- Apply progressive disclosure: recover only the minimum useful local or memory
  context needed to fill NotebookLM parameters.
- Artifact creation is recommendation-first. Suggest the right artifact, but do
  not launch generation until the user explicitly confirms.
- Treat asynchronous generation as scheduled work, not immediate output. Do not
  claim an artifact is ready until a later status or download step proves it.
- Keep wrapper-only details as implementation concerns outside the bundled
  contract. This skill orchestrates NotebookLM usage; it does not define
  Open WebUI glue, deployment-specific identity mapping, or project-local
  wrappers.
- Stay domain-agnostic. This bundled skill must not contain domain-specific or
  repository-specific assumptions.

## Inputs
- The user goal: investigate, query, ingest, generate, export, or automate.
- Optional notebook identifier provided by the user.
- Optional project-specific notebook hints recovered from the current project.
- The live NotebookLM MCP tool surface exposed by the current host.
- The minimum context needed to answer or generate the requested artifact.

## Outputs
- One validated NotebookLM workflow choice.
- One direct answer for synchronous notebook queries when that is enough.
- One artifact recommendation when generation would help but has not been
  confirmed yet.
- One scheduled generation request only after explicit user confirmation.
- One concise user-facing status that distinguishes investigation,
  recommendation, generation, and export phases.

## Steps
1. Identify whether the user needs investigation, direct notebook querying,
   notebook discovery, artifact generation, export, or automation.
2. Inspect the active NotebookLM tool surface and route to the tools that
   really exist.
3. Resolve the notebook target using explicit user input first, then any local
   project-specific hint, then notebook inspection.
4. Recover only the minimum local context needed to inform notebook selection
   or generation parameters.
5. Prefer synchronous notebook answers for normal information requests.
6. If context is weak or the correct notebook is unclear, investigate before
   recommending or creating any artifact.
7. Recommend one artifact only when it clearly improves the user outcome.
8. Stop and wait for confirmation before any artifact-generation request.
9. After confirmation, submit exactly one generation request with supported
   parameters only.
10. Report the correct handoff: answer returned, recommendation pending,
    generation scheduled, status checked, or artifact ready for export.

## Guardrails
- Do not invent NotebookLM tool names, parameters, or workflow steps.
- Do not assume a project default notebook unless the current environment
  explicitly provides one.
- Do not auto-create artifacts or queue multiple artifacts by default.
- Do not promise a finished artifact from the same turn that scheduled it.
- Do not embed domain-specific, repository-specific, or deployment-wrapper
  logic into the bundled skill contract.

## Resources
- `docs/components.md`
- `docs/intended-usage.md`
- `internal/assets/skills/mcp-notebooklm-orchestrator/`
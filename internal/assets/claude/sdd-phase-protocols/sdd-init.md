# Phase Protocol: sdd-init

## Dependencies
- **Reads**: project files for detection
- **Writes**: `sdd-init` artifact (project context)

## Cognitive Posture
None — detection and configuration.

## Model
sonnet — needs to read project structure intelligently

## When Triggered
- User invokes `/sdd-init` explicitly
- Orchestrator auto-runs when `mem_search(query: "sdd-init/{project}")` returns nothing before any other SDD command

## Sub-Agent Launch Template

```
## Project Standards (auto-resolved)
{matching compact rules — will likely be empty on first init}

## Available Tools
{verified tool list}

## Phase: sdd-init

Task: Detect project context and persist for the session.

## Detection Procedure
1. Project name: use the directory name or package.json/manifest name
2. Language/framework: detect from build files (package.json, __manifest__.py, go.mod, pyproject.toml, Cargo.toml, etc.)
3. Test runner: detect from scripts in build files (pytest, jest, go test, cargo test, etc.)
4. Artifact store mode: default to `engram` if Engram available, else `none`
5. Strict TDD capability:
   - Look for .strict-tdd file or ".strict-tdd: true" in project config
   - Look for TDD-oriented test runners (jest --watch, pytest-watch, etc.)
   - Ask user if ambiguous
6. Active overlays: check for `.atl/overlays/` directory

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd-init/{project}",
  topic_key: "sdd-init/{project}",
  type: "project-context",
  project: "{project}",
  content: "Project: {name}\nLanguage: {lang}\nFramework: {framework}\nTest runner: {cmd}\nArtifact mode: {mode}\nStrict TDD: {true|false}\nActive overlays: {list}\nInit date: {date}"
)

## Return Envelope per sdd-phase-common.md Section D
```

## Result Processing

- Cache project context for the session
- Orchestrator uses this for all subsequent phase delegations
- Update state: `uninitialized` → `idle`

## Failure Handling

- If project root cannot be determined → return `blocked`, ask user to run from project root
- If detection is ambiguous → return `partial`, ask user specific questions
- If Engram is unavailable → fall back to `none` mode silently, note in return envelope

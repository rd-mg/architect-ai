# Mode-Branching Protocol

Shared artifact-store branching logic consumed by all SDD skills.
Every skill defines its own **Artifact Metadata** (Name, Topic Key, Type) and references this file for the retrieval/persistence algorithm.

## Artifact Metadata Format

Each skill declares its artifacts in a `## Persistence` block:

```
## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

- **Artifact Name**: {filename}          — e.g. apply-progress.md
- **Topic Key**: sdd/{change-name}/{type} — e.g. sdd/{change-name}/apply-progress
- **Type**: {observation type}            — usually "architecture"

- Update `tasks.md` with `[x]` marks in `openspec/hybrid` modes.
```

For skills with multiple artifacts, label them:

```
## Persistence

Follow `_shared/mode-branching.md` for artifact-store branching.

**Artifact 1**: proposal
- **Artifact Name**: proposal.md
- **Topic Key**: sdd/{change-name}/proposal
- **Type**: architecture

**Artifact 2**: tasks
- **Artifact Name**: tasks.md
- **Topic Key**: sdd/{change-name}/tasks
- **Type**: architecture
```

Callers then use `"Follow persistence rules in Step 2 using **Artifact 1** metadata."`.

---

## Step 1: Retrieval (Symmetric Resumption)

Called when a skill MUST check for prior progress before starting work.
Follow rules based on `artifact_store` mode:

### engram
1. `mem_search(query: "{topic_key}", project: "{project}")` → returns observation ID (preview is truncated)
2. `mem_get_observation(id: {id})` → full content (MANDATORY — previews are 300 chars, insufficient for resumption)
3. If no result: proceed as first execution (no prior progress)

### openspec
1. Read `openspec/changes/{change-name}/{artifact-name}` from filesystem
2. If file exists: parse content for progress markers
3. If file does not exist: proceed as first execution

### hybrid
1. Try **engram** retrieval first (see engram rules above)
2. If engram returns no result, fall back to **openspec** filesystem read
3. **Conflict rule**: If both stores have content and they disagree, filesystem is authority. Log warning, use filesystem.

### none
Use context provided in orchestrator prompt only. No external retrieval.

---

## Step 2: Persistence (Saving Artifacts)

**MANDATORY — do NOT skip.** Every phase MUST persist its artifact.
Without persistence, the next phase CANNOT find the artifact and the pipeline BREAKS.

### engram
```
mem_save(
  title: "{topic_key}",
  topic_key: "{topic_key}",
  type: "{type}",
  project: "{project}",
  content: "{full artifact markdown}"
)
```
- Uses `topic_key`-based upsert. Re-running a phase OVERWRITES previous version.
- No revision history kept.

### openspec
1. Write content to `openspec/changes/{change-name}/{artifact-name}`
2. Use **Atomic Write Pattern** (see below)
3. Update `state.yaml` with phase status
4. Files are version-controlled via git

### hybrid
1. Write to filesystem FIRST (per openspec rules above)
2. THEN persist to Engram (per engram rules above)
3. Both writes MUST succeed for operation to be complete
4. Token cost is higher — use only when both cross-session and file artifacts are needed

### none
Return results inline only. No storage. Warn user that progress will be lost on session end.

---

### Merge Protocol (Cumulative Artifacts)

For skills with cumulative artifacts (e.g., apply-progress, verify-report):

1. **Read prior**: Follow Step 1 retrieval to load existing artifact
2. **Merge**: Prepend/append new work to existing content — do NOT overwrite
3. **CUMULATIVE RULE**: Final artifact MUST include ALL previously completed work PLUS new completions
4. **Store**: Write merged content through Step 2

Example (apply-progress):
```
Prior state: 3 tasks completed, 2 pending
New work:    completes tasks 4-5
Merge:       all 5 tasks completed, 0 pending
Persist:     full merged artifact (not just delta)
```

### STORE SYNC (Cross-Store Branching)

For **hybrid** mode, after persisting to both stores:

1. Compare content between Engram and filesystem
2. If they differ:
   - **Filesystem is authority** — use filesystem content as canonical
   - Overwrite Engram with filesystem content
   - Log sync action for traceability
3. If identical: no action needed

---

## Atomic Write Pattern (Filesystem)

To prevent data loss during session interrupts:

```
1. Write content to {filename}.tmp
2. Verify write success (check file exists, non-empty)
3. Rename {filename}.tmp to {filename}
   os.Rename(tmp, target) — atomic on Unix, prevents partial writes
```

Applied in `openspec` and `hybrid` modes for ALL filesystem writes.

---

## Branching Rules (Store Divergence)

When both stores exist and must be reconciled:

| Situation | Rule |
|-----------|------|
| Engram has newer data than filesystem | Filesystem wins — Engram is scratch, filesystem is source of truth |
| Filesystem has data, Engram is empty | Sync filesystem → Engram (resume capability) |
| Engram has data, filesystem is empty | Sync Engram → filesystem (recovery from crash) |
| Both have data, content differs | Filesystem wins. Overwrite Engram. Log divergence. |
| Both have data, content matches | No action needed — stores are in sync |

### Recovery from Session Interrupt

If a session crashes mid-phase:

1. On next start, orchestrator checks `state.yaml` or Engram state observation
2. If phase status is `in_progress`:
   - Read the phase's artifact for partial progress
   - Resume from last saved state
   - Do NOT restart from scratch
3. If phase status is `pending` but artifact exists:
   - Artifact was created but state wasn't updated
   - Mark phase as `completed` and advance DAG

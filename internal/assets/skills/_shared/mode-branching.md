# Mode-Branching Protocol

### Step 1: Retrieval (Reading Context)

Follow these rules based on the `artifact_store` mode:
- **engram**: Use `mem_get_observation(id)` for previous artifacts. Use `mem_search` only to find IDs.
- **openspec**: Read from `openspec/changes/{change-name}/` or `openspec/specs/`.
- **hybrid**: Read from Engram (primary) with filesystem fallback. **Filesystem is the authority** if stores differ.
- **none**: Use provided prompt context only.

### Step 2: Persistence (Saving Artifacts)

Follow these rules based on the `artifact_store` mode:
- **engram**: Save via `mem_save` or `mem_update` using the stable `topic_key`.
- **openspec**: Write to filesystem. Update `state.yaml` using the **Atomic Write Pattern**.
- **hybrid**: Persist to BOTH Engram and filesystem. Filesystem write MUST complete first.
- **none**: Return results inline only. No storage.

---

### Atomic Write Pattern (Filesystem)
To prevent data loss during session interrupts:
1. Write content to `{filename}.tmp`
2. Verify write success
3. Rename `{filename}.tmp` to `{filename}`

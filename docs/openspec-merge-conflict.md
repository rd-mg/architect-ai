# OpenSpec Merge Conflict Runbook

If you see `merge-conflict.md` in an active change folder, `sdd-archive` refused to merge one or more delta specs because the main spec changed under them. This is a correctness guard to prevent silent data loss.

## Diagnosis

Read `openspec/changes/{name}/merge-conflict.md`. It names the domains that conflict, the expected base SHA, and the actual SHA.

## Resolution path A: Re-run sdd-spec (Recommended)

This is the cleanest path. It re-authors the delta against current main.

1. In your session, re-run `sdd-spec` for this change.
2. The agent will recompute `base_sha` from current main and rewrite the delta, preserving your intent but matching the new base.
3. Re-run `architect-ai sdd-archive-preflight {change-name}` to verify.
4. Re-run `sdd-archive`.

**Trade-off**: You may need to reconcile your delta content if the new main spec significantly changed the behavior you were modifying.

## Resolution path B: Manual edit

For small changes where re-running `sdd-spec` is overkill.

1. Identify the new SHA of the main spec:
   ```bash
   sha256sum openspec/specs/{domain}/spec.md
   ```
2. Open the delta spec in `openspec/changes/{name}/specs/{domain}/spec.md`.
3. Update the `base_sha` in the YAML front-matter with the new SHA.
4. Verify your delta body still makes sense against the updated main spec.
5. Re-run `architect-ai sdd-archive-preflight {change-name}`.

## Emergency Override

If you are absolutely certain that overwriting the main spec's current version is safe, you can use the `--force-merge` flag (if implemented in your version) or manually update the `base_sha` as described in Path B.

> [!CAUTION]
> Manual overrides bypass the safety guards. Ensure you have reviewed the `git log` of the main spec to understand what you are overwriting.

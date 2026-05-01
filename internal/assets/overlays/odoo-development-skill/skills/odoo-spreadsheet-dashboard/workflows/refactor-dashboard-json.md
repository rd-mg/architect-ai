# Workflow: Refactor or Upgrade an Existing Dashboard

## Purpose

Safely modify an existing `.osheet.json` — add scorecards, fix formulas, change models,
add global filters — without breaking the existing structure.

## Steps

1. **Profile the original** to understand current state:
   ```bash
   python3 scripts/osheet_profile.py original.osheet.json --markdown
   ```

2. **Validate the original** to establish a baseline:
   ```bash
   python3 scripts/osheet_validate.py original.osheet.json --json > baseline-errors.json
   ```

3. **Make changes** directly to the JSON (with the editor) or via `osheet_build.py` for
   structural additions. Preserve all keys the profiler found. Do not delete unknown keys.

4. **Validate the refactored output**:
   ```bash
   python3 scripts/osheet_validate.py refactored.osheet.json --strict
   ```

5. **Compare semantic** to confirm the refactored dashboard still has equivalent structure:
   ```bash
   python3 scripts/osheet_compare.py original.osheet.json refactored.osheet.json --mode semantic
   ```

6. **Compare profile** to confirm figure/pivot/filter counts are as intended:
   ```bash
   python3 scripts/osheet_compare.py original.osheet.json refactored.osheet.json --mode profile
   ```

7. **Deliver** the refactored file.

## Safety rules

- Never delete a `globalFilters` entry without confirming it is unused in all pivot/list `fields`.
- Never renumber `formulaId` without updating every matching `PIVOT.VALUE` / `PIVOT.HEADER` formula.
- Never change a `pivot` UUID key without updating `odooLinkReferences` that reference it.
- Preserve `revisionId` and `uniqueFigureIds` unless intentionally resetting version history.

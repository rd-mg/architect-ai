# Workflow: Recreate a Sample Dashboard JSON (Regression)

## Purpose

Export a lossless recipe from an existing `.osheet.json`, rebuild from that recipe, and
confirm byte-for-byte equivalence. Used to generate golden fixtures for CI regression tests.

## Steps

1. **Export the recipe**:
   ```bash
   python3 scripts/osheet_recipe.py export original.osheet.json recipe.json
   ```
   The recipe captures the full payload in a normalised, deterministic JSON form.

2. **Rebuild from the recipe**:
   ```bash
   python3 scripts/osheet_recipe.py build recipe.json recreated.osheet.json
   ```

3. **Compare exact**:
   ```bash
   python3 scripts/osheet_compare.py original.osheet.json recreated.osheet.json --mode exact
   ```
   A `PASS` means the canonical SHA-256 of the normalised payloads matches.

4. **If FAIL**: inspect the diff section in the compare output. Common causes:
   - Non-deterministic key ordering in the original (fix: normalise before export).
   - Floating-point rounding in cell values.
   - `revisionId` or `uniqueFigureIds` generated at export time.

5. **Catalog all samples**:
   ```bash
   python3 scripts/osheet_recipe.py catalog assets/sample-osheets testdata/golden/odoo-osheet/recipes
   ```

6. **Store golden fixtures** in `testdata/golden/odoo-osheet/` for CI.

## Acceptance

- All 13 reference samples pass `--mode exact` roundtrip.
- Recipes are deterministic across Python versions (3.10+).
- The catalog script produces one recipe per sample with consistent file naming.

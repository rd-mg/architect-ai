# Workflow: Recreate a Sample Dashboard JSON (Regression)

## Purpose

Export lossless recipe from existing `.osheet.json`, rebuild from that recipe, and
confirm byte-for-byte equivalence. Used to generate golden fixtures for CI regression tests.

## Steps

1. **Export recipe**:
   ```bash
   python3 scripts/osheet_recipe.py export original.osheet.json recipe.json
   ```
   Recipe captures full payload in normalised, deterministic JSON form.

2. **Rebuild from recipe**:
   ```bash
   python3 scripts/osheet_recipe.py build recipe.json recreated.osheet.json
   ```

3. **Compare exact**:
   ```bash
   python3 scripts/osheet_compare.py original.osheet.json recreated.osheet.json --mode exact
   ```
   `PASS` means canonical SHA-256 of normalised payloads matches.

4. **If FAIL**: inspect diff section in compare output. Common causes:
   - Non-deterministic key ordering in original (fix: normalise before export).
   - Floating-point rounding in cell values.
   - `revisionId` or `uniqueFigureIds` generated at export time.

5. **Catalog all samples**:
   ```bash
   python3 scripts/osheet_recipe.py catalog assets/sample-osheets testdata/golden/odoo-osheet/recipes
   ```

6. **Store golden fixtures** in `testdata/golden/odoo-osheet/` for CI.

## Acceptance

- All 13 reference samples pass `--mode exact` roundtrip.
- Recipes deterministic across Python versions (3.10+).
- Catalog script produces one recipe per sample with consistent file naming.

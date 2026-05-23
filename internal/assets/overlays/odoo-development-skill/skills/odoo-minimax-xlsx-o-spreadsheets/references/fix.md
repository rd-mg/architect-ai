# FIX — Repair Broken Formulas in Existing xlsx

EDIT task. MUST preserve all original sheets and data. Never create new workbook.

## Workflow

```bash
# Step 1: Identify errors
python3 SKILL_DIR/scripts/formula_check.py input.xlsx --json

# Step 2: Unpack
python3 SKILL_DIR/scripts/xlsx_unpack.py input.xlsx /tmp/xlsx_work/

# Step 3: Fix each broken <f> element in worksheet XML using Edit tool
#   (see Error-to-Fix mapping below)

# Step 4: Pack and validate
python3 SKILL_DIR/scripts/xlsx_pack.py /tmp/xlsx_work/ output.xlsx
python3 SKILL_DIR/scripts/formula_check.py output.xlsx
```

## Error-to-Fix Mapping

| Error | Fix Strategy |
|-------|-------------|
| `#DIV/0!` | Wrap: `IFERROR(original_formula, "-")` |
| `#NAME?` | Fix misspelled function (e.g. `SUMM` → `SUM`) |
| `#REF!` | Reconstruct broken reference |
| `#VALUE!` | Fix type mismatch |

For full list of Excel error types and advanced diagnostics, see `validate.md`.

## Critical Rules

- Output MUST contain same sheets as input. Do NOT create new workbook.
- Only modify specific `<f>` elements that are broken — everything else must be untouched.
- After packing, always run `formula_check.py` to confirm all errors resolved.

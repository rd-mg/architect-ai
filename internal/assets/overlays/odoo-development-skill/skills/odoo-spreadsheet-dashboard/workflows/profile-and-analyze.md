# Workflow: Profile and Analyze an Existing Dashboard

## Purpose

Inspect existing Odoo 19 `.osheet` / `.osps` JSON file and produce structured profile
showing its structure, formulas, data sources, figures, and Odoo model coverage.

## Steps

1. **Receive file path** from user. Accept `.osheet.json`, `.osheet`, or `.osps` extensions.

2. **Run profiler**:
   ```bash
   python3 scripts/osheet_profile.py <file> --markdown
   ```
   For directory of samples:
   ```bash
   python3 scripts/osheet_profile.py assets/sample-osheets --markdown --out references/sample-pattern-catalog.md
   ```

3. **Read profile output**. Profiler emits:
   - Top-level keys present / absent
   - Sheet names and types (Dashboard / Data / other)
   - Formula function inventory with counts
   - Pivot dictionary keys and `formulaId` values
   - List `id` values
   - Global filter types and `fields` matching
   - Figure/chart/scorecard counts and types
   - Odoo model references
   - `odoo://view` drill-down links

4. **Summarise findings** in LITE caveman: key models, dominant figures, filter count, formula
   complexity, any structural anomalies.

5. **Identify risks**: missing `formulaId`, undefined style references, malformed `odoo://view`
   payloads, scorecard without `keyValue`, chart figure without `data.type`.

6. **Optionally run validator** for richer error list:
   ```bash
   python3 scripts/osheet_validate.py <file> --json
   ```

## Output

- Markdown profile table (printed or written to `--out` path).
- Optional JSON validation report.
- Human summary of findings with risk flags.

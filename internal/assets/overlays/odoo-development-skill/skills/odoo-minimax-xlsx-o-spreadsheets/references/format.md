# Financial Formatting & Output Standards — Complete Agent Guide

> Complete reference manual for agent when applying professional financial formatting to xlsx files. All operations target direct XML surgery on `xl/styles.xml` without using openpyxl. Every operational step provides ready-to-use XML snippets.

---

## 1. When to Use This Path

This document (FORMAT path) applies to following two scenarios:

**Scenario A — Dedicated Formatting of Existing File**
User provides existing xlsx file and requests financial modeling formatting standards applied or unified. Starting point: unpack file, audit existing `styles.xml`, then append missing styles and batch-update cell `s` attributes. No cell values or formulas modified.

**Scenario B — Applying Format Standards After CREATE/EDIT**
After completing data entry or formula writing, formatting applied as final step. At this point, `styles.xml` may come from minimal_xlsx template (pre-defines 13 style slots) or from user file. In either case, follow "append only, never modify existing xf entries."

**Not applicable**: Reading or analyzing file contents only (use READ path); modifying formulas or data (use EDIT path).

---

## 2. Financial Format Semantic System

### 2.1 Font Color = Cell Role (Color = Role)

Primary convention of financial modeling: **font color encodes cell's role, not decoration**. Reviewer can glance at colors to determine which cells are adjustable parameters and which are model-calculated results. Industry-wide convention (followed by investment banks, Big Four, corporate finance teams).

| Role | Font Color | AARRGGBB | Use Case |
|------|-----------|----------|----------|
| Hard-coded input / assumption | Blue | `000000FF` | Growth rates, discount rates, tax rates, other user-modifiable parameters |
| Formula / calculated result | Black | `00000000` | All cells containing `<f>` element |
| Same-workbook cross-sheet reference | Green | `00008000` | Cells whose formula starts with `SheetName!` |
| External file link | Red | `00FF0000` | Cells whose formula contains `[FileName.xlsx]` (flagged as fragile links) |
| Label / text | Black (default) | theme color | Row labels, category headings |
| Key assumption requiring review | Blue font + yellow fill | Font `000000FF` / Fill `00FFFF00` | Provisional values, parameters pending confirmation |

**Decision tree**:
```
Does cell contain <f> element?
  +-- Yes -> Does formula start with [FileName]?
  |           +-- Yes -> Red (external link)
  |           +-- No  -> Does formula contain SheetName!?
  |                       +-- Yes -> Green (cross-sheet reference)
  |                       +-- No  -> Black (same-sheet formula)
  +-- No  -> Is value user-adjustable parameter?
              +-- Yes -> Blue (input/assumption)
              +-- No  -> Black default (label)
```

**Strictly prohibited**: Blue font + `<f>` element coexisting (color role contradiction — must be corrected).

### 2.2 Number Format Matrix

| Data Type | formatCode | numFmtId | Display Example | Applicable Scenario |
|-----------|-----------|----------|-----------------|---------------------|
| Standard currency (whole dollars) | `$#,##0;($#,##0);"-"` | 164 | $1,234 / ($1,234) / - | P&L, balance sheet amount rows |
| Standard currency (with cents) | `$#,##0.00;($#,##0.00);"-"` | 169 | $1,234.56 / ($1,234.56) / - | Unit prices, detailed costs |
| Thousands (K) | `#,##0,"K"` | 171 | 1,234K | Simplified display for management reports |
| Millions (M) | `#,##0,,"M"` | 172 | 1M | Macro-level summary rows |
| Percentage (1 decimal) | `0.0%` | 165 | 12.5% | Growth rates, gross margins |
| Percentage (2 decimals) | `0.00%` | 170 | 12.50% | IRR, precise interest rates |
| Multiple / valuation multiplier | `0.0x` | 166 | 8.5x | EV/EBITDA, P/E |
| Integer (thousands separator) | `#,##0` | 167 | 12,345 | Employee count, unit quantities |
| Year | `0` | 1 (built-in, no declaration needed) | 2024 | Column header years, prevents 2,024 |
| Date | `m/d/yyyy` | 14 (built-in, no declaration needed) | 3/21/2026 | Timelines |
| General text | General | 0 (built-in, no declaration needed) | — | Label rows, cells with no format requirement |

numFmtId 169–172 are custom formats needing to be appended beyond 4 formats (164–167) pre-defined in minimal_xlsx template. When appending, assign IDs according to rules (see Section 3.4).

**Built-in format IDs do not need to be declared in `<numFmts>`** (IDs 0–163 built into Excel/LibreOffice; simply reference numFmtId in `<xf>`):

| numFmtId | formatCode | Description |
|----------|-----------|-------------|
| 0 | General | General format |
| 1 | `0` | Integer, no thousands separator (use this ID for years) |
| 3 | `#,##0` | Thousands-separated integer (no decimals) |
| 9 | `0%` | Percentage integer |
| 10 | `0.00%` | Percentage with two decimals |
| 14 | `m/d/yyyy` | Short date |

### 2.3 Negative Number Display Standards

Financial reports have two mainstream conventions for negative numbers — choose one and **maintain consistency** throughout entire workbook:

**Parenthetical style (investment banking standard, recommended for external deliverables)**

```
Positive: $1,234    Negative: ($1,234)    Zero: -
formatCode: $#,##0;($#,##0);"-"
```

**Red minus sign style (suitable for internal operational analysis reports)**

```
Positive: $1,234    Negative: -$1,234 (red)
formatCode: $#,##0;[Red]-$#,##0;"-"
```

Rule: Once style determined, maintain it across entire workbook. Do not mix two negative number display styles within same workbook.

### 2.4 Zero Value Display Standards

In financial models, "0" and "no data" have different semantics and should be visually distinct:

| Scenario | Recommended Display | formatCode Third Segment |
|----------|-------------------|--------------------------|
| Sparse matrix (most rows have zero-value periods) | Dash `-` | `"-"` |
| Quantity counts (zero itself meaningful) | `0` | `0` or omit |
| Placeholder row (explicitly empty) | Leave blank | Do not write to cell |

Four-segment format syntax: `positive format;negative format;zero value format;text format`

Zero as dash: `$#,##0;($#,##0);"-"`
Zero preserved as 0: `#,##0;(#,##0);0`

---

## 3. styles.xml Surgical Operations

### 3.1 Auditing Existing Styles: Understanding cellXfs Indirect Reference Chain

Cell's `s` attribute points to position index (0-based) in `cellXfs`, and each `<xf>` entry in `cellXfs` references its respective definition libraries through `fontId`, `fillId`, `borderId`, and `numFmtId`.

Reference chain diagram:

```
Cell <c s="6">
    | Look up cellXfs by 0-based index
cellXfs[6] -> numFmtId="164" fontId="2" fillId="0" borderId="0"
    |            |               |          |
numFmts         fonts[2]      fills[0]   borders[0]
id=164          color=00000000  (no fill)  (no border)
$#,##0...       black
```

Audit steps:

**Step 1**: Read `<numFmts>` and record all declared custom formats and their IDs:
```xml
<numFmts count="4">
  <numFmt numFmtId="164" formatCode="$#,##0;($#,##0);&quot;-&quot;"/>
  <numFmt numFmtId="165" formatCode="0.0%"/>
  <numFmt numFmtId="166" formatCode="0.0x"/>
  <numFmt numFmtId="167" formatCode="#,##0"/>
</numFmts>
```
Record: current max custom numFmtId = 167, next available ID = 168.

**Step 2**: Read `<fonts>` and list each `<font>` by 0-based index with its color and style:
```
fontId=0 -> No explicit color (theme default black)
fontId=1 -> color rgb="000000FF" (blue, input role)
fontId=2 -> color rgb="00000000" (black, formula role)
fontId=3 -> color rgb="00008000" (green, cross-sheet reference role)
fontId=4 -> <b/> + color rgb="00000000" (bold black, header)
```

**Step 3**: Read `<fills>` and confirm fills[0] and fills[1] are spec-mandated reserved entries (never delete):
```
fillId=0 -> patternType="none" (spec-mandated)
fillId=1 -> patternType="gray125" (spec-mandated)
fillId=2 -> Yellow highlight (if present)
```

**Step 4**: Read `<cellXfs>` and list each `<xf>` entry by 0-based index with its combination:
```
index 0 -> numFmtId=0,   fontId=0, fillId=0 -> Default style
index 1 -> numFmtId=0,   fontId=1, fillId=0 -> Blue font general (input)
index 5 -> numFmtId=164, fontId=1, fillId=0 -> Blue font currency (currency input)
index 6 -> numFmtId=164, fontId=2, fillId=0 -> Black font currency (currency formula)
...
```

**Step 5**: Verify all count attributes match actual number of elements (count mismatches cause Excel to refuse to open file).

### 3.2 Safely Appending New Styles (Golden Rule: Append Only, Never Modify Existing xf)

**Never modify existing `<xf>` entries**. Modifications affect all cells already referencing that index, breaking existing formatting. Only append new entries at end.

Complete atomic operation sequence for appending new styles (all 5 steps must be executed):

**Step 1**: Determine if new `<numFmt>` needed

Built-in formats (ID 0–163) skip this step. Custom formats appended to end of `<numFmts>`:
```xml
<numFmts count="5">  <!-- count +1 -->
  <!-- Keep existing entries unchanged -->
  <numFmt numFmtId="164" formatCode="$#,##0;($#,##0);&quot;-&quot;"/>
  <numFmt numFmtId="165" formatCode="0.0%"/>
  <numFmt numFmtId="166" formatCode="0.0x"/>
  <numFmt numFmtId="167" formatCode="#,##0"/>
  <!-- Newly appended -->
  <numFmt numFmtId="168" formatCode="$#,##0.00;($#,##0.00);&quot;-&quot;"/>
</numFmts>
```

**Step 2**: Determine if new `<font>` needed

Check whether existing fonts already contain matching color+style combination. If not, append to end of `<fonts>`:
```xml
<fonts count="6">  <!-- count +1 -->
  <!-- Keep existing entries unchanged -->
  ...
  <!-- Newly appended: red font (external link role), new fontId = 5 -->
  <font>
    <sz val="11"/>
    <name val="Calibri"/>
    <color rgb="00FF0000"/>
  </font>
</fonts>
```
New fontId = count value before appending (when original count=5, new fontId=5).

**Step 3**: Determine if new `<fill>` needed

If new background color needed, append to end of `<fills>` (note: fills[0] and fills[1] must never be modified):
```xml
<fills count="4">  <!-- count +1 -->
  <fill><patternFill patternType="none"/></fill>       <!-- 0: spec-mandated -->
  <fill><patternFill patternType="gray125"/></fill>    <!-- 1: spec-mandated -->
  <fill>                                               <!-- 2: yellow highlight -->
    <patternFill patternType="solid">
      <fgColor rgb="00FFFF00"/>
      <bgColor indexed="64"/>
    </patternFill>
  </fill>
  <!-- Newly appended: light gray fill (projection period distinction), new fillId = 3 -->
  <fill>
    <patternFill patternType="solid">
      <fgColor rgb="00D3D3D3"/>
      <bgColor indexed="64"/>
    </patternFill>
  </fill>
</fills>
```

**Step 4**: Append new `<xf>` combination at end of `<cellXfs>`
```xml
<cellXfs count="14">  <!-- count +1 -->
  <!-- Keep existing entries 0-12 unchanged -->
  ...
  <!-- Newly appended index=13: currency with cents formula (black font + numFmtId=168) -->
  <xf numFmtId="168" fontId="2" fillId="0" borderId="0" xfId="0"
      applyFont="1" applyNumberFormat="1"/>
</cellXfs>
```
New style index = count value before appending (when original count=13, new index=13).

**Step 5**: Record new style index; subsequently set `s` attribute of corresponding cells in sheet XML to this value.

### 3.3 AARRGGBB Color Format Explanation

OOXML's `rgb` attribute uses **8-digit hexadecimal AARRGGBB** format (not HTML's 6-digit RRGGBB):

```
AA  RR  GG  BB
|   |   |   |
Alpha Red Green Blue
```

- Alpha channel: `00` = fully opaque (normal use value); `FF` = fully transparent (invisible, never use)
- Financial color standards always use `00` as Alpha prefix

| Color | AARRGGBB | Corresponding Role |
|-------|----------|-------------------|
| Blue (input) | `000000FF` | Hard-coded assumptions |
| Black (formula) | `00000000` | Calculated results |
| Green (cross-sheet reference) | `00008000` | Same-workbook cross-sheet |
| Red (external link) | `00FF0000` | References to other files |
| Yellow (review-required fill) | `00FFFF00` | Key assumption highlight |
| Light gray (projection period fill) | `00D3D3D3` | Distinguishing historical vs. forecast periods |
| White | `00FFFFFF` | Pure white fill |

**Common mistake**: Mistakenly writing HTML format `#0000FF` as `FF0000FF` (Alpha=FF makes color fully transparent and invisible). Correct format: `000000FF`.

### 3.4 numFmtId Assignment Rules

```
ID 0-163    -> Excel/LibreOffice built-in formats, no declaration needed in <numFmts>, reference directly in <xf>
ID 164+     -> Custom formats, must be explicitly declared as <numFmt> elements in <numFmts>
```

Rules for assigning new IDs:
1. Read all `numFmtId` attribute values in current `<numFmts>`
2. Take maximum value + 1 as next custom format ID
3. Do not reuse existing IDs; do not skip numbers

minimal_xlsx template pre-defines IDs: 164, 165, 166, 167. Next available ID is 168.

---

## 4. Pre-defined Style Index Complete Reference Table (13 Slots)

Following are 13 style slots (cellXfs index 0–12) pre-defined in minimal_xlsx template's `styles.xml`, directly referenceable in cell `s` attribute in sheet XML:

| Index | Semantic Role | Font Color | Fill | numFmtId | Format Display | Typical Use |
|-------|--------------|------------|------|----------|---------------|-------------|
| **0** | Default style | Theme black | None | 0 | General | Cells requiring no special formatting |
| **1** | Input / assumption (general) | Blue `000000FF` | None | 0 | General | Text-type assumptions, flags |
| **2** | Formula / calculated result (general) | Black `00000000` | None | 0 | General | Text concatenation formulas, non-numeric calculations |
| **3** | Cross-sheet reference (general) | Green `00008000` | None | 0 | General | Values pulled cross-sheet (general format) |
| **4** | Header (bold) | Bold black | None | 0 | General | Row/column headings |
| **5** | Currency input | Blue `000000FF` | None | 164 | $1,234 / ($1,234) / - | Amount inputs in assumptions area |
| **6** | Currency formula | Black `00000000` | None | 164 | $1,234 / ($1,234) / - | Amount calculations in model area (revenue, EBITDA) |
| **7** | Percentage input | Blue `000000FF` | None | 165 | 12.5% | Rate inputs in assumptions area (growth rate, gross margin assumptions) |
| **8** | Percentage formula | Black `00000000` | None | 165 | 12.5% | Rate calculations in model area (actual gross margin) |
| **9** | Integer (comma) input | Blue `000000FF` | None | 167 | 12,345 | Quantity inputs in assumptions area (employee count) |
| **10** | Integer (comma) formula | Black `00000000` | None | 167 | 12,345 | Quantity calculations in model area |
| **11** | Year input | Blue `000000FF` | None | 1 | 2024 | Column header years (no thousands separator) |
| **12** | Key assumption highlight | Blue `000000FF` | Yellow `00FFFF00` | 0 | General | Key parameters pending review or confirmation |

**Selection guide**:
- Determine "input" vs. "formula" -> Choose odd-numbered (input/blue) or even-numbered (formula/black) paired slots
- Determine data type -> Choose corresponding currency (5/6) / percentage (7/8) / integer (9/10) / year (11) slot
- Cross-sheet reference needing number format -> Append new green + number format combination (see Section 5.4)
- Parameter pending review -> index 12

---

## 5. Assumption Separation Principle: XML-Level Implementation

### 5.1 Structural Design

Assumption separation principle: **Input assumptions centralized in dedicated area (sheet or block); model calculation area contains only formulas, no hard-coded values**.

Recommended structure:
```
Workbook sheet layout
  sheet 1 "Assumptions"  -> All blue-font cells (style 1/5/7/9/11/12)
  sheet 2 "Model"        -> All black or green-font cells (style 2/3/4/6/8/10)
```

Same-sheet zoning approach for simple models:
```
Rows 1-5:   [Assumptions block - blue font]
Row 6:      [Empty row separator]
Rows 7+:    [Model block - black/green font formulas referencing assumptions area]
```

### 5.2 Assumptions Area XML Example

```xml
<!-- Assumptions sheet (sheet1.xml) example -->

<!-- Row 1: Block title -->
<row r="1">
  <c r="A1" s="4" t="inlineStr"><is><t>Model Assumptions</t></is></c>
</row>

<!-- Row 2: Growth rate assumption - blue font percentage input, s="7" -->
<row r="2">
  <c r="A2" t="inlineStr"><is><t>Revenue Growth Rate</t></is></c>
  <c r="B2" s="7"><v>0.08</v></c>
</row>

<!-- Row 3: Gross margin assumption - blue font percentage input, s="7" -->
<row r="3">
  <c r="A3" t="inlineStr"><is><t>Gross Margin</t></is></c>
  <c r="B3" s="7"><v>0.65</v></c>
</row>

<!-- Row 4: Base revenue - blue font currency input, s="5" -->
<row r="4">
  <c r="A4" t="inlineStr"><is><t>Base Revenue (Year 0)</t></is></c>
  <c r="B4" s="5"><v>1000000</v></c>
</row>

<!-- Row 5: Key assumption (pending review) - blue font yellow fill, s="12" -->
<row r="5">
  <c r="A5" t="inlineStr"><is><t>Terminal Growth Rate</t></is></c>
  <c r="B5" s="12"><v>0.03</v></c>
</row>
```

### 5.3 Model Area XML Example (Referencing Assumptions Area)

```xml
<!-- Model sheet (sheet2.xml) example -->

<!-- Row 1: Column headers (years) - bold header, s="4"; year cells, s="11" -->
<row r="1">
  <c r="A1" s="4" t="inlineStr"><is><t>Metric</t></is></c>
  <c r="B1" s="11"><v>2024</v></c>
  <c r="C1" s="11"><v>2025</v></c>
  <c r="D1" s="11"><v>2026</v></c>
</row>

<!-- Row 2: Revenue row -->
<row r="2">
  <c r="A2" t="inlineStr"><is><t>Revenue</t></is></c>
  <!-- B2: Base year revenue, cross-sheet reference from Assumptions, green, s="3" (general format) -->
  <!-- If currency format needed, append new style s="13" (see Section 5.4) -->
  <c r="B2" s="3"><f>Assumptions!B4</f><v></v></c>
  <!-- C2, D2: Next year revenue = prior year * (1 + growth rate), black font currency formula, s="6" -->
  <c r="C2" s="6"><f>B2*(1+Assumptions!B2)</f><v></v></c>
  <c r="D2" s="6"><f>C2*(1+Assumptions!B2)</f><v></v></c>
</row>

<!-- Row 3: Gross profit row - black font currency formula, s="6" -->
<row r="3">
  <c r="A3" t="inlineStr"><is><t>Gross Profit</t></is></c>
  <c r="B3" s="6"><f>B2*Assumptions!B3</f><v></v></c>
  <c r="C3" s="6"><f>C2*Assumptions!B3</f><v></v></c>
  <c r="D3" s="6"><f>D2*Assumptions!B3</f><v></v></c>
</row>

<!-- Row 4: Gross margin row - black font percentage formula, s="8" -->
<row r="4">
  <c r="A4" t="inlineStr"><is><t>Gross Margin %</t></is></c>
  <c r="B4" s="8"><f>B3/B2</f><v></v></c>
  <c r="C4" s="8"><f>C3/C2</f><v></v></c>
  <c r="D4" s="8"><f>D3/D2</f><v></v></c>
</row>
```

### 5.4 Appending "Green + Number Format" Combinations

Pre-defined index 3 is green font + general format. If cross-sheet reference involves currency amount, green style with number format must be appended:

```xml
<!-- Append at end of <cellXfs> in styles.xml (assuming current count=13, new index=13) -->
<!-- index 13: cross-sheet reference + currency format (green font + $#,##0) -->
<xf numFmtId="164" fontId="3" fillId="0" borderId="0" xfId="0"
    applyFont="1" applyNumberFormat="1"/>
<!-- Update count to 14 -->
```

After appending, cross-sheet reference currency cells use `s="13"`.

---

## 6. Complete Operational Workflow

### 6.1 Workflow Overview

```
[Existing xlsx or file after CREATE/EDIT]
        |
  Step 1: Unpack (extract to temporary directory)
        |
  Step 2: Audit styles.xml (review existing styles, build index mapping table)
        |
  Step 3: Audit sheet XML (identify cells needing formatting and their semantic roles)
        |
  Step 4: Append missing styles (numFmt -> font -> fill -> xf, update counts)
        |
  Step 5: Batch-update s attribute of each cell in sheet XML
        |
  Step 6: XML validity + style reference integrity verification
        |
  Step 7: Pack (recompress as xlsx)
```

### 6.2 Step 1 — Unpack

```bash
python3 SKILL_DIR/scripts/xlsx_unpack.py input.xlsx /tmp/xlsx_fmt/
```

If script unavailable, unpack manually:
```bash
mkdir -p /tmp/xlsx_fmt && cp input.xlsx /tmp/xlsx_fmt/input.xlsx
cd /tmp/xlsx_fmt && unzip input.xlsx -d unpacked/
```

### 6.3 Step 2 — Audit styles.xml

Execute according to method in Section 3.1. Quick check for minimal_xlsx template initial state:
- `<cellXfs count="13">` and `<numFmts count="4">` -> Template initial state, all 13 pre-defined slots usable directly
- Otherwise -> Complete review of existing index mapping required

### 6.4 Step 3 — Audit Sheet XML, Build Formatting Plan

Read `xl/worksheets/sheet*.xml` and evaluate each cell:
1. Does it contain `<f>` element (formula)? -> Requires black/green/red style
2. Is it hard-coded numeric parameter? -> Requires blue style
3. Is data type currency/percentage/integer/year? -> Select corresponding number format slot
4. Is it header? -> Bold style (index 4)

Build formatting mapping table: `{cell coordinate: target style index}`

### 6.5 Step 4 — Append Styles

Execute according to atomic operation sequence in Section 3.2. Update corresponding count attribute immediately after appending each component.

### 6.6 Step 5 — Batch-Update Cell s Attributes

```xml
<!-- Before formatting: no style -->
<c r="B5"><v>0.08</v></c>

<!-- After formatting: growth rate assumption, blue font percentage, s="7" -->
<c r="B5" s="7"><v>0.08</v></c>
```

```xml
<!-- Before formatting: formula without style -->
<c r="C10"><f>B10*(1+Assumptions!B2)</f><v></v></c>

<!-- After formatting: currency formula, black font, s="6" -->
<c r="C10" s="6"><f>B10*(1+Assumptions!B2)</f><v></v></c>
```

For consecutive rows of same type, row-level default styles can reduce repetition:
```xml
<!-- Entire row uses style=6, only override for exception cells -->
<row r="5" s="6" customFormat="1">
  <c r="A5" s="0" t="inlineStr"><is><t>Operating Income</t></is></c>  <!-- Text overridden to default -->
  <c r="B5"><f>B3-B4</f><v></v></c>   <!-- Inherits row-level s=6 -->
  <c r="C5"><f>C3-C4</f><v></v></c>
</row>
```

### 6.7 Step 6 — Verification

```bash
# XML validity verification handled automatically by xlsx_pack.py, no need to manually run xmllint
# Pack script validates styles.xml and sheet XML legality before packaging; aborts and reports errors

# Style audit (optional, audit entire unpacked directory after formatting complete)
python3 SKILL_DIR/scripts/style_audit.py /tmp/xlsx_fmt/unpacked/

# Formula error static scan (must specify single .xlsx file, does not accept directories)
# Pack first, then scan:
python3 SKILL_DIR/scripts/xlsx_pack.py /tmp/xlsx_fmt/unpacked/ /tmp/output.xlsx
python3 SKILL_DIR/scripts/formula_check.py /tmp/output.xlsx
```

Manual style reference integrity check:
```bash
# Find max s attribute value in sheet XML
grep -o 's="[0-9]*"' /tmp/xlsx_fmt/unpacked/xl/worksheets/sheet1.xml \
  | grep -o '[0-9]*' | sort -n | tail -1

# Compare with cellXfs count attribute (max s value must be < count)
grep 'cellXfs count' /tmp/xlsx_fmt/unpacked/xl/styles.xml
```

### 6.8 Step 7 — Pack

```bash
python3 SKILL_DIR/scripts/xlsx_pack.py /tmp/xlsx_fmt/unpacked/ output.xlsx
```

If script unavailable, pack manually:
```bash
cd /tmp/xlsx_fmt/unpacked/
zip -r ../output.xlsx . -x "*.DS_Store"
```

---

## 7. Formatting Completeness Checklist

Verify each item before delivery:

### Color Role Consistency
- [ ] All numeric cells containing `<f>` elements: fontId corresponds to black (formula) or green (cross-sheet reference)
- [ ] All hard-coded numeric values that are user-adjustable parameters: fontId corresponds to blue (input)
- [ ] Cross-sheet references (formula contains `SheetName!`): fontId corresponds to green
- [ ] External file references (formula contains `[FileName.xlsx]`): fontId corresponds to red
- [ ] No cell simultaneously contains `<f>` element and uses blue font (color role contradiction)

### Number Format Correctness
- [ ] Year columns: numFmtId="1" (`0` format), displays as 2024 not 2,024
- [ ] Currency rows: numFmtId="164" or variant, negative numbers display as ($1,234) not -$1,234
- [ ] Percentage rows: values stored as decimals (0.08 = 8%), format numFmtId="165", displays as 8.0%
- [ ] Zero values: displayed as `-` in sparse matrices rather than `0` (formatCode third segment contains `"-"`)
- [ ] Multiple rows (EV/EBITDA, etc.): numFmtId="166" (`0.0x` format)
- [ ] Negative number display style consistent throughout entire workbook (parenthetical or red minus sign)

### styles.xml Structural Integrity
- [ ] `<numFmts count>` = actual number of `<numFmt>` elements
- [ ] `<fonts count>` = actual number of `<font>` elements
- [ ] `<fills count>` = actual number of `<fill>` elements (including spec-mandated fills[0] and fills[1])
- [ ] `<cellXfs count>` = actual number of `<xf>` elements
- [ ] fills[0] is `patternType="none"`, fills[1] is `patternType="gray125"` (spec-mandated)
- [ ] All `<xf>` referenced fontId / fillId / borderId within valid range of their respective collections
- [ ] All cell `s` attribute values < `cellXfs count` (no out-of-bounds references)

### Assumption Separation Verification
- [ ] No black-font numeric cells in assumptions area/sheet (black numeric = formula, should not be in assumptions)
- [ ] No blue-font non-year numeric cells in model area/sheet (blue numeric = hard-coded, should be in assumptions)
- [ ] Input parameters in model area reference assumptions area via formulas, not by directly copying values

### Formula and Format Linkage
- [ ] All cells with `<f>` elements have explicit `s` attribute (must not use default style=0, whose font color not explicitly black)
- [ ] SUM summary rows: style uses black font + corresponding number format (e.g., s="6" for currency summaries)
- [ ] Percentage formulas: values stored as decimals, format is `0.0%`; do not multiply values by 100 before applying percentage format

### Visual Hierarchy
- [ ] Header rows (years/metric names): style=4 (bold black)
- [ ] Summary rows (Total/EBITDA/Net Income): bold + corresponding number format (append style if needed)
- [ ] Unit description rows (e.g., "$ thousands"): use style=0 or style=2 (blue not needed)

---

## 8. Prohibited Actions (What You Must NOT Do)

- **Do not modify existing `<xf>` entries**: Batch-changes style of all cells referencing that index
- **Do not delete fills[0] and fills[1]**: Required by OOXML specification; deletion causes file corruption
- **Do not modify cell values or formulas**: FORMAT path only changes styles, not content
- **Do not use openpyxl for formatting**: openpyxl rewrites entire styles.xml on save, losing unsupported features
- **Do not apply global override styles**: Do not cover entire workbook with single style; assign precisely by semantic role
- **Do not write FF in Alpha channel**: `rgb="FF0000FF"` makes color fully transparent; correct format is `rgb="000000FF"`

---

## 9. Common Errors and Fixes

### Error 1: Year displays as 2,024

Cause: Year cell's `s` attribute uses format with thousands separator (e.g., numFmtId="3" or numFmtId="167").

```xml
<!-- Incorrect -->
<c r="B1" s="9"><v>2024</v></c>

<!-- Fix: Change to s="11" (numFmtId="1", format 0) -->
<c r="B1" s="11"><v>2024</v></c>
```

### Error 2: Percentage displays as 800% (value multiplied by 100)

Cause: 8% stored as `<v>8</v>` instead of `<v>0.08</v>`. Excel's `%` format automatically multiplies value by 100 for display.

```xml
<!-- Incorrect -->
<c r="B2" s="7"><v>8</v></c>

<!-- Fix: Value must be stored in decimal form -->
<c r="B2" s="7"><v>0.08</v></c>
```

### Error 3: File corruption after appending styles without updating count

Cause: `<font>` or `<xf>` element appended but count attribute not updated; Excel reads beyond bounds using old count.

Fix: Update corresponding count immediately after appending each element:
```xml
<!-- After appending 6th font, count must be changed from 5 to 6 -->
<fonts count="6">
  ...
</fonts>
```

### Error 4: Blue font + formula (color role contradiction)

Cause: Formula cell mistakenly uses input style (e.g., s="5" for currency input).

```xml
<!-- Incorrect: Formula cell uses blue input style -->
<c r="C5" s="5"><f>B5*1.08</f><v></v></c>

<!-- Fix: Change formula cell to corresponding black formula style (5->6, 7->8, 9->10) -->
<c r="C5" s="6"><f>B5*1.08</f><v></v></c>
```

### Error 5: AARRGGBB color missing Alpha (only 6 digits)

```xml
<!-- Incorrect: 6-digit format, behavior depends on implementation, usually causes wrong color -->
<color rgb="0000FF"/>

<!-- Fix: Always use 8-digit AARRGGBB, Alpha fixed at 00 -->
<color rgb="000000FF"/>
```

### Error 6: Modifying existing xf (affects all cells referencing that index)

Cause: Directly modifying attributes of Nth `<xf>` in cellXfs, causing all cells with `s="N"` batch-changed.

Fix: Keep existing entries unchanged, append new entry at end, only change `s` attribute of cells needing new style to new index:
```xml
<!-- Incorrect: Modified existing xf at index=6 -->
<xf numFmtId="164" fontId="2" fillId="0" borderId="0" xfId="0"
    applyFont="1" applyNumberFormat="1" applyAlignment="1">
  <alignment horizontal="right"/>  <!-- New attribute added, affects ALL cells already using s="6" -->
</xf>

<!-- Fix: Append new index (when original count=13, new index=13), only change s attribute of cells needing right alignment -->
<!-- Keep index=6 as-is -->
<xf numFmtId="164" fontId="2" fillId="0" borderId="0" xfId="0"
    applyFont="1" applyNumberFormat="1" applyAlignment="1">
  <alignment horizontal="right"/>
</xf>  <!-- New index=13 -->
```

---

## 10. Financial Model Structure Conventions

### 10.1 Header Rows

- Bold font (corresponds to style index 4 in this skill's template)
- Year columns: use number format `0` (numFmtId="1", no thousands separator) to prevent 2024 displaying as 2,024
- Unit description row may be added below headers: gray or italic text, e.g., "$ thousands" or "% of Revenue"

### 10.2 Row Type Standards

| Row Type | Style Recommendation | Example |
|----------|---------------------|---------|
| Category heading row | Bold, optionally with fill color | "Revenue" |
| Line item row | Normal style | "Product A", "Product B" |
| Subtotal row | Bold + top border | "Total Revenue" |
| Operating metric row | Normal style | "Gross Margin %" |
| Separator row | Empty row | (empty) |

### 10.3 Multi-Year Model Column Layout

```
Col A: Label column          (width 28, left-aligned text, s="4" for headers or s="0" for labels)
Col B: FY2022 Actual         (width 12, year header s="11", data cells styled by semantic role)
Col C: FY2023 Actual
Col D: FY2024E               (forecast period - can use light gray fill fillId=3 to differentiate)
Col E: FY2025E
Col F: FY2026E
```

### 10.4 Cross-Sheet Reference Patterns

Complete XML example of parameters passing from assumptions sheet to model sheet:

```xml
<!-- Assumptions sheet, cell B5: 8% growth rate, blue percentage input -->
<c r="B5" s="7"><v>0.08</v></c>

<!-- Model sheet, cell C10: references assumption area growth rate, green percentage formula -->
<!-- Requires appending index=13: green + percentage format (fontId=3, numFmtId=165) -->
<c r="C10" s="13"><f>Assumptions!B5</f><v></v></c>
```

---

## 11. Assumption Categories

In assumptions area (Assumptions sheet or assumptions block), organize assumptions in following standard order for ease of review and maintenance:

1. **Revenue assumptions**: Growth rates, pricing, sales volume
2. **Cost assumptions**: Gross margin, fixed/variable cost ratios
3. **Working capital**: DSO (Days Sales Outstanding), DPO (Days Payable Outstanding), inventory days
4. **Capital expenditures (CapEx)**: As percentage of revenue or absolute amounts
5. **Financing assumptions**: Interest rates, debt repayment schedules
6. **Tax and other**: Effective tax rate, depreciation & amortization (D&A)

---

## 12. Audit Trail Best Practices

- Use `s="12"` (blue font + yellow fill highlight) to mark cells requiring review or pending changes, making them immediately visible to reviewers
- In sensitivity analysis rows or separate Sensitivity tab, show impact of +/-1% changes in key assumptions on results
- **Do not hide rows containing assumptions**: Assumption rows must be visible to reviewers; do not use `hidden="1"` attribute
- Note "Last Updated" date at top of assumptions area or in dedicated cell, recording last modification time of model

---

## 13. Pre-Delivery Checklist (Common Financial Model Checklist)

Before outputting final file, confirm each item:

- [ ] Formula rows contain no hard-coded values (can use `formula_check.py` to scan packaged `.xlsx` file)
- [ ] Year columns display as 2024 not 2,024 (numFmtId="1", format `0`)
- [ ] Negative numbers display as (1,234) not -1,234 (use parenthetical style for externally delivered financial reports)
- [ ] Zero values display as `-` in sparse rows rather than `0` (formatCode third segment is `"-"`)
- [ ] Growth rates and percentages stored as decimals (0.08 = 8%), format is `0.0%`
- [ ] All cross-sheet reference cells use green font (style index 3 or appended green + number format combination)
- [ ] Assumptions block and model block clearly separated (different sheets or separated by empty rows within same sheet)
- [ ] Summary rows use `SUM()` formulas, not manually hard-coded totals
- [ ] Balance verification: summary rows = sum of respective line items (check row can be added at end of model to verify)

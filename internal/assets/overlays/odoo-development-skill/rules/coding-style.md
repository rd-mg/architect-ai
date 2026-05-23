# Coding Style

**Apply when writing or editing code in any project.**

> [!IMPORTANT]
> **ENGLISH LANGUAGE ONLY**: All code, comments, manifest fields, variable names, classes, DB schemas, and dev communication MUST be in ENGLISH. Spanish or other translations/comments strictly prohibited.

---

## Rule 1: Dependencies at the Top

All import, include, or require statements at top of every file.

**Placement order:**

1. After file header comment or docstring (if present)
2. Before any code, classes, or functions
3. Group by type: standard library → third-party → local imports
4. Separate groups with blank lines

**Never import inside functions, methods, loops, or conditionals.**

---

## Rule 2: Meaningful Variable Names

Every variable, function parameter, and constant MUST have descriptive name revealing purpose.

**Good names:**

- Reveal intent: what value represents
- Use full words, not abbreviations
- Include units when applicable (e.g., `timeout_seconds`, `price_usd`)
- Use domain terminology

**Forbidden names:**

- Generic: `data`, `info`, `temp`, `tmp`, `var`, `obj`, `item`, `thing`, `stuff`
- Single letters: `x`, `y`, `z`, `a`, `b`, `c` (except loop indices or math)
- Type names: `list`, `array`, `dict`, `map`, `string`
- Abbreviations: `vals`, `recs`, `acc`, `cnt`, `amt`, `num`
- Non-descriptive: `result`, `output`, `value`, `list1`, `data2`

**Acceptable short names:**

- Loop indices: `i`, `j`, `k` in simple loops
- Math/geometry: `x`, `y`, `z` for coordinates
- Exception variables: `e` or `err` in catch blocks

---

## Enforcement

**When writing new code:**

- Place all imports at top before any code
- Name every variable descriptively
- If name needs comment to explain it, name is wrong

**When editing existing code:**

- Fix misplaced imports
- Rename unclear variables while editing
- Do not propagate bad patterns

**Stop and refactor immediately when you see:**

- Import statements after first function or class
- Import statements inside functions
- Variables named with forbidden names
- Abbreviations requiring mental translation

---

## Rationale

Code read far more often than written. Proper organization and meaningful names make code self-documenting, reduce cognitive load, enable faster maintenance.

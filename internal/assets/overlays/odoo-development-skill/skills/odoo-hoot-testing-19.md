# Skill: Odoo Hoot Testing Framework (v19)

## 1. Description
The `odoo-hoot-testing-19` skill defines the architectural mandates for the new reactive frontend testing framework, replacing the legacy QUnit runner.

## 2. Core Mandatory Rules
- **Framework**: Prohibited to use `QUnit` or `web.test_utils`. Use `@odoo/hoot`.
- **Async Pattern**: All DOM interactions (`click`, `edit`) MUST be `await`-ed.
- **Reactivity Flush**: Mandatory `await animationFrame()` after state changes.

## 3. Implementation Patterns (Bad vs Good)

### A. Test Structure
```javascript
// ❌ LEGACY: QUnit approach (deprecated)
QUnit.test("my test", async (assert) => { ... });

// ✅ v19 STANDARD: Hoot structure
import { describe, expect, test } from "@odoo/hoot";

test("My Component Test", async () => {
    // Test logic here
});
```

### B. Reactive Reactivity Flush
```javascript
// ❌ LEGACY: Assuming sync DOM updates
await click(".o_save");
expect(".o_form").toHaveText("Saved");

// ✅ v19 STANDARD: Explicitly wait for OWL reactivity
await click(".o_save");
await animationFrame(); // Flush reactive updates
expect(".o_form").toHaveText("Saved");
```

## 4. Verification Workflow
- Ensure all tests are tagged with environment tags (`describe.current.tags("desktop")`).
- Validate that mocking uses `hoot-mock` modules (e.g., `mockService`, `onRpc`, `mockDate`).

## 5. Maintenance
- Track breaking changes in `@odoo/hoot` API.
- Migrate remaining legacy tours as Hoot becomes mandatory in v19.1+.

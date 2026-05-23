# Odoo OWL Migration: 16.0 → 17.0 (OWL 2.x Continued)

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  OWL MIGRATION GUIDE: 16.0 → 17.0                                            ║
║  OWL 2.x continues with enhancements and best practice refinements           ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Overview

v17 continues OWL 2.x with refinements. **No breaking changes.** v16 components work in v17.

## Changes Summary

| Feature | v16 | v17 |
|---------|-----|-----|
| OWL Version | 2.x | 2.x (same) |
| Props validation | Recommended | Strongly recommended |
| Services | Standard | Enhanced patterns |
| Error handling | Basic | Improved |
| JSDoc | Optional | Recommended |

## Enhanced Patterns for v17

### Props Validation
```javascript
// v17: Enhanced props with validation
export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    static props = {
        recordId: { type: Number, optional: true },
        mode: { type: String, optional: true,
            validate: (v) => ["view","edit","create"].includes(v) },
        onConfirm: { type: Function, optional: true },
    };
    static defaultProps = { mode: "view", config: {} };
}
```

### JSDoc Type Annotations
```javascript
/** @odoo-module **/
import { Component } from "@odoo/owl";

/**
 * @typedef {Object} MyComponentProps
 * @property {number} [recordId]
 * @property {'view'|'edit'|'create'} [mode]
 */

export class MyComponent extends Component {
    /** @type {string} */
    static template = "my_module.MyComponent";
    /** @type {MyComponentProps} */
    static props = { ... };
    static defaultProps = { mode: "view" };
}
```

### Enhanced Service Usage
```javascript
setup() {
    this.orm = useService("orm");
    this.action = useService("action");
    this.notification = useService("notification");
    this.dialog = useService("dialog");
    this.user = useService("user");
    this.company = useService("company");  // Enhanced in v17
}
```

### Error Handling
```javascript
async loadDataWithRetry(retries = 3) {
    for (let i = 0; i < retries; i++) {
        try {
            this.state.data = await this.orm.searchRead("my.model", [], ["name"]);
            return;
        } catch (error) {
            if (i === retries - 1) {
                this.notification.add(`Failed: ${error.message}`, { type: "danger" });
            }
            await new Promise(r => setTimeout(r, 1000 * (i + 1)));
        }
    }
}
```

### Lifecycle Best Practices
```javascript
setup() {
    this.orm = useService("orm");
    this.state = useState({ data: null, loading: true });
    this._cleanup = [];

    onWillStart(async () => { await this.loadData(); });
    onMounted(() => { this.setupEventListeners(); });
    onWillUpdateProps(async (nextProps) => { if (nextProps.id !== this.props.id) await this.loadData(); });
    onWillUnmount(() => { this._cleanup.forEach(fn => fn()); });
}
```

## Template Patterns for v17

```xml
<t t-name="my_module.MyComponent">
    <div class="my-component">
        <t t-if="state.loading"><div class="spinner-border"/></t>
        <t t-elif="state.error">
            <div class="alert alert-danger" t-esc="state.error"/>
        </t>
        <t t-else="">
            <t t-foreach="state.data" t-as="item" t-key="item.id">
                <div t-on-click="() => this.onItemClick(item)">
                    <span t-esc="item.name"/>
                </div>
            </t>
        </t>
    </div>
</t>
```

## Checklist (v16 → v17)

No breaking changes. Focus on improvements:
- [ ] Add JSDoc type annotations
- [ ] Add props validation with validate functions
- [ ] Add static defaultProps
- [ ] Implement cleanup in onWillUnmount
- [ ] Use enhanced service patterns
- [ ] Add error handling with retry logic
- [ ] Update templates with Bootstrap 5

## Best Practices

1. Always validate props — static props with types + validators
2. Document with JSDoc — type annotations for IDE support
3. Handle errors gracefully — user-friendly messages
4. Clean up resources — remove listeners in onWillUnmount
5. Use services correctly — company, user, and other services
6. Follow lifecycle hooks — appropriate hooks for each task

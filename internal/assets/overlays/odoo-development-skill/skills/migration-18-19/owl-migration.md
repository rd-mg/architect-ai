# Odoo OWL Migration: 18.0 → 19.0 (OWL 2.x → 3.x)

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  OWL MIGRATION GUIDE: 2.x → 3.x                                              ║
║  Significant update with enhanced patterns.                                  ║
║  Note: v19 in development — patterns may change.                             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Changes Summary

| Feature | OWL 2.x (v18) | OWL 3.x (v19) |
|---------|---------------|---------------|
| Reactivity | Standard | Enhanced |
| Props validation | Runtime | Enhanced + TS-like |
| Type annotations | JSDoc optional | JSDoc recommended |
| Error boundaries | Basic | Enhanced |
| Performance | Good | Improved |

## Enhanced Reactivity

**OWL 2.x:** Set re-creation needed for reactivity
**OWL 3.x:** Direct mutation works (Set/Map fully reactive)

```javascript
// v19: Direct mutation works
setup() {
    this.state = useState({ items: [], selectedIds: new Set() });
}
toggleSelect(id) {
    if (this.state.selectedIds.has(id))
        this.state.selectedIds.delete(id);
    else
        this.state.selectedIds.add(id);
    // No need to recreate Set
}
```

## Enhanced Props Validation

```javascript
static props = {
    recordId: { type: Number, optional: true },
    onConfirm: { type: Function, optional: true },
    mode: {
        type: String, optional: true,
        validate: (v) => ['view','edit'].includes(v),
    },
};
static defaultProps = { mode: 'view' };
```

## Complete Migration

**OWL 2.x (v18):**
```javascript
import { Component, useState, onWillStart } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";

export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    static props = { recordId: { type: Number, optional: true } };
    setup() {
        this.orm = useService("orm");
        this.state = useState({ data: [], loading: true });
        onWillStart(async () => { await this.loadData(); });
    }
}
registry.category("actions").add("my_module.my_action", MyComponent);
```

**OWL 3.x (v19):**
```javascript
import { Component, useState, onWillStart, onWillUnmount } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";
import { registry } from "@web/core/registry";

/**
 * @typedef {Object} MyComponentProps
 * @property {number} [recordId]
 * @property {(data: Array) => void} [onDataLoad]
 */

export class MyComponent extends Component {
    /** @type {string} */
    static template = "my_module.MyComponent";
    /** @type {MyComponentProps} */
    static props = {
        recordId: { type: Number, optional: true },
        onDataLoad: { type: Function, optional: true },
    };

    setup() {
        this.orm = useService("orm");
        this.notification = useService("notification");
        this.state = useState({ data: [], loading: true, error: null });
        this._abortController = null;

        onWillStart(async () => { await this.loadData(); });
        onWillUnmount(() => { this._abortController?.abort(); });
    }

    async loadData() {
        this._abortController = new AbortController();
        try {
            this.state.data = await this.orm.searchRead("my.model", [], ["name","state"]);
            this.state.error = null;
            this.props.onDataLoad?.(this.state.data);
        } catch (error) {
            if (error.name !== 'AbortError') {
                this.state.error = error.message;
                this.notification.add("Failed", { type: "danger" });
            }
        } finally { this.state.loading = false; }
    }
}
registry.category("actions").add("my_module.my_action", MyComponent);
```

## Key Differences

1. **Enhanced type annotations** — JSDoc for props/state/methods
2. **Better error handling** — AbortController for cancellation, cleanup
3. **Callback props** — `this.props.onDataLoad?.(data)`
4. **State type definitions** — `@typedef` for state

## Checklist

- [ ] Add comprehensive JSDoc annotations
- [ ] Define `@typedef` for component props
- [ ] Define `@typedef` for component state
- [ ] Add return/param types to all methods
- [ ] Implement cleanup in `onWillUnmount`
- [ ] Use AbortController for cancellable requests
- [ ] Update Set/Map usage (now fully reactive)
- [ ] Add validation functions to props
- [ ] Use `static defaultProps`

## Common Issues

**Set/Map not updating UI** → v18 needed recreate, v19 direct mutations work
**Missing type annotations** → Add JSDoc for all methods, state, props
**Uncancelled requests on unmount** → Use AbortController + onWillUnmount cleanup

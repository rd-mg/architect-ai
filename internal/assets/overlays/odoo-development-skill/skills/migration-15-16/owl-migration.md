# Odoo OWL Migration: 15.0 → 16.0 (OWL 1.x → 2.x)

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  OWL MIGRATION GUIDE: 1.x → 2.x                                              ║
║  MAJOR breaking change - all OWL components must be rewritten.               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Breaking Changes Summary

| Feature | OWL 1.x (v15) | OWL 2.x (v16) |
|---------|---------------|---------------|
| Module | `odoo.define()` | ES modules `/** @odoo-module **/` |
| Imports | `require()` | `import` statements |
| Hooks | `owl.hooks` | Direct from `@odoo/owl` |
| Template | Property on class | `static template` |
| RPC | `require('web.rpc')` | `useService("orm")` |
| Registration | `core.action_registry` | `registry.category("actions")` |

## Complete Migration

### Before (OWL 1.x - v15)
```javascript
odoo.define('my_module.MyComponent', function (require) {
    const { Component, useState } = owl;
    const { onWillStart } = owl.hooks;
    const rpc = require('web.rpc');
    const core = require('web.core');

    class MyComponent extends Component {
        setup() {
            this.state = useState({ records: [], loading: true });
            onWillStart(async () => { await this.loadRecords(); });
        }
        async loadRecords() {
            this.state.records = await rpc.query({model:'my.model', method:'search_read'});
        }
    }
    MyComponent.template = 'my_module.MyComponent';
    core.action_registry.add('my_module.my_action', MyComponent);
    return MyComponent;
});
```

### After (OWL 2.x - v16)
```javascript
/** @odoo-module **/
import { Component, useState, onWillStart } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";
import { registry } from "@web/core/registry";

export class MyComponent extends Component {
    static template = "my_module.MyComponent";
    setup() {
        this.orm = useService("orm");
        this.state = useState({ records: [], loading: true });
        onWillStart(async () => { await this.loadRecords(); });
    }
    async loadRecords() {
        this.state.records = await this.orm.searchRead("my.model", [], ["name","state"]);
    }
}
registry.category("actions").add("my_module.my_action", MyComponent);
```

## Template Migration

**OWL 1.x:** `<t t-name="my_module.MyComponent" owl="1">`
**OWL 2.x:** `<t t-name="my_module.MyComponent">` (remove `owl="1"`)

## Import Mapping

| OWL 1.x | OWL 2.x |
|---------|---------|
| `const { Component } = owl;` | `import { Component } from "@odoo/owl";` |
| `const { useState } = owl;` | `import { useState } from "@odoo/owl";` |
| `const { onWillStart } = owl.hooks;` | `import { onWillStart } from "@odoo/owl";` |
| `const rpc = require('web.rpc');` | `this.orm = useService("orm");` |
| `const core = require('web.core');` | `import { registry } from "@web/core/registry";` |

## RPC Method Migration

| OWL 1.x | OWL 2.x |
|---------|---------|
| `rpc.query({model, method:'search_read'})` | `this.orm.searchRead(model, domain, fields)` |
| `rpc.query({model, method:'read'})` | `this.orm.read(model, ids, fields)` |
| `rpc.query({model, method:'create'})` | `this.orm.create(model, vals)` |
| `rpc.query({model, method:'write'})` | `this.orm.write(model, ids, vals)` |
| `rpc.query({model, method:'unlink'})` | `this.orm.unlink(model, ids)` |

## Registration Migration

| OWL 1.x | OWL 2.x |
|---------|---------|
| `core.action_registry.add(key, C)` | `registry.category("actions").add(key, C)` |

## Checklist

- [ ] Replace `odoo.define()` with `/** @odoo-module **/`
- [ ] Convert `require()` → `import` statements
- [ ] Import hooks from `@odoo/owl` (not `owl.hooks`)
- [ ] Add `static template` to class
- [ ] Replace `rpc.query()` with `useService("orm")`
- [ ] Replace `core.action_registry` with `registry.category("actions")`
- [ ] Remove `owl="1"` from templates
- [ ] Update manifest assets glob patterns
- [ ] Test all component functionality

## Common Errors

**`odoo.define is not defined`** → Use ES module syntax
**`owl is not defined`** → Import from `@odoo/owl`
**`rpc is not defined`** → Use `useService("orm")`

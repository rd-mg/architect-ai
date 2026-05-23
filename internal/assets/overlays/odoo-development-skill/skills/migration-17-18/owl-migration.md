# Odoo OWL Migration: 17.0 → 18.0 (OWL 2.x Enhanced)

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  OWL MIGRATION GUIDE: 17.0 → 18.0                                            ║
║  Minor updates - OWL 2.x remains same with enhancements.                     ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Overview

OWL 2.x same as v17 with minor enhancements. **No breaking changes.**

## Changes

| Feature | v17 | v18 |
|---------|-----|-----|
| OWL Version | 2.x | 2.x (enhanced) |
| Services | Standard | Additional services |
| Props validation | Recommended | Recommended |
| JSDoc | Optional | Recommended |

## New/Enhanced Services

### Company Service
```javascript
setup() {
    this.company = useService("company");
    const current = this.company.currentCompany;
    const allowed = this.company.allowedCompanies;
}
```

### Enhanced ORM
```javascript
const records = await this.orm.searchRead("my.model", domain, fields, {
    limit: 100, offset: 0, order: "create_date DESC",
    context: { ...this.user.context },
});
```

### Enhanced Notification
```javascript
this.notification.add("Success", {
    type: "success", sticky: false,
    buttons: [{ name: "Undo", onClick: () => this.undo() }],
});
```

## Best Practices

### JSDoc Type Annotations
```javascript
/** @odoo-module **/
import { Component } from "@odoo/owl";
/** @typedef {Object} Props @property {number} [recordId] */

export class MyComponent extends Component {
    /** @type {string} */
    static template = "my_module.MyComponent";
    /** @type {Props} */
    static props = { recordId: { type: Number, optional: true } };
}
```

### Static Props Validation
```javascript
static props = {
    recordId: { type: Number },
    mode: { type: String, optional: true },
    onConfirm: { type: Function, optional: true },
    items: { type: Array, optional: true },
    value: { type: [String, Number], optional: true },
};
static defaultProps = { mode: "view", items: [] };
```

### Cleanup
```javascript
setup() {
    this._cleanup = null;
    onMounted(() => {
        const handler = (e) => this.handleKeydown(e);
        document.addEventListener("keydown", handler);
        this._cleanup = () => document.removeEventListener("keydown", handler);
    });
    onWillUnmount(() => { this._cleanup?.(); });
}
```

## Checklist

- [ ] Add JSDoc type annotations
- [ ] Add comprehensive `static props` validation
- [ ] Add `static defaultProps` where needed
- [ ] Use new company service if needed
- [ ] Add cleanup logic in `onWillUnmount`
- [ ] Test with new view Python expressions

## No Breaking Changes

v17 OWL components work in v18 without modification. Changes are additive.

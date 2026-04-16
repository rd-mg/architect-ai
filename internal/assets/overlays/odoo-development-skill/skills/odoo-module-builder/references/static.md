> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Static Assets Reference

## Table of Contents
- [Directory Structure](#directory-structure)
- [Module Icon](#module-icon)
- [Web Assets (JS, CSS, XML)](#web-assets-js-css-xml)
- [OWL Components (Odoo 17)](#owl-components-odoo-17)
- [Widget Registration](#widget-registration)

## Directory Structure

```
static/
├── description/
│   ├── icon.png            # Module icon (required for app modules)
│   └── index.html          # Module description page (optional)
├── src/
│   ├── css/
│   │   └── module_styles.css
│   ├── js/
│   │   └── my_widget.js
│   ├── xml/
│   │   └── my_widget.xml   # OWL/QWeb templates for JS
│   └── img/
│       └── logo.png
└── lib/                     # Third-party libraries (rare)
```

## Module Icon

Place a 128x128 PNG at `static/description/icon.png`. This is displayed in the Apps menu.

Reference in menu:
```xml
<menuitem id="menu_root" name="My App"
          web_icon="module_name,static/description/icon.png"/>
```

## Web Assets (JS, CSS, XML)

Register in `__manifest__.py`:

```python
'assets': {
    'web.assets_backend': [
        # Backend (internal) assets
        'module_name/static/src/css/my_styles.css',
        'module_name/static/src/js/my_widget.js',
        'module_name/static/src/xml/my_widget.xml',
    ],
    'web.assets_frontend': [
        # Website/portal assets
        'module_name/static/src/css/website_styles.css',
        'module_name/static/src/js/website_script.js',
    ],
},
```

### Asset Bundles

| Bundle | Used For |
|--------|----------|
| `web.assets_backend` | Internal/backend UI |
| `web.assets_frontend` | Website and portal pages |
| `web.assets_common` | Shared between frontend and backend |
| `web.report_assets_common` | PDF reports |

## OWL Components (Odoo 17)

Odoo 17 uses OWL 2 for frontend components:

### JavaScript Component

```javascript
/** @odoo-module **/

import { registry } from "@web/core/registry";
import { Component, useState } from "@odoo/owl";

export class BookCounter extends Component {
    static template = "library_management.BookCounter";
    static props = {
        record: Object,
    };

    setup() {
        this.state = useState({ count: 0 });
    }

    increment() {
        this.state.count++;
    }
}

// Register as a view widget
registry.category("view_widgets").add("book_counter", {
    component: BookCounter,
});
```

### OWL Template (XML)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<templates xml:space="preserve">
    <t t-name="library_management.BookCounter">
        <div class="book-counter">
            <span t-esc="state.count"/>
            <button t-on-click="increment" class="btn btn-sm btn-primary">
                +1
            </button>
        </div>
    </t>
</templates>
```

### Usage in Form View

```xml
<widget name="book_counter"/>
```

## Widget Registration

### Custom Field Widget (Odoo 17)

```javascript
/** @odoo-module **/

import { registry } from "@web/core/registry";
import { standardFieldProps } from "@web/views/fields/standard_field_props";
import { Component } from "@odoo/owl";

export class StarRatingField extends Component {
    static template = "library_management.StarRatingField";
    static props = { ...standardFieldProps };

    get stars() {
        return Array.from({ length: 5 }, (_, i) => i < this.props.record.data[this.props.name]);
    }

    onStarClick(rating) {
        this.props.record.update({ [this.props.name]: rating });
    }
}

registry.category("fields").add("star_rating", {
    component: StarRatingField,
    supportedTypes: ["integer"],
});
```

Usage:
```xml
<field name="rating" widget="star_rating"/>
```

### Odoo 16 Widget Style (legacy)

Odoo 16 uses a mix of legacy widgets (`Widget.extend()`) and early OWL. Prefer OWL when possible, but legacy pattern:

```javascript
odoo.define('module_name.MyWidget', function (require) {
    "use strict";
    const AbstractField = require('web.AbstractField');
    const fieldRegistry = require('web.field_registry');

    const MyWidget = AbstractField.extend({
        template: 'module_name.MyWidget',
        // ...
    });

    fieldRegistry.add('my_widget', MyWidget);
    return MyWidget;
});
```

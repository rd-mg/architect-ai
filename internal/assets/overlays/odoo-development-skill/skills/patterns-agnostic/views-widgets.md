# Views & Widgets Patterns

Consolidated from the following source files:
- `form-list-kanban-patterns.md` (architect-fix)
- `xpath-inheritance-patterns.md` (architect-fix)
- `search-filter-patterns.md` (architect-fix)
- `action-menu-patterns.md` (architect-fix)
- `qweb-report-view-patterns.md` (architect-fix)
- `owl-widget-patterns.md` (architect-fix)

> **Version-specific syntax** → `patterns-{version}/owl-components.md`
> `attrs=` removed in v17+ · `<tree>` renamed `<list>` in v18+ · OWL 1→2 in v15, OWL 2→3 in v17+

---

## Core XML Views

### Form View (Smart Buttons & Chatter)
```xml
<form>
    <header>
        <button name="action_confirm" type="object" class="btn-primary" invisible="state != 'draft'"/>
        <field name="state" widget="statusbar" statusbar_visible="draft,done"/>
    </header>
    <sheet>
        <div class="oe_button_box" name="button_box">
            <button name="action_view_related" type="object" class="oe_stat_button" icon="fa-list">
                <field name="related_count" widget="statinfo"/>
            </button>
        </div>
        <group>
            <field name="partner_id"/>
            <field name="date"/>
        </group>
        <notebook>
            <page string="Lines">
                <field name="line_ids"><list editable="bottom">...</list></field>
            </page>
        </notebook>
    </sheet>
    <div class="oe_chatter">
        <field name="message_follower_ids"/><field name="activity_ids"/><field name="message_ids"/>
    </div>
</form>
```

### Kanban Card Template
```xml
<kanban default_group_by="state">
    <templates>
        <t t-name="kanban-card">
            <div class="oe_kanban_global_click">
                <div class="o_kanban_record_top">
                    <strong class="o_kanban_record_title"><field name="name"/></strong>
                </div>
                <div class="o_kanban_record_bottom">
                    <div class="oe_kanban_bottom_left"><field name="amount_total" widget="monetary"/></div>
                    <div class="oe_kanban_bottom_right"><field name="activity_ids" widget="kanban_activity"/></div>
                </div>
            </div>
        </t>
    </templates>
</kanban>
```

---

## XPath & Inheritance

### Reliable Selectors
```xml
<!-- ✅ GOOD: use hasclass() — robust against class order changes -->
<xpath expr="//div[hasclass('oe_button_box')]" position="inside">
    <button .../>
</xpath>

<!-- ✅ GOOD: use named attributes -->
<xpath expr="//field[@name='partner_id']" position="after">
    <field name="custom_field"/>
</xpath>

<!-- ❌ BAD: position-based xpath breaks when siblings are added -->
<xpath expr="//group[2]/field[1]" position="replace">
```

---

## Actions & OWL Widgets

### Window Action & Menu
```xml
<record id="action_my_model" model="ir.actions.act_window">
    <field name="res_model">my.model</field>
    <field name="view_mode">list,form,kanban</field>
</record>

<menuitem id="menu_root" name="My App" sequence="10"/>
<menuitem id="menu_sub" parent="menu_root" action="action_my_model"/>
```

### OWL Widget Skeleton (v17+)
```javascript
/** @odoo-module **/
import { registry } from "@web/core/registry";
import { Component, useState } from "@odoo/owl";

export class MyWidget extends Component {
    static template = "my_module.MyWidget";
    setup() { this.state = useState({ value: this.props.value }); }
}
registry.category("fields").add("my_widget", { component: MyWidget });
```

---

## Anti-Patterns

```xml
<!-- ❌ NEVER use contains(@class) in XPath — use hasclass() instead. -->

<!-- ❌ NEVER use attrs= in v17+ — Use invisible="state == 'done'" directly. -->

<!-- ❌ NEVER use <tree> in v18+ — Use <list> instead. -->
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| `attrs=` | ✅ | ❌ removed | ❌ | ❌ |
| `<tree>` | ✅ | ✅ | `<list>` | `<list>` |
| `invisible=` | partial | ✅ full | ✅ | ✅ |
| OWL | 1.x / 2.x | 2.x | 2.x | 3.x |

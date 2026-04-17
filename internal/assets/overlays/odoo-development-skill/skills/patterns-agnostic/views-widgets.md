# Views Widgets Patterns

Consolidated from the following source files:
- `xml-view-patterns.md`
- `widget-field-patterns.md`
- `qweb-template-patterns.md`
- `action-patterns.md`
- `menu-navigation-patterns.md`
- `dashboard-kpi-patterns.md`

---


## Source: xml-view-patterns.md

# XML View Patterns Reference

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  XML VIEW PATTERNS                                                           ║
║  Complete reference for Odoo view definitions with version-specific syntax   ║
║  Critical: visibility syntax differs between versions                        ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## View Types Overview

| View Type | Purpose | Element |
|-----------|---------|---------|
| Form | Single record editing | `<form>` |
| Tree/List | Multiple records display | `<tree>` |
| Kanban | Card-based view | `<kanban>` |
| Search | Filtering/grouping | `<search>` |
| Graph | Charts/analytics | `<graph>` |
| Pivot | Pivot tables | `<pivot>` |
| Calendar | Date-based display | `<calendar>` |
| Gantt | Timeline view | `<gantt>` |

---

## Form View

### Basic Structure
```xml
<record id="my_model_view_form" model="ir.ui.view">
    <field name="name">my.model.form</field>
    <field name="model">my.model</field>
    <field name="arch" type="xml">
        <form string="My Model">
            <header>
                <!-- Status bar and buttons -->
            </header>
            <sheet>
                <!-- Main content -->
            </sheet>
            <div class="oe_chatter">
                <!-- Mail integration -->
            </div>
        </form>
    </field>
</record>
```

### Complete Form Example (v18)
```xml
<record id="my_model_view_form" model="ir.ui.view">
    <field name="name">my.model.form</field>
    <field name="model">my.model</field>
    <field name="arch" type="xml">
        <form string="My Model">
            <header>
                <button name="action_confirm"
                        type="object"
                        string="Confirm"
                        class="btn-primary"
                        invisible="state != 'draft'"/>
                <button name="action_cancel"
                        type="object"
                        string="Cancel"
                        invisible="state not in ('draft', 'confirmed')"/>
                <field name="state" widget="statusbar"
                       statusbar_visible="draft,confirmed,done"/>
            </header>
            <sheet>
                <div class="oe_button_box" name="button_box">
                    <button name="action_view_invoices"
                            type="object"
                            class="oe_stat_button"
                            icon="fa-pencil-square-o">
                        <field name="invoice_count" widget="statinfo"
                               string="Invoices"/>
                    </button>
                </div>
                <widget name="web_ribbon" title="Archived"
                        bg_color="bg-danger"
                        invisible="active"/>
                <div class="oe_title">
                    <h1>
                        <field name="name" placeholder="Name"/>
                    </h1>
                </div>
                <group>
                    <group string="General">
                        <field name="partner_id"/>
                        <field name="date"/>
                        <field name="user_id"/>
                    </group>
                    <group string="Details">
                        <field name="company_id" groups="base.group_multi_company"/>
                        <field name="currency_id" invisible="1"/>
                        <field name="amount"/>
                    </group>
                </group>
                <notebook>
                    <page string="Lines" name="lines">
                        <field name="line_ids">
                            <tree editable="bottom">
                                <field name="sequence" widget="handle"/>
                                <field name="name"/>
                                <field name="quantity"/>
                                <field name="price_unit"/>
                                <field name="subtotal"/>
                            </tree>
                        </field>
                    </page>
                    <page string="Notes" name="notes">
                        <field name="notes" placeholder="Internal notes..."/>
                    </page>
                </notebook>
            </sheet>
            <div class="oe_chatter">
                <field name="message_follower_ids"/>
                <field name="activity_ids"/>
                <field name="message_ids"/>
            </div>
        </form>
    </field>
</record>
```

---

## Visibility Syntax by Version

### v14-v16: attrs Syntax
```xml
<!-- DEPRECATED in v16, REMOVED in v17 -->
<field name="partner_id"
       attrs="{'invisible': [('state', '=', 'draft')],
               'readonly': [('state', '!=', 'draft')],
               'required': [('type', '=', 'customer')]}"/>

<button name="action"
        attrs="{'invisible': [('state', '!=', 'draft')]}"/>

<group attrs="{'invisible': [('show_details', '=', False)]}">
    <field name="detail"/>
</group>
```

### v17+: Inline Expression Syntax
```xml
<!-- REQUIRED in v17+ -->
<field name="partner_id"
       invisible="state == 'draft'"
       readonly="state != 'draft'"
       required="type == 'customer'"/>

<button name="action"
        invisible="state != 'draft'"/>

<group invisible="not show_details">
    <field name="detail"/>
</group>
```

### Expression Conversion Table

| attrs Domain | v17+ Expression |
|--------------|-----------------|
| `[('field', '=', 'value')]` | `field == 'value'` |
| `[('field', '!=', 'value')]` | `field != 'value'` |
| `[('field', '=', True)]` | `field` |
| `[('field', '=', False)]` | `not field` |
| `[('field', 'in', ['a','b'])]` | `field in ('a', 'b')` |
| `[('field', '>', 0)]` | `field > 0` |
| `['&', A, B]` | `A and B` |
| `['|', A, B]` | `A or B` |

### Complex Expressions (v17+)
```xml
<!-- AND condition -->
<field name="x" invisible="state == 'draft' and not is_manager"/>

<!-- OR condition -->
<field name="x" invisible="state == 'done' or state == 'cancel'"/>

<!-- Nested -->
<field name="x" invisible="state == 'draft' or (type == 'service' and qty == 0)"/>

<!-- Parent access in One2many -->
<field name="x" invisible="parent.state != 'draft'"/>

<!-- Context access -->
<field name="x" invisible="context.get('hide_field')"/>
```

---

## Tree/List View

### Basic Tree
```xml
<record id="my_model_view_tree" model="ir.ui.view">
    <field name="name">my.model.tree</field>
    <field name="model">my.model</field>
    <field name="arch" type="xml">
        <tree string="My Models">
            <field name="name"/>
            <field name="partner_id"/>
            <field name="date"/>
            <field name="state"/>
            <field name="amount" sum="Total"/>
        </tree>
    </field>
</record>
```

### Advanced Tree (v17+)
```xml
<tree string="My Models"
      decoration-danger="state == 'cancel'"
      decoration-warning="state == 'draft'"
      decoration-success="state == 'done'"
      default_order="date desc">
    <field name="sequence" widget="handle"/>
    <field name="name"/>
    <field name="partner_id"/>
    <field name="date"/>
    <field name="state" widget="badge"
           decoration-success="state == 'done'"
           decoration-info="state == 'confirmed'"
           decoration-warning="state == 'draft'"/>
    <field name="amount" sum="Total"/>
    <field name="company_id" column_invisible="True"/>
    <field name="internal_notes" optional="hide"/>
</tree>
```

### Editable Tree
```xml
<tree editable="bottom">  <!-- or "top" -->
    <field name="product_id"/>
    <field name="quantity"/>
    <field name="price_unit"/>
    <field name="subtotal" readonly="1"/>
</tree>
```

### Column Visibility (v17+)
```xml
<!-- Hide column completely -->
<field name="internal_id" column_invisible="True"/>

<!-- Optional column (user can show/hide) -->
<field name="notes" optional="hide"/>
<field name="important" optional="show"/>

<!-- Conditional column visibility -->
<field name="cost" column_invisible="not context.get('show_cost')"/>
```

---

## Search View

```xml
<record id="my_model_view_search" model="ir.ui.view">
    <field name="name">my.model.search</field>
    <field name="model">my.model</field>
    <field name="arch" type="xml">
        <search string="Search My Model">
            <!-- Search fields -->
            <field name="name"/>
            <field name="partner_id"/>
            <field name="user_id"/>

            <!-- Filters -->
            <separator/>
            <filter name="draft" string="Draft"
                    domain="[('state', '=', 'draft')]"/>
            <filter name="confirmed" string="Confirmed"
                    domain="[('state', '=', 'confirmed')]"/>
            <separator/>
            <filter name="my_records" string="My Records"
                    domain="[('user_id', '=', uid)]"/>
            <separator/>
            <filter name="today" string="Today"
                    domain="[('date', '=', context_today().strftime('%Y-%m-%d'))]"/>
            <filter name="this_month" string="This Month"
                    domain="[('date', '>=', (context_today() - relativedelta(day=1)).strftime('%Y-%m-%d')),
                             ('date', '&lt;', (context_today() + relativedelta(months=1, day=1)).strftime('%Y-%m-%d'))]"/>

            <!-- Group By -->
            <group expand="0" string="Group By">
                <filter name="group_state" string="Status"
                        context="{'group_by': 'state'}"/>
                <filter name="group_partner" string="Partner"
                        context="{'group_by': 'partner_id'}"/>
                <filter name="group_date" string="Date"
                        context="{'group_by': 'date:month'}"/>
            </group>

            <!-- Search Panel (left sidebar) -->
            <searchpanel>
                <field name="state" icon="fa-filter" enable_counters="1"/>
                <field name="category_id" icon="fa-folder" enable_counters="1"/>
            </searchpanel>
        </search>
    </field>
</record>
```

---

## Kanban View

```xml
<record id="my_model_view_kanban" model="ir.ui.view">
    <field name="name">my.model.kanban</field>
    <field name="model">my.model</field>
    <field name="arch" type="xml">
        <kanban default_group_by="state"
                class="o_kanban_small_column"
                on_create="quick_create"
                quick_create_view="my_module.my_model_view_form_quick_create">
            <field name="id"/>
            <field name="name"/>
            <field name="partner_id"/>
            <field name="state"/>
            <field name="color"/>
            <templates>
                <t t-name="kanban-box">
                    <div t-attf-class="oe_kanban_card oe_kanban_global_click #{kanban_color(record.color.raw_value)}">
                        <div class="oe_kanban_content">
                            <div class="o_kanban_record_top">
                                <div class="o_kanban_record_headings">
                                    <strong class="o_kanban_record_title">
                                        <field name="name"/>
                                    </strong>
                                </div>
                                <div class="o_dropdown_kanban dropdown">
                                    <a role="button" class="dropdown-toggle o-no-caret btn"
                                       data-bs-toggle="dropdown" href="#">
                                        <span class="fa fa-ellipsis-v"/>
                                    </a>
                                    <div class="dropdown-menu" role="menu">
                                        <a t-if="widget.editable" role="menuitem"
                                           type="edit" class="dropdown-item">Edit</a>
                                        <a t-if="widget.deletable" role="menuitem"
                                           type="delete" class="dropdown-item">Delete</a>
                                    </div>
                                </div>
                            </div>
                            <div class="o_kanban_record_body">
                                <field name="partner_id"/>
                            </div>
                            <div class="o_kanban_record_bottom">
                                <div class="oe_kanban_bottom_left">
                                    <field name="priority" widget="priority"/>
                                </div>
                                <div class="oe_kanban_bottom_right">
                                    <field name="user_id" widget="many2one_avatar_user"/>
                                </div>
                            </div>
                        </div>
                    </div>
                </t>
            </templates>
        </kanban>
    </field>
</record>
```

---

## View Inheritance

### Basic Inheritance
```xml
<record id="view_partner_form_inherit" model="ir.ui.view">
    <field name="name">res.partner.form.inherit.my_module</field>
    <field name="model">res.partner</field>
    <field name="inherit_id" ref="base.view_partner_form"/>
    <field name="arch" type="xml">
        <!-- Add field after existing field -->
        <xpath expr="//field[@name='email']" position="after">
            <field name="x_custom_field"/>
        </xpath>

        <!-- Add field before existing field -->
        <xpath expr="//field[@name='phone']" position="before">
            <field name="x_another_field"/>
        </xpath>

        <!-- Replace field -->
        <xpath expr="//field[@name='website']" position="replace">
            <field name="website" widget="url"/>
        </xpath>

        <!-- Add attributes -->
        <xpath expr="//field[@name='name']" position="attributes">
            <attribute name="required">1</attribute>
        </xpath>

        <!-- Add inside element -->
        <xpath expr="//group[@name='sale']" position="inside">
            <field name="x_sales_field"/>
        </xpath>

        <!-- Add new page to notebook -->
        <xpath expr="//notebook" position="inside">
            <page string="Custom" name="custom">
                <group>
                    <field name="x_custom_field"/>
                </group>
            </page>
        </xpath>
    </field>
</record>
```

### XPath Expressions

| Expression | Matches |
|------------|---------|
| `//field[@name='x']` | Field with name='x' |
| `//group[@name='x']` | Group with name='x' |
| `//page[@name='x']` | Page with name='x' |
| `//button[@name='x']` | Button with name='x' |
| `//notebook` | First notebook |
| `//sheet` | The sheet element |
| `//div[@class='x']` | Div with specific class |

### Position Values

| Position | Action |
|----------|--------|
| `before` | Insert before matched element |
| `after` | Insert after matched element |
| `inside` | Insert as last child |
| `replace` | Replace entire element |
| `attributes` | Modify attributes only |

### CRITICAL: Always Verify XPath Expressions

**ALWAYS read the parent view structure before writing inheritance code.** XPath expressions must match the ACTUAL view structure, not assumptions.

#### Common Mistakes
```xml
<!-- ❌ WRONG: Assuming structure without verification -->
<xpath expr="//div[hasclass('flex-row')]" position="inside">
    <field name="x_custom_field"/>
</xpath>

<!-- The actual view might use QWeb templates: -->
<!-- <t t-name="card"><div><div class="flex-row">... -->
```

#### Correct Workflow
```python
# 1. FIRST: Read the parent view to understand structure
# Read base.res_users_apikeys_view_kanban

# 2. THEN: Write correct xpath based on actual structure
```

```xml
<!-- ✅ CORRECT: Verified against actual view structure -->
<record id="res_users_apikeys_view_kanban_inherit" model="ir.ui.view">
    <field name="name">res.users.apikeys.kanban.inherit</field>
    <field name="model">res.users.apikeys</field>
    <field name="inherit_id" ref="base.res_users_apikeys_view_kanban"/>
    <field name="arch" type="xml">
        <!-- Correct xpath after reading actual view structure -->
        <xpath expr="//t[@t-name='card']/div/div" position="inside">
            <span t-if="record.is_readonly.raw_value"
                  class="badge text-bg-warning ms-2"
                  title="This API key can only perform read operations"/>
        </xpath>
    </field>
</record>
```

#### Best Practice Checklist
- ✅ Read parent view XML file first
- ✅ Identify exact element structure (div, t, group, etc.)
- ✅ Note QWeb templates (t-name, t-if, etc.)
- ✅ Verify class names and attributes
- ✅ Test xpath matches target element
- ❌ Never assume structure based on common patterns

---

## Actions and Menus

### Window Action
```xml
<record id="my_model_action" model="ir.actions.act_window">
    <field name="name">My Models</field>
    <field name="res_model">my.model</field>
    <field name="view_mode">tree,form,kanban</field>
    <field name="domain">[('active', '=', True)]</field>
    <field name="context">{'search_default_my_records': 1}</field>
    <field name="help" type="html">
        <p class="o_view_nocontent_smiling_face">
            Create your first record
        </p>
        <p>
            Click the button to get started.
        </p>
    </field>
</record>
```

### Menu Items
```xml
<!-- Root menu -->
<menuitem id="my_module_menu_root"
          name="My Module"
          sequence="10"
          web_icon="my_module,static/description/icon.png"/>

<!-- Submenu -->
<menuitem id="my_module_menu_main"
          name="Main Menu"
          parent="my_module_menu_root"
          sequence="10"/>

<!-- Action menu item -->
<menuitem id="my_model_menu"
          name="My Models"
          parent="my_module_menu_main"
          action="my_model_action"
          sequence="10"/>
```

---

## Common Widgets

| Widget | Field Types | Purpose |
|--------|-------------|---------|
| `statusbar` | Selection | Status bar display |
| `badge` | Selection | Colored badge |
| `priority` | Selection | Star rating |
| `many2one_avatar_user` | Many2one | User avatar |
| `many2many_tags` | Many2many | Tag chips |
| `monetary` | Float/Monetary | Currency display |
| `handle` | Integer | Drag handle |
| `boolean_toggle` | Boolean | Toggle switch |
| `date` | Date | Date picker |
| `datetime` | Datetime | Datetime picker |
| `image` | Binary | Image display |
| `url` | Char | Clickable URL |
| `email` | Char | Mailto link |
| `phone` | Char | Tel link |
| `html` | Html | Rich text editor |
| `progressbar` | Float/Integer | Progress bar |

---

## Version-Specific Summary

| Feature | v14-v16 | v17+ |
|---------|---------|------|
| Visibility | `attrs="{'invisible': [...]}"` | `invisible="expr"` |
| Readonly | `attrs="{'readonly': [...]}"` | `readonly="expr"` |
| Required | `attrs="{'required': [...]}"` | `required="expr"` |
| Column hide | N/A | `column_invisible="True"` |
| Optional cols | Limited | `optional="show/hide"` |

---


## Source: widget-field-patterns.md

# Widget and Field Rendering Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  WIDGET & FIELD RENDERING PATTERNS                                           ║
║  Field widgets, custom rendering, and UI components                          ║
║  Use for customizing field display in forms, trees, and kanban views         ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Common Widgets Reference

### Text and Selection Widgets
| Widget | Field Types | Description |
|--------|-------------|-------------|
| `char` | Char | Default text input |
| `text` | Text | Multiline textarea |
| `html` | Html | Rich text editor |
| `email` | Char | Email with mailto link |
| `url` | Char | URL with clickable link |
| `phone` | Char | Phone with tel: link |
| `selection` | Selection | Dropdown select |
| `radio` | Selection | Radio buttons |
| `badge` | Selection | Colored badge display |
| `statusbar` | Selection | Status bar progression |

### Numeric Widgets
| Widget | Field Types | Description |
|--------|-------------|-------------|
| `integer` | Integer | Default integer |
| `float` | Float | Default decimal |
| `monetary` | Float/Monetary | Currency formatted |
| `percentage` | Float | Percentage display |
| `progressbar` | Float/Integer | Progress bar |
| `float_time` | Float | Hours:minutes format |
| `handle` | Integer | Drag handle for reordering |

### Date and Time Widgets
| Widget | Field Types | Description |
|--------|-------------|-------------|
| `date` | Date | Date picker |
| `datetime` | Datetime | Date and time picker |
| `daterange` | Date | Date range picker |
| `remaining_days` | Date | Days remaining display |

### Relational Widgets
| Widget | Field Types | Description |
|--------|-------------|-------------|
| `many2one` | Many2one | Default dropdown |
| `many2one_avatar` | Many2one | With avatar image |
| `many2one_avatar_user` | Many2one | User with avatar |
| `many2one_avatar_employee` | Many2one | Employee with avatar |
| `many2many_tags` | Many2many | Tag pills |
| `many2many_tags_avatar` | Many2many | Tags with avatars |
| `many2many_checkboxes` | Many2many | Checkbox list |
| `one2many` | One2many | Inline list/table |
| `section_and_note_one2many` | One2many | With sections (sale lines) |

### Binary Widgets
| Widget | Field Types | Description |
|--------|-------------|-------------|
| `binary` | Binary | File upload |
| `image` | Binary | Image display/upload |
| `signature` | Binary | Signature pad |
| `pdf_viewer` | Binary | PDF inline viewer |

### Special Widgets
| Widget | Field Types | Description |
|--------|-------------|-------------|
| `boolean_toggle` | Boolean | Toggle switch |
| `boolean_favorite` | Boolean | Star icon |
| `priority` | Selection | Star priority |
| `color_picker` | Integer | Color selection |
| `domain` | Char | Domain builder |
| `ace` | Text | Code editor |
| `copy_clipboard_char` | Char | Copy button |
| `copy_clipboard_text` | Text | Copy button |

---

## Widget Usage in Views

### Form View Widgets
```xml
<form>
    <sheet>
        <group>
            <!-- Text widgets -->
            <field name="name"/>
            <field name="email" widget="email"/>
            <field name="website" widget="url"/>
            <field name="phone" widget="phone"/>
            <field name="description" widget="html"/>

            <!-- Selection widgets -->
            <field name="type" widget="radio"/>
            <field name="priority" widget="priority"/>
            <field name="state" widget="badge"/>

            <!-- Numeric widgets -->
            <field name="amount" widget="monetary"/>
            <field name="discount" widget="percentage"/>
            <field name="progress" widget="progressbar"/>
            <field name="duration" widget="float_time"/>

            <!-- Date widgets -->
            <field name="date_deadline" widget="remaining_days"/>

            <!-- Relational widgets -->
            <field name="user_id" widget="many2one_avatar_user"/>
            <field name="tag_ids" widget="many2many_tags"/>
            <field name="category_ids" widget="many2many_checkboxes"/>

            <!-- Binary widgets -->
            <field name="image" widget="image"/>
            <field name="signature" widget="signature"/>
            <field name="document" widget="binary"/>

            <!-- Special widgets -->
            <field name="is_favorite" widget="boolean_favorite" nolabel="1"/>
            <field name="active" widget="boolean_toggle"/>
            <field name="color" widget="color_picker"/>
        </group>
    </sheet>
</form>
```

### Tree View Widgets
```xml
<tree>
    <!-- Handle for drag-drop reordering -->
    <field name="sequence" widget="handle"/>

    <field name="name"/>
    <field name="state" widget="badge" decoration-success="state == 'done'"
           decoration-warning="state == 'pending'"
           decoration-danger="state == 'cancel'"/>
    <field name="amount" widget="monetary"/>
    <field name="progress" widget="progressbar"/>
    <field name="user_id" widget="many2one_avatar_user"/>
    <field name="tag_ids" widget="many2many_tags"/>
    <field name="is_favorite" widget="boolean_favorite" nolabel="1"/>
</tree>
```

### Kanban View Widgets
```xml
<kanban>
    <templates>
        <t t-name="kanban-box">
            <div class="oe_kanban_card">
                <!-- Avatar widget in kanban -->
                <field name="user_id" widget="many2one_avatar_user"/>

                <!-- Priority stars -->
                <field name="priority" widget="priority"/>

                <!-- Progress bar -->
                <field name="progress" widget="progressbar"/>

                <!-- Tags -->
                <field name="tag_ids" widget="many2many_tags"
                       options="{'color_field': 'color'}"/>
            </div>
        </t>
    </templates>
</kanban>
```

---

## Widget Options

### Many2one Options
```xml
<!-- No create option -->
<field name="partner_id" options="{'no_create': True}"/>

<!-- No create, no edit, no open -->
<field name="partner_id" options="{
    'no_create': True,
    'no_create_edit': True,
    'no_open': True
}"/>

<!-- Custom create label -->
<field name="partner_id" options="{'create_name_field': 'display_name'}"/>
```

### Many2many Tags Options
```xml
<!-- With color -->
<field name="tag_ids" widget="many2many_tags"
       options="{'color_field': 'color'}"/>

<!-- No create, limited -->
<field name="tag_ids" widget="many2many_tags"
       options="{'no_create': True, 'limit': 5}"/>
```

### Monetary Options
```xml
<!-- Specify currency field -->
<field name="amount" widget="monetary"
       options="{'currency_field': 'currency_id'}"/>
```

### Image Options
```xml
<!-- With size and preview -->
<field name="image" widget="image"
       options="{'size': [128, 128], 'preview_image': 'image_128'}"/>

<!-- Zoom on click -->
<field name="image" widget="image" options="{'zoom': true}"/>
```

### Progressbar Options
```xml
<!-- With current/max values -->
<field name="progress" widget="progressbar"
       options="{'current_value': 'done_count', 'max_value': 'total_count'}"/>

<!-- Editable -->
<field name="progress" widget="progressbar"
       options="{'editable': true}"/>
```

### Statusbar Options
```xml
<!-- Visible states -->
<field name="state" widget="statusbar"
       statusbar_visible="draft,confirmed,done"/>

<!-- Clickable states -->
<field name="state" widget="statusbar"
       options="{'clickable': '1'}"/>
```

---

## Field Decorations

### Tree View Decorations
```xml
<tree decoration-success="state == 'done'"
      decoration-warning="state == 'pending'"
      decoration-danger="state == 'cancel'"
      decoration-info="state == 'draft'"
      decoration-muted="not active"
      decoration-bf="is_important">
    <field name="name"/>
    <field name="state"/>
    <field name="active" column_invisible="True"/>
    <field name="is_important" column_invisible="True"/>
</tree>
```

### Available Decorations
| Decoration | Style |
|------------|-------|
| `decoration-bf` | Bold |
| `decoration-it` | Italic |
| `decoration-success` | Green |
| `decoration-info` | Blue |
| `decoration-warning` | Orange |
| `decoration-danger` | Red |
| `decoration-muted` | Gray |
| `decoration-primary` | Primary color |
| `decoration-secondary` | Secondary color |

---

## Conditional Widget Display

### Readonly Conditions
```xml
<!-- Readonly based on state -->
<field name="amount" readonly="state != 'draft'"/>

<!-- Readonly based on field value -->
<field name="partner_id" readonly="is_locked"/>
```

### Invisible Conditions
```xml
<!-- Hide based on type -->
<field name="product_id" invisible="type != 'product'"/>

<!-- Hide based on state -->
<field name="cancel_reason" invisible="state != 'cancel'"/>
```

### Required Conditions
```xml
<!-- Required based on type -->
<field name="partner_id" required="type == 'customer'"/>

<!-- Required based on state -->
<field name="date_done" required="state == 'done'"/>
```

---

## Special Field Patterns

### Statusbar with Buttons
```xml
<header>
    <button name="action_confirm" string="Confirm"
            type="object" invisible="state != 'draft'"
            class="oe_highlight"/>
    <button name="action_done" string="Done"
            type="object" invisible="state != 'confirmed'"
            class="oe_highlight"/>
    <button name="action_cancel" string="Cancel"
            type="object" invisible="state in ['done', 'cancel']"/>
    <field name="state" widget="statusbar"
           statusbar_visible="draft,confirmed,done"/>
</header>
```

### Priority Stars
```xml
<!-- Model definition -->
priority = fields.Selection([
    ('0', 'Normal'),
    ('1', 'Low'),
    ('2', 'High'),
    ('3', 'Very High'),
], default='0')

<!-- In view -->
<field name="priority" widget="priority"/>
```

### Favorite Star
```xml
<!-- Model definition -->
is_favorite = fields.Boolean(default=False)

<!-- In view -->
<field name="is_favorite" widget="boolean_favorite" nolabel="1"/>
```

### Color Picker
```xml
<!-- Model definition -->
color = fields.Integer(string='Color Index')

<!-- In view -->
<field name="color" widget="color_picker"/>
```

### Handle for Reordering
```xml
<!-- Model definition -->
sequence = fields.Integer(default=10)

<!-- In tree view -->
<tree>
    <field name="sequence" widget="handle"/>
    <field name="name"/>
</tree>
```

---

## Badge Decorations

### State Badge
```xml
<!-- Selection field -->
state = fields.Selection([
    ('draft', 'Draft'),
    ('confirmed', 'Confirmed'),
    ('done', 'Done'),
    ('cancel', 'Cancelled'),
], default='draft')

<!-- In view with colors -->
<field name="state" widget="badge"
       decoration-success="state == 'done'"
       decoration-info="state == 'draft'"
       decoration-warning="state == 'confirmed'"
       decoration-danger="state == 'cancel'"/>
```

### Type Badge
```xml
<field name="type" widget="badge"
       decoration-primary="type == 'product'"
       decoration-secondary="type == 'service'"/>
```

---

## Monetary Field Pattern

### Complete Setup
```python
# Model definition
class MyModel(models.Model):
    _name = 'my.model'

    currency_id = fields.Many2one(
        'res.currency',
        string='Currency',
        default=lambda self: self.env.company.currency_id,
    )
    amount = fields.Monetary(currency_field='currency_id')
    amount_untaxed = fields.Monetary(currency_field='currency_id')
    amount_tax = fields.Monetary(currency_field='currency_id')
    amount_total = fields.Monetary(
        currency_field='currency_id',
        compute='_compute_amount_total',
        store=True,
    )

    @api.depends('amount_untaxed', 'amount_tax')
    def _compute_amount_total(self):
        for record in self:
            record.amount_total = record.amount_untaxed + record.amount_tax
```

```xml
<!-- View -->
<group>
    <field name="currency_id" groups="base.group_multi_currency"/>
    <field name="amount_untaxed"/>
    <field name="amount_tax"/>
    <field name="amount_total"/>
</group>
```

---

## Version-Specific Notes

### Odoo 17+ (visibility attributes)
```xml
<!-- New syntax -->
<field name="field1" invisible="condition"/>
<field name="field2" readonly="condition"/>
<field name="field3" required="condition"/>
<field name="field4" column_invisible="condition"/>
```

### Odoo 14-16 (attrs)
```xml
<!-- Old syntax -->
<field name="field1" attrs="{'invisible': [('condition', '=', True)]}"/>
<field name="field2" attrs="{'readonly': [('state', '!=', 'draft')]}"/>
<field name="field3" attrs="{'required': [('type', '=', 'customer')]}"/>
```

---

## Best Practices

1. **Use appropriate widgets** - Match widget to data type and UX need
2. **Set widget options** - Configure behavior with options dict
3. **Use decorations** - Visual cues for states and priorities
4. **Handle currency** - Always specify currency_field for monetary
5. **Use avatar widgets** - Better UX for user/partner fields
6. **Status progression** - Use statusbar for state workflows
7. **Drag reordering** - Add handle widget for sequences
8. **Tags with colors** - Use color_field option for tags
9. **Conditional display** - Use invisible/readonly/required
10. **Version awareness** - attrs vs direct attributes

---


## Source: qweb-template-patterns.md

# QWeb Template Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  QWEB TEMPLATE PATTERNS                                                      ║
║  QWeb templating language for views, reports, and website pages              ║
║  Use for dynamic HTML generation in Odoo                                     ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## QWeb Directives Overview

| Directive | Purpose |
|-----------|---------|
| `t-if` | Conditional rendering |
| `t-elif` | Else-if condition |
| `t-else` | Else branch |
| `t-foreach` | Loop iteration |
| `t-set` | Variable assignment |
| `t-esc` | Escaped output |
| `t-out` | Output (escaped by default) |
| `t-raw` | Unescaped HTML output |
| `t-call` | Include another template |
| `t-att` | Dynamic attribute |
| `t-attf` | Format string attribute |
| `t-field` | Field rendering |
| `t-options` | Field options |

---

## Conditional Rendering

### Basic If/Else
```xml
<t t-if="record.state == 'draft'">
    <span class="badge bg-secondary">Draft</span>
</t>
<t t-elif="record.state == 'confirmed'">
    <span class="badge bg-primary">Confirmed</span>
</t>
<t t-elif="record.state == 'done'">
    <span class="badge bg-success">Done</span>
</t>
<t t-else="">
    <span class="badge bg-danger">Cancelled</span>
</t>
```

### Conditional on Element
```xml
<div t-if="record.partner_id">
    Partner: <t t-esc="record.partner_id.name"/>
</div>

<span t-if="record.amount > 0" class="text-success">
    <t t-esc="record.amount"/>
</span>
<span t-else="" class="text-danger">No amount</span>
```

### Multiple Conditions
```xml
<t t-if="record.active and record.state == 'confirmed'">
    Active and confirmed
</t>

<t t-if="record.type in ['product', 'service']">
    Valid type
</t>

<t t-if="record.amount >= 1000 or record.is_priority">
    High value or priority
</t>
```

---

## Loops and Iteration

### Basic For Loop
```xml
<t t-foreach="records" t-as="record">
    <div class="record-item">
        <span t-esc="record.name"/>
    </div>
</t>
```

### Loop with Index
```xml
<table>
    <t t-foreach="lines" t-as="line">
        <tr>
            <!-- line_index: 0-based index -->
            <td t-esc="line_index + 1"/>
            <!-- line_value: same as line -->
            <td t-esc="line.name"/>
            <!-- line_size: total count -->
            <td t-if="line_index == line_size - 1">Last item</td>
        </tr>
    </t>
</table>
```

### Loop Variables
| Variable | Description |
|----------|-------------|
| `{name}` | Current item |
| `{name}_index` | 0-based index |
| `{name}_size` | Total count |
| `{name}_first` | True if first |
| `{name}_last` | True if last |
| `{name}_odd` | True if odd index |
| `{name}_even` | True if even index |
| `{name}_value` | Same as {name} |

### Loop with Parity
```xml
<t t-foreach="items" t-as="item">
    <tr t-attf-class="#{item_odd and 'odd' or 'even'}">
        <td t-esc="item.name"/>
    </tr>
</t>
```

### Nested Loops
```xml
<t t-foreach="orders" t-as="order">
    <div class="order">
        <h3 t-esc="order.name"/>
        <t t-foreach="order.line_ids" t-as="line">
            <div class="line">
                <span t-esc="line.product_id.name"/>
                <span t-esc="line.quantity"/>
            </div>
        </t>
    </div>
</t>
```

---

## Variables and Expressions

### Setting Variables
```xml
<!-- Simple assignment -->
<t t-set="total" t-value="0"/>

<!-- Expression assignment -->
<t t-set="total" t-value="sum(line.amount for line in lines)"/>

<!-- Content assignment -->
<t t-set="greeting">
    Hello, <t t-esc="user.name"/>!
</t>
```

### Using Variables
```xml
<t t-set="has_lines" t-value="bool(record.line_ids)"/>

<div t-if="has_lines">
    <t t-set="line_count" t-value="len(record.line_ids)"/>
    <span>Total lines: <t t-esc="line_count"/></span>
</div>
```

### Calculations
```xml
<t t-set="subtotal" t-value="record.quantity * record.price"/>
<t t-set="tax" t-value="subtotal * 0.21"/>
<t t-set="total" t-value="subtotal + tax"/>

<div>
    Subtotal: <t t-esc="subtotal"/>
    Tax: <t t-esc="tax"/>
    Total: <t t-esc="total"/>
</div>
```

---

## Output and Escaping

### Escaped Output (Safe)
```xml
<!-- t-esc escapes HTML characters -->
<span t-esc="record.name"/>

<!-- t-out also escapes by default -->
<span t-out="record.description"/>
```

### Unescaped HTML Output
```xml
<!-- t-raw renders HTML as-is (use carefully!) -->
<div t-raw="record.html_content"/>

<!-- t-out with Markup object -->
<div t-out="record.rendered_html"/>
```

### Field Output (Reports/Website)
```xml
<!-- t-field renders with formatting -->
<span t-field="record.date"/>
<span t-field="record.amount"/>
<span t-field="record.partner_id"/>

<!-- t-field with options -->
<span t-field="record.amount" t-options="{'widget': 'monetary'}"/>
<span t-field="record.date" t-options="{'format': 'dd/MM/yyyy'}"/>
```

---

## Dynamic Attributes

### t-att: Dynamic Attribute Value
```xml
<!-- Single attribute -->
<div t-att-class="record.state"/>
<input t-att-value="record.name"/>
<a t-att-href="record.url"/>

<!-- Conditional attribute -->
<div t-att-class="record.active and 'active' or 'inactive'"/>
```

### t-attf: Format String Attribute
```xml
<!-- String interpolation with #{} -->
<div t-attf-class="card state-#{record.state}"/>
<a t-attf-href="/web#id=#{record.id}&amp;model=my.model"/>

<!-- Combine static and dynamic -->
<div t-attf-class="container #{record.is_important and 'important' or ''}"/>
```

### t-att for Multiple Attributes
```xml
<!-- Dictionary of attributes -->
<div t-att="{'class': 'my-class', 'data-id': record.id}"/>
```

### Conditional Attributes
```xml
<!-- Attribute only if condition true -->
<input type="checkbox" t-att-checked="record.active and 'checked'"/>
<button t-att-disabled="not record.can_edit and 'disabled'"/>

<!-- Class list based on conditions -->
<div t-attf-class="
    card
    #{record.state == 'done' and 'bg-success' or ''}
    #{record.is_priority and 'border-warning' or ''}
"/>
```

---

## Template Inheritance and Calls

### Calling Other Templates
```xml
<!-- Define a template -->
<template id="address_block">
    <div class="address">
        <t t-esc="partner.street"/>
        <t t-esc="partner.city"/>
        <t t-esc="partner.country_id.name"/>
    </div>
</template>

<!-- Call the template -->
<t t-call="my_module.address_block">
    <t t-set="partner" t-value="record.partner_id"/>
</t>
```

### Template with Parameters
```xml
<!-- Template expecting parameters -->
<template id="price_display">
    <span t-attf-class="price #{highlight and 'text-success' or ''}">
        <t t-esc="amount"/>
        <t t-esc="currency"/>
    </span>
</template>

<!-- Call with parameters -->
<t t-call="my_module.price_display">
    <t t-set="amount" t-value="record.amount_total"/>
    <t t-set="currency" t-value="record.currency_id.symbol"/>
    <t t-set="highlight" t-value="True"/>
</t>
```

### Inheriting Templates
```xml
<!-- Extend existing template -->
<template id="custom_layout" inherit_id="web.layout">
    <xpath expr="//head" position="inside">
        <link rel="stylesheet" href="/my_module/static/src/css/custom.css"/>
    </xpath>
</template>
```

---

## Report-Specific Patterns

### Report Document Structure
```xml
<template id="report_my_document">
    <t t-call="web.html_container">
        <t t-foreach="docs" t-as="doc">
            <t t-call="web.external_layout">
                <div class="page">
                    <h1 t-field="doc.name"/>
                    <!-- Report content -->
                </div>
            </t>
        </t>
    </t>
</template>
```

### Field Formatting in Reports
```xml
<!-- Date formatting -->
<span t-field="doc.date_order" t-options="{'format': 'dd MMMM yyyy'}"/>

<!-- Monetary formatting -->
<span t-field="doc.amount_total"
      t-options="{'widget': 'monetary', 'display_currency': doc.currency_id}"/>

<!-- Duration formatting -->
<span t-field="doc.duration" t-options="{'widget': 'duration'}"/>

<!-- Address formatting -->
<div t-field="doc.partner_id"
     t-options="{'widget': 'contact', 'fields': ['address', 'phone', 'email']}"/>
```

### Table with Totals
```xml
<table class="table table-sm">
    <thead>
        <tr>
            <th>Product</th>
            <th class="text-end">Quantity</th>
            <th class="text-end">Price</th>
            <th class="text-end">Subtotal</th>
        </tr>
    </thead>
    <tbody>
        <t t-foreach="doc.line_ids" t-as="line">
            <tr>
                <td t-esc="line.product_id.name"/>
                <td class="text-end" t-esc="line.quantity"/>
                <td class="text-end" t-field="line.price_unit"/>
                <td class="text-end" t-field="line.price_subtotal"/>
            </tr>
        </t>
    </tbody>
    <tfoot>
        <tr>
            <td colspan="3" class="text-end"><strong>Total:</strong></td>
            <td class="text-end" t-field="doc.amount_total"/>
        </tr>
    </tfoot>
</table>
```

---

## Kanban View QWeb

### Kanban Card Template
```xml
<kanban>
    <field name="name"/>
    <field name="state"/>
    <field name="partner_id"/>
    <field name="color"/>
    <templates>
        <t t-name="kanban-box">
            <div t-attf-class="oe_kanban_card oe_kanban_global_click
                               o_kanban_record_has_image_fill
                               #{record.color.raw_value ? 'oe_kanban_color_' + record.color.raw_value : ''}">
                <div class="oe_kanban_content">
                    <div class="o_kanban_record_top">
                        <div class="o_kanban_record_headings">
                            <strong class="o_kanban_record_title">
                                <field name="name"/>
                            </strong>
                        </div>
                    </div>
                    <div class="o_kanban_record_body">
                        <field name="partner_id"/>
                    </div>
                    <div class="o_kanban_record_bottom">
                        <div class="oe_kanban_bottom_left">
                            <field name="priority" widget="priority"/>
                        </div>
                        <div class="oe_kanban_bottom_right">
                            <field name="user_id" widget="many2one_avatar_user"/>
                        </div>
                    </div>
                </div>
            </div>
        </t>
    </templates>
</kanban>
```

### Kanban with Dropdown Menu
```xml
<div class="o_dropdown_kanban dropdown">
    <a role="button" class="dropdown-toggle o-no-caret btn"
       data-bs-toggle="dropdown" href="#">
        <span class="fa fa-ellipsis-v"/>
    </a>
    <div class="dropdown-menu" role="menu">
        <a role="menuitem" type="edit" class="dropdown-item">Edit</a>
        <a role="menuitem" type="delete" class="dropdown-item">Delete</a>
        <div role="separator" class="dropdown-divider"/>
        <a role="menuitem" type="object" name="action_archive"
           class="dropdown-item">Archive</a>
    </div>
</div>
```

---

## Website QWeb Patterns

### Page Template
```xml
<template id="my_page" name="My Page">
    <t t-call="website.layout">
        <div id="wrap" class="oe_structure">
            <section class="container py-5">
                <h1>My Page Title</h1>
                <t t-foreach="records" t-as="record">
                    <div class="card mb-3">
                        <div class="card-body">
                            <h5 t-esc="record.name"/>
                            <p t-raw="record.description"/>
                        </div>
                    </div>
                </t>
            </section>
        </div>
    </t>
</template>
```

### Portal Template
```xml
<template id="portal_my_records" name="My Records">
    <t t-call="portal.portal_layout">
        <t t-set="breadcrumbs_searchbar" t-value="True"/>
        <t t-call="portal.portal_searchbar">
            <t t-set="title">My Records</t>
        </t>
        <t t-if="records">
            <t t-foreach="records" t-as="record">
                <div class="card mb-2">
                    <div class="card-body">
                        <a t-attf-href="/my/records/#{record.id}">
                            <t t-esc="record.name"/>
                        </a>
                    </div>
                </div>
            </t>
        </t>
        <t t-else="">
            <p>No records found.</p>
        </t>
    </t>
</template>
```

---

## Useful Expressions

### String Operations
```xml
<t t-esc="record.name.upper()"/>
<t t-esc="record.name[:20]"/>
<t t-esc="', '.join(record.tag_ids.mapped('name'))"/>
<t t-esc="record.name or 'No name'"/>
```

### Number Formatting
```xml
<t t-esc="'%.2f' % record.amount"/>
<t t-esc="'{:,.2f}'.format(record.amount)"/>
<t t-esc="int(record.progress)"/>
```

### Date Formatting
```xml
<!-- Using format_date helper (reports) -->
<t t-esc="format_date(env, record.date)"/>

<!-- Using strftime -->
<t t-esc="record.date.strftime('%d/%m/%Y') if record.date else ''"/>
```

### List Operations
```xml
<t t-esc="len(record.line_ids)"/>
<t t-esc="sum(record.line_ids.mapped('amount'))"/>
<t t-esc="record.line_ids.filtered(lambda l: l.state == 'done')"/>
```

---

## Best Practices

1. **Use t-esc for safety** - Always escape user content
2. **Use t-field in reports** - Proper formatting and translation
3. **Keep logic minimal** - Complex logic belongs in Python
4. **Use t-call for reuse** - DRY principle for templates
5. **Name templates clearly** - Descriptive IDs for maintenance
6. **Use semantic HTML** - Proper structure for accessibility
7. **Handle empty states** - Always check for missing data
8. **Use Bootstrap classes** - Consistent styling with Odoo
9. **Test with real data** - Verify edge cases
10. **Version awareness** - QWeb syntax stable across versions

---


## Source: action-patterns.md

# Action Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ACTION PATTERNS                                                             ║
║  Window actions, server actions, client actions, and URL actions             ║
║  Use for navigation, automation, and user interface interactions             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Action Types Overview

| Type | Model | Use Case |
|------|-------|----------|
| Window | `ir.actions.act_window` | Open views (form, tree, kanban) |
| Server | `ir.actions.server` | Execute Python code |
| Client | `ir.actions.client` | JavaScript/OWL actions |
| URL | `ir.actions.act_url` | Open external URLs |
| Report | `ir.actions.report` | Generate PDF reports |

---

## Window Actions

### Basic Window Action (XML)
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="action_my_model" model="ir.actions.act_window">
        <field name="name">My Records</field>
        <field name="res_model">my.model</field>
        <field name="view_mode">tree,form,kanban</field>
        <field name="help" type="html">
            <p class="o_view_nocontent_smiling_face">
                Create your first record!
            </p>
        </field>
    </record>
</odoo>
```

### Window Action with Domain and Context
```xml
<record id="action_my_model_active" model="ir.actions.act_window">
    <field name="name">Active Records</field>
    <field name="res_model">my.model</field>
    <field name="view_mode">tree,form</field>
    <field name="domain">[('state', '=', 'active')]</field>
    <field name="context">{
        'default_state': 'active',
        'search_default_my_filter': 1,
        'group_by': 'category_id',
    }</field>
    <field name="limit">80</field>
</record>
```

### Window Action with Specific Views
```xml
<record id="action_my_model_custom" model="ir.actions.act_window">
    <field name="name">My Records (Custom)</field>
    <field name="res_model">my.model</field>
    <field name="view_mode">tree,form</field>
    <field name="view_ids" eval="[
        (5, 0, 0),
        (0, 0, {'view_mode': 'tree', 'view_id': ref('my_model_view_tree_custom')}),
        (0, 0, {'view_mode': 'form', 'view_id': ref('my_model_view_form_custom')}),
    ]"/>
</record>
```

### Open Specific Record
```xml
<record id="action_open_partner" model="ir.actions.act_window">
    <field name="name">Partner</field>
    <field name="res_model">res.partner</field>
    <field name="view_mode">form</field>
    <field name="res_id" ref="base.main_partner"/>
    <field name="target">current</field>
</record>
```

### Open as Dialog (Wizard)
```xml
<record id="action_wizard" model="ir.actions.act_window">
    <field name="name">My Wizard</field>
    <field name="res_model">my.wizard</field>
    <field name="view_mode">form</field>
    <field name="target">new</field>
    <field name="context">{
        'active_id': active_id,
        'active_ids': active_ids,
        'active_model': active_model,
    }</field>
</record>
```

### Target Options
| Target | Effect |
|--------|--------|
| `current` | Replace current view (default) |
| `new` | Open in dialog/popup |
| `inline` | Inline in current form |
| `fullscreen` | Fullscreen mode |
| `main` | Open in main content area |

---

## Window Actions from Python

### Return Action Dictionary
```python
def action_open_related(self):
    """Open related records."""
    self.ensure_one()
    return {
        'type': 'ir.actions.act_window',
        'name': 'Related Records',
        'res_model': 'related.model',
        'view_mode': 'tree,form',
        'domain': [('parent_id', '=', self.id)],
        'context': {
            'default_parent_id': self.id,
        },
    }

def action_open_single(self):
    """Open single record in form view."""
    return {
        'type': 'ir.actions.act_window',
        'res_model': 'my.model',
        'res_id': self.id,
        'view_mode': 'form',
        'target': 'current',
    }

def action_open_wizard(self):
    """Open wizard with context."""
    return {
        'type': 'ir.actions.act_window',
        'name': 'Configure',
        'res_model': 'my.wizard',
        'view_mode': 'form',
        'target': 'new',
        'context': {
            'default_record_id': self.id,
            'default_amount': self.amount_total,
        },
    }
```

### Use Existing Action
```python
def action_open_partners(self):
    """Use predefined action."""
    action = self.env.ref('base.action_partner_form').read()[0]
    action['domain'] = [('id', 'in', self.partner_ids.ids)]
    action['context'] = {'default_company_id': self.company_id.id}
    return action
```

---

## Server Actions

### Execute Python Code
```xml
<record id="action_mark_done" model="ir.actions.server">
    <field name="name">Mark as Done</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="binding_view_types">list,form</field>
    <field name="state">code</field>
    <field name="code">
if records:
    records.write({'state': 'done'})
    </field>
</record>
```

### Server Action with Notification
```xml
<record id="action_process_with_notify" model="ir.actions.server">
    <field name="name">Process Records</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="state">code</field>
    <field name="code">
count = len(records)
records.action_process()
action = {
    'type': 'ir.actions.client',
    'tag': 'display_notification',
    'params': {
        'title': 'Success',
        'message': f'Processed {count} records.',
        'type': 'success',
        'sticky': False,
    }
}
    </field>
</record>
```

### Server Action Opening Window
```xml
<record id="action_open_related" model="ir.actions.server">
    <field name="name">Open Related</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="state">code</field>
    <field name="code">
if records:
    action = {
        'type': 'ir.actions.act_window',
        'name': 'Related Records',
        'res_model': 'related.model',
        'view_mode': 'tree,form',
        'domain': [('parent_id', 'in', records.ids)],
    }
    </field>
</record>
```

### Server Action Types
| State | Description |
|-------|-------------|
| `code` | Execute Python code |
| `object_create` | Create new record |
| `object_write` | Update records |
| `multi` | Execute multiple actions |
| `email` | Send email |
| `sms` | Send SMS |
| `next_activity` | Schedule activity |

---

## Client Actions

### Display Notification
```python
def action_notify(self):
    """Show notification to user."""
    return {
        'type': 'ir.actions.client',
        'tag': 'display_notification',
        'params': {
            'title': 'Information',
            'message': 'Operation completed successfully.',
            'type': 'success',  # success, warning, danger, info
            'sticky': False,
            'next': {'type': 'ir.actions.act_window_close'},
        }
    }
```

### Notification with Link
```python
def action_notify_with_link(self):
    """Notification with clickable link."""
    return {
        'type': 'ir.actions.client',
        'tag': 'display_notification',
        'params': {
            'title': 'Created',
            'message': 'Record created successfully.',
            'type': 'success',
            'links': [{
                'label': 'View Record',
                'url': f'/web#id={self.id}&model={self._name}&view_type=form',
            }],
        }
    }
```

### Reload Page
```python
def action_reload(self):
    """Reload current view."""
    return {
        'type': 'ir.actions.client',
        'tag': 'reload',
    }
```

### Custom Client Action (OWL)
```xml
<!-- Register client action -->
<record id="action_custom_dashboard" model="ir.actions.client">
    <field name="name">My Dashboard</field>
    <field name="tag">my_module.dashboard</field>
</record>
```

```javascript
/** @odoo-module **/
import { registry } from "@web/core/registry";
import { Component } from "@odoo/owl";

class MyDashboard extends Component {
    static template = "my_module.Dashboard";
}

registry.category("actions").add("my_module.dashboard", MyDashboard);
```

---

## URL Actions

### Open External URL
```xml
<record id="action_open_docs" model="ir.actions.act_url">
    <field name="name">Documentation</field>
    <field name="url">https://www.odoo.com/documentation</field>
    <field name="target">new</field>
</record>
```

### Dynamic URL from Python
```python
def action_open_external(self):
    """Open external URL."""
    return {
        'type': 'ir.actions.act_url',
        'url': f'https://example.com/record/{self.external_id}',
        'target': 'new',  # or 'self' for same window
    }

def action_download_file(self):
    """Download file via URL."""
    return {
        'type': 'ir.actions.act_url',
        'url': f'/web/content/{attachment.id}?download=true',
        'target': 'self',
    }
```

---

## Action Binding

### Bind to Model (Action Menu)
```xml
<record id="action_batch_update" model="ir.actions.server">
    <field name="name">Batch Update</field>
    <field name="model_id" ref="model_my_model"/>
    <!-- Binding creates "Action" menu entry -->
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="binding_view_types">list</field>
    <field name="state">code</field>
    <field name="code">records.action_batch_update()</field>
</record>
```

### Binding View Types
| Type | Where Available |
|------|-----------------|
| `list` | Tree/list view only |
| `form` | Form view only |
| `list,form` | Both views |

---

## Close Action

### Close Dialog
```python
def action_close(self):
    """Close dialog/wizard."""
    return {'type': 'ir.actions.act_window_close'}
```

### Close with Notification
```python
def action_save_and_close(self):
    """Save and close with notification."""
    self._do_save()
    return {
        'type': 'ir.actions.client',
        'tag': 'display_notification',
        'params': {
            'title': 'Saved',
            'message': 'Changes saved successfully.',
            'type': 'success',
            'next': {'type': 'ir.actions.act_window_close'},
        }
    }
```

---

## Multi-Action Pattern

### Chain Actions
```python
def action_process_and_open(self):
    """Process then open related view."""
    self._process()

    # Return next action
    return {
        'type': 'ir.actions.act_window',
        'name': 'Processed Records',
        'res_model': 'processed.model',
        'view_mode': 'tree,form',
        'domain': [('source_id', '=', self.id)],
    }
```

### Conditional Actions
```python
def action_smart_open(self):
    """Open appropriate view based on record count."""
    related = self.env['related.model'].search([
        ('parent_id', '=', self.id),
    ])

    if len(related) == 1:
        # Single record - open form
        return {
            'type': 'ir.actions.act_window',
            'res_model': 'related.model',
            'res_id': related.id,
            'view_mode': 'form',
        }
    else:
        # Multiple records - open list
        return {
            'type': 'ir.actions.act_window',
            'name': 'Related Records',
            'res_model': 'related.model',
            'view_mode': 'tree,form',
            'domain': [('id', 'in', related.ids)],
        }
```

---

## Best Practices

1. **Use XML for static actions** - Easier to maintain and override
2. **Use Python for dynamic actions** - When domains/context depend on data
3. **Always set res_model** - Required for window actions
4. **Use context wisely** - Pass defaults and search filters
5. **binding_view_types** - Specify where action appears
6. **target=new for wizards** - Opens as dialog
7. **Return dict, not action record** - More flexible
8. **Handle empty recordsets** - Check before processing
9. **Use notifications** - Give user feedback
10. **Close dialogs properly** - Return act_window_close

---


## Source: menu-navigation-patterns.md

# Menu and Navigation Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MENU & NAVIGATION PATTERNS                                                  ║
║  Menu structure, navigation, and application organization                    ║
║  Use for module UI organization and user navigation                          ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Menu Structure Overview

```
Root Menu (App)
├── Category Menu 1
│   ├── Submenu 1.1 → Action
│   └── Submenu 1.2 → Action
├── Category Menu 2
│   ├── Submenu 2.1 → Action
│   └── Submenu 2.2 → Action
└── Configuration
    ├── Settings → Action
    └── Data → Action
```

---

## Basic Menu Definition

### Root Menu (Application)
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- Root menu (appears in app switcher) -->
    <menuitem id="menu_my_module_root"
              name="My Application"
              web_icon="my_module,static/description/icon.png"
              sequence="10"/>
</odoo>
```

### Category Menus
```xml
<!-- Category under root -->
<menuitem id="menu_my_module_main"
          name="Records"
          parent="menu_my_module_root"
          sequence="10"/>

<menuitem id="menu_my_module_config"
          name="Configuration"
          parent="menu_my_module_root"
          sequence="100"
          groups="base.group_system"/>
```

### Action Menus (Leaf Nodes)
```xml
<!-- Menu with action -->
<menuitem id="menu_my_model"
          name="My Records"
          parent="menu_my_module_main"
          action="action_my_model"
          sequence="10"/>

<menuitem id="menu_my_model_archived"
          name="Archived"
          parent="menu_my_module_main"
          action="action_my_model_archived"
          sequence="20"/>
```

---

## Complete Menu Example

### Full Module Menu Structure
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- ============================================ -->
    <!-- ROOT MENU (Application)                      -->
    <!-- ============================================ -->
    <menuitem id="menu_my_module_root"
              name="My Application"
              web_icon="my_module,static/description/icon.png"
              sequence="50"/>

    <!-- ============================================ -->
    <!-- MAIN MENUS                                   -->
    <!-- ============================================ -->

    <!-- Records Category -->
    <menuitem id="menu_records"
              name="Records"
              parent="menu_my_module_root"
              sequence="10"/>

    <menuitem id="menu_my_model"
              name="All Records"
              parent="menu_records"
              action="action_my_model"
              sequence="10"/>

    <menuitem id="menu_my_model_draft"
              name="Draft"
              parent="menu_records"
              action="action_my_model_draft"
              sequence="20"/>

    <menuitem id="menu_my_model_confirmed"
              name="Confirmed"
              parent="menu_records"
              action="action_my_model_confirmed"
              sequence="30"/>

    <!-- Reports Category -->
    <menuitem id="menu_reports"
              name="Reports"
              parent="menu_my_module_root"
              sequence="50"/>

    <menuitem id="menu_report_analysis"
              name="Analysis"
              parent="menu_reports"
              action="action_report_analysis"
              sequence="10"/>

    <!-- ============================================ -->
    <!-- CONFIGURATION MENUS                          -->
    <!-- ============================================ -->

    <menuitem id="menu_configuration"
              name="Configuration"
              parent="menu_my_module_root"
              sequence="100"
              groups="base.group_system"/>

    <menuitem id="menu_config_settings"
              name="Settings"
              parent="menu_configuration"
              action="action_my_module_config"
              sequence="10"/>

    <menuitem id="menu_config_categories"
              name="Categories"
              parent="menu_configuration"
              action="action_my_category"
              sequence="20"/>

    <menuitem id="menu_config_tags"
              name="Tags"
              parent="menu_configuration"
              action="action_my_tag"
              sequence="30"/>
</odoo>
```

---

## Menu Attributes

### Common Attributes
| Attribute | Description |
|-----------|-------------|
| `id` | Unique XML ID (required) |
| `name` | Display name |
| `parent` | Parent menu XML ID |
| `action` | Action to execute |
| `sequence` | Order (lower = first) |
| `groups` | Security groups (comma-separated) |
| `web_icon` | App icon (root menu only) |
| `active` | Enable/disable menu |

### Sequence Guidelines
| Range | Use For |
|-------|---------|
| 1-20 | Primary menus (most used) |
| 20-50 | Secondary menus |
| 50-80 | Reports and analysis |
| 80-100 | Configuration |

---

## Menu Security

### Restrict by Group
```xml
<!-- Only managers see this menu -->
<menuitem id="menu_sensitive"
          name="Sensitive Data"
          parent="menu_my_module_main"
          action="action_sensitive"
          groups="my_module.group_manager"/>

<!-- Multiple groups (OR) -->
<menuitem id="menu_admin"
          name="Admin"
          parent="menu_my_module_config"
          action="action_admin"
          groups="base.group_system,my_module.group_admin"/>
```

### Hide Menu Conditionally
```xml
<!-- Menu visible based on settings -->
<menuitem id="menu_optional_feature"
          name="Optional Feature"
          parent="menu_my_module_main"
          action="action_optional"
          groups="my_module.group_use_optional_feature"/>
```

---

## Extending Existing Menus

### Add to Existing App
```xml
<!-- Add menu under Sales app -->
<menuitem id="menu_my_sale_extension"
          name="My Extension"
          parent="sale.sale_menu_root"
          action="action_my_sale_extension"
          sequence="50"/>

<!-- Add under Sales > Configuration -->
<menuitem id="menu_my_sale_config"
          name="My Settings"
          parent="sale.menu_sale_config"
          action="action_my_sale_config"
          sequence="100"/>
```

### Common Parent Menus
```xml
<!-- Sales -->
parent="sale.sale_menu_root"
parent="sale.sale_order_menu"
parent="sale.menu_sale_config"

<!-- Purchase -->
parent="purchase.menu_purchase_root"
parent="purchase.menu_purchase_config"

<!-- Inventory -->
parent="stock.menu_stock_root"
parent="stock.menu_stock_config"

<!-- Accounting -->
parent="account.menu_finance"
parent="account.menu_finance_configuration"

<!-- CRM -->
parent="crm.crm_menu_root"
parent="crm.crm_menu_config"

<!-- HR -->
parent="hr.menu_hr_root"
parent="hr.menu_human_resources_configuration"

<!-- Project -->
parent="project.menu_main_pm"
parent="project.menu_project_config"

<!-- Settings (general) -->
parent="base.menu_administration"
```

---

## Menu with Filters

### Pre-filtered Menus
```xml
<!-- Action with domain -->
<record id="action_my_model_draft" model="ir.actions.act_window">
    <field name="name">Draft Records</field>
    <field name="res_model">my.model</field>
    <field name="view_mode">tree,form</field>
    <field name="domain">[('state', '=', 'draft')]</field>
    <field name="context">{'default_state': 'draft'}</field>
</record>

<menuitem id="menu_my_model_draft"
          name="Draft"
          parent="menu_records"
          action="action_my_model_draft"/>

<!-- Action with search filter -->
<record id="action_my_model_my_records" model="ir.actions.act_window">
    <field name="name">My Records</field>
    <field name="res_model">my.model</field>
    <field name="view_mode">tree,form</field>
    <field name="context">{'search_default_my_records': 1}</field>
</record>
```

---

## Dynamic Menus

### Create Menu from Python
```python
def _create_dynamic_menu(self, name, parent_id, action_id):
    """Create menu dynamically."""
    return self.env['ir.ui.menu'].create({
        'name': name,
        'parent_id': parent_id,
        'action': f'ir.actions.act_window,{action_id}',
        'sequence': 100,
    })
```

### Update Menu Visibility
```python
def _update_menu_visibility(self):
    """Show/hide menu based on configuration."""
    menu = self.env.ref('my_module.menu_optional_feature')
    config_param = self.env['ir.config_parameter'].sudo()
    show_menu = config_param.get_param('my_module.show_optional', 'False')
    menu.write({'active': show_menu == 'True'})
```

---

## App Icon

### Icon Requirements
- Format: PNG
- Size: 256x256 pixels (recommended)
- Location: `static/description/icon.png`

### Setting App Icon
```xml
<menuitem id="menu_my_module_root"
          name="My Application"
          web_icon="my_module,static/description/icon.png"
          sequence="50"/>
```

---

## Menu Order Patterns

### Standard App Layout
```
1. Main Operations (seq 10-20)
   - All Records
   - My Records
   - To Do / Pending

2. Secondary Operations (seq 30-50)
   - By Category views
   - Filtered views

3. Reports (seq 60-80)
   - Analysis
   - Dashboards

4. Configuration (seq 90-100)
   - Settings
   - Master Data
```

### Example Implementation
```xml
<!-- Main Operations -->
<menuitem id="menu_operations" name="Operations"
          parent="menu_root" sequence="10"/>

<menuitem id="menu_all" name="All Records"
          parent="menu_operations" action="action_all" sequence="10"/>

<menuitem id="menu_my" name="My Records"
          parent="menu_operations" action="action_my" sequence="20"/>

<!-- Reports -->
<menuitem id="menu_reporting" name="Reporting"
          parent="menu_root" sequence="60"/>

<menuitem id="menu_analysis" name="Analysis"
          parent="menu_reporting" action="action_analysis" sequence="10"/>

<!-- Configuration -->
<menuitem id="menu_config" name="Configuration"
          parent="menu_root" sequence="90"
          groups="base.group_system"/>
```

---

## Best Practices

1. **Consistent naming** - Use clear, action-oriented names
2. **Logical grouping** - Group related menus together
3. **Sequence numbers** - Leave gaps (10, 20, 30) for future insertions
4. **Security groups** - Restrict sensitive menus
5. **Configuration last** - Always sequence 90+
6. **App icon** - Always provide for root menu
7. **Extend, don't duplicate** - Add to existing apps when appropriate
8. **Keep it shallow** - Max 3 levels of nesting
9. **Use filters** - Pre-filtered views for common use cases
10. **Test permissions** - Verify menu visibility for different users

---


## Source: dashboard-kpi-patterns.md

# Dashboard and KPI Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  DASHBOARD & KPI PATTERNS                                                    ║
║  Analytics views, KPI displays, and business intelligence                    ║
║  Use for data visualization and executive dashboards                         ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Dashboard Model (Pivot/Graph Base)

### Analytics Model for Reporting
```python
from odoo import api, fields, models, tools


class SaleAnalysis(models.Model):
    _name = 'sale.analysis'
    _description = 'Sales Analysis'
    _auto = False  # No table created - it's a database view
    _order = 'date desc'

    # Dimension fields
    date = fields.Date(string='Date', readonly=True)
    partner_id = fields.Many2one('res.partner', string='Customer', readonly=True)
    product_id = fields.Many2one('product.product', string='Product', readonly=True)
    categ_id = fields.Many2one('product.category', string='Category', readonly=True)
    user_id = fields.Many2one('res.users', string='Salesperson', readonly=True)
    company_id = fields.Many2one('res.company', string='Company', readonly=True)
    state = fields.Selection([
        ('draft', 'Draft'),
        ('sale', 'Confirmed'),
        ('done', 'Done'),
        ('cancel', 'Cancelled'),
    ], string='Status', readonly=True)

    # Measure fields
    order_count = fields.Integer(string='# Orders', readonly=True)
    product_qty = fields.Float(string='Qty Sold', readonly=True)
    price_subtotal = fields.Float(string='Untaxed Total', readonly=True)
    price_total = fields.Float(string='Total', readonly=True)

    def init(self):
        """Create database view for analysis."""
        tools.drop_view_if_exists(self.env.cr, self._table)
        self.env.cr.execute("""
            CREATE OR REPLACE VIEW %s AS (
                SELECT
                    row_number() OVER () as id,
                    so.date_order::date as date,
                    so.partner_id,
                    sol.product_id,
                    pt.categ_id,
                    so.user_id,
                    so.company_id,
                    so.state,
                    COUNT(DISTINCT so.id) as order_count,
                    SUM(sol.product_uom_qty) as product_qty,
                    SUM(sol.price_subtotal) as price_subtotal,
                    SUM(sol.price_total) as price_total
                FROM sale_order so
                JOIN sale_order_line sol ON sol.order_id = so.id
                JOIN product_product pp ON pp.id = sol.product_id
                JOIN product_template pt ON pt.id = pp.product_tmpl_id
                GROUP BY
                    so.date_order::date,
                    so.partner_id,
                    sol.product_id,
                    pt.categ_id,
                    so.user_id,
                    so.company_id,
                    so.state
            )
        """ % self._table)
```

---

## Dashboard Views

### Pivot View
```xml
<record id="sale_analysis_view_pivot" model="ir.ui.view">
    <field name="name">sale.analysis.pivot</field>
    <field name="model">sale.analysis</field>
    <field name="arch" type="xml">
        <pivot string="Sales Analysis" display_quantity="true">
            <field name="date" type="row" interval="month"/>
            <field name="categ_id" type="row"/>
            <field name="user_id" type="col"/>
            <field name="price_total" type="measure"/>
            <field name="product_qty" type="measure"/>
            <field name="order_count" type="measure"/>
        </pivot>
    </field>
</record>
```

### Graph View
```xml
<record id="sale_analysis_view_graph" model="ir.ui.view">
    <field name="name">sale.analysis.graph</field>
    <field name="model">sale.analysis</field>
    <field name="arch" type="xml">
        <graph string="Sales Analysis" type="bar" stacked="True">
            <field name="date" type="row" interval="month"/>
            <field name="price_total" type="measure"/>
        </graph>
    </field>
</record>

<!-- Line Chart -->
<record id="sale_analysis_view_graph_line" model="ir.ui.view">
    <field name="name">sale.analysis.graph.line</field>
    <field name="model">sale.analysis</field>
    <field name="arch" type="xml">
        <graph string="Sales Trend" type="line">
            <field name="date" type="row" interval="day"/>
            <field name="price_total" type="measure"/>
        </graph>
    </field>
</record>

<!-- Pie Chart -->
<record id="sale_analysis_view_graph_pie" model="ir.ui.view">
    <field name="name">sale.analysis.graph.pie</field>
    <field name="model">sale.analysis</field>
    <field name="arch" type="xml">
        <graph string="Sales by Category" type="pie">
            <field name="categ_id" type="row"/>
            <field name="price_total" type="measure"/>
        </graph>
    </field>
</record>
```

### Dashboard Action
```xml
<record id="action_sale_analysis" model="ir.actions.act_window">
    <field name="name">Sales Analysis</field>
    <field name="res_model">sale.analysis</field>
    <field name="view_mode">graph,pivot</field>
    <field name="context">{
        'search_default_current_month': 1,
        'group_by': ['date:month'],
    }</field>
    <field name="help" type="html">
        <p class="o_view_nocontent_smiling_face">
            No data to display
        </p>
    </field>
</record>
```

---

## KPI Stat Buttons

### Button Box Pattern
```python
class Partner(models.Model):
    _inherit = 'res.partner'

    # KPI counters
    sale_order_count = fields.Integer(
        compute='_compute_sale_count',
        string='Sales',
    )
    sale_total = fields.Monetary(
        compute='_compute_sale_count',
        string='Total Sales',
    )
    invoice_count = fields.Integer(
        compute='_compute_invoice_count',
        string='Invoices',
    )
    open_invoice_amount = fields.Monetary(
        compute='_compute_invoice_count',
        string='Due Amount',
    )

    def _compute_sale_count(self):
        for partner in self:
            orders = self.env['sale.order'].search([
                ('partner_id', '=', partner.id),
                ('state', 'in', ['sale', 'done']),
            ])
            partner.sale_order_count = len(orders)
            partner.sale_total = sum(orders.mapped('amount_total'))

    def _compute_invoice_count(self):
        for partner in self:
            invoices = self.env['account.move'].search([
                ('partner_id', '=', partner.id),
                ('move_type', '=', 'out_invoice'),
            ])
            partner.invoice_count = len(invoices)
            partner.open_invoice_amount = sum(
                inv.amount_residual
                for inv in invoices
                if inv.payment_state != 'paid'
            )

    def action_view_sales(self):
        """Open related sales."""
        return {
            'type': 'ir.actions.act_window',
            'name': 'Sales',
            'res_model': 'sale.order',
            'view_mode': 'tree,form',
            'domain': [('partner_id', '=', self.id)],
        }

    def action_view_invoices(self):
        """Open related invoices."""
        return {
            'type': 'ir.actions.act_window',
            'name': 'Invoices',
            'res_model': 'account.move',
            'view_mode': 'tree,form',
            'domain': [
                ('partner_id', '=', self.id),
                ('move_type', '=', 'out_invoice'),
            ],
        }
```

### Button Box View
```xml
<form>
    <sheet>
        <div class="oe_button_box" name="button_box">
            <!-- Sales stat button -->
            <button name="action_view_sales" type="object"
                    class="oe_stat_button" icon="fa-dollar">
                <div class="o_field_widget o_stat_info">
                    <span class="o_stat_value">
                        <field name="sale_order_count"/>
                    </span>
                    <span class="o_stat_text">Sales</span>
                </div>
                <div class="o_stat_info" invisible="not sale_total">
                    <span class="o_stat_value">
                        <field name="sale_total" widget="monetary"/>
                    </span>
                </div>
            </button>

            <!-- Invoice stat button -->
            <button name="action_view_invoices" type="object"
                    class="oe_stat_button" icon="fa-book"
                    invisible="invoice_count == 0">
                <div class="o_field_widget o_stat_info">
                    <span class="o_stat_value">
                        <field name="invoice_count"/>
                    </span>
                    <span class="o_stat_text">Invoices</span>
                </div>
            </button>

            <!-- Alert indicator -->
            <button name="action_view_open_invoices" type="object"
                    class="oe_stat_button" icon="fa-exclamation-triangle"
                    invisible="open_invoice_amount == 0">
                <div class="o_stat_info text-danger">
                    <span class="o_stat_value">
                        <field name="open_invoice_amount" widget="monetary"/>
                    </span>
                    <span class="o_stat_text">Due</span>
                </div>
            </button>
        </div>
    </sheet>
</form>
```

---

## OWL Dashboard Component

### Dashboard Action (v16+)
```javascript
/** @odoo-module **/
import { registry } from "@web/core/registry";
import { Component, useState, onWillStart } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";

class MyDashboard extends Component {
    static template = "my_module.Dashboard";

    setup() {
        this.orm = useService("orm");
        this.action = useService("action");

        this.state = useState({
            kpis: {},
            loading: true,
        });

        onWillStart(async () => {
            await this.loadKPIs();
        });
    }

    async loadKPIs() {
        this.state.loading = true;
        try {
            this.state.kpis = await this.orm.call(
                "my.dashboard",
                "get_dashboard_data",
                []
            );
        } finally {
            this.state.loading = false;
        }
    }

    openAction(action) {
        this.action.doAction(action);
    }
}

registry.category("actions").add("my_module.dashboard", MyDashboard);
```

### Dashboard Template
```xml
<?xml version="1.0" encoding="UTF-8"?>
<templates xml:space="preserve">
    <t t-name="my_module.Dashboard">
        <div class="o_my_dashboard container-fluid">
            <div class="row mt-4">
                <!-- KPI Cards -->
                <div class="col-lg-3 col-md-6 mb-4">
                    <div class="card bg-primary text-white h-100"
                         t-on-click="() => this.openAction('sale.action_orders')">
                        <div class="card-body">
                            <div class="d-flex justify-content-between align-items-center">
                                <div>
                                    <h6 class="text-white-50">Sales</h6>
                                    <h2 t-esc="state.kpis.sale_count || 0"/>
                                </div>
                                <i class="fa fa-shopping-cart fa-3x opacity-50"/>
                            </div>
                        </div>
                        <div class="card-footer bg-transparent border-0">
                            <small>This Month: <t t-esc="state.kpis.sale_amount || 0"/></small>
                        </div>
                    </div>
                </div>

                <div class="col-lg-3 col-md-6 mb-4">
                    <div class="card bg-success text-white h-100">
                        <div class="card-body">
                            <div class="d-flex justify-content-between align-items-center">
                                <div>
                                    <h6 class="text-white-50">Revenue</h6>
                                    <h2 t-esc="state.kpis.revenue || 0"/>
                                </div>
                                <i class="fa fa-dollar fa-3x opacity-50"/>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="col-lg-3 col-md-6 mb-4">
                    <div class="card bg-warning text-dark h-100">
                        <div class="card-body">
                            <div class="d-flex justify-content-between align-items-center">
                                <div>
                                    <h6>Pending</h6>
                                    <h2 t-esc="state.kpis.pending_count || 0"/>
                                </div>
                                <i class="fa fa-clock-o fa-3x opacity-50"/>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="col-lg-3 col-md-6 mb-4">
                    <div class="card bg-danger text-white h-100">
                        <div class="card-body">
                            <div class="d-flex justify-content-between align-items-center">
                                <div>
                                    <h6 class="text-white-50">Overdue</h6>
                                    <h2 t-esc="state.kpis.overdue_count || 0"/>
                                </div>
                                <i class="fa fa-exclamation-triangle fa-3x opacity-50"/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Chart Area -->
            <div class="row">
                <div class="col-lg-8 mb-4">
                    <div class="card">
                        <div class="card-header">
                            <h5>Sales Trend</h5>
                        </div>
                        <div class="card-body">
                            <!-- Embed graph view or custom chart -->
                            <div id="sales_chart" style="height: 300px;"/>
                        </div>
                    </div>
                </div>
                <div class="col-lg-4 mb-4">
                    <div class="card">
                        <div class="card-header">
                            <h5>Top Products</h5>
                        </div>
                        <div class="card-body">
                            <ul class="list-group list-group-flush">
                                <t t-foreach="state.kpis.top_products || []" t-as="product">
                                    <li class="list-group-item d-flex justify-content-between">
                                        <span t-esc="product.name"/>
                                        <span class="badge bg-primary rounded-pill"
                                              t-esc="product.count"/>
                                    </li>
                                </t>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </t>
</templates>
```

### Dashboard Data Provider
```python
class MyDashboard(models.TransientModel):
    _name = 'my.dashboard'
    _description = 'Dashboard Data Provider'

    @api.model
    def get_dashboard_data(self):
        """Return KPI data for dashboard."""
        today = fields.Date.today()
        month_start = today.replace(day=1)

        # Sales KPIs
        sales = self.env['sale.order'].search([
            ('date_order', '>=', month_start),
            ('state', 'in', ['sale', 'done']),
        ])

        # Pending orders
        pending = self.env['sale.order'].search_count([
            ('state', '=', 'draft'),
        ])

        # Overdue invoices
        overdue = self.env['account.move'].search_count([
            ('move_type', '=', 'out_invoice'),
            ('payment_state', '!=', 'paid'),
            ('invoice_date_due', '<', today),
        ])

        # Top products
        top_products = self._get_top_products(limit=5)

        return {
            'sale_count': len(sales),
            'sale_amount': sum(sales.mapped('amount_total')),
            'revenue': self._get_revenue(),
            'pending_count': pending,
            'overdue_count': overdue,
            'top_products': top_products,
        }

    def _get_top_products(self, limit=5):
        """Get top selling products."""
        query = """
            SELECT pp.id, pt.name, SUM(sol.product_uom_qty) as qty
            FROM sale_order_line sol
            JOIN product_product pp ON pp.id = sol.product_id
            JOIN product_template pt ON pt.id = pp.product_tmpl_id
            JOIN sale_order so ON so.id = sol.order_id
            WHERE so.state IN ('sale', 'done')
            AND so.date_order >= %s
            GROUP BY pp.id, pt.name
            ORDER BY qty DESC
            LIMIT %s
        """
        month_start = fields.Date.today().replace(day=1)
        self.env.cr.execute(query, (month_start, limit))

        return [
            {'id': row[0], 'name': row[1], 'count': int(row[2])}
            for row in self.env.cr.fetchall()
        ]

    def _get_revenue(self):
        """Calculate monthly revenue."""
        month_start = fields.Date.today().replace(day=1)
        invoices = self.env['account.move'].search([
            ('move_type', '=', 'out_invoice'),
            ('state', '=', 'posted'),
            ('invoice_date', '>=', month_start),
        ])
        return sum(invoices.mapped('amount_total'))
```

---

## Search Filters for Dashboard

### Search View with Defaults
```xml
<record id="sale_analysis_view_search" model="ir.ui.view">
    <field name="name">sale.analysis.search</field>
    <field name="model">sale.analysis</field>
    <field name="arch" type="xml">
        <search>
            <!-- Filters -->
            <filter name="current_month" string="This Month"
                    domain="[('date', '>=', (context_today() - relativedelta(day=1)).strftime('%Y-%m-%d'))]"/>
            <filter name="current_quarter" string="This Quarter"
                    domain="[('date', '>=', (context_today() - relativedelta(months=(context_today().month - 1) % 3, day=1)).strftime('%Y-%m-%d'))]"/>
            <filter name="current_year" string="This Year"
                    domain="[('date', '>=', (context_today()).strftime('%Y-01-01'))]"/>
            <separator/>
            <filter name="confirmed" string="Confirmed"
                    domain="[('state', '=', 'sale')]"/>

            <!-- Group By -->
            <group expand="1" string="Group By">
                <filter name="group_by_date" string="Date"
                        context="{'group_by': 'date:month'}"/>
                <filter name="group_by_partner" string="Customer"
                        context="{'group_by': 'partner_id'}"/>
                <filter name="group_by_product" string="Product"
                        context="{'group_by': 'product_id'}"/>
                <filter name="group_by_category" string="Category"
                        context="{'group_by': 'categ_id'}"/>
                <filter name="group_by_user" string="Salesperson"
                        context="{'group_by': 'user_id'}"/>
            </group>
        </search>
    </field>
</record>
```

---

## Cohort View (Enterprise)

### Cohort Analysis
```xml
<record id="sale_analysis_view_cohort" model="ir.ui.view">
    <field name="name">sale.analysis.cohort</field>
    <field name="model">sale.analysis</field>
    <field name="arch" type="xml">
        <cohort string="Sales Cohort"
                date_start="date"
                date_stop="date"
                interval="month"
                measure="price_total"/>
    </field>
</record>
```

---

## Best Practices

1. **Use database views** - _auto=False for aggregated models
2. **Index key columns** - Add indexes to frequently filtered fields
3. **Pre-aggregate** - Calculate totals in SQL, not Python
4. **Cache expensive** - Use Redis/Memcached for heavy queries
5. **Limit date ranges** - Default to current month/quarter
6. **Add search filters** - Make it easy to drill down
7. **Use measures wisely** - Choose meaningful KPIs
8. **Refresh async** - Use background jobs for heavy data
9. **Mobile friendly** - Design for responsive display
10. **Test performance** - Verify with production data volumes

---


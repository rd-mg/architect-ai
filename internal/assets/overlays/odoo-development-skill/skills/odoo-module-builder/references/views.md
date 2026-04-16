> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Views Reference

## Table of Contents
- [View Types](#view-types)
- [Form Views](#form-views)
- [Tree/List Views](#treelist-views)
- [Kanban Views](#kanban-views)
- [Search Views](#search-views)
- [Actions and Menus](#actions-and-menus)
- [Inherited Views](#inherited-views)

## View Types

| Type | Purpose |
|------|---------|
| `form` | Single record editing |
| `tree` | List/table of records |
| `kanban` | Card-based board |
| `search` | Filters and grouping |
| `calendar` | Calendar display |
| `pivot` | Pivot table analysis |
| `graph` | Charts |
| `activity` | Activity timeline |

## Form Views

```xml
<record id="view_library_book_form" model="ir.ui.view">
    <field name="name">library.book.form</field>
    <field name="model">library.book</field>
    <field name="arch" type="xml">
        <form string="Book">
            <header>
                <button name="action_borrow" string="Borrow"
                        type="object" class="oe_highlight"
                        invisible="state != 'available'"/>
                <button name="action_return" string="Return"
                        type="object"
                        invisible="state != 'borrowed'"/>
                <field name="state" widget="statusbar"
                       statusbar_visible="draft,available,borrowed"/>
            </header>
            <sheet>
                <div class="oe_button_box" name="button_box">
                    <button name="action_view_loans" type="object"
                            class="oe_stat_button" icon="fa-book">
                        <field name="loan_count" widget="statinfo"
                               string="Loans"/>
                    </button>
                </div>
                <widget name="web_ribbon" title="Archived"
                        bg_color="text-bg-danger"
                        invisible="active"/>
                <field name="image_1920" widget="image"
                       class="oe_avatar" options='{"preview_image": "image_128"}'/>
                <div class="oe_title">
                    <label for="name"/>
                    <h1><field name="name" placeholder="Book Title"/></h1>
                </div>
                <group>
                    <group>
                        <field name="isbn"/>
                        <field name="author_ids" widget="many2many_tags"/>
                        <field name="category_id"/>
                    </group>
                    <group>
                        <field name="date_published"/>
                        <field name="price"/>
                        <field name="currency_id" invisible="1"/>
                    </group>
                </group>
                <notebook>
                    <page string="Description" name="description">
                        <field name="description"/>
                    </page>
                    <page string="Loans" name="loans">
                        <field name="loan_ids">
                            <tree editable="bottom">
                                <field name="partner_id"/>
                                <field name="date_borrow"/>
                                <field name="date_return"/>
                            </tree>
                        </field>
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

### Key Form Elements

| Element | Purpose |
|---------|---------|
| `<header>` | Status bar and action buttons |
| `<sheet>` | Main content area |
| `<group>` | Two-column layout container |
| `<notebook>` / `<page>` | Tabbed sections |
| `<div class="oe_chatter">` | Mail thread (requires mail.thread mixin) |
| `<div class="oe_button_box">` | Stat buttons area |
| `<separator>` | Visual separator with optional label |

### Visibility (Odoo 17)

In Odoo 17, use `invisible` attribute with Python-like expressions (replaces `attrs`):

```xml
<field name="field_a" invisible="state != 'draft'"/>
<field name="field_b" readonly="state == 'done'"/>
<field name="field_c" required="type == 'service'"/>
```

### Visibility (Odoo 16)

In Odoo 16, use `attrs` dict:

```xml
<field name="field_a" attrs="{'invisible': [('state', '!=', 'draft')]}"/>
<field name="field_b" attrs="{'readonly': [('state', '=', 'done')]}"/>
<field name="field_c" attrs="{'required': [('type', '=', 'service')]}"/>
```

## Tree/List Views

```xml
<record id="view_library_book_tree" model="ir.ui.view">
    <field name="name">library.book.tree</field>
    <field name="model">library.book</field>
    <field name="arch" type="xml">
        <tree string="Books" decoration-danger="state == 'lost'"
              decoration-info="state == 'available'"
              multi_edit="1">
            <field name="name"/>
            <field name="isbn"/>
            <field name="author_ids" widget="many2many_tags"/>
            <field name="category_id"/>
            <field name="state"
                   decoration-success="state == 'available'"
                   decoration-warning="state == 'borrowed'"
                   decoration-danger="state == 'lost'"
                   widget="badge"/>
            <field name="price" sum="Total Price"/>
        </tree>
    </field>
</record>
```

### Tree Decorations

| Attribute | Color |
|-----------|-------|
| `decoration-success` | Green |
| `decoration-info` | Blue |
| `decoration-warning` | Orange |
| `decoration-danger` | Red |
| `decoration-muted` | Gray |

## Kanban Views

```xml
<record id="view_library_book_kanban" model="ir.ui.view">
    <field name="name">library.book.kanban</field>
    <field name="model">library.book</field>
    <field name="arch" type="xml">
        <kanban default_group_by="state" class="o_kanban_small_column">
            <field name="name"/>
            <field name="state"/>
            <field name="image_128"/>
            <field name="category_id"/>
            <templates>
                <t t-name="kanban-card">
                    <field name="image_128" widget="image"
                           class="o_kanban_image"
                           options='{"preview_image": "image_128"}'/>
                    <div class="oe_kanban_details">
                        <strong class="o_kanban_record_title">
                            <field name="name"/>
                        </strong>
                        <div class="o_kanban_tags_section">
                            <field name="category_id"/>
                        </div>
                    </div>
                </t>
            </templates>
        </kanban>
    </field>
</record>
```

## Search Views

```xml
<record id="view_library_book_search" model="ir.ui.view">
    <field name="name">library.book.search</field>
    <field name="model">library.book</field>
    <field name="arch" type="xml">
        <search string="Search Books">
            <field name="name" string="Title"
                   filter_domain="['|', ('name', 'ilike', self), ('isbn', 'ilike', self)]"/>
            <field name="author_ids"/>
            <field name="category_id"/>
            <separator/>
            <filter name="available" string="Available"
                    domain="[('state', '=', 'available')]"/>
            <filter name="borrowed" string="Borrowed"
                    domain="[('state', '=', 'borrowed')]"/>
            <separator/>
            <filter name="archived" string="Archived"
                    domain="[('active', '=', False)]"/>
            <group expand="0" string="Group By">
                <filter name="group_by_category" string="Category"
                        context="{'group_by': 'category_id'}"/>
                <filter name="group_by_state" string="State"
                        context="{'group_by': 'state'}"/>
            </group>
        </search>
    </field>
</record>
```

## Actions and Menus

### Window Action

```xml
<record id="action_library_book" model="ir.actions.act_window">
    <field name="name">Books</field>
    <field name="res_model">library.book</field>
    <field name="view_mode">tree,form,kanban</field>
    <field name="search_view_id" ref="view_library_book_search"/>
    <field name="context">{'search_default_available': 1}</field>
    <field name="help" type="html">
        <p class="o_view_nocontent_smiling_face">
            Create your first book!
        </p>
    </field>
</record>
```

### Menu Items

```xml
<!-- Top-level menu -->
<menuitem id="menu_library_root"
          name="Library"
          web_icon="library_management,static/description/icon.png"
          sequence="10"/>

<!-- Sub-menu -->
<menuitem id="menu_library_catalog"
          name="Catalog"
          parent="menu_library_root"
          sequence="10"/>

<!-- Action menu item -->
<menuitem id="menu_library_book"
          name="Books"
          parent="menu_library_catalog"
          action="action_library_book"
          sequence="10"/>
```

## Inherited Views

### Extending an Existing View

```xml
<record id="view_partner_form_inherit_library" model="ir.ui.view">
    <field name="name">res.partner.form.inherit.library</field>
    <field name="model">res.partner</field>
    <field name="inherit_id" ref="base.view_partner_form"/>
    <field name="arch" type="xml">
        <!-- Add field after existing field -->
        <field name="website" position="after">
            <field name="loyalty_points"/>
        </field>

        <!-- Add page to notebook -->
        <xpath expr="//notebook" position="inside">
            <page string="Library" name="library">
                <field name="borrowed_book_ids"/>
            </page>
        </xpath>

        <!-- Replace existing element -->
        <field name="phone" position="replace">
            <field name="phone" widget="phone"/>
        </field>

        <!-- Add attributes to existing field -->
        <field name="email" position="attributes">
            <attribute name="required">1</attribute>
        </field>
    </field>
</record>
```

### XPath Position Values

| Position | Effect |
|----------|--------|
| `after` | Insert after the matched element |
| `before` | Insert before the matched element |
| `inside` | Append inside the matched element (last child) |
| `replace` | Replace the matched element entirely |
| `attributes` | Modify attributes of the matched element |

#!/usr/bin/env python3
"""
Odoo Module Scaffold Generator

Creates a complete Odoo module directory structure with all boilerplate files.

Usage:
    scaffold_module.py <module_technical_name> --path <output_directory> [--odoo-version 17]

Examples:
    scaffold_module.py library_management --path /home/user/addons
    scaffold_module.py library_management --path ./addons --odoo-version 16
"""

import sys
import os
from pathlib import Path

DEFAULT_ODOO_VERSION = "17"


def scaffold_module(module_name, output_path, odoo_version="19"):
    """Create a complete Odoo module scaffold."""
    module_dir = Path(output_path).resolve() / module_name

    if module_dir.exists():
        print(f"Error: Directory already exists: {module_dir}")
        return False

    # Derive human-readable name
    human_name = module_name.replace('_', ' ').title()

    # Create directory structure
    dirs = [
        module_dir,
        module_dir / 'models',
        module_dir / 'views',
        module_dir / 'security',
        module_dir / 'data',
        module_dir / 'static' / 'description',
    ]
    for d in dirs:
        d.mkdir(parents=True, exist_ok=True)

    # Determine attrs style based on version
    use_invisible_expr = int(odoo_version) >= 17

    # __manifest__.py
    (module_dir / '__manifest__.py').write_text(f"""{'{'}
    'name': '{human_name}',
    'version': '{odoo_version}.0.1.0.0',
    'category': 'Uncategorized',
    'summary': 'TODO: Short module summary',
    'description': \"\"\"
        {human_name}
        {'=' * len(human_name)}
        TODO: Add module description
    \"\"\",
    'author': 'TODO: Author Name',
    'website': '',
    'license': 'LGPL-3',
    'depends': [
        'base',
        'mail',
    ],
    'data': [
        'security/security.xml',
        'security/ir.model.access.csv',
        'views/{module_name}_views.xml',
        'views/menu.xml',
    ],
    'demo': [],
    'installable': True,
    'application': True,
    'auto_install': False,
{'}'}
""")

    # Root __init__.py
    (module_dir / '__init__.py').write_text("from . import models\n")

    # models/__init__.py
    (module_dir / 'models' / '__init__.py').write_text(
        f"from . import {module_name}\n"
    )

    # models/<module_name>.py — example model
    model_name = module_name.replace('_', '.')
    model_class = ''.join(word.capitalize() for word in module_name.split('_'))
    (module_dir / 'models' / f'{module_name}.py').write_text(f"""from odoo import models, fields, api


class {model_class}(models.Model):
    _name = '{model_name}'
    _description = '{human_name}'
    _inherit = ['mail.thread', 'mail.activity.mixin']
    _order = 'name'

    name = fields.Char(string='Name', required=True, tracking=True)
    description = fields.Text(string='Description')
    active = fields.Boolean(default=True)
    state = fields.Selection([
        ('draft', 'Draft'),
        ('confirmed', 'Confirmed'),
        ('done', 'Done'),
        ('cancelled', 'Cancelled'),
    ], string='Status', default='draft', required=True, tracking=True)
    company_id = fields.Many2one(
        'res.company', string='Company',
        default=lambda self: self.env.company,
    )

    def action_confirm(self):
        for record in self:
            record.state = 'confirmed'

    def action_done(self):
        for record in self:
            record.state = 'done'

    def action_cancel(self):
        for record in self:
            record.state = 'cancelled'

    def action_draft(self):
        for record in self:
            record.state = 'draft'
""")

    # security/security.xml
    (module_dir / 'security' / 'security.xml').write_text(f"""<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="group_{module_name}_user" model="res.groups">
        <field name="name">{human_name} User</field>
        <field name="category_id" ref="base.module_category_services"/>
        <field name="implied_ids" eval="[(4, ref('base.group_user'))]"/>
    </record>

    <record id="group_{module_name}_manager" model="res.groups">
        <field name="name">{human_name} Manager</field>
        <field name="category_id" ref="base.module_category_services"/>
        <field name="implied_ids" eval="[(4, ref('group_{module_name}_user'))]"/>
        <field name="users" eval="[(4, ref('base.user_root')), (4, ref('base.user_admin'))]"/>
    </record>
</odoo>
""")

    # security/ir.model.access.csv
    model_id = f"model_{module_name.replace('.', '_')}"
    (module_dir / 'security' / 'ir.model.access.csv').write_text(
        f"""id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
access_{module_name}_user,{model_name}.user,{model_id},group_{module_name}_user,1,1,1,0
access_{module_name}_manager,{model_name}.manager,{model_id},group_{module_name}_manager,1,1,1,1
""")

    # views/<module_name>_views.xml
    if use_invisible_expr:
        confirm_invisible = 'invisible="state != \'draft\'"'
        done_invisible = 'invisible="state != \'confirmed\'"'
        cancel_invisible = 'invisible="state in (\'done\', \'cancelled\')"'
        draft_invisible = 'invisible="state not in (\'confirmed\', \'cancelled\')"'
    else:
        confirm_invisible = "attrs=\"{'invisible': [('state', '!=', 'draft')]}\""
        done_invisible = "attrs=\"{'invisible': [('state', '!=', 'confirmed')]}\""
        cancel_invisible = "attrs=\"{'invisible': [('state', 'in', ('done', 'cancelled'))]}\""
        draft_invisible = "attrs=\"{'invisible': [('state', 'not in', ('confirmed', 'cancelled'))]}\""

    (module_dir / 'views' / f'{module_name}_views.xml').write_text(f"""<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- Form View -->
    <record id="view_{module_name}_form" model="ir.ui.view">
        <field name="name">{model_name}.form</field>
        <field name="model">{model_name}</field>
        <field name="arch" type="xml">
            <form string="{human_name}">
                <header>
                    <button name="action_confirm" string="Confirm"
                            type="object" class="oe_highlight"
                            {confirm_invisible}/>
                    <button name="action_done" string="Done"
                            type="object" class="oe_highlight"
                            {done_invisible}/>
                    <button name="action_cancel" string="Cancel"
                            type="object"
                            {cancel_invisible}/>
                    <button name="action_draft" string="Reset to Draft"
                            type="object"
                            {draft_invisible}/>
                    <field name="state" widget="statusbar"
                           statusbar_visible="draft,confirmed,done"/>
                </header>
                <sheet>
                    <div class="oe_title">
                        <label for="name"/>
                        <h1><field name="name" placeholder="Name"/></h1>
                    </div>
                    <group>
                        <group>
                            <field name="company_id"
                                   groups="base.group_multi_company"/>
                        </group>
                        <group>
                        </group>
                    </group>
                    <notebook>
                        <page string="Description" name="description">
                            <field name="description"/>
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

    <!-- Tree View -->
    <record id="view_{module_name}_tree" model="ir.ui.view">
        <field name="name">{model_name}.tree</field>
        <field name="model">{model_name}</field>
        <field name="arch" type="xml">
            <tree string="{human_name}" multi_edit="1">
                <field name="name"/>
                <field name="state" widget="badge"
                       decoration-success="state == 'done'"
                       decoration-info="state == 'confirmed'"
                       decoration-warning="state == 'draft'"/>
            </tree>
        </field>
    </record>

    <!-- Search View -->
    <record id="view_{module_name}_search" model="ir.ui.view">
        <field name="name">{model_name}.search</field>
        <field name="model">{model_name}</field>
        <field name="arch" type="xml">
            <search string="Search {human_name}">
                <field name="name"/>
                <separator/>
                <filter name="draft" string="Draft"
                        domain="[('state', '=', 'draft')]"/>
                <filter name="confirmed" string="Confirmed"
                        domain="[('state', '=', 'confirmed')]"/>
                <filter name="done" string="Done"
                        domain="[('state', '=', 'done')]"/>
                <separator/>
                <filter name="archived" string="Archived"
                        domain="[('active', '=', False)]"/>
                <group expand="0" string="Group By">
                    <filter name="group_by_state" string="Status"
                            context="{'{'}' group_by': 'state'{'}'}"/>
                </group>
            </search>
        </field>
    </record>

    <!-- Action -->
    <record id="action_{module_name}" model="ir.actions.act_window">
        <field name="name">{human_name}</field>
        <field name="res_model">{model_name}</field>
        <field name="view_mode">tree,form</field>
        <field name="search_view_id" ref="view_{module_name}_search"/>
        <field name="help" type="html">
            <p class="o_view_nocontent_smiling_face">
                Create your first {human_name.lower()} record!
            </p>
        </field>
    </record>
</odoo>
""")

    # views/menu.xml
    (module_dir / 'views' / 'menu.xml').write_text(f"""<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <menuitem id="menu_{module_name}_root"
              name="{human_name}"
              web_icon="{module_name},static/description/icon.png"
              sequence="100"/>

    <menuitem id="menu_{module_name}_main"
              name="{human_name}"
              parent="menu_{module_name}_root"
              sequence="10"/>

    <menuitem id="menu_{module_name}_list"
              name="{human_name}"
              parent="menu_{module_name}_main"
              action="action_{module_name}"
              sequence="10"/>
</odoo>
""")

    print(f"Module '{module_name}' scaffolded at {module_dir}")
    print(f"  Odoo version: {odoo_version}")
    print(f"  Model: {model_name}")
    print(f"  Visibility style: {'expressions' if use_invisible_expr else 'attrs'}")
    print("\nCreated files:")
    for f in sorted(module_dir.rglob('*')):
        if f.is_file():
            print(f"  {f.relative_to(module_dir)}")

    return True


def main():
    if len(sys.argv) < 4 or sys.argv[2] != '--path':
        print(__doc__)
        sys.exit(1)

    module_name = sys.argv[1]
    output_path = sys.argv[3]
    odoo_version = DEFAULT_ODOO_VERSION

    # Parse optional --odoo-version
    if '--odoo-version' in sys.argv:
        idx = sys.argv.index('--odoo-version')
        if idx + 1 < len(sys.argv):
            odoo_version = sys.argv[idx + 1]

    # Validate module name
    if not module_name.replace('_', '').isalnum() or module_name[0].isdigit():
        print(f"Error: Invalid module name '{module_name}'. Use snake_case (letters, digits, underscores).")
        sys.exit(1)

    if scaffold_module(module_name, output_path, odoo_version):
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == "__main__":
    main()

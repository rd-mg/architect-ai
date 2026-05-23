<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="view_{{ MODULE_NAME }}_form" model="ir.ui.view">
        <field name="name">{{ MODEL_NAME }}.form</field>
        <field name="model">{{ MODEL_NAME }}</field>
        <field name="arch" type="xml">
            <form string="{{ MODULE_DESC }}">
                <header>
                    <button name="action_done" string="Mark as Done" type="object" class="oe_highlight" invisible="state == 'done'"/>
                    <field name="state" widget="statusbar"/>
                </header>
                <sheet>
                    <group>
                        <field name="name"/>
                    </group>
                </sheet>
            </form>
        </field>
    </record>

    <record id="view_{{ MODULE_NAME }}_tree" model="ir.ui.view">
        <field name="name">{{ MODEL_NAME }}.tree</field>
        <field name="model">{{ MODEL_NAME }}</field>
        <field name="arch" type="xml">
            <tree string="{{ MODULE_DESC }}">
                <field name="name"/>
                <field name="state"/>
            </tree>
        </field>
    </record>

    <record id="action_{{ MODULE_NAME }}" model="ir.actions.act_window">
        <field name="name">{{ MODULE_DESC }}</field>
        <field name="res_model">{{ MODEL_NAME }}</field>
        <field name="view_mode">tree,form</field>
    </record>

    <menuitem id="menu_{{ MODULE_NAME }}_root" name="{{ MODULE_DESC }}" sequence="10"/>
    <menuitem id="menu_{{ MODULE_NAME }}_list" name="{{ MODULE_DESC }}" parent="menu_{{ MODULE_NAME }}_root" action="action_{{ MODULE_NAME }}"/>
</odoo>

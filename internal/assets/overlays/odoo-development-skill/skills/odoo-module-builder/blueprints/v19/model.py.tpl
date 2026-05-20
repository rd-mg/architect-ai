from odoo import models, fields, api
from odoo.tools import SQL

class {{ MODEL_CLASS }}(models.Model):
    _name = '{{ MODEL_NAME }}'
    _description = '{{ MODULE_DESC }}'
    _inherit = ['mail.thread']

    name: str = fields.Char(string='Name', required=True, tracking=True)
    active: bool = fields.Boolean(default=True)
    state: str = fields.Selection([
        ('draft', 'Draft'),
        ('done', 'Done'),
    ], string='Status', default='draft', tracking=True)

    def action_done(self) -> None:
        # Using Odoo 19 SQL builder example
        query = SQL("UPDATE %I SET state = 'done' WHERE id IN %s", self._table, tuple(self.ids))
        self.env.cr.execute(query)
        self.invalidate_recordset(['state'])

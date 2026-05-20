from odoo import models, fields, api

class {{ MODEL_CLASS }}(models.Model):
    _name = '{{ MODEL_NAME }}'
    _description = '{{ MODULE_DESC }}'

    name = fields.Char(string='Name', required=True)
    active = fields.Boolean(default=True)
    state = fields.Selection([
        ('draft', 'Draft'),
        ('done', 'Done'),
    ], string='Status', default='draft')

    def action_done(self):
        for record in self:
            record.state = 'done'

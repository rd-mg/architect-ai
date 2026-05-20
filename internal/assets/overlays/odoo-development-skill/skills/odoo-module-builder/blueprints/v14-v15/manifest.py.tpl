{
    'name': '{{ MODULE_NAME }}',
    'version': '14.0.1.0.0',
    'category': 'Uncategorized',
    'summary': 'Scaffolded Odoo v14-v15 Module',
    'author': 'Architect-AI',
    'website': 'https://github.com/rd-mg/architect-ai',
    'license': 'LGPL-3',
    'depends': ['base'],
    'data': [
        'security/ir.model.access.csv',
        'views/{{ MODULE_NAME }}_views.xml',
    ],
    'installable': True,
    'application': False,
    'auto_install': False,
}

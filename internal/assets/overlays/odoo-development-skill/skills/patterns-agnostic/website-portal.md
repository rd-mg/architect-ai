# Website & Portal Patterns

Consolidated from the following source files:
- `portal-user-patterns.md` (architect-fix)
- `public-route-patterns.md` (architect-fix)
- `website-controller-patterns.md` (architect-fix)
- `portal-mixin-patterns.md` (architect-fix)

> **Version-specific syntax** → `patterns-{version}/model-patterns.md`
> `http.route()` API stable across v14-v19 · OWL widget integration changes per version

---

## Portal & Mixins

### Making a Model Portal-Aware
```python
class MyDocument(models.Model):
    _name = 'my.document'
    _inherit = ['portal.mixin', 'mail.thread']

    def _compute_access_url(self):
        super()._compute_access_url()
        for doc in self:
            doc.access_url = f'/my/document/{doc.id}'
```

### Granting Portal Access
```python
def grant_access(self, partner):
    # Standard Odoo portal wizard flow
    wizard = self.env['portal.wizard'].create({})
    self.env['portal.wizard.user'].create({
        'wizard_id': wizard.id,
        'partner_id': partner.id,
        'email': partner.email,
        'in_portal': True,
    })
    wizard.action_apply()
```

---

## Controllers & Routes

### Portal Controller (Inheritance)
```python
from odoo.addons.portal.controllers.portal import CustomerPortal

class MyPortal(CustomerPortal):
    def _prepare_home_portal_values(self, counters):
        values = super()._prepare_home_portal_values(counters)
        if 'my_count' in counters:
            values['my_count'] = request.env['my.model'].search_count([])
        return values

    @http.route('/my/docs', type='http', auth='user', website=True)
    def list_docs(self, **kw):
        docs = request.env['my.model'].search([('partner_id', '=', request.env.user.partner_id.id)])
        return request.render('my_mod.portal_list', {'docs': docs})
```

### Public API / JSON Endpoint
```python
@http.route('/api/v1/data', type='json', auth='public', methods=['POST'], csrf=False)
def public_api(self, **kwargs):
    # MUST use sudo() for public auth
    data = request.env['my.model'].sudo().search_read([], ['name', 'value'])
    return {'status': 'success', 'data': data}
```

---

## Anti-Patterns

```python
# ❌ NEVER use auth='none' for customer data routes — use auth='user'.

# ❌ NEVER forget sudo() in public/website routes — context has no user by default.

# ❌ NEVER return None in type='json' routes — always return a dict.

# ❌ NEVER disable csrf=False on HTML forms — use type='json' for AJAX instead.
```

---

## Version Matrix

| Feature | v14-v16 | v17 | v18 | v19 |
|---------|---------|-----|-----|-----|
| Routing | `http.route` | `http.route` | `http.route` | `http.route` |
| CSRF | Default On | Default On | Default On | Default On |
| Pager | `portal_pager`| `portal_pager`| `portal_pager`| `portal_pager`|
| OWL in Portal| Legacy | OWL 2 | OWL 2 | OWL 3 |

# Website Portal Patterns

Consolidated from the following source files:
- `website-integration-patterns.md`
- `portal-access-patterns.md`

---


## Source: website-integration-patterns.md

# Website Integration Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  WEBSITE INTEGRATION PATTERNS                                                ║
║  Website pages, snippets, themes, and ecommerce integration                  ║
║  Use for customer-facing pages, portals, and web shops                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Module Structure

```
my_website_module/
├── controllers/
│   ├── __init__.py
│   └── main.py
├── data/
│   └── website_data.xml
├── models/
│   ├── __init__.py
│   └── website.py
├── static/
│   └── src/
│       ├── js/
│       │   └── my_script.js
│       ├── scss/
│       │   └── my_styles.scss
│       └── img/
│           └── banner.png
├── views/
│   ├── templates.xml
│   ├── snippets.xml
│   └── pages.xml
└── __manifest__.py
```

### Manifest
```python
{
    'name': 'My Website Module',
    'version': '18.0.1.0.0',
    'depends': ['website'],
    'data': [
        'views/templates.xml',
        'views/pages.xml',
        'views/snippets.xml',
        'data/website_data.xml',
    ],
    'assets': {
        'web.assets_frontend': [
            'my_website_module/static/src/scss/my_styles.scss',
            'my_website_module/static/src/js/my_script.js',
        ],
    },
}
```

---

## Website Controllers

### Basic Page Controller
```python
from odoo import http
from odoo.http import request


class MyWebsiteController(http.Controller):

    @http.route('/my-page', type='http', auth='public', website=True)
    def my_page(self, **kw):
        """Render custom page."""
        values = {
            'page_title': 'My Custom Page',
            'records': request.env['my.model'].sudo().search([]),
        }
        return request.render('my_module.my_page_template', values)

    @http.route('/my-page/<int:id>', type='http', auth='public', website=True)
    def my_page_detail(self, id, **kw):
        """Render detail page."""
        record = request.env['my.model'].sudo().browse(id)
        if not record.exists():
            raise request.not_found()

        values = {
            'record': record,
        }
        return request.render('my_module.my_page_detail', values)
```

### Form Submission
```python
@http.route('/my-form/submit', type='http', auth='public',
            methods=['POST'], website=True, csrf=True)
def submit_form(self, **post):
    """Handle form submission."""
    # Validate input
    name = post.get('name', '').strip()
    email = post.get('email', '').strip()

    if not name or not email:
        return request.redirect('/my-form?error=missing_fields')

    # Create record
    request.env['my.model'].sudo().create({
        'name': name,
        'email': email,
    })

    return request.redirect('/my-form/thank-you')
```

### AJAX Endpoint
```python
@http.route('/my-api/search', type='json', auth='public', website=True)
def ajax_search(self, query='', limit=10):
    """AJAX search endpoint."""
    records = request.env['my.model'].sudo().search([
        ('name', 'ilike', query),
        ('website_published', '=', True),
    ], limit=limit)

    return [{
        'id': r.id,
        'name': r.name,
        'url': f'/my-page/{r.id}',
    } for r in records]
```

---

## QWeb Templates

### Basic Page Template
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <template id="my_page_template" name="My Page">
        <t t-call="website.layout">
            <div id="wrap" class="oe_structure">
                <section class="container py-5">
                    <h1 t-esc="page_title"/>

                    <div class="row">
                        <t t-foreach="records" t-as="record">
                            <div class="col-md-4 mb-4">
                                <div class="card h-100">
                                    <div class="card-body">
                                        <h5 class="card-title" t-field="record.name"/>
                                        <p class="card-text" t-field="record.description"/>
                                        <a t-attf-href="/my-page/#{record.id}"
                                           class="btn btn-primary">
                                            View Details
                                        </a>
                                    </div>
                                </div>
                            </div>
                        </t>
                    </div>
                </section>
            </div>
        </t>
    </template>
</odoo>
```

### Detail Page Template
```xml
<template id="my_page_detail" name="My Page Detail">
    <t t-call="website.layout">
        <div id="wrap">
            <section class="container py-5">
                <nav aria-label="breadcrumb">
                    <ol class="breadcrumb">
                        <li class="breadcrumb-item">
                            <a href="/">Home</a>
                        </li>
                        <li class="breadcrumb-item">
                            <a href="/my-page">My Page</a>
                        </li>
                        <li class="breadcrumb-item active" t-esc="record.name"/>
                    </ol>
                </nav>

                <div class="row">
                    <div class="col-lg-8">
                        <h1 t-field="record.name"/>
                        <div t-field="record.description"
                             class="lead"/>

                        <div t-if="record.image"
                             class="my-4">
                            <img t-att-src="image_data_uri(record.image)"
                                 class="img-fluid rounded"
                                 t-att-alt="record.name"/>
                        </div>

                        <div t-field="record.content"
                             class="mt-4"/>
                    </div>

                    <div class="col-lg-4">
                        <div class="card">
                            <div class="card-body">
                                <h5 class="card-title">Details</h5>
                                <ul class="list-unstyled">
                                    <li>
                                        <strong>Date:</strong>
                                        <span t-field="record.date"/>
                                    </li>
                                    <li t-if="record.category_id">
                                        <strong>Category:</strong>
                                        <span t-field="record.category_id.name"/>
                                    </li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    </t>
</template>
```

### Form Template
```xml
<template id="my_form_template" name="Contact Form">
    <t t-call="website.layout">
        <div id="wrap">
            <section class="container py-5">
                <h1>Contact Us</h1>

                <t t-if="error">
                    <div class="alert alert-danger">
                        Please fill in all required fields.
                    </div>
                </t>

                <form action="/my-form/submit" method="post"
                      class="s_website_form">
                    <input type="hidden" name="csrf_token"
                           t-att-value="request.csrf_token()"/>

                    <div class="mb-3">
                        <label for="name" class="form-label">Name *</label>
                        <input type="text" name="name" id="name"
                               class="form-control" required="required"/>
                    </div>

                    <div class="mb-3">
                        <label for="email" class="form-label">Email *</label>
                        <input type="email" name="email" id="email"
                               class="form-control" required="required"/>
                    </div>

                    <div class="mb-3">
                        <label for="message" class="form-label">Message</label>
                        <textarea name="message" id="message"
                                  class="form-control" rows="5"/>
                    </div>

                    <button type="submit" class="btn btn-primary">
                        Send Message
                    </button>
                </form>
            </section>
        </div>
    </t>
</template>
```

---

## Website Snippets

### Snippet Definition
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- Snippet Template -->
    <template id="s_my_custom_snippet" name="My Custom Snippet">
        <section class="s_my_custom_snippet pt-5 pb-5">
            <div class="container">
                <div class="row">
                    <div class="col-lg-6">
                        <h2 class="h1-fs">Your Heading Here</h2>
                        <p class="lead">
                            Your description text goes here.
                        </p>
                        <a href="#" class="btn btn-primary">
                            Learn More
                        </a>
                    </div>
                    <div class="col-lg-6">
                        <img src="/web/image/website.library_image_08"
                             class="img-fluid rounded"
                             alt="Feature image"/>
                    </div>
                </div>
            </div>
        </section>
    </template>

    <!-- Register in Snippet Menu -->
    <template id="s_my_custom_snippet_register"
              inherit_id="website.snippets"
              name="My Custom Snippet Register">
        <xpath expr="//snippets[@id='snippet_structure']" position="inside">
            <t t-snippet="my_module.s_my_custom_snippet"
               t-thumbnail="/my_module/static/src/img/snippet_thumb.png"/>
        </xpath>
    </template>

    <!-- Snippet Options -->
    <template id="s_my_custom_snippet_options"
              inherit_id="website.snippet_options">
        <xpath expr="." position="inside">
            <div data-selector=".s_my_custom_snippet">
                <we-select string="Layout">
                    <we-button data-select-class="">Default</we-button>
                    <we-button data-select-class="s_my_snippet_reverse">
                        Reversed
                    </we-button>
                </we-select>
            </div>
        </xpath>
    </template>
</odoo>
```

### Dynamic Snippet
```xml
<!-- Snippet with dynamic content -->
<template id="s_dynamic_products" name="Featured Products">
    <section class="s_dynamic_products pt-5 pb-5">
        <div class="container">
            <h2 class="text-center mb-4">Featured Products</h2>
            <div class="row" t-ignore="true">
                <!-- Content populated by controller -->
            </div>
        </div>
    </section>
</template>

<!-- Snippet content rendering -->
<template id="s_dynamic_products_content" name="Products Content">
    <t t-foreach="products" t-as="product">
        <div class="col-md-3 mb-4">
            <div class="card h-100">
                <t t-if="product.image_1920">
                    <img t-att-src="'/web/image/product.template/%s/image_256' % product.id"
                         class="card-img-top"
                         t-att-alt="product.name"/>
                </t>
                <div class="card-body">
                    <h5 class="card-title" t-field="product.name"/>
                    <p class="card-text">
                        <span t-field="product.list_price"
                              t-options='{"widget": "monetary",
                                         "display_currency": product.currency_id}'/>
                    </p>
                    <a t-att-href="product.website_url"
                       class="btn btn-primary">
                        View
                    </a>
                </div>
            </div>
        </div>
    </t>
</template>
```

### Snippet Controller
```python
@http.route('/snippet/featured_products', type='json', auth='public', website=True)
def get_featured_products(self, limit=4):
    """Get featured products for snippet."""
    products = request.env['product.template'].sudo().search([
        ('website_published', '=', True),
        ('is_featured', '=', True),
    ], limit=limit)

    return request.env['ir.qweb']._render(
        'my_module.s_dynamic_products_content',
        {'products': products}
    )
```

---

## Website Mixin

### Add Website Fields to Model
```python
class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'
    _inherit = ['website.published.mixin', 'website.seo.metadata']

    name = fields.Char(string='Name', required=True)
    description = fields.Html(string='Description')
    website_id = fields.Many2one(
        comodel_name='website',
        string='Website',
        ondelete='restrict',
    )

    def _compute_website_url(self):
        """Compute public URL."""
        for record in self:
            record.website_url = f'/my-page/{record.id}'
```

### Website Published Mixin Fields
```python
# Inherited from website.published.mixin:
# - website_published (Boolean)
# - website_url (Char, computed)
# - is_published (Boolean)

# Inherited from website.seo.metadata:
# - website_meta_title
# - website_meta_description
# - website_meta_keywords
# - website_meta_og_img
```

---

## Portal Integration

### Portal Controller
```python
from odoo.addons.portal.controllers.portal import CustomerPortal


class MyPortalController(CustomerPortal):

    def _prepare_home_portal_values(self, counters):
        """Add counter to portal home."""
        values = super()._prepare_home_portal_values(counters)

        if 'my_model_count' in counters:
            partner = request.env.user.partner_id
            values['my_model_count'] = request.env['my.model'].search_count([
                ('partner_id', '=', partner.id),
            ])

        return values

    @http.route('/my/records', type='http', auth='user', website=True)
    def portal_my_records(self, page=1, sortby=None, **kw):
        """Portal page listing user's records."""
        partner = request.env.user.partner_id

        domain = [('partner_id', '=', partner.id)]
        records = request.env['my.model'].search(domain)

        values = {
            'records': records,
            'page_name': 'my_records',
        }
        return request.render('my_module.portal_my_records', values)

    @http.route('/my/record/<int:record_id>', type='http', auth='user', website=True)
    def portal_record_detail(self, record_id, **kw):
        """Portal detail page."""
        record = request.env['my.model'].browse(record_id)

        # Security check
        if record.partner_id != request.env.user.partner_id:
            raise request.not_found()

        values = {
            'record': record,
            'page_name': 'my_record_detail',
        }
        return request.render('my_module.portal_record_detail', values)
```

### Portal Templates
```xml
<!-- Portal Home Counter -->
<template id="portal_my_home_counter"
          inherit_id="portal.portal_my_home">
    <xpath expr="//div[hasclass('o_portal_docs')]" position="inside">
        <t t-call="portal.portal_docs_entry">
            <t t-set="icon" t-value="'fa fa-file-text'"/>
            <t t-set="title">My Records</t>
            <t t-set="url" t-value="'/my/records'"/>
            <t t-set="text">View your records</t>
            <t t-set="count" t-value="my_model_count"/>
        </t>
    </xpath>
</template>

<!-- Portal Records List -->
<template id="portal_my_records" name="My Records">
    <t t-call="portal.portal_layout">
        <t t-set="breadcrumbs_searchbar" t-value="True"/>

        <t t-call="portal.portal_searchbar">
            <t t-set="title">My Records</t>
        </t>

        <t t-if="records">
            <table class="table table-hover">
                <thead>
                    <tr>
                        <th>Reference</th>
                        <th>Date</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    <t t-foreach="records" t-as="record">
                        <tr>
                            <td>
                                <a t-attf-href="/my/record/#{record.id}">
                                    <span t-field="record.name"/>
                                </a>
                            </td>
                            <td><span t-field="record.date"/></td>
                            <td>
                                <span t-field="record.state"
                                      class="badge rounded-pill"/>
                            </td>
                        </tr>
                    </t>
                </tbody>
            </table>
        </t>
        <t t-else="">
            <div class="alert alert-info">
                No records found.
            </div>
        </t>
    </t>
</template>
```

---

## Frontend JavaScript

### Basic Script
```javascript
/** @odoo-module **/

import publicWidget from "@web/legacy/js/public/public_widget";

publicWidget.registry.MyWidget = publicWidget.Widget.extend({
    selector: '.s_my_custom_snippet',
    events: {
        'click .btn-action': '_onClickAction',
    },

    start() {
        this._super(...arguments);
        this._loadData();
    },

    async _loadData() {
        const result = await this._rpc({
            route: '/my-api/search',
            params: { query: '', limit: 10 },
        });
        this._renderResults(result);
    },

    _onClickAction(ev) {
        ev.preventDefault();
        // Handle click
    },

    _renderResults(data) {
        // Render data to DOM
    },
});

export default publicWidget.registry.MyWidget;
```

### OWL Frontend Component (v16+)
```javascript
/** @odoo-module **/

import { Component, useState, onMounted } from "@odoo/owl";
import { registry } from "@web/core/registry";
import { useService } from "@web/core/utils/hooks";

class MyPublicComponent extends Component {
    static template = "my_module.MyPublicComponent";

    setup() {
        this.rpc = useService("rpc");
        this.state = useState({
            items: [],
            loading: true,
        });

        onMounted(() => this.loadItems());
    }

    async loadItems() {
        const result = await this.rpc("/my-api/search", {
            query: "",
            limit: 10,
        });
        this.state.items = result;
        this.state.loading = false;
    }
}

registry
    .category("public_components")
    .add("my_module.MyPublicComponent", MyPublicComponent);
```

---

## Best Practices

1. **Use `website=True`** for public-facing routes
2. **Use `sudo()`** carefully for public access
3. **Check permissions** in portal pages
4. **Use CSRF tokens** for all forms
5. **Optimize queries** for public pages
6. **Cache heavy computations**
7. **Use CDN** for static assets
8. **Test multi-website** if applicable
9. **SEO metadata** for all public pages
10. **Mobile responsive** design

---


## Source: portal-access-patterns.md

# Portal and Access Token Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  PORTAL & ACCESS TOKEN PATTERNS                                              ║
║  Customer portal, public access, and secure token-based sharing              ║
║  Use for external user access and document sharing                           ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Portal Mixin

### Basic Portal Access
```python
from odoo import api, fields, models


class MyDocument(models.Model):
    _name = 'my.document'
    _inherit = ['portal.mixin', 'mail.thread']

    name = fields.Char(required=True)
    partner_id = fields.Many2one('res.partner')

    # Portal mixin provides:
    # - access_token: Char field for secure access
    # - access_url: Computed URL for portal access
    # - access_warning: Warning message for portal

    def _compute_access_url(self):
        """Override to set portal URL."""
        super()._compute_access_url()
        for record in self:
            record.access_url = f'/my/documents/{record.id}'
```

### Access Token Generation
```python
import uuid
from odoo import api, fields, models


class MyDocument(models.Model):
    _name = 'my.document'
    _inherit = 'portal.mixin'

    # Token is auto-generated by portal.mixin
    # But you can customize generation:

    def _get_access_token(self):
        """Generate new access token."""
        return str(uuid.uuid4())

    def action_regenerate_token(self):
        """Regenerate access token for security."""
        for record in self:
            record.access_token = record._get_access_token()
```

---

## Portal Controller

### Basic Portal Page
```python
from odoo import http
from odoo.http import request
from odoo.addons.portal.controllers.portal import CustomerPortal


class MyPortal(CustomerPortal):

    def _prepare_home_portal_values(self, counters):
        """Add document count to portal home."""
        values = super()._prepare_home_portal_values(counters)
        if 'document_count' in counters:
            values['document_count'] = request.env['my.document'].search_count([
                ('partner_id', '=', request.env.user.partner_id.id),
            ])
        return values

    @http.route(['/my/documents', '/my/documents/page/<int:page>'],
                type='http', auth='user', website=True)
    def portal_my_documents(self, page=1, sortby=None, **kw):
        """Portal documents list."""
        Document = request.env['my.document']
        partner = request.env.user.partner_id

        domain = [('partner_id', '=', partner.id)]

        # Sorting
        searchbar_sortings = {
            'date': {'label': 'Date', 'order': 'create_date desc'},
            'name': {'label': 'Name', 'order': 'name'},
        }
        sortby = sortby or 'date'
        order = searchbar_sortings[sortby]['order']

        # Count and pagination
        document_count = Document.search_count(domain)
        pager = request.website.pager(
            url='/my/documents',
            total=document_count,
            page=page,
            step=10,
        )

        documents = Document.search(
            domain,
            order=order,
            limit=10,
            offset=pager['offset'],
        )

        return request.render('my_module.portal_my_documents', {
            'documents': documents,
            'page_name': 'documents',
            'pager': pager,
            'default_url': '/my/documents',
            'searchbar_sortings': searchbar_sortings,
            'sortby': sortby,
        })

    @http.route(['/my/documents/<int:document_id>'],
                type='http', auth='public', website=True)
    def portal_document_detail(self, document_id, access_token=None, **kw):
        """Single document detail page."""
        try:
            document_sudo = self._document_check_access(
                'my.document', document_id, access_token
            )
        except AccessError:
            return request.redirect('/my')

        return request.render('my_module.portal_document_detail', {
            'document': document_sudo,
            'page_name': 'document',
        })

    def _document_check_access(self, model, document_id, access_token=None):
        """Check access to document."""
        document = request.env[model].browse(document_id)

        if access_token:
            # Public access with token
            document_sudo = document.sudo()
            if not document_sudo.access_token or \
               document_sudo.access_token != access_token:
                raise AccessError("Invalid token")
            return document_sudo
        else:
            # Authenticated user access
            document.check_access_rights('read')
            document.check_access_rule('read')
            return document
```

---

## Portal Templates

### Document List Template
```xml
<template id="portal_my_documents" name="My Documents">
    <t t-call="portal.portal_layout">
        <t t-set="breadcrumbs_searchbar" t-value="True"/>
        <t t-call="portal.portal_searchbar">
            <t t-set="title">My Documents</t>
        </t>
        <t t-if="documents">
            <t t-call="portal.portal_table">
                <thead>
                    <tr>
                        <th>Reference</th>
                        <th>Date</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    <t t-foreach="documents" t-as="document">
                        <tr>
                            <td>
                                <a t-attf-href="/my/documents/#{document.id}?{{ keep_query() }}">
                                    <t t-esc="document.name"/>
                                </a>
                            </td>
                            <td>
                                <span t-field="document.create_date" t-options="{'widget': 'date'}"/>
                            </td>
                            <td>
                                <span t-attf-class="badge #{document.state == 'done' and 'bg-success' or 'bg-secondary'}">
                                    <t t-esc="document.state"/>
                                </span>
                            </td>
                        </tr>
                    </t>
                </tbody>
            </t>
        </t>
        <t t-else="">
            <div class="alert alert-info">
                No documents found.
            </div>
        </t>
    </t>
</template>
```

### Document Detail Template
```xml
<template id="portal_document_detail" name="Document Detail">
    <t t-call="portal.portal_layout">
        <t t-set="o_portal_fullwidth_alert" groups="my_module.group_manager">
            <t t-call="portal.portal_back_in_edit_mode">
                <t t-set="backend_url"
                   t-value="'/web#model=my.document&amp;id=%d&amp;view_type=form' % document.id"/>
            </t>
        </t>

        <t t-call="portal.portal_record_layout">
            <t t-set="card_header">
                <h3 class="mb-0">
                    <span t-field="document.name"/>
                    <span t-attf-class="badge #{document.state == 'done' and 'bg-success' or 'bg-info'}">
                        <t t-field="document.state"/>
                    </span>
                </h3>
            </t>

            <t t-set="card_body">
                <div class="row">
                    <div class="col-lg-6">
                        <strong>Partner:</strong>
                        <span t-field="document.partner_id"/>
                    </div>
                    <div class="col-lg-6">
                        <strong>Date:</strong>
                        <span t-field="document.create_date"/>
                    </div>
                </div>
                <div class="row mt-3">
                    <div class="col-12">
                        <strong>Description:</strong>
                        <div t-field="document.description"/>
                    </div>
                </div>
            </t>
        </t>

        <!-- Chatter -->
        <div class="o_portal_chatter">
            <t t-call="portal.message_thread">
                <t t-set="object" t-value="document"/>
            </t>
        </div>
    </t>
</template>
```

---

## Access Token Sharing

### Share Link Generation
```python
class MyDocument(models.Model):
    _name = 'my.document'
    _inherit = 'portal.mixin'

    share_url = fields.Char(compute='_compute_share_url')

    def _compute_share_url(self):
        """Generate shareable URL with access token."""
        base_url = self.env['ir.config_parameter'].sudo().get_param('web.base.url')
        for record in self:
            if record.access_token:
                record.share_url = (
                    f"{base_url}/my/documents/{record.id}"
                    f"?access_token={record.access_token}"
                )
            else:
                record.share_url = False

    def action_share(self):
        """Generate token and return share URL."""
        self.ensure_one()
        if not self.access_token:
            self.access_token = self._get_access_token()

        return {
            'type': 'ir.actions.act_window',
            'res_model': 'document.share.wizard',
            'view_mode': 'form',
            'target': 'new',
            'context': {
                'default_document_id': self.id,
                'default_share_url': self.share_url,
            },
        }
```

### Share Wizard
```python
class DocumentShareWizard(models.TransientModel):
    _name = 'document.share.wizard'

    document_id = fields.Many2one('my.document', required=True)
    share_url = fields.Char(readonly=True)
    email = fields.Char(string='Send to Email')

    def action_send_email(self):
        """Send share link via email."""
        template = self.env.ref('my_module.email_template_document_share')
        template.send_mail(self.document_id.id, email_values={
            'email_to': self.email,
        })
        return {'type': 'ir.actions.act_window_close'}

    def action_copy_link(self):
        """Copy link to clipboard (handled by JS)."""
        return {
            'type': 'ir.actions.client',
            'tag': 'display_notification',
            'params': {
                'title': 'Link Copied',
                'message': 'Share link copied to clipboard',
                'type': 'success',
            }
        }
```

---

## Public Access Controller

### Token-Based Public Access
```python
class PublicDocumentController(http.Controller):

    @http.route('/document/view/<string:token>',
                type='http', auth='public', website=True)
    def view_document_public(self, token, **kw):
        """View document with access token (no login required)."""
        document = request.env['my.document'].sudo().search([
            ('access_token', '=', token),
        ], limit=1)

        if not document:
            return request.render('http_routing.404')

        return request.render('my_module.document_public_view', {
            'document': document,
        })

    @http.route('/document/download/<string:token>',
                type='http', auth='public')
    def download_document_public(self, token, **kw):
        """Download document attachment with token."""
        document = request.env['my.document'].sudo().search([
            ('access_token', '=', token),
        ], limit=1)

        if not document or not document.attachment_id:
            return request.not_found()

        attachment = document.attachment_id
        return request.make_response(
            base64.b64decode(attachment.datas),
            headers=[
                ('Content-Type', attachment.mimetype),
                ('Content-Disposition', f'attachment; filename="{attachment.name}"'),
            ]
        )
```

---

## Secure Token Patterns

### Time-Limited Tokens
```python
import hashlib
import time
from odoo import api, fields, models


class TimedAccessToken(models.Model):
    _name = 'timed.access.token'

    document_id = fields.Many2one('my.document', required=True)
    token = fields.Char(required=True)
    expires_at = fields.Datetime(required=True)
    created_by = fields.Many2one('res.users', default=lambda self: self.env.uid)

    @api.model
    def create_token(self, document_id, validity_hours=24):
        """Create time-limited access token."""
        token = hashlib.sha256(
            f"{document_id}-{time.time()}-{self.env.uid}".encode()
        ).hexdigest()

        expires = fields.Datetime.now() + timedelta(hours=validity_hours)

        return self.create({
            'document_id': document_id,
            'token': token,
            'expires_at': expires,
        })

    @api.model
    def validate_token(self, token):
        """Validate token and return document if valid."""
        access = self.search([
            ('token', '=', token),
            ('expires_at', '>', fields.Datetime.now()),
        ], limit=1)

        if access:
            return access.document_id
        return False

    @api.model
    def cleanup_expired(self):
        """Cron job to remove expired tokens."""
        expired = self.search([
            ('expires_at', '<', fields.Datetime.now()),
        ])
        expired.unlink()
```

### Signed URLs
```python
import hmac
import hashlib
from urllib.parse import urlencode


class SignedUrlMixin(models.AbstractModel):
    _name = 'signed.url.mixin'

    def _get_signing_key(self):
        """Get secret key for signing."""
        return self.env['ir.config_parameter'].sudo().get_param(
            'my_module.signing_key',
            default='default-secret-key'
        )

    def generate_signed_url(self, base_path, params, validity_seconds=3600):
        """Generate signed URL with expiration."""
        expires = int(time.time()) + validity_seconds
        params['expires'] = expires

        # Create signature
        message = f"{base_path}?{urlencode(sorted(params.items()))}"
        signature = hmac.new(
            self._get_signing_key().encode(),
            message.encode(),
            hashlib.sha256
        ).hexdigest()

        params['signature'] = signature
        return f"{base_path}?{urlencode(params)}"

    def verify_signature(self, base_path, params):
        """Verify URL signature."""
        signature = params.pop('signature', None)
        expires = int(params.get('expires', 0))

        if not signature or expires < time.time():
            return False

        message = f"{base_path}?{urlencode(sorted(params.items()))}"
        expected = hmac.new(
            self._get_signing_key().encode(),
            message.encode(),
            hashlib.sha256
        ).hexdigest()

        return hmac.compare_digest(signature, expected)
```

---

## Portal User Management

### Create Portal User
```python
def action_grant_portal_access(self):
    """Create portal user for partner."""
    self.ensure_one()
    partner = self.partner_id

    if not partner.email:
        raise UserError("Partner must have an email address.")

    # Check if user exists
    existing = self.env['res.users'].sudo().search([
        ('partner_id', '=', partner.id),
    ])

    if existing:
        raise UserError("Partner already has a user account.")

    # Create portal user
    portal_wizard = self.env['portal.wizard'].create({
        'partner_ids': [(6, 0, [partner.id])],
    })

    portal_wizard.action_apply()
```

### Record Rule for Portal
```xml
<record id="document_portal_rule" model="ir.rule">
    <field name="name">Portal: Own Documents</field>
    <field name="model_id" ref="model_my_document"/>
    <field name="domain_force">[
        ('partner_id', '=', user.partner_id.id)
    ]</field>
    <field name="groups" eval="[(4, ref('base.group_portal'))]"/>
    <field name="perm_read" eval="True"/>
    <field name="perm_write" eval="False"/>
    <field name="perm_create" eval="False"/>
    <field name="perm_unlink" eval="False"/>
</record>
```

---

## Best Practices

1. **Always validate tokens** - Check existence and expiration
2. **Use HTTPS** - Tokens in URLs require encrypted transport
3. **Regenerate on sharing** - New token for each share request
4. **Limit token scope** - One token per document
5. **Set expiration** - Time-limit tokens when possible
6. **Log access** - Track who accessed what and when
7. **Clean up expired** - Regular cron to remove old tokens
8. **Check permissions** - Even with token, verify allowed actions
9. **Use sudo carefully** - Only for token validation
10. **Test portal views** - Verify as portal user, not admin

---


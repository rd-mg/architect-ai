> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Controllers & Website Reference

## Table of Contents
- [Controller Basics](#controller-basics)
- [HTTP Routes](#http-routes)
- [JSON-RPC Endpoints](#json-rpc-endpoints)
- [Website Pages](#website-pages)
- [Portal Pages](#portal-pages)
- [File Structure](#file-structure)

## Controller Basics

Controllers handle HTTP requests. File structure:

```
controllers/
├── __init__.py
└── main.py
```

Register in module `__init__.py`:
```python
from . import controllers
```

And `controllers/__init__.py`:
```python
from . import main
```

## HTTP Routes

```python
from odoo import http
from odoo.http import request

class LibraryController(http.Controller):

    @http.route('/library/books', type='http', auth='user', website=True)
    def list_books(self, **kwargs):
        books = request.env['library.book'].search([('state', '=', 'available')])
        return request.render('library_management.books_list_template', {
            'books': books,
        })

    @http.route('/library/book/<int:book_id>', type='http', auth='user', website=True)
    def book_detail(self, book_id, **kwargs):
        book = request.env['library.book'].browse(book_id)
        if not book.exists():
            raise request.not_found()
        return request.render('library_management.book_detail_template', {
            'book': book,
        })
```

### Route Parameters

| Parameter | Values | Description |
|-----------|--------|-------------|
| `type` | `'http'`, `'json'` | HTTP returns HTML/files; JSON returns JSON-RPC |
| `auth` | `'user'`, `'public'`, `'none'` | `user` = logged in, `public` = guest or logged in, `none` = no session |
| `website` | `True/False` | Enable website features (layout, SEO, etc.) |
| `methods` | `['GET']`, `['POST']` | Allowed HTTP methods |
| `csrf` | `True/False` | CSRF protection (default True for `type='http'`) |
| `cors` | `'*'` | CORS header value |
| `sitemap` | `True/False/function` | Include in sitemap |

## JSON-RPC Endpoints

For API-style endpoints returning JSON:

```python
class LibraryAPI(http.Controller):

    @http.route('/api/library/books', type='json', auth='user', methods=['POST'])
    def get_books(self, domain=None, limit=10, offset=0):
        domain = domain or [('state', '=', 'available')]
        books = request.env['library.book'].search(domain, limit=limit, offset=offset)
        return [{
            'id': book.id,
            'name': book.name,
            'isbn': book.isbn,
            'state': book.state,
        } for book in books]

    @http.route('/api/library/book/borrow', type='json', auth='user', methods=['POST'])
    def borrow_book(self, book_id):
        book = request.env['library.book'].browse(book_id)
        if book.state != 'available':
            return {'error': 'Book is not available'}
        book.write({'state': 'borrowed'})
        return {'success': True, 'book_id': book.id}
```

JSON-RPC requests must POST a JSON body with `jsonrpc`, `method`, and `params` keys:

```json
{
    "jsonrpc": "2.0",
    "method": "call",
    "params": {"book_id": 1}
}
```

## Website Pages

### QWeb Template for Website Page

```xml
<template id="books_list_template" name="Library Books">
    <t t-call="website.layout">
        <div class="container mt-4">
            <h1>Library Books</h1>
            <div class="row">
                <div t-foreach="books" t-as="book" class="col-md-4 mb-3">
                    <div class="card">
                        <div class="card-body">
                            <h5 class="card-title" t-esc="book.name"/>
                            <p class="card-text">
                                <span class="badge bg-success" t-if="book.state == 'available'">Available</span>
                                <span class="badge bg-warning" t-if="book.state == 'borrowed'">Borrowed</span>
                            </p>
                            <a t-attf-href="/library/book/#{book.id}" class="btn btn-primary btn-sm">
                                View Details
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </t>
</template>
```

Use `website.layout` to wrap pages in the website theme (header, footer, etc.).

## Portal Pages

For authenticated customer/vendor portal pages:

```python
from odoo.addons.portal.controllers.portal import CustomerPortal

class LibraryPortal(CustomerPortal):

    def _prepare_home_portal_values(self, counters):
        values = super()._prepare_home_portal_values(counters)
        if 'book_count' in counters:
            values['book_count'] = request.env['library.book'].search_count([
                ('borrower_id', '=', request.env.user.partner_id.id),
            ])
        return values

    @http.route('/my/books', type='http', auth='user', website=True)
    def portal_my_books(self, **kwargs):
        books = request.env['library.book'].search([
            ('borrower_id', '=', request.env.user.partner_id.id),
        ])
        return request.render('library_management.portal_my_books', {
            'books': books,
            'page_name': 'books',
        })
```

### Portal Template

```xml
<template id="portal_my_books" name="My Books">
    <t t-call="portal.portal_layout">
        <t t-set="breadcrumbs_searchbar" t-value="True"/>
        <t t-call="portal.portal_table">
            <thead>
                <tr>
                    <th>Book</th>
                    <th>Borrow Date</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                <tr t-foreach="books" t-as="book">
                    <td><t t-esc="book.name"/></td>
                    <td><span t-field="book.borrow_date"/></td>
                    <td><span class="badge" t-esc="book.state"/></td>
                </tr>
            </tbody>
        </t>
    </t>
</template>
```

## File Structure

Register controller views in `__manifest__.py`:

```python
'data': [
    'views/website_templates.xml',
    'views/portal_templates.xml',
],
```

Add website dependency if using website features:

```python
'depends': ['website', 'portal'],
```

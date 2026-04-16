> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Tests Reference

## Table of Contents
- [Test Structure](#test-structure)
- [Test Classes](#test-classes)
- [Running Tests](#running-tests)
- [Common Patterns](#common-patterns)
- [HTTP Tests](#http-tests)
- [Test Tags](#test-tags)

## Test Structure

```
tests/
├── __init__.py
├── test_library_book.py
└── test_library_wizard.py
```

Register in `tests/__init__.py`:
```python
from . import test_library_book
from . import test_library_wizard
```

The `tests/` package does NOT go in `__manifest__.py` `data` — Odoo auto-discovers it.

## Test Classes

### TransactionCase (most common)

Each test method runs in its own rolled-back transaction:

```python
from odoo.tests.common import TransactionCase
from odoo.exceptions import ValidationError, UserError


class TestLibraryBook(TransactionCase):

    @classmethod
    def setUpClass(cls):
        """Set up test data shared across all test methods."""
        super().setUpClass()
        cls.Book = cls.env['library.book']
        cls.partner = cls.env['res.partner'].create({'name': 'Test Author'})
        cls.book = cls.Book.create({
            'name': 'Test Book',
            'isbn': '978-0-123456-78-9',
            'author_ids': [(4, cls.partner.id)],
        })

    def test_create_book(self):
        """Test book creation with required fields."""
        book = self.Book.create({'name': 'New Book'})
        self.assertEqual(book.state, 'draft')
        self.assertTrue(book.active)

    def test_borrow_book(self):
        """Test borrowing workflow."""
        self.book.action_confirm()
        self.assertEqual(self.book.state, 'confirmed')

    def test_constraint_isbn_unique(self):
        """Test ISBN uniqueness constraint."""
        with self.assertRaises(Exception):
            self.Book.create({
                'name': 'Duplicate ISBN Book',
                'isbn': '978-0-123456-78-9',
            })

    def test_validation_error(self):
        """Test custom validation raises error."""
        with self.assertRaises(ValidationError):
            self.book.write({
                'date_start': '2024-12-31',
                'date_end': '2024-01-01',
            })
```

### SavepointCase / SingleTransactionCase

All test methods share one transaction (faster, but tests depend on each other):

```python
from odoo.tests.common import SingleTransactionCase

class TestLibraryFlow(SingleTransactionCase):

    @classmethod
    def setUpClass(cls):
        super().setUpClass()
        cls.book = cls.env['library.book'].create({'name': 'Flow Test'})

    def test_01_draft(self):
        self.assertEqual(self.book.state, 'draft')

    def test_02_confirm(self):
        self.book.action_confirm()
        self.assertEqual(self.book.state, 'confirmed')
```

## Running Tests

### Specific module
```bash
./odoo-bin -d testdb --test-enable --stop-after-init -i library_management
```

### Specific test file
```bash
./odoo-bin -d testdb --test-enable --stop-after-init -i library_management --test-file=tests/test_library_book.py
```

### With test tags
```bash
./odoo-bin -d testdb --test-tags=library_management
```

### Specific tagged tests
```bash
./odoo-bin -d testdb --test-tags=at_install,post_install,-standard
```

## Common Patterns

### Testing Access Rights

```python
def test_user_cannot_delete(self):
    """Test that regular users cannot delete books."""
    user = self.env['res.users'].create({
        'name': 'Test User',
        'login': 'test_user',
        'groups_id': [(6, 0, [self.env.ref('library_management.group_library_user').id])],
    })
    book_as_user = self.book.with_user(user)
    with self.assertRaises(Exception):
        book_as_user.unlink()
```

### Testing Computed Fields

```python
def test_compute_total(self):
    """Test total computation."""
    order = self.env['library.order'].create({
        'line_ids': [
            (0, 0, {'book_id': self.book.id, 'quantity': 2, 'price': 10.0}),
            (0, 0, {'book_id': self.book.id, 'quantity': 1, 'price': 15.0}),
        ],
    })
    self.assertAlmostEqual(order.total, 35.0, places=2)
```

### Testing Record Rules

```python
def test_multi_company_rule(self):
    """Test users only see their company's books."""
    company_2 = self.env['res.company'].create({'name': 'Company 2'})
    book_c2 = self.Book.with_company(company_2).create({
        'name': 'Company 2 Book',
        'company_id': company_2.id,
    })
    user_c1 = self.env['res.users'].create({
        'name': 'C1 User',
        'login': 'c1_user',
        'company_id': self.env.company.id,
        'company_ids': [(6, 0, [self.env.company.id])],
    })
    visible = self.Book.with_user(user_c1).search([])
    self.assertNotIn(book_c2, visible)
```

### Form Simulation with `Form`

```python
from odoo.tests.common import Form

def test_form_onchange(self):
    """Test onchange behavior via form simulation."""
    form = Form(self.env['library.book'])
    form.name = 'New Book'
    form.category_id = self.env.ref('library_management.category_fiction')
    book = form.save()
    self.assertEqual(book.name, 'New Book')
```

## HTTP Tests

For testing controllers:

```python
from odoo.tests.common import HttpCase

class TestLibraryWebsite(HttpCase):

    def test_book_page(self):
        """Test book list page returns 200."""
        self.authenticate('admin', 'admin')
        response = self.url_open('/library/books')
        self.assertEqual(response.status_code, 200)

    def test_json_api(self):
        """Test JSON-RPC endpoint."""
        self.authenticate('admin', 'admin')
        response = self.url_open(
            '/api/library/books',
            data=json.dumps({'jsonrpc': '2.0', 'params': {}}),
            headers={'Content-Type': 'application/json'},
        )
        self.assertEqual(response.status_code, 200)
```

### Tour Tests (UI tests)

```python
class TestLibraryTour(HttpCase):

    def test_library_tour(self):
        self.start_tour("/web", 'library_management_tour', login="admin")
```

## Test Tags

```python
from odoo.tests import tagged

@tagged('post_install', '-at_install')
class TestLibraryPostInstall(TransactionCase):
    """Runs after all modules are installed."""
    pass

@tagged('at_install')
class TestLibraryAtInstall(TransactionCase):
    """Runs during module installation (default)."""
    pass
```

| Tag | When |
|-----|------|
| `at_install` | During module install (default) |
| `post_install` | After all modules installed |
| `standard` | Default tag, runs in CI |
| `-standard` | Exclude from standard runs |

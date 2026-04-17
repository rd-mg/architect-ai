# Data Operations Patterns

Consolidated from the following source files:
- `import-export-patterns.md`
- `data-migration-patterns.md`
- `sequence-numbering-patterns.md`
- `context-environment-patterns.md`
- `external-api-patterns.md`

---


## Source: import-export-patterns.md

# Import/Export Data Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  IMPORT/EXPORT DATA PATTERNS                                                 ║
║  CSV import, Excel export, data migration, and bulk operations               ║
║  Use for data loading, reporting exports, and system integration             ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## CSV Import Basics

### Standard Import Format
```csv
id,name,email,phone,country_id/id
partner_001,Acme Corp,info@acme.com,+1-555-0100,base.us
partner_002,Beta Inc,contact@beta.com,+1-555-0200,base.us
__import__.partner_003,Gamma LLC,hello@gamma.com,+1-555-0300,base.ca
```

### Import Column Patterns
| Pattern | Description | Example |
|---------|-------------|---------|
| `field` | Direct field | `name`, `email` |
| `field/id` | External ID reference | `country_id/id` |
| `field.subfield` | Related field | `partner_id.name` |
| `field/0/subfield` | One2many line | `line_ids/0/product_id` |

---

## Programmatic Import

### Import CSV Data
```python
import base64
import csv
from io import StringIO


def import_partners_from_csv(self, csv_content):
    """Import partners from CSV string."""
    reader = csv.DictReader(StringIO(csv_content))

    created = []
    errors = []

    for row in reader:
        try:
            # Find country by code
            country = self.env['res.country'].search([
                ('code', '=', row.get('country_code', 'US'))
            ], limit=1)

            partner = self.env['res.partner'].create({
                'name': row['name'],
                'email': row.get('email'),
                'phone': row.get('phone'),
                'country_id': country.id,
            })
            created.append(partner.id)
        except Exception as e:
            errors.append({
                'row': row,
                'error': str(e),
            })

    return {
        'created': created,
        'errors': errors,
        'total': len(created) + len(errors),
    }
```

### Using base_import
```python
def import_with_base_import(self, model_name, csv_content, fields):
    """Use Odoo's base import functionality."""
    import_wizard = self.env['base_import.import'].create({
        'res_model': model_name,
        'file': base64.b64encode(csv_content.encode()),
        'file_name': 'import.csv',
        'file_type': 'text/csv',
    })

    result = import_wizard.execute_import(
        fields,  # ['name', 'email', 'country_id/id']
        [],  # columns (auto-detect)
        {'quoting': '"', 'separator': ',', 'headers': True}
    )

    return result
```

---

## Export Data

### Export to CSV
```python
import csv
from io import StringIO
import base64


def export_partners_to_csv(self, domain=None):
    """Export partners to CSV."""
    partners = self.env['res.partner'].search(domain or [])

    output = StringIO()
    writer = csv.writer(output)

    # Header
    writer.writerow(['ID', 'Name', 'Email', 'Phone', 'Country'])

    # Data
    for partner in partners:
        writer.writerow([
            partner.id,
            partner.name,
            partner.email or '',
            partner.phone or '',
            partner.country_id.name or '',
        ])

    content = output.getvalue()
    output.close()

    return content
```

### Export to Excel
```python
import xlsxwriter
from io import BytesIO
import base64


def export_to_excel(self, records, fields, filename='export.xlsx'):
    """Export records to Excel file."""
    output = BytesIO()
    workbook = xlsxwriter.Workbook(output, {'in_memory': True})
    worksheet = workbook.add_worksheet('Data')

    # Styles
    header_format = workbook.add_format({
        'bold': True,
        'bg_color': '#4472C4',
        'font_color': 'white',
        'border': 1,
    })
    date_format = workbook.add_format({'num_format': 'yyyy-mm-dd'})
    money_format = workbook.add_format({'num_format': '#,##0.00'})

    # Write header
    for col, field in enumerate(fields):
        worksheet.write(0, col, field, header_format)

    # Write data
    for row, record in enumerate(records, start=1):
        for col, field in enumerate(fields):
            value = record[field]
            if isinstance(value, (int, float)):
                worksheet.write_number(row, col, value)
            elif hasattr(value, 'id'):  # Many2one
                worksheet.write(row, col, value.display_name or '')
            else:
                worksheet.write(row, col, str(value) if value else '')

    workbook.close()
    output.seek(0)

    return base64.b64encode(output.read())
```

---

## Import Wizard

### Create Import Wizard
```python
class ImportWizard(models.TransientModel):
    _name = 'import.wizard'
    _description = 'Import Wizard'

    file = fields.Binary(string='File', required=True)
    file_name = fields.Char(string='File Name')
    import_type = fields.Selection([
        ('create', 'Create Only'),
        ('update', 'Update Only'),
        ('create_update', 'Create or Update'),
    ], default='create', required=True)

    result_ids = fields.One2many(
        'import.wizard.result',
        'wizard_id',
        string='Results',
    )

    def action_import(self):
        """Process the import."""
        self.ensure_one()

        # Decode file
        content = base64.b64decode(self.file).decode('utf-8')
        reader = csv.DictReader(StringIO(content))

        results = []
        for row_num, row in enumerate(reader, start=2):
            try:
                record = self._process_row(row)
                results.append({
                    'wizard_id': self.id,
                    'row_number': row_num,
                    'status': 'success',
                    'record_id': record.id,
                    'message': f'Created: {record.display_name}',
                })
            except Exception as e:
                results.append({
                    'wizard_id': self.id,
                    'row_number': row_num,
                    'status': 'error',
                    'message': str(e),
                })

        self.env['import.wizard.result'].create(results)

        return {
            'type': 'ir.actions.act_window',
            'res_model': self._name,
            'res_id': self.id,
            'view_mode': 'form',
            'target': 'new',
        }

    def _process_row(self, row):
        """Process single row. Override in subclass."""
        raise NotImplementedError()


class ImportWizardResult(models.TransientModel):
    _name = 'import.wizard.result'
    _description = 'Import Result'

    wizard_id = fields.Many2one('import.wizard')
    row_number = fields.Integer()
    status = fields.Selection([
        ('success', 'Success'),
        ('error', 'Error'),
        ('skipped', 'Skipped'),
    ])
    record_id = fields.Integer()
    message = fields.Char()
```

### Wizard View
```xml
<record id="import_wizard_form" model="ir.ui.view">
    <field name="name">import.wizard.form</field>
    <field name="model">import.wizard</field>
    <field name="arch" type="xml">
        <form>
            <group invisible="result_ids">
                <field name="file" filename="file_name"/>
                <field name="file_name" invisible="1"/>
                <field name="import_type"/>
            </group>
            <group string="Results" invisible="not result_ids">
                <field name="result_ids" nolabel="1">
                    <tree>
                        <field name="row_number"/>
                        <field name="status" decoration-success="status == 'success'"
                               decoration-danger="status == 'error'"/>
                        <field name="message"/>
                    </tree>
                </field>
            </group>
            <footer>
                <button name="action_import" string="Import" type="object"
                        class="btn-primary" invisible="result_ids"/>
                <button string="Close" class="btn-secondary" special="cancel"/>
            </footer>
        </form>
    </field>
</record>
```

---

## External ID Handling

### Create with External ID
```python
def import_with_xmlid(self, model, xmlid, vals):
    """Create/update record with external ID."""
    # Check if record exists
    record = self.env.ref(xmlid, raise_if_not_found=False)

    if record:
        record.write(vals)
    else:
        # Create with external ID
        record = self.env[model].create(vals)

        # Create external ID
        module, name = xmlid.split('.')
        self.env['ir.model.data'].create({
            'name': name,
            'module': module,
            'model': model,
            'res_id': record.id,
            'noupdate': False,
        })

    return record
```

### Lookup by External ID
```python
def get_by_xmlid(self, xmlid, model=None):
    """Get record by external ID."""
    record = self.env.ref(xmlid, raise_if_not_found=False)
    if record and model:
        if record._name != model:
            return None
    return record

def get_xmlid(self, record):
    """Get external ID for record."""
    data = self.env['ir.model.data'].search([
        ('model', '=', record._name),
        ('res_id', '=', record.id),
    ], limit=1)
    return f'{data.module}.{data.name}' if data else None
```

---

## Batch Processing

### Import Large Files
```python
def import_large_file(self, file_path, batch_size=1000):
    """Import large file in batches."""
    with open(file_path, 'r') as f:
        reader = csv.DictReader(f)
        batch = []
        total_created = 0

        for row in reader:
            batch.append(self._prepare_vals(row))

            if len(batch) >= batch_size:
                self.env['my.model'].create(batch)
                total_created += len(batch)
                batch = []
                self.env.cr.commit()  # Commit batch

        # Process remaining
        if batch:
            self.env['my.model'].create(batch)
            total_created += len(batch)
            self.env.cr.commit()

    return total_created
```

### Background Import
```python
def action_import_async(self):
    """Queue import for background processing."""
    self.ensure_one()

    # Create attachment for the file
    attachment = self.env['ir.attachment'].create({
        'name': self.file_name,
        'datas': self.file,
        'res_model': self._name,
        'res_id': self.id,
    })

    # Schedule cron job or use queue_job
    self.env['ir.cron'].create({
        'name': f'Import: {self.file_name}',
        'model_id': self.env.ref('my_module.model_import_wizard').id,
        'state': 'code',
        'code': f'model.browse({self.id})._process_import_async()',
        'interval_number': 1,
        'interval_type': 'minutes',
        'numbercall': 1,
        'doall': True,
    })

    return {'type': 'ir.actions.client', 'tag': 'reload'}
```

---

## Export Reports

### Excel Report with Multiple Sheets
```python
def export_report_excel(self):
    """Export multi-sheet Excel report."""
    output = BytesIO()
    workbook = xlsxwriter.Workbook(output, {'in_memory': True})

    # Summary sheet
    summary = workbook.add_worksheet('Summary')
    summary.write(0, 0, 'Total Records')
    summary.write(0, 1, self.search_count([]))

    # Detail sheet
    detail = workbook.add_worksheet('Details')
    records = self.search([])

    # Headers
    headers = ['ID', 'Name', 'Date', 'Amount']
    for col, header in enumerate(headers):
        detail.write(0, col, header)

    # Data
    for row, rec in enumerate(records, start=1):
        detail.write(row, 0, rec.id)
        detail.write(row, 1, rec.name)
        detail.write(row, 2, str(rec.date) if rec.date else '')
        detail.write(row, 3, rec.amount)

    workbook.close()
    output.seek(0)

    # Create attachment
    attachment = self.env['ir.attachment'].create({
        'name': 'report.xlsx',
        'datas': base64.b64encode(output.read()),
        'mimetype': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })

    return {
        'type': 'ir.actions.act_url',
        'url': f'/web/content/{attachment.id}?download=true',
        'target': 'self',
    }
```

---

## JSON Import/Export

### Export to JSON
```python
import json


def export_to_json(self, records, fields):
    """Export records to JSON."""
    data = []
    for record in records:
        row = {}
        for field in fields:
            value = record[field]
            if hasattr(value, 'id'):  # Relational
                row[field] = {'id': value.id, 'name': value.display_name}
            elif hasattr(value, 'ids'):  # X2many
                row[field] = [{'id': r.id, 'name': r.display_name} for r in value]
            elif isinstance(value, (date, datetime)):
                row[field] = value.isoformat()
            else:
                row[field] = value
        data.append(row)

    return json.dumps(data, indent=2, default=str)
```

### Import from JSON
```python
def import_from_json(self, json_content):
    """Import records from JSON."""
    data = json.loads(json_content)

    created = []
    for row in data:
        # Process relational fields
        vals = {}
        for key, value in row.items():
            if isinstance(value, dict) and 'id' in value:
                vals[key] = value['id']
            elif isinstance(value, list) and value and isinstance(value[0], dict):
                vals[key] = [(6, 0, [v['id'] for v in value])]
            else:
                vals[key] = value

        record = self.create(vals)
        created.append(record.id)

    return created
```

---

## Data Validation

### Validate Import Data
```python
def validate_import_row(self, row):
    """Validate a single import row."""
    errors = []

    # Required fields
    if not row.get('name'):
        errors.append('Name is required')

    # Email format
    if row.get('email'):
        import re
        if not re.match(r'^[\w\.-]+@[\w\.-]+\.\w+$', row['email']):
            errors.append(f"Invalid email: {row['email']}")

    # Reference lookup
    if row.get('country_code'):
        country = self.env['res.country'].search([
            ('code', '=', row['country_code'])
        ], limit=1)
        if not country:
            errors.append(f"Unknown country: {row['country_code']}")

    return errors
```

---

## Best Practices

1. **Validate first** - Check all rows before creating any
2. **Use transactions** - Rollback on errors
3. **Batch processing** - Commit in chunks for large imports
4. **External IDs** - Use for data that may be re-imported
5. **Error reporting** - Show row numbers and specific errors
6. **Preview mode** - Let users review before committing
7. **Template download** - Provide import template
8. **Encoding** - Handle UTF-8 properly
9. **Progress feedback** - Show import progress
10. **Logging** - Log imports for audit trail

---


## Source: data-migration-patterns.md

# Data Migration and Upgrade Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  DATA MIGRATION PATTERNS                                                     ║
║  Version upgrades, data transformation, and migration scripts                ║
║  Use for module upgrades, data fixes, and version transitions                ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Migration Script Structure

### Directory Layout
```
my_module/
├── migrations/
│   ├── 14.0.1.1/
│   │   ├── pre-migrate.py
│   │   └── post-migrate.py
│   ├── 15.0.1.0/
│   │   ├── pre-migrate.py
│   │   └── post-migrate.py
│   └── 16.0.2.0/
│       └── post-migrate.py
└── __manifest__.py
```

### Version Numbering
```python
# __manifest__.py
{
    'name': 'My Module',
    'version': '16.0.2.0.1',  # odoo_version.module_version
    #           ^^ ^^ ^ ^
    #           |  |  | └── patch
    #           |  |  └──── minor
    #           |  └─────── major
    #           └────────── Odoo version
}
```

---

## Pre-Migration Scripts

Pre-migrations run BEFORE the module is updated. Use for:
- Renaming tables/columns
- Preserving data before schema changes
- Removing constraints that would block updates

### Basic Pre-Migration
```python
# migrations/16.0.2.0/pre-migrate.py
import logging
from odoo import SUPERUSER_ID
from odoo.api import Environment

_logger = logging.getLogger(__name__)


def migrate(cr, version):
    """Pre-migration: prepare database for update."""
    if not version:
        # Fresh install, no migration needed
        return

    _logger.info("Starting pre-migration from %s", version)

    # Rename column before ORM sees it
    cr.execute("""
        ALTER TABLE my_model
        RENAME COLUMN old_field TO x_old_field_backup
    """)

    _logger.info("Pre-migration completed")
```

### Rename Table
```python
def migrate(cr, version):
    """Rename table before model rename."""
    cr.execute("""
        SELECT EXISTS (
            SELECT FROM information_schema.tables
            WHERE table_name = 'old_model_name'
        )
    """)
    if cr.fetchone()[0]:
        cr.execute("ALTER TABLE old_model_name RENAME TO new_model_name")
        _logger.info("Renamed table old_model_name to new_model_name")
```

### Preserve Data Before Removal
```python
def migrate(cr, version):
    """Backup data before field removal."""
    cr.execute("""
        SELECT EXISTS (
            SELECT FROM information_schema.columns
            WHERE table_name = 'my_model'
            AND column_name = 'deprecated_field'
        )
    """)
    if cr.fetchone()[0]:
        # Create backup table
        cr.execute("""
            CREATE TABLE IF NOT EXISTS my_model_field_backup AS
            SELECT id, deprecated_field
            FROM my_model
            WHERE deprecated_field IS NOT NULL
        """)
        _logger.info("Backed up deprecated_field data")
```

### Remove Constraints
```python
def migrate(cr, version):
    """Remove constraint before schema change."""
    cr.execute("""
        SELECT constraint_name
        FROM information_schema.table_constraints
        WHERE table_name = 'my_model'
        AND constraint_name LIKE '%_check'
    """)
    for (constraint_name,) in cr.fetchall():
        cr.execute(f"ALTER TABLE my_model DROP CONSTRAINT {constraint_name}")
        _logger.info("Dropped constraint %s", constraint_name)
```

---

## Post-Migration Scripts

Post-migrations run AFTER the module is updated. Use for:
- Data transformation
- Setting default values
- Migrating data between fields
- Cleanup operations

### Basic Post-Migration
```python
# migrations/16.0.2.0/post-migrate.py
import logging
from odoo import SUPERUSER_ID, api

_logger = logging.getLogger(__name__)


def migrate(cr, version):
    """Post-migration: transform data after update."""
    if not version:
        return

    _logger.info("Starting post-migration from %s", version)

    env = api.Environment(cr, SUPERUSER_ID, {})

    # Use ORM for data operations
    records = env['my.model'].search([('new_field', '=', False)])
    for record in records:
        record.new_field = record.old_field

    _logger.info("Migrated %d records", len(records))
```

### Migrate Field Data
```python
def migrate(cr, version):
    """Migrate data from old field to new field."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    # Direct SQL for large datasets
    cr.execute("""
        UPDATE my_model
        SET new_state = CASE state
            WHEN 'draft' THEN 'new'
            WHEN 'open' THEN 'in_progress'
            WHEN 'done' THEN 'completed'
            ELSE 'unknown'
        END
        WHERE new_state IS NULL
    """)
    _logger.info("Migrated %d state values", cr.rowcount)
```

### Migrate Many2one to Many2many
```python
def migrate(cr, version):
    """Convert single relation to multiple."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    # Get records with old single value
    cr.execute("""
        SELECT id, old_partner_id
        FROM my_model
        WHERE old_partner_id IS NOT NULL
    """)

    for record_id, partner_id in cr.fetchall():
        # Insert into relation table
        cr.execute("""
            INSERT INTO my_model_partner_rel (my_model_id, partner_id)
            VALUES (%s, %s)
            ON CONFLICT DO NOTHING
        """, (record_id, partner_id))

    _logger.info("Migrated partner relations")
```

### Set Computed Field Store
```python
def migrate(cr, version):
    """Initialize stored computed field."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    # Trigger recomputation
    records = env['my.model'].search([])
    records._compute_total_amount()

    # Or use SQL for performance
    cr.execute("""
        UPDATE my_model m
        SET total_amount = (
            SELECT COALESCE(SUM(l.amount), 0)
            FROM my_model_line l
            WHERE l.model_id = m.id
        )
    """)
```

### Migrate XML IDs
```python
def migrate(cr, version):
    """Rename XML IDs after module rename."""
    cr.execute("""
        UPDATE ir_model_data
        SET module = 'new_module_name'
        WHERE module = 'old_module_name'
    """)

    # Update specific record references
    cr.execute("""
        UPDATE ir_model_data
        SET name = REPLACE(name, 'old_prefix_', 'new_prefix_')
        WHERE module = 'my_module'
        AND name LIKE 'old_prefix_%'
    """)
```

---

## openupgrade Patterns

Using OpenUpgrade library for complex migrations:

### Rename Field
```python
from openupgradelib import openupgrade


def migrate(cr, version):
    openupgrade.rename_fields(
        cr,
        [
            ('my.model', 'my_model', 'old_field', 'new_field'),
        ]
    )
```

### Rename Model
```python
def migrate(cr, version):
    openupgrade.rename_models(
        cr,
        [
            ('old.model', 'new.model'),
        ]
    )
    openupgrade.rename_tables(
        cr,
        [
            ('old_model', 'new_model'),
        ]
    )
```

### Merge Records
```python
def migrate(cr, version):
    """Merge duplicate records."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    duplicates = env['my.model'].search([])
    groups = {}
    for record in duplicates:
        key = record.name.lower().strip()
        if key not in groups:
            groups[key] = []
        groups[key].append(record)

    for key, records in groups.items():
        if len(records) > 1:
            main = records[0]
            for duplicate in records[1:]:
                openupgrade.merge_records(
                    env,
                    'my.model',
                    [duplicate.id],
                    main.id,
                )
```

---

## Batch Processing

### Large Dataset Migration
```python
def migrate(cr, version):
    """Process large dataset in batches."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    batch_size = 1000
    offset = 0
    total = 0

    while True:
        cr.execute("""
            SELECT id FROM my_model
            WHERE needs_migration = true
            ORDER BY id
            LIMIT %s OFFSET %s
        """, (batch_size, offset))

        ids = [row[0] for row in cr.fetchall()]
        if not ids:
            break

        records = env['my.model'].browse(ids)
        for record in records:
            record._migrate_data()
            total += 1

        # Commit batch
        cr.commit()
        env.invalidate_all()

        offset += batch_size
        _logger.info("Processed %d records...", total)

    _logger.info("Migration complete: %d records", total)
```

### Parallel Migration (Advanced)
```python
def migrate(cr, version):
    """Use SQL for parallel-safe operations."""
    # Atomic update with RETURNING
    while True:
        cr.execute("""
            WITH to_update AS (
                SELECT id FROM my_model
                WHERE migrated = false
                LIMIT 100
                FOR UPDATE SKIP LOCKED
            )
            UPDATE my_model m
            SET
                new_field = old_field * 1.1,
                migrated = true
            FROM to_update
            WHERE m.id = to_update.id
            RETURNING m.id
        """)

        updated = cr.fetchall()
        if not updated:
            break

        cr.commit()
```

---

## Testing Migrations

### Migration Test Pattern
```python
# tests/test_migration.py
from odoo.tests import TransactionCase


class TestMigration(TransactionCase):

    def setUp(self):
        super().setUp()
        # Create test data with old structure
        self.test_record = self.env['my.model'].create({
            'name': 'Test',
            'old_field': 'value',
        })

    def test_field_migration(self):
        """Test that old field migrates to new field."""
        # Simulate migration
        self.test_record._migrate_field()

        self.assertEqual(
            self.test_record.new_field,
            'value',
            "Field value should be migrated"
        )

    def test_state_migration(self):
        """Test state value mapping."""
        self.test_record.old_state = 'draft'
        self.test_record._migrate_state()

        self.assertEqual(
            self.test_record.state,
            'new',
            "State 'draft' should map to 'new'"
        )
```

---

## Common Migration Tasks

### Add Default Value
```python
def migrate(cr, version):
    """Set default for new required field."""
    cr.execute("""
        UPDATE my_model
        SET new_required_field = 'default_value'
        WHERE new_required_field IS NULL
    """)
```

### Convert Data Type
```python
def migrate(cr, version):
    """Convert char to integer."""
    # Add temporary column
    cr.execute("""
        ALTER TABLE my_model
        ADD COLUMN IF NOT EXISTS numeric_code INTEGER
    """)

    # Convert with error handling
    cr.execute("""
        UPDATE my_model
        SET numeric_code = CASE
            WHEN char_code ~ '^[0-9]+$' THEN char_code::INTEGER
            ELSE 0
        END
    """)
```

### Migrate Attachments
```python
def migrate(cr, version):
    """Move attachments to new model."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    attachments = env['ir.attachment'].search([
        ('res_model', '=', 'old.model'),
    ])

    for attachment in attachments:
        # Find corresponding new record
        cr.execute("""
            SELECT new_id FROM model_mapping
            WHERE old_id = %s
        """, (attachment.res_id,))
        result = cr.fetchone()

        if result:
            attachment.write({
                'res_model': 'new.model',
                'res_id': result[0],
            })
```

### Update Mail Followers
```python
def migrate(cr, version):
    """Update followers after model rename."""
    cr.execute("""
        UPDATE mail_followers
        SET res_model = 'new.model'
        WHERE res_model = 'old.model'
    """)

    cr.execute("""
        UPDATE mail_message
        SET model = 'new.model'
        WHERE model = 'old.model'
    """)
```

---

## Best Practices

### 1. Always Check Version
```python
def migrate(cr, version):
    if not version:
        return  # Fresh install
    # Migration code
```

### 2. Use Logging
```python
_logger.info("Starting migration from %s", version)
_logger.info("Migrated %d records", count)
_logger.warning("Skipped %d invalid records", skipped)
```

### 3. Handle Errors Gracefully
```python
def migrate(cr, version):
    try:
        # Migration code
    except Exception as e:
        _logger.error("Migration failed: %s", e)
        raise
```

### 4. Make Idempotent
```python
# Good - can run multiple times
cr.execute("""
    UPDATE my_model
    SET new_field = old_field
    WHERE new_field IS NULL  -- Only unmigrated
""")

# Bad - breaks on re-run
cr.execute("""
    UPDATE my_model
    SET new_field = old_field
""")
```

### 5. Commit Large Operations
```python
for i, record in enumerate(records):
    record.migrate()
    if i % 1000 == 0:
        cr.commit()
        env.invalidate_all()
```

### 6. Test Before Production
```python
# Run on copy of production database
# Verify data integrity after migration
# Check performance with realistic data volumes
```

### 7. Installing Modules During Migration
```python
# Good - use button_install() during migration
def migrate(cr, version):
    """Install dependency module during migration."""
    env = api.Environment(cr, SUPERUSER_ID, {})

    module = env['ir.module.module'].search([
        ('name', '=', 'required_module'),
        ('state', '!=', 'installed'),
    ])

    if module:
        module.button_install()
        _logger.info("Queued installation of %s", module.name)

# Bad - causes UserError during migration
def migrate(cr, version):
    env = api.Environment(cr, SUPERUSER_ID, {})
    module = env['ir.module.module'].search([
        ('name', '=', 'required_module'),
    ])
    module._button_immediate_install()  # ERROR: Cannot be called on non-loaded registries
```

**Why**: During migration, the registry is not fully loaded. The `_button_immediate_install()` method requires a complete registry and will raise:
```
odoo.exceptions.UserError: The method _button_immediate_install cannot be called on init or non loaded registries. Please use button_install instead.
```

Use `button_install()` which queues the installation to happen after the registry is properly initialized.

---

## Version-Specific Notes

| Version | Migration Notes |
|---------|-----------------|
| 14→15 | `@api.multi` removed, update method signatures |
| 15→16 | OWL 2.x, `Command` class for x2many |
| 16→17 | `attrs` removed, use inline expressions |
| 17→18 | `_check_company_auto`, `SQL()` builder |
| 18→19 | Type hints required, `SQL()` mandatory |

---


## Source: sequence-numbering-patterns.md

# Sequence and Numbering Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  SEQUENCE & NUMBERING PATTERNS                                               ║
║  Automatic numbering, reference generation, and sequence management          ║
║  Use for document numbers, codes, and unique identifiers                     ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Basic Sequence Setup

### Define Sequence (XML)
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <data noupdate="1">
        <record id="sequence_my_model" model="ir.sequence">
            <field name="name">My Model Sequence</field>
            <field name="code">my.model</field>
            <field name="prefix">MM/%(year)s/</field>
            <field name="padding">5</field>
            <field name="number_increment">1</field>
            <field name="number_next">1</field>
        </record>
    </data>
</odoo>
```

### Sequence Placeholders
| Placeholder | Description | Example |
|-------------|-------------|---------|
| `%(year)s` | 4-digit year | 2024 |
| `%(y)s` | 2-digit year | 24 |
| `%(month)s` | 2-digit month | 01-12 |
| `%(day)s` | 2-digit day | 01-31 |
| `%(doy)s` | Day of year | 001-366 |
| `%(woy)s` | Week of year | 01-53 |
| `%(weekday)s` | Day of week | 0-6 |
| `%(h24)s` | Hour (24h) | 00-23 |
| `%(h12)s` | Hour (12h) | 01-12 |
| `%(min)s` | Minutes | 00-59 |
| `%(sec)s` | Seconds | 00-59 |

### Sequence with Company
```xml
<record id="sequence_my_model_company" model="ir.sequence">
    <field name="name">My Model Sequence</field>
    <field name="code">my.model</field>
    <field name="prefix">%(company_code)s/%(year)s/</field>
    <field name="padding">4</field>
    <field name="company_id" ref="base.main_company"/>
</record>
```

---

## Using Sequences in Models

### Basic Usage
```python
from odoo import api, fields, models


class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'

    name = fields.Char(
        string='Reference',
        required=True,
        copy=False,
        readonly=True,
        default='New',
    )

    @api.model_create_multi
    def create(self, vals_list):
        for vals in vals_list:
            if vals.get('name', 'New') == 'New':
                vals['name'] = self.env['ir.sequence'].next_by_code(
                    'my.model'
                ) or 'New'
        return super().create(vals_list)
```

### With Date Context
```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if vals.get('name', 'New') == 'New':
            # Use specific date for sequence
            sequence_date = vals.get('date') or fields.Date.today()
            vals['name'] = self.env['ir.sequence'].with_context(
                ir_sequence_date=sequence_date
            ).next_by_code('my.model') or 'New'
    return super().create(vals_list)
```

### Company-Specific Sequence
```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if vals.get('name', 'New') == 'New':
            company_id = vals.get('company_id') or self.env.company.id
            vals['name'] = self.env['ir.sequence'].with_company(
                company_id
            ).next_by_code('my.model') or 'New'
    return super().create(vals_list)
```

### Conditional Sequence
```python
@api.model_create_multi
def create(self, vals_list):
    for vals in vals_list:
        if vals.get('name', 'New') == 'New':
            record_type = vals.get('type', 'standard')
            if record_type == 'internal':
                code = 'my.model.internal'
            else:
                code = 'my.model'
            vals['name'] = self.env['ir.sequence'].next_by_code(code) or 'New'
    return super().create(vals_list)
```

---

## Advanced Sequence Patterns

### Sequence per Partner
```python
class ResPartner(models.Model):
    _inherit = 'res.partner'

    sequence_id = fields.Many2one(
        'ir.sequence',
        string='Invoice Sequence',
        copy=False,
    )

    def _get_or_create_sequence(self):
        """Get or create partner-specific sequence."""
        self.ensure_one()
        if not self.sequence_id:
            self.sequence_id = self.env['ir.sequence'].create({
                'name': f'Invoice Sequence - {self.name}',
                'code': f'account.move.partner.{self.id}',
                'prefix': f'{self.ref or "CUST"}/%(year)s/',
                'padding': 4,
            })
        return self.sequence_id


class AccountMove(models.Model):
    _inherit = 'account.move'

    def _get_sequence(self):
        """Get appropriate sequence for invoice."""
        if self.partner_id.sequence_id:
            return self.partner_id.sequence_id
        return super()._get_sequence()
```

### Multi-Level Sequence
```python
class MyModel(models.Model):
    _name = 'my.model'

    name = fields.Char(string='Reference', readonly=True, default='New')
    department_id = fields.Many2one('hr.department')
    year = fields.Char(compute='_compute_year', store=True)

    @api.depends('create_date')
    def _compute_year(self):
        for record in self:
            record.year = str(record.create_date.year) if record.create_date else ''

    @api.model_create_multi
    def create(self, vals_list):
        for vals in vals_list:
            if vals.get('name', 'New') == 'New':
                dept_id = vals.get('department_id')
                dept = self.env['hr.department'].browse(dept_id) if dept_id else None
                dept_code = dept.x_code if dept else 'GEN'

                # Create sequence: DEPT/YEAR/NUMBER
                year = fields.Date.today().year
                prefix = f'{dept_code}/{year}/'

                # Get next number for this prefix
                last_record = self.search([
                    ('name', 'like', f'{prefix}%'),
                ], order='name desc', limit=1)

                if last_record:
                    last_num = int(last_record.name.split('/')[-1])
                    next_num = last_num + 1
                else:
                    next_num = 1

                vals['name'] = f'{prefix}{next_num:05d}'

        return super().create(vals_list)
```

### Sequence with Reset
```xml
<!-- Reset yearly -->
<record id="sequence_yearly_reset" model="ir.sequence">
    <field name="name">Yearly Reset Sequence</field>
    <field name="code">my.model.yearly</field>
    <field name="prefix">INV/%(year)s/</field>
    <field name="padding">5</field>
    <field name="use_date_range">True</field>
</record>
```

```python
# The sequence will auto-create date ranges when use_date_range=True
# Each year gets its own counter starting from 1
```

### Sequence for Sub-Records
```python
class MyModelLine(models.Model):
    _name = 'my.model.line'

    model_id = fields.Many2one('my.model', required=True, ondelete='cascade')
    line_number = fields.Integer(string='Line #', readonly=True)
    name = fields.Char(string='Description')

    @api.model_create_multi
    def create(self, vals_list):
        for vals in vals_list:
            if not vals.get('line_number'):
                model_id = vals.get('model_id')
                if model_id:
                    last_line = self.search([
                        ('model_id', '=', model_id),
                    ], order='line_number desc', limit=1)
                    vals['line_number'] = (last_line.line_number or 0) + 1
        return super().create(vals_list)
```

---

## Custom Reference Generation

### UUID-Based Reference
```python
import uuid


class MyModel(models.Model):
    _name = 'my.model'

    reference = fields.Char(
        string='Reference',
        default=lambda self: str(uuid.uuid4())[:8].upper(),
        readonly=True,
        copy=False,
    )
```

### Hash-Based Reference
```python
import hashlib


class MyModel(models.Model):
    _name = 'my.model'

    @api.model_create_multi
    def create(self, vals_list):
        records = super().create(vals_list)
        for record in records:
            if not record.reference:
                # Create hash from record data
                data = f'{record.id}{record.create_date}{record.partner_id.id}'
                hash_val = hashlib.md5(data.encode()).hexdigest()[:8].upper()
                record.reference = f'REF-{hash_val}'
        return records
```

### Checksum Reference
```python
class MyModel(models.Model):
    _name = 'my.model'

    def _generate_reference_with_checksum(self):
        """Generate reference with Luhn checksum."""
        base_number = self.env['ir.sequence'].next_by_code('my.model.base')
        checksum = self._luhn_checksum(base_number)
        return f'{base_number}{checksum}'

    def _luhn_checksum(self, number_str):
        """Calculate Luhn checksum digit."""
        digits = [int(d) for d in str(number_str)]
        odd_digits = digits[-1::-2]
        even_digits = digits[-2::-2]
        total = sum(odd_digits)
        for d in even_digits:
            total += sum(divmod(d * 2, 10))
        return (10 - (total % 10)) % 10
```

---

## Sequence Views

### Sequence Form View
```xml
<record id="view_sequence_form_inherit" model="ir.ui.view">
    <field name="name">ir.sequence.form.inherit</field>
    <field name="model">ir.sequence</field>
    <field name="inherit_id" ref="base.sequence_view"/>
    <field name="arch" type="xml">
        <field name="company_id" position="after">
            <field name="x_department_id"/>
        </field>
    </field>
</record>
```

### Menu for Sequence Management
```xml
<menuitem id="menu_sequence_config"
          name="Sequences"
          parent="base.menu_custom"
          action="base.ir_sequence_form"
          groups="base.group_system"/>
```

---

## Sequence Security

### Per-Company Sequences
```python
def _setup_company_sequences(self, company):
    """Create sequences for new company."""
    sequences = [
        {
            'name': f'My Model - {company.name}',
            'code': 'my.model',
            'prefix': f'{company.x_code}/%(year)s/',
            'padding': 5,
            'company_id': company.id,
        },
    ]

    for seq_vals in sequences:
        existing = self.env['ir.sequence'].search([
            ('code', '=', seq_vals['code']),
            ('company_id', '=', company.id),
        ])
        if not existing:
            self.env['ir.sequence'].create(seq_vals)
```

### Sequence Access Rights
```xml
<!-- Allow users to see sequence in selection -->
<record id="rule_sequence_user" model="ir.rule">
    <field name="name">Sequence: User Access</field>
    <field name="model_id" ref="base.model_ir_sequence"/>
    <field name="domain_force">[
        '|',
        ('company_id', '=', False),
        ('company_id', 'in', company_ids)
    ]</field>
    <field name="groups" eval="[(4, ref('base.group_user'))]"/>
    <field name="perm_read" eval="True"/>
    <field name="perm_write" eval="False"/>
    <field name="perm_create" eval="False"/>
    <field name="perm_unlink" eval="False"/>
</record>
```

---

## Best Practices

### 1. Use noupdate for Sequences
```xml
<data noupdate="1">
    <!-- Sequences should not be updated on module upgrade -->
    <record id="sequence_my_model" model="ir.sequence">
        ...
    </record>
</data>
```

### 2. Handle Concurrent Access
```python
# Sequences are thread-safe by design
# But custom numbering may need locking
def _get_next_number(self):
    self.env.cr.execute("""
        SELECT COALESCE(MAX(sequence_number), 0) + 1
        FROM my_model
        WHERE parent_id = %s
        FOR UPDATE
    """, (self.parent_id.id,))
    return self.env.cr.fetchone()[0]
```

### 3. Never Reuse Numbers
```python
# Bad - gaps are OK, reusing is not
def _fill_gap(self):
    # Never do this!
    pass

# Good - let gaps exist
# Document that gaps are normal and expected
```

### 4. Meaningful Prefixes
```python
# Good - meaningful prefixes
'prefix': 'INV/%(year)s/'  # Invoice
'prefix': 'SO/%(year)s/'   # Sale Order
'prefix': 'PO/%(year)s/'   # Purchase Order

# Bad - cryptic prefixes
'prefix': 'X1/%(year)s/'
```

### 5. Consistent Padding
```python
# Choose padding based on expected volume
# 4 digits = up to 9,999 per period
# 5 digits = up to 99,999 per period
# 6 digits = up to 999,999 per period
```

### 6. Document Sequence Logic
```python
class MyModel(models.Model):
    """
    Reference format: TYPE/YEAR/NUMBER
    - TYPE: 2-letter type code (IN=Internal, EX=External)
    - YEAR: 4-digit year
    - NUMBER: 5-digit sequential number, resets yearly

    Examples: IN/2024/00001, EX/2024/00042
    """
    _name = 'my.model'
```

---

## Troubleshooting

### Reset Sequence
```sql
-- Reset sequence to specific number (use with caution!)
UPDATE ir_sequence
SET number_next = 1
WHERE code = 'my.model';

-- Check current value
SELECT code, number_next, prefix, suffix
FROM ir_sequence
WHERE code = 'my.model';
```

### Fix Gaps
```python
# Gaps are normal - don't fix them
# But if required for reporting:
def _get_missing_numbers(self):
    """Find gaps in sequence (for audit only)."""
    all_nums = self.search([]).mapped(
        lambda r: int(r.name.split('/')[-1])
    )
    full_range = set(range(1, max(all_nums) + 1))
    missing = full_range - set(all_nums)
    return sorted(missing)
```

---


## Source: context-environment-patterns.md

# Context and Environment Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  CONTEXT & ENVIRONMENT PATTERNS                                              ║
║  Using context, environment, and recordset manipulation                      ║
║  Use for passing data, changing behavior, and managing state                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Understanding the Environment

### Environment Components
```python
# self.env contains:
# - env.cr: Database cursor
# - env.uid: Current user ID
# - env.context: Context dictionary
# - env.user: Current user record
# - env.company: Current company
# - env.companies: Accessible companies
# - env.lang: Current language
# - env.ref(): Get record by XML ID
# - env['model.name']: Access model
```

### Accessing Environment
```python
class MyModel(models.Model):
    _name = 'my.model'

    def example_method(self):
        # Database cursor
        cr = self.env.cr

        # Current user
        user = self.env.user
        user_id = self.env.uid

        # Current company
        company = self.env.company
        companies = self.env.companies

        # Language
        lang = self.env.lang

        # Access other models
        partners = self.env['res.partner'].search([])

        # Get record by XML ID
        admin = self.env.ref('base.user_admin')

        # Context
        ctx = self.env.context
```

---

## Context Usage

### Reading Context Values
```python
def my_method(self):
    # Get context value with default
    active_id = self.env.context.get('active_id')
    active_ids = self.env.context.get('active_ids', [])
    active_model = self.env.context.get('active_model')

    # Boolean context flags
    skip_validation = self.env.context.get('skip_validation', False)

    # With default
    limit = self.env.context.get('limit', 100)
```

### Passing Context
```python
# Using with_context()
def action_with_context(self):
    # Add to existing context
    record = self.with_context(my_flag=True)
    record.do_something()

    # Replace entire context
    record = self.with_context({'lang': 'en_US'})

    # Multiple values
    record = self.with_context(
        active_test=False,
        lang='fr_FR',
        custom_value=42,
    )
```

### Context in Fields
```python
# Default from context
partner_id = fields.Many2one(
    'res.partner',
    default=lambda self: self.env.context.get('default_partner_id'),
)

# In XML views
"""
<field name="partner_id"
       context="{'default_company_id': company_id,
                 'show_archived': True}"/>
"""
```

### Context in Actions
```python
def action_open_wizard(self):
    return {
        'type': 'ir.actions.act_window',
        'name': 'My Wizard',
        'res_model': 'my.wizard',
        'view_mode': 'form',
        'target': 'new',
        'context': {
            'default_partner_id': self.partner_id.id,
            'default_amount': self.amount_total,
            'active_id': self.id,
            'active_ids': self.ids,
            'active_model': self._name,
        },
    }
```

---

## Common Context Keys

### Standard Keys
```python
# active_id/active_ids - Current record(s) from action
active_id = self.env.context.get('active_id')
active_ids = self.env.context.get('active_ids', [])

# active_model - Model name
active_model = self.env.context.get('active_model')

# default_* - Default values for fields
default_name = self.env.context.get('default_name')

# search_default_* - Default search filters
# In action: context="{'search_default_my_filter': 1}"

# active_test - Include archived records
# False = show archived, True/missing = hide archived
records = self.with_context(active_test=False).search([])

# lang - Language code
translated = self.with_context(lang='fr_FR').name

# tz - Timezone
# Automatically used for datetime display

# mail_create_nosubscribe - Don't auto-subscribe creator
# mail_create_nolog - Don't create "created" message
# mail_notrack - Don't track field changes
# tracking_disable - Disable all tracking
```

### Custom Context Patterns
```python
class MyModel(models.Model):
    _name = 'my.model'

    def create(self, vals):
        # Check for context flags
        if self.env.context.get('import_mode'):
            # Skip validations during import
            pass

        if self.env.context.get('from_cron'):
            # Different behavior for scheduled actions
            pass

        return super().create(vals)

    def _compute_field(self):
        # Use context to modify computation
        if self.env.context.get('simplified_calculation'):
            # Simplified logic
            pass
        else:
            # Full calculation
            pass
```

---

## Recordset Operations

### Creating Records
```python
# Single record
record = self.env['my.model'].create({
    'name': 'Test',
    'partner_id': partner.id,
})

# Multiple records (v17+)
records = self.env['my.model'].create([
    {'name': 'Record 1'},
    {'name': 'Record 2'},
])

# With context
record = self.env['my.model'].with_context(
    mail_create_nosubscribe=True
).create(vals)
```

### Searching Records
```python
# Basic search
records = self.env['my.model'].search([
    ('state', '=', 'draft'),
])

# With limit and order
records = self.env['my.model'].search(
    [('state', '=', 'draft')],
    limit=10,
    order='create_date desc',
)

# Search and read in one call
data = self.env['my.model'].search_read(
    [('state', '=', 'draft')],
    ['name', 'state', 'amount'],
    limit=10,
)

# Count
count = self.env['my.model'].search_count([('state', '=', 'draft')])

# Include archived
all_records = self.env['my.model'].with_context(
    active_test=False
).search([])
```

### Browsing Records
```python
# By ID
record = self.env['my.model'].browse(record_id)

# Multiple IDs
records = self.env['my.model'].browse([1, 2, 3])

# From context
records = self.env['my.model'].browse(
    self.env.context.get('active_ids', [])
)

# Check existence
if record.exists():
    # Record exists in database
    pass
```

### Recordset Manipulation
```python
# Combine recordsets (OR/union)
combined = records1 | records2

# Intersection
common = records1 & records2

# Difference
diff = records1 - records2

# Filter
draft_records = records.filtered(lambda r: r.state == 'draft')
draft_records = records.filtered('is_draft')  # Boolean field

# Map
names = records.mapped('name')
partner_ids = records.mapped('partner_id.id')
partners = records.mapped('partner_id')  # Returns recordset

# Sort
sorted_records = records.sorted(key=lambda r: r.date)
sorted_records = records.sorted('date', reverse=True)

# Iterate
for record in records:
    record.do_something()

# Check if recordset
if records:
    # Has at least one record
    pass

# Ensure single record
record.ensure_one()
```

---

## Changing User/Company Context

### Change User
```python
# Execute as specific user
admin = self.env.ref('base.user_admin')
record_as_admin = self.with_user(admin)
record_as_admin.action_confirm()

# Execute with sudo (superuser)
self.sudo().write({'internal_field': value})

# IMPORTANT: sudo() bypasses access rights
# Use sparingly and carefully
```

### Change Company
```python
# Execute in different company context
other_company = self.env['res.company'].browse(2)
record_in_company = self.with_company(other_company)
record_in_company.create_in_company()

# Get company from context
company = self.env.company  # Current company
companies = self.env.companies  # All accessible
```

### Combining Context Changes
```python
# Change multiple aspects
result = self.sudo().with_company(company).with_context(
    skip_validation=True,
    lang='en_US',
).create(vals)
```

---

## Environment in Cron Jobs

### Proper Cron Setup
```python
@api.model
def _cron_process(self):
    """Cron method with proper environment handling."""
    # Cron runs as OdooBot or specific user

    # Process records
    records = self.search([('state', '=', 'pending')])

    for record in records:
        try:
            # Use with_context for isolation
            record.with_context(from_cron=True)._process()

            # Commit after each to preserve progress
            self.env.cr.commit()

        except Exception as e:
            _logger.error("Failed to process %s: %s", record.id, e)
            self.env.cr.rollback()
```

### Multi-Company Cron
```python
@api.model
def _cron_multi_company(self):
    """Process for all companies."""
    companies = self.env['res.company'].search([])

    for company in companies:
        self.with_company(company)._process_company()
        self.env.cr.commit()

def _process_company(self):
    """Process for current company context."""
    records = self.search([
        ('company_id', '=', self.env.company.id),
    ])
    # Process records
```

---

## Cache and Invalidation

### Cache Behavior
```python
# Records are cached in environment
record = self.env['my.model'].browse(1)
name1 = record.name  # DB query
name2 = record.name  # From cache

# Invalidate specific field
self.env['my.model'].invalidate_model(['name'])

# Invalidate all cache
self.env.invalidate_all()

# Clear all caches
self.env.cache.clear()
```

### Refresh from Database
```python
# Invalidate and re-read
record.invalidate_recordset()
fresh_value = record.name

# Or browse again
record = self.env['my.model'].browse(record.id)
```

---

## Best Practices

### 1. Don't Modify Context Directly
```python
# Bad
self.env.context['key'] = value

# Good
self = self.with_context(key=value)
```

### 2. Use sudo() Sparingly
```python
# Only when necessary, with minimal scope
partner = self.sudo().partner_id  # Just read related
partner.sudo().write({'internal': True})  # Just this write
```

### 3. Preserve Context in Overrides
```python
def create(self, vals):
    # Context is automatically preserved
    return super().create(vals)

# If you need to modify:
def create(self, vals):
    self = self.with_context(creating=True)
    return super(MyModel, self).create(vals)
```

### 4. Use ensure_one() for Single Records
```python
def action_confirm(self):
    self.ensure_one()  # Raises if not exactly one record
    # Process single record
```

### 5. Handle Empty Recordsets
```python
def get_partner_name(self):
    partner = self.partner_id
    # Bad - error if no partner
    return partner.name

    # Good
    return partner.name if partner else ''
```

### 6. Check Record Existence
```python
record = self.env['my.model'].browse(potentially_deleted_id)
if record.exists():
    # Safe to use
    pass
```

---


## Source: external-api-patterns.md

# External API Integration Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  EXTERNAL API INTEGRATION PATTERNS                                           ║
║  Connecting to third-party services, REST/SOAP APIs, and webhooks            ║
║  Use for payment gateways, shipping providers, CRM sync, etc.                ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Configuration Model

### API Credentials Storage
```python
from odoo import api, fields, models
from odoo.exceptions import ValidationError
import requests


class ExternalAPIConfig(models.Model):
    _name = 'external.api.config'
    _description = 'External API Configuration'

    name = fields.Char(string='Name', required=True)
    api_url = fields.Char(string='API URL', required=True)
    api_key = fields.Char(string='API Key', groups='base.group_system')
    api_secret = fields.Char(string='API Secret', groups='base.group_system')
    environment = fields.Selection(
        selection=[
            ('sandbox', 'Sandbox'),
            ('production', 'Production'),
        ],
        string='Environment',
        default='sandbox',
        required=True,
    )
    company_id = fields.Many2one(
        comodel_name='res.company',
        string='Company',
        default=lambda self: self.env.company,
    )
    active = fields.Boolean(default=True)
    last_sync = fields.Datetime(string='Last Sync', readonly=True)

    @api.constrains('api_url')
    def _check_api_url(self):
        for config in self:
            if not config.api_url.startswith(('http://', 'https://')):
                raise ValidationError("API URL must start with http:// or https://")

    def action_test_connection(self):
        """Test API connection."""
        self.ensure_one()
        try:
            response = self._make_request('GET', '/health')
            if response.status_code == 200:
                return {
                    'type': 'ir.actions.client',
                    'tag': 'display_notification',
                    'params': {
                        'title': 'Success',
                        'message': 'Connection successful!',
                        'type': 'success',
                    }
                }
        except Exception as e:
            raise ValidationError(f"Connection failed: {str(e)}")
```

### System Parameters (Alternative)
```python
# Store in ir.config_parameter
def _get_api_key(self):
    """Get API key from system parameters."""
    return self.env['ir.config_parameter'].sudo().get_param(
        'my_module.api_key', default=''
    )

def _set_api_key(self, value):
    """Set API key in system parameters."""
    self.env['ir.config_parameter'].sudo().set_param(
        'my_module.api_key', value
    )
```

---

## HTTP Client Mixin

### Reusable API Client
```python
import json
import logging
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

_logger = logging.getLogger(__name__)


class APIClientMixin(models.AbstractModel):
    _name = 'api.client.mixin'
    _description = 'API Client Mixin'

    def _get_session(self):
        """Get requests session with retry logic."""
        session = requests.Session()

        retries = Retry(
            total=3,
            backoff_factor=0.5,
            status_forcelist=[500, 502, 503, 504],
        )
        adapter = HTTPAdapter(max_retries=retries)
        session.mount('http://', adapter)
        session.mount('https://', adapter)

        return session

    def _get_headers(self):
        """Get default headers."""
        config = self._get_api_config()
        return {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {config.api_key}',
            'X-API-Version': '2024-01',
        }

    def _get_api_config(self):
        """Get API configuration for current company."""
        config = self.env['external.api.config'].search([
            ('company_id', '=', self.env.company.id),
            ('active', '=', True),
        ], limit=1)

        if not config:
            raise ValidationError("No API configuration found for this company.")

        return config

    def _make_request(self, method, endpoint, data=None, params=None):
        """Make HTTP request to external API."""
        config = self._get_api_config()
        url = f"{config.api_url.rstrip('/')}/{endpoint.lstrip('/')}"

        session = self._get_session()
        headers = self._get_headers()

        _logger.info("API Request: %s %s", method, url)

        try:
            response = session.request(
                method=method,
                url=url,
                headers=headers,
                json=data,
                params=params,
                timeout=30,
            )

            _logger.info("API Response: %s", response.status_code)

            if response.status_code >= 400:
                self._handle_error(response)

            return response

        except requests.Timeout:
            _logger.error("API Timeout: %s", url)
            raise ValidationError("API request timed out. Please try again.")

        except requests.RequestException as e:
            _logger.error("API Error: %s", str(e))
            raise ValidationError(f"API request failed: {str(e)}")

    def _handle_error(self, response):
        """Handle API error response."""
        try:
            error_data = response.json()
            message = error_data.get('message', response.text)
        except json.JSONDecodeError:
            message = response.text

        _logger.error("API Error %s: %s", response.status_code, message)

        if response.status_code == 401:
            raise ValidationError("Authentication failed. Check your API credentials.")
        elif response.status_code == 403:
            raise ValidationError("Access forbidden. Check your API permissions.")
        elif response.status_code == 404:
            raise ValidationError("Resource not found.")
        elif response.status_code == 429:
            raise ValidationError("Rate limit exceeded. Please try again later.")
        else:
            raise ValidationError(f"API Error ({response.status_code}): {message}")
```

---

## Sync Patterns

### Pull Sync (Import from External)
```python
class ExternalProduct(models.Model):
    _name = 'external.product'
    _description = 'External Product Sync'
    _inherit = ['api.client.mixin']

    external_id = fields.Char(string='External ID', index=True)
    product_id = fields.Many2one('product.product', string='Odoo Product')
    sync_date = fields.Datetime(string='Last Sync')
    sync_status = fields.Selection([
        ('pending', 'Pending'),
        ('synced', 'Synced'),
        ('error', 'Error'),
    ], default='pending')
    sync_error = fields.Text(string='Sync Error')

    @api.model
    def _cron_sync_products(self):
        """Cron job to sync products from external API."""
        _logger.info("Starting product sync from external API")

        try:
            response = self._make_request('GET', '/products', params={
                'updated_since': self._get_last_sync_date(),
                'limit': 100,
            })
            products = response.json().get('data', [])

            for product_data in products:
                self._sync_single_product(product_data)

            _logger.info("Synced %d products", len(products))

        except Exception as e:
            _logger.error("Product sync failed: %s", str(e))

    def _sync_single_product(self, data):
        """Sync single product from external data."""
        external_id = str(data['id'])

        # Find or create mapping
        mapping = self.search([('external_id', '=', external_id)], limit=1)
        if not mapping:
            mapping = self.create({'external_id': external_id})

        try:
            # Find or create Odoo product
            product = mapping.product_id
            if not product:
                product = self.env['product.product'].create({
                    'name': data['name'],
                    'default_code': data.get('sku'),
                    'list_price': data.get('price', 0),
                })
                mapping.product_id = product
            else:
                product.write({
                    'name': data['name'],
                    'list_price': data.get('price', 0),
                })

            mapping.write({
                'sync_date': fields.Datetime.now(),
                'sync_status': 'synced',
                'sync_error': False,
            })

        except Exception as e:
            mapping.write({
                'sync_status': 'error',
                'sync_error': str(e),
            })
            _logger.error("Failed to sync product %s: %s", external_id, str(e))
```

### Push Sync (Export to External)
```python
class ResPartner(models.Model):
    _inherit = 'res.partner'

    external_customer_id = fields.Char(string='External Customer ID')
    sync_to_external = fields.Boolean(string='Sync to External', default=True)

    def write(self, vals):
        """Override write to trigger external sync."""
        result = super().write(vals)

        # Sync if relevant fields changed
        sync_fields = {'name', 'email', 'phone', 'street', 'city'}
        if self.sync_to_external and sync_fields & set(vals.keys()):
            self._sync_to_external_api()

        return result

    def _sync_to_external_api(self):
        """Push customer data to external API."""
        for partner in self:
            if not partner.sync_to_external:
                continue

            data = {
                'name': partner.name,
                'email': partner.email,
                'phone': partner.phone,
                'address': {
                    'street': partner.street,
                    'city': partner.city,
                    'zip': partner.zip,
                    'country': partner.country_id.code,
                },
            }

            try:
                api_client = self.env['api.client.mixin']

                if partner.external_customer_id:
                    # Update existing
                    response = api_client._make_request(
                        'PUT',
                        f'/customers/{partner.external_customer_id}',
                        data=data
                    )
                else:
                    # Create new
                    response = api_client._make_request(
                        'POST', '/customers', data=data
                    )
                    result = response.json()
                    partner.external_customer_id = result['id']

            except Exception as e:
                _logger.error("Failed to sync partner %s: %s", partner.id, str(e))
```

### Bidirectional Sync
```python
class SyncManager(models.Model):
    _name = 'sync.manager'
    _description = 'Bidirectional Sync Manager'
    _inherit = ['api.client.mixin']

    @api.model
    def _cron_full_sync(self):
        """Full bidirectional sync."""
        self._pull_changes()
        self._push_changes()

    def _pull_changes(self):
        """Pull changes from external system."""
        last_sync = self._get_last_sync_timestamp('pull')

        response = self._make_request('GET', '/changes', params={
            'since': last_sync,
            'types': 'customer,product,order',
        })

        for change in response.json().get('changes', []):
            self._process_incoming_change(change)

        self._set_last_sync_timestamp('pull')

    def _push_changes(self):
        """Push local changes to external system."""
        # Get records modified since last push
        last_sync = self._get_last_sync_timestamp('push')

        modified_partners = self.env['res.partner'].search([
            ('write_date', '>', last_sync),
            ('sync_to_external', '=', True),
        ])

        for partner in modified_partners:
            partner._sync_to_external_api()

        self._set_last_sync_timestamp('push')
```

---

## Webhook Handling

### Incoming Webhooks
```python
from odoo import http
from odoo.http import request
import hmac
import hashlib


class WebhookController(http.Controller):

    @http.route('/webhook/external', type='json', auth='none',
                methods=['POST'], csrf=False)
    def handle_webhook(self):
        """Handle incoming webhook from external service."""
        # Verify signature
        signature = request.httprequest.headers.get('X-Signature')
        if not self._verify_signature(signature):
            return {'error': 'Invalid signature'}, 401

        data = request.jsonrequest
        event_type = data.get('event')

        _logger.info("Received webhook: %s", event_type)

        try:
            if event_type == 'customer.created':
                self._handle_customer_created(data['payload'])
            elif event_type == 'customer.updated':
                self._handle_customer_updated(data['payload'])
            elif event_type == 'order.completed':
                self._handle_order_completed(data['payload'])
            else:
                _logger.warning("Unknown webhook event: %s", event_type)

            return {'status': 'success'}

        except Exception as e:
            _logger.error("Webhook processing failed: %s", str(e))
            return {'status': 'error', 'message': str(e)}

    def _verify_signature(self, signature):
        """Verify webhook signature."""
        if not signature:
            return False

        secret = request.env['ir.config_parameter'].sudo().get_param(
            'my_module.webhook_secret'
        )
        if not secret:
            return False

        raw_body = request.httprequest.get_data()
        expected = hmac.new(
            secret.encode(),
            raw_body,
            hashlib.sha256
        ).hexdigest()

        return hmac.compare_digest(signature, expected)

    def _handle_customer_created(self, payload):
        """Process customer creation webhook."""
        partner = request.env['res.partner'].sudo().create({
            'name': payload['name'],
            'email': payload['email'],
            'external_customer_id': payload['id'],
        })
        _logger.info("Created partner %s from webhook", partner.id)
```

### Outgoing Webhooks
```python
class WebhookSender(models.Model):
    _name = 'webhook.sender'
    _description = 'Outgoing Webhook Sender'

    @api.model
    def send_webhook(self, event_type, payload, url=None):
        """Send webhook to external endpoint."""
        if not url:
            url = self.env['ir.config_parameter'].sudo().get_param(
                'my_module.webhook_url'
            )

        if not url:
            _logger.warning("No webhook URL configured")
            return False

        data = {
            'event': event_type,
            'timestamp': fields.Datetime.now().isoformat(),
            'payload': payload,
        }

        # Sign the payload
        secret = self.env['ir.config_parameter'].sudo().get_param(
            'my_module.webhook_secret'
        )
        signature = hmac.new(
            secret.encode(),
            json.dumps(data).encode(),
            hashlib.sha256
        ).hexdigest()

        headers = {
            'Content-Type': 'application/json',
            'X-Signature': signature,
        }

        try:
            response = requests.post(
                url, json=data, headers=headers, timeout=10
            )
            response.raise_for_status()
            _logger.info("Webhook sent successfully: %s", event_type)
            return True

        except Exception as e:
            _logger.error("Webhook failed: %s", str(e))
            # Queue for retry
            self._queue_webhook_retry(event_type, payload, url)
            return False
```

---

## OAuth2 Integration

### OAuth2 Token Management
```python
from datetime import timedelta


class OAuth2Config(models.Model):
    _name = 'oauth2.config'
    _description = 'OAuth2 Configuration'

    name = fields.Char(string='Name', required=True)
    client_id = fields.Char(string='Client ID', required=True)
    client_secret = fields.Char(
        string='Client Secret',
        required=True,
        groups='base.group_system',
    )
    auth_url = fields.Char(string='Authorization URL')
    token_url = fields.Char(string='Token URL', required=True)
    scope = fields.Char(string='Scope')

    access_token = fields.Char(string='Access Token', groups='base.group_system')
    refresh_token = fields.Char(string='Refresh Token', groups='base.group_system')
    token_expiry = fields.Datetime(string='Token Expiry')

    def get_access_token(self):
        """Get valid access token, refreshing if needed."""
        self.ensure_one()

        if self.access_token and self.token_expiry:
            if fields.Datetime.now() < self.token_expiry - timedelta(minutes=5):
                return self.access_token

        # Token expired or missing, refresh
        if self.refresh_token:
            self._refresh_token()
        else:
            self._get_new_token()

        return self.access_token

    def _refresh_token(self):
        """Refresh the access token."""
        response = requests.post(self.token_url, data={
            'grant_type': 'refresh_token',
            'refresh_token': self.refresh_token,
            'client_id': self.client_id,
            'client_secret': self.client_secret,
        })

        if response.status_code != 200:
            raise ValidationError("Token refresh failed")

        self._process_token_response(response.json())

    def _get_new_token(self):
        """Get new token using client credentials."""
        response = requests.post(self.token_url, data={
            'grant_type': 'client_credentials',
            'client_id': self.client_id,
            'client_secret': self.client_secret,
            'scope': self.scope,
        })

        if response.status_code != 200:
            raise ValidationError("Token acquisition failed")

        self._process_token_response(response.json())

    def _process_token_response(self, data):
        """Process token response and store tokens."""
        expires_in = data.get('expires_in', 3600)
        self.write({
            'access_token': data['access_token'],
            'refresh_token': data.get('refresh_token', self.refresh_token),
            'token_expiry': fields.Datetime.now() + timedelta(seconds=expires_in),
        })
```

---

## Rate Limiting

### Rate Limiter
```python
import time
from collections import deque


class RateLimiter:
    """Simple rate limiter for API calls."""

    def __init__(self, max_calls, period):
        self.max_calls = max_calls
        self.period = period  # seconds
        self.calls = deque()

    def wait_if_needed(self):
        """Wait if rate limit would be exceeded."""
        now = time.time()

        # Remove old calls outside the window
        while self.calls and self.calls[0] < now - self.period:
            self.calls.popleft()

        if len(self.calls) >= self.max_calls:
            sleep_time = self.period - (now - self.calls[0])
            if sleep_time > 0:
                _logger.info("Rate limit reached, sleeping %.2fs", sleep_time)
                time.sleep(sleep_time)

        self.calls.append(time.time())


# Usage in API client
class APIClient(models.AbstractModel):
    _name = 'api.client'
    _rate_limiter = RateLimiter(max_calls=100, period=60)

    def _make_request(self, method, endpoint, **kwargs):
        self._rate_limiter.wait_if_needed()
        # ... rest of request logic
```

---

## Error Handling & Retry

### Retry with Exponential Backoff
```python
import time
from functools import wraps


def retry_on_failure(max_retries=3, backoff_factor=2):
    """Decorator for retry with exponential backoff."""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            last_exception = None
            for attempt in range(max_retries):
                try:
                    return func(*args, **kwargs)
                except (requests.Timeout, requests.ConnectionError) as e:
                    last_exception = e
                    if attempt < max_retries - 1:
                        sleep_time = backoff_factor ** attempt
                        _logger.warning(
                            "Attempt %d failed, retrying in %ds: %s",
                            attempt + 1, sleep_time, str(e)
                        )
                        time.sleep(sleep_time)
            raise last_exception
        return wrapper
    return decorator


class APIClientWithRetry(models.AbstractModel):
    _name = 'api.client.retry'

    @retry_on_failure(max_retries=3, backoff_factor=2)
    def _make_request(self, method, endpoint, **kwargs):
        # Request implementation
        pass
```

---

## Best Practices

1. **Never hardcode credentials** - Use ir.config_parameter or dedicated config model
2. **Use HTTPS** - Always use secure connections
3. **Implement retry logic** - Handle transient failures
4. **Log all API calls** - For debugging and audit
5. **Handle rate limits** - Implement backoff strategies
6. **Validate responses** - Don't trust external data
7. **Use timeouts** - Prevent hanging requests
8. **Queue heavy operations** - Don't block user actions
9. **Test with sandbox** - Use environment switching
10. **Secure webhooks** - Always verify signatures

---


# Infrastructure Patterns

Consolidated from the following source files:
- `controller-api-patterns.md`
- `cron-automation-patterns.md`
- `mail-notification-patterns.md`
- `assets-bundling-patterns.md`
- `logging-debugging-patterns.md`
- `error-handling-patterns.md`
- `config-settings-patterns.md`
- `attachment-binary-patterns.md`
- `report-patterns.md`
- `multi-company-patterns.md`

---


## Source: controller-api-patterns.md

# Controller and API Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  CONTROLLER & API PATTERNS                                                   ║
║  HTTP controllers, REST endpoints, and web routes                            ║
║  Use for APIs, webhooks, and custom web endpoints                            ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## File Structure

```
my_module/
├── controllers/
│   ├── __init__.py
│   └── main.py
└── __manifest__.py
```

### __init__.py
```python
from . import main
```

---

## Basic Controller

```python
from odoo import http
from odoo.http import request


class MyController(http.Controller):

    @http.route('/my_module/hello', type='http', auth='public')
    def hello(self):
        """Simple public endpoint."""
        return "Hello, World!"

    @http.route('/my_module/data', type='json', auth='user')
    def get_data(self):
        """JSON endpoint requiring authentication."""
        records = request.env['my.model'].search([])
        return {
            'status': 'success',
            'count': len(records),
            'data': records.read(['name', 'state']),
        }
```

---

## Route Decorators

### Basic Parameters
```python
@http.route(
    route='/my_module/endpoint',
    type='http',           # 'http' or 'json'
    auth='user',           # 'public', 'user', 'none'
    methods=['GET', 'POST'],
    website=False,         # True for website controllers
    csrf=True,             # CSRF protection (default True)
)
```

### Auth Types
| Auth | Description |
|------|-------------|
| `public` | No login required, uses public user |
| `user` | Login required |
| `none` | No user context, manual handling |

### Route Parameters
```python
# URL parameters
@http.route('/my_module/record/<int:record_id>')
def get_record(self, record_id):
    record = request.env['my.model'].browse(record_id)
    return record.name

# Multiple parameters
@http.route('/my_module/<model>/<int:id>/action')
def model_action(self, model, id):
    record = request.env[model].browse(id)
    return str(record)

# Optional parameters
@http.route('/my_module/search')
def search(self, query='', limit=10, **kw):
    records = request.env['my.model'].search(
        [('name', 'ilike', query)],
        limit=int(limit)
    )
    return str(records.ids)
```

---

## HTTP Controllers

### GET Endpoint
```python
@http.route('/api/v1/records', type='http', auth='user', methods=['GET'])
def list_records(self, **kw):
    """List records with pagination."""
    limit = int(kw.get('limit', 20))
    offset = int(kw.get('offset', 0))

    records = request.env['my.model'].search(
        [], limit=limit, offset=offset
    )

    data = []
    for record in records:
        data.append({
            'id': record.id,
            'name': record.name,
            'state': record.state,
        })

    return request.make_response(
        json.dumps({'data': data}),
        headers=[('Content-Type', 'application/json')]
    )
```

### POST Endpoint
```python
@http.route('/api/v1/records', type='http', auth='user',
            methods=['POST'], csrf=False)
def create_record(self, **post):
    """Create new record."""
    try:
        record = request.env['my.model'].create({
            'name': post.get('name'),
            'description': post.get('description'),
        })

        return request.make_response(
            json.dumps({
                'status': 'success',
                'id': record.id,
            }),
            headers=[('Content-Type', 'application/json')]
        )
    except Exception as e:
        return request.make_response(
            json.dumps({
                'status': 'error',
                'message': str(e),
            }),
            status=400,
            headers=[('Content-Type', 'application/json')]
        )
```

---

## JSON-RPC Controllers

### Basic JSON Endpoint
```python
@http.route('/api/v1/json/records', type='json', auth='user')
def json_list_records(self, domain=None, fields=None, limit=100):
    """JSON-RPC endpoint for listing records."""
    domain = domain or []
    fields = fields or ['name', 'state']

    records = request.env['my.model'].search_read(
        domain, fields, limit=limit
    )

    return {
        'status': 'success',
        'count': len(records),
        'records': records,
    }
```

### JSON with Validation
```python
@http.route('/api/v1/json/create', type='json', auth='user')
def json_create(self, name, **kwargs):
    """Create record with validation."""
    if not name:
        return {
            'status': 'error',
            'message': 'Name is required',
        }

    try:
        vals = {'name': name}
        if kwargs.get('partner_id'):
            vals['partner_id'] = int(kwargs['partner_id'])

        record = request.env['my.model'].create(vals)

        return {
            'status': 'success',
            'id': record.id,
            'name': record.name,
        }
    except Exception as e:
        return {
            'status': 'error',
            'message': str(e),
        }
```

---

## REST API Pattern

### Complete CRUD Controller
```python
import json
from odoo import http
from odoo.http import request, Response


class MyAPIController(http.Controller):
    """REST API for my.model"""

    def _get_record(self, record_id):
        """Helper to get record with error handling."""
        record = request.env['my.model'].browse(record_id)
        if not record.exists():
            return None
        return record

    def _json_response(self, data, status=200):
        """Helper to create JSON response."""
        return Response(
            json.dumps(data),
            status=status,
            mimetype='application/json'
        )

    # LIST
    @http.route('/api/v1/mymodel', type='http', auth='user',
                methods=['GET'], csrf=False)
    def list(self, **kw):
        """GET /api/v1/mymodel - List all records."""
        domain = []
        if kw.get('state'):
            domain.append(('state', '=', kw['state']))

        records = request.env['my.model'].search_read(
            domain,
            ['id', 'name', 'state', 'create_date'],
            limit=int(kw.get('limit', 100)),
            offset=int(kw.get('offset', 0)),
        )

        return self._json_response({
            'status': 'success',
            'data': records,
        })

    # GET
    @http.route('/api/v1/mymodel/<int:id>', type='http', auth='user',
                methods=['GET'], csrf=False)
    def get(self, id):
        """GET /api/v1/mymodel/{id} - Get single record."""
        record = self._get_record(id)
        if not record:
            return self._json_response(
                {'status': 'error', 'message': 'Not found'},
                status=404
            )

        return self._json_response({
            'status': 'success',
            'data': {
                'id': record.id,
                'name': record.name,
                'state': record.state,
                'partner_id': record.partner_id.id,
                'partner_name': record.partner_id.name,
            },
        })

    # CREATE
    @http.route('/api/v1/mymodel', type='http', auth='user',
                methods=['POST'], csrf=False)
    def create(self, **post):
        """POST /api/v1/mymodel - Create record."""
        try:
            required = ['name']
            for field in required:
                if not post.get(field):
                    return self._json_response(
                        {'status': 'error', 'message': f'{field} is required'},
                        status=400
                    )

            vals = {
                'name': post['name'],
            }
            if post.get('partner_id'):
                vals['partner_id'] = int(post['partner_id'])

            record = request.env['my.model'].create(vals)

            return self._json_response({
                'status': 'success',
                'id': record.id,
            }, status=201)

        except Exception as e:
            return self._json_response(
                {'status': 'error', 'message': str(e)},
                status=500
            )

    # UPDATE
    @http.route('/api/v1/mymodel/<int:id>', type='http', auth='user',
                methods=['PUT', 'PATCH'], csrf=False)
    def update(self, id, **post):
        """PUT/PATCH /api/v1/mymodel/{id} - Update record."""
        record = self._get_record(id)
        if not record:
            return self._json_response(
                {'status': 'error', 'message': 'Not found'},
                status=404
            )

        try:
            vals = {}
            if 'name' in post:
                vals['name'] = post['name']
            if 'state' in post:
                vals['state'] = post['state']

            if vals:
                record.write(vals)

            return self._json_response({
                'status': 'success',
                'id': record.id,
            })

        except Exception as e:
            return self._json_response(
                {'status': 'error', 'message': str(e)},
                status=500
            )

    # DELETE
    @http.route('/api/v1/mymodel/<int:id>', type='http', auth='user',
                methods=['DELETE'], csrf=False)
    def delete(self, id):
        """DELETE /api/v1/mymodel/{id} - Delete record."""
        record = self._get_record(id)
        if not record:
            return self._json_response(
                {'status': 'error', 'message': 'Not found'},
                status=404
            )

        try:
            record.unlink()
            return self._json_response({
                'status': 'success',
                'message': 'Deleted',
            })

        except Exception as e:
            return self._json_response(
                {'status': 'error', 'message': str(e)},
                status=500
            )
```

---

## Webhook Endpoint

```python
import hmac
import hashlib

class WebhookController(http.Controller):

    @http.route('/webhook/my_module', type='json', auth='none',
                methods=['POST'], csrf=False)
    def webhook_handler(self):
        """Handle incoming webhook."""
        # Get raw data
        data = request.jsonrequest

        # Verify signature (example)
        signature = request.httprequest.headers.get('X-Signature')
        secret = request.env['ir.config_parameter'].sudo().get_param(
            'my_module.webhook_secret'
        )

        if not self._verify_signature(data, signature, secret):
            return {'status': 'error', 'message': 'Invalid signature'}

        # Process webhook
        try:
            event_type = data.get('event')
            payload = data.get('payload', {})

            if event_type == 'order.created':
                self._handle_order_created(payload)
            elif event_type == 'payment.received':
                self._handle_payment_received(payload)

            return {'status': 'success'}

        except Exception as e:
            return {'status': 'error', 'message': str(e)}

    def _verify_signature(self, data, signature, secret):
        """Verify webhook signature."""
        if not signature or not secret:
            return False
        expected = hmac.new(
            secret.encode(),
            json.dumps(data).encode(),
            hashlib.sha256
        ).hexdigest()
        return hmac.compare_digest(signature, expected)

    def _handle_order_created(self, payload):
        """Handle order created event."""
        request.env['my.model'].sudo().create({
            'name': payload.get('order_id'),
            'external_id': payload.get('id'),
        })
```

---

## File Download/Upload

### File Download
```python
@http.route('/my_module/download/<int:id>', type='http', auth='user')
def download_file(self, id):
    """Download file attachment."""
    record = request.env['my.model'].browse(id)
    if not record.exists() or not record.file:
        return request.not_found()

    return request.make_response(
        base64.b64decode(record.file),
        headers=[
            ('Content-Type', 'application/octet-stream'),
            ('Content-Disposition', f'attachment; filename="{record.filename}"'),
        ]
    )
```

### File Upload
```python
@http.route('/my_module/upload', type='http', auth='user',
            methods=['POST'], csrf=False)
def upload_file(self, **post):
    """Upload file."""
    file = post.get('file')
    if not file:
        return json.dumps({'error': 'No file provided'})

    try:
        content = base64.b64encode(file.read())
        filename = file.filename

        record = request.env['my.model'].create({
            'name': filename,
            'file': content,
            'filename': filename,
        })

        return json.dumps({
            'status': 'success',
            'id': record.id,
        })

    except Exception as e:
        return json.dumps({'error': str(e)})
```

---

## Authentication

### API Key Authentication
```python
class APIKeyController(http.Controller):

    def _check_api_key(self):
        """Validate API key from header."""
        api_key = request.httprequest.headers.get('X-API-Key')
        if not api_key:
            return False

        valid_key = request.env['ir.config_parameter'].sudo().get_param(
            'my_module.api_key'
        )
        return api_key == valid_key

    @http.route('/api/secure/data', type='json', auth='none', csrf=False)
    def secure_endpoint(self):
        """Endpoint with API key auth."""
        if not self._check_api_key():
            return {'error': 'Invalid API key'}, 401

        # Process request with sudo (no user context)
        data = request.env['my.model'].sudo().search_read([], ['name'])
        return {'data': data}
```

---

## Best Practices

1. **Use appropriate auth** - `public` for public APIs, `user` for authenticated
2. **Handle errors gracefully** - Return proper HTTP status codes
3. **Validate input** - Check required fields and types
4. **Use sudo carefully** - Only when necessary for public endpoints
5. **CSRF protection** - Disable only for legitimate API endpoints
6. **Rate limiting** - Implement for public APIs
7. **Logging** - Log API requests for debugging
8. **Documentation** - Document endpoints for consumers

---


## Source: cron-automation-patterns.md

# Scheduled Actions and Automation Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  CRON & AUTOMATION PATTERNS                                                  ║
║  Scheduled actions, server actions, and automated rules                      ║
║  Use for background jobs, triggers, and workflow automation                  ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Scheduled Actions (Cron Jobs)

### Basic Cron Definition (XML)
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="ir_cron_process_pending" model="ir.cron">
        <field name="name">My Module: Process Pending Records</field>
        <field name="model_id" ref="model_my_model"/>
        <field name="state">code</field>
        <field name="code">model._cron_process_pending()</field>
        <field name="interval_number">1</field>
        <field name="interval_type">hours</field>
        <field name="numbercall">-1</field>
        <field name="active">True</field>
        <field name="doall">False</field>
    </record>
</odoo>
```

### Interval Types
| Type | Description |
|------|-------------|
| `minutes` | Run every X minutes |
| `hours` | Run every X hours |
| `days` | Run every X days |
| `weeks` | Run every X weeks |
| `months` | Run every X months |

### Cron Attributes
| Attribute | Description |
|-----------|-------------|
| `interval_number` | Number of intervals |
| `interval_type` | Type of interval |
| `numbercall` | -1 for infinite, or count |
| `active` | Enable/disable cron |
| `doall` | Run missed executions |
| `nextcall` | Next execution datetime |
| `priority` | Execution priority (lower = first) |

---

## Python Cron Methods (v18)

### Basic Cron Method
```python
from odoo import api, models
import logging

_logger = logging.getLogger(__name__)


class MyModel(models.Model):
    _name = 'my.model'

    @api.model
    def _cron_process_pending(self) -> None:
        """Process pending records - called by scheduled action."""
        _logger.info("Starting cron: process pending records")

        records = self.search([('state', '=', 'pending')], limit=100)
        _logger.info(f"Found {len(records)} pending records")

        for record in records:
            try:
                record._process_single()
            except Exception as e:
                _logger.error(f"Error processing {record.id}: {e}")

        _logger.info("Cron completed: process pending records")
```

### Batch Processing Cron
```python
@api.model
def _cron_batch_process(self) -> None:
    """Process records in batches with commits."""
    batch_size = 100
    offset = 0
    processed = 0

    while True:
        # Fetch batch
        records = self.search(
            [('state', '=', 'pending')],
            limit=batch_size,
            offset=offset,
        )

        if not records:
            break

        for record in records:
            try:
                record.with_context(from_cron=True)._do_process()
                processed += 1
            except Exception as e:
                _logger.error(f"Error on record {record.id}: {e}")

        # Commit batch and clear cache
        self.env.cr.commit()
        self.env.invalidate_all()

        offset += batch_size
        _logger.info(f"Processed {processed} records so far...")

    _logger.info(f"Batch process complete: {processed} records")
```

### Time-Limited Cron
```python
import time

@api.model
def _cron_time_limited_process(self) -> None:
    """Process with time limit to prevent long-running jobs."""
    max_duration = 300  # 5 minutes
    start_time = time.time()
    processed = 0

    records = self.search([('needs_sync', '=', True)])

    for record in records:
        # Check time limit
        if time.time() - start_time > max_duration:
            _logger.warning(
                f"Time limit reached after {processed} records. "
                f"Remaining: {len(records) - processed}"
            )
            break

        try:
            record._sync_external()
            processed += 1
        except Exception as e:
            _logger.error(f"Sync error for {record.id}: {e}")

        # Periodic commit
        if processed % 50 == 0:
            self.env.cr.commit()

    _logger.info(f"Processed {processed}/{len(records)} records")
```

---

## Server Actions

### Python Code Action
```xml
<record id="action_mark_done" model="ir.actions.server">
    <field name="name">Mark as Done</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="binding_view_types">list,form</field>
    <field name="state">code</field>
    <field name="code">
if records:
    records.write({'state': 'done'})
    </field>
</record>
```

### Multi-Record Action
```xml
<record id="action_batch_confirm" model="ir.actions.server">
    <field name="name">Confirm Selected</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="binding_view_types">list</field>
    <field name="state">code</field>
    <field name="code">
for record in records:
    if record.state == 'draft':
        record.action_confirm()
    </field>
</record>
```

### Action with Notification
```xml
<record id="action_notify_users" model="ir.actions.server">
    <field name="name">Notify Users</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="binding_model_id" ref="model_my_model"/>
    <field name="state">code</field>
    <field name="code">
count = len(records)
records.action_send_notification()
action = {
    'type': 'ir.actions.client',
    'tag': 'display_notification',
    'params': {
        'title': 'Success',
        'message': f'Notified {count} users.',
        'type': 'success',
        'sticky': False,
    }
}
    </field>
</record>
```

---

## Automated Actions (Base Automation)

### On Create Trigger
```xml
<record id="automation_on_create" model="base.automation">
    <field name="name">Auto-assign on Create</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="trigger">on_create</field>
    <field name="state">code</field>
    <field name="code">
for record in records:
    if not record.user_id:
        record.user_id = record.create_uid
    </field>
</record>
```

### On Write Trigger
```xml
<record id="automation_on_state_change" model="base.automation">
    <field name="name">Notify on State Change</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="trigger">on_write</field>
    <field name="trigger_field_ids" eval="[(6, 0, [ref('field_my_model__state')])]"/>
    <field name="filter_domain">[('state', '=', 'confirmed')]</field>
    <field name="state">code</field>
    <field name="code">
records.message_post(
    body="Record has been confirmed.",
    message_type='notification',
)
    </field>
</record>
```

### Time-Based Trigger
```xml
<record id="automation_overdue_check" model="base.automation">
    <field name="name">Mark Overdue Records</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="trigger">on_time</field>
    <field name="trg_date_id" ref="field_my_model__deadline"/>
    <field name="trg_date_range">1</field>
    <field name="trg_date_range_type">day</field>
    <field name="filter_domain">[('state', 'not in', ['done', 'cancel'])]</field>
    <field name="state">code</field>
    <field name="code">
records.write({'is_overdue': True})
records.message_post(body="This record is now overdue!")
    </field>
</record>
```

### Trigger Types
| Trigger | When Executed |
|---------|---------------|
| `on_create` | After record creation |
| `on_write` | After record update |
| `on_create_or_write` | After create or update |
| `on_unlink` | Before record deletion |
| `on_time` | Based on date field |

---

## Queue Jobs (for Heavy Processing)

### Using ir.cron with Batching
```python
@api.model
def _cron_heavy_process(self) -> None:
    """Heavy process with queue-like behavior."""
    # Get unprocessed records
    to_process = self.search([
        ('processed', '=', False),
        ('attempts', '<', 3),  # Max retry attempts
    ], limit=50)

    for record in to_process:
        try:
            record.with_context(processing=True)._heavy_operation()
            record.processed = True
            record.processed_date = fields.Datetime.now()
        except Exception as e:
            record.attempts += 1
            record.last_error = str(e)
            _logger.error(f"Processing failed for {record.id}: {e}")

        # Commit after each to preserve progress
        self.env.cr.commit()
```

### Deferred Processing Pattern
```python
class MyModel(models.Model):
    _name = 'my.model'

    process_state = fields.Selection([
        ('pending', 'Pending'),
        ('processing', 'Processing'),
        ('done', 'Done'),
        ('error', 'Error'),
    ], default='pending')
    process_error = fields.Text()

    def action_queue_for_processing(self) -> None:
        """Queue records for cron processing."""
        self.write({
            'process_state': 'pending',
            'process_error': False,
        })

    @api.model
    def _cron_process_queue(self) -> None:
        """Process queued records."""
        records = self.search([
            ('process_state', '=', 'pending')
        ], limit=20)

        for record in records:
            record.process_state = 'processing'
            self.env.cr.commit()

            try:
                record._do_heavy_work()
                record.process_state = 'done'
            except Exception as e:
                record.process_state = 'error'
                record.process_error = str(e)

            self.env.cr.commit()
```

---

## Best Practices

### 1. Logging
```python
import logging
_logger = logging.getLogger(__name__)

@api.model
def _cron_task(self) -> None:
    _logger.info("Cron started: %s", self._name)
    try:
        # Work here
        _logger.info("Cron completed successfully")
    except Exception as e:
        _logger.exception("Cron failed: %s", e)
        raise
```

### 2. Transaction Safety
```python
@api.model
def _cron_safe_process(self) -> None:
    """Process with proper transaction handling."""
    records = self.search([('pending', '=', True)])

    for record in records:
        # Use new cursor for isolation
        try:
            with self.env.cr.savepoint():
                record._process()
        except Exception as e:
            _logger.error(f"Failed {record.id}: {e}")
            # Savepoint rollback - continue with next
            continue
```

### 3. Idempotency
```python
@api.model
def _cron_idempotent_sync(self) -> None:
    """Idempotent sync - safe to run multiple times."""
    records = self.search([
        ('needs_sync', '=', True),
        ('last_sync_attempt', '<', fields.Datetime.now() - timedelta(minutes=5)),
    ])

    for record in records:
        record.last_sync_attempt = fields.Datetime.now()
        self.env.cr.commit()

        try:
            record._sync()
            record.needs_sync = False
        except Exception:
            pass  # Will retry on next run
```

### 4. Monitoring
```python
@api.model
def _cron_with_monitoring(self) -> None:
    """Cron with execution tracking."""
    start = fields.Datetime.now()

    try:
        count = self._do_work()
        status = 'success'
        error = False
    except Exception as e:
        count = 0
        status = 'error'
        error = str(e)

    # Log execution
    self.env['my.cron.log'].create({
        'cron_name': 'process_pending',
        'start_time': start,
        'end_time': fields.Datetime.now(),
        'records_processed': count,
        'status': status,
        'error_message': error,
    })
```

---

## Cron Security

### Manifest Declaration
```python
# Cron data file must be in 'data' section
'data': [
    'data/cron.xml',
]
```

### Access Rights
Crons run as the user who created them (usually admin). For specific user context:

```python
@api.model
def _cron_as_specific_user(self) -> None:
    """Run as specific user for proper access rights."""
    cron_user = self.env.ref('my_module.cron_service_user')
    self_as_user = self.with_user(cron_user)
    self_as_user._do_work()
```

---


## Source: mail-notification-patterns.md

# Mail and Notification Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MAIL & NOTIFICATION PATTERNS                                                ║
║  Email templates, chatter integration, and activity management               ║
║  Use for automated emails, discussions, and workflow notifications           ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Mail Mixin Integration

### Adding Chatter to Model
```python
from odoo import api, fields, models


class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'
    _inherit = ['mail.thread', 'mail.activity.mixin']

    name = fields.Char(string='Name', required=True, tracking=True)
    state = fields.Selection(
        selection=[
            ('draft', 'Draft'),
            ('confirmed', 'Confirmed'),
            ('done', 'Done'),
        ],
        string='Status',
        default='draft',
        tracking=True,
    )
    partner_id = fields.Many2one(
        comodel_name='res.partner',
        string='Customer',
        tracking=True,
    )
    user_id = fields.Many2one(
        comodel_name='res.users',
        string='Assigned To',
        tracking=True,
    )
    description = fields.Html(string='Description')
```

### View with Chatter
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="my_model_view_form" model="ir.ui.view">
        <field name="name">my.model.form</field>
        <field name="model">my.model</field>
        <field name="arch" type="xml">
            <form string="My Model">
                <sheet>
                    <group>
                        <field name="name"/>
                        <field name="state"/>
                        <field name="partner_id"/>
                        <field name="user_id"/>
                    </group>
                    <notebook>
                        <page string="Description">
                            <field name="description"/>
                        </page>
                    </notebook>
                </sheet>
                <!-- Chatter -->
                <chatter/>
            </form>
        </field>
    </record>
</odoo>
```

---

## Email Templates

### Template Syntax Evolution (Odoo 15+)

Starting with Odoo 15, email templates migrated from Jinja2 (Mako-style) syntax to QWeb rendering. This brought significant syntax changes and improvements:

**Before Odoo 15 (Jinja2/Mako syntax):**
```xml
<record id="email_template_example" model="mail.template">
    <field name="name">Example Template</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="subject">Order ${object.name} - ${object.state}</field>
    <field name="body_html" type="html">
<![CDATA[
<p>Dear ${object.student_id.name},</p>
<p>Your order ${object.name} is now ${object.state}.</p>
<p>Amount: ${object.amount_total}</p>
]]>
    </field>
</record>
```

**Odoo 15+ (QWeb syntax):**
```xml
<record id="email_template_example" model="mail.template">
    <field name="name">Example Template</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="subject"><t t-out="object.name"/> - <t t-out="object.state"/></field>
    <field name="body_html" type="html">
<![CDATA[
<p>Dear <t t-out="object.student_id.name"/>,</p>
<p>Your order <t t-out="object.name"/> is now <t t-out="object.state"/>.</p>
<p>Amount: <t t-out="object.amount_total"/></p>
]]>
    </field>
</record>
```

**Key Changes:**
- **Syntax**: `${expression}` → `<t t-out="expression"/>`
- **Default behavior**: `t-out` escapes HTML by default (like old `t-esc`)
- **Raw HTML**: Use `t-out` with `Markup` objects for safe unescaped rendering
- **Conditionals**: `% if` → `<t t-if="condition">`
- **Loops**: `% for` → `<t t-foreach="items" t-as="item">`

**The t-out Directive:**
The `t-out` directive was introduced in Odoo 15 as a unified replacement for:
- `t-esc` (HTML-escaped output) - deprecated but still works
- `t-raw` (unescaped output) - deprecated but still works

`t-out` escapes by default but accepts `Markup` objects for safe HTML rendering, providing better security while maintaining flexibility.

**Migration Checklist:**
- Replace `${object.field}` with `<t t-out="object.field"/>`
- Replace `${object.field or ''}` with `<t t-out="object.field or ''"/>`
- Convert `% if` blocks to `<t t-if="condition">`
- Convert `% for` loops to `<t t-foreach="items" t-as="item">`
- Update any raw HTML rendering to use `Markup` objects with `t-out`

### Basic Email Template
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <record id="email_template_my_model_confirm" model="mail.template">
        <field name="name">My Model: Confirmation</field>
        <field name="model_id" ref="model_my_model"/>
        <field name="subject">{{ object.name }} - Confirmed</field>
        <field name="email_from">{{ (object.company_id.email or user.email) }}</field>
        <field name="email_to">{{ object.partner_id.email }}</field>
        <field name="body_html" type="html">
<![CDATA[
<div style="margin: 0px; padding: 0px;">
    <p style="margin: 0px; padding: 0px; font-size: 13px;">
        Dear {{ object.partner_id.name }},
    </p>
    <br/>
    <p>
        Your request <strong>{{ object.name }}</strong> has been confirmed.
    </p>
    <br/>
    <p>
        <strong>Details:</strong>
    </p>
    <ul>
        <li>Reference: {{ object.name }}</li>
        <li>Date: {{ object.create_date.strftime('%Y-%m-%d') }}</li>
        <li>Status: {{ object.state }}</li>
    </ul>
    <br/>
    <p>
        Best regards,<br/>
        {{ object.company_id.name }}
    </p>
</div>
]]>
        </field>
        <field name="auto_delete" eval="True"/>
    </record>
</odoo>
```

### Template with Attachments
```xml
<record id="email_template_with_report" model="mail.template">
    <field name="name">My Model: Send Report</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="subject">Report: {{ object.name }}</field>
    <field name="email_from">{{ user.email }}</field>
    <field name="email_to">{{ object.partner_id.email }}</field>
    <field name="body_html" type="html">
<![CDATA[
<p>Please find attached the report for {{ object.name }}.</p>
]]>
    </field>
    <!-- Attach PDF report -->
    <field name="report_template_id" ref="my_module.report_my_model"/>
    <field name="report_name">Report_{{ object.name }}</field>
</record>
```

### Dynamic Template (Python)
```python
def _get_email_template_body(self) -> str:
    """Generate dynamic email body."""
    lines_html = ""
    for line in self.line_ids:
        lines_html += f"""
        <tr>
            <td>{line.name}</td>
            <td style="text-align: right;">{line.quantity}</td>
            <td style="text-align: right;">{line.price_unit:.2f}</td>
        </tr>
        """

    return f"""
    <p>Dear {self.partner_id.name},</p>
    <table border="1" cellpadding="5">
        <thead>
            <tr>
                <th>Description</th>
                <th>Qty</th>
                <th>Price</th>
            </tr>
        </thead>
        <tbody>
            {lines_html}
        </tbody>
    </table>
    <p>Total: {self.amount_total:.2f}</p>
    """
```

---

## Sending Emails

### Using Template
```python
def action_send_email(self) -> dict:
    """Send email using template."""
    self.ensure_one()

    template = self.env.ref('my_module.email_template_my_model_confirm')
    template.send_mail(self.id, force_send=True)

    return True
```

### Open Email Composer
```python
def action_send_email_wizard(self) -> dict:
    """Open email composer with template."""
    self.ensure_one()

    template = self.env.ref('my_module.email_template_my_model_confirm')

    return {
        'type': 'ir.actions.act_window',
        'name': 'Send Email',
        'res_model': 'mail.compose.message',
        'view_mode': 'form',
        'target': 'new',
        'context': {
            'default_model': self._name,
            'default_res_ids': self.ids,
            'default_template_id': template.id,
            'default_composition_mode': 'comment',
            'force_email': True,
        },
    }
```

### Send Without Template
```python
def action_notify_partner(self) -> None:
    """Send email without template."""
    self.ensure_one()

    mail_values = {
        'subject': f'Update: {self.name}',
        'body_html': f'<p>Your record {self.name} has been updated.</p>',
        'email_from': self.env.company.email or self.env.user.email,
        'email_to': self.partner_id.email,
        'model': self._name,
        'res_id': self.id,
    }

    mail = self.env['mail.mail'].sudo().create(mail_values)
    mail.send()
```

### Batch Email
```python
def action_send_batch_emails(self) -> None:
    """Send emails to multiple records."""
    template = self.env.ref('my_module.email_template_my_model_confirm')

    for record in self:
        if record.partner_id.email:
            template.send_mail(record.id, force_send=False)

    # Process mail queue
    self.env['mail.mail'].sudo().process_email_queue()
```

---

## Message Posting

### Post Simple Message
```python
def action_post_note(self) -> None:
    """Post internal note."""
    self.ensure_one()

    self.message_post(
        body="This is an internal note.",
        message_type='comment',
        subtype_xmlid='mail.mt_note',
    )
```

### Post with Tracking
```python
def action_confirm(self) -> None:
    """Confirm and post message."""
    self.ensure_one()

    old_state = self.state
    self.write({'state': 'confirmed'})

    # Post message with details
    self.message_post(
        body=f"Record confirmed. State changed from {old_state} to confirmed.",
        message_type='notification',
        subtype_xmlid='mail.mt_comment',
    )
```

### Post with Attachments
```python
def action_post_with_attachment(self) -> None:
    """Post message with file attachment."""
    self.ensure_one()

    attachment = self.env['ir.attachment'].create({
        'name': 'document.pdf',
        'type': 'binary',
        'datas': self.document_file,  # base64 encoded
        'res_model': self._name,
        'res_id': self.id,
    })

    self.message_post(
        body="Document attached for review.",
        attachment_ids=[attachment.id],
    )
```

### Post from Template
```python
def action_post_from_template(self) -> None:
    """Post message using template."""
    self.ensure_one()

    template = self.env.ref('my_module.email_template_my_model_confirm')

    self.message_post_with_source(
        source_ref=template,
        subtype_xmlid='mail.mt_comment',
    )
```

---

## Followers and Subscriptions

### Add Followers
```python
def action_add_followers(self) -> None:
    """Add partners as followers."""
    self.ensure_one()

    partners_to_add = self.team_id.member_ids.mapped('partner_id')
    self.message_subscribe(partner_ids=partners_to_add.ids)
```

### Remove Followers
```python
def action_remove_follower(self, partner_id: int) -> None:
    """Remove specific follower."""
    self.ensure_one()
    self.message_unsubscribe(partner_ids=[partner_id])
```

### Custom Subtypes
```xml
<!-- data/mail_subtype.xml -->
<odoo>
    <!-- Subtype for confirmed notifications -->
    <record id="mt_my_model_confirmed" model="mail.message.subtype">
        <field name="name">Confirmed</field>
        <field name="res_model">my.model</field>
        <field name="default" eval="True"/>
        <field name="description">Record has been confirmed</field>
    </record>

    <!-- Subtype for assignment -->
    <record id="mt_my_model_assigned" model="mail.message.subtype">
        <field name="name">Assigned</field>
        <field name="res_model">my.model</field>
        <field name="default" eval="False"/>
        <field name="description">Record has been assigned</field>
    </record>
</odoo>
```

### Use Custom Subtype
```python
def action_confirm(self) -> None:
    """Confirm with custom notification."""
    self.ensure_one()
    self.write({'state': 'confirmed'})

    self.message_post(
        body="Record confirmed.",
        subtype_xmlid='my_module.mt_my_model_confirmed',
    )
```

---

## Activities

### Schedule Activity
```python
def action_schedule_followup(self) -> None:
    """Schedule follow-up activity."""
    self.ensure_one()

    self.activity_schedule(
        'mail.mail_activity_data_todo',
        date_deadline=fields.Date.today() + timedelta(days=7),
        summary='Follow up with customer',
        note='Check if customer needs assistance.',
        user_id=self.user_id.id,
    )
```

### Schedule with Feedback
```python
def action_request_approval(self) -> None:
    """Request approval via activity."""
    self.ensure_one()

    activity_type = self.env.ref('mail.mail_activity_data_todo')

    self.activity_schedule(
        activity_type_id=activity_type.id,
        date_deadline=fields.Date.today() + timedelta(days=3),
        summary='Approval Required',
        note=f'Please review and approve: {self.name}',
        user_id=self.env.ref('base.user_admin').id,
    )
```

### Mark Activity Done
```python
def action_mark_activity_done(self) -> None:
    """Mark all activities as done."""
    self.ensure_one()

    activities = self.activity_ids.filtered(
        lambda a: a.activity_type_id.name == 'To Do'
    )
    activities.action_feedback(feedback='Completed by workflow.')
```

### Custom Activity Type
```xml
<!-- data/mail_activity_type.xml -->
<odoo>
    <record id="mail_activity_type_review" model="mail.activity.type">
        <field name="name">Review Required</field>
        <field name="summary">Review this record</field>
        <field name="res_model">my.model</field>
        <field name="icon">fa-check-square</field>
        <field name="delay_count">3</field>
        <field name="delay_unit">days</field>
        <field name="default_user_id" ref="base.user_admin"/>
    </record>
</odoo>
```

---

## Automated Notifications

### On State Change (Automated Action)
```xml
<record id="automation_notify_on_confirm" model="base.automation">
    <field name="name">Notify on Confirmation</field>
    <field name="model_id" ref="model_my_model"/>
    <field name="trigger">on_write</field>
    <field name="trigger_field_ids" eval="[(6, 0, [ref('field_my_model__state')])]"/>
    <field name="filter_domain">[('state', '=', 'confirmed')]</field>
    <field name="state">code</field>
    <field name="code">
template = env.ref('my_module.email_template_my_model_confirm')
for record in records:
    if record.partner_id.email:
        template.send_mail(record.id)
    </field>
</record>
```

### Override Tracking (Python)
```python
def _track_subtype(self, init_values) -> str:
    """Return subtype for tracking notifications."""
    self.ensure_one()

    if 'state' in init_values:
        if self.state == 'confirmed':
            return self.env.ref('my_module.mt_my_model_confirmed')
        elif self.state == 'done':
            return self.env.ref('my_module.mt_my_model_done')

    return super()._track_subtype(init_values)
```

### Custom Notification Logic
```python
def _notify_get_recipients(self, message, msg_vals, **kwargs):
    """Override to customize notification recipients."""
    recipients = super()._notify_get_recipients(message, msg_vals, **kwargs)

    # Add manager to important notifications
    if self.state == 'confirmed' and self.amount_total > 10000:
        manager = self.env.ref('my_module.group_manager').users
        for user in manager:
            if user.partner_id.id not in [r['id'] for r in recipients]:
                recipients.append({
                    'id': user.partner_id.id,
                    'active': True,
                    'share': False,
                    'notif': 'email',
                    'type': 'user',
                })

    return recipients
```

---

## In-App Notifications

### Display Notification
```python
def action_with_notification(self) -> dict:
    """Action with success notification."""
    self.ensure_one()

    # Do something
    self.write({'state': 'done'})

    return {
        'type': 'ir.actions.client',
        'tag': 'display_notification',
        'params': {
            'title': 'Success',
            'message': f'{self.name} has been processed.',
            'type': 'success',  # success, warning, danger, info
            'sticky': False,
            'next': {'type': 'ir.actions.act_window_close'},
        }
    }
```

### Notification with Link
```python
def action_notify_with_link(self) -> dict:
    """Notification with clickable link."""
    return {
        'type': 'ir.actions.client',
        'tag': 'display_notification',
        'params': {
            'title': 'Record Created',
            'message': 'Click to view the new record.',
            'type': 'success',
            'links': [{
                'label': self.name,
                'url': f'/web#id={self.id}&model={self._name}&view_type=form',
            }],
        }
    }
```

---

## Bus Notifications (Real-time)

### Send Bus Notification
```python
def action_notify_users(self) -> None:
    """Send real-time notification via bus."""
    self.ensure_one()

    # Notify specific user
    self.env['bus.bus']._sendone(
        self.user_id.partner_id,
        'simple_notification',
        {
            'title': 'New Assignment',
            'message': f'You have been assigned to {self.name}',
            'type': 'info',
            'sticky': False,
        }
    )
```

### Broadcast to Channel
```python
def action_broadcast(self) -> None:
    """Broadcast to all users in group."""
    channel = f'my_module_{self.env.company.id}'

    self.env['bus.bus']._sendone(
        channel,
        'my_module/notification',
        {
            'record_id': self.id,
            'message': f'Record {self.name} updated',
        }
    )
```

---

## Best Practices

1. **Use templates** - Define email templates in XML for maintainability
2. **Handle missing emails** - Always check `partner_id.email` before sending
3. **Use queues** - Set `force_send=False` for batch operations
4. **Track important fields** - Add `tracking=True` to key fields
5. **Custom subtypes** - Create subtypes for different notification types
6. **Activity scheduling** - Use activities for task management
7. **Follower management** - Auto-subscribe relevant parties
8. **Test email rendering** - Verify templates render correctly

---

## Manifest Dependencies

```python
{
    'depends': [
        'mail',  # Required for mail.thread
    ],
    'data': [
        'data/mail_template.xml',
        'data/mail_subtype.xml',
        'data/mail_activity_type.xml',
        'views/my_model_views.xml',
    ],
}
```

---


## Source: assets-bundling-patterns.md

# Assets Bundling Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ASSETS BUNDLING PATTERNS                                                    ║
║  JavaScript, CSS, and SCSS asset management                                  ║
║  Use for frontend customization and OWL component registration               ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Asset Bundles Overview

### Main Bundles
| Bundle | Used In | Purpose |
|--------|---------|---------|
| `web.assets_backend` | Backend UI | JS/CSS for Odoo backend |
| `web.assets_frontend` | Website | JS/CSS for public website |
| `web.assets_common` | Both | Shared resources |
| `web.assets_qweb` | Both | QWeb templates |
| `point_of_sale.assets` | POS | Point of Sale |

---

## Manifest Asset Declaration

### Basic Assets in __manifest__.py
```python
{
    'name': 'My Module',
    'version': '17.0.1.0.0',
    'assets': {
        # Backend assets
        'web.assets_backend': [
            'my_module/static/src/js/**/*.js',
            'my_module/static/src/css/**/*.css',
            'my_module/static/src/scss/**/*.scss',
            'my_module/static/src/xml/**/*.xml',
        ],

        # Website/frontend assets
        'web.assets_frontend': [
            'my_module/static/src/frontend/js/*.js',
            'my_module/static/src/frontend/css/*.css',
        ],

        # QWeb templates for OWL components
        'web.assets_qweb': [
            'my_module/static/src/xml/*.xml',
        ],
    },
}
```

### Odoo 16+ Asset Declaration
```python
{
    'assets': {
        'web.assets_backend': [
            # Include specific files
            'my_module/static/src/js/my_component.js',

            # Include all files with glob pattern
            'my_module/static/src/**/*.js',
            'my_module/static/src/**/*.xml',
            'my_module/static/src/**/*.scss',

            # Prepend (load before others)
            ('prepend', 'my_module/static/src/js/early_load.js'),

            # After specific file
            ('after', 'web/static/src/core/main.js', 'my_module/static/src/js/after_main.js'),

            # Before specific file
            ('before', 'web/static/src/core/main.js', 'my_module/static/src/js/before_main.js'),

            # Replace existing file
            ('replace', 'other_module/static/src/js/old.js', 'my_module/static/src/js/new.js'),

            # Remove file
            ('remove', 'other_module/static/src/js/unwanted.js'),
        ],
    },
}
```

---

## Directory Structure

### Recommended Layout
```
my_module/
├── static/
│   ├── description/
│   │   └── icon.png          # Module icon (256x256)
│   └── src/
│       ├── js/               # JavaScript files
│       │   ├── my_widget.js
│       │   └── my_component.js
│       ├── css/              # Plain CSS
│       │   └── my_styles.css
│       ├── scss/             # SCSS files
│       │   └── my_styles.scss
│       ├── xml/              # QWeb templates
│       │   └── my_templates.xml
│       ├── img/              # Images
│       │   └── logo.png
│       └── fonts/            # Custom fonts
│           └── custom.woff2
```

---

## JavaScript Patterns

### ES6 Module (Odoo 14+)
```javascript
/** @odoo-module **/

import { registry } from "@web/core/registry";
import { Component } from "@odoo/owl";

export class MyComponent extends Component {
    static template = "my_module.MyComponent";

    setup() {
        // Component setup
    }
}

// Register component
registry.category("actions").add("my_module.my_action", MyComponent);
```

### Service Registration
```javascript
/** @odoo-module **/

import { registry } from "@web/core/registry";

const myService = {
    dependencies: ["orm", "notification"],

    start(env, { orm, notification }) {
        return {
            async doSomething(recordId) {
                const result = await orm.call("my.model", "my_method", [recordId]);
                notification.add("Done!", { type: "success" });
                return result;
            },
        };
    },
};

registry.category("services").add("myService", myService);
```

### Widget Extension
```javascript
/** @odoo-module **/

import { patch } from "@web/core/utils/patch";
import { FormController } from "@web/views/form/form_controller";

patch(FormController.prototype, {
    setup() {
        super.setup();
        // Additional setup
    },

    async onRecordSaved(record) {
        await super.onRecordSaved(record);
        // Custom logic after save
        console.log("Record saved:", record.resId);
    },
});
```

---

## CSS/SCSS Patterns

### Basic SCSS
```scss
// my_module/static/src/scss/my_styles.scss

// Use Odoo variables
@import "bootstrap/scss/functions";
@import "bootstrap/scss/variables";

// Module namespace
.o_my_module {
    // Component styles
    .my-card {
        border-radius: $border-radius;
        padding: $spacer;
        background: $white;
        box-shadow: $box-shadow-sm;

        &-header {
            font-weight: $font-weight-bold;
            border-bottom: 1px solid $border-color;
        }

        &-body {
            padding: $spacer;
        }
    }

    // Form customization
    .o_form_view {
        .my-custom-field {
            background-color: $light;
        }
    }

    // Kanban customization
    .o_kanban_view {
        .o_kanban_record {
            &.my-highlight {
                border-left: 3px solid $primary;
            }
        }
    }
}
```

### Dark Mode Support
```scss
// Support dark mode (Odoo 16+)
.o_my_module {
    .my-component {
        background: var(--o-view-background-color);
        color: var(--o-main-text-color);
        border-color: var(--o-color-border);
    }
}

// Or using media query
@media (prefers-color-scheme: dark) {
    .o_my_module {
        .my-component {
            background: #1e1e1e;
        }
    }
}

// Using Odoo's dark mode class
html.dark {
    .o_my_module {
        .my-component {
            background: #2d2d2d;
        }
    }
}
```

---

## QWeb Templates

### OWL Component Template
```xml
<?xml version="1.0" encoding="UTF-8"?>
<templates xml:space="preserve">

    <t t-name="my_module.MyComponent">
        <div class="o_my_module my-component">
            <div class="my-component-header">
                <h3 t-esc="props.title"/>
            </div>
            <div class="my-component-body">
                <t t-if="state.loading">
                    <span class="fa fa-spinner fa-spin"/> Loading...
                </t>
                <t t-else="">
                    <t t-foreach="state.items" t-as="item" t-key="item.id">
                        <div class="item" t-on-click="() => this.onItemClick(item)">
                            <t t-esc="item.name"/>
                        </div>
                    </t>
                </t>
            </div>
            <div class="my-component-footer">
                <button class="btn btn-primary" t-on-click="onSave">
                    Save
                </button>
            </div>
        </div>
    </t>

</templates>
```

### Extending Existing Templates
```xml
<?xml version="1.0" encoding="UTF-8"?>
<templates xml:space="preserve">

    <!-- Extend kanban record -->
    <t t-inherit="web.KanbanRecord" t-inherit-mode="extension">
        <xpath expr="//div[hasclass('o_kanban_record_bottom')]" position="inside">
            <t t-if="props.record.resModel === 'my.model'">
                <div class="my-custom-info">
                    <span t-esc="record.my_field.value"/>
                </div>
            </t>
        </xpath>
    </t>

</templates>
```

---

## Legacy JavaScript (pre-Odoo 14)

### AMD Module Pattern
```javascript
odoo.define('my_module.my_widget', function (require) {
    "use strict";

    var Widget = require('web.Widget');
    var core = require('web.core');
    var _t = core._t;

    var MyWidget = Widget.extend({
        template: 'my_module.MyWidgetTemplate',
        events: {
            'click .my-button': '_onButtonClick',
        },

        init: function (parent, options) {
            this._super.apply(this, arguments);
            this.options = options || {};
        },

        start: function () {
            var self = this;
            return this._super.apply(this, arguments).then(function () {
                self._renderContent();
            });
        },

        _onButtonClick: function (ev) {
            ev.preventDefault();
            // Handle click
        },

        _renderContent: function () {
            // Render logic
        },
    });

    return MyWidget;
});
```

---

## Asset Compilation

### Debug Mode Assets
```python
# In development, assets are not minified
# Enable debug mode: ?debug=assets

# Or in URL: /web?debug=1
# Or in URL: /web?debug=assets (assets only)
```

### Force Asset Regeneration
```python
# Clear assets cache
def clear_assets_cache(self):
    """Clear compiled assets."""
    self.env['ir.qweb'].clear_caches()

    # Delete asset bundles
    attachments = self.env['ir.attachment'].search([
        ('name', 'ilike', 'web.assets'),
    ])
    attachments.unlink()
```

---

## External Libraries

### Include External Library
```python
{
    'assets': {
        'web.assets_backend': [
            # Include from CDN (not recommended)
            # Better: download and put in static/lib/

            # Local library
            'my_module/static/lib/chart.js/chart.min.js',
            'my_module/static/lib/chart.js/chart.min.css',

            # Then your code that uses it
            'my_module/static/src/js/my_chart.js',
        ],
    },
}
```

### Library Wrapper
```javascript
/** @odoo-module **/

// Wrap external library for Odoo
import { loadJS, loadCSS } from "@web/core/assets";

export async function loadChartJS() {
    await loadJS("/my_module/static/lib/chart.js/chart.min.js");
    return window.Chart;
}

// Usage in component
import { loadChartJS } from "./chart_loader";

class ChartComponent extends Component {
    async setup() {
        this.Chart = await loadChartJS();
    }
}
```

---

## Lazy Loading

### Load Assets On Demand
```javascript
/** @odoo-module **/

import { loadJS, loadCSS } from "@web/core/assets";

export class LazyComponent extends Component {
    static template = "my_module.LazyComponent";

    async setup() {
        // Load heavy library only when needed
        if (this.props.needsChart) {
            await loadJS("/my_module/static/lib/heavy-lib.js");
            await loadCSS("/my_module/static/lib/heavy-lib.css");
        }
    }
}
```

---

## Version-Specific Patterns

### Odoo 17+ (ES Modules)
```javascript
/** @odoo-module **/

import { Component, useState, onMounted } from "@odoo/owl";
import { useService } from "@web/core/utils/hooks";

export class ModernComponent extends Component {
    static template = "my_module.ModernComponent";
    static props = {
        recordId: { type: Number },
        onSave: { type: Function, optional: true },
    };

    setup() {
        this.orm = useService("orm");
        this.state = useState({ data: null });

        onMounted(() => {
            this.loadData();
        });
    }

    async loadData() {
        this.state.data = await this.orm.read(
            "my.model",
            [this.props.recordId],
            ["name", "value"]
        );
    }
}
```

### Odoo 14-16 (Transition Period)
```javascript
/** @odoo-module **/

// May need to import from different paths
import { registry } from "@web/core/registry";
import { Component } from "@odoo/owl";

// Or legacy compatibility
const { Component } = owl;
```

---

## Testing Assets

### QUnit Tests
```javascript
/** @odoo-module **/

import { getFixture, mount } from "@web/../tests/helpers/utils";
import { MyComponent } from "@my_module/js/my_component";

QUnit.module("MyComponent", (hooks) => {
    let target;

    hooks.beforeEach(() => {
        target = getFixture();
    });

    QUnit.test("renders correctly", async (assert) => {
        await mount(MyComponent, target, {
            props: { title: "Test" },
        });

        assert.containsOnce(target, ".my-component");
        assert.strictEqual(
            target.querySelector(".my-component-header h3").textContent,
            "Test"
        );
    });
});
```

---

## Best Practices

1. **Use namespaced classes** - Prefix with `.o_my_module`
2. **Follow directory structure** - Consistent file organization
3. **Use SCSS variables** - Leverage Bootstrap/Odoo variables
4. **Minimize bundle size** - Only include what's needed
5. **Lazy load heavy libs** - Don't slow initial load
6. **Support dark mode** - Use CSS variables
7. **Test with debug=assets** - Catch compilation errors
8. **Version compatibility** - Check asset syntax for version
9. **Document dependencies** - Note external libraries
10. **Clean up on uninstall** - Remove generated assets

---


## Source: logging-debugging-patterns.md

# Logging and Debugging Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  LOGGING & DEBUGGING PATTERNS                                                ║
║  Proper logging, error tracking, and debugging techniques                    ║
║  Use for troubleshooting, monitoring, and audit trails                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Logging Setup

### Module Logger
```python
import logging

_logger = logging.getLogger(__name__)


class MyModel(models.Model):
    _name = 'my.model'

    def process_record(self):
        """Process with proper logging."""
        _logger.info("Processing record %s", self.id)

        try:
            result = self._do_work()
            _logger.debug("Work completed with result: %s", result)
            return result

        except ValueError as e:
            _logger.warning("Invalid value for record %s: %s", self.id, e)
            raise

        except Exception as e:
            _logger.error("Failed to process record %s: %s", self.id, e)
            _logger.exception("Full traceback:")
            raise
```

### Log Levels
| Level | Use Case |
|-------|----------|
| `DEBUG` | Detailed diagnostic info (disabled in production) |
| `INFO` | Normal operation events |
| `WARNING` | Something unexpected but not breaking |
| `ERROR` | Error that affects the operation |
| `CRITICAL` | System-level failure |

### Logging Best Practices
```python
# Good - Use lazy formatting
_logger.info("Processing order %s for customer %s", order.id, customer.name)

# Bad - Eager string formatting (computed even if not logged)
_logger.info(f"Processing order {order.id} for customer {customer.name}")

# Good - Log exceptions with traceback
try:
    risky_operation()
except Exception as e:
    _logger.exception("Operation failed: %s", e)

# Bad - Lose traceback information
try:
    risky_operation()
except Exception as e:
    _logger.error("Operation failed: %s", e)

# Good - Structured context
_logger.info(
    "Order %s: state changed from %s to %s",
    self.name, old_state, new_state
)

# Good - Performance-sensitive debug
if _logger.isEnabledFor(logging.DEBUG):
    _logger.debug("Full data: %s", expensive_data_format())
```

---

## Audit Logging

### Audit Trail Model
```python
class AuditLog(models.Model):
    _name = 'audit.log'
    _description = 'Audit Log'
    _order = 'create_date desc'

    model_name = fields.Char(string='Model', required=True, index=True)
    record_id = fields.Integer(string='Record ID', index=True)
    action = fields.Selection([
        ('create', 'Create'),
        ('write', 'Update'),
        ('unlink', 'Delete'),
        ('action', 'Action'),
    ], string='Action', required=True)
    user_id = fields.Many2one(
        'res.users', string='User',
        default=lambda self: self.env.user,
    )
    timestamp = fields.Datetime(
        string='Timestamp',
        default=fields.Datetime.now,
    )
    old_values = fields.Text(string='Old Values')
    new_values = fields.Text(string='New Values')
    ip_address = fields.Char(string='IP Address')
    description = fields.Text(string='Description')


class AuditMixin(models.AbstractModel):
    _name = 'audit.mixin'
    _description = 'Audit Mixin'

    def write(self, vals):
        """Log changes with audit trail."""
        for record in self:
            old_values = record._get_audit_values(vals.keys())
            result = super(AuditMixin, record).write(vals)
            new_values = record._get_audit_values(vals.keys())
            record._create_audit_log('write', old_values, new_values)
        return result

    @api.model_create_multi
    def create(self, vals_list):
        """Log creation with audit trail."""
        records = super().create(vals_list)
        for record in records:
            record._create_audit_log('create', {}, record._get_audit_values())
        return records

    def unlink(self):
        """Log deletion with audit trail."""
        for record in self:
            record._create_audit_log('unlink', record._get_audit_values(), {})
        return super().unlink()

    def _get_audit_values(self, field_names=None):
        """Get values for audit logging."""
        self.ensure_one()
        if field_names is None:
            field_names = self._get_audit_fields()
        return {
            field: getattr(self, field)
            for field in field_names
            if hasattr(self, field)
        }

    def _get_audit_fields(self):
        """Override to specify fields to audit."""
        return ['name', 'state']

    def _create_audit_log(self, action, old_values, new_values):
        """Create audit log entry."""
        self.env['audit.log'].sudo().create({
            'model_name': self._name,
            'record_id': self.id,
            'action': action,
            'old_values': json.dumps(old_values, default=str),
            'new_values': json.dumps(new_values, default=str),
            'ip_address': self._get_client_ip(),
        })

    def _get_client_ip(self):
        """Get client IP from request."""
        try:
            from odoo.http import request
            if request:
                return request.httprequest.remote_addr
        except Exception:
            pass
        return None
```

---

## Performance Logging

### Query Profiling
```python
import time


class PerformanceLogger:
    """Context manager for performance logging."""

    def __init__(self, operation_name, logger=None):
        self.operation_name = operation_name
        self.logger = logger or _logger
        self.start_time = None

    def __enter__(self):
        self.start_time = time.time()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        duration = time.time() - self.start_time
        if duration > 1.0:  # Log slow operations
            self.logger.warning(
                "Slow operation: %s took %.2fs",
                self.operation_name, duration
            )
        else:
            self.logger.debug(
                "Operation %s completed in %.3fs",
                self.operation_name, duration
            )


# Usage
def compute_report(self):
    with PerformanceLogger("compute_report"):
        # Heavy computation
        pass
```

### SQL Query Logging
```python
def _log_query_count(self, operation_name):
    """Log number of SQL queries in operation."""
    cr = self.env.cr

    initial_count = cr.sql_log_count if hasattr(cr, 'sql_log_count') else 0

    yield

    final_count = cr.sql_log_count if hasattr(cr, 'sql_log_count') else 0
    query_count = final_count - initial_count

    if query_count > 100:
        _logger.warning(
            "High query count in %s: %d queries",
            operation_name, query_count
        )
```

### Memory Profiling (Development)
```python
import tracemalloc


def profile_memory(func):
    """Decorator for memory profiling."""
    @wraps(func)
    def wrapper(*args, **kwargs):
        tracemalloc.start()
        try:
            result = func(*args, **kwargs)
            current, peak = tracemalloc.get_traced_memory()
            _logger.info(
                "%s memory: current=%.1fMB, peak=%.1fMB",
                func.__name__,
                current / 1024 / 1024,
                peak / 1024 / 1024
            )
            return result
        finally:
            tracemalloc.stop()
    return wrapper
```

---

## Error Tracking

### Error Log Model
```python
class ErrorLog(models.Model):
    _name = 'error.log'
    _description = 'Error Log'
    _order = 'create_date desc'

    name = fields.Char(string='Error', required=True)
    model_name = fields.Char(string='Model')
    record_id = fields.Integer(string='Record ID')
    method_name = fields.Char(string='Method')
    error_type = fields.Char(string='Error Type')
    error_message = fields.Text(string='Error Message')
    traceback = fields.Text(string='Traceback')
    user_id = fields.Many2one('res.users', string='User')
    resolved = fields.Boolean(string='Resolved', default=False)
    occurrence_count = fields.Integer(string='Occurrences', default=1)

    @api.model
    def log_error(self, error, model_name=None, record_id=None, method_name=None):
        """Log error with deduplication."""
        import traceback as tb

        error_type = type(error).__name__
        error_message = str(error)
        traceback_str = tb.format_exc()

        # Check for existing similar error
        existing = self.search([
            ('error_type', '=', error_type),
            ('error_message', '=', error_message),
            ('resolved', '=', False),
        ], limit=1)

        if existing:
            existing.occurrence_count += 1
            return existing

        return self.create({
            'name': f"{error_type}: {error_message[:100]}",
            'model_name': model_name,
            'record_id': record_id,
            'method_name': method_name,
            'error_type': error_type,
            'error_message': error_message,
            'traceback': traceback_str,
            'user_id': self.env.uid,
        })
```

### Error Handling Decorator
```python
def log_exceptions(model_name=None, method_name=None):
    """Decorator to log exceptions to error.log."""
    def decorator(func):
        @wraps(func)
        def wrapper(self, *args, **kwargs):
            try:
                return func(self, *args, **kwargs)
            except Exception as e:
                _logger.exception("Error in %s: %s", func.__name__, e)
                self.env['error.log'].sudo().log_error(
                    error=e,
                    model_name=model_name or self._name,
                    record_id=self.id if hasattr(self, 'id') else None,
                    method_name=method_name or func.__name__,
                )
                raise
        return wrapper
    return decorator


# Usage
class MyModel(models.Model):
    _name = 'my.model'

    @log_exceptions()
    def risky_operation(self):
        # Code that might fail
        pass
```

---

## Debug Helpers

### Debug Mode Check
```python
from odoo.tools import config


def is_debug_mode():
    """Check if server is in debug mode."""
    return config.get('dev_mode') or config.get('debug')


class MyModel(models.Model):
    _name = 'my.model'

    def process(self):
        if is_debug_mode():
            _logger.setLevel(logging.DEBUG)
            _logger.debug("Debug mode enabled - verbose logging")

        # Processing logic
```

### Temporary Debug Output
```python
def debug_record(self):
    """Debug helper to print record details."""
    self.ensure_one()

    info = {
        'id': self.id,
        'name': self.name,
        'state': self.state,
        'create_date': str(self.create_date),
        'write_date': str(self.write_date),
    }

    _logger.info("=== DEBUG RECORD ===")
    for key, value in info.items():
        _logger.info("  %s: %s", key, value)
    _logger.info("====================")
```

### SQL Debug
```python
def debug_sql(self, query):
    """Execute and log SQL query for debugging."""
    _logger.info("Executing SQL: %s", query)
    self.env.cr.execute(query)
    result = self.env.cr.fetchall()
    _logger.info("Result: %s rows", len(result))
    return result
```

---

## Request Logging

### HTTP Request Logger
```python
from odoo import http
from odoo.http import request
import time


class RequestLogger(http.Controller):

    @http.route('/api/endpoint', type='json', auth='user')
    def my_endpoint(self, **kwargs):
        start_time = time.time()
        request_id = self._generate_request_id()

        _logger.info(
            "[%s] Request: user=%s, params=%s",
            request_id, request.env.user.login, kwargs
        )

        try:
            result = self._process_request(**kwargs)

            duration = time.time() - start_time
            _logger.info(
                "[%s] Response: success, duration=%.3fs",
                request_id, duration
            )

            return result

        except Exception as e:
            duration = time.time() - start_time
            _logger.error(
                "[%s] Response: error=%s, duration=%.3fs",
                request_id, str(e), duration
            )
            raise

    def _generate_request_id(self):
        import uuid
        return str(uuid.uuid4())[:8]
```

---

## Configuration

### Odoo Logging Configuration
```ini
# odoo.conf
[options]
log_level = info
log_handler = :INFO,odoo.addons.my_module:DEBUG
log_db = True
log_db_level = warning
logfile = /var/log/odoo/odoo.log
```

### Per-Module Log Level
```python
# At module init, set specific log level
import logging

# Set module-specific log level
logging.getLogger('odoo.addons.my_module').setLevel(logging.DEBUG)

# Or conditionally
if config.get('my_module_debug'):
    logging.getLogger('odoo.addons.my_module').setLevel(logging.DEBUG)
```

---

## Best Practices

1. **Use module logger** - `_logger = logging.getLogger(__name__)`
2. **Lazy formatting** - `_logger.info("x=%s", x)` not `f"x={x}"`
3. **Include context** - Log record IDs, user, operation name
4. **Use appropriate levels** - DEBUG for dev, INFO for operations
5. **Log exceptions** - Use `_logger.exception()` to include traceback
6. **Don't log sensitive data** - Mask passwords, tokens, PII
7. **Performance aware** - Check log level before expensive formatting
8. **Structured logging** - Consistent format for parsing
9. **Audit critical operations** - Financial, security, compliance
10. **Clean up debug logs** - Remove temporary logging before commit

---


## Source: error-handling-patterns.md

# Error Handling Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ERROR HANDLING PATTERNS                                                     ║
║  Exceptions, validation, and error recovery                                  ║
║  Use for robust error handling, user feedback, and data integrity            ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Odoo Exception Types

### Standard Exceptions
```python
from odoo.exceptions import (
    UserError,        # User-facing errors (shown in dialog)
    ValidationError,  # Constraint violations
    AccessError,      # Permission denied
    MissingError,     # Record doesn't exist
    AccessDenied,     # Login/authentication failure
    RedirectWarning,  # Error with action button
)
```

### When to Use Each
| Exception | Use Case |
|-----------|----------|
| `UserError` | Business logic errors, invalid operations |
| `ValidationError` | Data validation failures in constraints |
| `AccessError` | Permission/security violations |
| `MissingError` | Record not found (browse deleted ID) |
| `RedirectWarning` | Error with corrective action link |

---

## Raising Exceptions

### UserError (Most Common)
```python
from odoo.exceptions import UserError


class MyModel(models.Model):
    _name = 'my.model'

    def action_confirm(self):
        """Confirm record with validation."""
        self.ensure_one()

        if not self.line_ids:
            raise UserError("Cannot confirm without lines.")

        if self.amount_total <= 0:
            raise UserError(
                f"Amount must be positive. Current: {self.amount_total}"
            )

        if self.state != 'draft':
            raise UserError(
                f"Only draft records can be confirmed. "
                f"Current state: {self.state}"
            )

        self.write({'state': 'confirmed'})
```

### ValidationError (Constraints)
```python
from odoo.exceptions import ValidationError


class MyModel(models.Model):
    _name = 'my.model'

    @api.constrains('date_start', 'date_end')
    def _check_dates(self):
        for record in self:
            if record.date_start and record.date_end:
                if record.date_start > record.date_end:
                    raise ValidationError(
                        "Start date must be before end date."
                    )

    @api.constrains('email')
    def _check_email(self):
        import re
        email_pattern = r'^[\w\.-]+@[\w\.-]+\.\w+$'
        for record in self:
            if record.email and not re.match(email_pattern, record.email):
                raise ValidationError(
                    f"Invalid email format: {record.email}"
                )

    @api.constrains('quantity')
    def _check_quantity(self):
        for record in self:
            if record.quantity < 0:
                raise ValidationError("Quantity cannot be negative.")
```

### RedirectWarning (With Action)
```python
from odoo.exceptions import RedirectWarning


class MyModel(models.Model):
    _name = 'my.model'

    def action_process(self):
        if not self.env.company.x_config_complete:
            action = self.env.ref('my_module.action_config_wizard')
            raise RedirectWarning(
                "Configuration is incomplete. Please complete setup first.",
                action.id,
                "Go to Configuration",
            )
```

### AccessError (Security)
```python
from odoo.exceptions import AccessError


class MyModel(models.Model):
    _name = 'my.model'

    def action_approve(self):
        if not self.env.user.has_group('my_module.group_approver'):
            raise AccessError(
                "You do not have permission to approve records."
            )
        self.write({'state': 'approved'})
```

---

## Try-Except Patterns

### Basic Error Handling
```python
def process_record(self):
    try:
        self._do_processing()
    except UserError:
        # Re-raise user errors (show to user)
        raise
    except Exception as e:
        _logger.error("Processing failed for %s: %s", self.id, e)
        raise UserError(f"Processing failed: {str(e)}")
```

### Handling Specific Exceptions
```python
def sync_external(self):
    import requests

    try:
        response = requests.get(self.api_url, timeout=30)
        response.raise_for_status()
        return response.json()

    except requests.Timeout:
        raise UserError(
            "External service timed out. Please try again later."
        )
    except requests.ConnectionError:
        raise UserError(
            "Cannot connect to external service. Check your connection."
        )
    except requests.HTTPError as e:
        if e.response.status_code == 401:
            raise UserError("Authentication failed. Check API credentials.")
        elif e.response.status_code == 404:
            raise UserError("Resource not found on external service.")
        else:
            raise UserError(f"External service error: {e.response.status_code}")
    except Exception as e:
        _logger.exception("Unexpected error in sync: %s", e)
        raise UserError(f"Sync failed: {str(e)}")
```

### Transaction Safety
```python
def process_batch(self):
    """Process batch with transaction safety."""
    for record in self:
        try:
            # Use savepoint for each record
            with self.env.cr.savepoint():
                record._process_single()

        except Exception as e:
            # Savepoint rolled back, continue with next
            _logger.error("Failed to process %s: %s", record.id, e)
            record.message_post(body=f"Processing failed: {e}")
            continue
```

### Graceful Degradation
```python
def get_external_data(self):
    """Get data with fallback."""
    try:
        # Try primary source
        return self._fetch_from_api()
    except Exception as e:
        _logger.warning("API fetch failed: %s, using cache", e)
        try:
            # Fallback to cache
            return self._get_from_cache()
        except Exception:
            _logger.error("Cache also failed")
            return None
```

---

## Validation Patterns

### Pre-Action Validation
```python
class MyModel(models.Model):
    _name = 'my.model'

    def action_confirm(self):
        """Confirm with pre-validation."""
        self._validate_confirm()
        self.write({'state': 'confirmed'})

    def _validate_confirm(self):
        """Validate before confirmation."""
        errors = []

        for record in self:
            if not record.partner_id:
                errors.append(f"[{record.name}] Partner is required.")
            if not record.line_ids:
                errors.append(f"[{record.name}] At least one line required.")
            if record.amount_total <= 0:
                errors.append(f"[{record.name}] Amount must be positive.")

        if errors:
            raise UserError("\n".join(errors))
```

### Field Validation Decorator
```python
def validate_required(field_name, message=None):
    """Decorator to validate required field."""
    def decorator(func):
        @wraps(func)
        def wrapper(self, *args, **kwargs):
            for record in self:
                if not getattr(record, field_name):
                    raise UserError(
                        message or f"{field_name} is required."
                    )
            return func(self, *args, **kwargs)
        return wrapper
    return decorator


class MyModel(models.Model):
    _name = 'my.model'

    @validate_required('partner_id', 'Please select a partner first.')
    def action_send(self):
        # Validation passed
        self._send_to_partner()
```

### SQL Constraint Handling
```python
class MyModel(models.Model):
    _name = 'my.model'

    _sql_constraints = [
        ('code_unique', 'UNIQUE(code, company_id)',
         'Code must be unique per company.'),
        ('amount_positive', 'CHECK(amount >= 0)',
         'Amount must be positive.'),
    ]

    @api.model_create_multi
    def create(self, vals_list):
        try:
            return super().create(vals_list)
        except IntegrityError as e:
            if 'code_unique' in str(e):
                raise UserError("A record with this code already exists.")
            raise
```

---

## Error Recovery

### Retry Pattern
```python
import time


def retry_on_error(max_retries=3, delay=1, exceptions=(Exception,)):
    """Decorator for retry on error."""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            last_error = None
            for attempt in range(max_retries):
                try:
                    return func(*args, **kwargs)
                except exceptions as e:
                    last_error = e
                    if attempt < max_retries - 1:
                        _logger.warning(
                            "Attempt %d failed: %s. Retrying...",
                            attempt + 1, e
                        )
                        time.sleep(delay * (attempt + 1))
            raise last_error
        return wrapper
    return decorator


class MyModel(models.Model):
    _name = 'my.model'

    @retry_on_error(max_retries=3, delay=2)
    def call_external_api(self):
        # May fail temporarily
        pass
```

### Rollback and Notify
```python
def process_with_rollback(self):
    """Process with proper rollback handling."""
    try:
        # Start work
        for record in self:
            record._process()

        # Commit if all successful
        self.env.cr.commit()

    except Exception as e:
        # Rollback transaction
        self.env.cr.rollback()

        # Notify about failure
        self.env['bus.bus']._sendone(
            self.env.user.partner_id,
            'simple_notification',
            {
                'title': 'Processing Failed',
                'message': str(e),
                'type': 'danger',
            }
        )
        raise
```

---

## User Feedback

### Progress Messages
```python
def action_process_batch(self):
    """Process with user feedback."""
    total = len(self)
    processed = 0
    errors = []

    for record in self:
        try:
            record._process()
            processed += 1
        except Exception as e:
            errors.append(f"{record.name}: {str(e)}")

    # Return notification
    if errors:
        return {
            'type': 'ir.actions.client',
            'tag': 'display_notification',
            'params': {
                'title': 'Processing Complete',
                'message': f'Processed {processed}/{total}. '
                          f'{len(errors)} errors occurred.',
                'type': 'warning',
                'sticky': True,
            }
        }
    else:
        return {
            'type': 'ir.actions.client',
            'tag': 'display_notification',
            'params': {
                'title': 'Success',
                'message': f'Successfully processed {total} records.',
                'type': 'success',
            }
        }
```

### Error Details in Chatter
```python
def process_with_logging(self):
    """Process with error logging to chatter."""
    try:
        result = self._do_work()
        self.message_post(
            body=f"Processing completed successfully: {result}",
            message_type='notification',
        )
    except Exception as e:
        self.message_post(
            body=f"<b>Processing Failed</b><br/>{str(e)}",
            message_type='notification',
            subtype_xmlid='mail.mt_note',
        )
        raise UserError(f"Processing failed: {str(e)}")
```

---

## Best Practices

1. **Use appropriate exception types** - UserError for business logic, ValidationError for constraints
2. **Provide clear messages** - Include context and what to do
3. **Log before raising** - Log technical details, show user-friendly message
4. **Don't catch too broadly** - Be specific about what you catch
5. **Always re-raise UserError** - Let it show to user
6. **Use savepoints for batches** - Isolate failures
7. **Validate early** - Check before doing work
8. **Return feedback** - Use notifications for batch operations
9. **Document expected errors** - In docstrings
10. **Test error paths** - Write tests for error scenarios

---

## Anti-Patterns

```python
# Bad - Too broad
try:
    do_something()
except:
    pass

# Bad - Silent failure
try:
    do_something()
except Exception:
    return False

# Bad - Losing error context
try:
    do_something()
except Exception:
    raise UserError("Something went wrong")  # Lost original error

# Good - Preserve and log
try:
    do_something()
except Exception as e:
    _logger.exception("Failed in do_something")
    raise UserError(f"Operation failed: {e}")
```

---


## Source: config-settings-patterns.md

# Configuration Settings Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  CONFIGURATION SETTINGS PATTERNS                                             ║
║  Module settings, res.config.settings, and system parameters                 ║
║  Use for user-configurable options and feature toggles                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Basic Settings Model

### Inherit res.config.settings
```python
from odoo import api, fields, models


class ResConfigSettings(models.TransientModel):
    _inherit = 'res.config.settings'

    # Simple config parameter
    my_module_api_key = fields.Char(
        string='API Key',
        config_parameter='my_module.api_key',
    )

    # Boolean setting
    my_module_enable_feature = fields.Boolean(
        string='Enable Feature X',
        config_parameter='my_module.enable_feature',
    )

    # Selection setting
    my_module_mode = fields.Selection([
        ('basic', 'Basic Mode'),
        ('advanced', 'Advanced Mode'),
    ], string='Operating Mode',
       config_parameter='my_module.mode',
       default='basic')

    # Integer setting
    my_module_max_items = fields.Integer(
        string='Maximum Items',
        config_parameter='my_module.max_items',
        default=100,
    )
```

---

## Company-Specific Settings

### Company Fields (Not Global)
```python
class ResConfigSettings(models.TransientModel):
    _inherit = 'res.config.settings'

    # Company-specific setting (stored on res.company)
    my_default_warehouse_id = fields.Many2one(
        'stock.warehouse',
        string='Default Warehouse',
        related='company_id.my_default_warehouse_id',
        readonly=False,
    )

    my_default_journal_id = fields.Many2one(
        'account.journal',
        string='Default Journal',
        related='company_id.my_default_journal_id',
        readonly=False,
    )


class ResCompany(models.Model):
    _inherit = 'res.company'

    my_default_warehouse_id = fields.Many2one(
        'stock.warehouse',
        string='Default Warehouse',
    )

    my_default_journal_id = fields.Many2one(
        'account.journal',
        string='Default Journal',
    )
```

---

## Feature Group Toggle

### Settings that Enable/Disable Features
```python
class ResConfigSettings(models.TransientModel):
    _inherit = 'res.config.settings'

    # Feature toggle linked to security group
    group_use_advanced_routing = fields.Boolean(
        string='Use Advanced Routing',
        implied_group='my_module.group_advanced_routing',
    )

    group_multi_warehouse = fields.Boolean(
        string='Multi-Warehouse',
        implied_group='stock.group_stock_multi_warehouses',
    )

    # Module toggle (installs module when enabled)
    module_my_optional_feature = fields.Boolean(
        string='Enable Optional Feature',
        help='Installs my_optional_feature module',
    )
```

### Security Group for Feature
```xml
<!-- Security group enabled by setting -->
<record id="group_advanced_routing" model="res.groups">
    <field name="name">Advanced Routing</field>
    <field name="category_id" ref="base.module_category_hidden"/>
    <field name="comment">Users with access to advanced routing features</field>
</record>
```

---

## Settings View

### Settings View Definition
```xml
<record id="res_config_settings_view_form" model="ir.ui.view">
    <field name="name">res.config.settings.view.form.inherit.my_module</field>
    <field name="model">res.config.settings</field>
    <field name="inherit_id" ref="base.res_config_settings_view_form"/>
    <field name="arch" type="xml">
        <xpath expr="//form" position="inside">
            <app data-string="My Module" string="My Module"
                 data-key="my_module" groups="base.group_system">
                <block title="General Settings">
                    <setting string="API Configuration">
                        <div class="text-muted">
                            Configure external API settings
                        </div>
                        <div class="content-group">
                            <div class="row mt-2">
                                <label for="my_module_api_key" class="col-lg-3"/>
                                <field name="my_module_api_key" class="col-lg-9"/>
                            </div>
                        </div>
                    </setting>

                    <setting string="Operating Mode">
                        <div class="text-muted">
                            Select the operating mode for the module
                        </div>
                        <div class="content-group">
                            <div class="row mt-2">
                                <field name="my_module_mode" class="col-lg-6"/>
                            </div>
                        </div>
                    </setting>
                </block>

                <block title="Features">
                    <setting help="Enable advanced features for power users">
                        <field name="group_use_advanced_routing"/>
                        <div invisible="not group_use_advanced_routing"
                             class="content-group mt-2">
                            <div class="row">
                                <label for="my_module_max_items" class="col-lg-4"/>
                                <field name="my_module_max_items" class="col-lg-8"/>
                            </div>
                        </div>
                    </setting>

                    <setting help="Install additional module for extended functionality">
                        <field name="module_my_optional_feature"/>
                    </setting>
                </block>

                <block title="Company Settings"
                       invisible="not company_id">
                    <setting string="Default Warehouse">
                        <div class="content-group">
                            <div class="row mt-2">
                                <label for="my_default_warehouse_id" class="col-lg-4"/>
                                <field name="my_default_warehouse_id" class="col-lg-8"/>
                            </div>
                        </div>
                    </setting>
                </block>
            </app>
        </xpath>
    </field>
</record>
```

---

## Reading Settings in Code

### Using Config Parameters
```python
class MyModel(models.Model):
    _name = 'my.model'

    def _get_api_key(self):
        """Read config parameter."""
        return self.env['ir.config_parameter'].sudo().get_param(
            'my_module.api_key',
            default='',
        )

    def _is_feature_enabled(self):
        """Check if feature is enabled."""
        param = self.env['ir.config_parameter'].sudo().get_param(
            'my_module.enable_feature',
            default='False',
        )
        # Parameters are stored as strings
        return param == 'True'

    def _get_max_items(self):
        """Read integer config."""
        param = self.env['ir.config_parameter'].sudo().get_param(
            'my_module.max_items',
            default='100',
        )
        return int(param)

    def do_something(self):
        """Example using settings."""
        if not self._is_feature_enabled():
            raise UserError("Feature is disabled in settings.")

        api_key = self._get_api_key()
        max_items = self._get_max_items()

        # Use settings...
```

### Reading Company Settings
```python
def _get_default_warehouse(self):
    """Get company default warehouse."""
    company = self.env.company
    return company.my_default_warehouse_id
```

### Checking Group Membership
```python
def has_advanced_access(self):
    """Check if user has advanced access."""
    return self.env.user.has_group('my_module.group_advanced_routing')
```

---

## Settings with Computed Fields

### Computed Settings Field
```python
class ResConfigSettings(models.TransientModel):
    _inherit = 'res.config.settings'

    # Stored in ir.config_parameter
    my_module_enabled = fields.Boolean(
        config_parameter='my_module.enabled',
    )

    # Computed based on another setting
    my_module_status = fields.Char(
        compute='_compute_status',
    )

    # Count for information display
    record_count = fields.Integer(
        compute='_compute_record_count',
    )

    @api.depends('my_module_enabled')
    def _compute_status(self):
        for record in self:
            record.my_module_status = 'Active' if record.my_module_enabled else 'Inactive'

    def _compute_record_count(self):
        for record in self:
            record.record_count = self.env['my.model'].search_count([])
```

---

## Settings with Actions

### Execute Action from Settings
```python
class ResConfigSettings(models.TransientModel):
    _inherit = 'res.config.settings'

    def action_sync_data(self):
        """Button action in settings."""
        # Perform sync operation
        self.env['my.model'].sync_all()
        return {
            'type': 'ir.actions.client',
            'tag': 'display_notification',
            'params': {
                'title': 'Sync Complete',
                'message': 'Data synchronization completed.',
                'type': 'success',
            }
        }

    def action_open_records(self):
        """Open related records."""
        return {
            'type': 'ir.actions.act_window',
            'name': 'My Records',
            'res_model': 'my.model',
            'view_mode': 'tree,form',
        }
```

### Action Button in View
```xml
<setting string="Data Synchronization">
    <div class="text-muted">
        Sync data with external system
    </div>
    <div class="content-group mt-2">
        <button name="action_sync_data" type="object"
                string="Sync Now" class="btn-primary"/>
    </div>
</setting>

<setting string="Quick Access">
    <div class="content-group">
        <button name="action_open_records" type="object"
                string="View Records" class="btn-link"
                icon="fa-external-link"/>
    </div>
</setting>
```

---

## Default Values Pattern

### Set Defaults Based on Settings
```python
class MyModel(models.Model):
    _name = 'my.model'

    warehouse_id = fields.Many2one(
        'stock.warehouse',
        default=lambda self: self._get_default_warehouse(),
    )

    mode = fields.Selection([
        ('basic', 'Basic'),
        ('advanced', 'Advanced'),
    ], default=lambda self: self._get_default_mode())

    def _get_default_warehouse(self):
        """Get default from company settings."""
        return self.env.company.my_default_warehouse_id

    def _get_default_mode(self):
        """Get default from system settings."""
        return self.env['ir.config_parameter'].sudo().get_param(
            'my_module.mode',
            default='basic',
        )
```

---

## XML Data for Settings

### Default Config Parameters
```xml
<!-- Set default config parameters -->
<record id="config_my_module_enabled" model="ir.config_parameter">
    <field name="key">my_module.enabled</field>
    <field name="value">True</field>
</record>

<record id="config_my_module_mode" model="ir.config_parameter">
    <field name="key">my_module.mode</field>
    <field name="value">basic</field>
</record>

<record id="config_my_module_max_items" model="ir.config_parameter">
    <field name="key">my_module.max_items</field>
    <field name="value">100</field>
</record>
```

---

## Settings Validation

### Validate on Save
```python
class ResConfigSettings(models.TransientModel):
    _inherit = 'res.config.settings'

    my_module_api_url = fields.Char(
        config_parameter='my_module.api_url',
    )

    @api.constrains('my_module_api_url')
    def _check_api_url(self):
        """Validate API URL format."""
        for record in self:
            if record.my_module_api_url:
                if not record.my_module_api_url.startswith('https://'):
                    raise ValidationError("API URL must use HTTPS.")

    def set_values(self):
        """Custom validation before saving."""
        super().set_values()

        # Additional validation
        if self.my_module_enabled and not self.my_module_api_key:
            raise UserError("API Key is required when feature is enabled.")
```

---

## Conditional Settings Display

### Show Settings Based on Other Settings
```xml
<setting string="Enable API Integration">
    <field name="my_module_api_enabled"/>
</setting>

<!-- Only show when API is enabled -->
<setting string="API Configuration"
         invisible="not my_module_api_enabled">
    <div class="content-group">
        <div class="row mt-2">
            <label for="my_module_api_url" class="col-lg-3"/>
            <field name="my_module_api_url" class="col-lg-9"
                   required="my_module_api_enabled"/>
        </div>
        <div class="row mt-2">
            <label for="my_module_api_key" class="col-lg-3"/>
            <field name="my_module_api_key" class="col-lg-9"
                   password="True"
                   required="my_module_api_enabled"/>
        </div>
    </div>
</setting>
```

---

## Menu Access

### Settings Menu Item
```xml
<!-- Menu to open settings -->
<record id="my_module_config_settings_action" model="ir.actions.act_window">
    <field name="name">Settings</field>
    <field name="res_model">res.config.settings</field>
    <field name="view_mode">form</field>
    <field name="target">inline</field>
    <field name="context">{'module': 'my_module'}</field>
</record>

<menuitem id="menu_my_module_configuration"
          name="Configuration"
          parent="menu_my_module_root"
          sequence="100"/>

<menuitem id="menu_my_module_settings"
          name="Settings"
          parent="menu_my_module_configuration"
          action="my_module_config_settings_action"
          groups="base.group_system"/>
```

---

## Best Practices

1. **Use config_parameter** - For global settings stored in ir.config_parameter
2. **Use related company fields** - For company-specific settings
3. **Use implied_group** - For feature toggles that grant permissions
4. **Use module_** prefix - For settings that install modules
5. **Validate settings** - Check values before saving
6. **Provide defaults** - Set sensible default values
7. **Group logically** - Use blocks to organize related settings
8. **Show/hide conditionally** - Use invisible attribute
9. **Document settings** - Use help text and descriptions
10. **Secure sensitive data** - Use password="True" for secrets

---


## Source: attachment-binary-patterns.md

# Attachment and Binary Field Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ATTACHMENT & BINARY FIELD PATTERNS                                          ║
║  File uploads, images, documents, and attachments                            ║
║  Use for handling files, images, and documents in Odoo                       ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Binary Field Types

| Field Type | Use Case |
|------------|----------|
| `Binary` | Generic file storage |
| `Image` | Image with auto-resize |

---

## Basic Binary Fields

### File Upload Field
```python
from odoo import fields, models


class MyModel(models.Model):
    _name = 'my.model'

    # Basic binary field
    document = fields.Binary(string='Document')
    document_name = fields.Char(string='Document Name')

    # Binary with specific attachment flag
    attachment = fields.Binary(
        string='Attachment',
        attachment=True,  # Store as ir.attachment
    )
    attachment_name = fields.Char(string='Attachment Name')
```

### Image Field
```python
class MyModel(models.Model):
    _name = 'my.model'

    # Image field (auto-resizes)
    image = fields.Image(string='Image')

    # Image with max dimensions
    image_1920 = fields.Image(
        string='Image',
        max_width=1920,
        max_height=1920,
    )

    # Multiple image sizes (common pattern)
    image_1920 = fields.Image(max_width=1920, max_height=1920)
    image_1024 = fields.Image(
        related='image_1920',
        max_width=1024,
        max_height=1024,
        store=True,
    )
    image_512 = fields.Image(
        related='image_1920',
        max_width=512,
        max_height=512,
        store=True,
    )
    image_256 = fields.Image(
        related='image_1920',
        max_width=256,
        max_height=256,
        store=True,
    )
    image_128 = fields.Image(
        related='image_1920',
        max_width=128,
        max_height=128,
        store=True,
    )
```

---

## Views for Binary Fields

### Form View - File Upload
```xml
<form>
    <sheet>
        <group>
            <!-- File upload with filename -->
            <field name="document" filename="document_name"/>
            <field name="document_name" invisible="1"/>
        </group>
    </sheet>
</form>
```

### Form View - Image
```xml
<form>
    <sheet>
        <!-- Image at top of form (like res.partner) -->
        <field name="image_1920" widget="image" class="oe_avatar"
               options="{'preview_image': 'image_128'}"/>

        <!-- Or in a group -->
        <group>
            <field name="image" widget="image"/>
        </group>
    </sheet>
</form>
```

### Tree View - Image
```xml
<tree>
    <field name="image_128" widget="image" options="{'size': [32, 32]}"/>
    <field name="name"/>
</tree>
```

### Kanban View - Image
```xml
<kanban>
    <field name="image_128"/>
    <templates>
        <t t-name="kanban-box">
            <div class="oe_kanban_card">
                <div class="o_kanban_image">
                    <img t-att-src="kanban_image('my.model', 'image_128', record.id.raw_value)"
                         alt="Image" class="o_image_64_cover"/>
                </div>
                <div class="oe_kanban_details">
                    <field name="name"/>
                </div>
            </div>
        </t>
    </templates>
</kanban>
```

---

## ir.attachment Model

### Working with Attachments
```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['mail.thread']  # For attachment tracking

    attachment_ids = fields.Many2many(
        'ir.attachment',
        string='Attachments',
    )

    # Or One2many for owned attachments
    document_ids = fields.One2many(
        'ir.attachment',
        'res_id',
        domain=[('res_model', '=', 'my.model')],
        string='Documents',
    )
```

### Creating Attachments
```python
def action_create_attachment(self):
    """Create attachment from binary data."""
    import base64

    attachment = self.env['ir.attachment'].create({
        'name': 'my_file.pdf',
        'type': 'binary',
        'datas': base64.b64encode(b'file content'),
        'res_model': self._name,
        'res_id': self.id,
        'mimetype': 'application/pdf',
    })
    return attachment

def action_attach_file(self, file_content, filename):
    """Attach file to record."""
    return self.env['ir.attachment'].create({
        'name': filename,
        'type': 'binary',
        'datas': base64.b64encode(file_content),
        'res_model': self._name,
        'res_id': self.id,
    })
```

### Reading Attachment Content
```python
def get_attachment_content(self, attachment_id):
    """Get attachment file content."""
    import base64

    attachment = self.env['ir.attachment'].browse(attachment_id)
    if attachment.exists():
        return base64.b64decode(attachment.datas)
    return None
```

### Deleting Attachments
```python
def action_cleanup_attachments(self):
    """Remove orphan attachments."""
    attachments = self.env['ir.attachment'].search([
        ('res_model', '=', self._name),
        ('res_id', '=', 0),  # Orphan attachments
    ])
    attachments.unlink()
```

---

## File Upload Controller

### Basic Upload Endpoint
```python
from odoo import http
from odoo.http import request
import base64


class FileUploadController(http.Controller):

    @http.route('/my_module/upload', type='http', auth='user',
                methods=['POST'], csrf=False)
    def upload_file(self, file, record_id, **kwargs):
        """Handle file upload."""
        if not file:
            return request.make_json_response({'error': 'No file'}, status=400)

        # Read file content
        file_content = file.read()
        file_name = file.filename

        # Create attachment
        attachment = request.env['ir.attachment'].sudo().create({
            'name': file_name,
            'type': 'binary',
            'datas': base64.b64encode(file_content),
            'res_model': 'my.model',
            'res_id': int(record_id),
        })

        return request.make_json_response({
            'success': True,
            'attachment_id': attachment.id,
        })
```

### Download Endpoint
```python
@http.route('/my_module/download/<int:attachment_id>', type='http',
            auth='user')
def download_file(self, attachment_id, **kwargs):
    """Download attachment."""
    attachment = request.env['ir.attachment'].sudo().browse(attachment_id)

    if not attachment.exists():
        return request.not_found()

    # Check access
    attachment.check('read')

    return request.make_response(
        base64.b64decode(attachment.datas),
        headers=[
            ('Content-Type', attachment.mimetype or 'application/octet-stream'),
            ('Content-Disposition', f'attachment; filename="{attachment.name}"'),
        ]
    )
```

---

## Image Processing

### Resize Image
```python
import base64
from io import BytesIO
from PIL import Image


def resize_image(self, image_data, max_width=1024, max_height=1024):
    """Resize image to max dimensions."""
    if not image_data:
        return image_data

    # Decode base64
    image_bytes = base64.b64decode(image_data)
    img = Image.open(BytesIO(image_bytes))

    # Calculate new size maintaining aspect ratio
    img.thumbnail((max_width, max_height), Image.LANCZOS)

    # Convert back to base64
    buffer = BytesIO()
    img_format = img.format or 'PNG'
    img.save(buffer, format=img_format)

    return base64.b64encode(buffer.getvalue())
```

### Generate Thumbnail
```python
def generate_thumbnail(self, image_data, size=(128, 128)):
    """Generate thumbnail from image."""
    if not image_data:
        return False

    image_bytes = base64.b64decode(image_data)
    img = Image.open(BytesIO(image_bytes))

    # Create thumbnail
    img.thumbnail(size, Image.LANCZOS)

    buffer = BytesIO()
    img.save(buffer, format='PNG')

    return base64.b64encode(buffer.getvalue())
```

### Image from URL
```python
import requests
import base64


def image_from_url(self, url):
    """Fetch image from URL."""
    try:
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        return base64.b64encode(response.content)
    except Exception:
        return False
```

---

## Document Preview Widget

### PDF Preview
```xml
<form>
    <sheet>
        <group>
            <!-- PDF viewer widget -->
            <field name="pdf_document" widget="pdf_viewer"/>
        </group>
    </sheet>
</form>
```

### Image Preview with Zoom
```xml
<field name="image" widget="image" options="{'zoom': true, 'zoom_delay': 500}"/>
```

---

## Signature Field

### Model Definition
```python
class MyModel(models.Model):
    _name = 'my.model'

    signature = fields.Binary(string='Signature')
```

### View
```xml
<form>
    <sheet>
        <group>
            <field name="signature" widget="signature"/>
        </group>
    </sheet>
</form>
```

---

## Many Attachments Pattern

### Attachment Button in Form
```xml
<form>
    <sheet>
        <div class="oe_button_box" name="button_box">
            <button name="action_view_attachments" type="object"
                    class="oe_stat_button" icon="fa-files-o">
                <div class="o_field_widget o_stat_info">
                    <span class="o_stat_value">
                        <field name="attachment_count" widget="statinfo"/>
                    </span>
                    <span class="o_stat_text">Attachments</span>
                </div>
            </button>
        </div>
    </sheet>
</form>
```

### Attachment Count and Action
```python
class MyModel(models.Model):
    _name = 'my.model'
    _inherit = ['mail.thread']

    attachment_count = fields.Integer(
        compute='_compute_attachment_count',
        string='Attachments',
    )

    def _compute_attachment_count(self):
        for record in self:
            record.attachment_count = self.env['ir.attachment'].search_count([
                ('res_model', '=', self._name),
                ('res_id', '=', record.id),
            ])

    def action_view_attachments(self):
        """Open attachments view."""
        self.ensure_one()
        return {
            'type': 'ir.actions.act_window',
            'name': 'Attachments',
            'res_model': 'ir.attachment',
            'view_mode': 'kanban,tree,form',
            'domain': [
                ('res_model', '=', self._name),
                ('res_id', '=', self.id),
            ],
            'context': {
                'default_res_model': self._name,
                'default_res_id': self.id,
            },
        }
```

---

## File Type Validation

### Validate File Extension
```python
from odoo.exceptions import ValidationError
import base64
import mimetypes


@api.constrains('document', 'document_name')
def _check_document(self):
    """Validate document type."""
    allowed_extensions = ['.pdf', '.doc', '.docx', '.xls', '.xlsx']
    allowed_mimetypes = [
        'application/pdf',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'application/vnd.ms-excel',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    ]

    for record in self:
        if record.document and record.document_name:
            # Check extension
            ext = '.' + record.document_name.rsplit('.', 1)[-1].lower()
            if ext not in allowed_extensions:
                raise ValidationError(
                    f"File type not allowed. Allowed: {', '.join(allowed_extensions)}"
                )

            # Check mimetype
            mimetype = mimetypes.guess_type(record.document_name)[0]
            if mimetype and mimetype not in allowed_mimetypes:
                raise ValidationError(f"Invalid file type: {mimetype}")
```

### Validate Image
```python
@api.constrains('image')
def _check_image(self):
    """Validate image format and size."""
    max_size = 10 * 1024 * 1024  # 10 MB
    allowed_formats = ['PNG', 'JPEG', 'JPG', 'GIF', 'WEBP']

    for record in self:
        if record.image:
            # Check size
            image_data = base64.b64decode(record.image)
            if len(image_data) > max_size:
                raise ValidationError(
                    f"Image too large. Maximum size: {max_size // 1024 // 1024} MB"
                )

            # Check format
            try:
                img = Image.open(BytesIO(image_data))
                if img.format.upper() not in allowed_formats:
                    raise ValidationError(
                        f"Invalid image format. Allowed: {', '.join(allowed_formats)}"
                    )
            except Exception as e:
                raise ValidationError(f"Invalid image: {str(e)}")
```

---

## Export/Import Binary Data

### Export Attachment to File
```python
import os


def export_attachments(self, path):
    """Export all attachments to filesystem."""
    attachments = self.env['ir.attachment'].search([
        ('res_model', '=', self._name),
        ('res_id', '=', self.id),
    ])

    for attachment in attachments:
        file_path = os.path.join(path, attachment.name)
        with open(file_path, 'wb') as f:
            f.write(base64.b64decode(attachment.datas))

    return len(attachments)
```

### Import Files as Attachments
```python
def import_files(self, file_paths):
    """Import files as attachments."""
    attachments = []

    for file_path in file_paths:
        with open(file_path, 'rb') as f:
            content = f.read()

        attachment = self.env['ir.attachment'].create({
            'name': os.path.basename(file_path),
            'type': 'binary',
            'datas': base64.b64encode(content),
            'res_model': self._name,
            'res_id': self.id,
        })
        attachments.append(attachment.id)

    return attachments
```

---

## Best Practices

1. **Use Image field for images** - Auto-resize and optimization
2. **Store as attachment** - `attachment=True` for large files
3. **Keep filename field** - Pair Binary with Char for filename
4. **Validate file types** - Security against malicious uploads
5. **Set size limits** - Prevent memory issues
6. **Generate thumbnails** - Multiple sizes for performance
7. **Use ir.attachment** - For multiple files per record
8. **Clean up orphans** - Remove unused attachments
9. **Check access rights** - Verify permissions on download
10. **Handle encoding** - Always use base64 for binary transport

---


## Source: report-patterns.md

# Report Generation Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  REPORT PATTERNS                                                             ║
║  PDF reports, QWeb templates, and document generation                        ║
║  Use for invoices, delivery slips, quotes, and custom documents              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Report Types

| Type | Use Case | Technology |
|------|----------|------------|
| PDF | Printable documents | QWeb + wkhtmltopdf |
| HTML | Screen display | QWeb |
| Excel | Data export | xlsxwriter/openpyxl |

---

## Basic Report Structure

### File Organization
```
my_module/
├── report/
│   ├── __init__.py
│   ├── my_report.py           # Report logic (optional)
│   └── my_report_templates.xml # QWeb templates
└── __manifest__.py
```

### Manifest Entry
```python
{
    'data': [
        'report/my_report_templates.xml',
    ],
}
```

---

## PDF Report Definition

### Report Action (XML)
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- Report Action -->
    <record id="report_my_model" model="ir.actions.report">
        <field name="name">My Report</field>
        <field name="model">my.model</field>
        <field name="report_type">qweb-pdf</field>
        <field name="report_name">my_module.report_my_model_document</field>
        <field name="report_file">my_module.report_my_model_document</field>
        <field name="print_report_name">'MyReport - %s' % object.name</field>
        <field name="binding_model_id" ref="model_my_model"/>
        <field name="binding_type">report</field>
        <field name="paperformat_id" ref="base.paperformat_euro"/>
    </record>

    <!-- Report Template -->
    <template id="report_my_model_document">
        <t t-call="web.html_container">
            <t t-foreach="docs" t-as="doc">
                <t t-call="web.external_layout">
                    <div class="page">
                        <h2>
                            <span t-field="doc.name"/>
                        </h2>

                        <div class="row mt-4">
                            <div class="col-6">
                                <strong>Customer:</strong>
                                <span t-field="doc.partner_id.name"/>
                            </div>
                            <div class="col-6 text-end">
                                <strong>Date:</strong>
                                <span t-field="doc.date"/>
                            </div>
                        </div>

                        <table class="table table-sm mt-4">
                            <thead>
                                <tr>
                                    <th>Description</th>
                                    <th class="text-end">Quantity</th>
                                    <th class="text-end">Price</th>
                                    <th class="text-end">Total</th>
                                </tr>
                            </thead>
                            <tbody>
                                <t t-foreach="doc.line_ids" t-as="line">
                                    <tr>
                                        <td><span t-field="line.name"/></td>
                                        <td class="text-end">
                                            <span t-field="line.quantity"/>
                                        </td>
                                        <td class="text-end">
                                            <span t-field="line.price_unit"
                                                  t-options='{"widget": "monetary",
                                                              "display_currency": doc.currency_id}'/>
                                        </td>
                                        <td class="text-end">
                                            <span t-field="line.subtotal"
                                                  t-options='{"widget": "monetary",
                                                              "display_currency": doc.currency_id}'/>
                                        </td>
                                    </tr>
                                </t>
                            </tbody>
                            <tfoot>
                                <tr>
                                    <td colspan="3" class="text-end">
                                        <strong>Total:</strong>
                                    </td>
                                    <td class="text-end">
                                        <strong>
                                            <span t-field="doc.amount_total"
                                                  t-options='{"widget": "monetary",
                                                              "display_currency": doc.currency_id}'/>
                                        </strong>
                                    </td>
                                </tr>
                            </tfoot>
                        </table>

                        <div t-if="doc.notes" class="mt-4">
                            <strong>Notes:</strong>
                            <p t-field="doc.notes"/>
                        </div>
                    </div>
                </t>
            </t>
        </t>
    </template>
</odoo>
```

---

## QWeb Template Syntax

### Basic Output
```xml
<!-- Text content -->
<span t-field="doc.name"/>

<!-- With formatting -->
<span t-field="doc.date" t-options='{"format": "dd/MM/yyyy"}'/>

<!-- Raw output (no escaping) -->
<span t-out="doc.html_content"/>

<!-- Escaped output -->
<span t-esc="doc.description"/>
```

### Field Widgets
```xml
<!-- Monetary -->
<span t-field="doc.amount"
      t-options='{"widget": "monetary",
                  "display_currency": doc.currency_id}'/>

<!-- Date -->
<span t-field="doc.date"
      t-options='{"format": "MMMM dd, yyyy"}'/>

<!-- Float precision -->
<span t-field="doc.quantity"
      t-options='{"precision": 2}'/>

<!-- Duration -->
<span t-field="doc.duration"
      t-options='{"widget": "duration",
                  "unit": "hour"}'/>

<!-- Image -->
<img t-att-src="image_data_uri(doc.image)"
     style="max-width: 200px;"/>

<!-- Barcode -->
<img t-att-src="'/report/barcode/?barcode_type=Code128&amp;value=%s&amp;width=200&amp;height=50' % doc.code"/>
```

### Conditionals
```xml
<!-- If -->
<div t-if="doc.state == 'draft'">
    <span class="badge bg-secondary">Draft</span>
</div>

<!-- If-Else -->
<t t-if="doc.amount > 0">
    <span class="text-success" t-field="doc.amount"/>
</t>
<t t-else="">
    <span class="text-muted">0.00</span>
</t>

<!-- If-Elif-Else -->
<t t-if="doc.state == 'done'">Done</t>
<t t-elif="doc.state == 'cancel'">Cancelled</t>
<t t-else="">In Progress</t>
```

### Loops
```xml
<!-- Basic loop -->
<t t-foreach="doc.line_ids" t-as="line">
    <tr>
        <td t-esc="line.name"/>
        <td t-esc="line.quantity"/>
    </tr>
</t>

<!-- Loop with index -->
<t t-foreach="doc.line_ids" t-as="line">
    <tr>
        <td t-esc="line_index + 1"/>  <!-- 0-based index -->
        <td t-esc="line.name"/>
    </tr>
</t>

<!-- Loop variables -->
<!-- line_index: current index (0-based) -->
<!-- line_first: True if first iteration -->
<!-- line_last: True if last iteration -->
<!-- line_odd: True if odd iteration (1-based) -->
<!-- line_even: True if even iteration -->
<!-- line_size: total items -->
```

### Attributes
```xml
<!-- Dynamic class -->
<tr t-att-class="'table-danger' if line.amount &lt; 0 else ''">

<!-- Dynamic style -->
<span t-att-style="'color: red;' if doc.overdue else ''"/>

<!-- Multiple attributes -->
<div t-attf-class="alert alert-#{doc.state == 'done' and 'success' or 'warning'}"/>
```

### Variables
```xml
<!-- Set variable -->
<t t-set="total" t-value="sum(doc.line_ids.mapped('amount'))"/>
<span t-esc="total"/>

<!-- String formatting -->
<t t-set="title" t-value="'Invoice %s' % doc.name"/>
```

---

## External Layout

### Using Company Header/Footer
```xml
<template id="report_my_document">
    <t t-call="web.html_container">
        <t t-foreach="docs" t-as="doc">
            <!-- Uses company letterhead -->
            <t t-call="web.external_layout">
                <div class="page">
                    <!-- Your content here -->
                </div>
            </t>
        </t>
    </t>
</template>
```

### Minimal Layout (No Header/Footer)
```xml
<template id="report_my_document">
    <t t-call="web.html_container">
        <t t-foreach="docs" t-as="doc">
            <t t-call="web.basic_layout">
                <div class="page">
                    <!-- Your content here -->
                </div>
            </t>
        </t>
    </t>
</template>
```

---

## Custom Report Logic

### Python Report Model
```python
# report/my_report.py
from odoo import api, models


class MyReport(models.AbstractModel):
    _name = 'report.my_module.report_my_model_document'
    _description = 'My Report'

    @api.model
    def _get_report_values(self, docids, data=None):
        """Prepare report data."""
        docs = self.env['my.model'].browse(docids)

        # Calculate totals
        totals = {
            'amount': sum(docs.mapped('amount_total')),
            'count': len(docs),
        }

        # Get additional data
        categories = docs.mapped('category_id')

        return {
            'doc_ids': docids,
            'doc_model': 'my.model',
            'docs': docs,
            'data': data,
            'totals': totals,
            'categories': categories,
            'company': self.env.company,
            'format_amount': self._format_amount,
        }

    def _format_amount(self, amount, currency):
        """Format amount with currency."""
        return currency.format(amount)
```

### Using Custom Values in Template
```xml
<template id="report_my_model_document">
    <t t-call="web.html_container">
        <t t-foreach="docs" t-as="doc">
            <t t-call="web.external_layout">
                <div class="page">
                    <!-- Use custom values -->
                    <p>Total Records: <t t-esc="totals['count']"/></p>
                    <p>Grand Total: <t t-esc="format_amount(totals['amount'], doc.currency_id)"/></p>

                    <!-- Access categories -->
                    <ul>
                        <t t-foreach="categories" t-as="cat">
                            <li t-esc="cat.name"/>
                        </t>
                    </ul>
                </div>
            </t>
        </t>
    </t>
</template>
```

---

## Paper Format

### Define Custom Paper Format
```xml
<record id="paperformat_custom" model="report.paperformat">
    <field name="name">Custom Format</field>
    <field name="default" eval="False"/>
    <field name="format">A4</field>
    <field name="orientation">Portrait</field>
    <field name="margin_top">40</field>
    <field name="margin_bottom">20</field>
    <field name="margin_left">7</field>
    <field name="margin_right">7</field>
    <field name="header_line" eval="False"/>
    <field name="header_spacing">35</field>
    <field name="dpi">90</field>
</record>

<!-- Use in report -->
<record id="report_my_model" model="ir.actions.report">
    <field name="paperformat_id" ref="paperformat_custom"/>
</record>
```

### Standard Paper Formats
- `base.paperformat_euro` - A4 Portrait
- `base.paperformat_us` - Letter Portrait

---

## Print from Python

### Single Record
```python
def action_print(self):
    """Print report for current record."""
    return self.env.ref('my_module.report_my_model').report_action(self)
```

### Multiple Records
```python
def action_print_selected(self):
    """Print report for selected records."""
    return self.env.ref('my_module.report_my_model').report_action(self.ids)
```

### With Custom Data
```python
def action_print_with_data(self):
    """Print with custom parameters."""
    data = {
        'date_from': self.date_from,
        'date_to': self.date_to,
        'include_draft': self.include_draft,
    }
    return self.env.ref('my_module.report_my_model').report_action(
        self, data=data
    )
```

---

## CSS Styling

### Inline Styles
```xml
<style>
    .my-report-table {
        width: 100%;
        border-collapse: collapse;
    }
    .my-report-table th {
        background-color: #f5f5f5;
        border-bottom: 2px solid #333;
    }
    .my-report-table td {
        border-bottom: 1px solid #ddd;
        padding: 8px;
    }
    .page-break {
        page-break-after: always;
    }
</style>
```

### External Stylesheet
```xml
<!-- In manifest assets -->
'assets': {
    'web.report_assets_common': [
        'my_module/static/src/scss/report.scss',
    ],
},
```

---

## Multi-Page Reports

### Page Breaks
```xml
<t t-foreach="docs" t-as="doc">
    <div class="page">
        <!-- Page content -->
    </div>
    <t t-if="not doc_last">
        <div class="page-break"/>
    </t>
</t>
```

### Grouped Reports
```xml
<t t-set="grouped" t-value="docs.grouped('category_id')"/>
<t t-foreach="grouped.items()" t-as="group">
    <div class="page">
        <h2 t-esc="group[0].name or 'Uncategorized'"/>
        <t t-foreach="group[1]" t-as="doc">
            <!-- Document content -->
        </t>
    </div>
</t>
```

---

## Report Inheritance

### Extend Existing Report
```xml
<template id="report_invoice_document_inherit"
          inherit_id="account.report_invoice_document">
    <xpath expr="//div[@name='invoice_address']" position="after">
        <div class="col-6">
            <strong>Custom Field:</strong>
            <span t-field="o.x_custom_field"/>
        </div>
    </xpath>
</template>
```

---

## Best Practices

1. **Use external_layout** for professional documents with company header
2. **Test PDF rendering** - wkhtmltopdf may render differently than browser
3. **Handle empty data** - Use `t-if` to check before displaying
4. **Format currencies properly** - Always use monetary widget with currency
5. **Escape user content** - Use `t-esc` or `t-field` to prevent XSS
6. **Add page breaks** - Use CSS `page-break-after: always` between records
7. **Optimize images** - Resize before including in reports
8. **Test multi-record** - Verify reports work with multiple records

---


## Source: multi-company-patterns.md

# Multi-Company and Multi-Currency Patterns

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  MULTI-COMPANY & MULTI-CURRENCY PATTERNS                                     ║
║  Company-aware models, cross-company rules, and currency handling            ║
║  Use for enterprise deployments with multiple business units                 ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Multi-Company Model Setup

### Basic Company-Aware Model (v18+)
```python
from odoo import api, fields, models


class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'
    _check_company_auto = True  # v18+ automatic company checking

    name = fields.Char(string='Name', required=True)
    company_id = fields.Many2one(
        comodel_name='res.company',
        string='Company',
        required=True,
        default=lambda self: self.env.company,
        index=True,
    )

    # Related records with company check
    partner_id = fields.Many2one(
        comodel_name='res.partner',
        string='Partner',
        check_company=True,  # Enforces same company
    )
    warehouse_id = fields.Many2one(
        comodel_name='stock.warehouse',
        string='Warehouse',
        check_company=True,
    )
```

### v17 and Earlier Pattern
```python
class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'

    company_id = fields.Many2one(
        comodel_name='res.company',
        string='Company',
        required=True,
        default=lambda self: self.env.company,
    )
    partner_id = fields.Many2one(
        comodel_name='res.partner',
        string='Partner',
        domain="[('company_id', 'in', [company_id, False])]",
    )

    @api.constrains('partner_id', 'company_id')
    def _check_company(self):
        for record in self:
            if record.partner_id.company_id and \
               record.partner_id.company_id != record.company_id:
                raise ValidationError(
                    "Partner company must match record company."
                )
```

---

## Company-Dependent Fields

### Company-Specific Values
```python
class ProductTemplate(models.Model):
    _inherit = 'product.template'

    # Different value per company
    x_internal_code = fields.Char(
        string='Internal Code',
        company_dependent=True,
    )
    x_local_price = fields.Float(
        string='Local Price',
        company_dependent=True,
        digits='Product Price',
    )
    x_local_supplier_id = fields.Many2one(
        comodel_name='res.partner',
        string='Local Supplier',
        company_dependent=True,
    )
```

### Accessing Company-Dependent Values
```python
def get_local_data(self):
    """Get company-specific values."""
    # Automatically returns value for current company
    code = self.x_internal_code

    # Get for specific company
    other_company = self.env['res.company'].browse(2)
    code_other = self.with_company(other_company).x_internal_code
```

---

## Record Rules for Multi-Company

### Basic Company Rule
```xml
<?xml version="1.0" encoding="utf-8"?>
<odoo>
    <!-- Users see only their company's records -->
    <record id="my_model_company_rule" model="ir.rule">
        <field name="name">My Model: Company Rule</field>
        <field name="model_id" ref="model_my_model"/>
        <field name="domain_force">[
            '|',
            ('company_id', '=', False),
            ('company_id', 'in', company_ids)
        ]</field>
        <field name="global" eval="True"/>
    </record>
</odoo>
```

### Multi-Company Access Patterns
```xml
<!-- Strict: Only own company -->
<field name="domain_force">[('company_id', '=', company_id)]</field>

<!-- Flexible: Own companies or no company -->
<field name="domain_force">[
    '|',
    ('company_id', '=', False),
    ('company_id', 'in', company_ids)
]</field>

<!-- Child companies included -->
<field name="domain_force">[
    ('company_id', 'child_of', company_id)
]</field>
```

---

## Cross-Company Operations

### Switch Company Context
```python
def action_process_all_companies(self):
    """Process records across all user's companies."""
    for company in self.env.user.company_ids:
        records = self.with_company(company).search([
            ('state', '=', 'pending'),
            ('company_id', '=', company.id),
        ])
        for record in records:
            record._process()
```

### Create in Specific Company
```python
def action_create_in_company(self, company_id: int) -> 'my.model':
    """Create record in specific company."""
    company = self.env['res.company'].browse(company_id)

    return self.with_company(company).create({
        'name': 'New Record',
        'company_id': company.id,
    })
```

### Inter-Company Transactions
```python
class InterCompanyTransfer(models.Model):
    _name = 'inter.company.transfer'
    _description = 'Inter-Company Transfer'

    source_company_id = fields.Many2one(
        comodel_name='res.company',
        string='Source Company',
        required=True,
    )
    dest_company_id = fields.Many2one(
        comodel_name='res.company',
        string='Destination Company',
        required=True,
    )

    def action_transfer(self):
        """Execute inter-company transfer."""
        self.ensure_one()

        # Create in source company
        source_record = self.with_company(self.source_company_id).sudo().create({
            'name': f'Transfer to {self.dest_company_id.name}',
            'type': 'outgoing',
            'company_id': self.source_company_id.id,
        })

        # Create in destination company
        dest_record = self.with_company(self.dest_company_id).sudo().create({
            'name': f'Transfer from {self.source_company_id.name}',
            'type': 'incoming',
            'company_id': self.dest_company_id.id,
            'source_ref': source_record.id,
        })

        return source_record, dest_record
```

---

## Multi-Currency Support

### Currency Fields
```python
class MyModel(models.Model):
    _name = 'my.model'
    _description = 'My Model'

    company_id = fields.Many2one(
        comodel_name='res.company',
        default=lambda self: self.env.company,
    )
    currency_id = fields.Many2one(
        comodel_name='res.currency',
        string='Currency',
        default=lambda self: self.env.company.currency_id,
        required=True,
    )

    # Monetary fields
    amount = fields.Monetary(
        string='Amount',
        currency_field='currency_id',
    )
    amount_tax = fields.Monetary(
        string='Tax Amount',
        currency_field='currency_id',
    )
    amount_total = fields.Monetary(
        string='Total',
        currency_field='currency_id',
        compute='_compute_amount_total',
        store=True,
    )

    # Company currency equivalent
    company_currency_id = fields.Many2one(
        related='company_id.currency_id',
        string='Company Currency',
    )
    amount_company_currency = fields.Monetary(
        string='Amount (Company Currency)',
        currency_field='company_currency_id',
        compute='_compute_amount_company_currency',
        store=True,
    )

    @api.depends('amount', 'amount_tax')
    def _compute_amount_total(self):
        for record in self:
            record.amount_total = record.amount + record.amount_tax

    @api.depends('amount_total', 'currency_id', 'company_id', 'date')
    def _compute_amount_company_currency(self):
        for record in self:
            if record.currency_id != record.company_currency_id:
                record.amount_company_currency = record.currency_id._convert(
                    record.amount_total,
                    record.company_currency_id,
                    record.company_id,
                    record.date or fields.Date.today(),
                )
            else:
                record.amount_company_currency = record.amount_total
```

### Currency Conversion
```python
def convert_to_currency(self, amount: float, target_currency) -> float:
    """Convert amount to target currency."""
    if self.currency_id == target_currency:
        return amount

    return self.currency_id._convert(
        amount,
        target_currency,
        self.company_id,
        self.date or fields.Date.today(),
    )

def get_rate(self) -> float:
    """Get exchange rate to company currency."""
    return self.currency_id._get_conversion_rate(
        self.currency_id,
        self.company_currency_id,
        self.company_id,
        self.date or fields.Date.today(),
    )
```

### Multi-Currency Reporting
```python
def _get_amounts_by_currency(self) -> dict:
    """Group amounts by currency for reporting."""
    result = {}
    for record in self:
        currency = record.currency_id
        if currency not in result:
            result[currency] = {
                'amount': 0.0,
                'amount_company': 0.0,
            }
        result[currency]['amount'] += record.amount_total
        result[currency]['amount_company'] += record.amount_company_currency
    return result
```

---

## Views for Multi-Company

### Form View with Company
```xml
<form string="My Model">
    <sheet>
        <group>
            <group>
                <field name="name"/>
                <field name="partner_id"
                       context="{'default_company_id': company_id}"
                       domain="[('company_id', 'in', [company_id, False])]"/>
            </group>
            <group>
                <field name="company_id"
                       groups="base.group_multi_company"
                       options="{'no_create': True}"/>
                <field name="currency_id"
                       groups="base.group_multi_currency"/>
            </group>
        </group>
        <group string="Amounts">
            <field name="amount"/>
            <field name="amount_total"/>
            <field name="amount_company_currency"
                   groups="base.group_multi_currency"
                   invisible="currency_id == company_currency_id"/>
        </group>
    </sheet>
</form>
```

### Search View with Company Filter
```xml
<search string="My Model">
    <field name="name"/>
    <field name="partner_id"/>
    <filter string="My Company" name="my_company"
            domain="[('company_id', '=', company_id)]"/>
    <group expand="0" string="Group By">
        <filter string="Company" name="group_company"
                context="{'group_by': 'company_id'}"
                groups="base.group_multi_company"/>
        <filter string="Currency" name="group_currency"
                context="{'group_by': 'currency_id'}"
                groups="base.group_multi_currency"/>
    </group>
</search>
```

---

## Scheduled Actions (Multi-Company)

### Process Each Company
```python
@api.model
def _cron_process_all_companies(self) -> None:
    """Cron that processes each company separately."""
    companies = self.env['res.company'].search([])

    for company in companies:
        self.with_company(company)._process_company_records()

def _process_company_records(self) -> None:
    """Process records for current company context."""
    records = self.search([
        ('company_id', '=', self.env.company.id),
        ('state', '=', 'pending'),
    ])

    for record in records:
        try:
            record._do_process()
        except Exception as e:
            _logger.error(f"Error processing {record.id}: {e}")
```

---

## Best Practices

### 1. Always Include company_id
```python
# Good - explicit company
company_id = fields.Many2one(
    'res.company',
    required=True,
    default=lambda self: self.env.company,
)

# Bad - no company field on business model
```

### 2. Use check_company (v18+)
```python
# Good - automatic validation
partner_id = fields.Many2one('res.partner', check_company=True)

# Manual (older versions)
@api.constrains('partner_id', 'company_id')
def _check_company(self):
    ...
```

### 3. Use with_company() for Context
```python
# Good - explicit company context
record.with_company(company)._process()

# Avoid - changing env.company directly
```

### 4. Handle Shared Records
```python
# Allow records without company (shared)
domain = [
    '|',
    ('company_id', '=', False),
    ('company_id', '=', self.env.company.id),
]
```

### 5. Currency Conversions
```python
# Always specify date for conversions
converted = currency._convert(
    amount,
    target_currency,
    company,
    date,  # Required for correct rate
)
```

---

## Version Differences

| Feature | v14-16 | v17 | v18+ |
|---------|--------|-----|------|
| Company check | Manual `@api.constrains` | Manual | `_check_company_auto = True` |
| Field validation | `domain=` | `domain=` | `check_company=True` |
| Company switch | `with_context(force_company=)` | `with_company()` | `with_company()` |
| Multi-company views | `groups="base.group_multi_company"` | Same | Same |

---


# Skill: Odoo Webhook Automation (v19)

## 1. Description
The `odoo-webhooks-automation-19` skill defines patterns for native webhook ingestion in Odoo 19 via `base_automation`. It replaces legacy custom HTTP controllers with declarative automation rules.

## 2. Core Mandatory Rules
- **No HTTP Controllers**: Prohibited to create `@http.route` for simple data ingestion.
- **Native Trigger**: MUST use `trigger="on_webhook"`.
- **Payload Handling**: Use the built-in `payload` object. Do not parse request bodies manually.

## 3. Implementation Patterns (Bad vs Good)

### A. Ingestion Pattern
```python
# ❌ LEGACY: Manual HTTP Controller
@http.route('/api/ingest', type='json', auth='none')
def ingest(self, **kw):
    # Logic...

# ✅ v19 STANDARD: Native base.automation
# Automation Rule Definition:
automation = self.env["base.automation"].create({
    "name": "Webhook Ingest",
    "trigger": "on_webhook",
    "record_getter": "model.search([('ref', '=', payload.get('ref'))], limit=1)",
    "action_server_ids": [Command.link(action_id)]
})
```

### B. Payload Access
```python
# ❌ LEGACY: Parsing request objects manually
data = json.loads(request.httprequest.data)

# ✅ v19 STANDARD: Direct native payload access
# Access 'payload' directly in the action scope
amount = payload.get('monto') 
```

## 4. Verification Workflow
- Check that no custom controllers exist for simple webhook integrations.
- Audit `base.automation` rules to ensure `record_getter` is performant (uses indexing).

## 5. Maintenance
- Ensure `payload` mapping is documented for external system integrations.
- Update `record_getter` expressions if schema changes.

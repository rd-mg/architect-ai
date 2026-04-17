# Odoo 19 Exclusive Features

Consolidated v19-specific micro-skills covering features that only exist in v19.0.


---

## Source: odoo-ai-server-actions-19.md

# Skill: Odoo AI Server Actions (v19)

## 1. Description
This skill defines architectural mandates for Odoo 19 native AI integration using the `ai.agent` framework. It ensures agents utilize the internal server action tool-calling mechanism instead of external, non-native APIs.

## 2. Core Mandatory Rules
- **Native Agents**: Only use `ai.agent` models. Never hallucinate third-party LLM service integrations.
- **Tool Calling**: Server Actions used as AI tools MUST set `use_in_ai=True`.
- **Result Injection**: NEVER use `return` statements in Server Actions. Use `ai['result'] = ...` to return data to the LLM.

## 3. Implementation Patterns (Bad vs Good)

### A. Exposing Server Actions as AI Tools
```python
# ❌ LEGACY: External API approach / Returning data
def get_info(self):
    return "Some data"

# ✅ v19 STANDARD: Native AI Server Action
# XML Definition: <field name="use_in_ai" eval="True"/>
# Python Code:
ai['result'] = f"Info: {self.get_info()}"
```

### B. AI Computed Fields
```python
# ❌ BAD: Manual logic in Python
def compute_summary(self):
    self.summary = openai.create_summary(self.description)

# ✅ v19 STANDARD: Declarative AI Automation
# Server Action Field:
# <field name="evaluation_type">ai_computed</field>
# <field name="ai_update_prompt">Summarize this text: /field</field>
```

## 4. Verification Workflow
- Ensure all AI-exposed actions define a JSON `ai_tool_schema` for input validation.
- Verify that `ai['result']` is populated in every tool-called method.

## 5. Maintenance
- Sync agent prompts with business requirements in `ai.topic`.
- Validate tool schema changes against native Odoo 19 `ai_server_actions` modules.

---

## Source: odoo-auth-passkeys-19.md

# Skill: Odoo Auth Passkeys (v19)

## 1. Description
The `odoo-auth-passkeys-19` skill defines the architectural mandates for WebAuthn/Passkey integration. It standardizes the two-step verification flow and prevents manual implementation of crypto primitives.

## 2. Core Mandatory Rules
- **Crypto Abstraction**: Agents MUST NOT implement manual WebAuthn/Crypto logic. Use `simplewebauthn` library.
- **Two-Step Auth**: Sensitive actions MUST use the `res.users.identitycheck` transient model flow.
- **RPC Safety**: Challenges MUST be fetched via secure RPC endpoints (`/auth/passkey/start-auth`).

## 3. Implementation Patterns (Bad vs Good)

### A. Frontend WebAuthn Flow
```javascript
// ❌ LEGACY: Manual implementation of browser crypto APIs
navigator.credentials.create(...) // Risk of implementation errors

// ✅ v19 STANDARD: Odoo native wrapper (simplewebauthn)
const serverOptions = await rpc("/auth/passkey/start-auth");
const auth = await passkeyLib.startAuthentication(serverOptions);
this.model.root.update({ password: JSON.stringify(auth) });
```

### B. Backend Identity Verification
```xml
<!-- ✅ v19 STANDARD: Integration in IdentityCheck views -->
<xpath expr="//footer/button[@id='password_confirm']" position="before">
    <button string="Use Passkey" type="object" name="run_check" class="btn btn-primary" 
            invisible="auth_method != 'webauthn'" context="{'password': password}"/>
</xpath>
```

## 4. Verification Workflow
- Ensure all sensitive `action_button` calls check `auth_method == 'webauthn'` in the view logic.
- Validate that identity check views override the JS controller with the v19 `auth_passkey_identity_check_view_form` class.

## 5. Maintenance
- Monitor security patches for `simplewebauthn` dependency.
- Validate Passkey identity flows whenever security policies change in the `res.users.identitycheck` model.

---

## Source: odoo-hoot-testing-19.md

# Skill: Odoo Hoot Testing Framework (v19)

## 1. Description
The `odoo-hoot-testing-19` skill defines the architectural mandates for the new reactive frontend testing framework, replacing the legacy QUnit runner.

## 2. Core Mandatory Rules
- **Framework**: Prohibited to use `QUnit` or `web.test_utils`. Use `@odoo/hoot`.
- **Async Pattern**: All DOM interactions (`click`, `edit`) MUST be `await`-ed.
- **Reactivity Flush**: Mandatory `await animationFrame()` after state changes.

## 3. Implementation Patterns (Bad vs Good)

### A. Test Structure
```javascript
// ❌ LEGACY: QUnit approach (deprecated)
QUnit.test("my test", async (assert) => { ... });

// ✅ v19 STANDARD: Hoot structure
import { describe, expect, test } from "@odoo/hoot";

test("My Component Test", async () => {
    // Test logic here
});
```

### B. Reactive Reactivity Flush
```javascript
// ❌ LEGACY: Assuming sync DOM updates
await click(".o_save");
expect(".o_form").toHaveText("Saved");

// ✅ v19 STANDARD: Explicitly wait for OWL reactivity
await click(".o_save");
await animationFrame(); // Flush reactive updates
expect(".o_form").toHaveText("Saved");
```

## 4. Verification Workflow
- Ensure all tests are tagged with environment tags (`describe.current.tags("desktop")`).
- Validate that mocking uses `hoot-mock` modules (e.g., `mockService`, `onRpc`, `mockDate`).

## 5. Maintenance
- Track breaking changes in `@odoo/hoot` API.
- Migrate remaining legacy tours as Hoot becomes mandatory in v19.1+.

---

## Source: odoo-iot-testing-19.md

# Odoo IoT Testing - Version 19.0 (Hoot)

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ODOO 19.0 IOT TESTING PATTERNS                                              ║
║  MANDATORY: MockHttpServiceDummy, animationFrame usage                       ║
║  VERIFY: Eventual consistency in real-time streams                           ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## 1. Mocking Hardware Components
Avoid using physical hardware in tests. Use Dummies and MockServices to simulate async IoT streams.

### MANDATORY RULE
Simulate success/failure with `setTimeout` and `Promise.resolve()` in mock services.

```javascript
// ✅ GOOD: Mocking IoT Service
class IotHttpServiceDummy {
    action(iotBoxId, deviceId, data, onSuccess) {
        setTimeout(() => onSuccess({ 
            status: { status: "connected" }, 
            result: 2.35 
        }), 1000);
    }
}
```

## 2. Testing Real-time Consistency
Tests must use `await animationFrame()` to flush OWL reactivity after IoT events.

```javascript
test("IoT Scale Measurement", async () => {
    // 1. Setup mock
    const myScale = new PosScaleDummy();
    myScale.iotHttpService = new IotHttpServiceDummy();
    
    // 2. Action
    await click(".o_iot_button");
    
    // 3. MANDATORY: Wait for reactivity
    await animationFrame();
    
    // 4. Assert
    expect(".gross-weight").toHaveText("2.35");
});
```

---

## Source: odoo-orm-extreme-19.md

# Skill: Odoo ORM Extreme Performance (v19)

## 1. Description
The `odoo-orm-extreme-19` skill defines the architectural mandates for ORM efficiency in Odoo 19. It enforces patterns that eliminate N+1 queries and high memory consumption.

## 2. Core Mandatory Rules
- **Indexing**: Composite and complex indexes **MUST** be defined declaratively via `models.Index`.
- **Query Efficiency**: `read_group()` is deprecated for backend logic. Use `_read_group()`.
- **Prefetching**: Use `search_fetch()` to perform search and prefetch in a single database round-trip.

## 3. Implementation Patterns (Bad vs Good)

### A. Declarative Indexing
```python
# ❌ LEGACY: Field-level index=True
name = fields.Char(index=True)

# ✅ v19 STANDARD: Declarative Index
class MyModel(models.Model):
    _indexes = [
        models.Index(name='custom_idx', expressions=['company_id', 'state'])
    ]
```

### B. Aggregation (The Performance Standard)
```python
# ❌ LEGACY: Expensive UI-metadata aggregation
results = self.env['model'].read_group(domain, ['amount:sum'], ['partner_id'])

# ✅ v19 STANDARD: Efficient Tuple Aggregation
results = self.env['model']._read_group(
    domain, 
    groupby=['partner_id'], 
    aggregates=['amount:sum']
)
```

## 4. Verification Workflow
- Ensure all custom modules replace `read_group` with `_read_group`.
- Audit database usage to ensure composite search queries use `_indexes`.

## 5. Maintenance
- Regularly audit performance-critical methods using `search_fetch()`.
- Keep the `ORM Performance Guide` in sync with Odoo 19 `models.py` structural changes.

---

## Source: odoo-security-hardening-19.md

# Skill: Odoo Security Hardening (v19)

## 1. Description
The `odoo-security-hardening-19` skill defines the architectural mandates for secure module development in Odoo 19. It enforces the transition from legacy security conventions to explicit API-level security.

## 2. Core Mandatory Rules
- **Method Exposure**: The underscore prefix (`_`) is officially insufficient for RPC blocking. **MUST** use `@api.private`.
- **Privilege Escalation**: Assigning groups during `res.users.create()` is blocked by the ORM. **MUST** use a two-step `create()` then `write()` flow.

## 3. Implementation Patterns (Bad vs Good)

### A. Blocking RPC Access
```python
# ❌ LEGACY: Underscore prefix is no longer sufficient for security.
def _sensitive_action(self):
    pass

# ✅ v19 STANDARD: Explicit security gating
from odoo import api

@api.private
def _sensitive_action(self):
    # This method is now blocked from external RPC calls.
    pass
```

### B. Safe User Creation
```python
# ❌ LEGACY: Privilege escalation risk (Blocked in Odoo 19)
user = self.env['res.users'].create({
    'name': 'User',
    'groups_id': [Command.link(self.env.ref('base.group_system').id)]
})

# ✅ v19 STANDARD: Secure two-step approach
user = self.env['res.users'].create({'name': 'User'})
user.write({'groups_id': [Command.link(self.env.ref('base.group_system').id)]})
```

## 4. Verification Workflow
- Ensure all business-logic methods that shouldn't be exposed are decorated.
- Use static analysis (Compliance Checker) to detect usage of restricted ORM parameters in user creation.

## 5. Maintenance
- Monitor Odoo 19 `security/ir.model.access.csv` changes.
- Ensure all custom modules audit their `groups_id` assignments during creation.

---

## Source: odoo-v19-compliance-checker.py.md

# Skill: Odoo v19 Compliance Checker

## 1. Description
The `odoo-v19-compliance-checker` is a static analysis tool that scans Odoo modules for Odoo 19 mandatory breaking changes. It acts as an automated "Compliance Gate" to prevent non-compliant code (legacy SQL patterns, deprecated XML, missing types) from entering the codebase.

## 2. Mandatory Rules (Compliance Gate)
- **SQL Security**: Raw `cr.execute()` calls with formatted strings are strictly prohibited. The `SQL()` builder is mandatory for all database interactions.
- **Python Typing**: All public methods and `create` methods MUST include type hints (`-> 'ModelName'`, `-> bool`, etc.).
- **XML Deprecation**: The `<tree>` tag is removed; `<list>` must be used. The `attrs` attribute is deprecated in favor of direct view attributes.
- **ORM Patterns**: `_sql_constraints` is removed; `models.Constraint` is mandatory for data integrity.

## 3. Implementation Patterns
The checker utilizes Python's `ast` module to walk the Abstract Syntax Tree and `lxml` for XML structure validation.

```python
# AST Analysis Example for SQL() enforcement
def check_sql_builder(self, node):
    if isinstance(node, ast.Call) and getattr(node.func, 'attr', '') == 'execute':
        if not any(isinstance(arg, ast.Call) and getattr(arg.func, 'attr', None) == 'SQL' for arg in node.args):
             self.add_violation(file, node.lineno, "Use SQL() builder")
```

## 4. Verification Workflow
1. Execute `python odoo-v19-compliance-checker.py [module_path]`.
2. Review the JSON report output.
3. If `status: FAIL`, address all `CRITICAL` violations before further review.

## 5. Maintenance
- Add new v19 deprecation patterns to the `check_xml_file` or `check_python_file` methods as Odoo releases updates.
- Keep the checker synchronized with the official `models.py` structural changes in Odoo 19.

---

## Source: odoo-webhooks-automation-19.md

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

---

## Source: odoo-webrtc-iot-19.md

# Odoo WebRTC & IoT - Version 19.0

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  ODOO 19.0 IOT/WEBRTC PATTERNS                                               ║
║  MANDATORY: Fallback Pipeline (WebRTC -> Longpolling -> WebSocket)           ║
║  VERIFY: IotHttpService integration, chunked message handling                ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## 1. Connection Pattern (Fallback Pipeline)
Agents MUST NOT implement direct `RTCPeerConnection` for IoT. Use `IotHttpService` to handle connection resilience.

### MANDATORY RULE
Always use the service orchestrator which automatically degrades connection type if WebRTC fails.

```javascript
// ✅ GOOD: Pattern for IotHttpService
async setup({ iot_http }) {
    this.iot = iot_http;
}

async sendAction(iotBoxId, deviceId, data) {
    // IotHttpService manages: _webRtc -> _longpolling -> _websocket
    await this.iot.action(
        iotBoxId, 
        deviceId, 
        data, 
        (res) => console.log("Success:", res),
        (err) => console.error("Fallback triggered:", err)
    );
}
```

## 2. Chunking for Industrial Streams
Large industrial data payloads MUST be chunked if they exceed `sctp.maxMessageSize`.

```javascript
// ✅ GOOD: Use chunking patterns defined in IotWebRtc
if (messageString.length >= rtcConnection.connection.sctp.maxMessageSize) {
    this._sendChunkedMessage(rtcConnection, messageString);
} else {
    rtcConnection.channel.send(messageString);
}
```

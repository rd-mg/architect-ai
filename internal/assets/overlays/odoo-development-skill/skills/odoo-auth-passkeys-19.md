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

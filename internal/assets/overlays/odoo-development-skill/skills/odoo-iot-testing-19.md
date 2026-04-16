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

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

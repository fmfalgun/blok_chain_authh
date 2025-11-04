# Chaincode API Reference

## AS Chaincode API

### RegisterDevice

Registers a new IoT device with the authentication server.

**Function**: `RegisterDevice(deviceID, publicKey, metadata)`

**Parameters**:
- `deviceID` (string): Unique device identifier (3-64 chars, alphanumeric + _ -)
- `publicKey` (string): PEM-encoded public key (100-4096 chars)
- `metadata` (string): Optional device information (max 1024 chars)

**Returns**: None (success) or Error

**Example**:
```bash
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["RegisterDevice","device_001","-----BEGIN PUBLIC KEY-----...","IoT Sensor v1.0"]}'
```

**Success Response**: Transaction ID

**Error Responses**:
- `device already exists`: Device ID is already registered
- `deviceID must be between 3 and 64 characters`: Invalid length
- `publicKey length is invalid`: Public key size out of bounds

---

### Authenticate

Authenticates a device and issues a TGT.

**Function**: `Authenticate(authRequestJSON)`

**Parameters**:
- `authRequestJSON` (string): JSON object containing:
  ```json
  {
    "deviceID": "device_001",
    "nonce": "base64_encoded_secure_random",
    "timestamp": 1672531200,
    "signature": "base64_encoded_signature"
  }
  ```

**Returns**: JSON string containing TGT details

**Example**:
```bash
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["Authenticate","{\"deviceID\":\"device_001\",\"nonce\":\"abc123\",\"timestamp\":1672531200,\"signature\":\"xyz789\"}"]}'
```

**Success Response**:
```json
{
  "tgtID": "tgt_abc123",
  "sessionKey": "base64_encoded_key",
  "expiresAt": 1672534800,
  "message": "Authentication successful"
}
```

**Error Responses**:
- `device not found`: Device ID doesn't exist
- `device is not active`: Device has been revoked or suspended
- `timestamp is invalid or too old`: Timestamp outside 5-minute window
- `invalid signature`: Signature verification failed

---

### GetDevice

Retrieves device information.

**Function**: `GetDevice(deviceID)`

**Parameters**:
- `deviceID` (string): Device identifier

**Returns**: JSON string with device details

**Example**:
```bash
peer chaincode query -C authchannel -n as \
  -c '{"Args":["GetDevice","device_001"]}'
```

**Success Response**:
```json
{
  "deviceID": "device_001",
  "publicKey": "-----BEGIN PUBLIC KEY-----...",
  "status": "active",
  "registrationTime": 1672531200,
  "lastAuthTime": 1672531500,
  "metadata": "IoT Sensor v1.0"
}
```

---

### RevokeDevice

Revokes a device's access.

**Function**: `RevokeDevice(deviceID)`

**Parameters**:
- `deviceID` (string): Device identifier to revoke

**Returns**: None (success) or Error

**Example**:
```bash
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["RevokeDevice","device_001"]}'
```

---

### GetAllDevices

Returns all registered devices (admin function).

**Function**: `GetAllDevices()`

**Parameters**: None

**Returns**: JSON array of device objects

**Example**:
```bash
peer chaincode query -C authchannel -n as \
  -c '{"Args":["GetAllDevices"]}'
```

---

## TGS Chaincode API

### RegisterService

Registers a new service.

**Function**: `RegisterService(serviceID, serviceName, description, requiredRole)`

**Parameters**:
- `serviceID` (string): Unique service identifier
- `serviceName` (string): Human-readable service name
- `description` (string): Service description
- `requiredRole` (string): Required role (e.g., "user", "admin")

**Returns**: None (success) or Error

**Example**:
```bash
peer chaincode invoke -C authchannel -n tgs \
  -c '{"Args":["RegisterService","service001","Data Access","Access IoT data streams","user"]}'
```

---

### IssueServiceTicket

Issues a service ticket to a device.

**Function**: `IssueServiceTicket(ticketRequestJSON)`

**Parameters**:
- `ticketRequestJSON` (string): JSON object containing:
  ```json
  {
    "deviceID": "device_001",
    "tgtID": "tgt_abc123",
    "serviceID": "service001",
    "timestamp": 1672531200,
    "signature": "base64_encoded_signature"
  }
  ```

**Returns**: JSON string with service ticket details

**Example**:
```bash
peer chaincode invoke -C authchannel -n tgs \
  -c '{"Args":["IssueServiceTicket","{\"deviceID\":\"device_001\",\"tgtID\":\"tgt_abc123\",\"serviceID\":\"service001\",\"timestamp\":1672531200,\"signature\":\"xyz789\"}"]}'
```

**Success Response**:
```json
{
  "ticketID": "ticket_xyz789",
  "deviceID": "device_001",
  "serviceID": "service001",
  "serviceKey": "base64_encoded_key",
  "issuedAt": 1672531200,
  "expiresAt": 1672533000,
  "status": "valid",
  "usageCount": 0,
  "maxUsageCount": 10
}
```

---

### ValidateServiceTicket

Validates and increments usage count for a service ticket.

**Function**: `ValidateServiceTicket(ticketID)`

**Parameters**:
- `ticketID` (string): Service ticket identifier

**Returns**: JSON string with ticket details or error

**Example**:
```bash
peer chaincode invoke -C authchannel -n tgs \
  -c '{"Args":["ValidateServiceTicket","ticket_xyz789"]}'
```

---

### GetAllServices

Returns all available services.

**Function**: `GetAllServices()`

**Parameters**: None

**Returns**: JSON array of service objects

**Example**:
```bash
peer chaincode query -C authchannel -n tgs \
  -c '{"Args":["GetAllServices"]}'
```

---

## ISV Chaincode API

### ValidateAccess

Validates a device's access request to a service.

**Function**: `ValidateAccess(accessRequestJSON)`

**Parameters**:
- `accessRequestJSON` (string): JSON object containing:
  ```json
  {
    "deviceID": "device_001",
    "serviceID": "service001",
    "ticketID": "ticket_xyz789",
    "action": "read",
    "timestamp": 1672531200,
    "ipAddress": "192.168.1.100",
    "userAgent": "IoT-Device/1.0",
    "signature": "base64_encoded_signature"
  }
  ```

**Returns**: JSON string with access response

**Example**:
```bash
peer chaincode invoke -C authchannel -n isv \
  -c '{"Args":["ValidateAccess","{\"deviceID\":\"device_001\",\"serviceID\":\"service001\",\"ticketID\":\"ticket_xyz789\",\"action\":\"read\",\"timestamp\":1672531200,\"ipAddress\":\"192.168.1.100\",\"userAgent\":\"IoT-Device/1.0\",\"signature\":\"xyz789\"}"]}'
```

**Success Response**:
```json
{
  "granted": true,
  "sessionID": "session_def456",
  "message": "Access granted",
  "expiresAt": 1672533000
}
```

---

### TerminateSession

Terminates an active session.

**Function**: `TerminateSession(sessionID)`

**Parameters**:
- `sessionID` (string): Session identifier to terminate

**Returns**: None (success) or Error

**Example**:
```bash
peer chaincode invoke -C authchannel -n isv \
  -c '{"Args":["TerminateSession","session_def456"]}'
```

---

### GetAccessLogs

Retrieves access logs for a device.

**Function**: `GetAccessLogs(deviceID)`

**Parameters**:
- `deviceID` (string): Device identifier

**Returns**: JSON array of access log entries

**Example**:
```bash
peer chaincode query -C authchannel -n isv \
  -c '{"Args":["GetAccessLogs","device_001"]}'
```

**Success Response**:
```json
[
  {
    "logID": "log_abc123",
    "deviceID": "device_001",
    "serviceID": "service001",
    "ticketID": "ticket_xyz789",
    "timestamp": 1672531200,
    "action": "read",
    "status": "success",
    "ipAddress": "192.168.1.100",
    "userAgent": "IoT-Device/1.0",
    "description": "Access granted"
  }
]
```

---

### GetActiveSessions

Returns all currently active sessions.

**Function**: `GetActiveSessions()`

**Parameters**: None

**Returns**: JSON array of active session objects

**Example**:
```bash
peer chaincode query -C authchannel -n isv \
  -c '{"Args":["GetActiveSessions"]}'
```

---

## Error Codes

| Code | Description |
|------|-------------|
| 400  | Bad Request - Invalid input parameters |
| 401  | Unauthorized - Authentication failed |
| 403  | Forbidden - Access denied |
| 404  | Not Found - Resource doesn't exist |
| 429  | Too Many Requests - Rate limit exceeded |
| 500  | Internal Server Error - Chaincode execution error |

## Rate Limits

- Default: 60 requests per minute per device
- Violation threshold: 3 violations trigger 5-minute ban
- Reset period: 1 minute

## Security Best Practices

1. **Always use HTTPS/TLS** for client-chaincode communication
2. **Validate all timestamps** to prevent replay attacks
3. **Rotate session keys** regularly
4. **Monitor for unusual patterns** in access logs
5. **Implement client-side caching** to reduce redundant requests
6. **Use strong signatures** (RSA 2048+ or ECDSA P-256+)

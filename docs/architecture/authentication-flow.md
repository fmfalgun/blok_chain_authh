# Authentication Flow Architecture

## Overview

The blockchain-based authentication framework implements a Kerberos-inspired authentication mechanism for IoT devices using Hyperledger Fabric. The system consists of three main chaincode components working together to provide secure authentication and access control.

## System Components

### 1. AS Chaincode (Authentication Server)
- **Purpose**: Primary authentication and device registration
- **Organization**: Org1 (AS)
- **Key Functions**:
  - Device registration
  - Initial authentication
  - TGT (Ticket Granting Ticket) issuance
  - Device lifecycle management

### 2. TGS Chaincode (Ticket Granting Server)
- **Purpose**: Service ticket issuance
- **Organization**: Org2 (TGS)
- **Key Functions**:
  - Service registration
  - Service ticket issuance
  - Ticket validation
  - Service access control

### 3. ISV Chaincode (IoT Service Validator)
- **Purpose**: Access validation and session management
- **Organization**: Org3 (ISV)
- **Key Functions**:
  - Access request validation
  - Session management
  - Access logging
  - Real-time monitoring

## Authentication Flow

### Phase 1: Device Registration

```
IoT Device -> AS Chaincode
├── RegisterDevice(deviceID, publicKey, metadata)
├── Validate input parameters
├── Check for duplicate registration
├── Store device record on ledger
└── Return: Registration confirmation
```

**Data Structure**:
```json
{
  "deviceID": "device_001",
  "publicKey": "-----BEGIN PUBLIC KEY-----...",
  "status": "active",
  "registrationTime": 1672531200,
  "lastAuthTime": 0,
  "metadata": "IoT Sensor v1.0"
}
```

### Phase 2: Initial Authentication

```
IoT Device -> AS Chaincode
├── Authenticate(authRequest)
│   ├── deviceID
│   ├── nonce (secure random)
│   ├── timestamp
│   └── signature
├── Validate device exists and is active
├── Verify timestamp (within 5 minutes)
├── Verify signature using device public key
├── Generate secure session key (256-bit)
├── Create TGT with 1-hour validity
├── Store TGT on ledger
└── Return: TGT + Session Key
```

**TGT Structure**:
```json
{
  "tgtID": "tgt_abc123",
  "deviceID": "device_001",
  "sessionKey": "base64_encoded_256bit_key",
  "issuedAt": 1672531200,
  "expiresAt": 1672534800,
  "status": "valid"
}
```

### Phase 3: Service Ticket Request

```
IoT Device -> TGS Chaincode
├── IssueServiceTicket(ticketRequest)
│   ├── deviceID
│   ├── tgtID
│   ├── serviceID
│   ├── timestamp
│   └── signature (encrypted with session key)
├── Validate TGT (cross-chaincode call to AS)
├── Check TGT expiration
├── Verify service exists and is active
├── Generate secure service key
├── Create service ticket with 30-minute validity
├── Store ticket on ledger
└── Return: Service Ticket + Service Key
```

**Service Ticket Structure**:
```json
{
  "ticketID": "ticket_xyz789",
  "deviceID": "device_001",
  "serviceID": "service001",
  "serviceKey": "base64_encoded_service_key",
  "issuedAt": 1672531200,
  "expiresAt": 1672533000,
  "status": "valid",
  "usageCount": 0,
  "maxUsageCount": 10
}
```

### Phase 4: Service Access

```
IoT Device -> ISV Chaincode
├── ValidateAccess(accessRequest)
│   ├── deviceID
│   ├── serviceID
│   ├── ticketID
│   ├── action (read/write/execute)
│   ├── timestamp
│   ├── ipAddress
│   └── signature (encrypted with service key)
├── Validate service ticket (cross-chaincode call to TGS)
├── Check ticket expiration and usage limits
├── Verify action permissions
├── Check for existing active session
├── Create/update session
├── Log access attempt
└── Return: Access granted + Session ID
```

**Session Structure**:
```json
{
  "sessionID": "session_def456",
  "deviceID": "device_001",
  "serviceID": "service001",
  "startTime": 1672531200,
  "lastActive": 1672531200,
  "status": "active"
}
```

## Security Features

### 1. Cryptographic Security
- **Random Generation**: All keys, nonces, and IDs use `crypto/rand`
- **Key Lengths**: 256-bit session keys and service keys
- **Public Key Cryptography**: Device authentication uses PKI
- **Signature Verification**: All requests must be signed

### 2. Temporal Security
- **Timestamp Validation**: Requests must be within 5-minute window
- **TGT Expiration**: 1-hour validity
- **Service Ticket Expiration**: 30-minute validity
- **Session Timeout**: 30 minutes of inactivity

### 3. Input Validation
- Length constraints on all input fields
- Pattern matching for IDs and identifiers
- Sanitization of user inputs
- Prevention of injection attacks

### 4. Rate Limiting
- Default: 60 requests per minute per device
- Automatic ban after 3 violations (5-minute ban)
- Configurable thresholds
- Per-device tracking

### 5. Audit Logging
- Comprehensive event logging
- Security event tracking
- Access attempt recording
- Immutable audit trail on blockchain

## Cross-Chaincode Communication

### AS ← → TGS
```
TGS validates TGT by querying AS:
TGS -> AS.GetTGT(tgtID)
AS returns TGT details or error
```

### TGS ← → ISV
```
ISV validates service ticket:
ISV -> TGS.ValidateServiceTicket(ticketID)
TGS increments usage count and returns ticket details
```

## Error Handling

### Common Error Scenarios
1. **Device Not Found**: Return 404 with descriptive message
2. **Expired Credentials**: Return 401 with expiration details
3. **Rate Limit Exceeded**: Return 429 with retry-after time
4. **Invalid Signature**: Return 403 with security alert
5. **Validation Failure**: Return 400 with field-specific errors

## Performance Considerations

### Optimization Strategies
1. **Caching**: Client-side caching of TGTs and tickets
2. **Batch Operations**: Group multiple requests when possible
3. **Session Reuse**: Minimize new session creation
4. **Efficient Queries**: Use indexed queries for lookups
5. **Cleanup**: Periodic cleanup of expired entries

## Scalability

### Horizontal Scaling
- Add more peers to each organization
- Load balancing across peers
- Sharding by device ID ranges (future enhancement)

### Vertical Scaling
- Increase peer resources
- Optimize chaincode execution
- Database tuning for world state

## Monitoring Points

### Key Metrics
1. **Authentication Rate**: Auths per second
2. **Ticket Issuance Rate**: Tickets per second
3. **Access Validation Latency**: Average response time
4. **Error Rate**: Failed requests percentage
5. **Active Sessions**: Current active session count
6. **Ledger Height**: Block height across peers

### Alert Triggers
- High error rate (> 5%)
- Unusual authentication patterns
- Repeated failed authentications
- Rate limit violations
- Service unavailability

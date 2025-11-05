# ISV Chaincode - IoT Service Validator

📍 **Location**: `chaincodes/isv-chaincode/`
🔗 **Parent**: [Chaincodes Overview](../README.md)
📚 **Related**: [AS Chaincode](../as-chaincode/README.md) | [TGS Chaincode](../tgs-chaincode/README.md)

## Overview

The ISV (IoT Service Validator) Chaincode is the **third and final tier** in the three-tier blockchain authentication framework. After a device receives a service ticket from TGS, it presents that ticket to ISV to actually access IoT services. Think of it as the "bouncer" - it validates service tickets, manages active sessions, and maintains comprehensive audit logs of all access attempts.

## Purpose

**Why does ISV Chaincode exist?**

The ISV provides the critical final layer of the authentication framework:

1. **Access Enforcement**: Actually grants or denies access to services based on valid tickets. Services delegate all authentication decisions to ISV.

2. **Session Management**: Tracks active device sessions, preventing duplicate logins and managing session timeouts (30 minutes).

3. **Audit Trail**: Records every access attempt (success/failure/denial) with detailed metadata (IP address, user agent, timestamp) for compliance and security monitoring.

4. **Service Isolation**: Services don't handle authentication - they just check with ISV. This separates business logic from security logic.

5. **Real-Time Monitoring**: Provides live view of active sessions, access patterns, and security events.

## Directory Structure

```
isv-chaincode/
├── isv-chaincode.go       # Main chaincode implementation
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
└── README.md             # This file
```

## Technologies Used

### Core Framework
- **Hyperledger Fabric Contract API (Go)** v1.2.1
  - *Why chosen*: Consistency with AS and TGS chaincodes
  - *What it does*: Provides transaction context for session and log management
  - *Benefit*: Built-in support for events and cross-chaincode calls

### Language
- **Go 1.21**
  - *Why chosen*: Same as other chaincodes for ecosystem consistency
  - *What it does*: Handles concurrent access validation for multiple devices

### Key Dependencies
- **fabric-protos-go** v0.3.3
  - *What it does*: Enables cross-chaincode ticket validation with TGS
  - *Why needed*: ISV calls TGS to validate service tickets

## Data Structures

### AccessLog
Immutable record of an access attempt.

```go
type AccessLog struct {
    LogID       string  // Unique log identifier
    DeviceID    string  // Device that attempted access
    ServiceID   string  // Service being accessed
    TicketID    string  // Service ticket presented
    Timestamp   int64   // Unix timestamp of access attempt
    Action      string  // "read", "write", "execute"
    Status      string  // "success", "failure", "denied"
    IPAddress   string  // Source IP address
    UserAgent   string  // Device user agent string
    Description string  // Human-readable description
}
```

**Storage Key**: `LOG_{logID}` (e.g., `"LOG_log_abc123"`)

**Why comprehensive logging?**
- **Compliance**: Regulatory requirements (GDPR, HIPAA, SOC2) mandate access logs
- **Security**: Detect suspicious patterns (brute force, privilege escalation)
- **Forensics**: Investigate security incidents with complete audit trail
- **Analytics**: Understand service usage patterns

**Why immutable?**
- Once written, logs never modified (blockchain guarantees)
- Prevents tampering or deletion of audit evidence
- Cryptographic proof of access events

### DeviceSession
Represents an active device-service session.

```go
type DeviceSession struct {
    SessionID  string  // Unique session identifier
    DeviceID   string  // Device in this session
    ServiceID  string  // Service being accessed
    StartTime  int64   // Session creation timestamp
    LastActive int64   // Last activity timestamp
    Status     string  // "active", "expired", "terminated"
}
```

**Storage Key**: `SESSION_{sessionID}` (e.g., `"SESSION_session_xyz789"`)

**Why track sessions?**
- **Prevent Duplicate Logins**: One device can't have multiple sessions for same service
- **Timeout Management**: Sessions expire after 30 minutes of inactivity
- **Resource Management**: Services can limit concurrent sessions
- **Session Hijacking Prevention**: Binds session to original device

**Session Lifecycle**:
1. **Created**: Device validates access → new session created (status="active")
2. **Active**: Each access updates LastActive timestamp
3. **Expired**: 30 minutes since LastActive → status="expired"
4. **Terminated**: Manual termination → status="terminated"

### AccessRequest
Request payload for access validation.

```go
type AccessRequest struct {
    DeviceID   string  // Device requesting access
    ServiceID  string  // Service to access
    TicketID   string  // Service ticket from TGS
    Action     string  // "read", "write", "execute"
    Timestamp  int64   // Request timestamp
    IPAddress  string  // Source IP address
    UserAgent  string  // Device user agent
    Signature  string  // Signature of entire request
}
```

**Why include IP and User Agent?**
- **Geofencing**: Can detect access from unusual locations
- **Device Fingerprinting**: Detect stolen credentials used from different device
- **Audit Trail**: Complete context for security investigations

**Action Types**:
- **read**: View data (GET operations)
- **write**: Modify data (POST/PUT/DELETE operations)
- **execute**: Trigger actions (device commands, analytics jobs)

### AccessResponse
Response payload after access validation.

```go
type AccessResponse struct {
    Granted    bool    // true if access granted, false otherwise
    SessionID  string  // Session identifier (empty if denied)
    Message    string  // Human-readable message
    ExpiresAt  int64   // Session expiration timestamp
}
```

## Functions

### Public Functions (Invokable)

#### 1. `InitLedger`
**Purpose**: Initialize the chaincode

```bash
peer chaincode invoke -C authchannel -n isv -c '{"Args":["InitLedger"]}'
```

**What it does**: Currently a no-op, reserved for future initialization.

**When to use**: Automatically called during deployment.

---

#### 2. `ValidateAccess`
**Purpose**: Validate a device's access request and create/update session

**Signature**:
```go
func ValidateAccess(ctx, accessRequestJSON string) (string, error)
```

**Parameters**:
- `accessRequestJSON`: JSON-encoded AccessRequest

**Example**:
```bash
ACCESS_REQUEST='{
  "deviceID": "device_001",
  "serviceID": "service001",
  "ticketID": "ticket_xyz789",
  "action": "read",
  "timestamp": 1672531200,
  "ipAddress": "192.168.1.100",
  "userAgent": "IoT-Device/1.0",
  "signature": "signed_access_request_base64"
}'

docker exec cli peer chaincode invoke \
  -C authchannel \
  -n isv \
  -c "{\"Args\":[\"ValidateAccess\",\"$ACCESS_REQUEST\"]}"
```

**Returns**: JSON-encoded AccessResponse

**Events Emitted**: `AccessGranted` with sessionID (only on success)

**Validation Steps**:
1. **Timestamp Validation**: Within ±5 minutes (prevents replay attacks)
2. **Signature Validation**: Verifies request authenticity
3. **Action Validation**: Must be "read", "write", or "execute"
4. **Ticket Validation**: Calls TGS to validate service ticket (cross-chaincode)
5. **Session Check**: Looks for existing active session

**Flow on Success**:
```
1. All validations pass
2. Check for existing session (deviceID + serviceID)
3. If existing session found and not expired:
   - Update LastActive timestamp
   - Log access (status="success", description="Using existing session")
   - Return response with existing sessionID
4. If no existing session:
   - Generate secure session ID
   - Create DeviceSession (StartTime=now, LastActive=now, Status="active")
   - Store session: Key=SESSION_{sessionID}
   - Log access (status="success", description="New session created")
   - Emit AccessGranted event
   - Return response with new sessionID
```

**Flow on Failure**:
```
1. Validation fails (e.g., invalid timestamp)
2. Log access (status="failure", description=reason)
3. Return AccessResponse{Granted: false, Message: reason}
```

**Session Reuse Logic**:
- **Benefit**: Reduces ticket validation overhead
- **Timeout**: 30 minutes since LastActive
- **Security**: Each access updates LastActive, extending session

**Production Enhancement** (Cross-Chaincode Ticket Validation):
```go
// Validate ticket with TGS
response := ctx.GetStub().InvokeChaincode("tgs", [][]byte{
    []byte("ValidateServiceTicket"),
    []byte(accessReq.TicketID),
}, "authchannel")

if response.Status != shim.OK {
    logAccess(ctx, ..., "denied", "Invalid ticket")
    return createAccessResponse(false, "", "Ticket validation failed", 0)
}
```

**When to use**: Every time a device wants to access a service. This is the primary function of ISV.

---

#### 3. `TerminateSession`
**Purpose**: Manually terminate an active session

**Signature**:
```go
func TerminateSession(ctx, sessionID string) error
```

**Parameters**:
- `sessionID`: Session identifier to terminate

**Example**:
```bash
peer chaincode invoke \
  -C authchannel \
  -n isv \
  -c '{"Args":["TerminateSession","session_xyz789"]}'
```

**Returns**: `nil` on success

**Events Emitted**: `SessionTerminated` with sessionID

**State Changes**:
- Updates session status to "terminated"
- Session record remains in ledger (audit trail)

**When to use**:
- **Logout**: Device explicitly logs out
- **Security**: Suspicious activity detected, forcibly terminate session
- **Admin**: Manual intervention required

**Difference from expiration**: Termination is immediate and manual, expiration is automatic after timeout.

---

#### 4. `GetAccessLogs`
**Purpose**: Retrieve access logs for a device

**Signature**:
```go
func GetAccessLogs(ctx, deviceID string) (string, error)
```

**Parameters**:
- `deviceID`: Device whose logs to retrieve

**Example**:
```bash
peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetAccessLogs","device_001"]}'
```

**Returns**: JSON-encoded array of AccessLog objects for the device

**Implementation**: Iterates all logs with prefix `LOG_`, filters by deviceID

**Performance Note**: O(n) operation where n = total logs. For production:
- Add composite key: `LOG_{deviceID}_{timestamp}_{logID}`
- Use `GetStateByPartialCompositeKey` for efficient filtering
- Implement pagination for large log sets

**When to use**:
- **Audit**: Review device activity history
- **Security**: Investigate suspicious behavior
- **Debugging**: Troubleshoot access issues

---

#### 5. `GetSession`
**Purpose**: Retrieve session information

**Signature**:
```go
func GetSession(ctx, sessionID string) (string, error)
```

**Example**:
```bash
peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetSession","session_xyz789"]}'
```

**Returns**: JSON-encoded DeviceSession object

**When to use**: Session validation, debugging, admin queries

---

#### 6. `GetActiveSessions`
**Purpose**: Retrieve all active sessions

**Signature**:
```go
func GetActiveSessions(ctx) (string, error)
```

**Example**:
```bash
peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetActiveSessions"]}'
```

**Returns**: JSON-encoded array of active DeviceSession objects

**Implementation**: Iterates all sessions, filters by status="active"

**When to use**:
- **Monitoring**: Real-time dashboard of active sessions
- **Capacity Planning**: How many concurrent sessions?
- **Security**: Detect anomalies (too many sessions, unexpected devices)

---

### Internal Helper Functions

#### `logAccess(...) error`
**Purpose**: Create an access log entry

**Parameters**:
- `ctx`: Transaction context
- `deviceID`, `serviceID`, `ticketID`, `action`: Request details
- `status`: "success", "failure", "denied"
- `ipAddress`, `userAgent`: Client metadata
- `description`: Human-readable message

**What it does**:
1. Generates secure log ID
2. Creates AccessLog object with all parameters
3. Stores log: Key = `LOG_{logID}`

**Why separate function?**
- **DRY**: Called from multiple places (success/failure paths)
- **Consistency**: Ensures all logs have same format
- **Audit Compliance**: Guarantees no access attempt goes unlogged

---

#### `findActiveSession(ctx, deviceID, serviceID) (string, error)`
**Purpose**: Find existing active session for device-service pair

**Algorithm**:
1. Iterate all sessions with prefix `SESSION_`
2. Filter by deviceID AND serviceID AND status="active"
3. Check if session hasn't timed out (currentTime - LastActive < 1800)
4. Return sessionID if found, error otherwise

**Why 30-minute timeout?**
- **Balance**: Long enough for typical operations, short enough for security
- **Consistency**: Same as service ticket validity
- **Industry Standard**: Common session timeout for APIs

---

#### `updateSession(ctx, sessionID) error`
**Purpose**: Update session's LastActive timestamp

**What it does**:
1. Retrieve session from ledger
2. Update LastActive to current timestamp
3. Store updated session

**Effect**: Extends session lifetime by 30 minutes from last activity

---

#### `createAccessResponse(granted, sessionID, message, expiresAt) (string, error)`
**Purpose**: Create JSON-encoded AccessResponse

**Why separate function?**
- **Consistency**: All responses have same format
- **Error Handling**: Centralized JSON marshaling with error handling
- **DRY**: Called from success and failure paths

---

#### `generateSecureSessionID() (string, error)`
**Current**: Placeholder returning `"session_" + random`

**Production**: Should generate 128-bit random ID using `crypto/rand`

**Why secure random?**
- **Session Hijacking Prevention**: Unpredictable session IDs
- **Brute Force Resistance**: 2^128 possible IDs

---

#### `generateSecureLogID() (string, error)`
**Current**: Placeholder

**Production**: Should use `crypto/rand` for unique log IDs

**Why unique IDs?**
- **Collision Avoidance**: Even with millions of logs
- **Audit Integrity**: Each log entry has unique identifier

## Security Considerations

### ✅ Implemented
1. **Comprehensive Logging**: Every access attempt logged (success/failure/denied)
2. **Timestamp Validation**: ±5 minute window prevents replay attacks
3. **Action Validation**: Only "read", "write", "execute" allowed
4. **Session Timeout**: 30-minute inactivity timeout
5. **Immutable Audit Trail**: Logs never deleted, only appended
6. **Event Emission**: Critical operations emit blockchain events

### ⚠️ Placeholder (Needs Production Implementation)
1. **Ticket Validation**: Currently basic checks, should use cross-chaincode call to TGS:
   ```go
   response := ctx.GetStub().InvokeChaincode("tgs", [][]byte{
       []byte("ValidateServiceTicket"),
       []byte(ticketID),
   }, "authchannel")
   ```

2. **Signature Verification**: Currently checks `len(signature) > 10`, should verify using service key from ticket

3. **Secure Random Generation**: Helper functions use placeholders, should use `crypto/rand`

### 🔒 Recommended Additions
1. **Geofencing**: Validate IPAddress against expected ranges
2. **Device Fingerprinting**: Detect ticket theft (compare UserAgent)
3. **Rate Limiting**: Prevent access flooding (use common/ratelimit.go)
4. **Anomaly Detection**: Detect unusual access patterns
5. **Log Archival**: Move old logs to off-chain storage (cold storage)
6. **Session Limits**: Max concurrent sessions per device

## Usage Flow

### Access Validation Flow (New Session)
```
1. Device obtains service ticket from TGS
2. Device creates AccessRequest with ticket, action, metadata
3. Device signs request with private key
4. Device → ValidateAccess(AccessRequest JSON)
5. ISV validates timestamp (±5 min)
6. ISV validates signature
7. ISV validates action ("read"/"write"/"execute")
8. ISV calls TGS to validate ticket (cross-chaincode)
9. ISV checks for existing active session
10. No existing session found
11. ISV generates secure session ID
12. ISV creates DeviceSession (StartTime=now, LastActive=now, Status="active")
13. Stores session: Key=SESSION_{sessionID}
14. Logs access (status="success", description="New session created")
15. Emits AccessGranted event
16. Returns AccessResponse{Granted: true, SessionID: sessionID, ExpiresAt: now+1800}
```

### Access Validation Flow (Existing Session)
```
1-8. Same as above
9. ISV finds existing active session (deviceID + serviceID)
10. Checks session timeout: (currentTime - LastActive < 1800)
11. Session still valid
12. Updates LastActive to currentTime
13. Logs access (status="success", description="Using existing session")
14. Returns AccessResponse{Granted: true, SessionID: existingSessionID, ExpiresAt: now+1800}

Note: No new session created, ticket validation skipped after first access
```

### Access Denial Flow
```
1-7. Same as above
8. Ticket validation fails (expired/revoked/invalid)
9. Logs access (status="denied", description="Invalid ticket")
10. Returns AccessResponse{Granted: false, Message: "Ticket validation failed"}

No session created, no event emitted
```

### Session Termination Flow
```
1. Admin/Device → TerminateSession(sessionID)
2. ISV retrieves session from ledger
3. Updates status to "terminated"
4. Stores updated session
5. Emits SessionTerminated event
6. Returns success

Future access requests will not find active session, require new ticket
```

## Integration with Other Chaincodes

### TGS Chaincode Integration
ISV validates service tickets via cross-chaincode call to TGS:

```go
// In ValidateAccess function (production implementation)
func validateTicketWithTGS(ctx contractapi.TransactionContextInterface, ticketID string) (ServiceTicket, error) {
    // Call TGS to validate ticket
    response := ctx.GetStub().InvokeChaincode(
        "tgs",
        [][]byte{[]byte("ValidateServiceTicket"), []byte(ticketID)},
        "authchannel",
    )

    if response.Status != shim.OK {
        return ServiceTicket{}, fmt.Errorf("ticket validation failed: %s", response.Message)
    }

    var ticket ServiceTicket
    err := json.Unmarshal(response.Payload, &ticket)
    if err != nil {
        return ServiceTicket{}, fmt.Errorf("failed to unmarshal ticket: %v", err)
    }

    // Ticket already validated by TGS (expiration, usage count)
    return ticket, nil
}
```

**Important**: TGS.ValidateServiceTicket increments usage counter, so call it only when actually granting access (not for queries).

### Service Integration
External IoT services delegate authentication to ISV:

```go
// In IoT Service (external application)
func handleDeviceRequest(deviceID, serviceID, ticketID, action string) (Response, error) {
    // Create access request
    accessReq := AccessRequest{
        DeviceID:  deviceID,
        ServiceID: serviceID,
        TicketID:  ticketID,
        Action:    action,
        Timestamp: time.Now().Unix(),
        IPAddress: request.RemoteAddr,
        UserAgent: request.UserAgent,
        Signature: signRequest(...),
    }

    // Call ISV chaincode
    response, err := invokeChaincode("isv", "ValidateAccess", accessReq)
    if err != nil {
        return Response{Error: "Access validation failed"}, err
    }

    var accessResp AccessResponse
    json.Unmarshal(response, &accessResp)

    if !accessResp.Granted {
        return Response{Error: accessResp.Message}, nil
    }

    // Access granted, proceed with service logic
    return processRequest(deviceID, action), nil
}
```

## Building and Deploying

### Build Chaincode
```bash
cd chaincodes/isv-chaincode
go mod tidy
go build -v .
```

### Package Chaincode
```bash
peer lifecycle chaincode package isv.tar.gz \
  --path chaincodes/isv-chaincode \
  --lang golang \
  --label isv_1.0
```

### Deploy (via automated script)
```bash
cd network
./scripts/deploy-chaincode.sh isv
```

**Deployment Steps**:
1. Package as `isv.tar.gz`
2. Install on all peers (Org1, Org2, Org3)
3. Approve for each organization
4. Commit to authchannel
5. Initialize ledger

### Verify Deployment
```bash
# Check committed
peer lifecycle chaincode querycommitted -C authchannel -n isv

# Test invocation
peer chaincode invoke \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetActiveSessions"]}'
```

## Testing

### Unit Tests
Located in `tests/unit/isv_chaincode_test.go`

```bash
cd tests
./run-tests.sh unit
```

**Test Cases**:
- Access validation (valid ticket, invalid ticket, expired timestamp)
- Session creation and reuse
- Session timeout handling
- Session termination
- Access log retrieval

### Integration Tests
Located in `tests/integration/authentication_flow_test.go`

```bash
./run-tests.sh integration
```

**Complete Flow**:
1. Register device (AS)
2. Authenticate device → get TGT (AS)
3. Request service ticket → get ticket (TGS)
4. Validate access → get session (ISV)
5. Access with existing session (ISV)
6. Wait 30 min → session expires
7. Access with new ticket → new session

### Manual Testing
```bash
# 1. Validate access (requires valid ticket from TGS)
docker exec cli peer chaincode invoke \
  -C authchannel -n isv \
  -c '{"Args":["ValidateAccess","{\"deviceID\":\"device_001\",\"serviceID\":\"service001\",\"ticketID\":\"ticket_xyz789\",\"action\":\"read\",\"timestamp\":1672531200,\"ipAddress\":\"192.168.1.100\",\"userAgent\":\"IoT-Device/1.0\",\"signature\":\"test_sig\"}"]}'

# 2. Get active sessions
docker exec cli peer chaincode query \
  -C authchannel -n isv \
  -c '{"Args":["GetActiveSessions"]}'

# 3. Get access logs for device
docker exec cli peer chaincode query \
  -C authchannel -n isv \
  -c '{"Args":["GetAccessLogs","device_001"]}'

# 4. Terminate session
docker exec cli peer chaincode invoke \
  -C authchannel -n isv \
  -c '{"Args":["TerminateSession","session_abc123"]}'
```

## Troubleshooting

### Common Issues

**Error**: `"timestamp is invalid or too old"`
- **Cause**: Request timestamp > 5 minutes old/future
- **Solution**: Synchronize device clock with NTP, use current Unix timestamp

**Error**: `"invalid signature"`
- **Cause**: Signature verification failed
- **Solution**: Check signing algorithm, ensure using correct private key

**Error**: `"invalid action"`
- **Cause**: Action not in ["read", "write", "execute"]
- **Solution**: Use valid action type

**Error**: `"invalid ticket"`
- **Cause**: Ticket validation with TGS failed
- **Solution**: Check ticket validity, expiration, usage count; request new ticket if needed

**Error**: `"session not found"`
- **Cause**: Session expired (30 min timeout) or never created
- **Solution**: Request new service ticket and validate access again

### Debugging

**Enable chaincode logs**:
```bash
docker logs -f peer0.org3.example.com
```

**Check active sessions**:
```bash
peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetActiveSessions"]}'
```

**Inspect access logs**:
```bash
peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetAccessLogs","device_001"]}'
```

**Check specific session**:
```bash
peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetSession","session_abc123"]}'
```

## Performance Considerations

### State Management
- **Sessions**: Prefixed with `SESSION_` for efficient range queries
- **Logs**: Prefixed with `LOG_` for namespacing
- **Session Lookup**: O(n) scan for active sessions (optimize with composite keys)

### Scalability
- **Concurrent Access**: Fabric handles concurrent validation via MVCC
- **Log Growth**: Logs grow indefinitely (implement archival for production)
- **Session Cleanup**: No automatic cleanup of expired sessions (add garbage collection)

### Optimization Tips
1. **Composite Keys**: Use `{deviceID}~{serviceID}~{sessionID}` for O(1) session lookup
2. **Log Pagination**: Add pagination to GetAccessLogs for large datasets
3. **Log Archival**: Move logs older than 30 days to off-chain storage
4. **Session Cache**: Client-side caching of sessionID reduces validation calls
5. **Batch Validation**: Validate multiple access requests in single transaction

## Learn More

- **[AS Chaincode](../as-chaincode/README.md)** - Device authentication
- **[TGS Chaincode](../tgs-chaincode/README.md)** - Service ticket issuance
- **[Common Utilities](../common/README.md)** - Shared validation functions
- **[HOW IT WORKS](../../HOW_IT_WORKS.md)** - Complete authentication flow
- **[API Reference](../../docs/api/chaincode-api.md)** - Full API documentation

## Navigation

📍 **Path**: [Main README](../../README.md) → [Chaincodes](../README.md) → **ISV Chaincode**

🔗 **Quick Links**:
- [← TGS Chaincode](../tgs-chaincode/README.md)
- [Common Utilities](../common/README.md)
- [Developer Guide](../../DEVELOPER_GUIDE.md)
- [Troubleshooting](../../docs/troubleshooting/common-issues.md)

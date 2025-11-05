# TGS Chaincode - Ticket Granting Server

📍 **Location**: `chaincodes/tgs-chaincode/`
🔗 **Parent**: [Chaincodes Overview](../README.md)
📚 **Related**: [AS Chaincode](../as-chaincode/README.md) | [ISV Chaincode](../isv-chaincode/README.md)

## Overview

The TGS (Ticket Granting Server) Chaincode is the **second tier** in the three-tier blockchain authentication framework. After a device authenticates with the AS and receives a TGT, it uses the TGS to request service tickets for specific IoT services. Think of it as the "ticket office" - devices present their TGT and request access to specific services, receiving time-limited service tickets in return.

## Purpose

**Why does TGS Chaincode exist?**

In Kerberos-inspired authentication, the Ticket Granting Server provides an important security layer:

1. **Separation of Concerns**: Separates initial authentication (AS) from service authorization (TGS). This allows revoking service access without re-authenticating devices.

2. **Fine-Grained Access Control**: Different services can have different access requirements (roles, permissions). TGS enforces these policies before issuing service tickets.

3. **Service Discovery**: Maintains a registry of available IoT services (data access, device control, analytics, etc.).

4. **Limited-Use Tickets**: Service tickets have shorter validity (30 minutes) and usage limits (10 uses) compared to TGTs, reducing security exposure.

5. **Decoupling**: Services don't need to know about device authentication - they only validate service tickets via ISV.

## Directory Structure

```
tgs-chaincode/
├── tgs-chaincode.go       # Main chaincode implementation
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
└── README.md             # This file
```

## Technologies Used

### Core Framework
- **Hyperledger Fabric Contract API (Go)** v1.2.1
  - *Why chosen*: Same as AS chaincode for consistency
  - *What it does*: Contract-based chaincode development with automatic transaction context handling
  - *Benefit*: Simplified cross-chaincode invocation support

### Language
- **Go 1.21**
  - *Why chosen*: Required for Fabric chaincodes, excellent standard library
  - *What it does*: Provides concurrent ticket issuance for multiple devices simultaneously

### Key Dependencies
- **fabric-protos-go** v0.3.3
  - *What it does*: Protocol buffers for Fabric communication
  - *Why needed*: Cross-chaincode calls to AS for TGT verification

## Data Structures

### ServiceTicket
Represents authorization to access a specific service.

```go
type ServiceTicket struct {
    TicketID        string  // Unique ticket identifier
    DeviceID        string  // Device authorized by this ticket
    ServiceID       string  // Service this ticket grants access to
    ServiceKey      string  // Encrypted session key for service communication
    IssuedAt        int64   // Unix timestamp of issuance
    ExpiresAt       int64   // Unix timestamp of expiration (IssuedAt + 1800)
    Status          string  // "valid", "expired", "revoked", "used"
    UsageCount      int     // How many times this ticket has been used
    MaxUsageCount   int     // Maximum allowed uses (0 = unlimited)
}
```

**Storage Key**: `TICKET_{ticketID}` (e.g., `"TICKET_ticket_xyz789"`)

**Why 30 minute validity?**
- **Shorter than TGT (1 hour)**: Limits exposure if ticket is compromised
- **Usage-based expiration**: Even if time limit not reached, tickets expire after 10 uses
- **Balance**: Long enough for typical IoT operations, short enough for security

**Why usage limits?**
- **Prevents ticket sharing**: Each ticket is bound to specific device and limited uses
- **Audit trail**: Each use increments counter, providing usage analytics
- **Revocation granularity**: Can revoke specific service access without affecting TGT

### Service
Represents an available IoT service.

```go
type Service struct {
    ServiceID    string  // Unique service identifier
    ServiceName  string  // Human-readable name
    Description  string  // Service description
    IsActive     bool    // Whether service is currently available
    RequiredRole string  // Required role to access ("user", "admin", etc.)
}
```

**Storage Key**: `SERVICE_{serviceID}` (e.g., `"SERVICE_service001"`)

**Why role-based access?**
- **Extensibility**: Future enhancement can check device roles in cross-chaincode call to AS
- **Policy Enforcement**: Different services can require different privilege levels
- **Compliance**: Meets regulatory requirements for access control (GDPR, HIPAA)

### TicketRequest
Request payload for service ticket issuance.

```go
type TicketRequest struct {
    DeviceID   string  // Device requesting the ticket
    TgtID      string  // TGT issued by AS (proves device authenticated)
    ServiceID  string  // Service the device wants to access
    Timestamp  int64   // Request timestamp (prevents replay)
    Signature  string  // Signature of (deviceID + tgtID + serviceID + timestamp)
}
```

**Why include TgtID?**
- **Proof of Authentication**: Only devices with valid TGT can request service tickets
- **Session Binding**: Links service ticket to original authentication session
- **Cross-Chaincode Verification**: TGS calls AS chaincode to verify TGT validity

## Functions

### Public Functions (Invokable)

#### 1. `InitLedger`
**Purpose**: Initialize the ledger with default services

```bash
peer chaincode invoke -C authchannel -n tgs -c '{"Args":["InitLedger"]}'
```

**What it does**: Pre-registers two default services:
- `service001`: "IoT Data Access" (role: user)
- `service002`: "Device Control" (role: admin)

**When to use**: Automatically called during chaincode deployment.

**State Changes**:
- Creates 2 Service records with keys `SERVICE_service001` and `SERVICE_service002`
- Both services set to `IsActive: true`

---

#### 2. `RegisterService`
**Purpose**: Register a new IoT service

**Signature**:
```go
func RegisterService(ctx, serviceID string, serviceName string, description string, requiredRole string) error
```

**Parameters**:
- `serviceID`: Unique service identifier (3-64 chars)
- `serviceName`: Human-readable name (e.g., "Analytics Dashboard")
- `description`: Service description (e.g., "Real-time analytics and reporting")
- `requiredRole`: Required access level ("user", "admin", "operator")

**Example**:
```bash
docker exec cli peer chaincode invoke \
  -C authchannel \
  -n tgs \
  -c '{"Args":["RegisterService","service003","Analytics Dashboard","Real-time analytics","admin"]}'
```

**Returns**: `nil` on success, `error` if service already exists

**Events Emitted**: `ServiceRegistered` with serviceID payload

**Validation Performed**:
1. Service doesn't already exist
2. ServiceID length: 3-64 characters

**When to use**: When deploying new IoT services that require authentication, register them here so devices can request access.

---

#### 3. `IssueServiceTicket`
**Purpose**: Issue a service ticket to an authenticated device

**Signature**:
```go
func IssueServiceTicket(ctx, ticketRequestJSON string) (string, error)
```

**Parameters**:
- `ticketRequestJSON`: JSON-encoded TicketRequest

**Example**:
```bash
TICKET_REQUEST='{
  "deviceID": "device_001",
  "tgtID": "tgt_abc123",
  "serviceID": "service001",
  "timestamp": 1672531200,
  "signature": "signed_request_base64"
}'

docker exec cli peer chaincode invoke \
  -C authchannel \
  -n tgs \
  -c "{\"Args\":[\"IssueServiceTicket\",\"$TICKET_REQUEST\"]}"
```

**Returns**: JSON-encoded ServiceTicket

**Events Emitted**: `ServiceTicketIssued` with ticketID payload

**Validation Performed**:
1. Timestamp within ±5 minutes (prevents replay attacks)
2. TGT ID format validation (production: cross-chaincode call to AS)
3. Service exists and is active
4. Signature validation (production: verify using session key from TGT)

**State Changes**:
- Creates ServiceTicket with 30-minute validity
- Sets UsageCount to 0, MaxUsageCount to 10
- Stores ticket: Key = `TICKET_{ticketID}`

**Security Features**:
- Service key generated using `crypto/rand` (see generateSecureServiceKey)
- Ticket ID generated with secure random (see generateSecureTicketID)
- Shorter validity than TGT (30 min vs 60 min)
- Usage limits prevent ticket reuse

**Cross-Chaincode Integration** (Production):
```go
// Verify TGT with AS chaincode
response := ctx.GetStub().InvokeChaincode("as", [][]byte{
    []byte("GetTGT"),
    []byte(ticketReq.TgtID),
}, "authchannel")

var tgt TGT
json.Unmarshal(response.Payload, &tgt)

if tgt.Status != "valid" || tgt.ExpiresAt < getCurrentTimestamp() {
    return "", fmt.Errorf("invalid or expired TGT")
}
```

**When to use**: After device authenticates with AS and receives TGT, it calls this function to get access to specific services.

---

#### 4. `ValidateServiceTicket`
**Purpose**: Validate a service ticket and increment usage counter

**Signature**:
```go
func ValidateServiceTicket(ctx, ticketID string) (string, error)
```

**Parameters**:
- `ticketID`: Ticket identifier to validate

**Example**:
```bash
peer chaincode invoke \
  -C authchannel \
  -n tgs \
  -c '{"Args":["ValidateServiceTicket","ticket_xyz789"]}'
```

**Returns**: JSON-encoded ServiceTicket with updated usage count

**Validation Performed**:
1. Ticket exists in ledger
2. Ticket status is "valid" (not expired/revoked/used)
3. Ticket hasn't expired (currentTime ≤ ExpiresAt)
4. Usage count < MaxUsageCount (if MaxUsageCount > 0)

**State Changes**:
- Increments UsageCount by 1
- If usage limit reached: sets status to "used"
- If expired: sets status to "expired"

**When to use**: Called by ISV chaincode during access validation. Can also be called directly for testing.

**Important**: This function mutates state (increments counter), so it should only be called when actually granting access.

---

#### 5. `RevokeServiceTicket`
**Purpose**: Revoke a service ticket (emergency access removal)

**Signature**:
```go
func RevokeServiceTicket(ctx, ticketID string) error
```

**Example**:
```bash
peer chaincode invoke \
  -C authchannel \
  -n tgs \
  -c '{"Args":["RevokeServiceTicket","ticket_xyz789"]}'
```

**Returns**: `nil` on success

**Events Emitted**: `ServiceTicketRevoked` with ticketID payload

**State Changes**:
- Updates ticket status to "revoked"
- Ticket record remains in ledger (audit trail)

**When to use**:
- Security incident detected
- Device behavior suspicious
- Manual admin intervention required

**Difference from expiration**: Revocation is immediate and manual, expiration is automatic and time-based.

---

#### 6. `GetService`
**Purpose**: Retrieve service information

**Signature**:
```go
func GetService(ctx, serviceID string) (string, error)
```

**Example**:
```bash
peer chaincode query \
  -C authchannel \
  -n tgs \
  -c '{"Args":["GetService","service001"]}'
```

**Returns**: JSON-encoded Service object

**When to use**: Service discovery, client applications listing available services

---

#### 7. `GetAllServices`
**Purpose**: Retrieve all registered services

**Signature**:
```go
func GetAllServices(ctx) (string, error)
```

**Example**:
```bash
peer chaincode query \
  -C authchannel \
  -n tgs \
  -c '{"Args":["GetAllServices"]}'
```

**Returns**: JSON-encoded array of Service objects

**Implementation Detail**: Uses `GetStateByRange("SERVICE_", "SERVICE_~")` for efficient prefix scanning

**When to use**: Service directory, admin dashboards, device configuration

---

### Internal Helper Functions

#### `getCurrentTimestamp() int64`
**Current**: Placeholder returning `1672531200`

**Production**: Should use transaction timestamp (see AS chaincode README)

---

#### `generateSecureServiceKey() (string, error)`
**Current**: Placeholder

**Production**: Should generate 256-bit random key using `crypto/rand`

**Why service-specific key?**
- Each service ticket has unique encryption key
- If one ticket compromised, others remain secure
- Enables per-service encryption/decryption

---

#### `generateSecureTicketID() (string, error)`
**Current**: Placeholder

**Production**: Should generate 128-bit random ID using `crypto/rand`

**Why unique IDs?**
- Prevents ticket ID collision
- Ensures unpredictable ticket identifiers
- Required for secure ticket management

## Security Considerations

### ✅ Implemented
1. **Timestamp Validation**: ±5 minute window prevents replay attacks
2. **Service Status Check**: Only active services can issue tickets
3. **Usage Limits**: Tickets expire after 10 uses
4. **Event Logging**: All ticket operations emit blockchain events
5. **Status Tracking**: Tickets have explicit states (valid/expired/revoked/used)

### ⚠️ Placeholder (Needs Production Implementation)
1. **TGT Verification**: Currently basic validation, should use cross-chaincode call to AS:
   ```go
   response := ctx.GetStub().InvokeChaincode("as", [][]byte{
       []byte("GetTGT"),
       []byte(tgtID),
   }, "authchannel")
   ```

2. **Signature Verification**: Currently checks `len(signature) > 10`, should verify using session key from TGT

3. **Secure Random Generation**: Helper functions use placeholders, should use `crypto/rand`

### 🔒 Recommended Additions
1. **Role-Based Access Control**: Check device role against `Service.RequiredRole`
2. **Rate Limiting**: Prevent ticket request flooding (use common/ratelimit.go)
3. **Audit Logging**: Use common/audit.go for comprehensive logging
4. **Ticket Cleanup**: Archive expired tickets to secondary storage
5. **Service Policies**: Add per-service policies (max tickets per device, time-of-day restrictions)

## Usage Flow

### Service Registration Flow
```
1. Admin → RegisterService(serviceID, name, description, role)
2. TGS validates inputs
3. Creates Service record with IsActive=true
4. Stores: Key=SERVICE_{serviceID}
5. Emits ServiceRegistered event
6. Returns success
```

### Service Ticket Issuance Flow
```
1. Device authenticates with AS → receives TGT
2. Device generates TicketRequest (deviceID, tgtID, serviceID, timestamp)
3. Device signs request with private key
4. Device → IssueServiceTicket(TicketRequest JSON)
5. TGS validates timestamp (±5 min window)
6. TGS verifies TGT with AS (cross-chaincode call)
7. TGS checks service exists and is active
8. TGS generates secure service key (crypto/rand)
9. TGS creates ServiceTicket (30 min validity, 10 use limit)
10. Stores: Key=TICKET_{ticketID}
11. Emits ServiceTicketIssued event
12. Returns ServiceTicket JSON
```

### Ticket Validation Flow
```
1. ISV/Client → ValidateServiceTicket(ticketID)
2. TGS retrieves ticket from ledger
3. Checks status == "valid"
4. Checks currentTime ≤ ExpiresAt
5. Checks UsageCount < MaxUsageCount
6. If all valid:
   - Increments UsageCount
   - If UsageCount == MaxUsageCount: status = "used"
   - Stores updated ticket
   - Returns ticket JSON
7. If invalid:
   - Updates status (expired/used)
   - Returns error
```

### Ticket Revocation Flow
```
1. Admin → RevokeServiceTicket(ticketID)
2. TGS retrieves ticket
3. Updates status = "revoked"
4. Stores updated ticket
5. Emits ServiceTicketRevoked event
6. Returns success

Future validations will fail due to status != "valid"
```

## Integration with Other Chaincodes

### AS Chaincode Integration
TGS verifies TGTs issued by AS via cross-chaincode invocation:

```go
// In IssueServiceTicket function (production implementation)
func verifyTGT(ctx contractapi.TransactionContextInterface, tgtID string) error {
    // Call AS chaincode to get TGT
    response := ctx.GetStub().InvokeChaincode(
        "as",
        [][]byte{[]byte("GetTGT"), []byte(tgtID)},
        "authchannel",
    )

    if response.Status != shim.OK {
        return fmt.Errorf("TGT verification failed: %s", response.Message)
    }

    var tgt TGT
    err := json.Unmarshal(response.Payload, &tgt)
    if err != nil {
        return fmt.Errorf("failed to unmarshal TGT: %v", err)
    }

    // Validate TGT
    currentTime := getCurrentTimestamp()
    if tgt.Status != "valid" {
        return fmt.Errorf("TGT status is %s, not valid", tgt.Status)
    }
    if currentTime > tgt.ExpiresAt {
        return fmt.Errorf("TGT expired at %d, current time %d", tgt.ExpiresAt, currentTime)
    }

    return nil
}
```

### ISV Chaincode Integration
ISV validates service tickets issued by TGS:

```go
// In ISV ValidateAccess function
func validateTicket(ctx, ticketID string) (ServiceTicket, error) {
    // Call TGS chaincode to validate ticket
    response := ctx.GetStub().InvokeChaincode(
        "tgs",
        [][]byte{[]byte("ValidateServiceTicket"), []byte(ticketID)},
        "authchannel",
    )

    if response.Status != shim.OK {
        return ServiceTicket{}, fmt.Errorf("ticket validation failed")
    }

    var ticket ServiceTicket
    json.Unmarshal(response.Payload, &ticket)
    return ticket, nil
}
```

## Building and Deploying

### Build Chaincode
```bash
cd chaincodes/tgs-chaincode
go mod tidy
go build -v .
```

### Package Chaincode
```bash
peer lifecycle chaincode package tgs.tar.gz \
  --path chaincodes/tgs-chaincode \
  --lang golang \
  --label tgs_1.0
```

### Deploy (via automated script)
```bash
cd network
./scripts/deploy-chaincode.sh tgs
```

**Deployment Steps**:
1. Package as `tgs.tar.gz`
2. Install on peers in Org1, Org2, Org3
3. Approve for each organization
4. Commit to authchannel
5. Initialize ledger (creates service001, service002)

### Verify Deployment
```bash
# Check committed
peer lifecycle chaincode querycommitted -C authchannel -n tgs

# Test query
peer chaincode query \
  -C authchannel \
  -n tgs \
  -c '{"Args":["GetAllServices"]}'
```

## Testing

### Unit Tests
Located in `tests/unit/tgs_chaincode_test.go`

```bash
cd tests
./run-tests.sh unit
```

**Test Cases**:
- Service registration (success, duplicate)
- Service ticket issuance (valid TGT, invalid TGT, inactive service)
- Ticket validation (valid, expired, usage limit exceeded)
- Ticket revocation

### Integration Tests
Located in `tests/integration/authentication_flow_test.go`

```bash
./run-tests.sh integration
```

**Test Scenarios**:
- AS authentication → TGS ticket issuance → ISV access validation
- Ticket expiration handling (time-based and usage-based)
- Service ticket revocation mid-session

### Manual Testing
```bash
# 1. Register service
docker exec cli peer chaincode invoke \
  -C authchannel -n tgs \
  -c '{"Args":["RegisterService","test_service","Test Service","Testing only","user"]}'

# 2. Request service ticket (requires valid TGT from AS)
docker exec cli peer chaincode invoke \
  -C authchannel -n tgs \
  -c '{"Args":["IssueServiceTicket","{\"deviceID\":\"device_001\",\"tgtID\":\"tgt_abc123\",\"serviceID\":\"test_service\",\"timestamp\":1672531200,\"signature\":\"test_sig\"}"]}'

# 3. Validate ticket
docker exec cli peer chaincode invoke \
  -C authchannel -n tgs \
  -c '{"Args":["ValidateServiceTicket","ticket_xyz789"]}'

# 4. Revoke ticket
docker exec cli peer chaincode invoke \
  -C authchannel -n tgs \
  -c '{"Args":["RevokeServiceTicket","ticket_xyz789"]}'
```

## Troubleshooting

### Common Issues

**Error**: `"service already exists"`
- **Cause**: Registering service with duplicate serviceID
- **Solution**: Use unique serviceID or query existing services first

**Error**: `"service not found"`
- **Cause**: Requesting ticket for unregistered service
- **Solution**: Register service using RegisterService

**Error**: `"service is not active"`
- **Cause**: Service exists but IsActive=false
- **Solution**: Check service configuration, update IsActive flag if needed

**Error**: `"invalid TGT ID"`
- **Cause**: TGT ID format invalid or TGT doesn't exist in AS
- **Solution**: Ensure device authenticated with AS first, check TGT ID format

**Error**: `"timestamp is invalid or too old"`
- **Cause**: Request timestamp > 5 minutes old/future
- **Solution**: Synchronize device clock, use current Unix timestamp

**Error**: `"ticket has expired"`
- **Cause**: Using ticket after 30-minute validity window
- **Solution**: Request new service ticket from TGS

**Error**: `"ticket usage limit exceeded"`
- **Cause**: Ticket used > 10 times
- **Solution**: Request new service ticket (max 10 uses per ticket)

### Debugging

**Enable chaincode logs**:
```bash
docker logs -f peer0.org2.example.com
```

**Check service registry**:
```bash
peer chaincode query \
  -C authchannel \
  -n tgs \
  -c '{"Args":["GetAllServices"]}'
```

**Inspect ticket state**:
```bash
peer chaincode query \
  -C authchannel \
  -n tgs \
  -c '{"Args":["ValidateServiceTicket","ticket_xyz789"]}'
```

## Performance Considerations

### State Management
- **Services**: Prefixed with `SERVICE_` for efficient range queries
- **Tickets**: Prefixed with `TICKET_` for namespacing
- **Usage Counters**: In-place updates (no history tracking for performance)

### Scalability
- **GetAllServices**: Uses range query, O(n) operation
- **Concurrent Ticket Issuance**: Fabric handles concurrent requests via MVCC
- **Ticket Archival**: Consider moving expired tickets to separate chaincode or off-chain storage

### Optimization Tips
1. **Service Caching**: Client-side caching of service list (rarely changes)
2. **Batch Ticket Requests**: Issue tickets for multiple services in one transaction
3. **Ticket Pooling**: Pre-issue tickets for frequently accessed services
4. **Pagination**: Add pagination to GetAllServices for >100 services

## Learn More

- **[AS Chaincode](../as-chaincode/README.md)** - Previous step: Device authentication
- **[ISV Chaincode](../isv-chaincode/README.md)** - Next step: Access validation
- **[Common Utilities](../common/README.md)** - Shared functions for ticket validation
- **[HOW IT WORKS](../../HOW_IT_WORKS.md)** - System architecture and flow diagrams
- **[API Reference](../../docs/api/chaincode-api.md)** - Complete API documentation

## Navigation

📍 **Path**: [Main README](../../README.md) → [Chaincodes](../README.md) → **TGS Chaincode** → [ISV Chaincode](../isv-chaincode/README.md)

🔗 **Quick Links**:
- [← AS Chaincode](../as-chaincode/README.md)
- [ISV Chaincode →](../isv-chaincode/README.md)
- [Developer Guide](../../DEVELOPER_GUIDE.md)
- [Production Deployment](../../docs/deployment/PRODUCTION_DEPLOYMENT.md)

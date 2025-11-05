# AS Chaincode - Authentication Server

📍 **Location**: `chaincodes/as-chaincode/`
🔗 **Parent**: [Chaincodes Overview](../README.md)
📚 **Related**: [API Reference](../../docs/api/chaincode-api.md) | [Developer Guide](../../DEVELOPER_GUIDE.md)

## Overview

The AS (Authentication Server) Chaincode is the **first tier** in the three-tier blockchain authentication framework. It handles initial device registration and authentication, issuing Ticket Granting Tickets (TGTs) to authenticated devices. Think of it as the "gatekeeper" - it verifies device identity using PKI (Public Key Infrastructure) and grants initial access credentials.

## Purpose

**Why does AS Chaincode exist?**

In a traditional Kerberos system, the Authentication Server (AS) is responsible for verifying user credentials and issuing TGTs. In our blockchain IoT adaptation:

1. **Decentralization**: Unlike centralized authentication servers, this chaincode runs on multiple peers across organizations, providing fault tolerance and eliminating single points of failure.

2. **Immutability**: All device registrations and authentication events are recorded on the immutable blockchain ledger, creating a permanent audit trail.

3. **PKI Authentication**: Uses public-key cryptography instead of passwords, which is more suitable for IoT devices.

4. **TGT Issuance**: Issues session credentials (TGTs) that devices use to request service tickets from the TGS chaincode.

## Directory Structure

```
as-chaincode/
├── as-chaincode.go        # Main chaincode implementation
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
└── README.md             # This file
```

## Technologies Used

### Core Framework
- **Hyperledger Fabric Contract API (Go)** v1.2.1
  - *Why chosen*: Provides high-level abstraction over Fabric's shim API
  - *What it does*: Simplifies chaincode development with contract-based programming model
  - *Alternative considered*: fabric-chaincode-go (low-level shim) - rejected for complexity

### Language
- **Go 1.21**
  - *Why chosen*: Native language for Fabric chaincodes, excellent concurrency support
  - *What it does*: Compiles to efficient binary, provides strong typing and performance
  - *Alternative considered*: Node.js - rejected for lower performance in cryptographic operations

### Key Dependencies
- **fabric-protos-go** v0.3.3
  - *What it does*: Provides Protobuf definitions for Fabric messages
  - *Why needed*: Required for chaincode-to-chaincode communication and transaction context

## Data Structures

### Device
Represents a registered IoT device.

```go
type Device struct {
    DeviceID         string  // Unique identifier (3-64 chars)
    PublicKey        string  // PEM-encoded RSA/ECDSA public key (100-4096 chars)
    Status           string  // "active", "suspended", "revoked"
    RegistrationTime int64   // Unix timestamp of registration
    LastAuthTime     int64   // Unix timestamp of last authentication
    Metadata         string  // Additional device info (model, version, etc.)
}
```

**Storage Key**: `deviceID` (e.g., `"device_001"`)

**Why this structure?**
- `PublicKey`: Enables signature verification without storing private keys
- `Status`: Allows revocation without deleting device records (audit trail preservation)
- `Metadata`: Extensible field for device classification and management

### TGT (Ticket Granting Ticket)
Represents authentication credentials issued to devices.

```go
type TGT struct {
    TgtID      string  // Unique TGT identifier
    DeviceID   string  // Device that owns this TGT
    SessionKey string  // Encrypted session key for TGS communication
    IssuedAt   int64   // Unix timestamp of issuance
    ExpiresAt  int64   // Unix timestamp of expiration (IssuedAt + 3600)
    Status     string  // "valid", "expired", "revoked"
}
```

**Storage Key**: `TGT_{tgtID}` (e.g., `"TGT_tgt_abc123"`)

**Why 1 hour validity?**
- Balance between security (shorter = less exposure) and usability (longer = fewer re-authentications)
- Industry standard from Kerberos (typically 8-10 hours, but IoT devices require tighter security)

### AuthRequest
Request payload for device authentication.

```go
type AuthRequest struct {
    DeviceID  string  // Device requesting authentication
    Nonce     string  // Random value to prevent replay attacks
    Timestamp int64   // Request timestamp (must be within 5 minutes)
    Signature string  // Signature of (deviceID + nonce + timestamp) using device's private key
}
```

**Why nonce + timestamp?**
- `Nonce`: Ensures uniqueness even if timestamp collision occurs
- `Timestamp`: Prevents replay attacks (old requests rejected if > 5 minutes old)

### AuthResponse
Response payload after successful authentication.

```go
type AuthResponse struct {
    TgtID      string  // Issued TGT identifier
    SessionKey string  // Session key for TGS communication
    ExpiresAt  int64   // When the TGT expires
    Message    string  // Human-readable message
}
```

## Functions

### Public Functions (Invokable)

#### 1. `InitLedger`
**Purpose**: Initialize the chaincode (called once during instantiation)

```bash
peer chaincode invoke -C authchannel -n as -c '{"Args":["InitLedger"]}'
```

**What it does**: Currently a no-op, but reserved for future initialization logic (e.g., pre-registering admin devices).

**When to use**: Automatically called during chaincode deployment. Don't call manually.

---

#### 2. `RegisterDevice`
**Purpose**: Register a new IoT device with the authentication server

**Signature**:
```go
func RegisterDevice(ctx, deviceID string, publicKey string, metadata string) error
```

**Parameters**:
- `deviceID`: Unique device identifier (3-64 alphanumeric chars)
- `publicKey`: PEM-encoded public key (100-4096 chars)
- `metadata`: Device information (e.g., `"IoT Sensor v1.0"`)

**Example**:
```bash
docker exec cli peer chaincode invoke \
  -C authchannel \
  -n as \
  -c '{"Args":["RegisterDevice","device_001","-----BEGIN PUBLIC KEY-----\nMIIBIjANBg...","IoT Sensor v1.0"]}'
```

**Returns**: `nil` on success, `error` if device already exists or validation fails

**Events Emitted**: `DeviceRegistered` with deviceID payload

**Validation Performed**:
1. Device doesn't already exist
2. DeviceID length: 3-64 characters
3. PublicKey length: 100-4096 characters

**State Changes**:
- Stores device with status "active"
- Sets RegistrationTime to current timestamp
- Initializes LastAuthTime to 0

**When to use**: Before a device can authenticate, it must be registered. Typically done during device provisioning.

---

#### 3. `Authenticate`
**Purpose**: Authenticate a device and issue a TGT

**Signature**:
```go
func Authenticate(ctx, authRequestJSON string) (string, error)
```

**Parameters**:
- `authRequestJSON`: JSON-encoded AuthRequest

**Example**:
```bash
AUTH_REQUEST='{
  "deviceID": "device_001",
  "nonce": "secure_random_nonce_base64",
  "timestamp": 1672531200,
  "signature": "device_signature_base64"
}'

docker exec cli peer chaincode invoke \
  -C authchannel \
  -n as \
  -c "{\"Args\":[\"Authenticate\",\"$AUTH_REQUEST\"]}"
```

**Returns**: JSON-encoded AuthResponse with TGT details

**Events Emitted**: `DeviceAuthenticated` with deviceID payload

**Validation Performed**:
1. Device exists in ledger
2. Device status is "active" (not suspended/revoked)
3. Timestamp is within 5 minutes of current time (±300 seconds)
4. Signature is valid (currently placeholder - production should verify using device's public key)

**State Changes**:
- Creates new TGT with 1-hour validity
- Updates device's LastAuthTime
- Stores TGT in ledger with key `TGT_{tgtID}`

**Security Features**:
- Session key generated using crypto/rand (see `generateSecureSessionKey`)
- TGT ID generated with secure random (see `generateSecureTgtID`)
- Timestamp validation prevents replay attacks

**When to use**: Every time a device needs to access services. The TGT obtained here is used to request service tickets from TGS.

---

#### 4. `GetDevice`
**Purpose**: Retrieve device information

**Signature**:
```go
func GetDevice(ctx, deviceID string) (string, error)
```

**Example**:
```bash
peer chaincode query -C authchannel -n as -c '{"Args":["GetDevice","device_001"]}'
```

**Returns**: JSON-encoded Device object

**When to use**: Administrative queries, debugging, device status checks

---

#### 5. `GetTGT`
**Purpose**: Retrieve TGT information

**Signature**:
```go
func GetTGT(ctx, tgtID string) (string, error)
```

**Example**:
```bash
peer chaincode query -C authchannel -n as -c '{"Args":["GetTGT","tgt_abc123"]}'
```

**Returns**: JSON-encoded TGT object

**When to use**: TGS chaincode validation (cross-chaincode call), debugging, audit

---

#### 6. `RevokeDevice`
**Purpose**: Revoke a device's access privileges

**Signature**:
```go
func RevokeDevice(ctx, deviceID string) error
```

**Example**:
```bash
peer chaincode invoke -C authchannel -n as -c '{"Args":["RevokeDevice","device_001"]}'
```

**Returns**: `nil` on success

**Events Emitted**: `DeviceRevoked` with deviceID payload

**State Changes**:
- Updates device status to "revoked"
- Device record remains in ledger (audit trail)

**When to use**: Compromised devices, decommissioned devices, security incidents

**Note**: Does NOT invalidate existing TGTs. In production, add TGT cleanup logic.

---

#### 7. `GetAllDevices`
**Purpose**: Retrieve all registered devices (admin function)

**Signature**:
```go
func GetAllDevices(ctx) (string, error)
```

**Example**:
```bash
peer chaincode query -C authchannel -n as -c '{"Args":["GetAllDevices"]}'
```

**Returns**: JSON-encoded array of Device objects

**Performance Note**: Uses `GetStateByRange` which iterates all keys. For large deployments (>10K devices), consider pagination using rich queries.

**When to use**: Device inventory, admin dashboards, reporting

---

### Internal Helper Functions

#### `getCurrentTimestamp() int64`
**Current Implementation**: Returns placeholder `1672531200`

**Production Implementation Should**:
```go
func getCurrentTimestamp() int64 {
    timestamp, err := ctx.GetStub().GetTxTimestamp()
    if err != nil {
        return time.Now().Unix()
    }
    return timestamp.Seconds
}
```

**Why use transaction timestamp?**
- Ensures all peers agree on the same timestamp (consensus)
- Prevents timestamp manipulation attacks

---

#### `generateSecureSessionKey() (string, error)`
**Current Implementation**: Placeholder returning `"secure_session_key_" + random`

**Production Implementation Should**:
```go
import "crypto/rand"

func generateSecureSessionKey() (string, error) {
    key := make([]byte, 32) // 256-bit key
    _, err := rand.Read(key)
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(key), nil
}
```

**Why 256-bit?**
- AES-256 is current industry standard for symmetric encryption
- Provides 128-bit security level (2^128 operations to break)

---

#### `generateSecureTgtID() (string, error)`
**Current Implementation**: Placeholder

**Production Implementation**: Same as `generateSecureSessionKey` but 128-bit (16 bytes)

**Why separate function?**
- Different security requirements (ID doesn't need to be as strong as encryption key)
- Allows independent modification of ID generation strategy

## Security Considerations

### ✅ Implemented
1. **Input Validation**: DeviceID and PublicKey length constraints
2. **Timestamp Validation**: ±5 minute window prevents replay attacks
3. **Status Checks**: Only "active" devices can authenticate
4. **Event Logging**: All critical operations emit blockchain events

### ⚠️ Placeholder (Needs Production Implementation)
1. **Signature Verification**: Currently checks `len(signature) > 10`, should verify using device's public key
2. **Secure Random Generation**: Helper functions use placeholders, should use `crypto/rand`
3. **Timestamp Source**: Uses hardcoded value, should use `ctx.GetStub().GetTxTimestamp()`

### 🔒 Recommended Additions
1. **Rate Limiting**: Prevent brute-force authentication attempts (use common/ratelimit.go)
2. **Audit Logging**: Use common/audit.go for comprehensive logging
3. **TGT Cleanup**: Automatically expire old TGTs (garbage collection)
4. **Cross-Chaincode Access Control**: Ensure only TGS can call GetTGT

## Usage Flow

### Device Registration Flow
```
1. Admin/Device → RegisterDevice(deviceID, publicKey, metadata)
2. AS Chaincode validates inputs
3. Creates Device record with status "active"
4. Stores in ledger: Key=deviceID
5. Emits DeviceRegistered event
6. Returns success
```

### Authentication Flow
```
1. Device generates nonce
2. Device signs (deviceID + nonce + timestamp) with private key
3. Device → Authenticate(AuthRequest JSON)
4. AS Chaincode retrieves Device record
5. Validates status, timestamp, signature
6. Generates secure session key (crypto/rand)
7. Creates TGT with 1-hour expiry
8. Stores TGT: Key=TGT_{tgtID}
9. Updates Device.LastAuthTime
10. Emits DeviceAuthenticated event
11. Returns AuthResponse with TGT details
```

### Revocation Flow
```
1. Admin → RevokeDevice(deviceID)
2. AS Chaincode retrieves Device
3. Updates status to "revoked"
4. Stores updated Device
5. Emits DeviceRevoked event
6. Returns success

Note: Existing TGTs remain valid until expiration
```

## Integration with Other Chaincodes

### TGS Chaincode Integration
The TGS chaincode needs to verify TGTs issued by AS. This is done via **cross-chaincode invocation**:

```go
// In TGS Chaincode
func verifyTGT(ctx, tgtID string) (TGT, error) {
    response := ctx.GetStub().InvokeChaincode("as", [][]byte{
        []byte("GetTGT"),
        []byte(tgtID),
    }, "authchannel")

    if response.Status != shim.OK {
        return TGT{}, fmt.Errorf("TGT verification failed")
    }

    var tgt TGT
    json.Unmarshal(response.Payload, &tgt)

    // Check expiration, status, etc.
    if tgt.Status != "valid" || getCurrentTimestamp() > tgt.ExpiresAt {
        return TGT{}, fmt.Errorf("invalid or expired TGT")
    }

    return tgt, nil
}
```

## Building and Deploying

### Build Chaincode
```bash
cd chaincodes/as-chaincode
go mod tidy
go build -v .
```

### Package Chaincode
```bash
peer lifecycle chaincode package as.tar.gz \
  --path chaincodes/as-chaincode \
  --lang golang \
  --label as_1.0
```

### Deploy (via automated script)
```bash
cd network
./scripts/deploy-chaincode.sh as
```

**What the deployment script does**:
1. Packages chaincode as `as.tar.gz`
2. Installs on all peers (Org1, Org2, Org3)
3. Approves for each organization
4. Commits chaincode definition to channel
5. Initializes ledger (calls InitLedger)

### Verify Deployment
```bash
# Check committed chaincodes
peer lifecycle chaincode querycommitted -C authchannel -n as

# Test invocation
peer chaincode invoke \
  -C authchannel \
  -n as \
  -c '{"Args":["InitLedger"]}'
```

## Testing

### Unit Tests
Located in `tests/unit/as_chaincode_test.go`

```bash
cd tests
./run-tests.sh unit
```

**Test Coverage**:
- Device registration (success, duplicate, validation)
- Authentication (success, invalid device, expired timestamp)
- Device retrieval
- Revocation

### Integration Tests
Located in `tests/integration/authentication_flow_test.go`

```bash
./run-tests.sh integration
```

**Test Scenarios**:
- Complete flow: Register → Authenticate → Request Ticket → Validate Access
- TGT expiration handling
- Revoked device authentication attempt

### Manual Testing
```bash
# 1. Register device
docker exec cli peer chaincode invoke \
  -C authchannel -n as \
  -c '{"Args":["RegisterDevice","test_device","-----BEGIN PUBLIC KEY-----\nMIIBIj...","Test Device"]}'

# 2. Authenticate
docker exec cli peer chaincode invoke \
  -C authchannel -n as \
  -c '{"Args":["Authenticate","{\"deviceID\":\"test_device\",\"nonce\":\"abc123\",\"timestamp\":1672531200,\"signature\":\"test_signature_abcdef\"}"]}'

# 3. Query device
docker exec cli peer chaincode query \
  -C authchannel -n as \
  -c '{"Args":["GetDevice","test_device"]}'

# 4. Revoke
docker exec cli peer chaincode invoke \
  -C authchannel -n as \
  -c '{"Args":["RevokeDevice","test_device"]}'
```

## Troubleshooting

### Common Issues

**Error**: `"device already exists"`
- **Cause**: Attempting to register a device with an existing deviceID
- **Solution**: Use a different deviceID or query existing device first

**Error**: `"device not found"`
- **Cause**: Authenticating with unregistered deviceID
- **Solution**: Register device first using RegisterDevice

**Error**: `"timestamp is invalid or too old"`
- **Cause**: Request timestamp differs from chaincode time by > 5 minutes
- **Solution**: Ensure device clock is synchronized (use NTP), check timestamp format (Unix seconds)

**Error**: `"device is not active (status: revoked)"`
- **Cause**: Attempting to authenticate with revoked device
- **Solution**: Check device status, re-register if necessary (requires admin approval)

**Error**: `"failed to marshal device"`
- **Cause**: Invalid characters in device fields
- **Solution**: Ensure deviceID is alphanumeric, publicKey is valid PEM format

### Debugging

**Enable chaincode logs**:
```bash
docker logs -f peer0.org1.example.com
```

**Check ledger state**:
```bash
peer chaincode query \
  -C authchannel \
  -n as \
  -c '{"Args":["GetAllDevices"]}'
```

**Inspect transactions**:
```bash
peer chaincode invoke \
  -C authchannel \
  -n as \
  -c '{"Args":["GetDevice","device_001"]}' \
  --waitForEvent
```

## Performance Considerations

### State Management
- **Devices**: Stored with simple key `deviceID` for O(1) lookup
- **TGTs**: Prefixed with `TGT_` for namespacing and range queries
- **Cleanup**: No automatic TGT deletion (implement garbage collection for production)

### Scalability
- **GetAllDevices**: O(n) operation, use sparingly in production
- **Concurrent Authentication**: Multiple devices can authenticate simultaneously (Fabric handles concurrency)
- **TGT Storage**: Consider TTL-based cleanup or periodic archival for large deployments

### Optimization Tips
1. **Pagination**: Modify GetAllDevices to support pagination for >1000 devices
2. **Rich Queries**: Use CouchDB state database for complex queries (requires JSON indexes)
3. **Caching**: Client-side caching of device public keys reduces GetDevice calls

## Learn More

- **[TGS Chaincode](../tgs-chaincode/README.md)** - Next step: Service ticket issuance
- **[ISV Chaincode](../isv-chaincode/README.md)** - Final step: Access validation
- **[Common Utilities](../common/README.md)** - Shared functions used by AS
- **[API Reference](../../docs/api/chaincode-api.md)** - Complete API documentation
- **[Developer Guide](../../DEVELOPER_GUIDE.md)** - In-depth technical guide

## Navigation

📍 **Path**: [Main README](../../README.md) → [Chaincodes](../README.md) → **AS Chaincode** → [TGS Chaincode](../tgs-chaincode/README.md)

🔗 **Quick Links**:
- [← Common Utilities](../common/README.md)
- [TGS Chaincode →](../tgs-chaincode/README.md)
- [HOW IT WORKS](../../HOW_IT_WORKS.md)
- [Deployment Guide](../../docs/deployment/PRODUCTION_DEPLOYMENT.md)

# Developer Guide - Blockchain Authentication Framework

**Target Audience**: Developers who want to understand, modify, or extend this codebase.

**Prerequisites**: Understanding of Go, Hyperledger Fabric, and blockchain concepts.

## Table of Contents

1. [Development Environment Setup](#development-environment-setup)
2. [Chaincode Development](#chaincode-development)
3. [Function-by-Function Breakdown](#function-by-function-breakdown)
4. [Design Patterns Used](#design-patterns-used)
5. [Testing Strategies](#testing-strategies)
6. [Debugging Techniques](#debugging-techniques)
7. [Extending the Framework](#extending-the-framework)
8. [Best Practices](#best-practices)
9. [Common Pitfalls](#common-pitfalls)

---

## Development Environment Setup

### IDE Configuration

**VS Code** (Recommended):
```json
// .vscode/settings.json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.formatTool": "goimports",
    "go.testFlags": ["-v", "-race"],
    "editor.formatOnSave": true
}
```

**Required VS Code Extensions**:
- Go (golang.go)
- Docker (ms-azuretools.vscode-docker)
- YAML (redhat.vscode-yaml)

### Local Development Setup

```bash
# 1. Install development tools
go install golang.org/x/tools/gopls@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# 2. Setup Git hooks
./scripts/install-hooks.sh

# 3. Install dependencies
make install-deps

# 4. Verify setup
go version  # Should be 1.21+
docker --version
make verify
```

---

## Chaincode Development

### Chaincode Structure

Each chaincode follows this structure:

```
as-chaincode/
├── as-chaincode.go       # Main chaincode implementation
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
└── README.md             # Chaincode-specific docs
```

### Implementing a New Chaincode Function

**Step 1: Define the Function Signature**

```go
// In as-chaincode.go

// MyNewFunction does something useful
// @param ctx Transaction context
// @param param1 First parameter
// @param param2 Second parameter
// @return result string or error
func (s *ASChaincode) MyNewFunction(
    ctx contractapi.TransactionContextInterface,
    param1 string,
    param2 string) (string, error) {

    // Implementation here
}
```

**Step 2: Input Validation**

```go
func (s *ASChaincode) MyNewFunction(
    ctx contractapi.TransactionContextInterface,
    param1 string,
    param2 string) (string, error) {

    // Validate param1
    if err := ValidateDeviceID(param1); err != nil {
        return "", fmt.Errorf("invalid param1: %v", err)
    }

    // Validate param2
    if len(param2) == 0 || len(param2) > 256 {
        return "", fmt.Errorf("param2 must be 1-256 chars")
    }

    // Continue...
}
```

**Step 3: State Access**

```go
func (s *ASChaincode) MyNewFunction(...) (string, error) {
    // ... validation ...

    // Read from world state
    existingData, err := ctx.GetStub().GetState(key)
    if err != nil {
        return "", fmt.Errorf("failed to read state: %v", err)
    }

    // Parse JSON
    var data MyDataStructure
    if existingData != nil {
        err = json.Unmarshal(existingData, &data)
        if err != nil {
            return "", fmt.Errorf("failed to parse data: %v", err)
        }
    }

    // Modify data
    data.Field1 = newValue
    data.UpdatedAt = getCurrentTimestamp()

    // Write back to state
    dataJSON, err := json.Marshal(data)
    if err != nil {
        return "", fmt.Errorf("failed to marshal data: %v", err)
    }

    err = ctx.GetStub().PutState(key, dataJSON)
    if err != nil {
        return "", fmt.Errorf("failed to write state: %v", err)
    }

    // Emit event
    ctx.GetStub().SetEvent("MyEventName", []byte(key))

    return "success", nil
}
```

---

## Function-by-Function Breakdown

### AS Chaincode: RegisterDevice

**Location**: `chaincodes/as-chaincode/as-chaincode.go:30`

**Purpose**: Register a new IoT device with the authentication server.

**Parameters**:
- `deviceID` (string): Unique identifier for the device
- `publicKey` (string): PEM-encoded RSA/ECDSA public key
- `metadata` (string): Optional device information

**Internal Logic**:

```go
func (s *ASChaincode) RegisterDevice(
    ctx contractapi.TransactionContextInterface,
    deviceID string,
    publicKey string,
    metadata string) error {

    // Step 1: Check if device already exists
    existing, err := ctx.GetStub().GetState(deviceID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if existing != nil {
        return fmt.Errorf("device %s already exists", deviceID)
    }

    // Step 2: Validate inputs
    if len(deviceID) < 3 || len(deviceID) > 64 {
        return fmt.Errorf("deviceID must be between 3 and 64 characters")
    }
    if len(publicKey) < 100 || len(publicKey) > 4096 {
        return fmt.Errorf("publicKey length is invalid")
    }

    // Step 3: Create device object
    device := Device{
        DeviceID:        deviceID,
        PublicKey:       publicKey,
        Status:          "active",
        RegistrationTime: getCurrentTimestamp(),
        LastAuthTime:    0,
        Metadata:        metadata,
    }

    // Step 4: Marshal to JSON
    deviceJSON, err := json.Marshal(device)
    if err != nil {
        return fmt.Errorf("failed to marshal device: %v", err)
    }

    // Step 5: Write to world state
    err = ctx.GetStub().PutState(deviceID, deviceJSON)
    if err != nil {
        return fmt.Errorf("failed to put device to world state: %v", err)
    }

    // Step 6: Emit event
    err = ctx.GetStub().SetEvent("DeviceRegistered", []byte(deviceID))
    if err != nil {
        return fmt.Errorf("failed to set event: %v", err)
    }

    log.Printf("Device %s registered successfully", deviceID)
    return nil
}
```

**Why This Implementation?**

1. **Idempotency Check**: Prevents duplicate registrations
2. **Input Validation**: Prevents malformed data from entering the ledger
3. **Structured Data**: Uses Go structs for type safety
4. **Event Emission**: Allows clients to listen for registration events
5. **Logging**: Helps with debugging and auditing

**Error Handling**:
- Returns descriptive errors for each failure case
- Fails fast on validation errors
- Logs success for audit trail

### AS Chaincode: Authenticate

**Location**: `chaincodes/as-chaincode/as-chaincode.go:73`

**Purpose**: Authenticate a device and issue a TGT.

**Internal Logic Breakdown**:

```go
func (s *ASChaincode) Authenticate(
    ctx contractapi.TransactionContextInterface,
    authRequestJSON string) (string, error) {

    // SECTION 1: Parse Request
    // =======================
    var authReq AuthRequest
    err := json.Unmarshal([]byte(authRequestJSON), &authReq)
    if err != nil {
        return "", fmt.Errorf("failed to unmarshal auth request: %v", err)
    }

    // SECTION 2: Device Lookup
    // =========================
    deviceJSON, err := ctx.GetStub().GetState(authReq.DeviceID)
    if err != nil {
        return "", fmt.Errorf("failed to read device: %v", err)
    }
    if deviceJSON == nil {
        return "", fmt.Errorf("device %s not found", authReq.DeviceID)
    }

    var device Device
    err = json.Unmarshal(deviceJSON, &device)
    if err != nil {
        return "", fmt.Errorf("failed to unmarshal device: %v", err)
    }

    // SECTION 3: Status Validation
    // =============================
    if device.Status != "active" {
        return "", fmt.Errorf("device is not active (status: %s)", device.Status)
    }

    // SECTION 4: Timestamp Validation
    // ================================
    currentTime := getCurrentTimestamp()
    if authReq.Timestamp < currentTime-300 ||
       authReq.Timestamp > currentTime+300 {
        return "", fmt.Errorf("timestamp is invalid or too old")
    }

    // SECTION 5: Signature Verification
    // ==================================
    // In production, verify signature using device.PublicKey:
    // 1. Reconstruct message: deviceID || nonce || timestamp
    // 2. Decode signature from base64
    // 3. Verify using RSA/ECDSA
    //
    // Example (production code):
    // message := []byte(authReq.DeviceID + authReq.Nonce + 
    //                   strconv.FormatInt(authReq.Timestamp, 10))
    // signatureBytes, _ := base64.StdEncoding.DecodeString(authReq.Signature)
    // publicKeyBytes := []byte(device.PublicKey)
    // block, _ := pem.Decode(publicKeyBytes)
    // pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
    // rsaPub := pub.(*rsa.PublicKey)
    // hashed := sha256.Sum256(message)
    // err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hashed[:], signatureBytes)

    if len(authReq.Signature) < 10 {
        return "", fmt.Errorf("invalid signature")
    }

    // SECTION 6: Generate Session Key
    // ================================
    sessionKey, err := generateSecureSessionKey()
    if err != nil {
        return "", fmt.Errorf("failed to generate session key: %v", err)
    }

    // SECTION 7: Generate TGT ID
    // ===========================
    tgtID, err := generateSecureTgtID()
    if err != nil {
        return "", fmt.Errorf("failed to generate TGT ID: %v", err)
    }

    // SECTION 8: Create TGT
    // ======================
    issuedAt := getCurrentTimestamp()
    expiresAt := issuedAt + 3600 // 1 hour

    tgt := TGT{
        TgtID:      tgtID,
        DeviceID:   authReq.DeviceID,
        SessionKey: sessionKey,
        IssuedAt:   issuedAt,
        ExpiresAt:  expiresAt,
        Status:     "valid",
    }

    // SECTION 9: Store TGT
    // =====================
    tgtJSON, err := json.Marshal(tgt)
    if err != nil {
        return "", fmt.Errorf("failed to marshal TGT: %v", err)
    }

    err = ctx.GetStub().PutState("TGT_"+tgtID, tgtJSON)
    if err != nil {
        return "", fmt.Errorf("failed to store TGT: %v", err)
    }

    // SECTION 10: Update Device
    // ==========================
    device.LastAuthTime = currentTime
    deviceJSON, err = json.Marshal(device)
    if err != nil {
        return "", fmt.Errorf("failed to marshal updated device: %v", err)
    }
    err = ctx.GetStub().PutState(authReq.DeviceID, deviceJSON)
    if err != nil {
        return "", fmt.Errorf("failed to update device: %v", err)
    }

    // SECTION 11: Create Response
    // ============================
    response := AuthResponse{
        TgtID:      tgtID,
        SessionKey: sessionKey,
        ExpiresAt:  expiresAt,
        Message:    "Authentication successful",
    }

    responseJSON, err := json.Marshal(response)
    if err != nil {
        return "", fmt.Errorf("failed to marshal response: %v", err)
    }

    // SECTION 12: Emit Event
    // =======================
    err = ctx.GetStub().SetEvent("DeviceAuthenticated", []byte(authReq.DeviceID))
    if err != nil {
        return "", fmt.Errorf("failed to set event: %v", err)
    }

    log.Printf("Device %s authenticated successfully, TGT: %s", authReq.DeviceID, tgtID)
    return string(responseJSON), nil
}
```

**Key Design Decisions**:

1. **JSON Input/Output**: Flexible, extensible, language-agnostic
2. **Timestamp Window**: 5-minute window allows for clock skew
3. **Separate TGT Key**: Using "TGT_" prefix prevents collision with device IDs
4. **Update Last Auth Time**: Tracks device activity
5. **Event Emission**: Enables real-time monitoring

### Common Package: GenerateSecureRandomBytes

**Location**: `chaincodes/common/utils.go:11`

**Purpose**: Generate cryptographically secure random bytes.

**Implementation**:

```go
func GenerateSecureRandomBytes(length int) ([]byte, error) {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to generate random bytes: %v", err)
    }
    return bytes, nil
}
```

**Why crypto/rand?**

```go
// BAD: Predictable pseudo-random
import "math/rand"
nonce := rand.Intn(1000000)

// GOOD: Cryptographically secure
import "crypto/rand"
nonceBytes := make([]byte, 32)
rand.Read(nonceBytes)
```

**crypto/rand** uses operating system entropy sources:
- `/dev/urandom` on Linux
- `CryptGenRandom` on Windows
- `getentropy()` on macOS

This ensures unpredictability even if attacker knows the algorithm.

---

## Design Patterns Used

### 1. Repository Pattern

**Purpose**: Abstract data access logic.

**Implementation**:

```go
// Instead of scattered GetState calls:
device1, _ := ctx.GetStub().GetState("device_001")
device2, _ := ctx.GetStub().GetState("device_002")

// Use repository pattern:
type DeviceRepository struct {
    ctx contractapi.TransactionContextInterface
}

func (r *DeviceRepository) GetByID(deviceID string) (*Device, error) {
    deviceJSON, err := r.ctx.GetStub().GetState(deviceID)
    if err != nil {
        return nil, err
    }
    if deviceJSON == nil {
        return nil, fmt.Errorf("device not found")
    }

    var device Device
    err = json.Unmarshal(deviceJSON, &device)
    return &device, err
}

func (r *DeviceRepository) Save(device *Device) error {
    deviceJSON, _ := json.Marshal(device)
    return r.ctx.GetStub().PutState(device.DeviceID, deviceJSON)
}
```

**Benefits**:
- Centralized data access
- Easier testing
- Consistent error handling

### 2. Factory Pattern

**Purpose**: Create objects with complex initialization.

**Example**:

```go
type TGTFactory struct {
    validityDuration time.Duration
}

func NewTGTFactory(duration time.Duration) *TGTFactory {
    return &TGTFactory{validityDuration: duration}
}

func (f *TGTFactory) CreateTGT(deviceID string) (*TGT, error) {
    tgtID, _ := generateSecureTgtID()
    sessionKey, _ := generateSecureSessionKey()
    
    now := getCurrentTimestamp()
    
    return &TGT{
        TgtID:      tgtID,
        DeviceID:   deviceID,
        SessionKey: sessionKey,
        IssuedAt:   now,
        ExpiresAt:  now + int64(f.validityDuration.Seconds()),
        Status:     "valid",
    }, nil
}
```

### 3. Builder Pattern

**Purpose**: Construct complex objects step-by-step.

**Example**:

```go
type AccessRequestBuilder struct {
    request AccessRequest
}

func NewAccessRequestBuilder() *AccessRequestBuilder {
    return &AccessRequestBuilder{
        request: AccessRequest{
            Timestamp: getCurrentTimestamp(),
        },
    }
}

func (b *AccessRequestBuilder) WithDevice(deviceID string) *AccessRequestBuilder {
    b.request.DeviceID = deviceID
    return b
}

func (b *AccessRequestBuilder) WithService(serviceID string) *AccessRequestBuilder {
    b.request.ServiceID = serviceID
    return b
}

func (b *AccessRequestBuilder) WithTicket(ticketID string) *AccessRequestBuilder {
    b.request.TicketID = ticketID
    return b
}

func (b *AccessRequestBuilder) Build() AccessRequest {
    return b.request
}

// Usage:
request := NewAccessRequestBuilder().
    WithDevice("device_001").
    WithService("service001").
    WithTicket("ticket_xyz").
    Build()
```

### 4. Strategy Pattern

**Purpose**: Define a family of algorithms and make them interchangeable.

**Example - Rate Limiting Strategies**:

```go
type RateLimitStrategy interface {
    AllowRequest(deviceID string) (bool, error)
}

// Token Bucket Strategy
type TokenBucketStrategy struct {
    tokensPerMinute int
}

func (s *TokenBucketStrategy) AllowRequest(deviceID string) (bool, error) {
    // Implementation
}

// Leaky Bucket Strategy
type LeakyBucketStrategy struct {
    capacity int
}

func (s *LeakyBucketStrategy) AllowRequest(deviceID string) (bool, error) {
    // Implementation
}

// Usage:
var strategy RateLimitStrategy
if useTokenBucket {
    strategy = &TokenBucketStrategy{tokensPerMinute: 60}
} else {
    strategy = &LeakyBucketStrategy{capacity: 100}
}

allowed, _ := strategy.AllowRequest("device_001")
```

---

## Testing Strategies

### Unit Testing

**Structure**:
```
tests/unit/
├── as_chaincode_test.go
├── tgs_chaincode_test.go
├── isv_chaincode_test.go
└── common_test.go
```

**Example Unit Test**:

```go
package unit

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestDeviceRegistration(t *testing.T) {
    t.Run("Valid device registration", func(t *testing.T) {
        // Arrange
        deviceID := "test_device_001"
        publicKey := "-----BEGIN PUBLIC KEY-----\nMIIBIj..."
        
        // Act
        // (In real test, invoke chaincode)
        
        // Assert
        assert.NotEmpty(t, deviceID)
    })
    
    t.Run("Reject duplicate device", func(t *testing.T) {
        // Test duplicate registration
    })
}
```

### Integration Testing

**Example**:

```go
func TestFullAuthenticationFlow(t *testing.T) {
    // Setup network
    network := SetupTestNetwork()
    defer network.Teardown()
    
    // Register device
    deviceID := "integration_test_001"
    err := network.RegisterDevice(deviceID, publicKey, metadata)
    assert.NoError(t, err)
    
    // Authenticate
    tgt, err := network.Authenticate(deviceID, nonce, timestamp, signature)
    assert.NoError(t, err)
    assert.NotEmpty(t, tgt.TgtID)
    
    // Request service ticket
    ticket, err := network.RequestServiceTicket(deviceID, tgt.TgtID, "service001")
    assert.NoError(t, err)
    
    // Validate access
    session, err := network.ValidateAccess(deviceID, ticket.TicketID, "read")
    assert.NoError(t, err)
    assert.True(t, session.Granted)
}
```

### Mocking

**Example Mock Stub**:

```go
type MockStub struct {
    state map[string][]byte
}

func (m *MockStub) GetState(key string) ([]byte, error) {
    return m.state[key], nil
}

func (m *MockStub) PutState(key string, value []byte) error {
    m.state[key] = value
    return nil
}

// Usage in tests:
stub := &MockStub{state: make(map[string][]byte)}
// Test chaincode functions with mock stub
```

---

## Debugging Techniques

### 1. Chaincode Logging

```go
import "log"

func (s *ASChaincode) MyFunction() error {
    log.Printf("[DEBUG] Entering MyFunction")
    log.Printf("[INFO] Processing device: %s", deviceID)
    log.Printf("[ERROR] Failed to validate: %v", err)
    return nil
}
```

View logs:
```bash
docker logs -f dev-peer0.org1.example.com-as-1.0-xyz
```

### 2. Debugging with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug chaincode
cd chaincodes/as-chaincode
dlv debug

# Set breakpoint
(dlv) break RegisterDevice
(dlv) continue
```

### 3. Network Debugging

```bash
# Check peer logs
docker logs peer0.org1.example.com

# Check orderer logs
docker logs orderer.example.com

# Check chaincode container logs
docker logs $(docker ps -f name=dev-peer0.org1 -q)

# Inspect world state (if using CouchDB)
curl http://localhost:5984/authchannel/_all_docs
```

### 4. Transaction Tracing

```bash
# Get transaction by ID
peer chaincode query \
  -C authchannel \
  -n qscc \
  -c '{"Args":["GetTransactionByID", "authchannel", "TX_ID"]}'

# Get block by number
peer chaincode query \
  -C authchannel \
  -n qscc \
  -c '{"Args":["GetBlockByNumber", "authchannel", "5"]}'
```

---

## Extending the Framework

### Adding a New Chaincode Function

**Example: Add Device Update Function**

1. **Define function in chaincode**:

```go
// UpdateDeviceMetadata updates device metadata
func (s *ASChaincode) UpdateDeviceMetadata(
    ctx contractapi.TransactionContextInterface,
    deviceID string,
    newMetadata string) error {
    
    // Get existing device
    deviceJSON, err := ctx.GetStub().GetState(deviceID)
    if err != nil {
        return fmt.Errorf("failed to read device: %v", err)
    }
    if deviceJSON == nil {
        return fmt.Errorf("device not found")
    }
    
    var device Device
    json.Unmarshal(deviceJSON, &device)
    
    // Update metadata
    device.Metadata = newMetadata
    
    // Save back
    deviceJSON, _ = json.Marshal(device)
    ctx.GetStub().PutState(deviceID, deviceJSON)
    
    // Emit event
    ctx.GetStub().SetEvent("DeviceUpdated", []byte(deviceID))
    
    return nil
}
```

2. **Add API documentation**:

Update `docs/api/chaincode-api.md` with new function details.

3. **Add tests**:

```go
func TestUpdateDeviceMetadata(t *testing.T) {
    // Test implementation
}
```

4. **Upgrade chaincode**:

```bash
# Package new version
peer lifecycle chaincode package as_v2.tar.gz \
  --path ./chaincodes/as-chaincode \
  --lang golang \
  --label as_2.0

# Install, approve, and commit with sequence incremented
```

### Adding Cross-Chaincode Communication

**Example: TGS validates TGT with AS**

```go
func (s *TGSChaincode) IssueServiceTicket(...) (string, error) {
    // Cross-chaincode call to AS
    response := ctx.GetStub().InvokeChaincode(
        "as",
        [][]byte{
            []byte("GetTGT"),
            []byte(ticketReq.TgtID),
        },
        "authchannel")
    
    if response.Status != shim.OK {
        return "", fmt.Errorf("TGT validation failed: %s", response.Message)
    }
    
    var tgt TGT
    json.Unmarshal(response.Payload, &tgt)
    
    // Verify TGT is valid
    if tgt.Status != "valid" {
        return "", fmt.Errorf("TGT is not valid")
    }
    
    // Continue with ticket issuance...
}
```

---

## Best Practices

### 1. Always Validate Input

```go
// BAD: No validation
func MyFunction(input string) error {
    data := processInput(input)
    saveData(data)
}

// GOOD: Comprehensive validation
func MyFunction(input string) error {
    if err := ValidateInput(input); err != nil {
        return fmt.Errorf("validation failed: %v", err)
    }
    
    data := processInput(input)
    if err := saveData(data); err != nil {
        return fmt.Errorf("save failed: %v", err)
    }
    
    return nil
}
```

### 2. Use Structured Logging

```go
// BAD: Unstructured logs
log.Printf("Error in function")

// GOOD: Structured logs
log.Printf("[ERROR] Function: %s, Device: %s, Error: %v", 
    "RegisterDevice", deviceID, err)
```

### 3. Handle Errors Gracefully

```go
// BAD: Silent failure
deviceJSON, err := ctx.GetStub().GetState(deviceID)
if err != nil {
    return nil
}

// GOOD: Proper error propagation
deviceJSON, err := ctx.GetStub().GetState(deviceID)
if err != nil {
    return fmt.Errorf("failed to read device %s: %v", deviceID, err)
}
```

### 4. Use Constants

```go
// Define constants
const (
    TGTValiditySeconds = 3600
    ServiceTicketValiditySeconds = 1800
    MaxDeviceIDLength = 64
    MinDeviceIDLength = 3
)

// Use in code
expiresAt := issuedAt + TGTValiditySeconds
```

### 5. Emit Events

```go
// Always emit events for important operations
ctx.GetStub().SetEvent("DeviceRegistered", []byte(deviceID))
ctx.GetStub().SetEvent("AuthenticationFailed", []byte(deviceID))
ctx.GetStub().SetEvent("TGTExpired", []byte(tgtID))
```

---

## Common Pitfalls

### 1. Determinism Violations

**Problem**: Chaincode must be deterministic (same input = same output).

```go
// BAD: Non-deterministic (uses current time)
timestamp := time.Now().Unix()

// GOOD: Use transaction timestamp
txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
timestamp := txTimestamp.Seconds
```

### 2. Large State Writes

**Problem**: Writing large objects to state is expensive.

```go
// BAD: Storing large binary data
device.ImageData = largeImageBytes // 10MB

// GOOD: Store reference
device.ImageURL = "ipfs://Qm..."
```

### 3. Inefficient Queries

```go
// BAD: Iterate all keys
iterator, _ := ctx.GetStub().GetStateByRange("", "")
for iterator.HasNext() {
    // Process every key
}

// GOOD: Use specific key patterns
iterator, _ := ctx.GetStub().GetStateByRange("device_", "device_~")
```

### 4. Not Checking for Nil

```go
// BAD: May panic
var device Device
json.Unmarshal(deviceJSON, &device)
status := device.Status // Panic if deviceJSON is nil

// GOOD: Check for nil
if deviceJSON == nil {
    return fmt.Errorf("device not found")
}
```

### 5. Hardcoded Values

```go
// BAD: Magic numbers
if counter > 60 { ... }

// GOOD: Named constants
const MaxRequestsPerMinute = 60
if counter > MaxRequestsPerMinute { ... }
```

---

## Performance Tips

### 1. Batch Operations

```go
// BAD: Multiple separate transactions
for _, device := range devices {
    RegisterDevice(device)
}

// GOOD: Single transaction with batch
func BatchRegisterDevices(devices []Device) error {
    for _, device := range devices {
        deviceJSON, _ := json.Marshal(device)
        ctx.GetStub().PutState(device.DeviceID, deviceJSON)
    }
    return nil
}
```

### 2. Use Pagination

```go
// Query with pagination
func GetDevicesPaginated(pageSize int32, bookmark string) (*QueryResult, error) {
    query := `{"selector":{"docType":"device"}}`
    iterator, metadata, _ := ctx.GetStub().GetQueryResultWithPagination(
        query, pageSize, bookmark)
    
    // Process results
    return results, nil
}
```

### 3. Optimize JSON

```go
// Use json tags to control serialization
type Device struct {
    DeviceID  string `json:"id"`           // Shorter field name
    PublicKey string `json:"key,omitempty"` // Omit if empty
    LargeData string `json:"-"`            // Never serialize
}
```

---

## Summary

This guide covers the essential aspects of developing and extending the blockchain authentication framework. Key takeaways:

1. **Validation is critical** - Always validate all inputs
2. **Determinism matters** - Avoid non-deterministic operations
3. **Events enable monitoring** - Emit events for important operations
4. **Error handling is not optional** - Always handle and propagate errors
5. **Test thoroughly** - Unit tests, integration tests, and performance tests
6. **Follow patterns** - Use established design patterns
7. **Document everything** - Code comments, API docs, and architectural docs

For more information:
- [HOW_IT_WORKS.md](HOW_IT_WORKS.md) - Understanding the system
- [API Reference](docs/api/chaincode-api.md) - Complete API documentation
- [Hyperledger Fabric Docs](https://hyperledger-fabric.readthedocs.io/) - Official documentation

Happy coding! 🚀

# Common Utilities Package

📍 **Location**: `chaincodes/common/`
🔗 **Parent Guide**: [Chaincodes Overview](../README.md)
📚 **Used By**: AS Chaincode | TGS Chaincode | ISV Chaincode

---

## 📋 Overview

This package contains **shared utility functions** used by all three chaincodes. Instead of duplicating code, we centralize common functionality here.

**DRY Principle**: Don't Repeat Yourself
- Write once, use everywhere
- Fix bugs in one place
- Consistent behavior across chaincodes

---

## 📁 Files in This Directory

### 1. `utils.go` - Core Utilities

**Purpose**: Cryptographic operations, random generation, timestamp handling

**Key Functions**:

```go
// Secure Random Generation
GenerateSecureRandomBytes(length int) ([]byte, error)
GenerateSecureNonce() (string, error)          // 256-bit
GenerateSessionKey() (string, error)           // 256-bit
GenerateTicketID() (string, error)
GenerateDeviceID() (string, error)

// Hashing
HashData(data string) string                   // SHA-256

// Time Operations
GetCurrentTimestamp() int64
IsExpired(timestamp, validity int64) bool
ValidateTimestamp(timestamp, maxAge int64) error

// Encoding
EncodeToHex(data []byte) string
DecodeFromHex(data string) ([]byte, error)
EncodeToBase64(data []byte) string
DecodeFromBase64(data string) ([]byte, error)
```

**Why crypto/rand?**
```go
// ❌ BAD: Predictable
import "math/rand"
nonce := rand.Intn(1000000)

// ✅ GOOD: Cryptographically secure
import "crypto/rand"
nonce, _ := GenerateSecureRandomBytes(32)
```

**Usage Example**:
```go
import "github.com/blockchain-auth/common"

// Generate secure session key
sessionKey, err := common.GenerateSessionKey()
if err != nil {
    return fmt.Errorf("key generation failed: %v", err)
}

// Validate timestamp
err = common.ValidateTimestamp(requestTime, 300) // 5 min window
if err != nil {
    return fmt.Errorf("timestamp invalid: %v", err)
}
```

---

### 2. `validation.go` - Input Validation

**Purpose**: Validate all user inputs before processing

**Constants**:
```go
const (
    MaxIDLength = 64
    MinIDLength = 3
    MaxPEMLength = 4096
    MaxMetadataLength = 1024
)
```

**Validation Functions**:
```go
ValidateDeviceID(deviceID string) error
ValidateServiceID(serviceID string) error
ValidatePublicKey(publicKey string) error
ValidateSignature(signature string) error
ValidateNonce(nonce string) error
ValidateMetadata(metadata string) error
ValidateAction(action string) error
ValidateIPAddress(ipAddress string) error
ValidateTimestampRange(timestamp, maxAge int64) error
```

**Regex Patterns**:
```go
ValidIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
ValidActionPattern = regexp.MustCompile(`^(read|write|execute|delete)$`)
ValidIPv4Pattern = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
```

**Why Validate Everything?**
- ✅ **Security**: Prevent injection attacks
- ✅ **Data Integrity**: Keep bad data out of blockchain
- ✅ **Early Detection**: Fail fast with clear errors
- ✅ **Documentation**: Rules are self-documenting

**Usage Example**:
```go
// Validate device ID
if err := common.ValidateDeviceID(deviceID); err != nil {
    return fmt.Errorf("invalid deviceID: %v", err)
}

// Validate public key
if err := common.ValidatePublicKey(publicKey); err != nil {
    return fmt.Errorf("invalid publicKey: %v", err)
}
```

---

### 3. `ratelimit.go` - Rate Limiting

**Purpose**: Prevent abuse by limiting requests per device

**Structure**:
```go
type RateLimiter struct {
    requestCounts      map[string]*RequestCounter
    bannedDevices      map[string]time.Time
    requestsPerMinute  int
    banDurationMinutes int
    mu                 sync.RWMutex
}

type RequestCounter struct {
    Count          int
    WindowStart    time.Time
    ViolationCount int
}
```

**How It Works**:
1. Track requests per device per minute
2. If limit exceeded → violation recorded
3. After 3 violations → 5-minute ban
4. Window resets every minute

**Configuration**:
```go
const (
    DefaultRequestsPerMinute = 60
    DefaultBanDurationMinutes = 5
)
```

**API**:
```go
NewRateLimiter(reqPerMin, banDurationMin int) *RateLimiter
AllowRequest(deviceID string) (bool, error)
GetDeviceStats(deviceID string) map[string]interface{}
UnbanDevice(deviceID string)
ResetDevice(deviceID string)
GetStats() map[string]interface{}
```

**Usage Example**:
```go
limiter := common.NewRateLimiter(60, 5)

// Check if request allowed
allowed, err := limiter.AllowRequest(deviceID)
if !allowed {
    return fmt.Errorf("rate limit exceeded: %v", err)
}

// Get stats
stats := limiter.GetDeviceStats(deviceID)
log.Printf("Device %s: %d requests", deviceID, stats["requestCount"])
```

---

### 4. `audit.go` - Audit Logging

**Purpose**: Comprehensive logging of all security events

**Event Types**:
```go
// Authentication Events
EventDeviceRegistered
EventDeviceAuthenticated  
EventAuthenticationFailed

// Service Events
EventServiceTicketIssued
EventServiceTicketRevoked

// Access Events
EventAccessGranted
EventAccessDenied
EventSessionCreated

// Security Events
EventRateLimitExceeded
EventValidationFailed
EventSignatureInvalid
```

**Severity Levels**:
```go
SeverityInfo     // Normal operations
SeverityWarning  // Potential issues
SeverityError    // Errors occurred
SeverityCritical // Security incidents
```

**Audit Event Structure**:
```go
type AuditEvent struct {
    EventID       string
    EventType     AuditEventType
    Timestamp     int64
    Severity      AuditSeverity
    Status        string
    DeviceID      string
    ResourceType  string
    ResourceID    string
    Action        string
    Description   string
    IPAddress     string
    Metadata      map[string]string
    ErrorCode     string
    ErrorMessage  string
}
```

**Logger API**:
```go
NewAuditLogger(chaincodeName string) *AuditLogger
LogEvent(eventType, severity, status, desc string) *AuditEvent
LogDeviceRegistration(deviceID, status string) *AuditEvent
LogAuthentication(deviceID string, success bool, reason string) *AuditEvent
LogServiceTicketIssuance(deviceID, serviceID, ticketID, status string) *AuditEvent
LogAccessAttempt(deviceID, serviceID, action string, granted bool, reason, ipAddress string) *AuditEvent
LogSecurityEvent(eventType AuditEventType, deviceID, reason string, severity AuditSeverity) *AuditEvent
LogRateLimitExceeded(deviceID, ipAddress string) *AuditEvent
```

**Usage Example**:
```go
logger := common.NewAuditLogger("AS")

// Log authentication
event := logger.LogAuthentication(deviceID, true, "Authentication successful")
event.SetIPAddress("192.168.1.100")
event.AddMetadata("method", "PKI")

// Log security event
logger.LogRateLimitExceeded(deviceID, ipAddress)
```

**Why Audit Everything?**
- ✅ **Compliance**: Required for many regulations
- ✅ **Forensics**: Investigate security incidents
- ✅ **Monitoring**: Real-time security monitoring
- ✅ **Immutability**: Blockchain provides tamper-proof logs

---

## 🛠️ Technologies & Dependencies

### Go Standard Library

```go
"crypto/rand"      // Cryptographic random number generation
"crypto/sha256"    // SHA-256 hashing
"encoding/base64"  // Base64 encoding/decoding
"encoding/hex"     // Hexadecimal encoding/decoding
"encoding/json"    // JSON serialization
"fmt"              // Formatted I/O
"regexp"           // Regular expressions
"sync"             // Synchronization primitives
"time"             // Time operations
```

### Why These Packages?

**crypto/rand**:
- OS-level entropy source
- FIPS 140-2 compliant on some systems
- Suitable for cryptographic keys

**crypto/sha256**:
- Standard hashing algorithm
- 256-bit output
- Collision-resistant

**sync.RWMutex**:
- Thread-safe data structures
- Multiple readers or single writer
- Essential for rate limiter

---

## 🔐 Security Considerations

### 1. Randomness Quality

```go
// Generate 32 bytes (256 bits) of entropy
bytes := make([]byte, 32)
_, err := rand.Read(bytes)

// Check for errors
if err != nil {
    // NEVER ignore this error
    return nil, fmt.Errorf("PRNG failed: %v", err)
}
```

**Why 256 bits?**
- AES-256 standard
- Resistant to brute force
- Quantum-resistant (for now)

### 2. Validation Edge Cases

```go
// Empty strings
if len(input) == 0 {
    return &ValidationError{Field: "input", Message: "cannot be empty"}
}

// Maximum lengths
if len(input) > MaxLength {
    return &ValidationError{Field: "input", Message: "too long"}
}

// Pattern matching
if !ValidPattern.MatchString(input) {
    return &ValidationError{Field: "input", Message: "invalid format"}
}
```

### 3. Rate Limiting Bypass Prevention

```go
// Use device ID + IP address combo
key := deviceID + ":" + ipAddress

// Time-based windows, not counter-based
if time.Since(window.Start) > time.Minute {
    window.Reset()
}

// Exponential ban duration
banDuration := time.Duration(math.Pow(2, violations)) * time.Minute
```

---

## 📊 Performance Considerations

### Memory Usage

```go
// Rate limiter stores data per device
// Cleanup old entries periodically
func (rl *RateLimiter) cleanup() {
    for deviceID, counter := range rl.requestCounts {
        if time.Since(counter.WindowStart) > time.Hour {
            delete(rl.requestCounts, deviceID)
        }
    }
}
```

### CPU Usage

```go
// Regex compilation done once
var ValidIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Reuse compiled regex, don't recompile each time
```

---

## 🧪 Testing Common Utilities

### Unit Tests

```bash
cd chaincodes/common
go test -v ./...
```

### Example Test

```go
func TestGenerateSecureRandomBytes(t *testing.T) {
    bytes1, _ := GenerateSecureRandomBytes(32)
    bytes2, _ := GenerateSecureRandomBytes(32)
    
    // Should be different each time
    assert.NotEqual(t, bytes1, bytes2)
    
    // Should be correct length
    assert.Equal(t, 32, len(bytes1))
}
```

---

## 🚀 Usage in Chaincodes

### Import Statement

```go
import (
    "github.com/blockchain-auth/common"
)
```

### Example: AS Chaincode Using Common Utils

```go
func (s *ASChaincode) RegisterDevice(
    ctx contractapi.TransactionContextInterface,
    deviceID string,
    publicKey string) error {
    
    // 1. Validate inputs
    if err := common.ValidateDeviceID(deviceID); err != nil {
        return fmt.Errorf("validation failed: %v", err)
    }
    
    if err := common.ValidatePublicKey(publicKey); err != nil {
        return fmt.Errorf("validation failed: %v", err)
    }
    
    // 2. Check rate limit
    allowed, err := rateLimiter.AllowRequest(deviceID)
    if !allowed {
        auditLogger.LogRateLimitExceeded(deviceID, "")
        return fmt.Errorf("rate limit exceeded")
    }
    
    // 3. Process registration
    // ...
    
    // 4. Log success
    auditLogger.LogDeviceRegistration(deviceID, "success")
    
    return nil
}
```

---

## 🔄 Next Steps

### Explore Chaincodes That Use These Utilities:
- 🔐 **AS Chaincode**: [../as-chaincode/README.md](../as-chaincode/README.md)
- 🎫 **TGS Chaincode**: [../tgs-chaincode/README.md](../tgs-chaincode/README.md)
- ✅ **ISV Chaincode**: [../isv-chaincode/README.md](../isv-chaincode/README.md)

### Learn More:
- 📚 **Developer Guide**: [../../DEVELOPER_GUIDE.md](../../DEVELOPER_GUIDE.md)
- 🧪 **Testing Guide**: [../../tests/README.md](../../tests/README.md)

---

📍 **Navigation**: [Chaincodes Overview](../README.md) | [AS Chaincode →](../as-chaincode/README.md)

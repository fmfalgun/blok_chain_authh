# Chaincodes Overview

📍 **Location**: `chaincodes/`
🔗 **Parent Guide**: [Back to Main README](../README.md)
📚 **Related**: [HOW_IT_WORKS.md](../HOW_IT_WORKS.md) | [DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md)

---

## 📋 Overview

This directory contains all the **smart contract code** (chaincodes) that run on the Hyperledger Fabric blockchain. These chaincodes implement the core authentication logic for IoT devices.

## 🎯 What are Chaincodes?

**Definition**: Chaincodes are programs that run on the blockchain and define the business logic for reading and writing to the ledger.

**Think of chaincodes as**:
- **Database + API**: They store data AND provide functions to access it
- **Smart Contracts**: Self-executing contracts with terms directly written into code
- **Microservices**: Independent services that communicate via blockchain

**Key Characteristics**:
- ✅ **Deterministic**: Same input always produces same output
- ✅ **Isolated**: Run in Docker containers
- ✅ **Endorsed**: Require approval from multiple organizations
- ✅ **Immutable**: Code changes require new version deployment

---

## 🏗️ Architecture: Three-Chaincode Design

```
┌─────────────────────────────────────────────────────────┐
│                    IoT Device                            │
└──────────────┬──────────────────────────────────────────┘
               │
               │ ① Register/Authenticate
               ▼
┌─────────────────────────────────────────────────────────┐
│  AS Chaincode (Authentication Server)                    │
│  Organization: Org1                                      │
│  Purpose: Device Registration & TGT Issuance            │
│  📁 Directory: as-chaincode/                            │
└──────────────┬──────────────────────────────────────────┘
               │
               │ ② Request Service Ticket
               ▼
┌─────────────────────────────────────────────────────────┐
│  TGS Chaincode (Ticket Granting Server)                 │
│  Organization: Org2                                      │
│  Purpose: Service Ticket Issuance                       │
│  📁 Directory: tgs-chaincode/                           │
└──────────────┬──────────────────────────────────────────┘
               │
               │ ③ Validate Access
               ▼
┌─────────────────────────────────────────────────────────┐
│  ISV Chaincode (IoT Service Validator)                  │
│  Organization: Org3                                      │
│  Purpose: Access Validation & Session Management        │
│  📁 Directory: isv-chaincode/                           │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  Common Utilities                                        │
│  Shared by all chaincodes                               │
│  📁 Directory: common/                                  │
└─────────────────────────────────────────────────────────┘
```

### Why Three Separate Chaincodes?

**1. Separation of Concerns**
- Each chaincode has ONE primary responsibility
- Easier to understand, test, and maintain
- Clear boundaries between authentication stages

**2. Security Isolation**
- Compromise of one doesn't expose all
- Different organizations control different stages
- Principle of least privilege

**3. Scalability**
- Each chaincode can scale independently
- Deploy on different peers if needed
- Optimize each for its specific workload

**4. Flexibility**
- Upgrade one without touching others
- Different endorsement policies per chaincode
- Mix-and-match for different use cases

---

## 📁 Directory Structure

```
chaincodes/
├── README.md                    ← You are here
│
├── common/                      ← Shared utilities
│   ├── README.md               ← Utility documentation
│   ├── utils.go                ← Crypto, timestamps, IDs
│   ├── validation.go           ← Input validation rules
│   ├── ratelimit.go            ← Rate limiting logic
│   ├── audit.go                ← Audit logging system
│   └── go.mod                  ← Go module definition
│
├── as-chaincode/               ← Authentication Server
│   ├── README.md               ← AS chaincode docs
│   ├── as-chaincode.go         ← Main chaincode code
│   └── go.mod                  ← Dependencies
│
├── tgs-chaincode/              ← Ticket Granting Server
│   ├── README.md               ← TGS chaincode docs
│   ├── tgs-chaincode.go        ← Main chaincode code
│   └── go.mod                  ← Dependencies
│
└── isv-chaincode/              ← Service Validator
    ├── README.md               ← ISV chaincode docs
    ├── isv-chaincode.go        ← Main chaincode code
    └── go.mod                  ← Dependencies
```

---

## 🔍 Chaincode Comparison

| Feature | AS Chaincode | TGS Chaincode | ISV Chaincode |
|---------|--------------|---------------|---------------|
| **Primary Role** | Device Identity | Service Authorization | Access Control |
| **Key Functions** | Register, Authenticate | Issue Tickets, Validate | Validate Access, Track Sessions |
| **Data Stored** | Devices, TGTs | Services, Tickets | Sessions, Access Logs |
| **Endorsement** | Org1 required | Org2 required | Org3 required |
| **Calls Others** | No | Queries AS (TGT validation) | Queries TGS (Ticket validation) |
| **Complexity** | Medium | Medium | High |
| **Lines of Code** | ~300 | ~350 | ~400 |

---

## 🛠️ Technologies Used

### Programming Language: Go 1.21

**Why Go?**
- ✅ **Required by Fabric**: Official language for chaincodes
- ✅ **Fast**: Compiled language, excellent performance
- ✅ **Concurrent**: Built-in goroutines for parallel processing
- ✅ **Type-Safe**: Catches errors at compile time
- ✅ **Great Tooling**: go fmt, go vet, golangci-lint

**Go Features Used**:
```go
// Structs for data modeling
type Device struct {
    DeviceID string `json:"deviceID"`
    Status   string `json:"status"`
}

// Interfaces for contracts
type contractapi.ContractInterface

// JSON encoding/decoding
json.Marshal(device)
json.Unmarshal(data, &device)

// Error handling
if err != nil {
    return fmt.Errorf("operation failed: %v", err)
}
```

### Hyperledger Fabric Contract API

**Package**: `github.com/hyperledger/fabric-contract-api-go`
**Version**: v1.2.1

**What it provides**:
```go
// Transaction context - access to ledger
ctx.GetStub().GetState(key)
ctx.GetStub().PutState(key, value)

// Query capabilities
ctx.GetStub().GetStateByRange(startKey, endKey)

// Events
ctx.GetStub().SetEvent("EventName", payload)

// Cross-chaincode calls
ctx.GetStub().InvokeChaincode("otherCC", args, "channel")
```

**Why use Contract API?**
- Simpler than raw shim API
- Automatic JSON serialization
- Better error handling
- Metadata generation

---

## 🔄 Chaincode Lifecycle

### 1. Development Phase
```bash
# Write Go code
vim chaincodes/as-chaincode/as-chaincode.go

# Test locally
cd chaincodes/as-chaincode
go test ./...

# Build to verify
go build
```

### 2. Package Phase
```bash
# Package for deployment
peer lifecycle chaincode package as.tar.gz \
  --path ./chaincodes/as-chaincode \
  --lang golang \
  --label as_1.0
```

### 3. Install Phase
```bash
# Install on each peer
peer lifecycle chaincode install as.tar.gz

# Get package ID
peer lifecycle chaincode queryinstalled
# Returns: as_1.0:abc123def456...
```

### 4. Approve Phase
```bash
# Each organization approves
peer lifecycle chaincode approveformyorg \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --package-id as_1.0:abc123def456... \
  --sequence 1
```

### 5. Commit Phase
```bash
# Commit to channel (requires majority)
peer lifecycle chaincode commit \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --sequence 1 \
  --peerAddresses peer0.org1:7051 \
  --peerAddresses peer0.org2:9051 \
  --peerAddresses peer0.org3:11051
```

### 6. Invoke Phase
```bash
# Now ready to invoke
peer chaincode invoke \
  -C authchannel \
  -n as \
  -c '{"Args":["RegisterDevice","device001","pubkey","meta"]}'
```

---

## 🔐 Security Features

### 1. Input Validation (common/validation.go)
```go
// All inputs validated before processing
ValidateDeviceID(deviceID)      // Length, characters
ValidatePublicKey(publicKey)    // PEM format, size
ValidateSignature(signature)    // Length, format
ValidateTimestamp(timestamp)    // Within time window
```

### 2. Rate Limiting (common/ratelimit.go)
```go
// Prevent abuse
RateLimiter{
    requestsPerMinute: 60,
    banDurationMinutes: 5,
    // Tracks per device
}
```

### 3. Audit Logging (common/audit.go)
```go
// All operations logged
AuditLogger.LogAuthentication(deviceID, success, reason)
AuditLogger.LogAccessAttempt(deviceID, serviceID, granted, reason)
```

### 4. Cryptographic Security (common/utils.go)
```go
// Secure random generation
GenerateSecureRandomBytes(32)   // Uses crypto/rand
GenerateSecureNonce()           // 256-bit nonces
GenerateSessionKey()            // 256-bit keys
```

---

## 📊 Data Flow

### Complete Transaction Flow
```
1. Client → Peer → AS Chaincode
   ├─ Proposal: RegisterDevice(deviceID, publicKey, metadata)
   ├─ Execution: Validate inputs, check duplicate, create device
   ├─ Read Set: Check if deviceID exists
   └─ Write Set: device_001 → Device{...}

2. Peer → Orderer
   ├─ Submit: Transaction with read/write sets
   └─ Order: Place in block with other transactions

3. Orderer → All Peers
   ├─ Deliver: Block with ordered transactions
   └─ Validate: Check read/write conflicts

4. Peers → Commit
   ├─ Validate: Endorsement signatures, no conflicts
   └─ Commit: Write to world state and blockchain

5. Event → Client
   └─ Notify: DeviceRegistered event emitted
```

---

## 🧪 Testing

### Unit Tests
```bash
cd chaincodes/as-chaincode
go test -v ./...
```

### Integration Tests
```bash
# Test complete flow
cd tests/integration
go test -v ./...
```

### Manual Testing
```bash
# Invoke functions directly
peer chaincode invoke \
  -C authchannel \
  -n as \
  -c '{"Args":["RegisterDevice","test001","key","meta"]}'

# Query results
peer chaincode query \
  -C authchannel \
  -n as \
  -c '{"Args":["GetDevice","test001"]}'
```

---

## 🚀 Quick Start Guide

### 1. Build All Chaincodes
```bash
cd chaincodes/as-chaincode && go build && cd ../..
cd chaincodes/tgs-chaincode && go build && cd ../..
cd chaincodes/isv-chaincode && go build && cd ../..
```

### 2. Deploy All Chaincodes
```bash
# From project root
make deploy-cc
```

### 3. Verify Deployment
```bash
docker exec cli peer lifecycle chaincode querycommitted -C authchannel
```

### 4. Test Functions
```bash
# Register device
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["RegisterDevice","device001","-----BEGIN PUBLIC KEY-----\n...","IoT Sensor"]}'

# Authenticate
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["Authenticate","{\"deviceID\":\"device001\",\"nonce\":\"abc\",\"timestamp\":1672531200,\"signature\":\"xyz\"}"]}'

# Request service ticket
peer chaincode invoke -C authchannel -n tgs \
  -c '{"Args":["IssueServiceTicket","{\"deviceID\":\"device001\",\"tgtID\":\"tgt_123\",\"serviceID\":\"service001\",\"timestamp\":1672531200,\"signature\":\"xyz\"}"]}'
```

---

## 📚 Detailed Documentation

### Chaincode-Specific READMEs

Each chaincode directory has its own detailed README:

1. **[common/README.md](common/README.md)**
   - Shared utilities documentation
   - Usage examples for each utility
   - Design decisions

2. **[as-chaincode/README.md](as-chaincode/README.md)**
   - AS chaincode detailed guide
   - Function-by-function breakdown
   - API reference

3. **[tgs-chaincode/README.md](tgs-chaincode/README.md)**
   - TGS chaincode detailed guide
   - Service ticket flow
   - Cross-chaincode communication

4. **[isv-chaincode/README.md](isv-chaincode/README.md)**
   - ISV chaincode detailed guide
   - Session management
   - Access logging

---

## 🐛 Troubleshooting

### Build Fails
```bash
# Check Go version
go version  # Should be 1.21+

# Tidy dependencies
cd chaincodes/as-chaincode
go mod tidy

# Clear cache
go clean -modcache
```

### Chaincode Won't Install
```bash
# Check peer logs
docker logs peer0.org1.example.com

# Verify package
tar -tzf as.tar.gz

# Check permissions
ls -la chaincodes/as-chaincode
```

### Function Invocation Fails
```bash
# Check chaincode container logs
docker logs $(docker ps -f name=dev-peer0.org1.*as -q)

# Query committed chaincodes
peer lifecycle chaincode querycommitted -C authchannel

# Test with simple query first
peer chaincode query -C authchannel -n as -c '{"Args":["GetAllDevices"]}'
```

---

## 🎯 Best Practices

### Code Organization
```go
// ✅ GOOD: Clear structure
type MyChaincode struct {
    contractapi.Contract
}

func (s *MyChaincode) MyFunction(ctx, param1, param2) error {
    // Validate
    // Process
    // Store
    // Event
    // Return
}

// ❌ BAD: Everything in one function
```

### Error Handling
```go
// ✅ GOOD: Descriptive errors
if err := ValidateInput(input); err != nil {
    return fmt.Errorf("validation failed for %s: %v", input, err)
}

// ❌ BAD: Silent failures
if err := ValidateInput(input); err != nil {
    return nil
}
```

### State Management
```go
// ✅ GOOD: Structured keys
ctx.GetStub().PutState("DEVICE_"+deviceID, data)
ctx.GetStub().PutState("TGT_"+tgtID, data)

// ❌ BAD: Potential collisions
ctx.GetStub().PutState(deviceID, data)
ctx.GetStub().PutState(tgtID, data)
```

---

## 🔄 Next Steps

### Learn More About Specific Chaincodes:
- 📖 **Common Utilities**: [common/README.md](common/README.md)
- 🔐 **AS Chaincode**: [as-chaincode/README.md](as-chaincode/README.md)
- 🎫 **TGS Chaincode**: [tgs-chaincode/README.md](tgs-chaincode/README.md)
- ✅ **ISV Chaincode**: [isv-chaincode/README.md](isv-chaincode/README.md)

### Understand the Network:
- 🌐 **Network Setup**: [../network/README.md](../network/README.md)
- 📊 **Monitoring**: [../monitoring/README.md](../monitoring/README.md)

### Start Developing:
- 📚 **Developer Guide**: [../DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md)
- 🧪 **Testing Guide**: [../tests/README.md](../tests/README.md)

---

📍 **Navigation**: [Main README](../README.md) | [← CI/CD](.github/workflows/README.md) | [Common Utils →](common/README.md)

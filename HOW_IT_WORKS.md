# How the Blockchain Authentication Framework Works

This document explains the internal workings of the blockchain-based authentication system, from high-level architecture to low-level implementation details.

## Table of Contents

1. [System Architecture](#system-architecture)
2. [Authentication Flow](#authentication-flow)
3. [Data Structures](#data-structures)
4. [Chaincode Internals](#chaincode-internals)
5. [Security Mechanisms](#security-mechanisms)
6. [Network Communication](#network-communication)
7. [State Management](#state-management)
8. [Performance Optimization](#performance-optimization)

---

## System Architecture

### Three-Tier Architecture

The system is built on a **Kerberos-inspired** three-tier authentication model:

```
┌─────────────────────────────────────────────────────────────┐
│                     IoT Device (Client)                      │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 1. Register/Authenticate
             ▼
┌─────────────────────────────────────────────────────────────┐
│  AS Chaincode (Authentication Server) - Organization 1      │
│  - Device Registration                                       │
│  - Primary Authentication                                    │
│  - TGT (Ticket Granting Ticket) Issuance                   │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 2. Request Service Ticket
             ▼
┌─────────────────────────────────────────────────────────────┐
│  TGS Chaincode (Ticket Granting Server) - Organization 2    │
│  - Service Registration                                      │
│  - Service Ticket Issuance                                  │
│  - Ticket Validation                                        │
└────────────┬────────────────────────────────────────────────┘
             │
             │ 3. Access Service
             ▼
┌─────────────────────────────────────────────────────────────┐
│  ISV Chaincode (IoT Service Validator) - Organization 3     │
│  - Access Validation                                         │
│  - Session Management                                        │
│  - Access Logging                                           │
└─────────────────────────────────────────────────────────────┘
```

### Why Three Chaincodes?

**Separation of Concerns**: Each chaincode has a distinct responsibility:
- **AS**: Identity management (who you are)
- **TGS**: Service authorization (what you can access)
- **ISV**: Access control enforcement (when and how you access)

**Security**: Compromise of one chaincode doesn't expose the entire system.

**Scalability**: Each chaincode can be deployed on different peers, distributing load.

### Hyperledger Fabric Network Topology

```
┌──────────────────────────────────────────────────────────────┐
│                      Ordering Service                         │
│              (EtcdRaft Consensus - 1 orderer)                 │
└──────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Organization 1│   │ Organization 2│   │ Organization 3│
│     (AS)      │   │     (TGS)     │   │     (ISV)     │
├───────────────┤   ├───────────────┤   ├───────────────┤
│ peer0.org1:   │   │ peer0.org2:   │   │ peer0.org3:   │
│   7051        │   │   9051        │   │   11051       │
│ peer1.org1:   │   │ peer1.org2:   │   │ peer1.org3:   │
│   8051        │   │   10051       │   │   12051       │
│ ca.org1:      │   │ ca.org2:      │   │ ca.org3:      │
│   7054        │   │   8054        │   │   9054        │
└───────────────┘   └───────────────┘   └───────────────┘
```

---

## Authentication Flow

### Phase 1: Device Registration

**Purpose**: Register a new IoT device with the authentication server.

**Flow**:
```
IoT Device                              AS Chaincode
     │                                       │
     │  RegisterDevice(deviceID,             │
     │    publicKey, metadata)               │
     ├──────────────────────────────────────>│
     │                                       │
     │                          ┌────────────┤
     │                          │ Validate   │
     │                          │ - ID length│
     │                          │ - PEM key  │
     │                          │ - No dup   │
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Store to   │
     │                          │ World State│
     │                          └────────────┤
     │                                       │
     │<────── Success / Error ───────────────┤
     │                                       │
```

**What Happens Internally**:

1. **Input Validation**:
   ```go
   // chaincodes/common/validation.go
   ValidateDeviceID(deviceID)     // 3-64 chars, alphanumeric
   ValidatePublicKey(publicKey)   // PEM format, 100-4096 chars
   ValidateMetadata(metadata)     // Max 1024 chars
   ```

2. **Duplicate Check**:
   ```go
   // chaincodes/as-chaincode/as-chaincode.go:32
   existing, err := ctx.GetStub().GetState(deviceID)
   if existing != nil {
       return fmt.Errorf("device %s already exists", deviceID)
   }
   ```

3. **Device Object Creation**:
   ```go
   device := Device{
       DeviceID:         deviceID,
       PublicKey:        publicKey,
       Status:           "active",
       RegistrationTime: getCurrentTimestamp(),
       LastAuthTime:     0,
       Metadata:         metadata,
   }
   ```

4. **State Storage**:
   ```go
   deviceJSON, _ := json.Marshal(device)
   ctx.GetStub().PutState(deviceID, deviceJSON)
   ```

5. **Event Emission**:
   ```go
   ctx.GetStub().SetEvent("DeviceRegistered", []byte(deviceID))
   ```

**Blockchain State After Registration**:
```
Key: "device_001"
Value: {
  "deviceID": "device_001",
  "publicKey": "-----BEGIN PUBLIC KEY-----...",
  "status": "active",
  "registrationTime": 1672531200,
  "lastAuthTime": 0,
  "metadata": "IoT Sensor v1.0"
}
```

### Phase 2: Initial Authentication

**Purpose**: Authenticate device and receive a TGT (Ticket Granting Ticket).

**Flow**:
```
IoT Device                              AS Chaincode
     │                                       │
     │  Authenticate({deviceID, nonce,       │
     │    timestamp, signature})             │
     ├──────────────────────────────────────>│
     │                                       │
     │                          ┌────────────┤
     │                          │ Verify:    │
     │                          │ - Device   │
     │                          │   exists   │
     │                          │ - Active   │
     │                          │ - Timestamp│
     │                          │ - Signature│
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Generate:  │
     │                          │ - TGT ID   │
     │                          │ - Session  │
     │                          │   Key      │
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Store TGT  │
     │                          └────────────┤
     │                                       │
     │<─── {tgtID, sessionKey, expiresAt} ───┤
     │                                       │
```

**What Happens Internally**:

1. **Request Parsing**:
   ```go
   var authReq AuthRequest
   json.Unmarshal([]byte(authRequestJSON), &authReq)
   ```

2. **Device Lookup**:
   ```go
   deviceJSON, _ := ctx.GetStub().GetState(authReq.DeviceID)
   var device Device
   json.Unmarshal(deviceJSON, &device)
   ```

3. **Status Check**:
   ```go
   if device.Status != "active" {
       return "", fmt.Errorf("device is not active")
   }
   ```

4. **Timestamp Validation**:
   ```go
   currentTime := getCurrentTimestamp()
   if authReq.Timestamp < currentTime-300 ||
      authReq.Timestamp > currentTime+300 {
       return "", fmt.Errorf("timestamp is invalid")
   }
   ```

5. **Signature Verification** (placeholder):
   ```go
   // In production, verify signature using device.PublicKey
   // and authReq.Signature against (deviceID + nonce + timestamp)
   ```

6. **Secure Random Generation**:
   ```go
   // chaincodes/common/utils.go
   func GenerateSecureSessionKey() (string, error) {
       key, _ := GenerateSecureRandomBytes(32) // 256 bits
       return base64.StdEncoding.EncodeToString(key), nil
   }
   ```

7. **TGT Creation**:
   ```go
   tgt := TGT{
       TgtID:      generateSecureTgtID(),
       DeviceID:   authReq.DeviceID,
       SessionKey: generateSecureSessionKey(),
       IssuedAt:   getCurrentTimestamp(),
       ExpiresAt:  getCurrentTimestamp() + 3600, // 1 hour
       Status:     "valid",
   }
   ```

8. **TGT Storage**:
   ```go
   tgtJSON, _ := json.Marshal(tgt)
   ctx.GetStub().PutState("TGT_"+tgtID, tgtJSON)
   ```

**Blockchain State After Authentication**:
```
Key: "TGT_tgt_abc123"
Value: {
  "tgtID": "tgt_abc123",
  "deviceID": "device_001",
  "sessionKey": "dGVzdF9zZXNzaW9uX2tleV8xMjM0NTY=",
  "issuedAt": 1672531200,
  "expiresAt": 1672534800,
  "status": "valid"
}
```

### Phase 3: Service Ticket Request

**Purpose**: Request a ticket to access a specific service.

**Flow**:
```
IoT Device                              TGS Chaincode
     │                                       │
     │  IssueServiceTicket({deviceID,        │
     │    tgtID, serviceID, timestamp,       │
     │    signature})                        │
     ├──────────────────────────────────────>│
     │                                       │
     │                          ┌────────────┤
     │                          │ Verify TGT │
     │                          │ (cross-    │
     │                          │  chaincode)│
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Check      │
     │                          │ Service    │
     │                          │ exists &   │
     │                          │ is active  │
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Generate   │
     │                          │ Service    │
     │                          │ Key        │
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Create &   │
     │                          │ Store      │
     │                          │ Ticket     │
     │                          └────────────┤
     │                                       │
     │<──── {ticketID, serviceKey} ──────────┤
     │                                       │
```

**What Happens Internally**:

1. **TGT Validation** (in production, cross-chaincode call):
   ```go
   // This would invoke AS chaincode to validate TGT
   // For now, basic validation
   if len(ticketReq.TgtID) < 5 {
       return "", fmt.Errorf("invalid TGT ID")
   }
   ```

2. **Service Lookup**:
   ```go
   serviceJSON, _ := ctx.GetStub().GetState("SERVICE_" + ticketReq.ServiceID)
   var service Service
   json.Unmarshal(serviceJSON, &service)
   ```

3. **Service Status Check**:
   ```go
   if !service.IsActive {
       return "", fmt.Errorf("service is not active")
   }
   ```

4. **Service Key Generation**:
   ```go
   serviceKey, _ := generateSecureServiceKey()
   ```

5. **Ticket Creation**:
   ```go
   ticket := ServiceTicket{
       TicketID:      generateSecureTicketID(),
       DeviceID:      ticketReq.DeviceID,
       ServiceID:     ticketReq.ServiceID,
       ServiceKey:    serviceKey,
       IssuedAt:      getCurrentTimestamp(),
       ExpiresAt:     getCurrentTimestamp() + 1800, // 30 min
       Status:        "valid",
       UsageCount:    0,
       MaxUsageCount: 10, // Max 10 uses
   }
   ```

### Phase 4: Access Validation

**Purpose**: Validate access request and manage session.

**Flow**:
```
IoT Device                              ISV Chaincode
     │                                       │
     │  ValidateAccess({deviceID,            │
     │    serviceID, ticketID, action,       │
     │    timestamp, ipAddress, signature})  │
     ├──────────────────────────────────────>│
     │                                       │
     │                          ┌────────────┤
     │                          │ Validate   │
     │                          │ Ticket     │
     │                          │ (cross-    │
     │                          │  chaincode)│
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Check      │
     │                          │ Expiry &   │
     │                          │ Usage      │
     │                          │ Limits     │
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Find/      │
     │                          │ Create     │
     │                          │ Session    │
     │                          └────────────┤
     │                                       │
     │                          ┌────────────┤
     │                          │ Log Access │
     │                          └────────────┤
     │                                       │
     │<─── {granted: true, sessionID} ───────┤
     │                                       │
```

---

## Data Structures

### World State Storage

Hyperledger Fabric uses a **key-value store** (LevelDB or CouchDB) for world state:

```
┌────────────────────┬──────────────────────────────────┐
│       Key          │            Value                  │
├────────────────────┼──────────────────────────────────┤
│ device_001         │ Device JSON                       │
│ device_002         │ Device JSON                       │
│ TGT_tgt_abc123     │ TGT JSON                          │
│ TGT_tgt_def456     │ TGT JSON                          │
│ SERVICE_service001 │ Service JSON                      │
│ TICKET_ticket_xyz  │ ServiceTicket JSON                │
│ SESSION_sess_123   │ DeviceSession JSON                │
│ LOG_log_456        │ AccessLog JSON                    │
└────────────────────┴──────────────────────────────────┘
```

### Key Naming Conventions

- **Devices**: `{deviceID}` (e.g., `device_001`)
- **TGTs**: `TGT_{tgtID}` (e.g., `TGT_tgt_abc123`)
- **Services**: `SERVICE_{serviceID}` (e.g., `SERVICE_service001`)
- **Tickets**: `TICKET_{ticketID}` (e.g., `TICKET_ticket_xyz789`)
- **Sessions**: `SESSION_{sessionID}` (e.g., `SESSION_session_def456`)
- **Logs**: `LOG_{logID}` (e.g., `LOG_log_123456`)

**Why Prefixes?**
- **Namespace Isolation**: Prevents key collisions
- **Efficient Queries**: Use `GetStateByRange` to query all items of a type
- **Clear Organization**: Easy to understand what each key contains

---

## Chaincode Internals

### Chaincode Lifecycle

```
┌─────────────────────────────────────────────────────────┐
│  1. Development                                          │
│     - Write Go code                                      │
│     - Implement contractapi.ContractInterface           │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  2. Packaging                                            │
│     - `peer lifecycle chaincode package`                │
│     - Creates .tar.gz with code + metadata              │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  3. Installation                                         │
│     - Install on all peers                              │
│     - `peer lifecycle chaincode install`                │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  4. Approval                                             │
│     - Each org approves chaincode                       │
│     - `peer lifecycle chaincode approveformyorg`        │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  5. Commit                                               │
│     - Commit to channel (requires majority approval)    │
│     - `peer lifecycle chaincode commit`                 │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────┐
│  6. Invoke/Query                                         │
│     - Chaincode ready for transactions                  │
└─────────────────────────────────────────────────────────┘
```

### Transaction Execution Flow

```
Client                  Peer              Orderer           Peer
  │                      │                   │               │
  │  1. Proposal         │                   │               │
  ├─────────────────────>│                   │               │
  │                      │                   │               │
  │                      │ 2. Execute        │               │
  │                      │    Chaincode      │               │
  │                      │    (Read-Only)    │               │
  │                      │                   │               │
  │  3. Proposal         │                   │               │
  │     Response         │                   │               │
  │<─────────────────────┤                   │               │
  │                      │                   │               │
  │  4. Submit           │                   │               │
  │     Transaction      │                   │               │
  ├──────────────────────┼──────────────────>│               │
  │                      │                   │               │
  │                      │                   │ 5. Order      │
  │                      │                   │    Txns       │
  │                      │                   │               │
  │                      │                   │ 6. Create     │
  │                      │                   │    Block      │
  │                      │                   │               │
  │                      │ 7. Deliver Block  │               │
  │                      │<──────────────────┤──────────────>│
  │                      │                   │               │
  │                      │ 8. Validate &     │               │
  │                      │    Commit         │               │
  │                      │                   │               │
  │  9. Event            │                   │               │
  │<─────────────────────┤                   │               │
  │                      │                   │               │
```

**Steps Explained**:

1. **Proposal**: Client sends transaction proposal to endorsing peers
2. **Execute**: Peer executes chaincode in simulation (no state changes)
3. **Proposal Response**: Peer returns read/write set + signature
4. **Submit**: Client sends endorsed transaction to orderer
5. **Order**: Orderer orders transactions from multiple clients
6. **Create Block**: Orderer creates block with ordered transactions
7. **Deliver Block**: Block delivered to all peers
8. **Validate & Commit**: Peers validate and commit to ledger
9. **Event**: Event notification sent to client

---

## Security Mechanisms

### 1. Input Validation

**Implementation**: `chaincodes/common/validation.go`

```go
// Example: Device ID validation
func ValidateDeviceID(deviceID string) error {
    if len(deviceID) < MinIDLength {
        return &ValidationError{
            Field:   "deviceID",
            Message: "too short",
        }
    }

    if !ValidIDPattern.MatchString(deviceID) {
        return &ValidationError{
            Field:   "deviceID",
            Message: "invalid characters",
        }
    }

    return nil
}
```

**Validation Rules**:
- Length constraints (prevent buffer overflow)
- Character whitelisting (prevent injection)
- Format validation (ensure data integrity)

### 2. Rate Limiting

**Implementation**: `chaincodes/common/ratelimit.go`

```go
type RateLimiter struct {
    requestCounts  map[string]*RequestCounter
    bannedDevices  map[string]time.Time
    requestsPerMinute int
    banDurationMinutes int
    mu sync.RWMutex
}

func (rl *RateLimiter) AllowRequest(deviceID string) (bool, error) {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    // Check if banned
    if banExpiry, isBanned := rl.bannedDevices[deviceID]; isBanned {
        if time.Now().Before(banExpiry) {
            return false, fmt.Errorf("device is banned")
        }
        delete(rl.bannedDevices, deviceID)
    }

    // Get or create counter
    counter := rl.getOrCreateCounter(deviceID)

    // Check rate limit
    if counter.Count >= rl.requestsPerMinute {
        counter.ViolationCount++
        if counter.ViolationCount >= 3 {
            // Ban device
            rl.bannedDevices[deviceID] = time.Now().Add(
                time.Duration(rl.banDurationMinutes) * time.Minute)
        }
        return false, fmt.Errorf("rate limit exceeded")
    }

    counter.Count++
    return true, nil
}
```

**How it Works**:
1. Track requests per device per minute
2. Exceed limit → violation recorded
3. Three violations → 5-minute ban
4. Window resets every minute

### 3. Audit Logging

**Implementation**: `chaincodes/common/audit.go`

```go
type AuditEvent struct {
    EventID       string
    EventType     AuditEventType
    Timestamp     int64
    Severity      AuditSeverity
    DeviceID      string
    Action        string
    Description   string
    IPAddress     string
    ErrorMessage  string
    // ... more fields
}

func (al *AuditLogger) LogAuthentication(
    deviceID string,
    success bool,
    reason string) *AuditEvent {

    event := al.LogEvent(
        EventDeviceAuthenticated,
        SeverityInfo,
        "success",
        reason,
    )
    event.DeviceID = deviceID
    return event
}
```

**What Gets Logged**:
- All authentication attempts (success/failure)
- Service ticket requests
- Access validation attempts
- Rate limit violations
- Device registration/revocation
- Configuration changes

---

## Network Communication

### Peer-to-Peer Communication

```
Peer0.Org1 <──── Gossip Protocol ───-> Peer1.Org1
     │                                      │
     │                                      │
     └────────── Block Dissemination ──────┘
     │                                      │
     └────────── State Reconciliation ─────┘
```

**Gossip Protocol**:
- Peers gossip ledger data to each other
- Efficient block dissemination
- State reconciliation for new/offline peers

### Client-Peer Communication

```
Client Application
      │
      │ gRPC + TLS
      │
      ▼
Peer (Endorser)
      │
      │ Execute Chaincode
      │
      ▼
World State (LevelDB/CouchDB)
```

---

## State Management

### World State vs Blockchain

```
┌──────────────────────────────────────────────────────────┐
│               World State (Current State)                 │
│                                                            │
│  device_001 → {status: "active", ...}                     │
│  TGT_abc123 → {status: "valid", ...}                      │
│                                                            │
│  ↑ Read/Write by Chaincode                                │
└──────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────┐
│            Blockchain (Historical Record)                 │
│                                                            │
│  Block N-2  →  Block N-1  →  Block N                      │
│  [Tx1, Tx2]    [Tx3, Tx4]    [Tx5, Tx6]                  │
│                                                            │
│  ↑ Append-Only, Immutable                                 │
└──────────────────────────────────────────────────────────┘
```

**World State**: Current key-value pairs (latest state)
**Blockchain**: Complete transaction history (audit trail)

---

## Performance Optimization

### 1. Minimize State Reads

```go
// BAD: Multiple reads
device1 := GetState("device_001")
device2 := GetState("device_002")
device3 := GetState("device_003")

// BETTER: Batch read with GetStateByRange
iterator := GetStateByRange("device_", "device_~")
```

### 2. Use Composite Keys

```go
// Create composite key
key, _ := ctx.GetStub().CreateCompositeKey(
    "device~timestamp",
    []string{deviceID, timestamp})

// Query by prefix
iterator, _ := ctx.GetStub().GetStateByPartialCompositeKey(
    "device~timestamp",
    []string{deviceID})
```

### 3. Pagination

```go
// Query with pagination
queryString := `{"selector":{"deviceID":"device_001"}}`
iterator, metadata := ctx.GetStub().GetQueryResultWithPagination(
    queryString,
    pageSize,
    bookmark)
```

---

## Summary

This system implements a **secure, scalable, and auditable** authentication framework for IoT devices using blockchain technology. Key takeaways:

1. **Three-tier architecture** separates concerns and enhances security
2. **Blockchain provides** immutability, transparency, and decentralization
3. **Security layers** include validation, rate limiting, and audit logging
4. **Fabric's endorsement policy** ensures transaction integrity
5. **World state** provides efficient current state access
6. **Blockchain** provides complete historical audit trail

The next document, [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md), provides in-depth technical details for developers implementing or extending this system.

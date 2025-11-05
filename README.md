# Blockchain Authentication Framework for IoT Devices

A secure, production-ready blockchain-based authentication system for IoT devices built on Hyperledger Fabric 2.5.5, implementing a Kerberos-inspired three-tier authentication architecture.

## 🎯 Overview

This framework provides a decentralized authentication and access control system for IoT devices using three specialized chaincodes:
- **AS (Authentication Server)**: Device registration and initial authentication
- **TGS (Ticket Granting Server)**: Service ticket issuance
- **ISV (IoT Service Validator)**: Access validation and session management

## ✨ Key Features

- 🔐 **Secure Authentication**: PKI-based device authentication with TGT (Ticket Granting Tickets)
- 🎫 **Service Tickets**: Fine-grained access control for different services
- 📊 **Session Management**: Active session tracking with timeout handling
- 🛡️ **Security Hardening**: Rate limiting, input validation, and comprehensive audit logging
- 📈 **Monitoring**: Prometheus metrics, Grafana dashboards, and alerting
- 🔄 **CI/CD**: Automated testing, security scanning, and deployment pipelines
- 📝 **Audit Trail**: Immutable blockchain-based logging of all authentication events

## 📋 Prerequisites

### Required Software
- **Docker**: 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose**: 1.29+ ([Install Docker Compose](https://docs.docker.com/compose/install/))
- **Go**: 1.21+ ([Install Go](https://go.dev/doc/install))
- **Node.js**: 16+ (optional, for SDK) ([Install Node.js](https://nodejs.org/))
- **Make**: GNU Make ([Usually pre-installed on Linux/Mac](https://www.gnu.org/software/make/))

### System Requirements
- **CPU**: 4+ cores recommended
- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 20GB free space
- **OS**: Linux (Ubuntu 20.04+), macOS (10.15+), or WSL2 on Windows

### Download Hyperledger Fabric Binaries

```bash
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.5 1.5.5
export PATH=$PATH:$(pwd)/fabric-samples/bin
export FABRIC_CFG_PATH=$(pwd)/fabric-samples/config
```

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/blok_chain_authh.git
cd blok_chain_authh
```

### 2. Install Dependencies

```bash
# Install Go dependencies for all chaincodes
make install-deps

# Verify installation
go version
docker --version
docker-compose --version
```

### 3. Start the Network

```bash
# Start the Hyperledger Fabric network
make network-up

# Create the authentication channel
make channel-create

# This will:
# - Generate crypto material (certificates, keys)
# - Start orderer and peer containers (11 containers total)
# - Create the 'authchannel' channel
# - Join all peers to the channel
```

### 4. Deploy Chaincodes

```bash
# Deploy all three chaincodes
make deploy-cc

# This deploys:
# - AS Chaincode (Authentication Server)
# - TGS Chaincode (Ticket Granting Server)
# - ISV Chaincode (IoT Service Validator)
```

### 5. Verify Deployment

```bash
# Check if all containers are running
docker ps

# Verify channel status
make verify

# Check chaincode is committed
docker exec cli peer lifecycle chaincode querycommitted -C authchannel
```

### 6. Start Monitoring (Optional)

```bash
# Start Prometheus, Grafana, and Alertmanager
make monitoring-up

# Access dashboards:
# - Grafana: http://localhost:3000 (admin/admin)
# - Prometheus: http://localhost:9090
# - Alertmanager: http://localhost:9093
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Run Specific Test Types

```bash
# Unit tests only
make test-unit

# Integration tests
make test-integration

# Performance tests
make test-performance
```

## 📖 Usage Examples

### Register a Device

```bash
# Using peer CLI
docker exec cli peer chaincode invoke \
  -C authchannel \
  -n as \
  -c '{"Args":["RegisterDevice","device_001","-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...","IoT Sensor v1.0"]}'
```

### Authenticate a Device

```bash
# Prepare authentication request
AUTH_REQUEST='{
  "deviceID": "device_001",
  "nonce": "secure_random_nonce_base64",
  "timestamp": 1672531200,
  "signature": "device_signature_base64"
}'

# Invoke authentication
docker exec cli peer chaincode invoke \
  -C authchannel \
  -n as \
  -c "{\"Args\":[\"Authenticate\",\"$AUTH_REQUEST\"]}"

# Returns: TGT ID, Session Key, and Expiration
```

### Request Service Ticket

```bash
# Prepare ticket request
TICKET_REQUEST='{
  "deviceID": "device_001",
  "tgtID": "tgt_abc123",
  "serviceID": "service001",
  "timestamp": 1672531200,
  "signature": "signed_request"
}'

# Invoke ticket issuance
docker exec cli peer chaincode invoke \
  -C authchannel \
  -n tgs \
  -c "{\"Args\":[\"IssueServiceTicket\",\"$TICKET_REQUEST\"]}"
```

### Validate Access

```bash
# Prepare access request
ACCESS_REQUEST='{
  "deviceID": "device_001",
  "serviceID": "service001",
  "ticketID": "ticket_xyz789",
  "action": "read",
  "timestamp": 1672531200,
  "ipAddress": "192.168.1.100",
  "userAgent": "IoT-Device/1.0",
  "signature": "signed_access_request"
}'

# Validate access
docker exec cli peer chaincode invoke \
  -C authchannel \
  -n isv \
  -c "{\"Args\":[\"ValidateAccess\",\"$ACCESS_REQUEST\"]}"
```

## 🏗️ Project Structure

```
blok_chain_authh/
├── .github/
│   └── workflows/          # CI/CD pipelines
├── chaincodes/
│   ├── common/             # Shared utilities
│   ├── as-chaincode/       # Authentication Server
│   ├── tgs-chaincode/      # Ticket Granting Server
│   └── isv-chaincode/      # Service Validator
├── network/
│   ├── config/             # Network configuration
│   └── scripts/            # Management scripts
├── monitoring/             # Prometheus, Grafana, Alertmanager
├── tests/                  # Unit, integration, performance tests
├── docs/                   # Documentation
└── Makefile                # Build automation
```

## 🛠️ Available Make Commands

```bash
make help                # Display all available commands
make install-deps        # Install dependencies
make network-up          # Start Fabric network
make network-down        # Stop network and clean up
make channel-create      # Create authentication channel
make deploy-cc           # Deploy all chaincodes
make test                # Run all tests
make monitoring-up       # Start monitoring stack
make clean               # Clean all artifacts
make restart             # Restart entire network
make logs                # Show container logs
make verify              # Verify network status
```

## 📚 Documentation

### Quick Reference
- **[HOW_IT_WORKS.md](HOW_IT_WORKS.md)** - Detailed explanation of project internals
- **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** - In-depth technical guide
- **[Architecture](docs/architecture/authentication-flow.md)** - System design
- **[API Reference](docs/api/chaincode-api.md)** - Complete API docs
- **[Production Deployment](docs/deployment/PRODUCTION_DEPLOYMENT.md)** - Deploy guide
- **[Troubleshooting](docs/troubleshooting/common-issues.md)** - Common issues

### 📖 Complete Documentation Roadmap

This repository has comprehensive documentation in every directory. Follow this path to gain complete understanding:

#### Level 1: Getting Started (You Are Here!)
**Current**: [Main README](README.md) - Overview, quick start, basic usage

**What you learned**: How to install, start the network, deploy chaincodes, and run basic operations

**Next Step**: Choose your path based on your goal:
- **Want to understand how it works?** → Go to [HOW_IT_WORKS.md](HOW_IT_WORKS.md)
- **Want to develop/extend the system?** → Go to [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)
- **Want to explore the code structure?** → Continue to Level 2 below

---

#### Level 2: Project Understanding

##### Path A: How It Works (Conceptual Understanding)
1. **[HOW_IT_WORKS.md](HOW_IT_WORKS.md)** (10,000 words)
   - System architecture and authentication flow
   - Data structures and chaincode internals
   - Security mechanisms and state management
   - **Time**: 30-45 minutes
   - **Next**: Understanding the implementation → [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)

##### Path B: Developer Guide (Technical Deep-Dive)
2. **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** (12,000 words)
   - Function-by-function code breakdown
   - Design patterns and best practices
   - Testing strategies and debugging
   - **Time**: 45-60 minutes
   - **Next**: Explore specific components → Level 3

---

#### Level 3: Component Deep-Dive

After understanding the big picture, dive into specific components:

##### 3A: Chaincodes (The Core Logic)
**Start**: [chaincodes/README.md](chaincodes/README.md) - Overview of all three chaincodes

Then explore each chaincode in order (following the authentication flow):

1. **[chaincodes/as-chaincode/README.md](chaincodes/as-chaincode/README.md)** (~5,000 words)
   - Authentication Server - Device registration and TGT issuance
   - Functions: RegisterDevice, Authenticate, GetDevice, RevokeDevice
   - Security features, integration patterns
   - **Time**: 20-30 minutes

2. **[chaincodes/tgs-chaincode/README.md](chaincodes/tgs-chaincode/README.md)** (~4,800 words)
   - Ticket Granting Server - Service ticket issuance
   - Functions: RegisterService, IssueServiceTicket, ValidateServiceTicket
   - Cross-chaincode communication with AS
   - **Time**: 20-30 minutes

3. **[chaincodes/isv-chaincode/README.md](chaincodes/isv-chaincode/README.md)** (~5,000 words)
   - IoT Service Validator - Access validation and session management
   - Functions: ValidateAccess, TerminateSession, GetAccessLogs
   - Audit logging and session tracking
   - **Time**: 20-30 minutes

4. **[chaincodes/common/README.md](chaincodes/common/README.md)** (~2,800 words)
   - Shared utilities used by all chaincodes
   - Validation, rate limiting, audit logging, cryptographic utilities
   - **Time**: 15-20 minutes

**Total Chaincodes Time**: ~2 hours

---

##### 3B: Network Infrastructure
**Start**: [network/README.md](network/README.md) - Network overview

Then explore network components:

1. **[network/config/README.md](network/config/README.md)** (~6,000 words)
   - `crypto-config.yaml` - PKI structure and certificate generation
   - `configtx.yaml` - Channel configuration and consensus
   - `docker-compose-network.yaml` - Container orchestration
   - **Time**: 30-40 minutes
   - **What you'll learn**: How to configure Fabric networks, what each config parameter means

2. **[network/scripts/README.md](network/scripts/README.md)** (~6,000 words)
   - `network.sh` - Network lifecycle automation
   - `deploy-chaincode.sh` - Chaincode deployment (Fabric 2.x lifecycle)
   - `verify-channel.sh` - Health checks
   - **Time**: 30-40 minutes
   - **What you'll learn**: How to automate network operations, chaincode deployment process

**Total Network Time**: ~1-1.5 hours

---

##### 3C: CI/CD and Automation
**[.github/workflows/README.md](.github/workflows/README.md)** (~2,800 words)
- Automated testing workflow (test.yml)
- Security scanning pipeline (security.yml)
- Deployment automation (deploy.yml)
- **Time**: 15-20 minutes
- **What you'll learn**: How CI/CD is set up, GitHub Actions workflows

---

##### 3D: Testing
**[tests/README.md](tests/README.md)** (~800 words)
- Unit tests, integration tests, performance tests
- How to run and write tests
- **Time**: 10 minutes
- **What you'll learn**: Testing strategies for blockchain applications

---

##### 3E: Monitoring
**[monitoring/README.md](monitoring/README.md)** (~1,000 words)
- Prometheus metrics collection
- Grafana dashboards
- Alertmanager configuration
- **Time**: 10-15 minutes
- **What you'll learn**: How to monitor blockchain networks

---

##### 3F: Documentation Hub
**[docs/README.md](docs/README.md)** (~500 words)
- Central hub linking to all specialized docs
- Architecture diagrams, API references, deployment guides
- **Time**: 5 minutes
- **What you'll learn**: Where to find specific documentation

---

#### Level 4: Specialized Topics

Based on your needs, explore:

**For Deployment**:
- [docs/deployment/PRODUCTION_DEPLOYMENT.md](docs/deployment/PRODUCTION_DEPLOYMENT.md)
- Learn: AWS/Azure/GCP deployment, Kubernetes, production hardening

**For API Reference**:
- [docs/api/chaincode-api.md](docs/api/chaincode-api.md)
- Learn: Complete API documentation for all chaincode functions

**For Architecture**:
- [docs/architecture/authentication-flow.md](docs/architecture/authentication-flow.md)
- Learn: System design, sequence diagrams, data flow

**For Troubleshooting**:
- [docs/troubleshooting/common-issues.md](docs/troubleshooting/common-issues.md)
- Learn: Common problems and solutions

---

### 🎓 Suggested Learning Paths

#### Path 1: "I Want to Use This System"
**Time**: ~2 hours
```
Main README (15 min)
↓
HOW_IT_WORKS.md (45 min)
↓
chaincodes/README.md (10 min)
↓
chaincodes/as-chaincode/README.md (30 min)
↓
chaincodes/tgs-chaincode/README.md (30 min)
↓
chaincodes/isv-chaincode/README.md (30 min)
```
**Outcome**: Understand how to use the authentication system, what each function does

---

#### Path 2: "I Want to Deploy This to Production"
**Time**: ~3 hours
```
Main README (15 min)
↓
HOW_IT_WORKS.md (45 min)
↓
network/config/README.md (40 min)
↓
network/scripts/README.md (40 min)
↓
docs/deployment/PRODUCTION_DEPLOYMENT.md (60 min)
↓
monitoring/README.md (15 min)
```
**Outcome**: Know how to configure, deploy, and monitor in production

---

#### Path 3: "I Want to Develop/Extend This System"
**Time**: ~4-5 hours
```
Main README (15 min)
↓
DEVELOPER_GUIDE.md (60 min)
↓
All chaincode READMEs (2 hours)
↓
chaincodes/common/README.md (20 min)
↓
network/config/README.md (40 min)
↓
network/scripts/README.md (40 min)
↓
tests/README.md (10 min)
↓
.github/workflows/README.md (20 min)
```
**Outcome**: Deep understanding of code, able to add features and fix bugs

---

#### Path 4: "I Want to Build My Own Fabric Project"
**Time**: ~5-6 hours (complete path)
```
Complete Path 3 above
↓
docs/architecture/authentication-flow.md (30 min)
↓
docs/api/chaincode-api.md (30 min)
↓
docs/deployment/PRODUCTION_DEPLOYMENT.md (60 min)
```
**Outcome**: Complete understanding, able to replicate Fabric setup for own use case

---

### 📊 Documentation Statistics

- **Total Documentation**: ~50,000 words
- **Number of READMEs**: 15 interconnected guides
- **Code Examples**: 100+ practical examples
- **Coverage**: Every directory explained with why/what/how

### 🔗 Quick Navigation

**By Topic**:
- **Blockchain Basics**: [HOW_IT_WORKS.md](HOW_IT_WORKS.md) → [network/config/README.md](network/config/README.md)
- **Chaincode Development**: [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) → [chaincodes/](chaincodes/README.md)
- **DevOps**: [network/scripts/README.md](network/scripts/README.md) → [.github/workflows/README.md](.github/workflows/README.md)
- **Production**: [docs/deployment/PRODUCTION_DEPLOYMENT.md](docs/deployment/PRODUCTION_DEPLOYMENT.md)

**By Role**:
- **Developers**: DEVELOPER_GUIDE.md → chaincodes/* → tests/README.md
- **DevOps Engineers**: network/* → monitoring/README.md → .github/workflows/README.md
- **Architects**: HOW_IT_WORKS.md → docs/architecture/* → network/config/README.md
- **End Users**: Main README → HOW_IT_WORKS.md → docs/api/chaincode-api.md

## 🔐 Security Features

- **Input Validation**: Length constraints, regex patterns, timestamp validation
- **Rate Limiting**: 60 req/min per device, automatic banning
- **Audit Logging**: All events logged to immutable blockchain
- **Cryptographic Security**: crypto/rand, 256-bit keys, PKI authentication

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/your-org/blok_chain_authh/issues)
- **Email**: support@example.com

---

**Built with Hyperledger Fabric 2.5.5** | **Secure by Design** | **Production Ready**

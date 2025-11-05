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

- **[HOW_IT_WORKS.md](HOW_IT_WORKS.md)** - Detailed explanation of project internals
- **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** - In-depth technical guide
- **[Architecture](docs/architecture/authentication-flow.md)** - System design
- **[API Reference](docs/api/chaincode-api.md)** - Complete API docs
- **[Production Deployment](docs/deployment/PRODUCTION_DEPLOYMENT.md)** - Deploy guide
- **[Troubleshooting](docs/troubleshooting/common-issues.md)** - Common issues

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

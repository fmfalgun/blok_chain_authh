# Blockchain Authentication Framework (BAF)

This README provides instructions for setting up, running, and testing the Blockchain Authentication Framework (BAF) for Hyperledger Fabric.

## Overview

The Blockchain Authentication Framework implements a Kerberos-like authentication system using Hyperledger Fabric blockchain technology. It consists of three main components:

1. **Authentication Server (AS)** - Registers clients and issues Ticket Granting Tickets (TGTs)
2. **Ticket Granting Server (TGS)** - Validates TGTs and issues service tickets
3. **IoT Service Validator (ISV)** - Validates service tickets and grants access to IoT devices

The system provides secure authentication and authorization for IoT devices using blockchain technology for transparency, immutability, and decentralization.

## Project Structure

```
/home/fm/projects/blok_chain_authh/
├── start-network.sh           # Script to start the Fabric network
├── docker-compose.yaml        # Docker Compose configuration
├── configtx.yaml              # Fabric channel configuration
├── crypto-config.yaml         # Crypto material configuration
├── chaincodes/                # Chaincode (smart contract) implementations
│   ├── as-chaincode-fixed-v4/  # Authentication Server chaincode
│   ├── tgs-chaincode-fixed-v4/ # Ticket Granting Server chaincode
│   └── isv-chaincode-fixed-v4/ # IoT Service Validator chaincode
├── blockchain-auth-framework/ # Client SDK and test scripts
│   ├── auth-framework.js      # Main authentication framework SDK
│   ├── auth-cli.sh            # CLI interface for authentication
│   ├── simple-auth.sh         # Simplified interface for testing
│   ├── test-auth.js           # Test utility for cryptographic operations
│   ├── test-authentication-flow.sh # Test for complete authentication flow
│   ├── test-rsa-keys.sh       # Test for RSA key operations
│   ├── setup-test-environment.sh # Prepares environment for testing
│   └── run-all-tests.sh       # Runs all tests and generates report
└── organizations/             # Generated crypto material for organizations
```

## Prerequisites

- Docker and Docker Compose
- Node.js and NPM
- Hyperledger Fabric binaries v2.2+
- Bash shell

## Setup Instructions

### 1. Start the Hyperledger Fabric Network

```bash
cd /home/fm/projects/blok_chain_authh
./start-network.sh
```

This script will:
- Generate crypto material for organizations
- Create channel artifacts
- Start the network containers
- Create and join the channel
- Deploy the chaincodes (smart contracts)

### 2. Set Up Client Environment

```bash
cd /home/fm/projects/blok_chain_authh/blockchain-auth-framework
npm install
node enrollAdmin.js
```

### 3. Make Test Scripts Executable

```bash
cd /home/fm/projects/blok_chain_authh/blockchain-auth-framework
chmod +x make-executable.sh
./make-executable.sh
```

## Testing the Authentication Framework

The framework includes several test scripts to verify its functionality:

### Run All Tests

```bash
cd /home/fm/projects/blok_chain_authh/blockchain-auth-framework
./run-all-tests.sh
```

This will run all tests and generate a comprehensive test report.

### Individual Tests

1. **Authentication Flow Test**:
   ```bash
   ./test-authentication-flow.sh
   ```

2. **RSA Key Test**:
   ```bash
   ./test-rsa-keys.sh
   ```

### Test Results

Test results are saved in a timestamped directory:
```
test-results-YYYYMMDDhhmmss/
├── test-summary.md           # Summary of all test results
├── test-authentication-flow-results.log # Detailed log of authentication flow test
├── test-rsa-keys-results.log # Detailed log of RSA key test
└── *-status.txt              # Status file for each test (PASS/FAIL)
```

## Using the Authentication System

### Client Registration

```bash
# Using CLI
./auth-cli.sh register-client client1

# Using Node.js SDK
node auth-framework.js register-client admin client1
```

### Device Registration

```bash
# Using CLI
./auth-cli.sh register-device device1 "temperature,humidity"

# Using Node.js SDK
node auth-framework.js register-device admin device1 temperature humidity
```

### Authentication Process

```bash
# Using Node.js SDK
node auth-framework.js authenticate admin client1 device1
```

This performs the complete authentication flow:
1. Get TGT from Authentication Server
2. Get Service Ticket from Ticket Granting Server
3. Authenticate with IoT Service Validator
4. Establish a session with the IoT device

### Accessing Device Data

```bash
# Using Node.js SDK
node auth-framework.js get-device-data admin client1 device1
```

### Closing the Session

```bash
# Using Node.js SDK
node auth-framework.js close-session admin client1 device1
```

## Troubleshooting

### Network Issues

If the network is not functioning properly, try:

```bash
# Stop the network
cd /home/fm/projects/blok_chain_authh
docker-compose down

# Remove old data
rm -rf organizations/peerOrganizations
rm -rf organizations/ordererOrganizations
rm -rf channel-artifacts

# Restart the network
./start-network.sh
```

### Client SDK Issues

If the client SDK is not functioning properly, try:

```bash
# Reset client environment
cd /home/fm/projects/blok_chain_authh/blockchain-auth-framework
rm -rf wallet/*.id
rm -f *-private.pem
rm -f *-tgt.json
rm -f *-serviceticket-*.json
rm -f *-session-*.txt

# Re-enroll admin
node enrollAdmin.js
```

## Security Considerations

- Private keys should be stored securely
- The system uses RSA encryption for authentication
- Session keys have a limited lifetime (1 hour by default)
- Nonce challenges prevent replay attacks

## Additional Documentation

For more detailed information, refer to:

- [Blockchain Authentication Framework Documentation](./blockchain-auth-framework/blockchain-auth-doc.md)
- [Test Plan](./blockchain-auth-framework/blockchain-auth-test-plan.md)

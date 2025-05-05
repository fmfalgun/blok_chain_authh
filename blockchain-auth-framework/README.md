# Blockchain Authentication Framework

This directory contains the client SDK and testing tools for the Blockchain Authentication Framework (BAF) built on Hyperledger Fabric.

## Quick Start

To quickly set up the environment and make all scripts executable:

```bash
chmod +x quick-setup.sh
./quick-setup.sh
```

This will make all scripts executable, check the network status, and set up the test environment.

## Authentication Flow

The authentication flow consists of the following steps:

1. **Client Registration**: Register a client with the Authentication Server (AS)
2. **Device Registration**: Register an IoT device with the IoT Service Validator (ISV)
3. **Get TGT**: Client obtains a Ticket Granting Ticket from AS
4. **Get Service Ticket**: Client uses TGT to get a Service Ticket from Ticket Granting Server (TGS)
5. **Access Device**: Client uses Service Ticket to access the IoT device through ISV
6. **Close Session**: Client closes the session when done

## Available Scripts

### Testing Scripts

- `test-authentication-flow.sh`: Tests the complete authentication flow
- `test-rsa-keys.sh`: Tests RSA key generation and validation
- `run-all-tests.sh`: Runs all tests and generates a comprehensive report
- `check-network-status.sh`: Checks the status of the Hyperledger Fabric network
- `setup-test-environment.sh`: Sets up the test environment

### Utility Scripts

- `auth-framework.js`: Node.js SDK for the authentication framework
- `auth-cli.sh`: CLI interface for authentication operations
- `simple-auth.sh`: Simplified interface for authentication operations
- `make-executable.sh`: Makes all scripts executable
- `quick-setup.sh`: Quick setup script for the authentication framework

## Using the SDK

### Client Registration

```javascript
const authFramework = require('./auth-framework');
authFramework.registerClient('admin', 'client1');
```

Or using CLI:
```bash
./auth-cli.sh register-client client1
```

### Device Registration

```javascript
const authFramework = require('./auth-framework');
authFramework.registerIoTDevice('admin', 'device1', ['temperature', 'humidity']);
```

Or using CLI:
```bash
./auth-cli.sh register-device device1 "temperature,humidity"
```

### Complete Authentication

```javascript
const authFramework = require('./auth-framework');

// Step 1: Get TGT from AS
const tgt = await authFramework.getTGT('admin', 'client1');

// Step 2: Get Service Ticket from TGS
const serviceTicket = await authFramework.getServiceTicket('admin', 'client1', 'iotservice1');

// Step 3: Access IoT device through ISV
const accessResult = await authFramework.accessIoTDevice('admin', 'client1', 'device1');

// Step 4: Get device data
const deviceData = await authFramework.getIoTDeviceData('admin', 'client1', 'device1');

// Step 5: Close session
await authFramework.closeSession('admin', 'client1', 'device1');
```

Or using CLI:
```bash
./auth-cli.sh authenticate client1 device1
```

## Testing

### Running All Tests

```bash
./run-all-tests.sh
```

This will run all tests and generate a test report in a timestamped directory.

### Running Individual Tests

```bash
./test-authentication-flow.sh
./test-rsa-keys.sh
```

## Network Management

### Checking Network Status

```bash
./check-network-status.sh
```

This will check if the Hyperledger Fabric network is running properly.

### Restarting the Network

If the network is not running properly, you can restart it using:

```bash
cd /home/fm/projects/blok_chain_authh
./start-network.sh
```

## Troubleshooting

### Client SDK Issues

If the client SDK is not functioning properly:

```bash
# Reset client environment
rm -rf wallet/*.id
rm -f *-private.pem
rm -f *-tgt.json
rm -f *-serviceticket-*.json
rm -f *-session-*.txt

# Re-enroll admin
node enrollAdmin.js
```

### Network Issues

If the network is not functioning properly:

```bash
# From the project root directory
docker-compose down
rm -rf organizations/peerOrganizations
rm -rf organizations/ordererOrganizations
rm -rf channel-artifacts
./start-network.sh
```

## Documentation

For more detailed information, refer to:
- [Blockchain Authentication Framework Documentation](../README-BAF.md)
- [Test Plan](./blockchain-auth-test-plan.md)

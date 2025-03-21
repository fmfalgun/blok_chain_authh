# Authentication Framework for Hyperledger Fabric (v3)

A comprehensive, modular Go implementation of a Kerberos-like authentication framework for IoT devices on Hyperledger Fabric.

## Overview

This framework provides secure authentication and access control for IoT devices using a Kerberos-inspired protocol on Hyperledger Fabric. It consists of three main components:

1. **Authentication Server (AS)** - Validates client identity and issues Ticket Granting Tickets
2. **Ticket Granting Server (TGS)** - Issues service tickets for accessing specific services
3. **IoT Service Validator (ISV)** - Manages IoT devices and validates service tickets

## Project Structure

```
v3/
├── cmd/                  # Command-line interface
│   └── authcli/          # Authentication CLI
├── config/               # Configuration files
├── internal/             # Internal packages
│   ├── auth/             # Authentication logic
│   ├── crypto/           # Cryptographic operations
│   └── fabric/           # Fabric network interaction
├── pkg/                  # Public packages
│   └── logger/           # Logging utility
├── scripts/              # Utility scripts
├── Makefile              # Build and execution targets
├── go.mod                # Go module dependencies
└── README.md             # This documentation
```

## Authentication Flow

The authentication flow follows these steps:

1. **Client Registration** - Clients register with the AS by generating RSA key pairs
2. **Device Registration** - IoT devices register with the ISV, specifying their capabilities
3. **Authentication** - Clients authenticate with the AS to get a Ticket Granting Ticket (TGT)
4. **Service Request** - Clients use the TGT to request a Service Ticket from the TGS
5. **Device Access** - Clients use the Service Ticket to access IoT devices through the ISV

## Prerequisites

- Go 1.18 or higher
- Hyperledger Fabric network with deployed chaincodes (AS, TGS, ISV)
- Fabric certificates and credentials for a user with chaincode access

## Installation

1. Clone this repository:
   ```bash
   git clone <repository-url>
   cd v3
   ```

2. Set up the environment:
   ```bash
   ./scripts/setup.sh
   ```

3. Initialize the wallet with your Fabric identity:
   ```bash
   ./scripts/init-wallet.sh
   ```

4. Build the CLI:
   ```bash
   make build
   ```

## Usage

The framework provides a command-line interface for all operations:

### Basic Authentication Flow

```bash
# Register a client
bin/authcli register-client --client-id client1

# Register a device
bin/authcli register-device --device-id device1 --capabilities temperature,humidity

# Authenticate client
bin/authcli authenticate --client-id client1 --device-id device1

# Access device
bin/authcli access-device --client-id client1 --device-id device1

# Get device data
bin/authcli get-device-data --device-id device1

# Close session
bin/authcli close-session --client-id client1 --device-id device1
```

### Simplified Flow with Make

```bash
# Run complete authentication flow
make auth-flow

# Individual steps
make register-client
make register-device
make authenticate
make access-device
make get-device-data
make close-session
```

## Configuration

The framework uses a Fabric connection profile for network configuration:

- `config/connection-profile.json` - Connection profile for the Fabric network

## Development

### Adding New Features

1. Implement the feature in the appropriate module
2. Add any new CLI commands to `cmd/authcli/main.go`
3. Update the Makefile with new targets if needed
4. Test the feature with the test network

### Running Tests

```bash
make test
```

## Troubleshooting

### Common Issues

1. **Authentication Fails**
   - Check that your client is registered and keys are correctly generated
   - Ensure your Fabric network connection profile is correct
   - Verify the AS, TGS, and ISV chaincodes are deployed correctly

2. **Wallet Initialization Fails**
   - Make sure you have valid Fabric certificates
   - Check the certificate paths in the connection profile

3. **RSA Key Compatibility**
   - Ensure your keys are in the correct format (PKCS#1 for private, PKIX for public)
   - Use debug tools in the crypto package to test key operations

## License

[MIT License](LICENSE)

## Acknowledgments

Based on the foundations of versions 1 and 2 of the authentication framework, with improvements for modularity, error handling, and production readiness.

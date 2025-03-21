# Go Authentication Framework for Hyperledger Fabric

This project provides a Go-based authentication framework for Hyperledger Fabric with Kerberos-like authentication for IoT devices.

## Quick Start

```bash
# Generate keys for a client and an IoT device
go run standalone-auth-framework.go generate-keys client1
go run standalone-auth-framework.go generate-keys device1

# Simulate authentication (for testing)
go run standalone-auth-framework.go simulate-auth client1 test-nonce-123

# Use the wrapper script for Fabric integration
./fabric-wrapper.sh register-client client1
./fabric-wrapper.sh register-device device1 temperature humidity
./fabric-wrapper.sh authenticate client1 device1
```

## Components

1. **Standalone Authentication Framework**: `standalone-auth-framework.go`
   - Provides key generation in PKCS#1 format compatible with Go chaincodes
   - Simulates authentication flow for testing
   - Debugging utilities for RSA operations

2. **Fabric Wrapper Script**: `fabric-wrapper.sh`
   - Integrates the standalone framework with your Fabric network
   - Orchestrates the full authentication workflow
   - Provides a unified command-line interface

3. **Integration Guide**: `integration-guide.md`
   - Detailed instructions for integrating Go-generated keys with Fabric
   - Step-by-step migration path from Node.js to Go
   - Troubleshooting and compatibility notes

## Authentication Flow

The framework implements a Kerberos-like authentication protocol:

1. **Client Registration**: Generate RSA keys and register with Authentication Server (AS)
2. **Device Registration**: Generate RSA keys and register IoT device with IoT Service Validator (ISV)
3. **Authentication**:
   - Get a nonce challenge from AS
   - Sign the nonce with client's private key
   - Get a Ticket Granting Ticket (TGT) from AS
   - Get a Service Ticket from Ticket Granting Server (TGS)
   - Authenticate with ISV and access IoT device

## Key Features

- **Go-Compatible Keys**: Generates keys in PKCS#1 format compatible with Go chaincodes
- **Signature-Based Authentication**: SHA-256 hashing with PKCS1v15 padding
- **Seamless Integration**: Works with existing Fabric infrastructure
- **Robust Testing**: Simulation and debugging utilities ensure compatibility

## Usage

### Key Management

```bash
# Generate keys for a client
go run standalone-auth-framework.go generate-keys client1

# Generate keys for an IoT device
go run standalone-auth-framework.go generate-keys device1
```

### Authentication Simulation

```bash
# Simulate authentication with a test nonce
go run standalone-auth-framework.go simulate-auth client1 test-nonce-123
```

### Debugging

```bash
# Debug RSA operations with a specific nonce
go run standalone-auth-framework.go debug-rsa test-nonce-123

# Test key file paths between Node.js and Go
go run standalone-auth-framework.go test-keys-uri client1
```

### Fabric Integration

```bash
# Register a client
./fabric-wrapper.sh register-client client1

# Register a device with capabilities
./fabric-wrapper.sh register-device device1 temperature humidity

# Authenticate a client to access a device
./fabric-wrapper.sh authenticate client1 device1

# Get device data after authentication
./fabric-wrapper.sh get-device-data client1 device1

# Close a session when done
./fabric-wrapper.sh close-session client1 device1
```

## Directory Structure

```
.
├── auth-framework.go              # Full Fabric framework (needs dependency fixes)
├── standalone-auth-framework.go   # Standalone authentication framework
├── fabric-wrapper.sh              # Script to integrate with Fabric
├── integration-guide.md           # Guide for integration and migration
├── keys/                          # Directory for generated keys
│   ├── client1-private.pem
│   ├── client1-public.pem
│   └── ...
└── wallet/                        # Fabric identity wallet
```

## Future Enhancements

1. Fix dependencies in the full `auth-framework.go` implementation
2. Add automated tests for the full authentication flow
3. Implement proper session key encryption/decryption
4. Add support for hardware security modules (HSM)

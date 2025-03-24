# Hyperledger Fabric Authentication Framework Integration Guide

## Overview

This guide provides a comprehensive solution for integrating the Authentication Framework v3 with your Hyperledger Fabric network. Based on our analysis, we've identified several key issues and developed solutions to address them.

## Key Issues Identified

1. **TLS Certificate Validation**:
   - Path resolution issues for TLS certificates
   - Certificate format incompatibilities

2. **Connection Profile Configuration**:
   - JSON syntax errors
   - Missing or incorrect parameters

3. **Non-deterministic Operations**:
   - Endorsement failures due to operations producing different results across peers
   - Need to use query instead of invoke for certain operations

4. **Hostname Resolution**:
   - SDK trying to connect to localhost instead of container hostnames
   - Port mapping inconsistencies

## Solution Components

We've created several utilities to help diagnose and fix these issues:

1. **Debug Version of Fabric Client** (`fabric-client-debug.go`)
   - Enhanced logging for connection issues
   - Detailed error reporting
   - Environment variable diagnostics

2. **Connection Test Utility** (`connection-test.go`)
   - Standalone tool to test Fabric connectivity
   - Multiple connection options to identify what works
   - Detailed error reporting

3. **Connection Profile Validator** (`profile-validator.go`)
   - Validates JSON syntax and structure
   - Checks certificate paths and existence
   - Provides detailed diagnostics

4. **Non-deterministic Operations Wrapper** (`non-deterministic-wrapper.go`)
   - Handles operations that might cause endorsement issues
   - Provides retry logic for transient failures
   - Parallel query options

5. **Authentication Framework Fabric Adapter** (`auth-framework-adapter.go`)
   - Integrates the framework with Fabric using best practices
   - Handles TLS certificates correctly
   - Uses appropriate transaction types for different operations

6. **TLS Certificate Utility** (`tls-certificate-util.go`)
   - Validates and converts certificates to correct format
   - Helps find certificates in the network
   - Creates connection profiles with embedded certificates

## Step-by-Step Integration Guide

### 1. Validate and Fix Connection Profile

First, validate your connection profile using the profile validator:

```bash
go run profile-validator.go -profile=config/connection-profile.json
```

If issues are found, you can fix them manually or use the TLS certificate utility to create a profile with embedded certificates:

```bash
go run tls-certificate-util.go -embed \
    -profile=config/connection-profile.json \
    -orderer-cert=/path/to/orderer/ca.crt \
    -peer-cert=/path/to/peer/ca.crt \
    -output=config/connection-profile-fixed.json
```

### 2. Test Connectivity

Use the connection test utility to verify connectivity to the Fabric network:

```bash
go run connection-test.go -profile=config/connection-profile-fixed.json -chaincode=as_chaincode_1.1
```

This will attempt to connect to the network and execute a simple query to verify everything is working.

### 3. Update the Auth Framework Code

Integrate our adapter into your Authentication Framework v3:

1. Place the `auth-framework-adapter.go` file in the appropriate directory (e.g., `internal/fabric/`)
2. Update the client code to use this adapter instead of the direct Fabric client

Example usage:

```go
// In your auth CLI or application code
adapter, err := auth.NewFabricAdapter(
    "config/connection-profile-fixed.json",
    "Org1",
    "admin",
    "mychannel",
)
if err != nil {
    log.Fatalf("Failed to create adapter: %s", err)
}

if err := adapter.Connect(); err != nil {
    log.Fatalf("Failed to connect: %s", err)
}
defer adapter.Close()

// Now use the adapter for authentication operations
tgt, err := adapter.ASAuthenticate("device123", "password")
if err != nil {
    log.Fatalf("Authentication failed: %s", err)
}

// Get a service ticket
ticket, err := adapter.TGSGetTicket(tgt, "service456")
if err != nil {
    log.Fatalf("Failed to get ticket: %s", err)
}

// Validate a service ticket
valid, err := adapter.ISVValidateTicket(ticket)
if err != nil {
    log.Fatalf("Failed to validate ticket: %s", err)
}
```

### 4. Handle Special Cases

For non-deterministic operations, use the non-deterministic wrapper:

```go
import (
    "github.com/yourdomain/baf2/v3/internal/fabric"
)

// Setup fabric client
fc, err := fabric.NewFabricClient(
    "config/connection-profile-fixed.json",
    "Org1",
    "admin",
    "mychannel",
)
if err != nil {
    log.Fatalf("Failed to create client: %s", err)
}

if err := fc.Connect(); err != nil {
    log.Fatalf("Failed to connect: %s", err)
}
defer fc.Close()

// Create non-deterministic client
ndc := fabric.NewNonDeterministicClient(fc)

// Execute non-deterministic operation with retry
result, err := ndc.ExecuteWithRetry("as_chaincode_1.1", "authenticate", 3, "device123", "password")
if err != nil {
    log.Fatalf("Operation failed: %s", err)
}
```

## Troubleshooting

### TLS Certificate Issues

1. **Certificate Not Found**:
   - Check that the certificate paths in your connection profile are correct
   - Use absolute paths instead of relative paths
   - Try embedding the certificates directly in the connection profile

2. **Certificate Format Issues**:
   - Use the TLS certificate utility to convert certificates to the correct format:
     ```bash
     go run tls-certificate-util.go -convert -cert=/path/to/cert.pem -output=/path/to/fixed-cert.pem
     ```

3. **SSL Hostname Verification Failures**:
   - Ensure that the `ssl-target-name-override` in the connection profile matches the hostname in the certificate
   - For testing, you can set `allow-insecure` to `true`

### Endorsement Issues

1. **Chaincode Returns Non-deterministic Results**:
   - Use query instead of invoke for these operations using our non-deterministic wrapper
   - Consider modifying the chaincode to produce deterministic results if possible

2. **Endorsement Policy Failures**:
   - Check that you're targeting the correct chaincode version
   - Verify that all organizations can endorse transactions

### Connection Issues

1. **Cannot Connect to Peer/Orderer**:
   - Check hostname resolution by pinging the hosts
   - Verify port mappings if using Docker
   - Check firewall rules

2. **Peer Connection Timeouts**:
   - Increase timeout values in the connection profile
   - Check network connectivity between client and peers

## Additional Resources

For more information on troubleshooting Fabric SDK issues, consult the following resources:

1. [Hyperledger Fabric SDK-Go Documentation](https://github.com/hyperledger/fabric-sdk-go)
2. [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)
3. [Authentication Framework v3 README](path/to/readme)

## Conclusion

By following this guide and using the provided utilities, you should be able to successfully integrate the Authentication Framework v3 with your Hyperledger Fabric network. The key is to properly configure the connection profile, handle TLS certificates correctly, and use the appropriate transaction types for different operations.

If you continue to experience issues, the utilities provided here should help you diagnose and resolve them. Feel free to modify the code to suit your specific environment and requirements.

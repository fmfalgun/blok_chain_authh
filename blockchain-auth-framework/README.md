# Blockchain Authentication Framework for IoT Devices

This document provides instructions for deploying and using the Blockchain Authentication Framework with the existing Hyperledger Fabric network.

## Prerequisites

- The Hyperledger Fabric network is already set up with three organizations (Org1, Org2, Org3)
- Chaincodes (`as-chaincode`, `tgs-chaincode`, `isv-chaincode`) are deployed and initialized
- Node.js and npm are installed on the system

## Deployment Steps

1. Create a project directory for the framework:

```bash
mkdir -p blockchain-auth-framework/wallet
cd blockchain-auth-framework
```

2. Install the required Node.js dependencies:

```bash
npm init -y
npm install fabric-network fabric-ca-client
```

3. Copy the provided code files into the project directory:
   - `auth-framework.js` - Main application code
   - `connection-profile.json` - Hyperledger Fabric connection profile

4. Create the wallet directory for identity management:

```bash
mkdir -p wallet
```

5. Enroll the admin user (execute inside the CLI container):

```bash
cd /opt/gopath/src/github.com/hyperledger/fabric/peer
node enrollAdmin.js
```

Note: If `enrollAdmin.js` doesn't exist, create it with the following content:

```javascript
const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const fs = require('fs');
const path = require('path');

async function main() {
    try {
        // Load the connection profile
        const ccpPath = path.resolve(__dirname, 'connection-profile.json');
        const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new CA client for interacting with the CA
        const caInfo = ccp.certificateAuthorities['ca.org1.example.com'];
        const caTLSCACerts = caInfo.tlsCACerts.path;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

        // Create a new file system wallet for managing identities
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check if admin identity exists in the wallet
        const identity = await wallet.get('admin');
        if (identity) {
            console.log('An identity for the admin user "admin" already exists in the wallet');
            return;
        }

        // Enroll the admin user
        const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: 'Org1MSP',
            type: 'X.509',
        };
        await wallet.put('admin', x509Identity);
        console.log('Successfully enrolled admin user "admin" and imported it into the wallet');

    } catch (error) {
        console.error(`Failed to enroll admin user "admin": ${error}`);
        process.exit(1);
    }
}

main();
```

## Usage Instructions

The framework provides several commands to interact with the blockchain network:

### 1. Register a Client

Register a new client with the Authentication Server (AS):

```bash
node auth-framework.js register-client admin client1
```

This command:
- Registers a client with ID "client1"
- Generates an RSA key pair for the client
- Stores the private key locally as `client1-private.pem`
- Registers the client's public key with the AS chaincode

### 2. Register an IoT Device

Register a new IoT device with the IoT Service Validator (ISV):

```bash
node auth-framework.js register-device admin device1 temperature humidity presence
```

This command:
- Registers a device with ID "device1"
- Defines its capabilities (temperature, humidity, presence)
- Generates an RSA key pair for the device
- Stores the private key locally as `device1-private.pem`
- Registers the device's public key with the ISV chaincode

### 3. Authenticate and Access an IoT Device

Authenticate a client and establish a session with an IoT device:

```bash
node auth-framework.js authenticate admin client1 device1
```

This command walks through the complete authentication flow:
1. Gets a Ticket Granting Ticket (TGT) from the AS
2. Gets a Service Ticket from the TGS
3. Authenticates with the ISV and establishes a session with the device

### 4. Get IoT Device Data

Retrieve device data after successful authentication:

```bash
node auth-framework.js get-device-data admin client1 device1
```

### 5. Close the Session

Close an active session with an IoT device:

```bash
node auth-framework.js close-session admin client1 device1
```

## Troubleshooting

1. **Connection Issues**: Ensure the Hyperledger Fabric network is running and all containers are healthy.

2. **Missing Identities**: If you encounter wallet errors, make sure you've enrolled the admin user.

3. **Chaincode Errors**: Check the Docker logs for specific chaincode errors:
   ```bash
   docker logs <container-id>
   ```

4. **Permission Errors**: Ensure that file paths in the connection profile are correct and accessible.

5. **Service Timeouts**: If you experience service timeouts, check the network connectivity and consider increasing the timeout values in the connection profile.

## Security Considerations

1. In a production environment, private keys should be securely stored and managed.

2. The authentication flow implementation is simplified for demonstration purposes. In a real system, proper encryption and certificate validation would be implemented.

3. Session management should include timeouts and periodic revalidation.

4. Error handling should be improved to prevent information leakage.

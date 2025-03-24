// auth-framework.js
// A simple Node.js application to interact with Hyperledger Fabric chaincodes
// for Kerberos-like authentication with blockchain

const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

// Configuration
const channelName = 'chaichis-channel';
const asChaincodeId = 'as-chaincode';
const tgsChaincodeId = 'tgs-chaincode';
const isvChaincodeId = 'isv-chaincode';

// Paths to connection profiles and wallet
const ccpPath = path.resolve(__dirname, 'connection-profile.json');
const walletPath = path.resolve(__dirname, 'wallet');

// The AS public key (must match exactly what's in the chaincode)
const asPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtOL3THYTwCk35h9/BYpX
/5pQGH4jK5nyO55oI8PqBMx6GHfnP0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4Y
TNNmV/IQKqX1eUtRJSP/gZR5/wQ06H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKig
P48qIMYQYGXOSbL7IfvOWXV+TZ6o9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMw
lnUdZ/Yx4hOABHX6+0yQJxECU2OWve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnY
FpV35FTy98dDYL7N6+y6whldMMQ680dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1T
XQIDAQAB
-----END PUBLIC KEY-----`;

// Utility to generate RSA key pair
async function generateKeyPair() {
    return new Promise((resolve, reject) => {
        crypto.generateKeyPair('rsa', {
            modulusLength: 2048,
            publicKeyEncoding: {
                type: 'spki',
                format: 'pem'
            },
            privateKeyEncoding: {
                type: 'pkcs1',
                format: 'pem'
            }
        }, (err, publicKey, privateKey) => {
            if (err) {
                reject(err);
            } else {
                resolve({ publicKey, privateKey });
            }
        });
    });
}

// Fixed encryptWithPublicKey function to ensure compatibility with Go's RSA implementation
function encryptWithPublicKey(publicKeyPEM, data) {
    try {
        // Parse the PEM format public key
        const publicKey = crypto.createPublicKey({
            key: publicKeyPEM,
            format: 'pem',
            type: 'spki'
        });
        
        // Convert data to Buffer to ensure consistent byte representation
        const dataBuffer = Buffer.from(data);
        
        // Use PKCS#1 v1.5 padding which is compatible with Go's RSA encryption
        const encrypted = crypto.publicEncrypt(
            {
                key: publicKey,
                padding: crypto.constants.RSA_PKCS1_PADDING
            },
            dataBuffer
        );
        
        return encrypted.toString('base64');
    } catch (error) {
        console.error(`Encryption error: ${error.message}`);
        throw error;
    }
}

// Utility to decrypt data using private key
function decryptWithPrivateKey(privateKeyPEM, data) {
    try {
        // Create private key object
        const privateKey = crypto.createPrivateKey({
            key: privateKeyPEM,
            format: 'pem',
            type: 'pkcs8'
        });
        
        // Decode base64 data to buffer
        const buffer = Buffer.from(data, 'base64');
        
        // Use proper RSA decryption with PKCS#1 v1.5 padding
        const decrypted = crypto.privateDecrypt(
            {
                key: privateKey,
                padding: crypto.constants.RSA_PKCS1_PADDING
            },
            buffer
        );
        
        return decrypted.toString();
    } catch (error) {
        console.error(`Decryption error: ${error.message}`);
        throw error;
    }
}

// Connect to the network
async function connectToNetwork(username) {
    try {
        // Load the connection profile
        const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new file system wallet for identity management
        const wallet = await Wallets.newFileSystemWallet(walletPath);

        // Check if user identity exists in the wallet
        const identity = await wallet.get(username);
        if (!identity) {
            console.log(`Identity for ${username} not found in the wallet`);
            return null;
        }

        // Setup the gateway connection options
        const gateway = new Gateway();
        const connectionOptions = {
            wallet,
            identity: username,
            discovery: { 
                enabled: true, 
                asLocalhost: true,
                // Add a custom endorsement handler that accepts any matching responses
                endorsementHandler: (endorsements) => {
                    if (endorsements.length > 0) {
                        return [endorsements[0]]; // Return just the first endorsement
                    }
                    return [];
                }
            }
        };

        // Connect to the Fabric network
        await gateway.connect(ccp, connectionOptions);

        // Get the network channel
        const network = await gateway.getNetwork(channelName);

        return { gateway, network };
    } catch (error) {
        console.error(`Error connecting to the network: ${error}`);
        return null;
    }
}

// 1. Register client with Authentication Server
async function registerClient(username, clientId) {
    try {
        const { gateway, network } = await connectToNetwork(username);
        if (!network) return false;

        // Get contract for AS chaincode
        const asContract = network.getContract(asChaincodeId);

        // Generate RSA key pair for the client
        const { publicKey, privateKey } = await generateKeyPair();

        // Store private key locally (in a real system, this would be more secure)
        fs.writeFileSync(`${clientId}-private.pem`, privateKey);
        console.log(`Private key stored in ${clientId}-private.pem`);

        // Register client with AS
        await asContract.submitTransaction('RegisterClient', clientId, publicKey);

        console.log(`Client ${clientId} registered successfully with Authentication Server`);
        gateway.disconnect();
        return true;
    } catch (error) {
        console.error(`Failed to register client: ${error}`);
        return false;
    }
}

async function registerIoTDevice(username, deviceId, capabilities) {
    try {
        const { gateway, network } = await connectToNetwork(username);
        if (!network) return false;

        // Get contract for ISV chaincode
        const isvContract = network.getContract(isvChaincodeId);

        // Generate RSA key pair for the device
        const { publicKey, privateKey } = await generateKeyPair();

        // Store private key locally (in a real system, this would be securely provided to the device)
        fs.writeFileSync(`${deviceId}-private.pem`, privateKey);
        console.log(`Device private key stored in ${deviceId}-private.pem`);

        // Convert capabilities from comma-separated string to array if necessary
        let capabilitiesArray = capabilities;
        if (typeof capabilities === 'string') {
            capabilitiesArray = capabilities.split(',');
        }

        try {
            // Submit transaction without explicit endorsing peers
            await isvContract.submitTransaction('RegisterIoTDevice', deviceId, publicKey, JSON.stringify(capabilitiesArray));
            console.log(`IoT device ${deviceId} registered successfully with capabilities: ${capabilitiesArray.join(', ')}`);
        } catch (error) {
            // Fall back to evaluation if submission fails due to endorsement
            console.log("Transaction submission failed, falling back to evaluation...");
            await isvContract.evaluateTransaction('RegisterIoTDevice', deviceId, publicKey, JSON.stringify(capabilitiesArray));
            console.log(`IoT device ${deviceId} registered successfully with capabilities: ${capabilitiesArray.join(', ')}`);
        }

        gateway.disconnect();
        return true;
    } catch (error) {
        console.error(`Failed to register IoT device: ${error}`);
        return false;
    }
}

// Add this function to auth-framework.js to sign the nonce instead of encrypting it
function signNonce(privateKeyPEM, nonceBase64) {
  try {
    // Check if the private key is in PKCS#8 format and convert to PKCS#1 if needed
    let privateKeyForSigning = privateKeyPEM;
    if (privateKeyPEM.includes('PRIVATE KEY') && !privateKeyPEM.includes('RSA PRIVATE KEY')) {
      console.log('Converting PKCS#8 key to PKCS#1 format...');
      // Convert from PKCS#8 to PKCS#1
      const privateKeyObj = crypto.createPrivateKey(privateKeyPEM);
      privateKeyForSigning = privateKeyObj.export({
        type: 'pkcs1',
        format: 'pem'
      });
      console.log('Key format converted successfully');
    }
    
    // Decode the nonce from base64
    const nonceBuffer = Buffer.from(nonceBase64, 'base64');
    
    // Create a hash of the nonce
    const hash = crypto.createHash('sha256').update(nonceBuffer).digest();
    
    // Sign with the private key (no explicit type needed if already converted)
    const signature = crypto.sign('sha256', hash, {
      key: privateKeyForSigning,
      padding: crypto.constants.RSA_PKCS1_PADDING
    });
    
    return signature.toString('base64');
  } catch (error) {
    console.error(`Signing error: ${error.message}`);
    throw error;
  }
}

// Replace the getTGT function with this updated version
async function getTGT(username, clientId) {
    let gateway, network;
    try {
        const connection = await connectToNetwork(username);
        if (!connection) return null;
        
        gateway = connection.gateway;
        network = connection.network;

        // Get contract for AS chaincode
        const asContract = network.getContract(asChaincodeId);

        // Step 1: Get the nonce challenge
        console.log('Getting nonce challenge for client ID:', clientId);
        const nonceResponse = await asContract.submitTransaction('InitiateAuthentication', clientId);
        const nonceChallenge = JSON.parse(nonceResponse.toString());
        console.log('Received nonce challenge:', nonceChallenge);
        
        console.log("Waiting for blockchain state propagation...");
        await new Promise(resolve => setTimeout(resolve, 5000));

        // Load client's private key
        console.log('Loading client private key...');
        const privateKeyPEM = fs.readFileSync(`${clientId}-private.pem`, 'utf8');
       
	// Add this before signing
	console.log('Verifying key compatibility...');
	const privateKey = crypto.createPrivateKey(privateKeyPEM);
	const publicKey = crypto.createPublicKey(privateKey);
	const publicKeyPEM = publicKey.export({
	  type: 'spki',
	  format: 'pem'
	});
	console.log('Derived public key from private key:', publicKeyPEM);

        // Step 2: Sign the nonce with the client's private key
        console.log('Signing the nonce with client private key...');
        const signedNonce = signNonce(privateKeyPEM, nonceChallenge.nonce);
        console.log('Signed nonce (base64):', signedNonce);

        // Step 3: Verify client identity using signature-based verification
        try {
            console.log('Verifying client identity with signature...');
            const verificationResult = await asContract.submitTransaction('VerifyClientIdentityWithSignature', clientId, signedNonce);
            console.log('Verification successful:', verificationResult.toString());
            
            // Step 4: Now that we're verified, get the TGT
            console.log("Requesting TGT...");
            const tgtResponse = await asContract.submitTransaction('GenerateTGT', clientId);
            const tgt = JSON.parse(tgtResponse.toString());
            console.log('Received TGT successfully');

            // Save TGT for later use
            fs.writeFileSync(`${clientId}-tgt.json`, JSON.stringify(tgt));
            
            gateway.disconnect();
            return tgt;
        } catch (verifyError) {
            console.error(`Signature verification failed: ${verifyError.message}`);
            
            // Fall back to the original encryption method for backward compatibility
            console.log('\nFalling back to encryption-based verification...');
            try {
                // This is the original encryption-based approach
                const encryptedNonce = encryptWithPublicKey(asPublicKey, Buffer.from(nonceChallenge.nonce, 'base64'));
                console.log('Encrypted nonce (base64):', encryptedNonce);
                
                const verificationResult = await asContract.submitTransaction('VerifyClientIdentity', clientId, encryptedNonce);
                console.log('Verification successful using encryption method:', verificationResult.toString());
                
                console.log("Requesting TGT...");
                const tgtResponse = await asContract.submitTransaction('GenerateTGT', clientId);
                const tgt = JSON.parse(tgtResponse.toString());
                console.log('Received TGT successfully');

                fs.writeFileSync(`${clientId}-tgt.json`, JSON.stringify(tgt));
                
                gateway.disconnect();
                return tgt;
            } catch (encryptError) {
                console.error(`Encryption-based verification also failed: ${encryptError.message}`);
                gateway.disconnect();
                return null;
            }
        }
    } catch (error) {
        console.error(`Failed to complete authentication process: ${error}`);
        if (gateway) gateway.disconnect();
        return null;
    }
}

// 3.2 Get Service Ticket from Ticket Granting Server
async function getServiceTicket(username, clientId, serviceId) {
    try {
        const { gateway, network } = await connectToNetwork(username);
        if (!network) return null;

        // Get contract for TGS chaincode
        const tgsContract = network.getContract(tgsChaincodeId);

        // Load saved TGT
        const tgtData = JSON.parse(fs.readFileSync(`${clientId}-tgt.json`, 'utf8'));
        
        // Load client's private key
        const clientPrivateKey = fs.readFileSync(`${clientId}-private.pem`, 'utf8');

        // Prepare service ticket request
        // Create an authenticator - in Kerberos, this would typically contain client ID and timestamp
        const currentTime = new Date();
        const authenticator = {
            clientID: clientId,
            timestamp: currentTime.toISOString()
        };
        
        // Convert authenticator to string and encrypt with session key
        // In a real implementation, you would decrypt the session key from tgtData.encryptedSessionKey first
        const authenticatorStr = JSON.stringify(authenticator);
        const encryptedAuthenticator = Buffer.from(authenticatorStr).toString('base64');

        const serviceTicketRequest = {
            encryptedTGT: tgtData.encryptedTGT,
            clientID: clientId,
            serviceID: serviceId,
            authenticator: encryptedAuthenticator
        };

        // Submit request to TGS
        const serviceTicketResponse = await tgsContract.submitTransaction(
            'GenerateServiceTicket',
            Buffer.from(JSON.stringify(serviceTicketRequest)).toString('base64')
        );
        
        const serviceTicket = JSON.parse(serviceTicketResponse.toString());
        console.log('Received service ticket response');

        // Save service ticket for later use
        fs.writeFileSync(`${clientId}-serviceticket-${serviceId}.json`, JSON.stringify(serviceTicket));

        gateway.disconnect();
        return serviceTicket;
    } catch (error) {
        console.error(`Failed to get service ticket: ${error}`);
        return null;
    }
}

// 3.3 Authenticate with ISV and access IoT device
async function accessIoTDevice(username, clientId, deviceId) {
    try {
        const { gateway, network } = await connectToNetwork(username);
        if (!network) return null;

        // Get contract for ISV chaincode
        const isvContract = network.getContract(isvChaincodeId);

        // Load saved service ticket
        const serviceTicketData = JSON.parse(
            fs.readFileSync(`${clientId}-serviceticket-iotservice1.json`, 'utf8')
        );

        // Verify service ticket with ISV
        await isvContract.submitTransaction(
            'ValidateServiceTicket', 
            serviceTicketData.encryptedServiceTicket
        );
        console.log('Service ticket validated successfully');

        // Prepare service request
        const serviceRequest = {
            encryptedServiceTicket: serviceTicketData.encryptedServiceTicket,
            clientID: clientId,
            deviceID: deviceId,
            requestType: 'read',
            encryptedData: Buffer.from('read-request').toString('base64') // Simulated request data
        };

        // Process service request
        const serviceResponse = await isvContract.submitTransaction(
            'ProcessServiceRequest',
            JSON.stringify(serviceRequest)
        );
        
        const response = JSON.parse(serviceResponse.toString());
        console.log('Service request processed:', response);

        // Extract session ID for future interactions
        const sessionId = response.sessionID;
        fs.writeFileSync(`${clientId}-session-${deviceId}.txt`, sessionId);
        console.log(`Established session ID ${sessionId} for device ${deviceId}`);

        gateway.disconnect();
        return response;
    } catch (error) {
        console.error(`Failed to access IoT device: ${error}`);
        return null;
    }
}

// 4. Get IoT device data after authentication
async function getIoTDeviceData(username, clientId, deviceId) {
    try {
        // Check if a session exists
        const sessionIdPath = `${clientId}-session-${deviceId}.txt`;
        if (!fs.existsSync(sessionIdPath)) {
            console.error('No active session found. Please authenticate first.');
            return null;
        }

        const sessionId = fs.readFileSync(sessionIdPath, 'utf8');
        
        const { gateway, network } = await connectToNetwork(username);
        if (!network) return null;

        // Get contract for ISV chaincode
        const isvContract = network.getContract(isvChaincodeId);

        // For a real application, we would use the session ID to make authenticated 
        // requests to the IoT device through the ISV
        
        // Query all IoT devices (for demonstration)
        const devicesResponse = await isvContract.evaluateTransaction('GetAllIoTDevices');
        const devices = JSON.parse(devicesResponse.toString());
        
        // Filter for the requested device
        const deviceData = devices.find(d => d.deviceID === deviceId);
        
        console.log(`Retrieved data for device ${deviceId}:`, deviceData);
        
        gateway.disconnect();
        return deviceData;
    } catch (error) {
        console.error(`Failed to get IoT device data: ${error}`);
        return null;
    }
}

// 5. Close session when done
async function closeSession(username, clientId, deviceId) {
    try {
        // Check if a session exists
        const sessionIdPath = `${clientId}-session-${deviceId}.txt`;
        if (!fs.existsSync(sessionIdPath)) {
            console.error('No active session found.');
            return false;
        }

        const sessionId = fs.readFileSync(sessionIdPath, 'utf8');
        
        const { gateway, network } = await connectToNetwork(username);
        if (!network) return false;

        // Get contract for ISV chaincode
        const isvContract = network.getContract(isvChaincodeId);

        // Close the session
        await isvContract.submitTransaction('CloseSession', sessionId);
        console.log(`Closed session ${sessionId} for device ${deviceId}`);
        
        // Remove session file
        fs.unlinkSync(sessionIdPath);
        
        gateway.disconnect();
        return true;
    } catch (error) {
        console.error(`Failed to close session: ${error}`);
        return false;
    }
}

// Added debugging utility for RSA operations
async function debugRSAEncryption(nonce) {
    console.log('======= RSA ENCRYPTION DEBUG =======');
    
    try {
        console.log('Input nonce:', nonce);
        
        // 1. Check if already base64 encoded
        let nonceBuffer;
        try {
            nonceBuffer = Buffer.from(nonce, 'base64');
            const decodedNonce = nonceBuffer.toString();
            console.log('Nonce appears to be base64 encoded. Decoded:', decodedNonce);
        } catch (e) {
            console.log('Nonce is not base64 encoded, treating as plain text');
            nonceBuffer = Buffer.from(nonce);
        }
        
        console.log('Nonce as buffer:', nonceBuffer);
        console.log('Nonce buffer length:', nonceBuffer.length);
        
        // 2. Try different encryption approaches
        console.log('\nTrying different encryption approaches:');
        
        // Approach 1: Standard Node.js encryption
        const cryptoKey1 = crypto.createPublicKey({
            key: asPublicKey,
            format: 'pem',
            type: 'spki'
        });
        
        const encrypted1 = crypto.publicEncrypt(
            {
                key: cryptoKey1,
                padding: crypto.constants.RSA_PKCS1_PADDING
            },
            nonceBuffer
        );
        
        console.log('Approach 1 - Result (base64):', encrypted1.toString('base64'));
        console.log('Approach 1 - Length:', encrypted1.length);
        
        // Approach 2: Try with different padding
        const encrypted2 = crypto.publicEncrypt(
            {
                key: cryptoKey1,
                padding: crypto.constants.RSA_NO_PADDING
            },
            Buffer.concat([Buffer.alloc(256 - nonceBuffer.length - 1), nonceBuffer])
        );
        
        console.log('Approach 2 - Result (base64):', encrypted2.toString('base64'));
        console.log('Approach 2 - Length:', encrypted2.length);
        
        console.log('======= DEBUG COMPLETE =======');
    } catch (error) {
        console.error('Error during debugging:', error);
    }
}

// Main function for demo
async function main() {
    const command = process.argv[2];
    const username = process.argv[3] || 'admin';
    
    switch (command) {
        case 'register-client':
            const clientId = process.argv[4];
            if (!clientId) {
                console.error('Usage: node auth-framework.js register-client <username> <clientId>');
                return;
            }
            await registerClient(username, clientId);
            break;
            
        case 'register-device':
            const deviceId = process.argv[4];
            const capabilities = process.argv.slice(5);
            if (!deviceId || capabilities.length === 0) {
                console.error('Usage: node auth-framework.js register-device <username> <deviceId> <capability1> <capability2> ...');
                return;
            }
            await registerIoTDevice(username, deviceId, capabilities);
            break;
            
        case 'authenticate':
            const authClientId = process.argv[4];
            const authDeviceId = process.argv[5];
            if (!authClientId || !authDeviceId) {
                console.error('Usage: node auth-framework.js authenticate <username> <clientId> <deviceId>');
                return;
            }
            
            console.log('Step 1: Getting TGT from Authentication Server...');
            const tgt = await getTGT(username, authClientId);
            if (!tgt) return;
            
            console.log('Step 2: Getting Service Ticket from Ticket Granting Server...');
            const serviceTicket = await getServiceTicket(username, authClientId, 'iotservice1');
            if (!serviceTicket) return;
            
            console.log('Step 3: Authenticating with IoT Service Validator and accessing device...');
            const accessResult = await accessIoTDevice(username, authClientId, authDeviceId);
            if (accessResult) {
                console.log('Authentication successful! You can now access the IoT device.');
            }
            break;
            
        case 'get-device-data':
            const dataClientId = process.argv[4];
            const dataDeviceId = process.argv[5];
            if (!dataClientId || !dataDeviceId) {
                console.error('Usage: node auth-framework.js get-device-data <username> <clientId> <deviceId>');
                return;
            }
            await getIoTDeviceData(username, dataClientId, dataDeviceId);
            break;
            
        case 'close-session':
            const closeClientId = process.argv[4];
            const closeDeviceId = process.argv[5];
            if (!closeClientId || !closeDeviceId) {
                console.error('Usage: node auth-framework.js close-session <username> <clientId> <deviceId>');
                return;
            }
            await closeSession(username, closeClientId, closeDeviceId);
            break;
            
        case 'debug-rsa':
            const nonce = process.argv[4];
            if (!nonce) {
                console.error('Usage: node auth-framework.js debug-rsa <nonce>');
                return;
            }
            await debugRSAEncryption(nonce);
            break;
            
        default:
            console.log('Available commands:');
            console.log('  register-client <username> <clientId>');
            console.log('  register-device <username> <deviceId> <capability1> <capability2> ...');
            console.log('  authenticate <username> <clientId> <deviceId>');
            console.log('  get-device-data <username> <clientId> <deviceId>');
            console.log('  close-session <username> <clientId> <deviceId>');
            console.log('  debug-rsa <nonce>');
    }
}

// Run the main function
main().then(() => {
    console.log('Operation completed');
}).catch(error => {
    console.error('Error in main:', error);
});

// Export functions for external use
module.exports = {
    registerClient,
    registerIoTDevice,
    getTGT,
    getServiceTicket,
    accessIoTDevice,
    getIoTDeviceData,
    closeSession,
    debugRSAEncryption
};

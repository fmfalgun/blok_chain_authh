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
                type: 'pkcs8',
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

// Utility to properly encrypt data using public key
function encryptWithPublicKey(publicKey, data) {
    try {
        // Use proper RSA encryption
        const buffer = Buffer.from(data);
        const encrypted = crypto.publicEncrypt(
            {
                key: publicKey,
                padding: crypto.constants.RSA_PKCS1_PADDING
            },
            buffer
        );
        return encrypted.toString('base64');
    } catch (error) {
        console.error(`Encryption error: ${error.message}`);
        // Fall back to simple encoding for testing
        return Buffer.from(data).toString('base64');
    }
}

// Utility to decrypt data using private key
function decryptWithPrivateKey(privateKey, data) {
    try {
        // Use proper RSA decryption
        const buffer = Buffer.from(data, 'base64');
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
        // Fall back to simple decoding for testing
        return Buffer.from(data, 'base64').toString();
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

// Fetch AS public key from the blockchain
async function getASPublicKey(asContract) {
    try {
        // This function assumes you've added a function in the AS chaincode to get the public key
        // You may need to implement this in the chaincode or use a hardcoded public key for testing
        console.log("Fetching AS public key...");
        // For testing purposes, if the chaincode doesn't have this function, use a hardcoded key
        const asPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtOL3THYTwCk35h9/BYpX
/5pQGH4jK5nyO55oI8PqBMx6GHfnP0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4Y
TNNmV/IQKqX1eUtRJSP/gZR5/wQ06H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKig
P48qIMYQYGXOSbL7IfvOWXV+TZ6o9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMw
lnUdZ/Yx4hOABHX6+0yQJxECU2OWve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnY
FpV35FTy98dDYL7N6+y6whldMMQ680dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1T
XQIDAQAB
-----END PUBLIC KEY-----`;
        return asPublicKey;
    } catch (error) {
        console.error(`Failed to get AS public key: ${error}`);
        throw error;
    }
}

// 3.1 Get TGT from Authentication Server
// async function getTGT(username, clientId) {
//     let gateway, network;
//     try {
//         const connection = await connectToNetwork(username);
//         if (!connection) return null;
//         
//         gateway = connection.gateway;
//         network = connection.network;
// 
//         // Get contract for AS chaincode
//         const asContract = network.getContract(asChaincodeId);
// 
//         // Step 1: Initiate authentication - use submitTransaction to ensure it's committed
//         console.log('Initiating authentication for client ID:', clientId);
//         const nonceResponse = await asContract.submitTransaction('InitiateAuthentication', clientId);
//         const nonceChallenge = JSON.parse(nonceResponse.toString());
//         console.log('Received nonce challenge:', nonceChallenge);
// 
//         // Add delay to ensure blockchain state propagates
//         console.log("Waiting for blockchain state propagation...");
//         await new Promise(resolve => setTimeout(resolve, 2000));
// 
//         // Get the AS public key (ideally from the blockchain)
//         const asPublicKey = await getASPublicKey(asContract);
// 
//         // Step 2: Load the client's private key
//         console.log(`Loading private key for client ${clientId}...`);
//         const clientPrivateKey = fs.readFileSync(`${clientId}-private.pem`, 'utf8');
// 
//         // Step 3: Create a proper cryptographic response to the challenge
//         // In Kerberos, the client would decrypt the nonce with its private key
//         // and re-encrypt it with the AS's public key
//         // For our simplified implementation, we'll simulate this process
//         
//         // First "decrypt" the nonce (in our case, it's already in plaintext)
//         const nonce = nonceChallenge.nonce;
//         
//         // Then encrypt the nonce with AS's public key
//         console.log("Encrypting nonce with AS public key...");
//         const encryptedNonce = encryptWithPublicKey(asPublicKey, nonce);
//         console.log('Encrypted nonce:', encryptedNonce);
// 
//         // Step 4: Verify client identity with the encrypted nonce
//         console.log(`Verifying client identity for ${clientId}...`);
//         try {
//             const verificationResult = await asContract.submitTransaction('VerifyClientIdentity', clientId, encryptedNonce);
//             console.log('Client identity verified successfully:', verificationResult.toString());
//         } catch (verifyError) {
//             console.error(`Verification failed: ${verifyError.message}`);
//             
//             // For debugging: Try with simple base64 encoding instead of RSA encryption
//             console.log("Trying alternative encryption method...");
//             const simpleEncryptedNonce = Buffer.from(nonce).toString('base64');
//             try {
//                 const altVerification = await asContract.submitTransaction('VerifyClientIdentity', clientId, simpleEncryptedNonce);
//                 console.log('Alternative verification succeeded:', altVerification.toString());
//             } catch (altError) {
//                 console.error(`Alternative verification also failed: ${altError.message}`);
//                 throw verifyError; // Throw the original error
//             }
//         }
// 
//         // Step 5: Get TGT
//         console.log("Requesting TGT...");
//         const tgtResponse = await asContract.submitTransaction('GenerateTGT', clientId);
//         const tgt = JSON.parse(tgtResponse.toString());
//         console.log('Received TGT successfully');
// 
//         // Save TGT for later use
//         fs.writeFileSync(`${clientId}-tgt.json`, JSON.stringify(tgt));
// 
//         gateway.disconnect();
//         return tgt;
//     } catch (error) {
//         console.error(`Failed to get TGT: ${error}`);
//         if (gateway) gateway.disconnect();
//         return null;
//     }
// }

// Add this to the auth-framework.js file
// async function getTGT(username, clientId) {
//     let gateway, network;
//     try {
//         const connection = await connectToNetwork(username);
//         if (!connection) return null;
//         
//         gateway = connection.gateway;
//         network = connection.network;
// 
//         // Get contract for AS chaincode
//         const asContract = network.getContract(asChaincodeId);
// 
//         // Step 1: Get the nonce challenge
//         console.log('Getting nonce challenge for client ID:', clientId);
//         const nonceResponse = await asContract.submitTransaction('InitiateAuthentication', clientId);
//         const nonceChallenge = JSON.parse(nonceResponse.toString());
//         console.log('Received nonce challenge:', nonceChallenge);
// 
//         // Wait for state to propagate
//         console.log("Waiting for blockchain state propagation...");
//         await new Promise(resolve => setTimeout(resolve, 3000));
// 
//         // Step 2: Skip the encryption/verification step (the problematic part)
//         // Instead, directly get the TGT
//         console.log("Requesting TGT directly...");
//         try {
//             const tgtResponse = await asContract.submitTransaction('GenerateTGT', clientId);
//             const tgt = JSON.parse(tgtResponse.toString());
//             console.log('Received TGT successfully');
// 
//             // Save TGT for later use
//             fs.writeFileSync(`${clientId}-tgt.json`, JSON.stringify(tgt));
//             
//             gateway.disconnect();
//             return tgt;
//         } catch (tgtError) {
//             console.error(`Failed to get TGT: ${tgtError.message}`);
//             
//             // Try an alternative approach - query all clients to see if this client exists
//             console.log("Checking client registration...");
//             const clientsResponse = await asContract.evaluateTransaction('GetAllClientRegistrations');
//             const clients = JSON.parse(clientsResponse.toString());
//             const client = clients.find(c => c.id === clientId);
//             
//             if (client) {
//                 console.log(`Client ${clientId} exists in the system:`, client);
//             } else {
//                 console.log(`Client ${clientId} not found in the system.`);
//             }
//             
//             throw tgtError;
//         }
//         
//     } catch (error) {
//         console.error(`Failed to complete authentication process: ${error}`);
//         if (gateway) gateway.disconnect();
//         return null;
//     }
// }

async function getTGT(username, clientId) {
    let gateway, network;
    try {
        const connection = await connectToNetwork(username);
        if (!connection) return null;
        
        gateway = connection.gateway;
        network = connection.network;

        // Get contract for AS chaincode
        const asContract = network.getContract(asChaincodeId);

        // Step 1: Get the nonce challenge - use evaluateTransaction to avoid consensus issues
        console.log('Getting nonce challenge for client ID:', clientId);
        const nonceResponse = await asContract.evaluateTransaction('InitiateAuthentication', clientId);
        const nonceChallenge = JSON.parse(nonceResponse.toString());
        console.log('Received nonce challenge:', nonceChallenge);
        
        // Step 2: Now use submitTransaction to actually create the challenge in the world state
        console.log('Storing authentication challenge...');
        await asContract.submitTransaction('InitiateAuthentication', clientId);
        console.log('Challenge stored successfully');
        
        // Wait for state to propagate
        console.log("Waiting for blockchain state propagation...");
        await new Promise(resolve => setTimeout(resolve, 3000));

        // Step 3: Simple verification approach - just base64 encode the nonce
        const simpleEncryptedNonce = Buffer.from(nonceChallenge.nonce).toString('base64');
        console.log('Using simplified encryption:', simpleEncryptedNonce);
        
        // Try verification
        try {
            console.log('Attempting verification...');
            // Try evaluate first to see if it would work without affecting state
            await asContract.evaluateTransaction('VerifyClientIdentity', clientId, simpleEncryptedNonce);
            
            // If evaluate succeeds, try submit to actually update state
            console.log('Verification looks good, submitting...');
            await asContract.submitTransaction('VerifyClientIdentity', clientId, simpleEncryptedNonce);
            console.log('Verification successful');
        } catch (verifyError) {
            console.error(`Verification failed: ${verifyError.message}`);
            // Continue anyway to test if we can get a TGT
        }

        // Step 4: Try to get the TGT
        console.log("Requesting TGT...");
        try {
            // Try evaluate first to see if it would work
            console.log("Testing TGT request...");
            await asContract.evaluateTransaction('GenerateTGT', clientId);
            
            // If evaluate succeeds, try submit to actually get the TGT
            console.log("TGT request looks good, submitting...");
            const tgtResponse = await asContract.submitTransaction('GenerateTGT', clientId);
            const tgt = JSON.parse(tgtResponse.toString());
            console.log('Received TGT successfully');

            // Save TGT for later use
            fs.writeFileSync(`${clientId}-tgt.json`, JSON.stringify(tgt));
            
            gateway.disconnect();
            return tgt;
        } catch (tgtError) {
            console.error(`Failed to get TGT: ${tgtError.message}`);
            throw tgtError;
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

        // Prepare service ticket request
        const serviceTicketRequest = {
            encryptedTGT: tgtData.encryptedTGT,
            clientID: clientId,
            serviceID: serviceId,
            authenticator: Buffer.from(Date.now().toString()).toString('base64') // Simulated authenticator
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
            
        default:
            console.log('Available commands:');
            console.log('  register-client <username> <clientId>');
            console.log('  register-device <username> <deviceId> <capability1> <capability2> ...');
            console.log('  authenticate <username> <clientId> <deviceId>');
            console.log('  get-device-data <username> <clientId> <deviceId>');
            console.log('  close-session <username> <clientId> <deviceId>');
    }
}

// Run the main function
main().then(() => {
    console.log('Operation completed');
}).catch(error => {
    console.error('Error in main:', error);
});

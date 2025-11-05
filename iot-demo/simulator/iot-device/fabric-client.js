/**
 * Fabric Client - Hyperledger Fabric SDK Wrapper
 *
 * Provides simplified interface for:
 * - Connecting to Fabric network
 * - Invoking chaincode transactions
 * - Querying chaincode
 * - Managing connection lifecycle
 */

const { Gateway, Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');
const fs = require('fs');

class FabricClient {
    constructor(config) {
        this.config = config;
        this.gateway = null;
        this.wallet = null;
        this.network = null;
        this.contracts = {};
    }

    /**
     * Connect to Fabric network
     */
    async connect() {
        try {
            // Create wallet
            const walletPath = path.join(process.cwd(), 'wallet');
            this.wallet = await Wallets.newFileSystemWallet(walletPath);

            // Check if identity exists
            const identity = await this.wallet.get(this.config.identity);
            if (!identity) {
                console.log(`⚠️  Identity not found. Creating new identity...`);
                await this.enrollUser();
            }

            // Load connection profile
            const ccpPath = this.getCCPPath();
            const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

            // Create gateway
            this.gateway = new Gateway();
            await this.gateway.connect(ccp, {
                wallet: this.wallet,
                identity: this.config.identity,
                discovery: { enabled: true, asLocalhost: true }
            });

            // Get network and contracts
            this.network = await this.gateway.getNetwork(this.config.channelName);

            // Get all chaincode contracts
            const chaincodes = ['as', 'tgs', 'isv', 'user-acl', 'iot-data'];
            for (const cc of chaincodes) {
                try {
                    this.contracts[cc] = this.network.getContract(cc);
                } catch (error) {
                    console.warn(`⚠️  Chaincode '${cc}' not found or not deployed`);
                }
            }

            console.log(`✅ Connected to Fabric network (Channel: ${this.config.channelName})`);

        } catch (error) {
            console.error(`❌ Failed to connect to Fabric network: ${error.message}`);
            throw error;
        }
    }

    /**
     * Enroll user with CA
     */
    async enrollUser() {
        try {
            const ccpPath = this.getCCPPath();
            const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

            // Get CA info
            const caInfo = ccp.certificateAuthorities[this.config.caName];
            const caTLSCACerts = caInfo.tlsCACerts.pem;
            const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

            // Enroll admin
            const enrollment = await ca.enroll({
                enrollmentID: this.config.enrollmentID || 'admin',
                enrollmentSecret: this.config.enrollmentSecret || 'adminpw'
            });

            // Create identity
            const x509Identity = {
                credentials: {
                    certificate: enrollment.certificate,
                    privateKey: enrollment.key.toBytes(),
                },
                mspId: this.config.mspId,
                type: 'X.509',
            };

            await this.wallet.put(this.config.identity, x509Identity);
            console.log(`✅ Identity enrolled and imported to wallet`);

        } catch (error) {
            console.error(`❌ Failed to enroll user: ${error.message}`);
            throw error;
        }
    }

    /**
     * Invoke chaincode transaction (write to ledger)
     */
    async invoke(chaincodeId, functionName, args = []) {
        try {
            const contract = this.contracts[chaincodeId];
            if (!contract) {
                throw new Error(`Chaincode ${chaincodeId} not found`);
            }

            const result = await contract.submitTransaction(functionName, ...args);
            return result.toString();

        } catch (error) {
            console.error(`❌ Invoke failed (${chaincodeId}.${functionName}): ${error.message}`);
            throw error;
        }
    }

    /**
     * Query chaincode (read from ledger)
     */
    async query(chaincodeId, functionName, args = []) {
        try {
            const contract = this.contracts[chaincodeId];
            if (!contract) {
                throw new Error(`Chaincode ${chaincodeId} not found`);
            }

            const result = await contract.evaluateTransaction(functionName, ...args);
            return result.toString();

        } catch (error) {
            console.error(`❌ Query failed (${chaincodeId}.${functionName}): ${error.message}`);
            throw error;
        }
    }

    /**
     * Disconnect from network
     */
    async disconnect() {
        if (this.gateway) {
            await this.gateway.disconnect();
            this.gateway = null;
            this.network = null;
            this.contracts = {};
        }
    }

    /**
     * Get connection profile path
     */
    getCCPPath() {
        // Try common locations
        const possiblePaths = [
            path.resolve(__dirname, '../../../network/organizations/peerOrganizations/org1.example.com/connection-org1.json'),
            path.resolve(__dirname, '../../config/connection-profile.json'),
            path.resolve(__dirname, './connection-profile.json'),
            this.config.connectionProfilePath
        ];

        for (const ccpPath of possiblePaths) {
            if (fs.existsSync(ccpPath)) {
                return ccpPath;
            }
        }

        // If not found, create a default one
        return this.createDefaultConnectionProfile();
    }

    /**
     * Create default connection profile
     */
    createDefaultConnectionProfile() {
        const profile = {
            name: "iot-demo-network",
            version: "1.0.0",
            client: {
                organization: "Org1",
                connection: {
                    timeout: {
                        peer: { endorser: "300" },
                        orderer: "300"
                    }
                }
            },
            channels: {
                authchannel: {
                    orderers: ["orderer.example.com"],
                    peers: {
                        "peer0.org1.example.com": {},
                        "peer1.org1.example.com": {}
                    }
                }
            },
            organizations: {
                Org1: {
                    mspid: "Org1MSP",
                    peers: ["peer0.org1.example.com", "peer1.org1.example.com"],
                    certificateAuthorities: ["ca.org1.example.com"]
                }
            },
            orderers: {
                "orderer.example.com": {
                    url: "grpc://localhost:7050"
                }
            },
            peers: {
                "peer0.org1.example.com": {
                    url: "grpc://localhost:7051"
                },
                "peer1.org1.example.com": {
                    url: "grpc://localhost:8051"
                }
            },
            certificateAuthorities: {
                "ca.org1.example.com": {
                    url: "http://localhost:7054",
                    caName: "ca-org1",
                    tlsCACerts: {
                        pem: ["-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"]
                    }
                }
            }
        };

        const profilePath = path.join(__dirname, 'connection-profile.json');
        fs.writeFileSync(profilePath, JSON.stringify(profile, null, 2));
        console.log(`ℹ️  Created default connection profile: ${profilePath}`);

        return profilePath;
    }
}

module.exports = FabricClient;

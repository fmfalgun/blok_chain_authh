/**
 * Fabric Client for Backend API
 * Same as device simulator but optimized for web backend
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

    async connect() {
        try {
            const walletPath = path.join(process.cwd(), 'wallet');
            this.wallet = await Wallets.newFileSystemWallet(walletPath);

            const identity = await this.wallet.get(this.config.identity);
            if (!identity) {
                await this.enrollAdmin();
            }

            const ccpPath = this.getCCPPath();
            const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

            this.gateway = new Gateway();
            await this.gateway.connect(ccp, {
                wallet: this.wallet,
                identity: this.config.identity,
                discovery: { enabled: true, asLocalhost: true }
            });

            this.network = await this.gateway.getNetwork(this.config.channelName);

            const chaincodes = ['as', 'tgs', 'isv', 'user-acl', 'iot-data'];
            for (const cc of chaincodes) {
                try {
                    this.contracts[cc] = this.network.getContract(cc);
                } catch (error) {
                    console.warn(`Warning: Chaincode '${cc}' not available`);
                }
            }

        } catch (error) {
            console.error('Failed to connect to Fabric:', error);
            throw error;
        }
    }

    async enrollAdmin() {
        const ccpPath = this.getCCPPath();
        const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        const caInfo = ccp.certificateAuthorities[this.config.caName];
        const caTLSCACerts = caInfo.tlsCACerts.pem;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

        const enrollment = await ca.enroll({
            enrollmentID: 'admin',
            enrollmentSecret: 'adminpw'
        });

        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: this.config.mspId,
            type: 'X.509',
        };

        await this.wallet.put(this.config.identity, x509Identity);
    }

    async invoke(chaincodeId, functionName, args = []) {
        const contract = this.contracts[chaincodeId];
        if (!contract) {
            throw new Error(`Chaincode ${chaincodeId} not found`);
        }
        const result = await contract.submitTransaction(functionName, ...args);
        return result.toString();
    }

    async query(chaincodeId, functionName, args = []) {
        const contract = this.contracts[chaincodeId];
        if (!contract) {
            throw new Error(`Chaincode ${chaincodeId} not found`);
        }
        const result = await contract.evaluateTransaction(functionName, ...args);
        return result.toString();
    }

    async disconnect() {
        if (this.gateway) {
            await this.gateway.disconnect();
        }
    }

    getCCPPath() {
        const possiblePaths = [
            path.resolve(__dirname, '../../../network/organizations/peerOrganizations/org1.example.com/connection-org1.json'),
            path.resolve(__dirname, '../../config/connection-profile.json'),
            './connection-profile.json'
        ];

        for (const ccpPath of possiblePaths) {
            if (fs.existsSync(ccpPath)) {
                return ccpPath;
            }
        }

        return this.createDefaultConnectionProfile();
    }

    createDefaultConnectionProfile() {
        const profile = {
            name: "iot-demo-network",
            version: "1.0.0",
            client: {
                organization: "Org1",
                connection: {
                    timeout: { peer: { endorser: "300" }, orderer: "300" }
                }
            },
            channels: {
                authchannel: {
                    orderers: ["orderer.example.com"],
                    peers: { "peer0.org1.example.com": {} }
                }
            },
            organizations: {
                Org1: {
                    mspid: "Org1MSP",
                    peers: ["peer0.org1.example.com"],
                    certificateAuthorities: ["ca.org1.example.com"]
                }
            },
            orderers: {
                "orderer.example.com": { url: "grpc://localhost:7050" }
            },
            peers: {
                "peer0.org1.example.com": { url: "grpc://localhost:7051" }
            },
            certificateAuthorities: {
                "ca.org1.example.com": {
                    url: "http://localhost:7054",
                    caName: "ca-org1",
                    tlsCACerts: { pem: [] }
                }
            }
        };

        const profilePath = path.join(__dirname, 'connection-profile.json');
        fs.writeFileSync(profilePath, JSON.stringify(profile, null, 2));
        return profilePath;
    }
}

module.exports = FabricClient;

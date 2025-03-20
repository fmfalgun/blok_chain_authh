// registerUser.js
// Script to register and enroll a new user with the Certificate Authority
const { Wallets, Gateway } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const fs = require('fs');
const path = require('path');

async function main() {
    try {
        // Get the username from command line argument
        const username = process.argv[2];
        if (!username) {
            console.log('Usage: node registerUser.js <username>');
            process.exit(1);
        }

        // Load the connection profile
        const ccpPath = path.resolve(__dirname, 'connection-profile.json');
        const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new CA client for interacting with the CA
        const caInfo = ccp.certificateAuthorities['ca.org1.example.com'];
        const caTLSCACerts = fs.readFileSync(caInfo.tlsCACerts.path, 'utf8');
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

        // Create a new file system wallet for managing identities
        const walletPath = path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check if user identity exists in the wallet
        const userIdentity = await wallet.get(username);
        if (userIdentity) {
            console.log(`An identity for the user "${username}" already exists in the wallet`);
            return;
        }

        // Check to see if admin identity exists in the wallet
        const adminIdentity = await wallet.get('admin');
        if (!adminIdentity) {
            console.log('An identity for the admin user "admin" does not exist in the wallet');
            console.log('Run the enrollAdmin.js application before retrying');
            return;
        }

        // Build a user object for authenticating with the CA
        const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
        const adminUser = await provider.getUserContext(adminIdentity, 'admin');

        // Register the user, enroll the user, and import the new identity into the wallet
        const secret = await ca.register({
            affiliation: 'org1.department1',
            enrollmentID: username,
            role: 'client'
        }, adminUser);
        
        const enrollment = await ca.enroll({
            enrollmentID: username,
            enrollmentSecret: secret
        });
        
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: 'Org1MSP',
            type: 'X.509',
        };
        
        await wallet.put(username, x509Identity);
        console.log(`Successfully registered and enrolled user "${username}" and imported it into the wallet`);

    } catch (error) {
        console.error(`Failed to register user "${process.argv[2]}": ${error}`);
        process.exit(1);
    }
}

main();

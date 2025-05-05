#!/bin/bash

# Network Status Check Script
# This script checks the status of the Hyperledger Fabric network

# Define color codes for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function for printing colored output
print_green() {
    echo -e "${GREEN}$1${NC}"
}

print_yellow() {
    echo -e "${YELLOW}$1${NC}"
}

print_red() {
    echo -e "${RED}$1${NC}"
}

print_yellow "Checking Hyperledger Fabric Network Status"
echo "==============================================="

# Check if Docker is running
print_yellow "Checking Docker service..."
if systemctl is-active docker >/dev/null 2>&1 || docker info >/dev/null 2>&1; then
    print_green "✓ Docker is running"
else
    print_red "✗ Docker is not running. Please start the Docker service."
    exit 1
fi

# Check network containers
print_yellow "Checking network containers..."

# Define expected containers
EXPECTED_CONTAINERS=(
    "orderer.example.com"
    "peer0.org1.example.com"
    "peer1.org1.example.com"
    "peer2.org1.example.com"
    "peer0.org2.example.com"
    "peer1.org2.example.com"
    "peer2.org2.example.com"
    "peer0.org3.example.com"
    "peer1.org3.example.com"
    "peer2.org3.example.com"
    "ca.org1.example.com"
    "ca.org2.example.com"
    "ca.org3.example.com"
    "cli"
)

MISSING_CONTAINERS=0

for container in "${EXPECTED_CONTAINERS[@]}"; do
    if docker ps -q --filter "name=${container}$" | grep -q .; then
        print_green "✓ ${container} is running"
    else
        print_red "✗ ${container} is not running"
        MISSING_CONTAINERS=$((MISSING_CONTAINERS + 1))
    fi
done

# Check chaincodes
print_yellow "Checking chaincode containers..."

CHAINCODE_CONTAINERS=$(docker ps -q --filter "name=dev-peer.*chaichis-channel" | wc -l)

if [ $CHAINCODE_CONTAINERS -ge 3 ]; then
    print_green "✓ $CHAINCODE_CONTAINERS chaincode containers are running"
else
    print_yellow "Only $CHAINCODE_CONTAINERS chaincode containers are running (expected at least 3)"
fi

# Check channel and chaincode status using CLI container
print_yellow "Checking channel and chaincode status..."

# Try to run peer channel list command
CHANNEL_LIST=$(docker exec cli peer channel list 2>/dev/null)
if [[ $CHANNEL_LIST == *"chaichis-channel"* ]]; then
    print_green "✓ chaichis-channel exists"
else
    print_red "✗ chaichis-channel does not exist or cannot be accessed"
fi

# Try to list installed chaincodes
print_yellow "Installed chaincodes:"
docker exec cli peer chaincode list --installed 2>/dev/null || print_red "Unable to list installed chaincodes"

print_yellow "Instantiated chaincodes on chaichis-channel:"
docker exec cli peer chaincode list --channelID chaichis-channel --instantiated 2>/dev/null || print_red "Unable to list instantiated chaincodes"

# Check network connection from JavaScript SDK
print_yellow "Checking network connection from JavaScript SDK..."

# Create temporary test script
cat > temp_network_test.js << EOF
const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');

async function testNetworkConnection() {
    try {
        // Load connection profile
        const ccpPath = path.resolve(__dirname, 'connection-profile.json');
        const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

        // Create a new file system wallet
        const walletPath = path.resolve(__dirname, 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);

        // Check if admin identity exists
        const identity = await wallet.get('admin');
        if (!identity) {
            console.log('Admin identity not found in the wallet');
            return false;
        }

        // Create a new gateway
        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet,
            identity: 'admin',
            discovery: { enabled: true, asLocalhost: true }
        });

        // Get the network
        const network = await gateway.getNetwork('chaichis-channel');
        
        // Try to get one of the contracts
        const asContract = network.getContract('as-chaincode');
        
        console.log('Successfully connected to the network and retrieved AS contract');
        
        // Disconnect
        gateway.disconnect();
        return true;
    } catch (error) {
        console.error('Error testing network connection:', error);
        return false;
    }
}

testNetworkConnection()
    .then(result => {
        console.log('Connection test result:', result ? 'SUCCESS' : 'FAILED');
        process.exit(result ? 0 : 1);
    })
    .catch(error => {
        console.error('Test threw an exception:', error);
        process.exit(1);
    });
EOF

# Run the test script
node temp_network_test.js
if [ $? -eq 0 ]; then
    print_green "✓ Successfully connected to the network from JavaScript SDK"
else
    print_red "✗ Failed to connect to the network from JavaScript SDK"
fi

# Clean up
rm temp_network_test.js

# Summary
echo "==============================================="
if [ $MISSING_CONTAINERS -eq 0 ] && [[ $CHANNEL_LIST == *"chaichis-channel"* ]]; then
    print_green "Network Status: HEALTHY"
    print_green "The Hyperledger Fabric network is running properly."
elif [ $MISSING_CONTAINERS -gt 0 ] && [ $MISSING_CONTAINERS -lt ${#EXPECTED_CONTAINERS[@]} ]; then
    print_yellow "Network Status: PARTIALLY RUNNING"
    print_yellow "Some containers are missing, but the network may still be functional."
else
    print_red "Network Status: NOT RUNNING"
    print_red "The network is not running properly. Please restart it."
fi

echo "==============================================="
print_yellow "To restart the network:"
echo "cd /home/fm/projects/blok_chain_authh"
echo "./start-network.sh"

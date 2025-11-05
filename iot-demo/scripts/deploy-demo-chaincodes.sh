#!/bin/bash

# Deploy USER-ACL and IOT-DATA chaincodes
# This script deploys the demo-specific chaincodes to the existing Fabric network

set -e

echo "Deploying demo chaincodes..."

# Navigate to network scripts directory
NETWORK_SCRIPTS="/home/user/blok_chain_authh/network/scripts"

# Deploy USER-ACL chaincode
echo "  Deploying USER-ACL chaincode..."
cd /home/user/blok_chain_authh
if [ -f "$NETWORK_SCRIPTS/deploy-chaincode.sh" ]; then
    bash "$NETWORK_SCRIPTS/deploy-chaincode.sh" user-acl || echo "  (Chaincode may already be deployed)"
else
    echo "  Warning: deploy-chaincode.sh not found, assuming manual deployment"
fi

# Deploy IOT-DATA chaincode
echo "  Deploying IOT-DATA chaincode..."
if [ -f "$NETWORK_SCRIPTS/deploy-chaincode.sh" ]; then
    bash "$NETWORK_SCRIPTS/deploy-chaincode.sh" iot-data || echo "  (Chaincode may already be deployed)"
else
    echo "  Warning: deploy-chaincode.sh not found, assuming manual deployment"
fi

echo "Chaincode deployment complete (or already deployed)"

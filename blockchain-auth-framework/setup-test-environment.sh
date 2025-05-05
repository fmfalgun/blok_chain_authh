#!/bin/bash

# Setup Test Environment Script
# This script prepares the environment for testing the authentication system

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

print_yellow "Setting up the test environment"
echo "-----------------------------------------"

# Step 1: Make all test scripts executable
print_yellow "Step 1: Making all test scripts executable"
chmod +x test-authentication-flow.sh
chmod +x test-rsa-keys.sh
chmod +x auth-cli.sh
chmod +x simple-auth.sh

if [ $? -eq 0 ]; then
    print_green "✓ All scripts are now executable"
else
    print_red "✗ Failed to make scripts executable"
    exit 1
fi

# Step 2: Check if the network is running
print_yellow "Step 2: Checking if the Hyperledger Fabric network is running"

RUNNING_CONTAINERS=$(docker ps -f name=peer --format '{{.Names}}' | wc -l)

if [ $RUNNING_CONTAINERS -gt 0 ]; then
    print_green "✓ Hyperledger Fabric network is running ($RUNNING_CONTAINERS peer containers found)"
else
    print_red "✗ Hyperledger Fabric network is not running"
    print_yellow "Please start the network using: cd /home/fm/projects/blok_chain_authh && ./start-network.sh"
    exit 1
fi

# Step 3: Check if wallet directory exists and has admin identity
print_yellow "Step 3: Checking admin identity in wallet"

if [ -d "wallet" ] && [ -d "wallet/admin" ]; then
    print_green "✓ Admin identity found in wallet directory"
else
    print_yellow "Admin identity not found in wallet. Running enrollAdmin.js..."
    node enrollAdmin.js
    
    if [ $? -eq 0 ]; then
        print_green "✓ Admin enrolled successfully"
    else
        print_red "✗ Failed to enroll admin"
        exit 1
    fi
fi

# Step 4: Clean up any previous test artifacts
print_yellow "Step 4: Cleaning up previous test artifacts"

# Remove old private keys and session files
rm -f test-*-private.pem client_test_*-private.pem device_test_*-private.pem
rm -f client_test_*-tgt.json client_test_*-serviceticket-*.json
rm -f client_test_*-session-*.txt

print_green "✓ Previous test artifacts cleaned up"

# Step 5: Create a clean test directory
print_yellow "Step 5: Creating clean test directory"

TEST_DIR="test-results-$(date +%Y%m%d%H%M%S)"
mkdir -p $TEST_DIR

if [ $? -eq 0 ]; then
    print_green "✓ Test directory created: $TEST_DIR"
else
    print_red "✗ Failed to create test directory"
    exit 1
fi

# Step 6: Create a README for the test environment
print_yellow "Step 6: Creating README for the test environment"

cat > $TEST_DIR/README.md << EOF
# Blockchain Authentication Framework Test Results

This directory contains test results for the Blockchain Authentication Framework.

## Test Scripts

1. **test-authentication-flow.sh** - Tests the complete authentication flow
2. **test-rsa-keys.sh** - Tests RSA key generation and validation
3. **auth-cli.sh** - CLI interface for authentication operations
4. **simple-auth.sh** - Simplified interface for authentication operations

## Running Tests

To run a test, execute the corresponding script:

\`\`\`bash
./test-authentication-flow.sh
\`\`\`

Test results will be stored in this directory.

## Test Environment

- Date: $(date)
- Hyperledger Fabric Network: Running
- Chaincodes: as-chaincode, tgs-chaincode, isv-chaincode
EOF

print_green "✓ README created for the test environment"

print_green "Test Environment Setup Completed Successfully!"
echo "✓ Scripts made executable"
echo "✓ Network status verified"
echo "✓ Admin identity checked"
echo "✓ Previous artifacts cleaned up"
echo "✓ Test directory created: $TEST_DIR"
echo "✓ README created"

print_yellow "You can now run the test scripts:"
echo "./test-authentication-flow.sh"
echo "./test-rsa-keys.sh"

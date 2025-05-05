#!/bin/bash

# Authentication Flow Test Script
# This script tests the complete authentication flow from client registration to device access

# Exit on any error
set -e

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

# Function for testing steps with validation
test_step() {
    local step_num=$1
    local description=$2
    local command=$3
    
    print_yellow "Step $step_num: $description"
    echo "Command: $command"
    
    eval $command
    
    if [ $? -eq 0 ]; then
        print_green "✓ Step $step_num completed successfully"
    else
        print_red "✗ Step $step_num failed"
        exit 1
    fi
    
    echo "-----------------------------------------"
}

# Generate unique identifiers for this test run
TIMESTAMP=$(date +%Y%m%d%H%M%S)
CLIENT_ID="client_test_${TIMESTAMP}"
DEVICE_ID="device_test_${TIMESTAMP}"
USER="admin"

print_yellow "Starting authentication flow test with:"
echo "Client ID: $CLIENT_ID"
echo "Device ID: $DEVICE_ID"
echo "User: $USER"
echo "-----------------------------------------"

# Step 1: Register a client
test_step 1 "Registering client" "node auth-framework.js register-client $USER $CLIENT_ID"

# Step 2: Register an IoT device
test_step 2 "Registering IoT device" "node auth-framework.js register-device $USER $DEVICE_ID temperature humidity light"

# Step 3: Get Ticket Granting Ticket from Authentication Server
test_step 3 "Getting TGT from Authentication Server" "node auth-framework.js getTGT $USER $CLIENT_ID"

# Check if TGT file exists
if [ ! -f "${CLIENT_ID}-tgt.json" ]; then
    print_red "TGT file not found: ${CLIENT_ID}-tgt.json"
    exit 1
fi

# Step 4: Get Service Ticket from Ticket Granting Server
test_step 4 "Getting Service Ticket from TGS" "node auth-framework.js getServiceTicket $USER $CLIENT_ID iotservice1"

# Check if service ticket file exists
if [ ! -f "${CLIENT_ID}-serviceticket-iotservice1.json" ]; then
    print_red "Service ticket file not found: ${CLIENT_ID}-serviceticket-iotservice1.json"
    exit 1
fi

# Step 5: Access IoT device through ISV
test_step 5 "Accessing IoT device through ISV" "node auth-framework.js accessIoTDevice $USER $CLIENT_ID $DEVICE_ID"

# Check if session file exists
if [ ! -f "${CLIENT_ID}-session-${DEVICE_ID}.txt" ]; then
    print_red "Session file not found: ${CLIENT_ID}-session-${DEVICE_ID}.txt"
    exit 1
fi

# Step 6: Get IoT device data using established session
test_step 6 "Getting IoT device data" "node auth-framework.js getIoTDeviceData $USER $CLIENT_ID $DEVICE_ID"

# Step 7: Close the session
test_step 7 "Closing session" "node auth-framework.js closeSession $USER $CLIENT_ID $DEVICE_ID"

# Check that session file is removed
if [ -f "${CLIENT_ID}-session-${DEVICE_ID}.txt" ]; then
    print_red "Session file still exists after closing: ${CLIENT_ID}-session-${DEVICE_ID}.txt"
    exit 1
fi

print_green "Authentication Flow Test Completed Successfully!"
echo "✓ Client registration"
echo "✓ Device registration"
echo "✓ TGT acquisition"
echo "✓ Service ticket acquisition"
echo "✓ Device access"
echo "✓ Data retrieval"
echo "✓ Session termination"

# Clean up test files (optional, comment out to keep files for inspection)
rm -f ${CLIENT_ID}-private.pem
rm -f ${CLIENT_ID}-tgt.json
rm -f ${CLIENT_ID}-serviceticket-*.json
rm -f ${DEVICE_ID}-private.pem

print_yellow "Test artifacts cleaned up"

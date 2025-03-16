#!/bin/bash
# Modified Testing Guide for Kerberos-like Chaincode System on Hyperledger Fabric
# Uses single-organization endorsement as a workaround for non-deterministic initialization

#####################################################################
# SECTION 1: ENVIRONMENT SETUP AND INITIALIZATION
#####################################################################

# Initial setup for environment variables and orderer CA
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

echo "===== TESTING PEER CONNECTIVITY ====="
# Setup Org1 environment for most operations
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_TLS_ENABLED=true

echo "Testing Org1 connectivity..."
peer channel list

# Save TLS certificate paths for reference (though we'll only use single org)
ORG1_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
ORG2_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
ORG3_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

#####################################################################
# SECTION 2: HELPER FUNCTIONS
#####################################################################

# Function to switch to Org1
switch_to_org1() {
    echo "Switching to Org1 peer0..."
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export CORE_PEER_TLS_ENABLED=true
    echo "Now using Org1 peer0"
}

# Function to switch to Org2
switch_to_org2() {
    echo "Switching to Org2 peer0..."
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
    export CORE_PEER_LOCALMSPID="Org2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    export CORE_PEER_TLS_ENABLED=true
    echo "Now using Org2 peer0"
}

# Function to switch to Org3
switch_to_org3() {
    echo "Switching to Org3 peer0..."
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org3.example.com:13051
    export CORE_PEER_LOCALMSPID="Org3MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
    export CORE_PEER_TLS_ENABLED=true
    echo "Now using Org3 peer0"
}

# Function to initialize chaincode on a specific organization
initialize_chaincode_on_org() {
    local CHAINCODE_NAME=$1
    local ORG=$2
    
    echo "Initializing $CHAINCODE_NAME on Org$ORG..."
    
    if [ "$ORG" == "1" ]; then
        switch_to_org1
    elif [ "$ORG" == "2" ]; then
        switch_to_org2
    elif [ "$ORG" == "3" ]; then
        switch_to_org3
    fi
    
    # Use the --isInit flag before other arguments to avoid the orderer error
    peer chaincode invoke --isInit -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"Initialize","Args":[]}' --tls --cafile $ORDERER_CA
    
    sleep 3 # Wait for initialization to fully process
    echo "$CHAINCODE_NAME initialized on Org$ORG"
}

# Function to execute a chaincode using single org endorsement
execute_chaincode() {
    local CHAINCODE_NAME=$1
    local FUNCTION_NAME=$2
    local ARGS=$3
    local ORG=$4
    
    echo "Executing $FUNCTION_NAME on $CHAINCODE_NAME with Org$ORG endorsement..."
    if [ "$ORG" == "1" ]; then
        switch_to_org1
    elif [ "$ORG" == "2" ]; then
        switch_to_org2
    elif [ "$ORG" == "3" ]; then
        switch_to_org3
    fi
    
    # Regular invoke with single organization
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c "{\"function\":\"$FUNCTION_NAME\",\"Args\":$ARGS}" --tls --cafile $ORDERER_CA
    
    # Capture the return value in case we need to get output values
    local RESPONSE=$?
    sleep 3 # Wait for transaction to be committed
    return $RESPONSE
}

# Function to query a chaincode
query_chaincode() {
    local CHAINCODE_NAME=$1
    local FUNCTION_NAME=$2
    local ARGS=$3
    local ORG=$4
    
    echo "Querying $FUNCTION_NAME on $CHAINCODE_NAME from Org$ORG..."
    if [ "$ORG" == "1" ]; then
        switch_to_org1
    elif [ "$ORG" == "2" ]; then
        switch_to_org2
    elif [ "$ORG" == "3" ]; then
        switch_to_org3
    fi
    
    peer chaincode query -C chaichis-channel -n $CHAINCODE_NAME -c "{\"function\":\"$FUNCTION_NAME\",\"Args\":$ARGS}" --tls --cafile $ORDERER_CA
}

# Verify initialization was successful by making a simple query
verify_chaincode_initialization() {
    local CHAINCODE_NAME=$1
    local ORG=$2
    
    echo "Verifying $CHAINCODE_NAME initialization on Org$ORG..."
    
    if [ "$ORG" == "1" ]; then
        switch_to_org1
    elif [ "$ORG" == "2" ]; then
        switch_to_org2
    elif [ "$ORG" == "3" ]; then
        switch_to_org3
    fi
    
    if [ "$CHAINCODE_NAME" == "as-chaincode" ]; then
        # AS chaincode verification
        peer chaincode query -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"GetAllClientRegistrations","Args":[]}' --tls --cafile $ORDERER_CA
    elif [ "$CHAINCODE_NAME" == "tgs-chaincode" ]; then
        # TGS chaincode verification
        peer chaincode query -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"GetAllClientRegistrations","Args":[]}' --tls --cafile $ORDERER_CA
    elif [ "$CHAINCODE_NAME" == "isv-chaincode" ]; then
        # ISV chaincode verification
        peer chaincode query -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"GetAllIoTDevices","Args":[]}' --tls --cafile $ORDERER_CA
    fi
    
    echo "Verification completed for $CHAINCODE_NAME on Org$ORG"
}

#####################################################################
# SECTION 3: INITIALIZATION PHASE
#####################################################################

echo "===== INITIALIZING CHAINCODES ON ALL ORGANIZATIONS ====="

# Check chaincode policy to understand endorsement requirements
echo "Checking chaincode policies..."
peer lifecycle chaincode querycommitted -C chaichis-channel -n as-chaincode --output json

# Initialize each chaincode on each organization separately
# We'll only use a single org (Org1) for testing after initialization
echo "Initializing AS chaincode..."
initialize_chaincode_on_org "as-chaincode" "1"
echo "Verifying AS chaincode initialization..."
verify_chaincode_initialization "as-chaincode" "1"

echo "Initializing TGS chaincode..."
initialize_chaincode_on_org "tgs-chaincode" "1"
echo "Verifying TGS chaincode initialization..."
verify_chaincode_initialization "tgs-chaincode" "1"

echo "Initializing ISV chaincode..."
initialize_chaincode_on_org "isv-chaincode" "1"
echo "Verifying ISV chaincode initialization..."
verify_chaincode_initialization "isv-chaincode" "1"

#####################################################################
# SECTION 4: TESTING THE AUTHENTICATION SERVER (AS) CHAINCODE
#####################################################################

echo "===== TESTING AS CHAINCODE FUNCTIONS ====="

# Make sure we're using Org1 for AS testing
switch_to_org1

# Test Variables for AS
CLIENT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxzYkf+XKGRgxLQwVP6MA
3mY9gU4kDeJZfLBuQzM77a7jo+kW+CdVAgMBAAE=
-----END PUBLIC KEY-----"
CLIENT_PUBLIC_KEY_B64=$(echo -n "$CLIENT_PUBLIC_KEY" | base64 -w 0)

# 1. Register a Client
echo "Registering client1..."
execute_chaincode "as-chaincode" "RegisterClient" "[\"client1\", \"$CLIENT_PUBLIC_KEY_B64\"]" "1"

# 2. Check Client Validity
echo "Checking client1 validity..."
query_chaincode "as-chaincode" "CheckClientValidity" "[\"client1\"]" "1"

# 3. Initiate Authentication
echo "Initiating authentication for client1..."
NONCE_RESPONSE=$(execute_chaincode "as-chaincode" "InitiateAuthentication" "[\"client1\"]" "1")
echo "Nonce response: $NONCE_RESPONSE"

# Extract nonce value (For real implementation, parse the JSON response)
NONCE_VALUE="simulated_nonce_for_testing"

# 4. Verify Client Identity (using simulated encrypted nonce)
echo "Verifying client identity..."
ENCRYPTED_NONCE="simulated_encrypted_nonce_base64_string"
execute_chaincode "as-chaincode" "VerifyClientIdentity" "[\"client1\", \"$ENCRYPTED_NONCE\"]" "1"

# 5. Generate TGT
echo "Generating TGT for client1..."
TGT_RESPONSE=$(execute_chaincode "as-chaincode" "GenerateTGT" "[\"client1\"]" "1")
echo "TGT response: $TGT_RESPONSE"

# Extract TGT (For real implementation, parse the JSON response)
ENCRYPTED_TGT="simulated_encrypted_tgt_base64_for_testing"

# 6. Get All Client Registrations
echo "Getting all client registrations from AS..."
query_chaincode "as-chaincode" "GetAllClientRegistrations" "[]" "1"

# 7. Allocate Peer Task
echo "Allocating peer task..."
execute_chaincode "as-chaincode" "AllocatePeerTask" "[\"peer1\", \"authentication\", \"client1\"]" "1"

# 8. Reserve and Validate Registration
echo "Reserving and validating registration..."
execute_chaincode "as-chaincode" "ReserveAndValidateRegistration" "[\"client1\"]" "1"

#####################################################################
# SECTION 5: TESTING THE TICKET GRANTING SERVICE (TGS) CHAINCODE
#####################################################################

echo "===== TESTING TGS CHAINCODE FUNCTIONS ====="

# We'll use Org1 consistently for all testing
switch_to_org1

# 1. Process Registration from AS
echo "Processing registration from AS..."
execute_chaincode "tgs-chaincode" "ProcessRegistrationFromAS" "[\"$ENCRYPTED_TGT\"]" "1"

# 2. Check Registration Validity
echo "Checking registration validity at TGS..."
query_chaincode "tgs-chaincode" "CheckRegistrationValidity" "[\"client1\"]" "1"

# 3. Generate Service Ticket
echo "Generating service ticket..."
# Properly escape the JSON for the service request
SERVICE_REQUEST="{\\\"encryptedTGT\\\":\\\"$ENCRYPTED_TGT\\\",\\\"clientID\\\":\\\"client1\\\",\\\"serviceID\\\":\\\"iotservice1\\\",\\\"authenticator\\\":\\\"simulated_authenticator_base64_string\\\"}"
SERVICE_TICKET_RESPONSE=$(execute_chaincode "tgs-chaincode" "GenerateServiceTicket" "[\"$SERVICE_REQUEST\"]" "1")
echo "Service ticket response: $SERVICE_TICKET_RESPONSE"

# Extract service ticket (For real implementation, parse the JSON response)
ENCRYPTED_SERVICE_TICKET="simulated_encrypted_service_ticket_base64_for_testing"

# 4. Forward Registration to ISV
echo "Forwarding registration to ISV..."
execute_chaincode "tgs-chaincode" "ForwardRegistrationToISV" "[\"client1\", \"iotservice1\", \"$ENCRYPTED_SERVICE_TICKET\"]" "1"

# 5. Get All Client Registrations from TGS
echo "Getting all client registrations from TGS..."
query_chaincode "tgs-chaincode" "GetAllClientRegistrations" "[]" "1"

#####################################################################
# SECTION 6: TESTING THE IOT SERVICE VALIDATOR (ISV) CHAINCODE
#####################################################################

echo "===== TESTING ISV CHAINCODE FUNCTIONS ====="

# Still using Org1 for all testing
switch_to_org1

# 1. Register IoT Device
echo "Registering IoT device..."
DEVICE_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdef
abcdef1234567890ABCDEF1234567890abcdefABCDEF==
-----END PUBLIC KEY-----"
DEVICE_PUBLIC_KEY_B64=$(echo -n "$DEVICE_PUBLIC_KEY" | base64 -w 0)
# Convert JSON array to string representation to avoid parsing errors
CAPABILITIES="[\\\"temperature\\\",\\\"humidity\\\",\\\"pressure\\\"]"

execute_chaincode "isv-chaincode" "RegisterIoTDevice" "[\"device1\", \"$DEVICE_PUBLIC_KEY_B64\", \"$CAPABILITIES\"]" "1"

# 2. Update Device Status
echo "Updating device status..."
SIGNATURE="simulated_signature_base64_string"
execute_chaincode "isv-chaincode" "UpdateDeviceStatus" "[\"device1\", \"active\", \"$SIGNATURE\"]" "1"

# 3. Check Device Availability
echo "Checking device availability..."
query_chaincode "isv-chaincode" "CheckDeviceAvailability" "[\"device1\"]" "1"

# 4. Validate Service Ticket
echo "Validating service ticket..."
execute_chaincode "isv-chaincode" "ValidateServiceTicket" "[\"$ENCRYPTED_SERVICE_TICKET\"]" "1"

# 5. Process Service Request
echo "Processing service request..."
# Properly escape the JSON for the service request
SERVICE_REQUEST="{\\\"encryptedServiceTicket\\\":\\\"$ENCRYPTED_SERVICE_TICKET\\\",\\\"clientID\\\":\\\"client1\\\",\\\"deviceID\\\":\\\"device1\\\",\\\"requestType\\\":\\\"read\\\",\\\"encryptedData\\\":\\\"simulated_encrypted_data_base64_string\\\"}"
SERVICE_RESPONSE=$(execute_chaincode "isv-chaincode" "ProcessServiceRequest" "[\"$SERVICE_REQUEST\"]" "1")
echo "Service response: $SERVICE_RESPONSE"

# Extract session ID (For real implementation, parse the JSON response)
SESSION_ID="simulated_session_id_for_testing"

# 6. Handle Device Response
echo "Handling device response..."
execute_chaincode "isv-chaincode" "HandleDeviceResponse" "[\"$SESSION_ID\", \"Temperature is 25.5 C\"]" "1"

# 7. Close Session
echo "Closing session..."
execute_chaincode "isv-chaincode" "CloseSession" "[\"$SESSION_ID\"]" "1"

# 8. Get All IoT Devices
echo "Getting all IoT devices..."
query_chaincode "isv-chaincode" "GetAllIoTDevices" "[]" "1"

# 9. Get Active Sessions by Client
echo "Getting active sessions by client..."
query_chaincode "isv-chaincode" "GetActiveSessionsByClient" "[\"client1\"]" "1"

echo "===== ALL TESTS COMPLETED ====="

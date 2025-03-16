#!/bin/bash
# Comprehensive Testing Guide for Kerberos-like Chaincode System on Hyperledger Fabric

#####################################################################
# SECTION 1: ENVIRONMENT SETUP AND INITIALIZATION
#####################################################################

# Enter the CLI container (run this outside the container)
# docker exec -it cli bash

# Initial setup for environment variables and orderer CA
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

#####################################################################
# SECTION 2: PRE-TESTING VALIDATION AND INITIALIZATION
#####################################################################

# This section ensures all chaincodes are properly initialized on all peers
# before we begin testing functionality

echo "===== VERIFYING CHAINCODE INSTALLATION AND INITIALIZATION ====="

# Function to save TLS cert paths for each organization's peers for endorsement
save_tls_cert_paths() {
    ORG1_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    ORG2_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    ORG3_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
    
    echo "TLS certificate paths saved for all organizations"
}

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
    export CORE_PEER_ADDRESS=peer0.org3.example.com:11051
    export CORE_PEER_LOCALMSPID="Org3MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
    export CORE_PEER_TLS_ENABLED=true
    echo "Now using Org3 peer0"
}

# Function to initialize chaincode on all peers of all organizations
initialize_chaincode() {
    CHAINCODE_NAME=$1
    
    echo "===== INITIALIZING $CHAINCODE_NAME ON ALL ORGANIZATIONS ====="
    
    # Check if chaincode is committed and endorsement policy
    echo "Checking chaincode policy for $CHAINCODE_NAME..."
    switch_to_org1
    peer lifecycle chaincode querycommitted -C chaichis-channel -n $CHAINCODE_NAME --output json
    
    # Initialize on Org1
    echo "Initializing $CHAINCODE_NAME on Org1..."
    switch_to_org1
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"Initialize","Args":[]}' --tls --cafile $ORDERER_CA --isInit
    sleep 2
    
    # Initialize on Org2
    echo "Initializing $CHAINCODE_NAME on Org2..."
    switch_to_org2
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"Initialize","Args":[]}' --tls --cafile $ORDERER_CA --isInit
    sleep 2
    
    # Initialize on Org3
    echo "Initializing $CHAINCODE_NAME on Org3..."
    switch_to_org3
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"Initialize","Args":[]}' --tls --cafile $ORDERER_CA --isInit
    sleep 2
    
    echo "$CHAINCODE_NAME initialized on all organizations"
}

# Function to execute a chaincode function with multi-org endorsement
execute_with_endorsement() {
    CHAINCODE_NAME=$1
    FUNCTION_NAME=$2
    ARGS=$3
    
    echo "Executing $FUNCTION_NAME on $CHAINCODE_NAME with multi-org endorsement..."
    switch_to_org1
    
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c "{\"function\":\"$FUNCTION_NAME\",\"Args\":$ARGS}" \
        --tls --cafile $ORDERER_CA \
        --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
        --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
        --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT
    
    # Capture the return value in case we need to get output values
    RESPONSE=$?
    sleep 2
    return $RESPONSE
}

# Function to query a chaincode
query_chaincode() {
    CHAINCODE_NAME=$1
    FUNCTION_NAME=$2
    ARGS=$3
    ORG=$4
    
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

# Save TLS cert paths for multi-org endorsement
save_tls_cert_paths

# Initialize all chaincodes on all organizations
initialize_chaincode "as-chaincode"
initialize_chaincode "tgs-chaincode"
initialize_chaincode "isv-chaincode"

#####################################################################
# SECTION 3: TESTING THE AUTHENTICATION SERVER (AS) CHAINCODE
#####################################################################

echo "===== TESTING AS CHAINCODE FUNCTIONS ====="

# Switch back to Org1 for AS testing
switch_to_org1

# Test Variables for AS
CLIENT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxzYkf+XKGRgxLQwVP6MA
3mY9gU4kDeJZfLBuQzM77a7jo+kW+CdVAgMBAAE=
-----END PUBLIC KEY-----"
CLIENT_PUBLIC_KEY_B64=$(echo -n "$CLIENT_PUBLIC_KEY" | base64 -w 0)

# 1. Register a Client
echo "Registering client1..."
execute_with_endorsement "as-chaincode" "RegisterClient" "[\"client1\", \"$CLIENT_PUBLIC_KEY_B64\"]"

# 2. Check Client Validity
echo "Checking client1 validity..."
query_chaincode "as-chaincode" "CheckClientValidity" "[\"client1\"]" "1"

# 3. Initiate Authentication
echo "Initiating authentication for client1..."
NONCE_RESPONSE=$(execute_with_endorsement "as-chaincode" "InitiateAuthentication" "[\"client1\"]")
echo "Nonce response: $NONCE_RESPONSE"

# Extract nonce value (For real implementation, parse the JSON response)
NONCE_VALUE="simulated_nonce_for_testing"

# 4. Verify Client Identity (using simulated encrypted nonce)
echo "Verifying client identity..."
ENCRYPTED_NONCE="simulated_encrypted_nonce_base64_string"
execute_with_endorsement "as-chaincode" "VerifyClientIdentity" "[\"client1\", \"$ENCRYPTED_NONCE\"]"

# 5. Generate TGT
echo "Generating TGT for client1..."
TGT_RESPONSE=$(execute_with_endorsement "as-chaincode" "GenerateTGT" "[\"client1\"]")
echo "TGT response: $TGT_RESPONSE"

# Extract TGT (For real implementation, parse the JSON response)
ENCRYPTED_TGT="simulated_encrypted_tgt_base64_for_testing"

# 6. Get All Client Registrations
echo "Getting all client registrations from AS..."
query_chaincode "as-chaincode" "GetAllClientRegistrations" "[]" "1"

# 7. Allocate Peer Task
echo "Allocating peer task..."
execute_with_endorsement "as-chaincode" "AllocatePeerTask" "[\"peer1\", \"authentication\", \"client1\"]"

# 8. Reserve and Validate Registration
echo "Reserving and validating registration..."
execute_with_endorsement "as-chaincode" "ReserveAndValidateRegistration" "[\"client1\"]"

#####################################################################
# SECTION 4: TESTING THE TICKET GRANTING SERVICE (TGS) CHAINCODE
#####################################################################

echo "===== TESTING TGS CHAINCODE FUNCTIONS ====="

# Switch to Org2 for TGS testing
switch_to_org2

# 1. Process Registration from AS
echo "Processing registration from AS..."
execute_with_endorsement "tgs-chaincode" "ProcessRegistrationFromAS" "[\"$ENCRYPTED_TGT\"]"

# 2. Check Registration Validity
echo "Checking registration validity at TGS..."
query_chaincode "tgs-chaincode" "CheckRegistrationValidity" "[\"client1\"]" "2"

# 3. Generate Service Ticket
echo "Generating service ticket..."
SERVICE_REQUEST="{\"encryptedTGT\": \"$ENCRYPTED_TGT\", \"clientID\": \"client1\", \"serviceID\": \"iotservice1\", \"authenticator\": \"simulated_authenticator_base64_string\"}"
SERVICE_TICKET_RESPONSE=$(execute_with_endorsement "tgs-chaincode" "GenerateServiceTicket" "[\"$SERVICE_REQUEST\"]")
echo "Service ticket response: $SERVICE_TICKET_RESPONSE"

# Extract service ticket (For real implementation, parse the JSON response)
ENCRYPTED_SERVICE_TICKET="simulated_encrypted_service_ticket_base64_for_testing"

# 4. Forward Registration to ISV
echo "Forwarding registration to ISV..."
execute_with_endorsement "tgs-chaincode" "ForwardRegistrationToISV" "[\"client1\", \"iotservice1\", \"$ENCRYPTED_SERVICE_TICKET\"]"

# 5. Get All Client Registrations from TGS
echo "Getting all client registrations from TGS..."
query_chaincode "tgs-chaincode" "GetAllClientRegistrations" "[]" "2"

#####################################################################
# SECTION 5: TESTING THE IOT SERVICE VALIDATOR (ISV) CHAINCODE
#####################################################################

echo "===== TESTING ISV CHAINCODE FUNCTIONS ====="

# Switch to Org3 for ISV testing
switch_to_org3

# 1. Register IoT Device
echo "Registering IoT device..."
DEVICE_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdef
abcdef1234567890ABCDEF1234567890abcdefABCDEF==
-----END PUBLIC KEY-----"
DEVICE_PUBLIC_KEY_B64=$(echo -n "$DEVICE_PUBLIC_KEY" | base64 -w 0)
CAPABILITIES='["temperature","humidity","pressure"]'

execute_with_endorsement "isv-chaincode" "RegisterIoTDevice" "[\"device1\", \"$DEVICE_PUBLIC_KEY_B64\", $CAPABILITIES]"

# 2. Update Device Status
echo "Updating device status..."
SIGNATURE="simulated_signature_base64_string"
execute_with_endorsement "isv-chaincode" "UpdateDeviceStatus" "[\"device1\", \"active\", \"$SIGNATURE\"]"

# 3. Check Device Availability
echo "Checking device availability..."
query_chaincode "isv-chaincode" "CheckDeviceAvailability" "[\"device1\"]" "3"

# 4. Validate Service Ticket
echo "Validating service ticket..."
execute_with_endorsement "isv-chaincode" "ValidateServiceTicket" "[\"$ENCRYPTED_SERVICE_TICKET\"]"

# 5. Process Service Request
echo "Processing service request..."
SERVICE_REQUEST="{\"encryptedServiceTicket\": \"$ENCRYPTED_SERVICE_TICKET\", \"clientID\": \"client1\", \"deviceID\": \"device1\", \"requestType\": \"read\", \"encryptedData\": \"simulated_encrypted_data_base64_string\"}"
SERVICE_RESPONSE=$(execute_with_endorsement "isv-chaincode" "ProcessServiceRequest" "[\"$SERVICE_REQUEST\"]")
echo "Service response: $SERVICE_RESPONSE"

# Extract session ID (For real implementation, parse the JSON response)
SESSION_ID="simulated_session_id_for_testing"

# 6. Handle Device Response
echo "Handling device response..."
execute_with_endorsement "isv-chaincode" "HandleDeviceResponse" "[\"$SESSION_ID\", \"Temperature is 25.5 C\"]"

# 7. Close Session
echo "Closing session..."
execute_with_endorsement "isv-chaincode" "CloseSession" "[\"$SESSION_ID\"]"

# 8. Get All IoT Devices
echo "Getting all IoT devices..."
query_chaincode "isv-chaincode" "GetAllIoTDevices" "[]" "3"

# 9. Get Active Sessions by Client
echo "Getting active sessions by client..."
query_chaincode "isv-chaincode" "GetActiveSessionsByClient" "[\"client1\"]" "3"

echo "===== ALL TESTS COMPLETED ====="

# Return to Org1 at the end
switch_to_org1

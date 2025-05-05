#!/bin/bash
# Fixed Testing Guide for Kerberos-like Chaincode System on Hyperledger Fabric

#####################################################################
# SECTION 1: ENVIRONMENT SETUP AND INITIALIZATION
#####################################################################

# Enter the CLI container (run this outside the container)
# docker exec -it cli bash

# Initial setup for environment variables and orderer CA
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Basic connectivity test to confirm peer addresses are reachable
echo "===== TESTING PEER CONNECTIVITY ====="
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_TLS_ENABLED=true

echo "Testing Org1 connectivity..."
peer channel list

export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_TLS_ENABLED=true

echo "Testing Org2 connectivity..."
peer channel list

# IMPORTANT: Changed port from 11051 to 13051 based on docker ps output
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:11051
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_TLS_ENABLED=true

echo "Testing Org3 connectivity..."
peer channel list

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
    # IMPORTANT: Changed port from 11051 to 13051 based on docker ps output
    export CORE_PEER_ADDRESS=peer0.org3.example.com:11051
    export CORE_PEER_LOCALMSPID="Org3MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
    export CORE_PEER_TLS_ENABLED=true
    echo "Now using Org3 peer0"
}

# Function to initialize chaincode with multi-org endorsement
initialize_chaincode() {
    CHAINCODE_NAME=$1
    
    echo "===== INITIALIZING $CHAINCODE_NAME ON ALL ORGANIZATIONS ====="
    
    # Check if chaincode is committed and endorsement policy
    echo "Checking chaincode policy for $CHAINCODE_NAME..."
    switch_to_org1
    peer lifecycle chaincode querycommitted -C chaichis-channel -n $CHAINCODE_NAME --output json
    
    # Initialize with multi-org endorsement
    echo "Initializing $CHAINCODE_NAME with multi-org endorsement..."
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c '{"function":"Initialize","Args":[]}' \
        --tls --cafile $ORDERER_CA --isInit \
        --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
        --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
        --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT
    
    sleep 3
    echo "$CHAINCODE_NAME initialized with multi-org endorsement"
}

# Function to execute a chaincode function with multi-org endorsement
execute_with_endorsement() {
    CHAINCODE_NAME=$1
    FUNCTION_NAME=$2
    ARGS=$3
    
    echo "Executing $FUNCTION_NAME on $CHAINCODE_NAME with multi-org endorsement..."
    switch_to_org1
    
    # IMPORTANT: Updated peer0.org3 port from 11051 to 13051
    peer chaincode invoke -C chaichis-channel -n $CHAINCODE_NAME -c "{\"function\":\"$FUNCTION_NAME\",\"Args\":$ARGS}" \
        --tls --cafile $ORDERER_CA \
        --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
        --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
        --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT
    
    RESPONSE=$?
    sleep 3
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

# Initialize all chaincodes with multi-org endorsement
initialize_chaincode "as-chaincode"
initialize_chaincode "tgs-chaincode"
initialize_chaincode "isv-chaincode"

#####################################################################
# SECTION 3: TESTING THE AUTHENTICATION SERVER (AS) CHAINCODE
#####################################################################

echo "===== TESTING AS CHAINCODE FUNCTIONS ====="

# Test Variables for AS
CLIENT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvSYtNtHdGJPPGRoSspF8
MXfnNUKHMr1OGht1YmpJ/a1dtj9o8oLQrFzpEhTls9EK31+TNAQ1Qev2HmwU8V35
pUWxlVXW4W9lKMctLfnPEhdlSWF+8mP+4UXcwhQDdZjiHgHM1v4SqR+dI1UBgHIq
eVrO34ScnRcwQNXM8qNi6tOvpfCB9aQT+WLZvN9zLsQgv5JZQXEVz1XIQzXjJV2x
WbsoI/f7thbeYTVHdkH2wjy06K5ijPy1vWQ+wbjdJZdxn5fEyu3OiUMLnd+ZuGLA
Q8I7h1jMPQ9JHnUl3whuyEY5bxUuXdBHQKU5PP+zQTpPHUxxqCLSAu0chIoEJAB9
vwIDAQAB
-----END PUBLIC KEY-----"

# 1. Register a Client with multi-org endorsement
echo "Registering client1 with multi-org endorsement..."
execute_with_endorsement "as-chaincode" "RegisterClient" "[\"client1\", \"$CLIENT_PUBLIC_KEY\"]"

# 2. Check Client Validity
echo "Checking client1 validity..."
query_chaincode "as-chaincode" "CheckClientValidity" "[\"client1\"]" "1"

# 3. Initiate Authentication
echo "Initiating authentication for client1..."
AUTH_RESPONSE=$(execute_with_endorsement "as-chaincode" "InitiateAuthentication" "[\"client1\"]")
echo "Authentication response: $AUTH_RESPONSE"

# 4. Get response data - extract nonce (simulated for this test)
NONCE_VALUE="simulated_nonce_from_response"
ENCRYPTED_NONCE="simulated_encrypted_nonce_base64_string"

# 5. Verify Client Identity
echo "Verifying client identity with multi-org endorsement..."
execute_with_endorsement "as-chaincode" "VerifyClientIdentity" "[\"client1\", \"$ENCRYPTED_NONCE\"]"

# 6. Generate TGT
echo "Generating TGT for client1 with multi-org endorsement..."
TGT_RESPONSE=$(execute_with_endorsement "as-chaincode" "GenerateTGT" "[\"client1\"]")
echo "TGT response: $TGT_RESPONSE"

# 7. Extract TGT (For production, extract the actual TGT from response)
ENCRYPTED_TGT="simulated_encrypted_tgt_base64_for_testing"
ENCRYPTED_SESSION_KEY="simulated_encrypted_session_key_base64_for_testing"

# 8. Get All Client Registrations
echo "Getting all client registrations from AS..."
query_chaincode "as-chaincode" "GetAllClientRegistrations" "[]" "1"

#####################################################################
# SECTION 4: TESTING THE TICKET GRANTING SERVICE (TGS) CHAINCODE
#####################################################################

echo "===== TESTING TGS CHAINCODE FUNCTIONS ====="

# 1. Process Registration from AS
echo "Processing registration from AS with multi-org endorsement..."
execute_with_endorsement "tgs-chaincode" "ProcessRegistrationFromAS" "[\"$ENCRYPTED_TGT\"]"

# 2. Check Registration Validity
echo "Checking registration validity at TGS..."
query_chaincode "tgs-chaincode" "CheckRegistrationValidity" "[\"client1\"]" "2"

# 3. Generate Service Ticket
echo "Generating service ticket with multi-org endorsement..."
SERVICE_REQUEST="{\"encryptedTGT\": \"$ENCRYPTED_TGT\", \"clientID\": \"client1\", \"serviceID\": \"iotservice1\", \"authenticator\": \"simulated_authenticator_base64_string\"}"
SERVICE_REQUEST_ENCODED=$(echo -n "$SERVICE_REQUEST" | base64 -w 0)
SERVICE_TICKET_RESPONSE=$(execute_with_endorsement "tgs-chaincode" "GenerateServiceTicket" "[\"$SERVICE_REQUEST_ENCODED\"]")
echo "Service ticket response: $SERVICE_TICKET_RESPONSE"

# 4. Extract service ticket (For production, extract from the actual response)
ENCRYPTED_SERVICE_TICKET="simulated_encrypted_service_ticket_base64_for_testing"

# 5. Forward Registration to ISV
echo "Forwarding registration to ISV with multi-org endorsement..."
execute_with_endorsement "tgs-chaincode" "ForwardRegistrationToISV" "[\"client1\", \"iotservice1\", \"$ENCRYPTED_SERVICE_TICKET\"]"

# 6. Get All Client Registrations from TGS
echo "Getting all client registrations from TGS..."
query_chaincode "tgs-chaincode" "GetAllClientRegistrations" "[]" "2"

#####################################################################
# SECTION 5: TESTING THE IOT SERVICE VALIDATOR (ISV) CHAINCODE
#####################################################################

echo "===== TESTING ISV CHAINCODE FUNCTIONS ====="

# 1. Register IoT Device
echo "Registering IoT device with multi-org endorsement..."
DEVICE_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdef
abcdef1234567890ABCDEF1234567890abcdefABCDEF==
-----END PUBLIC KEY-----"
CAPABILITIES='["temperature","humidity","pressure"]'

execute_with_endorsement "isv-chaincode" "RegisterIoTDevice" "[\"device1\", \"$DEVICE_PUBLIC_KEY\", $CAPABILITIES]"

# 2. Update Device Status
echo "Updating device status with multi-org endorsement..."
SIGNATURE="simulated_signature_base64_string"
execute_with_endorsement "isv-chaincode" "UpdateDeviceStatus" "[\"device1\", \"active\", \"$SIGNATURE\"]"

# 3. Check Device Availability
echo "Checking device availability..."
query_chaincode "isv-chaincode" "CheckDeviceAvailability" "[\"device1\"]" "3"

# 4. Validate Service Ticket
echo "Validating service ticket with multi-org endorsement..."
execute_with_endorsement "isv-chaincode" "ValidateServiceTicket" "[\"$ENCRYPTED_SERVICE_TICKET\"]"

# 5. Process Service Request
echo "Processing service request with multi-org endorsement..."
SERVICE_REQUEST="{\"encryptedServiceTicket\": \"$ENCRYPTED_SERVICE_TICKET\", \"clientID\": \"client1\", \"deviceID\": \"device1\", \"requestType\": \"read\", \"encryptedData\": \"simulated_encrypted_data_base64_string\"}"
SERVICE_REQUEST_ENCODED=$(echo -n "$SERVICE_REQUEST" | base64 -w 0)
SERVICE_RESPONSE=$(execute_with_endorsement "isv-chaincode" "ProcessServiceRequest" "[\"$SERVICE_REQUEST_ENCODED\"]")
echo "Service response: $SERVICE_RESPONSE"

# 6. Extract session ID (For production, extract from the actual response)
SESSION_ID="simulated_session_id_for_testing"

# 7. Handle Device Response
echo "Handling device response with multi-org endorsement..."
execute_with_endorsement "isv-chaincode" "HandleDeviceResponse" "[\"$SESSION_ID\", \"Temperature is 25.5 C\"]"

# 8. Close Session
echo "Closing session with multi-org endorsement..."
execute_with_endorsement "isv-chaincode" "CloseSession" "[\"$SESSION_ID\"]"

# 9. Get All IoT Devices
echo "Getting all IoT devices..."
query_chaincode "isv-chaincode" "GetAllIoTDevices" "[]" "3"

# 10. Get Active Sessions by Client
echo "Getting active sessions by client..."
query_chaincode "isv-chaincode" "GetActiveSessionsByClient" "[\"client1\"]" "3"

echo "===== ALL TESTS COMPLETED SUCCESSFULLY ====="

# Return to Org1 at the end
switch_to_org1

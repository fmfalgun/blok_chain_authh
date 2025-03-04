#!/bin/bash

# Exit on first error
set -e

# Set environment variables for peer commands
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export PEER0_ORG2_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export PEER0_ORG3_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

# Define variables
CHANNEL_NAME="chaichis-channel"
CC_NAME="iot-auth"

# Function to set environment variables for each organization
setOrg1Env() {
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
}

setOrg2Env() {
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
}

setOrg3Env() {
  export CORE_PEER_LOCALMSPID="Org3MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG3_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org3.example.com:13051
}

echo "===== Testing IoT Authentication Chaincode Functions ====="

# Generate unique IDs for testing
CLIENT_ID="client_$(date +%s)"
DEVICE_ID="device_$(date +%s)"

echo "Using client ID: $CLIENT_ID"
echo "Using device ID: $DEVICE_ID"

# 1. Test AS (Org1) Functions
echo -e "\n\n===== 1. Testing Authentication Server (Org1) Functions ====="
setOrg1Env

echo "1.1 Registering new client..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c "{\"function\":\"RegisterClient\",\"Args\":[\"$CLIENT_ID\", \"client-public-key-123\", \"org1\"]}"

sleep 3

echo "1.2 Validating client..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c "{\"function\":\"ValidateClient\",\"Args\":[\"$CLIENT_ID\"]}"

sleep 3

echo "1.3 Issuing Ticket Granting Ticket..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c "{\"function\":\"IssueTicketGrantingTicket\",\"Args\":[\"$CLIENT_ID\"]}"

sleep 3

echo "1.4 Getting client info..."
CLIENT_INFO=$(peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c "{\"function\":\"GetAllClients\",\"Args\":[]}")
echo "Client info: $CLIENT_INFO"

# Extract the TGT for later use
TGT=$(echo $CLIENT_INFO | grep -o "\"tgt\":\"[^\"]*\"" | cut -d':' -f2 | tr -d '",')
echo "Retrieved TGT: $TGT"

# 2. Test TGS (Org2) Functions
echo -e "\n\n===== 2. Testing Ticket Granting Server (Org2) Functions ====="
setOrg2Env

echo "2.1 Verifying TGT..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA -c "{\"function\":\"VerifyTicketGrantingTicket\",\"Args\":[\"$CLIENT_ID\", \"$TGT\"]}"

sleep 3

echo "2.2 Issuing Service Ticket..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA -c "{\"function\":\"IssueServiceTicket\",\"Args\":[\"$CLIENT_ID\", \"$DEVICE_ID\"]}"

sleep 3

echo "2.3 Forwarding to ISV..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA -c "{\"function\":\"ForwardToISV\",\"Args\":[\"$CLIENT_ID\"]}"

sleep 3

# Check client info after TGS operations
setOrg1Env
echo "Client info after TGS operations:"
CLIENT_INFO=$(peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c "{\"function\":\"GetAllClients\",\"Args\":[]}")
echo $CLIENT_INFO

# Extract the Service Ticket for later use
SERVICE_TICKET=$(echo $CLIENT_INFO | grep -o "\"serviceTicket\":\"[^\"]*\"" | cut -d':' -f2 | tr -d '",')
echo "Retrieved Service Ticket: $SERVICE_TICKET"

# 3. Test ISV (Org3) Functions
echo -e "\n\n===== 3. Testing IoT Service Validator (Org3) Functions ====="
setOrg3Env

echo "3.1 Registering IoT Device..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c "{\"function\":\"RegisterIoTDevice\",\"Args\":[\"$DEVICE_ID\", \"device-public-key-456\", \"org3\", \"temperature-sensor\"]}"

sleep 3

echo "3.2 Checking Device Availability..."
peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c "{\"function\":\"CheckDeviceAvailability\",\"Args\":[\"$DEVICE_ID\"]}"

sleep 3

echo "3.3 Verifying Service Ticket..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c "{\"function\":\"VerifyServiceTicket\",\"Args\":[\"$CLIENT_ID\", \"$SERVICE_TICKET\", \"$DEVICE_ID\"]}"

sleep 3

echo "3.4 Granting Device Access..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c "{\"function\":\"GrantDeviceAccess\",\"Args\":[\"$CLIENT_ID\", \"$DEVICE_ID\"]}"

# Final check of client status
setOrg1Env
echo -e "\n\n===== Final Client Status ====="
CLIENT_INFO=$(peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c "{\"function\":\"GetAllClients\",\"Args\":[]}")
echo $CLIENT_INFO

echo -e "\n===== All Tests Completed Successfully! ====="

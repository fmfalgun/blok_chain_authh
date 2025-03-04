#!/bin/bash

# Exit on first error
set -e

# Set environment variables for peer commands
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_ORG3_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG3_CA
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

# Variables
CHANNEL_NAME="chaichis-channel"
CC_NAME="iot-auth"

echo "===== Testing Org3 (IoT Service Validator) Functions ====="

# Test RegisterIoTDevice function
echo "1. Registering IoT device..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c '{"function":"RegisterIoTDevice","Args":["iot-device1", "device1-public-key", "org3", "temperature-sensor"]}'
sleep 3

# Test CheckDeviceAvailability function
echo "2. Checking device availability..."
peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"CheckDeviceAvailability","Args":["iot-device1"]}'

# Get client service ticket (for reference - in a real implementation, the client would send this)
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

CLIENT_REG=$(peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"GetAllClients","Args":[]}' | jq -r '.[0]')
SERVICE_TICKET=$(echo $CLIENT_REG | jq -r '.serviceTicket')
echo "Retrieved service ticket for client1: $SERVICE_TICKET"

# Reset to Org3 peer
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG3_CA
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

# Test VerifyServiceTicket function
echo "3. Verifying service ticket..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c "{\"function\":\"VerifyServiceTicket\",\"Args\":[\"client1\", \"$SERVICE_TICKET\", \"iot-device1\"]}"
sleep 3

# Test GrantDeviceAccess function
echo "4. Granting device access..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c '{"function":"GrantDeviceAccess","Args":["client1", "iot-device1"]}'
sleep 3

echo "===== Org3 (ISV) Functions Testing Completed ====="

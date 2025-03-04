#!/bin/bash

# Exit on first error
set -e

# Set environment variables for peer commands
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

# Variables
CHANNEL_NAME="chaichis-channel"
CC_NAME="iot-auth"

echo "===== Testing Org1 (Authentication Server) Functions ====="

# Test RegisterClient function
echo "1. Registering client1..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c '{"function":"RegisterClient","Args":["client1", "client1-public-key", "org1"]}'
sleep 3

# Test ValidateClient function
echo "2. Validating client1..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c '{"function":"ValidateClient","Args":["client1"]}'
sleep 3

# Test IssueTicketGrantingTicket function
echo "3. Issuing TGT to client1..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c '{"function":"IssueTicketGrantingTicket","Args":["client1"]}'
sleep 3

# Test GetAllClients function
echo "4. Getting all clients..."
peer chaincode query -C $CHANNEL_NAME -n $CC_NAME -c '{"function":"GetAllClients","Args":[]}'

# Test AllocatePeerTasks function
echo "5. Allocating peer tasks..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CC_NAME --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c '{"function":"AllocatePeerTasks","Args":["peer1.org1.example.com", "validation", "client1"]}'
sleep 3

echo "===== Org1 (AS) Functions Testing Completed ====="

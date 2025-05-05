#!/bin/bash

# Set environment variables for TLS
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Function to switch to Org1
switch_to_org1() {
    export CORE_PEER_LOCALMSPID="Org1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
}

# Function to switch to Org2
switch_to_org2() {
    export CORE_PEER_LOCALMSPID="Org2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
}

# Function to switch to Org3
switch_to_org3() {
    export CORE_PEER_LOCALMSPID="Org3MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org3.example.com:11051
}

echo "Packaging chaincode..."
docker exec cli peer lifecycle chaincode package as-chaincode.tar.gz --path /opt/gopath/src/github.com/hyperledger/fabric/peer/chaincodes/as-chaincode-fixed-v4 --lang golang --label as-chaincode

# Install on Org1
echo "Installing chaincode on Org1..."
switch_to_org1
docker exec cli peer lifecycle chaincode install as-chaincode.tar.gz

# Install on Org2
echo "Installing chaincode on Org2..."
switch_to_org2
docker exec cli peer lifecycle chaincode install as-chaincode.tar.gz

# Install on Org3
echo "Installing chaincode on Org3..."
switch_to_org3
docker exec cli peer lifecycle chaincode install as-chaincode.tar.gz

# Query installed chaincode
echo "Querying installed chaincode..."
docker exec cli peer lifecycle chaincode queryinstalled > package_id.txt
cat package_id.txt 
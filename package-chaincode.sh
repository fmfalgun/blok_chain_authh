#!/bin/bash

# Set environment variables
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Function to switch to Org1 context
switch_to_org1() {
    export CORE_PEER_LOCALMSPID=Org1MSP
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
}

# Check channel configuration
echo "Checking channel configuration..."
switch_to_org1
docker exec cli peer channel fetch config config_block.pb -o orderer.example.com:7050 -c channel1 --tls --cafile $ORDERER_CA

# Check if chaincode is already installed
echo "Checking installed chaincodes..."
docker exec cli peer lifecycle chaincode queryinstalled

# Get the package ID
PACKAGE_ID=$(docker exec cli peer lifecycle chaincode queryinstalled | grep "as-chaincode" | awk '{print $3}' | tr -d ',')
echo "Package ID: $PACKAGE_ID"

# Approve chaincode for Org1
echo "Approving chaincode for Org1..."
docker exec cli peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile $ORDERER_CA --channelID channel1 --name as-chaincode --version 1.0 --package-id $PACKAGE_ID --sequence 1 --init-required --waitForEvent --waitForEventTimeout 60s

# Function to switch to Org2 context
switch_to_org2() {
    export CORE_PEER_LOCALMSPID=Org2MSP
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
}

# Function to switch to Org3 context
switch_to_org3() {
    export CORE_PEER_LOCALMSPID=Org3MSP
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org3.example.com:13051
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
}

# Approve chaincode for Org2
echo "Approving chaincode for Org2..."
switch_to_org2
docker exec cli peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile $ORDERER_CA --channelID channel1 --name as-chaincode --version 1.0 --package-id $PACKAGE_ID --sequence 1 --init-required --waitForEvent --waitForEventTimeout 60s

# Approve chaincode for Org3
echo "Approving chaincode for Org3..."
switch_to_org3
docker exec cli peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile $ORDERER_CA --channelID channel1 --name as-chaincode --version 1.0 --package-id $PACKAGE_ID --sequence 1 --init-required --waitForEvent --waitForEventTimeout 60s

# Check approval status
echo "Checking approval status..."
switch_to_org1
docker exec cli peer lifecycle chaincode checkcommitreadiness --channelID channel1 --name as-chaincode --version 1.0 --sequence 1 --init-required --output json

# Commit chaincode definition
echo "Committing chaincode definition..."
docker exec cli peer lifecycle chaincode commit -o orderer.example.com:7050 --tls --cafile $ORDERER_CA --channelID channel1 --name as-chaincode --version 1.0 --sequence 1 --init-required --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt --waitForEvent --waitForEventTimeout 60s

# Initialize the chaincode
echo "Initializing chaincode..."
docker exec cli peer chaincode invoke -o orderer.example.com:7050 --tls --cafile $ORDERER_CA -C channel1 -n as-chaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt --isInit -c '{"Args":["Init"]}' 
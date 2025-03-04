#!/bin/bash
set -e

CHANNEL_NAME="chaichis-channel"
CC_NAME="iot-auth"
CC_VERSION="1.0"
CC_SEQUENCE="1"

# Set environment variables for peer commands
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export PEER0_ORG2_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export PEER0_ORG3_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

# This function sets the environment variables for the peer org1
setOrg1() {
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
}

# This function sets the environment variables for the peer org2
setOrg2() {
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
}

# This function sets the environment variables for the peer org3
setOrg3() {
  export CORE_PEER_LOCALMSPID="Org3MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG3_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org3.example.com:13051
}

# Try to install abstore chaincode directly (old method)
setOrg1
echo "==================== Installing chaincode on peer0.org1 ===================="
peer chaincode install -n ${CC_NAME} -v ${CC_VERSION} -p github.com/hyperledger/fabric/examples/chaincode/go/abstore/

setOrg2
echo "==================== Installing chaincode on peer0.org2 ===================="
peer chaincode install -n ${CC_NAME} -v ${CC_VERSION} -p github.com/hyperledger/fabric/examples/chaincode/go/abstore/

setOrg3
echo "==================== Installing chaincode on peer0.org3 ===================="
peer chaincode install -n ${CC_NAME} -v ${CC_VERSION} -p github.com/hyperledger/fabric/examples/chaincode/go/abstore/

# Initialize the chaincode
setOrg1
echo "==================== Instantiating chaincode ===================="
peer chaincode instantiate -o orderer.example.com:7050 --tls --cafile $ORDERER_CA -C ${CHANNEL_NAME} -n ${CC_NAME} -v ${CC_VERSION} -c '{"Args":["init","a","100","b","200"]}' -P "OR ('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')"

echo "Waiting for chaincode instantiation to complete..."
sleep 10

echo "==================== Testing chaincode ===================="
# Query the chaincode
peer chaincode query -C ${CHANNEL_NAME} -n ${CC_NAME} -c '{"Args":["query","a"]}'

# Invoke the chaincode
peer chaincode invoke -o orderer.example.com:7050 --tls --cafile $ORDERER_CA -C ${CHANNEL_NAME} -n ${CC_NAME} --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA -c '{"Args":["invoke","a","b","10"]}'

echo "==================== Chaincode deployment completed ===================="

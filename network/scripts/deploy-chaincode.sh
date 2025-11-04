#!/bin/bash

CHANNEL_NAME="authchannel"
CC_NAME=$1
CC_VERSION="1.0"
CC_SEQUENCE="1"
CC_SRC_PATH="../../chaincodes/${CC_NAME}-chaincode"

if [ -z "$CC_NAME" ]; then
  echo "Usage: ./deploy-chaincode.sh <chaincode-name>"
  echo "Example: ./deploy-chaincode.sh as"
  exit 1
fi

echo "=========================================="
echo "Deploying chaincode: $CC_NAME"
echo "=========================================="

# Package chaincode
echo "1. Packaging chaincode..."
docker exec cli peer lifecycle chaincode package ${CC_NAME}.tar.gz \
  --path /opt/gopath/src/github.com/hyperledger/fabric/chaincodes/${CC_NAME}-chaincode \
  --lang golang \
  --label ${CC_NAME}_${CC_VERSION}

# Install on all peers
echo "2. Installing chaincode on all peers..."
for org in 1 2 3; do
  for peer in 0 1; do
    echo "Installing on peer${peer}.org${org}..."
    docker exec cli bash -c "
      export CORE_PEER_LOCALMSPID=Org${org}MSP
      export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/peers/peer${peer}.org${org}.example.com/tls/ca.crt
      export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/users/Admin@org${org}.example.com/msp
      export CORE_PEER_ADDRESS=peer${peer}.org${org}.example.com:$((7051 + (org-1)*2000 + peer*1000))
      peer lifecycle chaincode install ${CC_NAME}.tar.gz
    "
  done
done

# Query installed chaincode to get package ID
echo "3. Querying installed chaincode..."
PACKAGE_ID=$(docker exec cli peer lifecycle chaincode queryinstalled | grep ${CC_NAME}_${CC_VERSION} | awk '{print $3}' | sed 's/,$//')
echo "Package ID: $PACKAGE_ID"

# Approve chaincode for each org
echo "4. Approving chaincode for each organization..."
for org in 1 2 3; do
  echo "Approving for Org${org}..."
  docker exec cli bash -c "
    export CORE_PEER_LOCALMSPID=Org${org}MSP
    export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/peers/peer0.org${org}.example.com/tls/ca.crt
    export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/users/Admin@org${org}.example.com/msp
    export CORE_PEER_ADDRESS=peer0.org${org}.example.com:$((7051 + (org-1)*2000))
    peer lifecycle chaincode approveformyorg \
      -o orderer.example.com:7050 \
      --channelID ${CHANNEL_NAME} \
      --name ${CC_NAME} \
      --version ${CC_VERSION} \
      --package-id ${PACKAGE_ID} \
      --sequence ${CC_SEQUENCE} \
      --tls \
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  "
done

# Check commit readiness
echo "5. Checking commit readiness..."
docker exec cli peer lifecycle chaincode checkcommitreadiness \
  --channelID ${CHANNEL_NAME} \
  --name ${CC_NAME} \
  --version ${CC_VERSION} \
  --sequence ${CC_SEQUENCE} \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Commit chaincode
echo "6. Committing chaincode..."
docker exec cli peer lifecycle chaincode commit \
  -o orderer.example.com:7050 \
  --channelID ${CHANNEL_NAME} \
  --name ${CC_NAME} \
  --version ${CC_VERSION} \
  --sequence ${CC_SEQUENCE} \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
  --peerAddresses peer0.org3.example.com:11051 \
  --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

# Query committed chaincode
echo "7. Verifying chaincode deployment..."
docker exec cli peer lifecycle chaincode querycommitted \
  --channelID ${CHANNEL_NAME} \
  --name ${CC_NAME} \
  --tls \
  --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

echo "=========================================="
echo "Chaincode $CC_NAME deployed successfully!"
echo "=========================================="

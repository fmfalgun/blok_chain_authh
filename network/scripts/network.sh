#!/bin/bash

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config

CHANNEL_NAME="authchannel"
DELAY="3"
MAX_RETRY="5"

function printHelp() {
  echo "Usage: "
  echo "  network.sh <Mode> [Flags]"
  echo "    <Mode>"
  echo "      - 'up' - Bring up Fabric network with docker-compose"
  echo "      - 'down' - Clear the network with docker-compose down"
  echo "      - 'restart' - Restart the network"
  echo "      - 'createChannel' - Create and join channel"
  echo ""
  echo "    Flags:"
  echo "    -ca <use CAs> - Create Org certificates with CAs"
  echo "    -c <channel name> - Channel name to use (defaults to \"authchannel\")"
  echo "    -s <dbtype> - State database to use (goleveldb|couchdb, defaults to goleveldb)"
  echo "    -r <max retry> - CLI times out after this number of attempts (defaults to 5)"
  echo "    -d <delay> - Delay between retries (defaults to 3s)"
  echo "    -verbose - Verbose output"
  echo ""
  echo "  network.sh -h (print this message)"
  echo ""
  echo " Examples:"
  echo "  network.sh up createChannel -ca -c authchannel -s couchdb"
}

function clearContainers() {
  CONTAINER_IDS=$(docker ps -a | awk '($2 ~ /hyperledger/) {print $1}')
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == " " ]; then
    echo "---- No containers available for deletion ----"
  else
    docker rm -f $CONTAINER_IDS
  fi
}

function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*/) {print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    echo "---- No images available for deletion ----"
  else
    docker rmi -f $DOCKER_IMAGE_IDS
  fi
}

function networkDown() {
  docker-compose -f $COMPOSE_FILE down --volumes --remove-orphans
  clearContainers
  removeUnwantedImages

  # Remove channel artifacts and crypto material
  rm -rf ../channel-artifacts/*.block ../channel-artifacts/*.tx
  rm -rf ../crypto-config
  rm -rf ../ledgers

  echo "---- Network shut down ----"
}

function generateCerts() {
  which cryptogen
  if [ "$?" -ne 0 ]; then
    echo "cryptogen tool not found. exiting"
    exit 1
  fi

  echo "##########################################################"
  echo "##### Generate certificates using cryptogen tool #########"
  echo "##########################################################"

  if [ -d "../crypto-config" ]; then
    rm -Rf ../crypto-config
  fi

  set -x
  cryptogen generate --config=../config/crypto-config.yaml --output="../crypto-config"
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "Failed to generate certificates..."
    exit 1
  fi

  echo "---- Generated certificates ----"
}

function generateChannelArtifacts() {
  which configtxgen
  if [ "$?" -ne 0 ]; then
    echo "configtxgen tool not found. exiting"
    exit 1
  fi

  if [ ! -d "../channel-artifacts" ]; then
    mkdir ../channel-artifacts
  fi

  echo "#################################################################"
  echo "### Generating channel configuration transaction 'channel.tx' ###"
  echo "#################################################################"

  set -x
  configtxgen -profile ThreeOrgsOrdererGenesis -channelID system-channel -outputBlock ../channel-artifacts/genesis.block -configPath ../config
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "Failed to generate orderer genesis block..."
    exit 1
  fi

  echo "###################################################################"
  echo "#######    Generating channel transaction file 'channel.tx'   ####"
  echo "###################################################################"

  set -x
  configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ../channel-artifacts/channel.tx -channelID $CHANNEL_NAME -configPath ../config
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "Failed to generate channel configuration transaction..."
    exit 1
  fi

  echo "#################################################################"
  echo "#######    Generating anchor peer update for Org1MSP   ##########"
  echo "#################################################################"

  set -x
  configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ../channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP -configPath ../config
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org1MSP..."
    exit 1
  fi

  echo "#################################################################"
  echo "#######    Generating anchor peer update for Org2MSP   ##########"
  echo "#################################################################"

  set -x
  configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ../channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP -configPath ../config
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org2MSP..."
    exit 1
  fi

  echo "#################################################################"
  echo "#######    Generating anchor peer update for Org3MSP   ##########"
  echo "#################################################################"

  set -x
  configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ../channel-artifacts/Org3MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org3MSP -configPath ../config
  res=$?
  set +x

  if [ $res -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org3MSP..."
    exit 1
  fi

  echo "---- Generated channel artifacts ----"
}

function networkUp() {
  if [ ! -d "../crypto-config" ]; then
    generateCerts
    generateChannelArtifacts
  fi

  COMPOSE_FILE=../config/docker-compose-network.yaml

  docker-compose -f $COMPOSE_FILE up -d 2>&1

  docker ps -a
  if [ $? -ne 0 ]; then
    echo "ERROR !!!! Unable to start network"
    exit 1
  fi

  echo "---- Network started ----"
}

function createChannel() {
  docker exec cli peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --outputBlock ./channel-artifacts/${CHANNEL_NAME}.block --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

  echo "---- Channel '$CHANNEL_NAME' created ----"
}

function joinChannel() {
  for org in 1 2 3; do
    for peer in 0 1; do
      echo "Joining peer${peer}.org${org} to channel..."
      docker exec cli bash -c "
        CORE_PEER_LOCALMSPID=Org${org}MSP
        CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/peers/peer${peer}.org${org}.example.com/tls/ca.crt
        CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/users/Admin@org${org}.example.com/msp
        CORE_PEER_ADDRESS=peer${peer}.org${org}.example.com:$((7051 + (org-1)*2000 + peer*1000))
        peer channel join -b ./channel-artifacts/${CHANNEL_NAME}.block
      "
    done
  done

  echo "---- All peers joined channel ----"
}

function updateAnchorPeers() {
  for org in 1 2 3; do
    echo "Updating anchor peers for Org${org}..."
    docker exec cli bash -c "
      CORE_PEER_LOCALMSPID=Org${org}MSP
      CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/peers/peer0.org${org}.example.com/tls/ca.crt
      CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org${org}.example.com/users/Admin@org${org}.example.com/msp
      CORE_PEER_ADDRESS=peer0.org${org}.example.com:$((7051 + (org-1)*2000))
      peer channel update -o orderer.example.com:7050 -c ${CHANNEL_NAME} -f ./channel-artifacts/Org${org}MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    "
  done

  echo "---- Anchor peers updated ----"
}

# Parse commandline args
if [[ $# -lt 1 ]] ; then
  printHelp
  exit 0
else
  MODE=$1
  shift
fi

# Parse flags
while [[ $# -ge 1 ]] ; do
  key="$1"
  case $key in
  -h )
    printHelp
    exit 0
    ;;
  -c )
    CHANNEL_NAME="$2"
    shift
    ;;
  -r )
    MAX_RETRY="$2"
    shift
    ;;
  -d )
    DELAY="$2"
    shift
    ;;
  * )
    echo
    echo "Unknown flag: $key"
    echo
    printHelp
    exit 1
    ;;
  esac
  shift
done

# Determine mode of operation and run
if [ "${MODE}" == "up" ]; then
  networkUp
elif [ "${MODE}" == "createChannel" ]; then
  createChannel
  joinChannel
  updateAnchorPeers
elif [ "${MODE}" == "down" ]; then
  networkDown
elif [ "${MODE}" == "restart" ]; then
  networkDown
  networkUp
else
  printHelp
  exit 1
fi

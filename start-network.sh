#!/bin/bash

# Exit on any error
set -e

# Define color codes for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function for printing colored output
print_green() {
    echo -e "${GREEN}$1${NC}"
}

print_yellow() {
    echo -e "${YELLOW}$1${NC}"
}

print_red() {
    echo -e "${RED}$1${NC}"
}

# Function to check if Docker is installed and running
check_prerequisites() {
    print_yellow "Checking prerequisites..."
    
    # Check Docker
    if ! [ -x "$(command -v docker)" ]; then
        print_red "Error: docker is not installed."
        exit 1
    fi

    # Check Docker Compose
    if ! [ -x "$(command -v docker-compose)" ]; then
        print_red "Error: docker-compose is not installed."
        exit 1
    fi
    
    # Check if Docker daemon is running
    docker info > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        print_red "Error: Docker daemon is not running."
        exit 1
    fi
    
    print_green "All prerequisites satisfied."
}

# Function to generate crypto material if not already present
generate_crypto() {
    print_yellow "Checking for crypto material..."
    
    if [ ! -d "organizations/peerOrganizations" ]; then
        print_yellow "Generating crypto material with cryptogen tool..."
        
        # Make sure the crypto-config file exists
        if [ ! -f "crypto-config.yaml" ]; then
            print_red "Error: crypto-config.yaml not found."
            exit 1
        fi
        
        # Run the cryptogen tool
        cryptogen generate --config=./crypto-config.yaml --output="organizations"
        
        if [ $? -ne 0 ]; then
            print_red "Failed to generate crypto material..."
            exit 1
        fi
    else
        print_green "Crypto material already exists."
    fi
}

# Function to generate channel artifacts
generate_channel_artifacts() {
    print_yellow "Generating channel artifacts..."
    
    mkdir -p channel-artifacts
    
    # Generate genesis block for orderer
    configtxgen -profile ThreeOrgsOrdererGenesis -channelID system-channel -outputBlock ./channel-artifacts/genesis.block
    
    if [ $? -ne 0 ]; then
        print_red "Failed to generate orderer genesis block..."
        exit 1
    fi
    
    # Generate channel creation transaction
    configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID chaichis-channel
    
    if [ $? -ne 0 ]; then
        print_red "Failed to generate channel configuration transaction..."
        exit 1
    fi
    
    # Generate anchor peer transactions for each org
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID chaichis-channel -asOrg Org1MSP
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID chaichis-channel -asOrg Org2MSP
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors.tx -channelID chaichis-channel -asOrg Org3MSP
    
    print_green "Channel artifacts generated."
}

# Function to start the network
start_network() {
    print_yellow "Starting the network..."
    
    # Start the containers
    docker-compose -f docker-compose.yaml up -d
    
    # Wait for the containers to start
    sleep 10
    
    print_green "Network started successfully."
}

# Function to create and join channel
create_and_join_channel() {
    print_yellow "Creating channel and joining peers..."
    
    # Create the channel
    docker exec cli peer channel create -o orderer.example.com:7050 -c chaichis-channel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/channel.tx \
        --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    
    # Join peer0.org1 to the channel
    docker exec cli peer channel join -b chaichis-channel.block
    
    # Join peer1.org1 to the channel
    docker exec -e CORE_PEER_ADDRESS=peer1.org1.example.com:8051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer2.org1 to the channel
    docker exec -e CORE_PEER_ADDRESS=peer2.org1.example.com:11051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer2.org1.example.com/tls/ca.crt \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer0.org2 to the channel
    docker exec -e CORE_PEER_LOCALMSPID=Org2MSP \
        -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer1.org2 to the channel
    docker exec -e CORE_PEER_LOCALMSPID=Org2MSP \
        -e CORE_PEER_ADDRESS=peer1.org2.example.com:10051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer2.org2 to the channel
    docker exec -e CORE_PEER_LOCALMSPID=Org2MSP \
        -e CORE_PEER_ADDRESS=peer2.org2.example.com:12051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer2.org2.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer0.org3 to the channel
    docker exec -e CORE_PEER_LOCALMSPID=Org3MSP \
        -e CORE_PEER_ADDRESS=peer0.org3.example.com:13051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer1.org3 to the channel
    docker exec -e CORE_PEER_LOCALMSPID=Org3MSP \
        -e CORE_PEER_ADDRESS=peer1.org3.example.com:14051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer1.org3.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp \
        cli peer channel join -b chaichis-channel.block
    
    # Join peer2.org3 to the channel
    docker exec -e CORE_PEER_LOCALMSPID=Org3MSP \
        -e CORE_PEER_ADDRESS=peer2.org3.example.com:15051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer2.org3.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp \
        cli peer channel join -b chaichis-channel.block
    
    print_green "All peers joined the channel."
}

# Function to update anchor peers
update_anchor_peers() {
    print_yellow "Updating anchor peers..."
    
    # Update anchor peers for Org1
    docker exec cli peer channel update -o orderer.example.com:7050 -c chaichis-channel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/Org1MSPanchors.tx \
        --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    
    # Update anchor peers for Org2
    docker exec -e CORE_PEER_LOCALMSPID=Org2MSP \
        -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
        cli peer channel update -o orderer.example.com:7050 -c chaichis-channel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/Org2MSPanchors.tx \
        --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    
    # Update anchor peers for Org3
    docker exec -e CORE_PEER_LOCALMSPID=Org3MSP \
        -e CORE_PEER_ADDRESS=peer0.org3.example.com:13051 \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt \
        -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp \
        cli peer channel update -o orderer.example.com:7050 -c chaichis-channel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/Org3MSPanchors.tx \
        --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    
    print_green "Anchor peers updated."
}

# Function to check network health
check_network_health() {
    print_yellow "Checking network health..."
    
    # Check if all containers are running
    RUNNING_CONTAINERS=$(docker ps -f status=running | grep -c "hyperledger/fabric")
    EXPECTED_CONTAINERS=13  # 1 orderer + 9 peers + 1 cli + 4 CAs
    
    if [ "$RUNNING_CONTAINERS" -lt "$EXPECTED_CONTAINERS" ]; then
        print_red "Error: Not all containers are running. Expected $EXPECTED_CONTAINERS, found $RUNNING_CONTAINERS."
        docker ps -a
        exit 1
    fi
    
    print_green "Network is healthy with $RUNNING_CONTAINERS containers running."
}

# Main execution flow
main() {
    print_yellow "Starting Chaichis Network Deployment..."
    
    check_prerequisites
    generate_crypto
    generate_channel_artifacts
    start_network
    create_and_join_channel
    update_anchor_peers
    check_network_health
    
    print_green "==== Chaichis Network Successfully Deployed ===="
    print_green "The network has 3 organizations with 3 peers each, all connected to a single channel."
    print_green "Channel name: chaichis-channel"
    print_green "You can now deploy chaincode and develop applications for this network."
}

# Execute main function
main

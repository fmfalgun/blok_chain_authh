#!/bin/bash
# copy-certs.sh - Script to copy TLS certificates from Docker container

# Exit on error
set -e

YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Copying TLS certificates from Docker container...${NC}"

# Create directories for certificates
mkdir -p organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls
mkdir -p organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls
mkdir -p organizations/peerOrganizations/org1.example.com/ca
mkdir -p organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls
mkdir -p organizations/peerOrganizations/org2.example.com/ca
mkdir -p organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls
mkdir -p organizations/peerOrganizations/org3.example.com/ca

# Copy orderer TLS certificate
echo -e "${YELLOW}Copying orderer TLS certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/

# Copy org1 peer0 TLS certificate
echo -e "${YELLOW}Copying org1 peer0 TLS certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/

# Copy org1 CA certificate
echo -e "${YELLOW}Copying org1 CA certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem organizations/peerOrganizations/org1.example.com/ca/

# Copy org2 peer0 TLS certificate
echo -e "${YELLOW}Copying org2 peer0 TLS certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/

# Copy org2 CA certificate
echo -e "${YELLOW}Copying org2 CA certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem organizations/peerOrganizations/org2.example.com/ca/

# Copy org3 peer0 TLS certificate
echo -e "${YELLOW}Copying org3 peer0 TLS certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/

# Copy org3 CA certificate
echo -e "${YELLOW}Copying org3 CA certificate...${NC}"
docker cp cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/ca/ca.org3.example.com-cert.pem organizations/peerOrganizations/org3.example.com/ca/

echo -e "${GREEN}Successfully copied TLS certificates!${NC}"

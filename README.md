# Blockchain-based Authentication System

## Overview
This project implements a secure authentication system leveraging blockchain technology. Built on Hyperledger Fabric, it provides a distributed, immutable, and transparent authentication mechanism that eliminates single points of failure while maintaining high security standards.

## Prerequisites
- Docker and Docker Compose
- Hyperledger Fabric binaries
- Node.js and npm

## Setup Instructions

### 1. Initialize the Network
Run the network setup script to deploy the Hyperledger Fabric blockchain network:

```bash
./start-network.sh
```

This script will:
- Generate cryptographic materials for organizations
- Create the channel artifacts
- Launch the network containers (orderers, peers, CAs)
- Create and join channels
- Deploy the authentication chaincode

### 2. Common Issues and Troubleshooting

#### TLS Certificate Verification Issues
If encountering TLS certificate errors with the orderer (error message: `tls: failed to verify certificate: x509: certificate signed by unknown authority`), follow these steps:

1. Check the orderer CA logs:
```bash
docker logs ca.orderer.example.com
```

2. Clean up existing crypto materials:
```bash
rm -rf organizations/
```

3. Regenerate crypto materials:
```bash
cryptogen generate --config=./crypto-config.yaml --output="./organizations"
```

4. Verify certificate paths in `configtx.yaml`:
   - Ensure paths in the EtcdRaft section match your actual certificate locations

5. Regenerate blockchain artifacts:
```bash
configtxgen -profile ThreeOrgsOrdererGenesis -channelID system-channel -outputBlock ./channel-artifacts/genesis.block
configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ./channel-artifacts/chaichis-channel.tx -channelID chaichis-channel
```

6. Clean Docker volumes and restart:
```bash
docker-compose down -v
docker volume prune -f
docker-compose up -d
```

7. Verify all containers are running:
```bash
docker-compose ps -a
```

## Features
- Decentralized user authentication
- Immutable audit trail of authentication events
- Role-based access control
- Multi-factor authentication support
- Self-sovereign identity management

## Architecture
The system consists of multiple organizations participating in a permissioned blockchain network. Each organization runs peers that maintain a copy of the distributed ledger and execute chaincode for identity validation and authentication.

## Security Considerations
- All network communications are secured with TLS
- Private keys are stored securely in the cryptographic materials
- Smart contract logic enforces access control policies
- Certificate-based authentication for all network participants

# Comprehensive Hyperledger Fabric Chaincode Deployment Guide

This detailed guide covers the end-to-end process for deploying chaincodes in a Hyperledger Fabric network, with in-depth troubleshooting instructions and command references to resolve common issues.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Network Components](#network-components)
- [Chaincode Structure](#chaincode-structure)
- [Deployment Process Overview](#deployment-process-overview)
- [Detailed Deployment Steps](#detailed-deployment-steps)
  - [Step 1: Copy Chaincode to CLI Container](#step-1-copy-chaincode-to-cli-container)
  - [Step 2: Package Chaincode](#step-2-package-chaincode)
  - [Step 3: Install Chaincode](#step-3-install-chaincode)
  - [Step 4: Channel Creation and Management](#step-4-channel-creation-and-management)
  - [Step 5: Approve Chaincode for Organizations](#step-5-approve-chaincode-for-organizations)
  - [Step 6: Commit the Chaincode Definition](#step-6-commit-the-chaincode-definition)
  - [Step 7: Initialize the Chaincode](#step-7-initialize-the-chaincode)
- [Multi-Org Deployment Considerations](#multi-org-deployment-considerations)
- [Advanced Troubleshooting Guide](#advanced-troubleshooting-guide)
  - [Diagnosing Environment Issues](#diagnosing-environment-issues)
  - [Path and Directory Issues](#path-and-directory-issues)
  - [Dependency Problems](#dependency-problems)
  - [Code Compilation Errors](#code-compilation-errors)
  - [Channel Not Found](#channel-not-found)
  - [Organization Approval Issues](#organization-approval-issues)
  - [TLS Certificate Problems](#tls-certificate-problems)
  - [Chaincode Initialization Failures](#chaincode-initialization-failures)
  - [Docker and Container Issues](#docker-and-container-issues)
  - [Network Connection Problems](#network-connection-problems)
- [Maintenance and Update Procedures](#maintenance-and-update-procedures)
  - [Upgrading Chaincode](#upgrading-chaincode)
  - [Monitoring Chaincode](#monitoring-chaincode)
  - [Backup and Recovery](#backup-and-recovery)
- [Appendix: Command Reference](#appendix-command-reference)

## Prerequisites

Before deploying chaincodes, ensure the following components are ready:

- Operating system with Docker and Docker Compose installed
- Hyperledger Fabric binaries and Docker images (fabric-ca, fabric-peer, fabric-orderer, fabric-tools)
- Network configuration files:
  - `crypto-config.yaml` for cryptographic material generation
  - `configtx.yaml` for channel artifact generation
  - `docker-compose.yaml` for network service definitions
- Chaincode developed in Go, JavaScript, or Java with proper dependencies

## Network Components

A typical Hyperledger Fabric network consists of:

- **Certificate Authorities (CAs)**: One per organization, responsible for issuing identities
- **Orderer nodes**: Consensus service that sequences transactions into blocks
- **Peer nodes**: Host ledgers and chaincodes, validate and commit transactions
- **CLI container**: Administrative tool for network operations
- **Channels**: Private communication paths between organizations

Verify your network components are running with:

```bash
docker-compose ps -a
```

Expected output should show containers with STATUS "Up":
```
NAME                     IMAGE                               COMMAND                  SERVICE                  STATUS
ca.orderer.example.com   hyperledger/fabric-ca:latest        "sh -c 'fabric-ca-se…"   ca.orderer.example.com   Up
ca.org1.example.com      hyperledger/fabric-ca:latest        "sh -c 'fabric-ca-se…"   ca.org1.example.com      Up
orderer.example.com      hyperledger/fabric-orderer:latest   "orderer"                orderer.example.com      Up
peer0.org1.example.com   hyperledger/fabric-peer:latest      "peer node start"        peer0.org1.example.com   Up
cli                      hyperledger/fabric-tools:latest     "/bin/bash"              cli                      Up
```

## Chaincode Structure

Each chaincode should have the following structure, organized by programming language:

**Go chaincode example**:
```
chaincodes/
├── as-chaincode/
│   ├── as-chaincode.go    # Main chaincode file with contract implementation
│   ├── go.mod             # Dependency management file
│   └── go.sum             # Dependency checksums
```

A basic Go chaincode file should look like:

```go
package main

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing assets
type SmartContract struct {
	contractapi.Contract
}

// Initialize is called during chaincode instantiation to set up any data
func (s *SmartContract) Initialize() error {
	return nil
}

// Create adds a new asset to the world state
func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, id string, value string) error {
	// Implementation
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
```

The go.mod file should include necessary dependencies:

```
module github.com/chaincodes/as-chaincode

go 1.17

require github.com/hyperledger/fabric-contract-api-go v1.2.0
```

## Deployment Process Overview

The chaincode deployment process in Fabric 2.x involves these steps:

1. **Copy chaincode** files to the CLI container
2. **Package the chaincode** into a deployable format
3. **Install the chaincode** on relevant peer nodes
4. **Create and join the channel** (if not already done)
5. **Approve chaincode** definition for each organization
6. **Commit the chaincode** definition to the channel
7. **Initialize the chaincode** through an invoke request

## Detailed Deployment Steps

### Step 1: Copy Chaincode to CLI Container

First, access your project directory containing the chaincode:

```bash
cd /path/to/project/chaincodes
```

Copy your chaincode directories from the host machine to the CLI container:

```bash
docker cp chaincodes/as-chaincode cli:/opt/gopath/src/github.com/chaincodes/
docker cp chaincodes/tgs-chaincode cli:/opt/gopath/src/github.com/chaincodes/
docker cp chaincodes/isv-chaincode cli:/opt/gopath/src/github.com/chaincodes/
```

Verify the files were copied correctly:

```bash
docker exec -it cli bash
ls -la /opt/gopath/src/github.com/chaincodes/
```

Expected output should show your chaincode directories with proper file structure.

### Step 2: Package Chaincode

Enter the CLI container:

```bash
docker exec -it cli bash
```

Set the environment variables for the organization that will submit the package:

```bash
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
```

For Go chaincodes, resolve dependencies to prevent build issues:

```bash
cd /opt/gopath/src/github.com/chaincodes/as-chaincode/
GO111MODULE=on go mod vendor
cd /opt/gopath/src/github.com/hyperledger/fabric/peer/
```

Package the chaincode:

```bash
peer lifecycle chaincode package as-chaincode.tar.gz --path /opt/gopath/src/github.com/chaincodes/as-chaincode/ --lang golang --label as-chaincode_1.0
```

The command creates a package file `as-chaincode.tar.gz` in the current directory.

### Step 3: Install Chaincode

Install the packaged chaincode on the peer:

```bash
peer lifecycle chaincode install as-chaincode.tar.gz
```

Successful installation will show output like:
```
2025-03-10 05:30:32.191 UTC 0001 INFO [cli.lifecycle.chaincode] submitInstallProposal -> Installed remotely: response:<status:200 payload:"\nQas-chaincode_1.0:3ca928fecee6d0d3a3d4e3b94ff71aa5834d0602bfbe53a0880cbf7de7956d17\022\020as-chaincode_1.0" >
```

After installation, query to obtain the package ID:

```bash
peer lifecycle chaincode queryinstalled
```

Output example:
```
Installed chaincodes on peer:
Package ID: as-chaincode_1.0:3ca928fecee6d0d3a3d4e3b94ff71aa5834d0602bfbe53a0880cbf7de7956d17, Label: as-chaincode_1.0
```

Set the package ID as an environment variable for use in later commands:

```bash
export CC_PACKAGE_ID=as-chaincode_1.0:3ca928fecee6d0d3a3d4e3b94ff71aa5834d0602bfbe53a0880cbf7de7956d17
```

### Step 4: Channel Creation and Management

#### Check Existing Channels

Verify what channels currently exist:

```bash
peer channel list
```

#### Create a New Channel

If the channel doesn't exist, create it:

```bash
cd /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts
peer channel create -o orderer.example.com:7050 -c chaichis-channel -f ./chaichis-channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
```

This will create a `chaichis-channel.block` genesis block file in the current directory.

#### Join Peers to the Channel

Have each organization's peers join the channel:

```bash
# For Org1
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer channel join -b chaichis-channel.block

# For Org2
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer channel join -b chaichis-channel.block

# For Org3
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

peer channel join -b chaichis-channel.block
```

#### Update Anchor Peers

Configure anchor peers for each organization:

```bash
# For Org1
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer channel update -o orderer.example.com:7050 -c chaichis-channel -f ./Org1MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# For Org2
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer channel update -o orderer.example.com:7050 -c chaichis-channel -f ./Org2MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# For Org3
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

peer channel update -o orderer.example.com:7050 -c chaichis-channel -f ./Org3MSPanchors.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
```

#### Verify Channel Configuration

To verify channel status and peer membership:

```bash
# Check what channels a peer has joined
peer channel list

# Get detailed channel information
peer channel getinfo -c chaichis-channel
```

### Step 5: Approve Chaincode for Organizations

Each organization must install and approve the chaincode.

For Org1:
```bash
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

# Ensure package ID environment variable is set
echo $CC_PACKAGE_ID

# Approve the chaincode definition
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name as-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

For Org2:
```bash
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

# Install the chaincode on Org2's peer
peer lifecycle chaincode install as-chaincode.tar.gz

# Approve the chaincode definition
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name as-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

For Org3:
```bash
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

# Install the chaincode on Org3's peer
peer lifecycle chaincode install as-chaincode.tar.gz

# Approve the chaincode definition
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name as-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

Check the commit readiness to ensure all organizations have approved:

```bash
peer lifecycle chaincode checkcommitreadiness --channelID chaichis-channel --name as-chaincode --version 1.0 --sequence 1 --init-required
```

This should show:
```
Chaincode definition for chaincode 'as-chaincode', version '1.0', sequence '1' on channel 'chaichis-channel' approval status by org:
Org1MSP: true
Org2MSP: true
Org3MSP: true
```

### Step 6: Commit the Chaincode Definition

Once all organizations have approved, commit the chaincode definition to the channel:

```bash
peer lifecycle chaincode commit -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name as-chaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt --version 1.0 --sequence 1 --init-required
```

If successful, the command will return without errors.

Verify the committed chaincode definitions on the channel:

```bash
peer lifecycle chaincode querycommitted --channelID chaichis-channel
```

### Step 7: Initialize the Chaincode

After successfully committing the chaincode definition, initialize the chaincode:

```bash
peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chaichis-channel -n as-chaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --isInit -c '{"function":"Initialize","Args":[]}'
```

A successful initialization should show:
```
2025-03-10 05:45:32.191 UTC 0001 INFO [chaincodeCmd] chaincodeInvokeOrQuery -> Chaincode invoke successful. result: status:200
```

### Step 8: Deploy Additional Chaincodes (TGS and ISV)

After successfully deploying the first chaincode (AS), follow the same process to deploy additional chaincodes for other organizations.

#### Deploy TGS Chaincode for Org2

```bash
# Set environment variables for organization 2 (TGS)
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

# Package the TGS chaincode
peer lifecycle chaincode package tgs-chaincode.tar.gz --path /opt/gopath/src/github.com/chaincodes/tgs-chaincode/ --lang golang --label tgs-chaincode_1.0

# Install the TGS chaincode on peer0.org2
peer lifecycle chaincode install tgs-chaincode.tar.gz

# Get the package ID
peer lifecycle chaincode queryinstalled
```

After obtaining the package ID, set it as an environment variable:

```bash
# Set the package ID variable (replace with your actual ID)
export CC_PACKAGE_ID=tgs-chaincode_1.0:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Approve chaincode for Org2
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name tgs-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

#### Install and Approve TGS Chaincode for All Organizations

The TGS chaincode needs to be installed on peers from all organizations and approved by each organization:

```bash
# Install and approve for Org1
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode install tgs-chaincode.tar.gz
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name tgs-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1

# Install and approve for Org3
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

peer lifecycle chaincode install tgs-chaincode.tar.gz
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name tgs-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

#### Commit and Initialize TGS Chaincode

Once all organizations have approved, commit and initialize the TGS chaincode:

```bash
# Check commit readiness
peer lifecycle chaincode checkcommitreadiness --channelID chaichis-channel --name tgs-chaincode --version 1.0 --sequence 1 --init-required

# Commit the chaincode definition
peer lifecycle chaincode commit -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name tgs-chaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt --version 1.0 --sequence 1 --init-required

# Switch back to Org2 for initialization
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

# Initialize the chaincode
peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chaichis-channel -n tgs-chaincode --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --isInit -c '{"function":"Initialize","Args":[]}'
```

#### Deploy ISV Chaincode for Org3

Follow a similar process to deploy the ISV chaincode for Organization 3:

```bash
# Set environment variables for organization 3 (ISV)
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

# Package the ISV chaincode
peer lifecycle chaincode package isv-chaincode.tar.gz --path /opt/gopath/src/github.com/chaincodes/isv-chaincode/ --lang golang --label isv-chaincode_1.0

# Install the ISV chaincode on peer0.org3
peer lifecycle chaincode install isv-chaincode.tar.gz

# Get the package ID
peer lifecycle chaincode queryinstalled
```

After obtaining the package ID, set it as an environment variable:

```bash
# Set the package ID variable (replace with your actual ID)
export CC_PACKAGE_ID=isv-chaincode_1.0:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Approve chaincode for Org3
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name isv-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

#### Install and Approve ISV Chaincode for All Organizations

```bash
# Install and approve for Org1
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

peer lifecycle chaincode install isv-chaincode.tar.gz
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name isv-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1

# Install and approve for Org2
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051

peer lifecycle chaincode install isv-chaincode.tar.gz
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name isv-chaincode --version 1.0 --init-required --package-id $CC_PACKAGE_ID --sequence 1
```

#### Commit and Initialize ISV Chaincode

```bash
# Check commit readiness
peer lifecycle chaincode checkcommitreadiness --channelID chaichis-channel --name isv-chaincode --version 1.0 --sequence 1 --init-required

# Commit the chaincode definition
peer lifecycle chaincode commit -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --channelID chaichis-channel --name isv-chaincode --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt --version 1.0 --sequence 1 --init-required

# Switch back to Org3 for initialization
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051

# Initialize the chaincode
peer chaincode invoke -o orderer.example.com:7050 --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C chaichis-channel -n isv-chaincode --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt --isInit -c '{"function":"Initialize","Args":[]}'
```

## Multi-Org Deployment Considerations

For networks with multiple organizations, follow these best practices:

1. **Install on all relevant peers**: Each organization should install the chaincode on their peers that will execute transactions
   
2. **Coordinate on chaincode parameters**: All organizations must approve the same chaincode definition (name, version, and sequence)
   
3. **Endorsement policies**: Define appropriate endorsement policies based on business requirements
   ```bash
   # Example: Requiring endorsement from Org1 AND Org2
   --signature-policy "AND('Org1MSP.peer','Org2MSP.peer')"
   
   # Example: Requiring endorsement from any 2 organizations out of 3
   --signature-policy "OutOf(2,'Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')"
   ```

4. **Sequence for upgrades**: Increment the sequence number for each upgrade
   
5. **Private data collections**: Configure private data collections if required
   ```bash
   --collections-config /path/to/collections_config.json
   ```

## Advanced Troubleshooting Guide

### Diagnosing Environment Issues

#### Check Environment Variables

```bash
# View all environment variables
env | grep CORE_PEER

# Verify specific variables
echo $CORE_PEER_LOCALMSPID
echo $CORE_PEER_MSPCONFIGPATH
```

#### Verify Container Status

```bash
# Check container status
docker ps -a | grep fabric

# Inspect container logs
docker logs peer0.org1.example.com
```

#### Test Network Connectivity

```bash
# Test peer connectivity from CLI container
ping peer0.org1.example.com

# Check peer port accessibility
telnet peer0.org1.example.com 7051
```

### Path and Directory Issues

#### Issue: Incorrect paths to crypto materials

Error:
```
Cannot run peer because cannot init crypto, specified path "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" does not exist or cannot be accessed
```

#### Solution:

1. Check the actual directory structure:
   ```bash
   ls -la /opt/gopath/src/github.com/hyperledger/fabric/peer/
   ```

2. Verify the correct crypto material location:
   ```bash
   find /opt/gopath/src/github.com/hyperledger/fabric/peer/ -name "msp" -type d
   ```

3. Update environment variables with correct paths:
   ```bash
   # If materials are in organizations directory instead of crypto
   export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
   ```

4. Verify the certificate files exist:
   ```bash
   ls -la $CORE_PEER_MSPCONFIGPATH
   ls -la $CORE_PEER_TLS_ROOTCERT_FILE
   ```

### Dependency Problems

#### Issue: Missing Go dependencies

Error:
```
could not build chaincode: docker build failed: docker image build failed: docker build failed: Error returned from build: 1 "cannot find module providing package github.com/hyperledger/fabric-contract-api-go/contractapi: import lookup disabled by -mod=readonly"
```

#### Solution:

1. Fix dependencies locally first:
   ```bash
   # Exit CLI container
   exit
   
   # Navigate to local chaincode directory
   cd /path/to/chaincode/directory
   
   # Update dependencies
   go mod tidy
   ```

2. Rebuild the go.mod file if necessary:
   ```bash
   # Initialize a new go.mod file
   go mod init github.com/chaincodes/as-chaincode
   
   # Add required dependencies
   go get github.com/hyperledger/fabric-contract-api-go@v1.2.0
   go get github.com/hyperledger/fabric-chaincode-go@v0.0.0-20230731094759-d626e9ab09b9
   ```

3. Copy updated chaincode to CLI container:
   ```bash
   docker cp chaincodes/as-chaincode cli:/opt/gopath/src/github.com/chaincodes/
   ```

4. Inside the CLI container, download dependencies before packaging:
   ```bash
   docker exec -it cli bash
   cd /opt/gopath/src/github.com/chaincodes/as-chaincode/
   GO111MODULE=on go mod vendor
   ```

5. Check vendor directory was created:
   ```bash
   ls -la vendor/
   ```

### Code Compilation Errors

#### Issue: Go compilation errors

Error:
```
./as-chaincode.go:12:2: "math/big" imported and not used
```

#### Solution:

1. Examine the chaincode file:
   ```bash
   cat /opt/gopath/src/github.com/chaincodes/as-chaincode/as-chaincode.go
   ```

2. Exit CLI container and edit the file locally:
   ```bash
   exit
   vim chaincodes/as-chaincode/as-chaincode.go
   ```

3. Remove unused imports or fix other compilation issues

4. Copy corrected file back to CLI container:
   ```bash
   docker cp chaincodes/as-chaincode cli:/opt/gopath/src/github.com/chaincodes/
   ```

5. Re-attempt packaging and installation with corrected code

### Channel Not Found

#### Issue: Referenced channel doesn't exist

Error:
```
Error: proposal failed with status: 500 - channel 'chaichis-channel' not found
```

#### Solution:

1. Check existing channels:
   ```bash
   peer channel list
   ```

2. If the channel doesn't exist, create it:
   ```bash
   # Make sure channel creation transaction exists
   ls -la /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/
   
   # Create the channel
   peer channel create -o orderer.example.com:7050 -c chaichis-channel -f ./channel-artifacts/chaichis-channel.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
   ```

3. Join peers to the channel:
   ```bash
   # For each organization, set environment variables and join
   peer channel join -b chaichis-channel.block
   ```

4. Verify channel creation:
   ```bash
   peer channel list
   peer channel getinfo -c chaichis-channel
   ```

5. Examine channel blocks for debugging:
   ```bash
   peer channel fetch config -c chaichis-channel
   ```

### Organization Approval Issues

#### Issue: Not all organizations have approved the chaincode

Error:
```
Error: proposal failed with status: 500 - failed to invoke backing implementation of 'CommitChaincodeDefinition': chaincodedefinition not agreed to by this org (Org1MSP)
```

#### Solution:

1. Check approval status of all organizations:
   ```bash
   peer lifecycle chaincode checkcommitreadiness --channelID chaichis-channel --name as-chaincode --version 1.0 --sequence 1 --init-required
   ```

2. If any organization shows `false`, approve for that organization:
   ```bash
   # Switch to the organization context
   export CORE_PEER_LOCALMSPID="Org1MSP"
   export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/

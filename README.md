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


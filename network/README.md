# Network Configuration

📍 **Location**: `network/`
🔗 **Parent**: [Main README](../README.md)

## Overview
This directory contains all configuration files and scripts to deploy and manage the Hyperledger Fabric blockchain network.

## Directory Structure
```
network/
├── config/           ← Network configuration files
│   ├── crypto-config.yaml
│   ├── configtx.yaml
│   └── docker-compose-network.yaml
└── scripts/          ← Management scripts
    ├── network.sh
    ├── deploy-chaincode.sh
    └── verify-channel.sh
```

## Quick Start
```bash
# Start network
./scripts/network.sh up

# Create channel
./scripts/network.sh createChannel

# Deploy chaincodes
./scripts/deploy-chaincode.sh as
./scripts/deploy-chaincode.sh tgs
./scripts/deploy-chaincode.sh isv
```

## Learn More
- 📁 [Configuration Files](config/README.md)
- 📜 [Management Scripts](scripts/README.md)
- 📚 [HOW_IT_WORKS](../HOW_IT_WORKS.md)

📍 **Navigation**: [Main](../README.md) | [Config →](config/README.md)

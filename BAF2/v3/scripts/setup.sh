#!/bin/bash
# setup.sh - Environment setup script for the authentication framework

# Exit on error
set -e

YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up authentication framework environment...${NC}"

# Create required directories
mkdir -p config keys wallet sessions

# Check if connection profile exists
if [ ! -f "config/connection-profile.json" ]; then
    echo -e "${YELLOW}Connection profile not found. Creating a sample connection profile...${NC}"
    
    # Create sample connection profile
    cat > config/connection-profile.json << EOF
{
    "name": "chaichis-network",
    "version": "1.0.0",
    "client": {
        "organization": "Org1",
        "connection": {
            "timeout": {
                "peer": {
                    "endorser": "300"
                },
                "orderer": "300"
            }
        }
    },
    "channels": {
        "chaichis-channel": {
            "orderers": [
                "orderer.example.com"
            ],
            "peers": {
                "peer0.org1.example.com": {
                    "endorsingPeer": true,
                    "chaincodeQuery": true,
                    "ledgerQuery": true,
                    "eventSource": true
                },
                "peer0.org2.example.com": {
                    "endorsingPeer": true,
                    "chaincodeQuery": true,
                    "ledgerQuery": true,
                    "eventSource": true
                },
                "peer0.org3.example.com": {
                    "endorsingPeer": true,
                    "chaincodeQuery": true,
                    "ledgerQuery": true,
                    "eventSource": true
                }
            }
        }
    },
    "organizations": {
        "Org1": {
            "mspid": "Org1MSP",
            "peers": [
                "peer0.org1.example.com",
                "peer1.org1.example.com",
                "peer2.org1.example.com"
            ],
            "certificateAuthorities": [
                "ca.org1.example.com"
            ]
        },
        "Org2": {
            "mspid": "Org2MSP",
            "peers": [
                "peer0.org2.example.com",
                "peer1.org2.example.com",
                "peer2.org2.example.com"
            ],
            "certificateAuthorities": [
                "ca.org2.example.com"
            ]
        },
        "Org3": {
            "mspid": "Org3MSP",
            "peers": [
                "peer0.org3.example.com",
                "peer1.org3.example.com",
                "peer2.org3.example.com"
            ],
            "certificateAuthorities": [
                "ca.org3.example.com"
            ]
        }
    },
    "orderers": {
        "orderer.example.com": {
            "url": "grpcs://orderer.example.com:7050",
            "tlsCACerts": {
                "path": "organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "orderer.example.com"
            }
        }
    },
    "peers": {
        "peer0.org1.example.com": {
            "url": "grpcs://peer0.org1.example.com:7051",
            "tlsCACerts": {
                "path": "organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org1.example.com"
            }
        },
        "peer0.org2.example.com": {
            "url": "grpcs://peer0.org2.example.com:9051",
            "tlsCACerts": {
                "path": "organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org2.example.com"
            }
        },
        "peer0.org3.example.com": {
            "url": "grpcs://peer0.org3.example.com:13051",
            "tlsCACerts": {
                "path": "organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org3.example.com"
            }
        }
    },
    "certificateAuthorities": {
        "ca.org1.example.com": {
            "url": "https://ca.org1.example.com:7054",
            "caName": "ca.org1.example.com",
            "tlsCACerts": {
                "path": "organizations/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem"
            },
            "httpOptions": {
                "verify": false
            }
        }
    }
}
EOF
    
    echo -e "${GREEN}Created sample connection profile. Please edit it to match your network configuration.${NC}"
else
    echo -e "${GREEN}Connection profile already exists.${NC}"
fi

# Check for previous version directory
if [ -d "../v1" ] || [ -d "../v2" ]; then
    echo -e "${YELLOW}Found previous version directories. Checking for existing keys...${NC}"
    
    # Check v1 keys
    if [ -d "../v1/keys" ]; then
        echo -e "${YELLOW}Copying keys from v1...${NC}"
        cp -r ../v1/keys/* keys/ 2>/dev/null || true
    fi
    
    # Check v2 keys
    if [ -d "../v2/keys" ]; then
        echo -e "${YELLOW}Copying keys from v2...${NC}"
        cp -r ../v2/keys/* keys/ 2>/dev/null || true
    fi
    
    # Check connection profile in v1
    if [ -f "../v1/connection-profile.json" ] && [ ! -f "config/connection-profile.json" ]; then
        echo -e "${YELLOW}Copying connection profile from v1...${NC}"
        cp ../v1/connection-profile.json config/
    fi
    
    # Check connection profile in v2
    if [ -f "../v2/connection-profile.json" ] && [ ! -f "config/connection-profile.json" ]; then
        echo -e "${YELLOW}Copying connection profile from v2...${NC}"
        cp ../v2/connection-profile.json config/
    fi
fi

# Make scripts executable
chmod +x scripts/*.sh

echo -e "${GREEN}Environment setup completed successfully!${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Edit config/connection-profile.json to match your network configuration"
echo -e "2. Run '${BLUE}./scripts/init-wallet.sh${NC}' to initialize your wallet with identity credentials"
echo -e "3. Run '${BLUE}make build${NC}' to build the authcli binary"

#!/bin/bash
# Script to extract raw certificates and create a proper connection profile

set -e  # Exit on any error

echo "Extracting raw certificates..."

# Create directories to store certificates
mkdir -p crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls
mkdir -p crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls
mkdir -p crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls
mkdir -p crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls
mkdir -p crypto-config/peerOrganizations/org1.example.com/ca
mkdir -p crypto-config/peerOrganizations/org2.example.com/ca
mkdir -p crypto-config/peerOrganizations/org3.example.com/ca

# Extract raw certificates
docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt" > crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt

docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" > crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem" > crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem

docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" > crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt

docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem" > crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem

docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt" > crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/ca/ca.org3.example.com-cert.pem" > crypto-config/peerOrganizations/org3.example.com/ca/ca.org3.example.com-cert.pem

echo "Creating connection profile with certificate paths..."

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
                "path": "crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "orderer.example.com",
                "hostnameOverride": "orderer.example.com"
            }
        }
    },
    "peers": {
        "peer0.org1.example.com": {
            "url": "grpcs://peer0.org1.example.com:7051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org1.example.com",
                "hostnameOverride": "peer0.org1.example.com"
            }
        },
        "peer1.org1.example.com": {
            "url": "grpcs://peer1.org1.example.com:8051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer1.org1.example.com",
                "hostnameOverride": "peer1.org1.example.com"
            }
        },
        "peer2.org1.example.com": {
            "url": "grpcs://peer2.org1.example.com:11051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer2.org1.example.com",
                "hostnameOverride": "peer2.org1.example.com"
            }
        },
        "peer0.org2.example.com": {
            "url": "grpcs://peer0.org2.example.com:9051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org2.example.com",
                "hostnameOverride": "peer0.org2.example.com"
            }
        },
        "peer1.org2.example.com": {
            "url": "grpcs://peer1.org2.example.com:10051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer1.org2.example.com",
                "hostnameOverride": "peer1.org2.example.com"
            }
        },
        "peer2.org2.example.com": {
            "url": "grpcs://peer2.org2.example.com:12051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer2.org2.example.com",
                "hostnameOverride": "peer2.org2.example.com"
            }
        },
        "peer0.org3.example.com": {
            "url": "grpcs://peer0.org3.example.com:13051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org3.example.com",
                "hostnameOverride": "peer0.org3.example.com"
            }
        },
        "peer1.org3.example.com": {
            "url": "grpcs://peer1.org3.example.com:14051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer1.org3.example.com",
                "hostnameOverride": "peer1.org3.example.com"
            }
        },
        "peer2.org3.example.com": {
            "url": "grpcs://peer2.org3.example.com:15051",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer2.org3.example.com",
                "hostnameOverride": "peer2.org3.example.com"
            }
        }
    },
    "certificateAuthorities": {
        "ca.org1.example.com": {
            "url": "https://ca.org1.example.com:7054",
            "caName": "ca.org1.example.com",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem"
            },
            "httpOptions": {
                "verify": false
            }
        },
        "ca.org2.example.com": {
            "url": "https://ca.org2.example.com:8054",
            "caName": "ca.org2.example.com",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem"
            },
            "httpOptions": {
                "verify": false
            }
        },
        "ca.org3.example.com": {
            "url": "https://ca.org3.example.com:9054",
            "caName": "ca.org3.example.com",
            "tlsCACerts": {
                "path": "crypto-config/peerOrganizations/org3.example.com/ca/ca.org3.example.com-cert.pem"
            },
            "httpOptions": {
                "verify": false
            }
        }
    }
}
EOF

echo "Connection profile created with raw certificate paths"

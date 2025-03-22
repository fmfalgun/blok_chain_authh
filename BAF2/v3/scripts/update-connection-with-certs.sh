#!/bin/bash
# Script to update connection profile with embedded certificates

set -e  # Exit on any error

echo "Creating connection profile with embedded certificates..."

# Create a temp directory
mkdir -p /tmp/fabric-certs

# Extract certificates directly and format them in one step
ORDERER_CA=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt" | sed 's/$/\\n/' | tr -d '\n')

ORG1_CA=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" | sed 's/$/\\n/' | tr -d '\n')

ORG1_CA_PEM=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/ca/ca.org1.example.com-cert.pem" | sed 's/$/\\n/' | tr -d '\n')

ORG2_CA=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" | sed 's/$/\\n/' | tr -d '\n')

ORG2_CA_PEM=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/ca/ca.org2.example.com-cert.pem" | sed 's/$/\\n/' | tr -d '\n')

ORG3_CA=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt" | sed 's/$/\\n/' | tr -d '\n')

ORG3_CA_PEM=$(docker exec cli bash -c "cat /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/ca/ca.org3.example.com-cert.pem" | sed 's/$/\\n/' | tr -d '\n')

# Create the connection profile with actual certificate content
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
                "pem": "${ORDERER_CA}"
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
                "pem": "${ORG1_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org1.example.com",
                "hostnameOverride": "peer0.org1.example.com"
            }
        },
        "peer1.org1.example.com": {
            "url": "grpcs://peer1.org1.example.com:8051",
            "tlsCACerts": {
                "pem": "${ORG1_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer1.org1.example.com",
                "hostnameOverride": "peer1.org1.example.com"
            }
        },
        "peer2.org1.example.com": {
            "url": "grpcs://peer2.org1.example.com:11051",
            "tlsCACerts": {
                "pem": "${ORG1_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer2.org1.example.com",
                "hostnameOverride": "peer2.org1.example.com"
            }
        },
        "peer0.org2.example.com": {
            "url": "grpcs://peer0.org2.example.com:9051",
            "tlsCACerts": {
                "pem": "${ORG2_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org2.example.com",
                "hostnameOverride": "peer0.org2.example.com"
            }
        },
        "peer1.org2.example.com": {
            "url": "grpcs://peer1.org2.example.com:10051",
            "tlsCACerts": {
                "pem": "${ORG2_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer1.org2.example.com",
                "hostnameOverride": "peer1.org2.example.com"
            }
        },
        "peer2.org2.example.com": {
            "url": "grpcs://peer2.org2.example.com:12051",
            "tlsCACerts": {
                "pem": "${ORG2_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer2.org2.example.com",
                "hostnameOverride": "peer2.org2.example.com"
            }
        },
        "peer0.org3.example.com": {
            "url": "grpcs://peer0.org3.example.com:13051",
            "tlsCACerts": {
                "pem": "${ORG3_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer0.org3.example.com",
                "hostnameOverride": "peer0.org3.example.com"
            }
        },
        "peer1.org3.example.com": {
            "url": "grpcs://peer1.org3.example.com:14051",
            "tlsCACerts": {
                "pem": "${ORG3_CA}"
            },
            "grpcOptions": {
                "ssl-target-name-override": "peer1.org3.example.com",
                "hostnameOverride": "peer1.org3.example.com"
            }
        },
        "peer2.org3.example.com": {
            "url": "grpcs://peer2.org3.example.com:15051",
            "tlsCACerts": {
                "pem": "${ORG3_CA}"
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
                "pem": ["${ORG1_CA_PEM}"]
            },
            "httpOptions": {
                "verify": false
            }
        },
        "ca.org2.example.com": {
            "url": "https://ca.org2.example.com:8054",
            "caName": "ca.org2.example.com",
            "tlsCACerts": {
                "pem": ["${ORG2_CA_PEM}"]
            },
            "httpOptions": {
                "verify": false
            }
        },
        "ca.org3.example.com": {
            "url": "https://ca.org3.example.com:9054",
            "caName": "ca.org3.example.com",
            "tlsCACerts": {
                "pem": ["${ORG3_CA_PEM}"]
            },
            "httpOptions": {
                "verify": false
            }
        }
    }
}
EOF

echo "Connection profile with embedded certificates created successfully!"

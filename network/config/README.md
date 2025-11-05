# Network Configuration

📍 **Location**: `network/config/`
🔗 **Parent**: [Network Overview](../README.md)
📚 **Related**: [Network Scripts](../scripts/README.md) | [Deployment Guide](../../docs/deployment/PRODUCTION_DEPLOYMENT.md)

## Overview

This directory contains all configuration files required to set up and run the Hyperledger Fabric network. These files define the network topology, cryptographic materials, consensus mechanism, and Docker container orchestration. Think of this as the "blueprint" - everything needed to recreate the exact network architecture.

## Purpose

**Why do these configuration files exist?**

1. **Declarative Infrastructure**: All network configuration is code (Infrastructure as Code)
2. **Reproducibility**: Same configs → same network, every time
3. **Version Control**: Track changes to network architecture over time
4. **Multi-Environment**: Same configs work for dev, staging, production with minimal changes
5. **Automation**: Scripts can parse these configs to automate network operations

## Directory Structure

```
config/
├── crypto-config.yaml             # PKI structure definition
├── configtx.yaml                  # Channel configuration
├── docker-compose-network.yaml    # Container orchestration
└── README.md                     # This file
```

## Files Explained

### 1. crypto-config.yaml

**Purpose**: Defines the PKI (Public Key Infrastructure) hierarchy for the network.

**What it generates**:
- CA certificates for each organization
- Peer certificates and private keys
- Orderer certificates and private keys
- Admin user certificates
- TLS certificates for secure communication

**Technology**: Used by `cryptogen` tool (part of Hyperledger Fabric binaries)

**Why this file exists**:
- **Trust Model**: Defines which organizations exist and how they trust each other
- **Identity Management**: Each entity (peer, orderer, user) gets unique identity
- **TLS Security**: Enables encrypted communication between all network components

**Structure Breakdown**:

```yaml
OrdererOrgs:
  - Name: Orderer
    Domain: example.com
    Specs:
      - Hostname: orderer
```

**Explanation**:
- **Name**: Organization name for the orderer
- **Domain**: DNS domain (not used in test networks, but required for production)
- **Specs**: List of orderer nodes (we have 1 orderer named "orderer.example.com")

**Generated Certificates**:
```
crypto-config/
├── ordererOrganizations/
│   └── example.com/
│       ├── ca/              # Certificate Authority cert
│       ├── msp/             # Organization MSP definition
│       ├── orderers/
│       │   └── orderer.example.com/
│       │       ├── msp/     # Orderer identity
│       │       └── tls/     # TLS certificates
│       └── users/
│           └── Admin@example.com/
│               └── msp/     # Admin identity
```

```yaml
PeerOrgs:
  - Name: Org1
    Domain: org1.example.com
    Template:
      Count: 2
    Users:
      Count: 2
```

**Explanation**:
- **Name**: Organization name (Org1, Org2, Org3 in our network)
- **Domain**: DNS domain for the organization
- **Template.Count**: Number of peer nodes (2 peers per organization)
- **Users.Count**: Number of non-admin users to generate (2 per org)

**Why 3 organizations?**
- **Separation of Duties**: Each chaincode runs on separate org
  - Org1: Hosts AS (Authentication Server) chaincode
  - Org2: Hosts TGS (Ticket Granting Server) chaincode
  - Org3: Hosts ISV (IoT Service Validator) chaincode
- **Realistic Simulation**: Mimics multi-party blockchain scenarios
- **Endorsement Policies**: Can require endorsements from multiple orgs

**Why 2 peers per organization?**
- **High Availability**: If one peer fails, the other continues
- **Load Balancing**: Distribute query load across peers
- **Disaster Recovery**: Ledger redundancy within organization

**Generated Certificates for Org1**:
```
crypto-config/
└── peerOrganizations/
    └── org1.example.com/
        ├── ca/              # CA certificate for Org1
        ├── msp/             # Org1 MSP definition
        ├── peers/
        │   ├── peer0.org1.example.com/
        │   │   ├── msp/     # Peer identity
        │   │   └── tls/     # TLS certificates
        │   └── peer1.org1.example.com/
        │       ├── msp/
        │       └── tls/
        └── users/
            ├── Admin@org1.example.com/
            │   └── msp/     # Org admin identity
            ├── User1@org1.example.com/
            │   └── msp/
            └── User2@org1.example.com/
                └── msp/
```

**How to generate certificates**:
```bash
# From network directory
cryptogen generate --config=./config/crypto-config.yaml
```

**When to regenerate**:
- Adding new organizations
- Adding new peers
- Certificate expiration (Fabric certs typically expire after 1 year)
- Testing scenarios (fresh start)

**Production Alternative**:
- **Fabric CA**: More realistic CA server (supports certificate renewal, revocation)
- **cryptogen**: Quick and easy for development, but limited in production

---

### 2. configtx.yaml

**Purpose**: Defines channel configuration, consensus mechanism, and network policies.

**What it generates**:
- Genesis block for orderer
- Channel configuration transaction
- Anchor peer configurations

**Technology**: Used by `configtxgen` tool (part of Hyperledger Fabric binaries)

**Why this file exists**:
- **Channel Governance**: Defines rules for channel creation and updates
- **Consensus Configuration**: Specifies EtcdRaft consensus parameters
- **Policies**: Defines who can do what (endorsement, lifecycle, etc.)
- **Capabilities**: Enables specific Fabric features (must match Fabric version)

**Structure Breakdown**:

#### Organizations Section
```yaml
Organizations:
  - &OrdererOrg
      Name: OrdererMSP
      ID: OrdererMSP
      MSPDir: ../crypto-config/ordererOrganizations/example.com/msp
      Policies:
          Readers:
              Type: Signature
              Rule: "OR('OrdererMSP.member')"
          Writers:
              Type: Signature
              Rule: "OR('OrdererMSP.member')"
          Admins:
              Type: Signature
              Rule: "OR('OrdererMSP.admin')"
```

**Explanation**:
- **&OrdererOrg**: YAML anchor for reuse elsewhere in file
- **Name/ID**: MSP (Membership Service Provider) identifier
- **MSPDir**: Path to MSP certificates (generated by cryptogen)
- **Policies**: Access control rules using MSP signatures

**Policy Types**:
- **Signature**: Requires signatures from MSP members
- **ImplicitMeta**: Aggregates policies from sub-groups (e.g., "MAJORITY Admins")

**Common Rules**:
- `OR('OrdererMSP.admin')`: Any orderer admin can perform action
- `OR('Org1MSP.member')`: Any Org1 member can perform action
- `AND('Org1MSP.admin', 'Org2MSP.admin')`: Requires both Org1 AND Org2 admins

#### Capabilities Section
```yaml
Capabilities:
    Channel: &ChannelCapabilities
        V2_0: true
    Orderer: &OrdererCapabilities
        V2_0: true
    Application: &ApplicationCapabilities
        V2_0: true
```

**Explanation**:
- **V2_0**: Enables Hyperledger Fabric 2.0+ features
  - New chaincode lifecycle
  - Private data enhancements
  - FabToken support

**Why capabilities matter**:
- **Version Compatibility**: Ensures all peers/orderers support required features
- **Upgrade Safety**: Prevents using new features on old nodes
- **Feature Gates**: Can enable/disable specific Fabric capabilities

**Our network uses**: V2_0 capabilities (compatible with Fabric 2.5.5)

#### Orderer Section
```yaml
Orderer: &OrdererDefaults
    OrdererType: etcdraft
    Addresses:
        - orderer.example.com:7050
    EtcdRaft:
        Consenters:
            - Host: orderer.example.com
              Port: 7050
              ClientTLSCert: ../crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
              ServerTLSCert: ../crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/server.crt
    BatchTimeout: 2s
    BatchSize:
        MaxMessageCount: 500
        AbsoluteMaxBytes: 10 MB
        PreferredMaxBytes: 2 MB
```

**Explanation**:
- **OrdererType**: Consensus algorithm
  - `etcdraft`: Crash Fault Tolerant (CFT) consensus (recommended for production)
  - Alternatives: `solo` (single orderer, dev only), `kafka` (deprecated)

**Why EtcdRaft?**
- **Production Ready**: Used by major Fabric deployments
- **Fault Tolerant**: Tolerates (n-1)/2 failures (e.g., 3 orderers tolerate 1 failure)
- **Performance**: ~3000-10000 TPS depending on hardware
- **No External Dependencies**: Unlike Kafka, Raft is built into orderer

**Consenters**: List of Raft cluster members (we have 1 orderer, production should have 3-5)

**Batch Settings**:
- **BatchTimeout**: Max time to wait before creating block (2 seconds)
- **MaxMessageCount**: Max transactions per block (500)
- **AbsoluteMaxBytes**: Max block size (10 MB)
- **PreferredMaxBytes**: Preferred block size (2 MB)

**Performance Tuning**:
- **Lower BatchTimeout**: Faster block creation, higher CPU usage
- **Higher MaxMessageCount**: Larger blocks, better throughput
- **Trade-off**: Throughput vs Latency vs Storage

#### Application Section
```yaml
Application: &ApplicationDefaults
    Organizations:
    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: "ANY Readers"
        Writers:
            Type: ImplicitMeta
            Rule: "ANY Writers"
        Admins:
            Type: ImplicitMeta
            Rule: "MAJORITY Admins"
        LifecycleEndorsement:
            Type: ImplicitMeta
            Rule: "MAJORITY Endorsement"
        Endorsement:
            Type: ImplicitMeta
            Rule: "MAJORITY Endorsement"
```

**Explanation**:
- **Readers**: Who can read channel data (ANY org member)
- **Writers**: Who can write to channel (ANY org member)
- **Admins**: Who can modify channel config (MAJORITY of org admins)
- **LifecycleEndorsement**: Chaincode approval policy (MAJORITY of orgs must approve)
- **Endorsement**: Default transaction endorsement policy (MAJORITY of orgs must endorse)

**ImplicitMeta Policies**:
- **ANY**: At least one sub-policy must be satisfied
- **ALL**: All sub-policies must be satisfied
- **MAJORITY**: More than half of sub-policies must be satisfied

**Why MAJORITY for endorsement?**
- **Byzantine Fault Tolerance**: Requires consensus from multiple orgs
- **Trust Distribution**: No single org can manipulate data
- **Production Best Practice**: Industry standard for permissioned blockchains

#### Profiles Section
```yaml
Profiles:
    ThreeOrgsOrdererGenesis:
        <<: *ChannelDefaults
        Orderer:
            <<: *OrdererDefaults
            Organizations:
                - *OrdererOrg
        Consortiums:
            AuthConsortium:
                Organizations:
                    - *Org1
                    - *Org2
                    - *Org3

    ThreeOrgsChannel:
        Consortium: AuthConsortium
        <<: *ChannelDefaults
        Application:
            <<: *ApplicationDefaults
            Organizations:
                - *Org1
                - *Org2
                - *Org3
```

**Explanation**:
- **ThreeOrgsOrdererGenesis**: Profile for creating genesis block
  - Defines orderer configuration
  - Creates "AuthConsortium" with Org1, Org2, Org3

- **ThreeOrgsChannel**: Profile for creating application channels
  - References "AuthConsortium"
  - Adds application-level policies

**How to use profiles**:
```bash
# Generate genesis block
configtxgen -profile ThreeOrgsOrdererGenesis -channelID system-channel -outputBlock ./system-genesis-block/genesis.block

# Generate channel transaction
configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ./channel-artifacts/authchannel.tx -channelID authchannel

# Generate anchor peer transactions
configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID authchannel -asOrg Org1MSP
```

**Production Modifications**:
1. **Multiple Orderers**: Add more consenters to EtcdRaft (3-5 recommended)
2. **Custom Policies**: Adjust endorsement policies per use case
3. **Resource Limits**: Configure BatchSize based on network capacity
4. **TLS**: Enable mutual TLS for all communications

---

### 3. docker-compose-network.yaml

**Purpose**: Orchestrates all Fabric network containers using Docker Compose.

**What it deploys**:
- 1 Orderer container
- 6 Peer containers (2 per organization)
- 4 CA (Certificate Authority) containers (1 per org + 1 for orderer)
- 1 CLI container (for manual operations)

**Total**: 12 containers

**Technology**: Docker Compose v2 specification

**Why Docker Compose?**
- **Reproducibility**: Same containers on any machine with Docker
- **Simplicity**: Single command to start/stop entire network
- **Development Speed**: Faster than Kubernetes for local development
- **Resource Efficiency**: All containers share host resources

**Container Breakdown**:

#### Orderer Container
```yaml
orderer.example.com:
    container_name: orderer.example.com
    image: hyperledger/fabric-orderer:2.5.5
    environment:
        - FABRIC_LOGGING_SPEC=INFO
        - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
        - ORDERER_GENERAL_LISTENPORT=7050
        - ORDERER_GENERAL_LOCALMSPID=OrdererMSP
        - ORDERER_GENERAL_LOCALMSPDIR=/var/hyperledger/orderer/msp
        - ORDERER_GENERAL_TLS_ENABLED=true
        - ORDERER_GENERAL_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key
        - ORDERER_GENERAL_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt
        - ORDERER_GENERAL_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
        - ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=/var/hyperledger/orderer/tls/server.crt
        - ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=/var/hyperledger/orderer/tls/server.key
        - ORDERER_GENERAL_CLUSTER_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
        - ORDERER_GENERAL_BOOTSTRAPMETHOD=file
        - ORDERER_GENERAL_BOOTSTRAPFILE=/var/hyperledger/orderer/orderer.genesis.block
        - ORDERER_CHANNELPARTICIPATION_ENABLED=true
        - ORDERER_ADMIN_TLS_ENABLED=true
        - ORDERER_ADMIN_TLS_CERTIFICATE=/var/hyperledger/orderer/tls/server.crt
        - ORDERER_ADMIN_TLS_PRIVATEKEY=/var/hyperledger/orderer/tls/server.key
        - ORDERER_ADMIN_TLS_ROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
        - ORDERER_ADMIN_TLS_CLIENTROOTCAS=[/var/hyperledger/orderer/tls/ca.crt]
        - ORDERER_ADMIN_LISTENADDRESS=0.0.0.0:7053
    volumes:
        - ../system-genesis-block/genesis.block:/var/hyperledger/orderer/orderer.genesis.block
        - ../crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp:/var/hyperledger/orderer/msp
        - ../crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/:/var/hyperledger/orderer/tls
    ports:
        - 7050:7050
        - 7053:7053
```

**Key Environment Variables**:
- **FABRIC_LOGGING_SPEC**: Log level (DEBUG, INFO, WARNING, ERROR)
- **ORDERER_GENERAL_LISTENPORT**: Orderer API port (7050)
- **ORDERER_GENERAL_TLS_ENABLED**: Enables TLS encryption
- **ORDERER_GENERAL_BOOTSTRAPMETHOD**: How to get genesis block (file, none)
- **ORDERER_CHANNELPARTICIPATION_ENABLED**: Enables dynamic channel participation (Fabric 2.3+)
- **ORDERER_ADMIN_LISTENADDRESS**: Admin API port (7053, for osnadmin commands)

**Volumes**:
- `genesis.block`: Genesis block containing initial channel config
- `msp/`: Orderer's identity certificates
- `tls/`: TLS certificates for encrypted communication

**Ports**:
- `7050`: Orderer service port (peers connect here)
- `7053`: Orderer admin port (channel management)

#### Peer Container (example: peer0.org1)
```yaml
peer0.org1.example.com:
    container_name: peer0.org1.example.com
    image: hyperledger/fabric-peer:2.5.5
    environment:
        - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
        - CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=blok_chain_authh_default
        - FABRIC_LOGGING_SPEC=INFO
        - CORE_PEER_TLS_ENABLED=true
        - CORE_PEER_PROFILE_ENABLED=false
        - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt
        - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key
        - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt
        - CORE_PEER_ID=peer0.org1.example.com
        - CORE_PEER_ADDRESS=peer0.org1.example.com:7051
        - CORE_PEER_LISTENADDRESS=0.0.0.0:7051
        - CORE_PEER_CHAINCODEADDRESS=peer0.org1.example.com:7052
        - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:7052
        - CORE_PEER_GOSSIP_BOOTSTRAP=peer1.org1.example.com:8051
        - CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer0.org1.example.com:7051
        - CORE_PEER_LOCALMSPID=Org1MSP
    volumes:
        - /var/run/:/host/var/run/
        - ../crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp:/etc/hyperledger/fabric/msp
        - ../crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls:/etc/hyperledger/fabric/tls
        - peer0.org1.example.com:/var/hyperledger/production
    ports:
        - 7051:7051
```

**Key Environment Variables**:
- **CORE_VM_ENDPOINT**: Docker socket for chaincode containers
- **CORE_PEER_TLS_ENABLED**: Enables TLS
- **CORE_PEER_ID**: Unique peer identifier
- **CORE_PEER_ADDRESS**: Peer service address (external clients connect here)
- **CORE_PEER_CHAINCODEADDRESS**: Chaincode container connection address
- **CORE_PEER_GOSSIP_BOOTSTRAP**: Peer to gossip with (for ledger sync)
- **CORE_PEER_LOCALMSPID**: MSP ID of peer's organization

**Gossip Protocol**:
- **Purpose**: Peers share ledger data via gossip (peer-to-peer sync)
- **BOOTSTRAP**: Initial peer to connect to (peer1 in same org)
- **EXTERNALENDPOINT**: How other orgs find this peer

**Volumes**:
- `/var/run/`: Docker socket (to launch chaincode containers)
- `msp/`: Peer identity
- `tls/`: TLS certificates
- `peer0.org1.example.com`: Named volume for ledger data (persistent)

**Port Mapping**:
- Org1: 7051, 8051 (peer0, peer1)
- Org2: 9051, 10051 (peer0, peer1)
- Org3: 11051, 12051 (peer0, peer1)

**Why different ports?**
- All containers run on same Docker host
- Need unique ports to avoid conflicts
- Production: Each peer would have own host, all use port 7051

#### CA Container (example: ca_org1)
```yaml
ca_org1:
    image: hyperledger/fabric-ca:1.5.5
    environment:
        - FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server
        - FABRIC_CA_SERVER_CA_NAME=ca-org1
        - FABRIC_CA_SERVER_TLS_ENABLED=true
        - FABRIC_CA_SERVER_PORT=7054
    ports:
        - 7054:7054
    command: sh -c 'fabric-ca-server start -b admin:adminpw -d'
    volumes:
        - ../crypto-config/peerOrganizations/org1.example.com/ca/:/etc/hyperledger/fabric-ca-server-config
```

**Purpose**: Certificate Authority server for runtime certificate operations

**Key Environment Variables**:
- **FABRIC_CA_HOME**: CA server home directory
- **FABRIC_CA_SERVER_CA_NAME**: CA name
- **FABRIC_CA_SERVER_TLS_ENABLED**: TLS for CA API
- **FABRIC_CA_SERVER_PORT**: CA API port

**Command**: Starts CA with bootstrap admin (admin:adminpw)

**When to use CA**:
- **Enrollment**: Register new users/devices
- **Re-enrollment**: Renew expiring certificates
- **Revocation**: Revoke compromised certificates

**Note**: Our network uses cryptogen-generated certs, so CA is optional (but good for production)

#### CLI Container
```yaml
cli:
    container_name: cli
    image: hyperledger/fabric-tools:2.5.5
    tty: true
    stdin_open: true
    environment:
        - GOPATH=/opt/gopath
        - CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
        - FABRIC_LOGGING_SPEC=INFO
        - CORE_PEER_ID=cli
        - CORE_PEER_ADDRESS=peer0.org1.example.com:7051
        - CORE_PEER_LOCALMSPID=Org1MSP
        - CORE_PEER_TLS_ENABLED=true
        - CORE_PEER_TLS_CERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt
        - CORE_PEER_TLS_KEY_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.key
        - CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
        - CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric/peer
    command: /bin/bash
    volumes:
        - /var/run/:/host/var/run/
        - ../../chaincodes/:/opt/gopath/src/github.com/hyperledger/fabric/peer/chaincodes/
        - ../crypto-config:/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/
        - ../scripts:/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/
        - ../channel-artifacts:/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts
```

**Purpose**: Interactive container for manual peer commands

**Usage**:
```bash
# Execute commands in CLI container
docker exec cli peer chaincode query -C authchannel -n as -c '{"Args":["GetAllDevices"]}'

# Interactive shell
docker exec -it cli bash
```

**Why needed?**
- **Manual Testing**: Quick chaincode invocation without SDK
- **Debugging**: Inspect channel, chaincode, ledger state
- **Administration**: Create channels, install chaincodes, update configs

## Usage Guide

### 1. Generate Cryptographic Material
```bash
cd network

# Generate all certificates
cryptogen generate --config=./config/crypto-config.yaml --output=crypto-config

# Verify
ls crypto-config/peerOrganizations/
# Should see: org1.example.com, org2.example.com, org3.example.com
```

**Output**: `crypto-config/` directory with all certificates and keys

### 2. Generate Genesis Block and Channel Transaction
```bash
# Create system-genesis-block directory
mkdir -p system-genesis-block
mkdir -p channel-artifacts

# Generate genesis block for orderer
configtxgen -profile ThreeOrgsOrdererGenesis \
  -channelID system-channel \
  -outputBlock ./system-genesis-block/genesis.block \
  -configPath ./config

# Generate channel creation transaction
configtxgen -profile ThreeOrgsChannel \
  -outputCreateChannelTx ./channel-artifacts/authchannel.tx \
  -channelID authchannel \
  -configPath ./config

# Generate anchor peer transactions (for gossip)
configtxgen -profile ThreeOrgsChannel \
  -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx \
  -channelID authchannel \
  -asOrg Org1MSP \
  -configPath ./config

configtxgen -profile ThreeOrgsChannel \
  -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx \
  -channelID authchannel \
  -asOrg Org2MSP \
  -configPath ./config

configtxgen -profile ThreeOrgsChannel \
  -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors.tx \
  -channelID authchannel \
  -asOrg Org3MSP \
  -configPath ./config
```

**Output**:
- `system-genesis-block/genesis.block`: Orderer genesis block
- `channel-artifacts/authchannel.tx`: Channel creation transaction
- `channel-artifacts/Org*MSPanchors.tx`: Anchor peer updates

### 3. Start Network
```bash
# From network directory
docker-compose -f config/docker-compose-network.yaml up -d

# Verify all containers running
docker ps

# Check logs
docker logs orderer.example.com
docker logs peer0.org1.example.com
```

**Expected**: 12 containers running (1 orderer, 6 peers, 4 CAs, 1 CLI)

### 4. Modify Configuration (Advanced)

**Adding a new organization**:
1. Edit `crypto-config.yaml`: Add new PeerOrg section
2. Edit `configtx.yaml`: Add new organization definition
3. Edit `docker-compose-network.yaml`: Add peer and CA containers
4. Regenerate all configs and restart network

**Changing consensus parameters**:
1. Edit `configtx.yaml` → Orderer section
2. Modify BatchTimeout, BatchSize, etc.
3. Regenerate genesis block
4. Restart orderer with new genesis block

**Adding more peers**:
1. Edit `crypto-config.yaml`: Increase Template.Count
2. Edit `docker-compose-network.yaml`: Add peer containers with unique ports
3. Regenerate crypto material
4. Restart network

## Troubleshooting

### Common Issues

**Error**: `configtxgen: command not found`
- **Cause**: Fabric binaries not in PATH
- **Solution**:
  ```bash
  export PATH=$PATH:$PWD/../bin
  export FABRIC_CFG_PATH=$PWD/config
  ```

**Error**: `Cannot run peer because cannot init crypto`
- **Cause**: Crypto material not generated or paths wrong
- **Solution**: Check `crypto-config/` exists and paths in docker-compose match

**Error**: `Port already in use`
- **Cause**: Another Fabric network or container using same ports
- **Solution**: Stop other networks, or change ports in docker-compose

**Error**: `Error starting orderer: failed to parse config`
- **Cause**: Invalid YAML in configtx.yaml
- **Solution**: Validate YAML syntax, check for tabs vs spaces

## Learn More

- **[Network Scripts](../scripts/README.md)** - Automation scripts for network operations
- **[Main README](../../README.md)** - Quick start guide
- **[Deployment Guide](../../docs/deployment/PRODUCTION_DEPLOYMENT.md)** - Production deployment
- **[Hyperledger Fabric Docs](https://hyperledger-fabric.readthedocs.io/)** - Official documentation

## Navigation

📍 **Path**: [Main README](../../README.md) → [Network](../README.md) → **Config** → [Scripts](../scripts/README.md)

🔗 **Quick Links**:
- [← Network Overview](../README.md)
- [Network Scripts →](../scripts/README.md)
- [HOW IT WORKS](../../HOW_IT_WORKS.md)

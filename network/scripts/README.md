# Network Management Scripts

📍 **Location**: `network/scripts/`
🔗 **Parent**: [Network Overview](../README.md)
📚 **Related**: [Network Config](../config/README.md) | [Makefile](../../Makefile)

## Overview

This directory contains shell scripts that automate Hyperledger Fabric network operations. These scripts handle everything from network startup/shutdown to chaincode deployment and channel management. Think of these as the "control panel" - high-level commands that orchestrate complex multi-step operations.

## Purpose

**Why do these scripts exist?**

1. **Automation**: Complex operations (20+ commands) reduced to single script call
2. **Consistency**: Same procedure every time, eliminates human error
3. **Abstraction**: Hide Fabric CLI complexity behind simple interfaces
4. **CI/CD Integration**: Scripts can be called from automated pipelines
5. **Documentation**: Scripts serve as executable documentation of network operations

## Directory Structure

```
scripts/
├── network.sh                 # Network lifecycle management
├── deploy-chaincode.sh        # Chaincode deployment automation
├── verify-channel.sh          # Channel health checks
└── README.md                 # This file
```

## Files Explained

### 1. network.sh

**Purpose**: Master script for network lifecycle - start, stop, restart, create channels

**Usage**:
```bash
./network.sh <mode> [flags]
```

**Modes**:
- `up`: Bring up Fabric network
- `down`: Tear down network and clean up
- `restart`: Down + Up
- `createChannel`: Create and join authchannel

**Flags**:
- `-ca`: Use Fabric CA instead of cryptogen
- `-c <channel>`: Channel name (default: authchannel)
- `-s <dbtype>`: State database (goleveldb|couchdb)
- `-r <retry>`: Max retry attempts (default: 5)
- `-d <delay>`: Delay between retries in seconds (default: 3)
- `-verbose`: Enable verbose output

**Examples**:
```bash
# Basic network startup
./network.sh up

# Start network and create channel
./network.sh up createChannel

# Use CouchDB instead of default LevelDB
./network.sh up createChannel -s couchdb

# Custom channel name
./network.sh up createChannel -c mychannel

# Restart network
./network.sh restart

# Shutdown network
./network.sh down
```

**Internal Functions**:

#### `generateCerts()`
**What it does**: Generates cryptographic material using cryptogen

**Steps**:
1. Checks if `cryptogen` binary exists in PATH
2. Removes old `crypto-config/` directory
3. Runs: `cryptogen generate --config=../config/crypto-config.yaml --output=../crypto-config`
4. Verifies generation succeeded

**Output**: `crypto-config/` directory with all certificates and keys

**Why needed**: All Fabric components need TLS certificates and MSP identities

---

#### `generateChannelArtifacts()`
**What it does**: Generates genesis block and channel transaction files

**Steps**:
1. Checks if `configtxgen` binary exists
2. Creates `channel-artifacts/` directory
3. Generates genesis block:
   ```bash
   configtxgen -profile ThreeOrgsOrdererGenesis \
     -channelID system-channel \
     -outputBlock ../system-genesis-block/genesis.block
   ```
4. Generates channel creation transaction:
   ```bash
   configtxgen -profile ThreeOrgsChannel \
     -outputCreateChannelTx ../channel-artifacts/authchannel.tx \
     -channelID authchannel
   ```
5. Generates anchor peer updates for each org:
   ```bash
   configtxgen -profile ThreeOrgsChannel \
     -outputAnchorPeersUpdate ../channel-artifacts/Org1MSPanchors.tx \
     -channelID authchannel \
     -asOrg Org1MSP
   ```

**Output**:
- `system-genesis-block/genesis.block`
- `channel-artifacts/authchannel.tx`
- `channel-artifacts/Org{1,2,3}MSPanchors.tx`

**Why needed**: Orderer needs genesis block, peers need channel transaction to join channel

---

#### `networkUp()`
**What it does**: Starts all Docker containers

**Steps**:
1. Calls `generateCerts()` if crypto-config doesn't exist
2. Calls `generateChannelArtifacts()` if channel artifacts don't exist
3. Sets `COMPOSE_FILE` variable:
   - Default: `../config/docker-compose-network.yaml`
   - With CouchDB: `../config/docker-compose-network.yaml -f ../config/docker-compose-couch.yaml`
4. Runs: `docker-compose -f $COMPOSE_FILE up -d`
5. Waits for containers to start (sleep 10s)
6. Verifies containers are running: `docker ps`

**Output**: 12 running containers (or more with CouchDB)

**Why separate CouchDB compose file?**
- **Modularity**: CouchDB is optional, not all deployments need it
- **Development Speed**: LevelDB faster for local dev, CouchDB for rich queries
- **Resource Efficiency**: CouchDB containers use more memory

---

#### `createChannel()`
**What it does**: Creates channel and joins all peers

**Steps**:
1. Sets environment variables for Org1 peer0:
   ```bash
   CORE_PEER_LOCALMSPID=Org1MSP
   CORE_PEER_ADDRESS=peer0.org1.example.com:7051
   CORE_PEER_MSPCONFIGPATH=/path/to/Admin@org1.example.com/msp
   ```

2. Creates channel using orderer:
   ```bash
   docker exec cli peer channel create \
     -o orderer.example.com:7050 \
     -c authchannel \
     -f ./channel-artifacts/authchannel.tx \
     --tls --cafile /path/to/orderer/tls/ca.crt
   ```

3. Joins all peers to channel (loops through all 6 peers):
   ```bash
   docker exec cli peer channel join -b authchannel.block
   ```

4. Updates anchor peers for each organization:
   ```bash
   docker exec cli peer channel update \
     -o orderer.example.com:7050 \
     -c authchannel \
     -f ./channel-artifacts/Org1MSPanchors.tx \
     --tls --cafile /path/to/orderer/tls/ca.crt
   ```

**Output**: All peers joined to authchannel, anchor peers configured

**Why anchor peers?**
- **Cross-Org Gossip**: Anchor peers are entry points for cross-organization communication
- **Service Discovery**: Clients can discover peers via anchor peers
- **Gossip Bootstrap**: External peers find anchor peers first, then discover other peers

---

#### `networkDown()`
**What it does**: Completely tears down network and removes all artifacts

**Steps**:
1. Stops and removes all containers: `docker-compose down --volumes --remove-orphans`
2. Removes Hyperledger containers: `docker rm -f $(docker ps -a | grep hyperledger)`
3. Removes chaincode images: `docker rmi -f $(docker images | grep dev-peer)`
4. Deletes directories:
   ```bash
   rm -rf ../channel-artifacts/*.block ../channel-artifacts/*.tx
   rm -rf ../crypto-config
   rm -rf ../ledgers
   rm -rf ../system-genesis-block
   ```

**Output**: Clean slate, ready for fresh network startup

**When to use**:
- Testing: Start with clean state
- Troubleshooting: Remove corrupted ledger data
- Development: Reset after major changes

**Warning**: This deletes ALL ledger data. Backup important data first!

---

#### `clearContainers()` and `removeUnwantedImages()`
**Purpose**: Cleanup helper functions

**clearContainers**: Removes all Hyperledger Fabric containers
**removeUnwantedImages**: Removes chaincode Docker images (dev-peer.*)

**Why needed**: Chaincode containers persist even after `docker-compose down`, these clean them up

---

### 2. deploy-chaincode.sh

**Purpose**: Automates chaincode deployment using new Fabric 2.x lifecycle

**Usage**:
```bash
./deploy-chaincode.sh <chaincode-name>
```

**Examples**:
```bash
# Deploy AS chaincode
./deploy-chaincode.sh as

# Deploy TGS chaincode
./deploy-chaincode.sh tgs

# Deploy ISV chaincode
./deploy-chaincode.sh isv

# Deploy all chaincodes
./deploy-chaincode.sh as && ./deploy-chaincode.sh tgs && ./deploy-chaincode.sh isv
```

**Chaincode Lifecycle Steps** (Fabric 2.x):

#### Step 1: Package Chaincode
```bash
peer lifecycle chaincode package ${CC_NAME}.tar.gz \
  --path ../chaincodes/${CC_NAME}-chaincode \
  --lang golang \
  --label ${CC_NAME}_1.0
```

**What it does**: Creates a compressed package of chaincode source code

**Output**: `as.tar.gz`, `tgs.tar.gz`, or `isv.tar.gz`

**Why package?**
- **Portability**: Same package deployed to all peers
- **Versioning**: Package includes label (e.g., as_1.0)
- **Integrity**: Package hash ensures code hasn't changed

---

#### Step 2: Install on All Peers
```bash
# Install on Org1 peer0
docker exec cli peer lifecycle chaincode install as.tar.gz

# Install on Org1 peer1
docker exec -e CORE_PEER_ADDRESS=peer1.org1.example.com:8051 \
  cli peer lifecycle chaincode install as.tar.gz

# Repeat for all 6 peers (Org1, Org2, Org3)
```

**What it does**: Copies chaincode package to each peer

**Output**: Package ID (e.g., `as_1.0:abc123...`)

**Why install on all peers?**
- **Redundancy**: If one peer fails, others can execute chaincode
- **Load Balancing**: Distribute chaincode execution across peers
- **Endorsement**: Multiple peers can endorse same transaction

---

#### Step 3: Approve for Each Organization
```bash
# Approve for Org1
docker exec cli peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --package-id as_1.0:abc123... \
  --sequence 1 \
  --tls --cafile /path/to/orderer/ca.crt

# Approve for Org2 (switch CORE_PEER_* env vars to Org2)
docker exec -e CORE_PEER_LOCALMSPID=Org2MSP \
  -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
  -e CORE_PEER_MSPCONFIGPATH=/path/to/Admin@org2.example.com/msp \
  cli peer lifecycle chaincode approveformyorg ...

# Approve for Org3
...
```

**What it does**: Each organization votes to approve chaincode definition

**Why approval needed?**
- **Decentralized Governance**: No single org controls chaincode deployment
- **Consensus**: Majority of orgs must agree on chaincode version, endorsement policy, etc.
- **Lifecycle Policy**: Defined in channel config (default: MAJORITY)

**Parameters**:
- `--name`: Chaincode name (as, tgs, isv)
- `--version`: Version number (1.0, 1.1, 2.0, etc.)
- `--package-id`: From install step
- `--sequence`: Increments with each chaincode update (1, 2, 3, ...)
- `--init-required`: Whether InitLedger must be called (false by default)

---

#### Step 4: Check Commit Readiness
```bash
docker exec cli peer lifecycle chaincode checkcommitreadiness \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --sequence 1 \
  --tls --cafile /path/to/orderer/ca.crt
```

**Output**:
```json
{
  "approvals": {
    "Org1MSP": true,
    "Org2MSP": true,
    "Org3MSP": true
  }
}
```

**What it checks**: Whether enough orgs have approved to meet lifecycle policy

**Why needed**: Prevents commit failure due to insufficient approvals

---

#### Step 5: Commit Chaincode Definition
```bash
docker exec cli peer lifecycle chaincode commit \
  -o orderer.example.com:7050 \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --sequence 1 \
  --tls --cafile /path/to/orderer/ca.crt \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /path/to/org1/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /path/to/org2/tls/ca.crt \
  --peerAddresses peer0.org3.example.com:11051 \
  --tlsRootCertFiles /path/to/org3/tls/ca.crt
```

**What it does**: Commits chaincode definition to channel (writes to ledger)

**Why `--peerAddresses` for all orgs?**
- **Endorsement Collection**: Commit transaction needs endorsements from multiple orgs
- **Lifecycle Policy**: Default requires MAJORITY of orgs to endorse commit

**Output**: Chaincode definition committed to authchannel

---

#### Step 6: Initialize Chaincode (Optional)
```bash
docker exec cli peer chaincode invoke \
  -o orderer.example.com:7050 \
  -C authchannel \
  -n as \
  -c '{"Args":["InitLedger"]}' \
  --tls --cafile /path/to/orderer/ca.crt \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /path/to/org1/tls/ca.crt
```

**What it does**: Calls InitLedger function to set up initial state

**When needed**: If chaincode needs default data (e.g., TGS creates default services)

**Note**: Our chaincodes have minimal InitLedger (just logging), but TGS creates 2 default services

---

#### Step 7: Verify Deployment
```bash
# Query committed chaincodes
docker exec cli peer lifecycle chaincode querycommitted \
  -C authchannel \
  --name as

# Test invocation
docker exec cli peer chaincode query \
  -C authchannel \
  -n as \
  -c '{"Args":["GetAllDevices"]}'
```

**What it checks**: Chaincode is committed and invokable

**Output**: Chaincode details (version, sequence, endorsement policy) and query result

---

**Script Variables**:
```bash
CHANNEL_NAME="authchannel"
CC_NAME=$1  # Chaincode name from argument
CC_VERSION="1.0"
CC_SEQUENCE="1"
CC_RUNTIME_LANGUAGE="golang"
CC_SRC_PATH="../chaincodes/${CC_NAME}-chaincode"
DELAY="3"
MAX_RETRY="5"
```

**Error Handling**:
- Retry logic for network delays (max 5 retries, 3s delay)
- Exit on critical failures (install, approve, commit)
- Verbose output for debugging (`set -x`)

**Upgrade Procedure**:
To upgrade chaincode to new version:
1. Modify chaincode source code
2. Increment `CC_VERSION` to `1.1`
3. Increment `CC_SEQUENCE` to `2`
4. Run `./deploy-chaincode.sh as` again
5. All steps repeat with new version/sequence

---

### 3. verify-channel.sh

**Purpose**: Health checks for channel and peer connectivity

**Usage**:
```bash
./verify-channel.sh [channel-name]
```

**Default**: Verifies `authchannel`

**What it checks**:

#### 1. Orderer Health
```bash
docker exec cli peer channel getinfo \
  -c authchannel \
  -o orderer.example.com:7050 \
  --tls --cafile /path/to/orderer/ca.crt
```

**Output**:
```json
{
  "height": 10,
  "currentBlockHash": "abc123...",
  "previousBlockHash": "def456..."
}
```

**What it means**:
- **height**: Number of blocks in channel (should be > 0)
- **currentBlockHash**: Hash of latest block
- **previousBlockHash**: Hash of previous block (forms blockchain)

**Health Indicator**: If height > 0 and no errors, orderer is healthy

---

#### 2. Peer Membership
```bash
docker exec cli peer channel list
```

**Output**:
```
Channels peers has joined:
authchannel
```

**What it checks**: Peer successfully joined authchannel

**Health Indicator**: authchannel appears in list

---

#### 3. Chaincode Deployment
```bash
docker exec cli peer lifecycle chaincode querycommitted \
  -C authchannel
```

**Output**:
```
Committed chaincodes on channel 'authchannel':
Name: as, Version: 1.0, Sequence: 1
Name: tgs, Version: 1.0, Sequence: 1
Name: isv, Version: 1.0, Sequence: 1
```

**What it checks**: All expected chaincodes are deployed

**Health Indicator**: as, tgs, isv all appear with version 1.0

---

#### 4. Ledger Consistency
```bash
# Query same data from multiple peers
docker exec cli peer chaincode query \
  -C authchannel \
  -n as \
  -c '{"Args":["GetAllDevices"]}'

# Switch to different peer
docker exec -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
  cli peer chaincode query \
  -C authchannel \
  -n as \
  -c '{"Args":["GetAllDevices"]}'
```

**What it checks**: All peers have same ledger state

**Health Indicator**: Both queries return identical results

---

**Script Output Example**:
```
========================================
Channel Verification: authchannel
========================================

[1/4] Checking orderer connectivity...
✓ Orderer is reachable
✓ Channel height: 15 blocks

[2/4] Checking peer membership...
✓ Peer has joined channel: authchannel

[3/4] Checking deployed chaincodes...
✓ Chaincode 'as' deployed (version 1.0)
✓ Chaincode 'tgs' deployed (version 1.0)
✓ Chaincode 'isv' deployed (version 1.0)

[4/4] Checking ledger consistency...
✓ Ledger state consistent across peers

========================================
Channel Status: HEALTHY
========================================
```

**When to run**:
- After network startup
- After chaincode deployment
- After adding new peers
- Troubleshooting connectivity issues
- Before running tests

---

## Common Workflows

### Initial Network Setup
```bash
# 1. Start network and create channel
cd network
./scripts/network.sh up createChannel

# 2. Verify network health
./scripts/verify-channel.sh

# 3. Deploy chaincodes
./scripts/deploy-chaincode.sh as
./scripts/deploy-chaincode.sh tgs
./scripts/deploy-chaincode.sh isv

# 4. Verify chaincodes deployed
./scripts/verify-channel.sh
```

### Complete Reset
```bash
# 1. Stop network and clean all data
./scripts/network.sh down

# 2. Start fresh network
./scripts/network.sh up createChannel

# 3. Redeploy chaincodes
./scripts/deploy-chaincode.sh as
./scripts/deploy-chaincode.sh tgs
./scripts/deploy-chaincode.sh isv
```

### Chaincode Update
```bash
# 1. Modify chaincode source in chaincodes/as-chaincode/

# 2. Update version in deploy-chaincode.sh
# CC_VERSION="1.1"
# CC_SEQUENCE="2"

# 3. Redeploy
./scripts/deploy-chaincode.sh as

# 4. Verify new version
docker exec cli peer lifecycle chaincode querycommitted -C authchannel --name as
```

### Troubleshooting Network Issues
```bash
# 1. Check network status
./scripts/verify-channel.sh

# 2. Check container logs
docker logs peer0.org1.example.com
docker logs orderer.example.com

# 3. Restart network if needed
./scripts/network.sh restart

# 4. Verify again
./scripts/verify-channel.sh
```

## Script Customization

### Adding New Channel
Edit `network.sh`:
```bash
# Add new channel name
CHANNEL_NAME_2="mychannel"

# Duplicate createChannel() function
createMyChannel() {
  # Change channel name in all commands
  peer channel create -c mychannel -f ./channel-artifacts/mychannel.tx ...
}
```

### Adding New Organization
1. Edit `network/config/crypto-config.yaml`: Add Org4
2. Edit `network/config/configtx.yaml`: Add Org4MSP
3. Edit `network/config/docker-compose-network.yaml`: Add Org4 peers and CA
4. Edit `network.sh`: Update createChannel() to include Org4 peers

### Custom Endorsement Policy
Edit `deploy-chaincode.sh`:
```bash
# Add endorsement policy to approve and commit
peer lifecycle chaincode approveformyorg \
  ... \
  --signature-policy "OR('Org1MSP.member', 'Org2MSP.member')"

peer lifecycle chaincode commit \
  ... \
  --signature-policy "OR('Org1MSP.member', 'Org2MSP.member')"
```

**Policy Examples**:
- `OR('Org1MSP.member', 'Org2MSP.member')`: Any org can endorse
- `AND('Org1MSP.member', 'Org2MSP.member')`: Both orgs must endorse
- `OutOf(2, 'Org1MSP.member', 'Org2MSP.member', 'Org3MSP.member')`: Any 2 of 3

## Troubleshooting

### Common Issues

**Error**: `cryptogen: command not found`
- **Cause**: Fabric binaries not in PATH
- **Solution**:
  ```bash
  export PATH=$PATH:$PWD/../bin
  ./scripts/network.sh up
  ```

**Error**: `Error: failed to create deliver client`
- **Cause**: Orderer not ready when creating channel
- **Solution**: Increase sleep delay in networkUp(), or run createChannel separately after waiting

**Error**: `Error: chaincode install failed`
- **Cause**: Chaincode source has compile errors
- **Solution**: Test build locally first:
  ```bash
  cd chaincodes/as-chaincode
  go build -v .
  ```

**Error**: `Error: insufficient approvals for chaincode definition`
- **Cause**: Not all orgs approved chaincode
- **Solution**: Check approvals with:
  ```bash
  peer lifecycle chaincode checkcommitreadiness ...
  ```

**Error**: `Error: could not find chaincode with name 'as'`
- **Cause**: Chaincode not deployed or wrong name
- **Solution**: Verify deployment:
  ```bash
  peer lifecycle chaincode querycommitted -C authchannel
  ```

## Learn More

- **[Network Config](../config/README.md)** - Configuration files explained
- **[Main README](../../README.md)** - Quick start guide
- **[Makefile](../../Makefile)** - High-level commands that call these scripts
- **[Hyperledger Fabric Docs](https://hyperledger-fabric.readthedocs.io/)** - Official documentation

## Navigation

📍 **Path**: [Main README](../../README.md) → [Network](../README.md) → **Scripts**

🔗 **Quick Links**:
- [← Network Config](../config/README.md)
- [Network Overview](../README.md)
- [Deployment Guide](../../docs/deployment/PRODUCTION_DEPLOYMENT.md)
- [Troubleshooting](../../docs/troubleshooting/common-issues.md)

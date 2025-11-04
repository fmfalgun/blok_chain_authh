# Production Deployment Guide

## Pre-Deployment Checklist

### Infrastructure Requirements
- [ ] **Compute**: Minimum 3 servers (one per organization)
  - 4 CPU cores per peer
  - 8GB RAM per peer
  - 100GB SSD storage per peer
- [ ] **Network**: Low-latency connectivity between all nodes (< 100ms)
- [ ] **Security**: Firewall rules configured
- [ ] **Monitoring**: Prometheus + Grafana setup ready
- [ ] **Backup**: Automated backup solution in place
- [ ] **SSL/TLS**: Valid certificates for all endpoints

### Security Checklist
- [ ] All default passwords changed
- [ ] SSH key-based authentication enabled
- [ ] Firewall rules restrict access to required ports only
- [ ] TLS enabled for all peer-to-peer communication
- [ ] Private keys stored securely (HSM or key vault)
- [ ] Certificate expiration monitoring configured
- [ ] Audit logging enabled and centralized

### Software Prerequisites
- [ ] Docker 20.10+ installed
- [ ] Docker Compose 1.29+ installed
- [ ] Hyperledger Fabric 2.5.5 binaries
- [ ] Go 1.21+ (for chaincode compilation)
- [ ] Node.js 16+ (for SDK applications)

---

## Deployment Architecture

### Recommended Topology

```
┌─────────────────────────────────────────────────────────────┐
│                     Production Network                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Organization 1 (AS)                                         │
│  ├── peer0.org1.example.com (Leader)                        │
│  ├── peer1.org1.example.com (Follower)                      │
│  └── ca.org1.example.com                                     │
│                                                               │
│  Organization 2 (TGS)                                        │
│  ├── peer0.org2.example.com (Leader)                        │
│  ├── peer1.org2.example.com (Follower)                      │
│  └── ca.org2.example.com                                     │
│                                                               │
│  Organization 3 (ISV)                                        │
│  ├── peer0.org3.example.com (Leader)                        │
│  ├── peer1.org3.example.com (Follower)                      │
│  └── ca.org3.example.com                                     │
│                                                               │
│  Ordering Service                                            │
│  ├── orderer0.example.com                                    │
│  ├── orderer1.example.com                                    │
│  └── orderer2.example.com                                    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Step 1: Infrastructure Setup

### AWS Deployment

```bash
# Create EC2 instances
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.xlarge \
  --count 9 \
  --key-name your-key-pair \
  --security-group-ids sg-xxxxxxxxx \
  --subnet-id subnet-xxxxxxxxx \
  --block-device-mappings '[{"DeviceName":"/dev/xvda","Ebs":{"VolumeSize":100,"VolumeType":"gp3"}}]'

# Create Elastic IPs for static addressing
aws ec2 allocate-address --domain vpc

# Associate Elastic IPs with instances
aws ec2 associate-address \
  --instance-id i-xxxxxxxxx \
  --allocation-id eipalloc-xxxxxxxxx
```

### Azure Deployment

```bash
# Create resource group
az group create --name fabric-production --location eastus

# Create virtual machines
az vm create \
  --resource-group fabric-production \
  --name peer0-org1 \
  --image UbuntuLTS \
  --size Standard_D4s_v3 \
  --admin-username azureuser \
  --ssh-key-values @~/.ssh/id_rsa.pub \
  --data-disk-sizes-gb 100

# Repeat for all 9 VMs
```

### GCP Deployment

```bash
# Create instances
gcloud compute instances create peer0-org1 \
  --machine-type=n1-standard-4 \
  --image-family=ubuntu-2004-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=100GB \
  --boot-disk-type=pd-ssd \
  --zone=us-central1-a

# Repeat for all 9 instances
```

---

## Step 2: Network Configuration

### Configure Firewall Rules

```bash
# Allow peer communication
sudo ufw allow 7051/tcp  # Peer0 Org1
sudo ufw allow 8051/tcp  # Peer1 Org1
sudo ufw allow 9051/tcp  # Peer0 Org2
sudo ufw allow 10051/tcp # Peer1 Org2
sudo ufw allow 11051/tcp # Peer0 Org3
sudo ufw allow 12051/tcp # Peer1 Org3

# Allow orderer communication
sudo ufw allow 7050/tcp

# Allow CA communication
sudo ufw allow 7054/tcp  # CA Org1
sudo ufw allow 8054/tcp  # CA Org2
sudo ufw allow 9054/tcp  # CA Org3

# Allow operations/metrics
sudo ufw allow 9443:9448/tcp

# Enable firewall
sudo ufw enable
```

### DNS Configuration

```bash
# Add to /etc/hosts or configure DNS server
10.0.1.10 peer0.org1.example.com
10.0.1.11 peer1.org1.example.com
10.0.2.10 peer0.org2.example.com
10.0.2.11 peer1.org2.example.com
10.0.3.10 peer0.org3.example.com
10.0.3.11 peer1.org3.example.com
10.0.0.10 orderer.example.com
```

---

## Step 3: Install Dependencies

### On Each Server

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Fabric binaries
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.5 1.5.5
export PATH=$PATH:$(pwd)/fabric-samples/bin

# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

---

## Step 4: Generate Certificates

### Production Certificate Generation

```bash
# Generate crypto material with proper SANs
cd network/config

# Edit crypto-config.yaml to include production hostnames
nano crypto-config.yaml

# Generate certificates
cryptogen generate --config=crypto-config.yaml --output=../crypto-config

# Backup certificates securely
tar czf crypto-backup-$(date +%Y%m%d).tar.gz ../crypto-config
# Store in secure location (S3, Azure Blob, GCS)
```

### Use Production CA (Recommended)

```bash
# Initialize Fabric CA for each organization
fabric-ca-server init -b admin:adminpw

# Configure CA for production
cat > fabric-ca-server-config.yaml <<EOF
csr:
  cn: ca.org1.example.com
  names:
    - C: US
      ST: "California"
      L: "San Francisco"
      O: "Org1"
      OU: "Blockchain"
signing:
  default:
    expiry: 8760h
  profiles:
    tls:
      expiry: 8760h
EOF

# Start CA
fabric-ca-server start -b admin:adminpw
```

---

## Step 5: Deploy Network

### Start Orderer

```bash
# On orderer server
cd /opt/fabric/network

# Start orderer container
docker-compose -f docker-compose-orderer.yaml up -d

# Verify orderer is running
docker logs orderer.example.com
curl http://localhost:8443/healthz
```

### Start Peers

```bash
# On each peer server
cd /opt/fabric/network

# Org1 - Peer0
docker-compose -f docker-compose-org1-peer0.yaml up -d

# Org1 - Peer1
docker-compose -f docker-compose-org1-peer1.yaml up -d

# Repeat for other organizations

# Verify peers are running
docker ps
docker logs peer0.org1.example.com
```

### Create Channel

```bash
# On any peer (e.g., peer0.org1)
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_TLS_CERT_FILE=/opt/crypto/peer0.org1.example.com/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=/opt/crypto/peer0.org1.example.com/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=/opt/crypto/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=/opt/crypto/users/Admin@org1.example.com/msp

# Create channel
peer channel create -o orderer.example.com:7050 \
  -c authchannel \
  -f ./channel-artifacts/channel.tx \
  --outputBlock ./channel-artifacts/authchannel.block \
  --tls --cafile /opt/crypto/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Join all peers to channel
peer channel join -b ./channel-artifacts/authchannel.block
```

---

## Step 6: Deploy Chaincodes

### Package Chaincodes

```bash
# On build server
cd chaincodes/as-chaincode
peer lifecycle chaincode package as.tar.gz \
  --path . \
  --lang golang \
  --label as_1.0

# Repeat for tgs and isv chaincodes
```

### Install on All Peers

```bash
# Install on each peer
peer lifecycle chaincode install as.tar.gz

# Get package ID
peer lifecycle chaincode queryinstalled

# Approve for organization
peer lifecycle chaincode approveformyorg \
  -o orderer.example.com:7050 \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --package-id $PACKAGE_ID \
  --sequence 1 \
  --tls \
  --cafile /opt/crypto/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Commit chaincode (after all orgs approve)
peer lifecycle chaincode commit \
  -o orderer.example.com:7050 \
  --channelID authchannel \
  --name as \
  --version 1.0 \
  --sequence 1 \
  --tls \
  --cafile /opt/crypto/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
  --peerAddresses peer0.org1.example.com:7051 \
  --tlsRootCertFiles /opt/crypto/peer0.org1.example.com/tls/ca.crt \
  --peerAddresses peer0.org2.example.com:9051 \
  --tlsRootCertFiles /opt/crypto/peer0.org2.example.com/tls/ca.crt \
  --peerAddresses peer0.org3.example.com:11051 \
  --tlsRootCertFiles /opt/crypto/peer0.org3.example.com/tls/ca.crt
```

---

## Step 7: Configure Monitoring

### Deploy Prometheus

```bash
# On monitoring server
cd monitoring

# Update prometheus.yml with production endpoints
docker-compose -f docker-compose-monitoring.yml up -d

# Access Prometheus
open http://monitoring.example.com:9090

# Verify targets are up
curl http://monitoring.example.com:9090/api/v1/targets
```

### Configure Grafana

```bash
# Access Grafana
open http://monitoring.example.com:3000

# Add Prometheus datasource
# Import dashboard from monitoring/grafana/dashboards/

# Set up alerting channels (Slack, Email, PagerDuty)
```

---

## Step 8: Configure Backups

### Automated Backup Script

```bash
#!/bin/bash
# /opt/fabric/backup.sh

BACKUP_DIR="/opt/backups"
DATE=$(date +%Y%m%d-%H%M%S)

# Backup ledger data
docker exec peer0.org1.example.com tar czf - /var/hyperledger/production > \
  $BACKUP_DIR/ledger-$DATE.tar.gz

# Backup crypto material
tar czf $BACKUP_DIR/crypto-$DATE.tar.gz /opt/fabric/network/crypto-config

# Upload to S3/Azure/GCS
aws s3 cp $BACKUP_DIR/ s3://your-backup-bucket/ --recursive

# Cleanup old backups (keep last 30 days)
find $BACKUP_DIR -mtime +30 -delete
```

### Schedule Backups

```bash
# Add to crontab
0 2 * * * /opt/fabric/backup.sh >> /var/log/fabric-backup.log 2>&1
```

---

## Step 9: Testing

### Smoke Tests

```bash
# Test device registration
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["RegisterDevice","test_device","-----BEGIN PUBLIC KEY-----...","test"]}'

# Test authentication
peer chaincode invoke -C authchannel -n as \
  -c '{"Args":["Authenticate","{\"deviceID\":\"test_device\",\"nonce\":\"abc123\",\"timestamp\":1672531200,\"signature\":\"xyz789\"}"]}'

# Test service ticket issuance
peer chaincode invoke -C authchannel -n tgs \
  -c '{"Args":["IssueServiceTicket","{\"deviceID\":\"test_device\",\"tgtID\":\"tgt_123\",\"serviceID\":\"service001\",\"timestamp\":1672531200,\"signature\":\"xyz789\"}"]}'
```

### Load Testing

```bash
# Use provided performance tests
cd tests/performance
go test -bench=. -benchtime=60s
```

---

## Step 10: Go Live

### Final Checklist
- [ ] All peers are synchronized (same ledger height)
- [ ] Monitoring dashboards show green status
- [ ] Backups are running successfully
- [ ] SSL certificates are valid and not expiring soon
- [ ] Rate limiting is configured appropriately
- [ ] Documentation is updated with production endpoints
- [ ] Runbook is prepared for common issues
- [ ] On-call rotation is established
- [ ] Load testing completed successfully
- [ ] Security audit completed

### Rollback Plan
1. Keep previous version containers running on standby
2. Have database snapshots ready
3. Document rollback procedure
4. Test rollback in staging environment first

---

## Maintenance

### Regular Tasks
- **Daily**: Monitor dashboards, check logs for errors
- **Weekly**: Review backup integrity, check disk space
- **Monthly**: Review security logs, rotate credentials
- **Quarterly**: Update dependencies, patch OS, conduct security audit

### Upgrades

```bash
# Zero-downtime chaincode upgrade
peer lifecycle chaincode package as_v2.tar.gz \
  --path ./chaincodes/as-chaincode \
  --lang golang \
  --label as_2.0

# Install new version on all peers
peer lifecycle chaincode install as_v2.tar.gz

# Approve and commit with sequence incremented
peer lifecycle chaincode approveformyorg \
  --channelID authchannel \
  --name as \
  --version 2.0 \
  --package-id $NEW_PACKAGE_ID \
  --sequence 2 \
  --tls

# Commit after all orgs approve
peer lifecycle chaincode commit \
  --channelID authchannel \
  --name as \
  --version 2.0 \
  --sequence 2 \
  --tls
```

---

## Support Contacts

- **Technical Lead**: tech-lead@example.com
- **Security Team**: security@example.com
- **DevOps Team**: devops@example.com
- **On-call**: +1-555-0100

## Additional Resources

- [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)
- [Project Architecture](../architecture/authentication-flow.md)
- [API Reference](../api/chaincode-api.md)
- [Troubleshooting Guide](../troubleshooting/common-issues.md)

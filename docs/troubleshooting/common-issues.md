# Troubleshooting Guide

## Network Issues

### Issue: Network fails to start

**Symptoms**:
- Docker containers not starting
- Error: "Cannot connect to Docker daemon"

**Solutions**:
1. Check Docker is running:
   ```bash
   docker ps
   systemctl status docker
   ```

2. Clean up existing containers:
   ```bash
   make network-down
   docker rm -f $(docker ps -aq)
   docker volume prune -f
   ```

3. Restart network:
   ```bash
   make network-up
   ```

---

### Issue: Channel creation fails

**Symptoms**:
- Error: "Channel already exists"
- Error: "Orderer not reachable"

**Solutions**:
1. Verify orderer is running:
   ```bash
   docker logs orderer.example.com
   ```

2. Check network connectivity:
   ```bash
   docker exec cli peer channel list
   ```

3. Recreate channel artifacts:
   ```bash
   cd network && ./scripts/network.sh down
   rm -rf channel-artifacts crypto-config
   ./scripts/network.sh up
   ```

---

## Chaincode Issues

### Issue: Chaincode deployment fails

**Symptoms**:
- Error: "Chaincode install failed"
- Error: "Endorsement policy not satisfied"

**Solutions**:
1. Check chaincode packaging:
   ```bash
   cd chaincodes/as-chaincode
   go mod tidy
   go build
   ```

2. Verify all dependencies:
   ```bash
   go mod download
   go mod verify
   ```

3. Redeploy chaincode:
   ```bash
   cd network
   ./scripts/deploy-chaincode.sh as
   ```

---

### Issue: Chaincode invocation timeout

**Symptoms**:
- Error: "Request timeout"
- Error: "Chaincode container not responding"

**Solutions**:
1. Check chaincode container logs:
   ```bash
   docker logs $(docker ps -f name=dev-peer0.org1.* -q)
   ```

2. Increase timeout in client:
   ```javascript
   const contract = network.getContract('as');
   const result = await contract.submitTransaction('RegisterDevice', deviceID, publicKey, metadata, {
     timeout: 60000 // 60 seconds
   });
   ```

3. Verify peer connectivity:
   ```bash
   docker exec cli peer chaincode query -C authchannel -n as -c '{"Args":["GetAllDevices"]}'
   ```

---

## Authentication Issues

### Issue: Device registration fails with "Invalid public key"

**Symptoms**:
- Error: "publicKey length is invalid"
- Error: "must be a valid PEM-encoded public key"

**Solutions**:
1. Verify public key format:
   ```bash
   # Key must start with -----BEGIN PUBLIC KEY-----
   # and end with -----END PUBLIC KEY-----
   cat device_public_key.pem
   ```

2. Generate valid key pair:
   ```bash
   openssl genrsa -out private_key.pem 2048
   openssl rsa -in private_key.pem -pubout -out public_key.pem
   ```

3. Ensure proper encoding:
   ```javascript
   const publicKeyPEM = fs.readFileSync('public_key.pem', 'utf8');
   // Do NOT base64 encode the PEM - use it as-is
   ```

---

### Issue: Authentication fails with "timestamp is invalid"

**Symptoms**:
- Error: "timestamp is invalid or too old"
- Error: "timestamp is too far in the future"

**Solutions**:
1. Synchronize system time:
   ```bash
   sudo ntpdate -s time.nist.gov
   # or
   sudo timedatectl set-ntp true
   ```

2. Check time skew:
   ```javascript
   const timestamp = Math.floor(Date.now() / 1000);
   console.log('Current Unix timestamp:', timestamp);
   ```

3. Use proper timestamp format:
   ```javascript
   // Correct: Unix timestamp in seconds
   const timestamp = Math.floor(Date.now() / 1000);

   // Incorrect: Milliseconds
   const timestamp = Date.now(); // DON'T DO THIS
   ```

---

### Issue: Rate limit exceeded

**Symptoms**:
- Error: "rate limit exceeded (60/60 requests per minute)"
- Error: "device banned for 5 minutes"

**Solutions**:
1. Implement request throttling:
   ```javascript
   const delay = ms => new Promise(resolve => setTimeout(resolve, ms));

   for (const request of requests) {
     await processRequest(request);
     await delay(1000); // 1 second between requests
   }
   ```

2. Use batch operations when possible:
   ```javascript
   // Instead of individual requests, batch them
   await contract.submitTransaction('BatchOperation', JSON.stringify(devices));
   ```

3. Monitor rate limit status:
   ```javascript
   // Implement exponential backoff
   async function retryWithBackoff(fn, maxRetries = 3) {
     for (let i = 0; i < maxRetries; i++) {
       try {
         return await fn();
       } catch (error) {
         if (error.message.includes('rate limit')) {
           await delay(Math.pow(2, i) * 1000);
         } else {
           throw error;
         }
       }
     }
   }
   ```

---

## Performance Issues

### Issue: Slow transaction processing

**Symptoms**:
- High latency for chaincode invocations
- Transactions taking > 5 seconds

**Solutions**:
1. Check peer resource usage:
   ```bash
   docker stats peer0.org1.example.com
   ```

2. Optimize chaincode queries:
   ```go
   // Use indexed queries instead of GetStateByRange
   resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("device", []string{deviceID})
   ```

3. Increase peer resources:
   ```yaml
   # docker-compose.yaml
   services:
     peer0.org1.example.com:
       deploy:
         resources:
           limits:
             cpus: '2'
             memory: 4G
   ```

---

### Issue: High memory usage

**Symptoms**:
- Peers consuming > 4GB memory
- Out of memory errors

**Solutions**:
1. Monitor memory usage:
   ```bash
   docker stats --no-stream
   ```

2. Implement periodic cleanup:
   ```bash
   # Add to cron
   0 2 * * * docker system prune -f
   ```

3. Optimize data structures:
   ```go
   // Use pagination for large result sets
   func (s *ISVChaincode) GetAccessLogsPaginated(ctx contractapi.TransactionContextInterface,
       deviceID string, pageSize int32, bookmark string) (*PaginatedQueryResult, error) {
       // Implementation with pagination
   }
   ```

---

## Monitoring Issues

### Issue: Prometheus not scraping metrics

**Symptoms**:
- Empty Grafana dashboards
- Error: "No data" in Prometheus

**Solutions**:
1. Verify metrics endpoints:
   ```bash
   curl http://peer0.org1.example.com:9443/metrics
   curl http://orderer.example.com:8443/metrics
   ```

2. Check Prometheus configuration:
   ```bash
   docker logs prometheus
   cat monitoring/prometheus/prometheus.yml
   ```

3. Restart monitoring stack:
   ```bash
   make monitoring-down
   make monitoring-up
   ```

---

## Security Issues

### Issue: TLS handshake failures

**Symptoms**:
- Error: "TLS handshake failed"
- Error: "Certificate verification failed"

**Solutions**:
1. Regenerate certificates:
   ```bash
   cd network
   rm -rf crypto-config
   cryptogen generate --config=config/crypto-config.yaml --output=crypto-config
   ```

2. Verify certificate paths:
   ```yaml
   # Check docker-compose.yaml
   volumes:
     - ./crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls:/etc/hyperledger/fabric/tls
   ```

3. Check certificate expiration:
   ```bash
   openssl x509 -in crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/server.crt -noout -dates
   ```

---

## Logging and Debugging

### Enable Debug Logging

```bash
# For peers
export FABRIC_LOGGING_SPEC=DEBUG

# For orderer
export ORDERER_GENERAL_LOGLEVEL=DEBUG

# For chaincode
export CORE_CHAINCODE_LOGGING_LEVEL=DEBUG
```

### View Logs

```bash
# Peer logs
docker logs -f peer0.org1.example.com

# Orderer logs
docker logs -f orderer.example.com

# Chaincode logs
docker logs -f $(docker ps -f name=dev-peer0.org1.* -q)

# All containers
docker-compose -f network/config/docker-compose-network.yaml logs -f
```

### Useful Debug Commands

```bash
# Check channel info
docker exec cli peer channel getinfo -c authchannel

# List installed chaincodes
docker exec cli peer lifecycle chaincode queryinstalled

# List committed chaincodes
docker exec cli peer lifecycle chaincode querycommitted -C authchannel

# Check peer status
docker exec cli peer node status
```

---

## Getting Help

1. **Check logs first**: Most issues can be diagnosed from container logs
2. **Review documentation**: docs/architecture/authentication-flow.md
3. **Test network health**: `cd network && ./scripts/verify-channel.sh`
4. **Clean slate**: If all else fails, `make clean && make network-up`
5. **Community support**: Open an issue on GitHub with logs and error messages

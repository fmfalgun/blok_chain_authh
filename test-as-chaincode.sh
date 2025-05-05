#!/bin/bash

# Test script for AS Chaincode functions

# Set environment variables
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export ORG1_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export ORG2_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export ORG3_TLS_ROOTCERT=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

# Test client public key
CLIENT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvSYtNtHdGJPPGRoSspF8
MXfnNUKHMr1OGht1YmpJ/a1dtj9o8oLQrFzpEhTls9EK31+TNAQ1Qev2HmwU8V35
pUWxlVXW4W9lKMctLfnPEhdlSWF+8mP+4UXcwhQDdZjiHgHM1v4SqR+dI1UBgHIq
eVrO34ScnRcwQNXM8qNi6tOvpfCB9aQT+WLZvN9zLsQgv5JZQXEVz1XIQzXjJV2x
WbsoI/f7thbeYTVHdkH2wjy06K5ijPy1vWQ+wbjdJZdxn5fEyu3OiUMLnd+ZuGLA
Q8I7h1jMPQ9JHnUl3whuyEY5bxUuXdBHQKU5PP+zQTpPHUxxqCLSAu0chIoEJAB9
vwIDAQAB
-----END PUBLIC KEY-----"

echo "===== Testing AS Chaincode Functions ====="

# 1. Test RegisterClient
echo "Testing RegisterClient..."
docker exec cli peer chaincode invoke -C channel1 -n as_1.0 -c "{\"function\":\"RegisterClient\",\"Args\":[\"client1\", \"$CLIENT_PUBLIC_KEY\"]}" \
    --tls --cafile $ORDERER_CA \
    --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
    --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
    --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT

# 2. Test CheckClientValidity
echo "Testing CheckClientValidity..."
docker exec cli peer chaincode query -C channel1 -n as_1.0 -c "{\"function\":\"CheckClientValidity\",\"Args\":[\"client1\"]}" \
    --tls --cafile $ORDERER_CA

# 3. Test InitiateAuthentication
echo "Testing InitiateAuthentication..."
AUTH_RESPONSE=$(docker exec cli peer chaincode invoke -C channel1 -n as_1.0 -c "{\"function\":\"InitiateAuthentication\",\"Args\":[\"client1\"]}" \
    --tls --cafile $ORDERER_CA \
    --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
    --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
    --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT)
echo "Authentication Response: $AUTH_RESPONSE"

# 4. Test VerifyClientIdentity
echo "Testing VerifyClientIdentity..."
# Note: In a real scenario, we would encrypt the nonce with the client's private key
# For testing, we'll use a simulated encrypted nonce
ENCRYPTED_NONCE="simulated_encrypted_nonce_base64"
docker exec cli peer chaincode invoke -C channel1 -n as_1.0 -c "{\"function\":\"VerifyClientIdentity\",\"Args\":[\"client1\", \"$ENCRYPTED_NONCE\"]}" \
    --tls --cafile $ORDERER_CA \
    --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
    --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
    --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT

# 5. Test VerifyClientIdentityWithSignature
echo "Testing VerifyClientIdentityWithSignature..."
# Note: In a real scenario, we would sign the nonce with the client's private key
# For testing, we'll use a simulated signature
SIGNED_NONCE="simulated_signed_nonce_base64"
docker exec cli peer chaincode invoke -C channel1 -n as_1.0 -c "{\"function\":\"VerifyClientIdentityWithSignature\",\"Args\":[\"client1\", \"$SIGNED_NONCE\"]}" \
    --tls --cafile $ORDERER_CA \
    --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
    --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
    --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT

# 6. Test GenerateTGT
echo "Testing GenerateTGT..."
TGT_RESPONSE=$(docker exec cli peer chaincode invoke -C channel1 -n as_1.0 -c "{\"function\":\"GenerateTGT\",\"Args\":[\"client1\"]}" \
    --tls --cafile $ORDERER_CA \
    --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $ORG1_TLS_ROOTCERT \
    --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $ORG2_TLS_ROOTCERT \
    --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles $ORG3_TLS_ROOTCERT)
echo "TGT Response: $TGT_RESPONSE"

# 7. Test GetAllClientRegistrations
echo "Testing GetAllClientRegistrations..."
docker exec cli peer chaincode query -C channel1 -n as_1.0 -c "{\"function\":\"GetAllClientRegistrations\",\"Args\":[]}" \
    --tls --cafile $ORDERER_CA

echo "===== AS Chaincode Testing Completed =====" 
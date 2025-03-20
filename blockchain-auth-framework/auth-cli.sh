#!/bin/bash

# Script to interact with authentication chaincodes through the CLI container

# Function to register a client
register_client() {
    CLIENT_ID=$1
    echo "Registering client $CLIENT_ID with AS chaincode..."
    
    # Create a single-line public key (escaped newlines)
    PUBKEY="-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvSYtNtHdGJPPGRoSspF8\\nMXfnNUKHMr1OGht1YmpJ/a1dtj9o8oLQrFzpEhTls9EK31+TNAQ1Qev2HmwU8V35\\npUWxlVXW4W9lKMctLfnPEhdlSWF+8mP+4UXcwhQDdZjiHgHM1v4SqR+dI1UBgHIq\\neVrO34ScnRcwQNXM8qNi6tOvpfCB9aQT+WLZvN9zLsQgv5JZQXEVz1XIQzXjJV2x\\nWbsoI/f7thbeYTVHdkH2wjy06K5ijPy1vWQ+wbjdJZdxn5fEyu3OiUMLnd+ZuGLA\\nQ8I7h1jMPQ9JHnUl3whuyEY5bxUuXdBHQKU5PP+zQTpPHUxxqCLSAu0chIoEJAB9\\nvwIDAQAB\\n-----END PUBLIC KEY-----"
    
    # Run the command in the CLI container
    docker exec cli peer chaincode invoke -C chaichis-channel -n as-chaincode -c "{\"function\":\"RegisterClient\",\"Args\":[\"$CLIENT_ID\", \"$PUBKEY\"]}" --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
}

# Function to register an IoT device
register_device() {
    DEVICE_ID=$1
    CAPABILITIES=$2
    echo "Registering device $DEVICE_ID with capabilities: $CAPABILITIES"
    
    # Create a single-line public key (escaped newlines)
    PUBKEY="-----BEGIN PUBLIC KEY-----\\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzg1xKfJUq6mD7GME5VWh\\n5pZZZQmJKZqPgpcaiKO5ob1gAPpQ2vG5OPRxRQmrfXQBUfJkhCHWrUWWTJ/lWB5p\\ndFx9KmrWr3NzqITulEy3CW4qy6FXtQL1iSfYxJGXTH5rLYpy1bjBkKMPzTnVlBxx\\nzgP0ZBzUcRcEHGBG4Uj/abjAntHSjXvBfr83Mt2sQtJGZlvRssNwUP6EFV0udt2u\\nORP/0kh5MpGMCBVQWmVbO2zZ9WOIOiA2FDDNvlFQNNxgRvF5pW9wRtOTwvKo1jGR\\nycjkS2+KfIyFNFKTY0XnAKpY3QrOqWqT1ucvWwNGG6Z77JEt6d5Y3PsZkpOE46ky\\nAQIDAQAB\\n-----END PUBLIC KEY-----"
    
    # Format capabilities as JSON array
    CAPABILITIES_JSON="[\"${CAPABILITIES//,/\",\"}\"]"
    
    # Run the command in the CLI container with proper endorsement
    docker exec cli peer chaincode invoke -C chaichis-channel -n isv-chaincode -c "{\"function\":\"RegisterIoTDevice\",\"Args\":[\"$DEVICE_ID\", \"$PUBKEY\", $CAPABILITIES_JSON]}" --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
}

# Function to check client validity
check_client() {
    CLIENT_ID=$1
    echo "Checking validity of client $CLIENT_ID..."
    
    # Query the AS chaincode
    docker exec cli peer chaincode query -C chaichis-channel -n as-chaincode -c "{\"function\":\"CheckClientValidity\",\"Args\":[\"$CLIENT_ID\"]}" --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
}

# Function to simulate full authentication flow
authenticate() {
    CLIENT_ID=$1
    DEVICE_ID=$2
    echo "Simulating authentication flow for client $CLIENT_ID to access device $DEVICE_ID"
    
    # Step 1: Get TGT from AS
    echo "Step 1: Getting TGT from Authentication Server..."
    TGT_RESPONSE=$(docker exec cli peer chaincode invoke -C chaichis-channel -n as-chaincode -c "{\"function\":\"GenerateTGT\",\"Args\":[\"$CLIENT_ID\"]}" --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt)
    echo "TGT Response: $TGT_RESPONSE"
    
    # Step 2: Get Service Ticket from TGS (simulated)
    echo "Step 2: Getting Service Ticket from TGS (simulated)..."
    echo "Service Ticket would be obtained from TGS chaincode"
    
    # Step 3: Access IoT device through ISV (simulated)
    echo "Step 3: Authenticating with ISV and accessing device (simulated)..."
    echo "Service ticket would be validated and device access would be granted"
    
    echo "Authentication flow simulation completed"
}

# Function to get IoT device data
get_device_data() {
    DEVICE_ID=$1
    echo "Getting data for device $DEVICE_ID..."
    
    # Query the ISV chaincode for all devices
    RESPONSE=$(docker exec cli peer chaincode query -C chaichis-channel -n isv-chaincode -c "{\"function\":\"GetAllIoTDevices\",\"Args\":[]}" --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem)
    
    echo "Devices data: $RESPONSE"
}

# Main function
case "$1" in
    register-client)
        register_client $2
        ;;
    register-device)
        register_device $2 "$3"
        ;;
    check-client)
        check_client $2
        ;;
    authenticate)
        authenticate $2 $3
        ;;
    get-device-data)
        get_device_data $2
        ;;
    *)
        echo "Usage:"
        echo "  ./auth-cli.sh register-client <clientId>"
        echo "  ./auth-cli.sh register-device <deviceId> \"capability1,capability2\""
        echo "  ./auth-cli.sh check-client <clientId>"
        echo "  ./auth-cli.sh authenticate <clientId> <deviceId>"
        echo "  ./auth-cli.sh get-device-data <deviceId>"
        ;;
esac

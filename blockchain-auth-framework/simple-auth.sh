#!/bin/bash

# This script uses test-ccs.sh to perform authentication actions

# Check if cli container is running
if ! docker ps | grep -q cli; then
    echo "CLI container is not running. Please start the Fabric network."
    exit 1
fi

# Function to register a client
register_client() {
    CLIENT_ID=$1
    
    # Generate a proper escaped public key
    PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvSYtNtHdGJPPGRoSspF8
MXfnNUKHMr1OGht1YmpJ/a1dtj9o8oLQrFzpEhTls9EK31+TNAQ1Qev2HmwU8V35
pUWxlVXW4W9lKMctLfnPEhdlSWF+8mP+4UXcwhQDdZjiHgHM1v4SqR+dI1UBgHIq
eVrO34ScnRcwQNXM8qNi6tOvpfCB9aQT+WLZvN9zLsQgv5JZQXEVz1XIQzXjJV2x
WbsoI/f7thbeYTVHdkH2wjy06K5ijPy1vWQ+wbjdJZdxn5fEyu3OiUMLnd+ZuGLA
Q8I7h1jMPQ9JHnUl3whuyEY5bxUuXdBHQKU5PP+zQTpPHUxxqCLSAu0chIoEJAB9
vwIDAQAB
-----END PUBLIC KEY-----"
    
    # Save the public key to a file and copy to CLI container
    echo "$PUBLIC_KEY" > temp_pubkey.pem
    docker cp temp_pubkey.pem cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/
    rm temp_pubkey.pem
    
    # Execute registration
    echo "Registering client $CLIENT_ID..."
    docker exec -it cli /bin/bash -c "peer chaincode invoke -C chaichis-channel -n as-chaincode -c '{\"function\":\"RegisterClient\",\"Args\":[\"$CLIENT_ID\", \"'$(cat << EOF
$(docker exec cli cat /opt/gopath/src/github.com/hyperledger/fabric/peer/temp_pubkey.pem)
EOF
)'\"]}'  --tls --cafile \$ORDERER_CA"
}

# Function to register a device
register_device() {
    DEVICE_ID=$1
    CAPS=$2
    
    # Generate a proper device public key
    PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdef
abcdef1234567890ABCDEF1234567890abcdefABCDEF
vwIDAQAB
-----END PUBLIC KEY-----"
    
    # Save the public key to a file and copy to CLI container
    echo "$PUBLIC_KEY" > temp_device_pubkey.pem
    docker cp temp_device_pubkey.pem cli:/opt/gopath/src/github.com/hyperledger/fabric/peer/
    rm temp_device_pubkey.pem
    
    # Execute registration
    echo "Registering device $DEVICE_ID with capabilities: $CAPS"
    docker exec -it cli /bin/bash -c "peer chaincode invoke -C chaichis-channel -n isv-chaincode -c '{\"function\":\"RegisterIoTDevice\",\"Args\":[\"$DEVICE_ID\", \"'$(cat << EOF
$(docker exec cli cat /opt/gopath/src/github.com/hyperledger/fabric/peer/temp_device_pubkey.pem)
EOF
)'\", \"$CAPS\"]}'  --tls --cafile \$ORDERER_CA"
}

# Function to check client validity
check_client() {
    CLIENT_ID=$1
    echo "Checking validity of client $CLIENT_ID..."
    docker exec cli peer chaincode query -C chaichis-channel -n as-chaincode -c "{\"function\":\"CheckClientValidity\",\"Args\":[\"$CLIENT_ID\"]}" --tls --cafile $ORDERER_CA
}

# Function to get TGT for a client
get_tgt() {
    CLIENT_ID=$1
    echo "Getting TGT for client $CLIENT_ID..."
    docker exec cli peer chaincode invoke -C chaichis-channel -n as-chaincode -c "{\"function\":\"GenerateTGT\",\"Args\":[\"$CLIENT_ID\"]}" --tls --cafile $ORDERER_CA
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
    get-tgt)
        get_tgt $2
        ;;
    *)
        echo "Usage:"
        echo "  ./auth-cli.sh register-client <clientId>"
        echo "  ./auth-cli.sh register-device <deviceId> \"capability1,capability2\""
        echo "  ./auth-cli.sh check-client <clientId>"
        echo "  ./auth-cli.sh get-tgt <clientId>"
        ;;
esac

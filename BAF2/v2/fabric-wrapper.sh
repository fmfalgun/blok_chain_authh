#!/bin/bash
# fabric-wrapper.sh
# A script to integrate standalone auth framework keys with the simplified Fabric client

# Configuration
KEYS_DIR="keys"
USER_NAME="admin"  # Default Fabric user
STANDALONE_FRAMEWORK="standalone-auth-framework.go"  # The standalone key management
FABRIC_CLIENT="simple-fabric-client.go"  # The simplified Fabric client

function show_usage {
    echo "Usage: ./fabric-wrapper.sh COMMAND [OPTIONS]"
    echo "Commands:"
    echo "  register-client CLIENT_ID              - Generate keys and register client with Fabric"
    echo "  register-device DEVICE_ID CAPABILITIES - Generate keys and register device with Fabric"
    echo "  authenticate CLIENT_ID DEVICE_ID       - Authenticate using the generated keys"
    echo "  get-device-data CLIENT_ID DEVICE_ID    - Get device data after authentication"
    echo "  close-session CLIENT_ID DEVICE_ID      - Close an active session"
    echo "  debug-rsa NONCE                        - Debug RSA operations with a specific nonce"
    echo ""
    echo "Examples:"
    echo "  ./fabric-wrapper.sh register-client client1"
    echo "  ./fabric-wrapper.sh register-device device1 temperature humidity"
    echo "  ./fabric-wrapper.sh authenticate client1 device1"
}

# Check arguments
if [ $# -lt 1 ]; then
    show_usage
    exit 1
fi

COMMAND=$1
shift

case $COMMAND in
    register-client)
        if [ $# -lt 1 ]; then
            echo "Error: Missing client ID"
            show_usage
            exit 1
        fi
        CLIENT_ID=$1
        
        # Step 1: Generate keys using standalone framework
        echo "Generating keys for client $CLIENT_ID..."
        go run $STANDALONE_FRAMEWORK generate-keys $CLIENT_ID
        
        # Step 2: Register client with Fabric using simplified client
        echo "Registering client with Fabric (using $USER_NAME)..."
        go run $FABRIC_CLIENT register-client $USER_NAME $CLIENT_ID
        if [ $? -ne 0 ]; then
            echo "ERROR: Failed to register client with Fabric. See error message above."
            exit 1
        fi
        
        echo "SUCCESS: Client $CLIENT_ID registered with Authentication Server"
        ;;
        
    register-device)
        if [ $# -lt 2 ]; then
            echo "Error: Missing device ID or capabilities"
            show_usage
            exit 1
        fi
        DEVICE_ID=$1
        shift
        CAPABILITIES=$@
        
        # Step 1: Generate keys using standalone framework
        echo "Generating keys for device $DEVICE_ID..."
        go run $STANDALONE_FRAMEWORK generate-keys $DEVICE_ID
        
        # Step 2: Register device with Fabric using simplified client
        echo "Registering device with Fabric (using $USER_NAME)..."
        go run $FABRIC_CLIENT register-device $USER_NAME $DEVICE_ID $CAPABILITIES
        if [ $? -ne 0 ]; then
            echo "ERROR: Failed to register device with Fabric. See error message above."
            exit 1
        fi
        
        echo "SUCCESS: IoT device $DEVICE_ID registered with capabilities: $CAPABILITIES"
        ;;
        
    authenticate)
        if [ $# -lt 2 ]; then
            echo "Error: Missing client ID or device ID"
            show_usage
            exit 1
        fi
        CLIENT_ID=$1
        DEVICE_ID=$2
        
        # Step 1: Verify keys exist
        if [ ! -f "$KEYS_DIR/$CLIENT_ID-private.pem" ]; then
            echo "Error: Client key not found. Please register client first."
            exit 1
        fi
        if [ ! -f "$KEYS_DIR/$DEVICE_ID-private.pem" ]; then
            echo "Error: Device key not found. Please register device first."
            exit 1
        fi
        
        # Step 2: Perform authentication using simplified client
        echo "Authenticating with Fabric (using $USER_NAME)..."
        go run $FABRIC_CLIENT authenticate $USER_NAME $CLIENT_ID $DEVICE_ID
        if [ $? -ne 0 ]; then
            echo "ERROR: Failed to authenticate with Fabric. See error message above."
            exit 1
        fi
        
        echo "SUCCESS: Authentication successful! You can now access the IoT device."
        ;;
        
    get-device-data)
        if [ $# -lt 2 ]; then
            echo "Error: Missing client ID or device ID"
            show_usage
            exit 1
        fi
        CLIENT_ID=$1
        DEVICE_ID=$2
        
        # Get device data using simplified client
        echo "Getting device data from Fabric (using $USER_NAME)..."
        go run $FABRIC_CLIENT get-device-data $USER_NAME $CLIENT_ID $DEVICE_ID
        if [ $? -ne 0 ]; then
            echo "ERROR: Failed to get device data from Fabric. See error message above."
            exit 1
        fi
        ;;
        
    close-session)
        if [ $# -lt 2 ]; then
            echo "Error: Missing client ID or device ID"
            show_usage
            exit 1
        fi
        CLIENT_ID=$1
        DEVICE_ID=$2
        
        # Close session using simplified client
        echo "Closing session in Fabric (using $USER_NAME)..."
        go run $FABRIC_CLIENT close-session $USER_NAME $CLIENT_ID $DEVICE_ID
        if [ $? -ne 0 ]; then
            echo "ERROR: Failed to close session in Fabric. See error message above."
            exit 1
        fi
        ;;
        
    debug-rsa)
        if [ $# -lt 1 ]; then
            echo "Error: Missing nonce"
            echo "Usage: ./fabric-wrapper.sh debug-rsa <nonce>"
            exit 1
        fi
        NONCE=$1
        
        # Run RSA debugging with the standalone framework
        echo "Running RSA debugging with nonce: $NONCE"
        go run $STANDALONE_FRAMEWORK debug-rsa $NONCE
        ;;
        
    *)
        echo "Error: Unknown command '$COMMAND'"
        show_usage
        exit 1
        ;;
esac

exit 0

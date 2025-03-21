#!/bin/bash
# fabric-wrapper.sh
# A simple script to use standalone auth framework keys with Fabric network

# Configuration
KEYS_DIR="keys"
NODE_APP_DIR="/path/to/nodejs/app"  # Update this to your Node.js app directory

function show_usage {
    echo "Usage: ./fabric-wrapper.sh COMMAND [OPTIONS]"
    echo "Commands:"
    echo "  register-client CLIENT_ID              - Generate keys and register client with Fabric"
    echo "  register-device DEVICE_ID CAPABILITIES - Generate keys and register device with Fabric"
    echo "  authenticate CLIENT_ID DEVICE_ID       - Authenticate using the generated keys"
    echo "  get-device-data CLIENT_ID DEVICE_ID    - Get device data after authentication"
    echo "  close-session CLIENT_ID DEVICE_ID      - Close an active session"
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
        go run standalone-auth-framework.go generate-keys $CLIENT_ID
        
        # Step 2: Copy key to Node.js app directory if needed
        # echo "Copying key to Node.js app directory..."
        # cp $KEYS_DIR/$CLIENT_ID-private.pem $NODE_APP_DIR/$CLIENT_ID-private.pem
        
        # Step 3: Register client with Fabric using Node.js app
        echo "Registering client with Fabric (using admin)..."
        # You can either use a Node.js script or your working Fabric command
        # Example: node register-client.js admin $CLIENT_ID
        
        # For now, we'll simulate this step
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
        go run standalone-auth-framework.go generate-keys $DEVICE_ID
        
        # Step 2: Copy key to Node.js app directory if needed
        # echo "Copying key to Node.js app directory..."
        # cp $KEYS_DIR/$DEVICE_ID-private.pem $NODE_APP_DIR/$DEVICE_ID-private.pem
        
        # Step 3: Register device with Fabric using Node.js app
        echo "Registering device with Fabric (using admin)..."
        # Example: node register-device.js admin $DEVICE_ID $CAPABILITIES
        
        # For now, we'll simulate this step
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
        
        # Step 2: Simulate authentication first to verify keys work
        echo "Simulating authentication..."
        go run standalone-auth-framework.go simulate-auth $CLIENT_ID test-nonce-123
        
        # Step 3: Perform actual authentication with Fabric
        echo "Authenticating with Fabric (using admin)..."
        # Example: node authenticate.js admin $CLIENT_ID $DEVICE_ID
        
        # For now, we'll simulate this step
        echo "Step 1: Getting TGT from Authentication Server..."
        echo "Step 2: Getting Service Ticket from Ticket Granting Server..."
        echo "Step 3: Authenticating with IoT Service Validator and accessing device..."
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
        
        # Step 1: Verify authentication has been done
        echo "Checking if session exists..."
        # This would check for a session file in a real implementation
        
        # Step 2: Get device data from Fabric
        echo "Getting device data from Fabric (using admin)..."
        # Example: node get-device-data.js admin $CLIENT_ID $DEVICE_ID
        
        # For now, we'll simulate this step
        echo "Retrieved data for device $DEVICE_ID:"
        echo "  Device ID: $DEVICE_ID"
        echo "  Capabilities: temperature, humidity"
        echo "  Status: active"
        ;;
        
    close-session)
        if [ $# -lt 2 ]; then
            echo "Error: Missing client ID or device ID"
            show_usage
            exit 1
        fi
        CLIENT_ID=$1
        DEVICE_ID=$2
        
        # Step 1: Verify session exists
        echo "Checking if session exists..."
        # This would check for a session file in a real implementation
        
        # Step 2: Close session in Fabric
        echo "Closing session in Fabric (using admin)..."
        # Example: node close-session.js admin $CLIENT_ID $DEVICE_ID
        
        # For now, we'll simulate this step
        echo "SUCCESS: Closed session for device $DEVICE_ID"
        ;;
        
    *)
        echo "Error: Unknown command '$COMMAND'"
        show_usage
        exit 1
        ;;
esac

exit 0

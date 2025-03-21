#!/bin/bash
# init-wallet.sh - Script to initialize the wallet with identity credentials

# Exit on error
set -e

YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Initializing wallet with identity credentials...${NC}"

# Create wallet directory if it doesn't exist
mkdir -p wallet

# Run the wallet initialization utility
go run cmd/authcli/main.go --log-level=debug help

# Check if admin identity exists in the wallet
ADMIN_EXISTS=false
if [ -d "wallet/admin" ]; then
    ADMIN_EXISTS=true
fi

if [ "$ADMIN_EXISTS" = false ]; then
    echo -e "${YELLOW}Admin identity not found in wallet. Attempting to find credentials...${NC}"
    
    # Check common locations for Fabric credentials
    CERTS_DIR="certs"
    
    # If certs directory doesn't exist, create it
    if [ ! -d "$CERTS_DIR" ]; then
        mkdir -p $CERTS_DIR
        echo -e "${YELLOW}Created $CERTS_DIR directory for credentials${NC}"
    fi
    
    # Search for certificates in parent directories if available
    for DIR in "../v1/certs" "../v2/certs" "../certs"; do
        if [ -d "$DIR" ]; then
            echo -e "${YELLOW}Found certificates in $DIR, copying to $CERTS_DIR...${NC}"
            cp -r $DIR/* $CERTS_DIR/ 2>/dev/null || true
        fi
    done
    
    # Check for wallet in parent directories
    for DIR in "../v1/wallet" "../v2/wallet" "../wallet"; do
        if [ -d "$DIR" ]; then
            echo -e "${YELLOW}Found wallet in $DIR, copying to wallet/...${NC}"
            cp -r $DIR/* wallet/ 2>/dev/null || true
        fi
    done
    
    # Check if admin identity now exists
    if [ -d "wallet/admin" ]; then
        ADMIN_EXISTS=true
        echo -e "${GREEN}Successfully imported admin identity from previous versions${NC}"
    fi
fi

if [ "$ADMIN_EXISTS" = false ]; then
    echo -e "${YELLOW}Need to manually set up admin identity${NC}"
    echo -e "Please provide the following information to create an admin identity:"
    
    # Run the utility to create admin identity
    #go run internal/fabric/wallet.go
    go run cmd/authcli/main.go --log-level=debug initialize-wallet

    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to initialize admin identity${NC}"
        echo -e "Please ensure you have valid certificates for your Fabric network"
        exit 1
    fi
fi

# Verify wallet initialization
if [ -d "wallet/admin" ]; then
    echo -e "${GREEN}Wallet initialized successfully with admin identity${NC}"
    echo -e "${YELLOW}You can now build and run the authentication framework:${NC}"
    echo -e "  ${BLUE}make build${NC}      - Build the authcli binary"
    echo -e "  ${BLUE}make auth-flow${NC}  - Run complete authentication flow"
else
    echo -e "${RED}Failed to initialize wallet. Please ensure you have valid certificates for your Fabric network${NC}"
    exit 1
fi

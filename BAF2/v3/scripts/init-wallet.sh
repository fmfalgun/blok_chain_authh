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

# Check if admin identity exists in the wallet
ADMIN_EXISTS=false
if [ -d "wallet/admin" ]; then
    ADMIN_EXISTS=true
    echo -e "${GREEN}Admin identity already exists in wallet${NC}"
    exit 0
fi

# Search for wallet in parent directories
for DIR in "../v1/wallet" "../v2/wallet" "../wallet"; do
    if [ -d "$DIR/admin" ]; then
        echo -e "${YELLOW}Found wallet in $DIR, copying to wallet/...${NC}"
        mkdir -p wallet
        cp -r $DIR/admin wallet/ 2>/dev/null || true
    fi
done

# Check if admin identity now exists
if [ -d "wallet/admin" ]; then
    ADMIN_EXISTS=true
    echo -e "${GREEN}Successfully imported admin identity from previous versions${NC}"
    exit 0
fi

# If admin identity still doesn't exist, run the wallet initialization tool
echo -e "${YELLOW}Need to manually set up admin identity${NC}"
go run cmd/initwallet/main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to initialize admin identity${NC}"
    echo -e "Please ensure you have valid certificates for your Fabric network"
    exit 1
fi

# Verify wallet initialization
if [ -d "wallet/admin" ] || [ "$(ls -A wallet/)" ]; then
    echo -e "${GREEN}Wallet initialized successfully with admin identity${NC}"
    echo -e "${YELLOW}You can now build and run the authentication framework:${NC}"
    echo -e "  ${BLUE}make build${NC}      - Build the authcli binary"
    echo -e "  ${BLUE}make auth-flow${NC}  - Run complete authentication flow"
else
    echo -e "${RED}Failed to initialize wallet. Please ensure you have valid certificates for your Fabric network${NC}"
    exit 1
fi

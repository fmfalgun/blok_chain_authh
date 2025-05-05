#!/bin/bash

# Quick Setup Script for Blockchain Authentication Framework
# This script makes all test scripts executable and sets up the test environment

# Define color codes for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function for printing colored output
print_green() {
    echo -e "${GREEN}$1${NC}"
}

print_yellow() {
    echo -e "${YELLOW}$1${NC}"
}

print_red() {
    echo -e "${RED}$1${NC}"
}

print_yellow "Quick Setup for Blockchain Authentication Framework"
echo "======================================================"

# Make all scripts executable
print_yellow "1. Making all scripts executable..."
chmod +x test-authentication-flow.sh
chmod +x test-rsa-keys.sh
chmod +x setup-test-environment.sh
chmod +x run-all-tests.sh
chmod +x auth-cli.sh
chmod +x simple-auth.sh
chmod +x make-executable.sh
chmod +x check-network-status.sh

print_green "✓ All scripts are now executable"

# Check network status
print_yellow "2. Checking network status..."
./check-network-status.sh

# Prepare test environment
print_yellow "3. Setting up test environment..."
./setup-test-environment.sh

if [ $? -eq 0 ]; then
    print_green "✓ Test environment setup complete"
else
    print_red "✗ Test environment setup failed"
    exit 1
fi

# Display available commands
print_yellow "4. Available commands:"
echo "--------------------------"
echo "- Run all tests: ./run-all-tests.sh"
echo "- Run authentication flow test: ./test-authentication-flow.sh"
echo "- Run RSA key test: ./test-rsa-keys.sh"
echo "- Check network status: ./check-network-status.sh"
echo "- Register client (CLI): ./auth-cli.sh register-client <clientId>"
echo "- Register device (CLI): ./auth-cli.sh register-device <deviceId> \"capability1,capability2\""
echo "- Authenticate (CLI): ./auth-cli.sh authenticate <clientId> <deviceId>"
echo "--------------------------"

print_green "Quick setup complete! You can now run tests and use the authentication framework."

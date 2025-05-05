#!/bin/bash

# Execute these commands to make all scripts executable and start testing

# Make run-executable.sh executable
chmod +x run-executable.sh

# Run it to make all other scripts executable
./run-executable.sh

echo "All scripts are now executable."
echo "You can now run the following commands:"
echo "  ./quick-setup.sh - Set up the test environment"
echo "  ./test-authentication-flow.sh - Test the authentication flow"
echo "  ./test-rsa-keys.sh - Test RSA key operations"
echo "  ./run-all-tests.sh - Run all tests and generate a report"
echo "  ./check-network-status.sh - Check the status of the Fabric network"

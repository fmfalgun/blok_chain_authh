#!/bin/bash

# Exit on first error
set -e

echo "===== Starting Full IoT Authentication Test Workflow ====="

echo "1. Running Org1 (Authentication Server) tests..."
./test-org1-as.sh

echo "2. Running Org2 (Ticket Granting Server) tests..."
./test-org2-tgs.sh

echo "3. Running Org3 (IoT Service Validator) tests..."
./test-org3-isv.sh

echo "===== Full Test Workflow Completed Successfully ====="
echo "The IoT authentication system is functioning correctly with all organizations."

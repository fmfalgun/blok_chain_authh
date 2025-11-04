#!/bin/bash

CHANNEL_NAME="authchannel"

echo "=========================================="
echo "Verifying Channel: $CHANNEL_NAME"
echo "=========================================="

echo "1. Checking channel info..."
docker exec cli peer channel getinfo -c $CHANNEL_NAME

echo ""
echo "2. Listing joined channels for peer0.org1..."
docker exec cli peer channel list

echo ""
echo "3. Verifying peer0.org1 can query the ledger..."
docker exec cli peer chaincode query -C $CHANNEL_NAME -n as -c '{"Args":["GetAllDevices"]}'

echo ""
echo "=========================================="
echo "Channel verification complete!"
echo "=========================================="

#!/bin/bash

# Cleanup IoT Demo - Stop and remove all containers

echo "🛑 Stopping IoT Demo..."

cd "$(dirname "$0")/../simulator"

# Stop all containers
if [ -f "docker-compose-demo.yml" ]; then
    docker-compose -f docker-compose-demo.yml down -v
    echo "✅ All demo containers stopped and removed"
else
    echo "⚠️  docker-compose-demo.yml not found"
fi

# Clean up generated config files
rm -f iot-device/config-*.json
echo "✅ Cleaned up generated config files"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║               🎉 Demo Cleanup Complete!                    ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Note: Base blockchain network (Module 1) is still running"
echo "To stop it: cd /home/user/blok_chain_authh && make network-down"
echo ""

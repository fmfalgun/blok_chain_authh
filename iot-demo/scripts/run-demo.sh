#!/bin/bash

# IoT Demo Launcher
# Usage: ./run-demo.sh [number_of_sensors]
# Default: 3 sensors

set -e

# Configuration
SENSOR_COUNT=${1:-3}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Validate sensor count
if [ "$SENSOR_COUNT" -lt 1 ] || [ "$SENSOR_COUNT" -gt 10 ]; then
    echo "❌ Error: Sensor count must be between 1 and 10"
    echo "Usage: ./run-demo.sh [number_of_sensors]"
    exit 1
fi

echo "╔════════════════════════════════════════════════════════════╗"
echo "║     IoT Blockchain Demo - Temperature Monitoring System    ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "🎯 Configuration:"
echo "   Sensors: $SENSOR_COUNT"
echo "   Web Frontend: http://localhost:3000"
echo "   API Backend: http://localhost:8080"
echo ""

# Check if Module 1 (base blockchain) is running
echo "🔍 Checking if base blockchain network (Module 1) is running..."
if ! docker ps | grep -q "orderer.example.com"; then
    echo "❌ Error: Base blockchain network is not running!"
    echo ""
    echo "Please start Module 1 first:"
    echo "  cd /home/user/blok_chain_authh"
    echo "  make network-up"
    echo "  make channel-create"
    echo "  make deploy-cc"
    echo ""
    exit 1
fi
echo "✅ Base blockchain network is running"
echo ""

# Step 1: Deploy demo chaincodes
echo "📦 Step 1/5: Deploying USER-ACL and IOT-DATA chaincodes..."
bash "$SCRIPT_DIR/deploy-demo-chaincodes.sh"
echo "✅ Chaincodes deployed"
echo ""

# Step 2: Setup demo users
echo "👥 Step 2/5: Setting up demo users (alice, bob, admin)..."
bash "$SCRIPT_DIR/setup-users.sh"
echo "✅ Users configured"
echo ""

# Step 3: Generate docker-compose with N sensors
echo "🔧 Step 3/5: Generating Docker Compose configuration..."
bash "$SCRIPT_DIR/generate-compose.sh" "$SENSOR_COUNT"
echo "✅ Docker Compose generated for $SENSOR_COUNT sensors"
echo ""

# Step 4: Start all services
echo "🚀 Step 4/5: Starting all services..."
cd "$ROOT_DIR/simulator"
docker-compose -f docker-compose-demo.yml up -d
echo "✅ All services started"
echo ""

# Step 5: Wait for services to be ready
echo "⏳ Step 5/5: Waiting for services to be ready..."
sleep 10

# Check service health
echo ""
echo "🔍 Checking service health..."

if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Backend API: Ready"
else
    echo "⚠️  Backend API: Not responding (may need more time)"
fi

if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Frontend UI: Ready"
else
    echo "⚠️  Frontend UI: Not responding (may need more time)"
fi

# Show running containers
echo ""
echo "📊 Running Containers:"
docker ps --filter "name=iot-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                    🎉 Demo Started!                        ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "🌐 Access Points:"
echo "   Web UI:       http://localhost:3000"
echo "   Backend API:  http://localhost:8080"
echo "   Health Check: http://localhost:8080/health"
echo ""
echo "👤 Demo Accounts:"
echo "   alice / alice123    (user role)"
echo "   bob / bob123        (user role)"
echo "   admin / admin123    (admin role)"
echo ""
echo "📊 $SENSOR_COUNT temperature sensors are now:"
echo "   - Authenticating with blockchain (AS → TGS → ISV)"
echo "   - Sending temperature data every 10-30 seconds"
echo "   - All data recorded on blockchain"
echo ""
echo "📝 View Logs:"
echo "   All services:  docker-compose -f simulator/docker-compose-demo.yml logs -f"
echo "   Backend API:   docker logs -f iot-backend"
echo "   Frontend:      docker logs -f iot-frontend"
echo "   Sensor 1:      docker logs -f iot-device-001"
echo ""
echo "🛑 Stop Demo:"
echo "   ./scripts/cleanup-demo.sh"
echo ""
echo "✨ Opening web browser in 3 seconds..."
sleep 3

# Open browser (works on Linux with xdg-open, macOS with open)
if command -v xdg-open > /dev/null; then
    xdg-open http://localhost:3000 2>/dev/null &
elif command -v open > /dev/null; then
    open http://localhost:3000 2>/dev/null &
fi

echo ""
echo "✅ Demo is ready! Visit http://localhost:3000 to get started"
echo ""

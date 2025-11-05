#!/bin/bash

# Generate Docker Compose file with N sensors
# Usage: ./generate-compose.sh [number_of_sensors]

SENSOR_COUNT=${1:-3}
OUTPUT_FILE="../simulator/docker-compose-demo.yml"

cat > "$OUTPUT_FILE" <<EOF
version: '3.8'

services:
  # Web Backend API
  web-backend:
    container_name: iot-backend
    build:
      context: ./web-backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - FRONTEND_URL=http://localhost:3000
      - CHANNEL_NAME=authchannel
      - FABRIC_IDENTITY=admin
      - MSP_ID=Org1MSP
      - CA_NAME=ca.org1.example.com
    networks:
      - iot-demo-network
    depends_on:
      - web-frontend

  # Web Frontend UI
  web-frontend:
    container_name: iot-frontend
    build:
      context: ./web-frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - VUE_APP_API_URL=http://localhost:8080/api
    networks:
      - iot-demo-network

EOF

# Generate IoT device simulators
for i in $(seq 1 $SENSOR_COUNT); do
    DEVICE_NUM=$(printf "%03d" $i)
    cat >> "$OUTPUT_FILE" <<EOF
  # IoT Device Simulator $DEVICE_NUM
  iot-device-$DEVICE_NUM:
    container_name: iot-device-simulator-$DEVICE_NUM
    build:
      context: ./iot-device
      dockerfile: Dockerfile
    environment:
      - CONFIG_PATH=/app/config.json
    volumes:
      - ./iot-device/config-$DEVICE_NUM.json:/app/config.json
    networks:
      - iot-demo-network
    restart: unless-stopped

EOF

    # Generate device-specific config
    cat > "../simulator/iot-device/config-$DEVICE_NUM.json" <<CONFIG
{
  "deviceID": "sensor-$DEVICE_NUM",
  "ownerID": "user_alice",
  "deviceName": "Temperature Sensor $DEVICE_NUM",
  "temperature": {
    "baseTemp": 22,
    "amplitude": 5,
    "noiseLevel": 0.5,
    "cycleHours": 24,
    "updateInterval": {
      "min": 10,
      "max": 30
    }
  },
  "session": {
    "duration": 300
  },
  "blockchain": {
    "channelName": "authchannel",
    "identity": "appUser",
    "mspId": "Org1MSP",
    "caName": "ca.org1.example.com",
    "enrollmentID": "admin",
    "enrollmentSecret": "adminpw",
    "connectionProfilePath": null
  }
}
CONFIG
done

# Add networks section
cat >> "$OUTPUT_FILE" <<EOF

networks:
  iot-demo-network:
    driver: bridge
EOF

echo "Generated docker-compose-demo.yml with $SENSOR_COUNT sensors"

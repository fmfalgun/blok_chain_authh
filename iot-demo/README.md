# IoT Demo Module - Real-Time Temperature Monitoring System

🎯 **Module 2: IoT Demonstration Layer**

A complete demonstration system showing blockchain-based IoT authentication and data management in action. This module builds on top of the base blockchain authentication framework (Module 1) to provide a working example with real IoT devices, users, and a web interface.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Module Dependencies](#module-dependencies)
- [Components](#components)
- [Usage Guide](#usage-guide)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Documentation](#documentation)

---

## 🎯 Overview

This module demonstrates a **real-world IoT scenario** where:

1. **IoT Devices (Temperature Sensors)**: Multiple simulated sensors send temperature readings every 10-30 seconds
2. **Users**: Register, login, and view data from sensors they own or have been granted access to
3. **Blockchain**: Records all authentication, authorization, and data submission events
4. **Web Interface**: Vue.js dashboard for real-time monitoring with access control

### What Gets Recorded on Blockchain:

✅ Device registrations
✅ User registrations
✅ Authentication events (every time device/user authenticates)
✅ Permission grants/revocations
✅ Temperature readings with timestamps
✅ Access control checks (who accessed what sensor)
✅ Session management (creation, activity, termination)

**Result**: Complete immutable audit trail of all IoT operations!

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    MODULE 1: Base Blockchain Framework              │
│  ┌──────┐  ┌──────┐  ┌──────┐                                      │
│  │  AS  │  │ TGS  │  │ ISV  │  Authentication & Authorization      │
│  └──────┘  └──────┘  └──────┘                                      │
└────────────────────────────────┬────────────────────────────────────┘
                                 │ Uses for authentication
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│              MODULE 2: IoT Demo Layer (This Module)                 │
│                                                                      │
│  Blockchain Layer:                                                  │
│  ┌──────────────────┐        ┌────────────────────┐               │
│  │  USER-ACL        │        │  IOT-DATA          │               │
│  │  Chaincode       │        │  Chaincode         │               │
│  │                  │        │                    │               │
│  │ - Register User  │        │ - Store Temp       │               │
│  │ - Login User     │        │ - Get Readings     │               │
│  │ - Grant Access   │        │ - Query Devices    │               │
│  │ - Check Perms    │        │                    │               │
│  └──────────────────┘        └────────────────────┘               │
│         ↑                              ↑                            │
│         │                              │                            │
│  ┌──────┴──────────────────────────────┴─────────────────────┐    │
│  │                    Application Layer                       │    │
│  │                                                            │    │
│  │  ┌─────────────────┐  ┌──────────────┐  ┌──────────────┐ │    │
│  │  │ IoT Devices     │  │ Web Backend  │  │ Web Frontend │ │    │
│  │  │ (1-10 sensors)  │  │ (Express API)│  │ (Vue.js UI)  │ │    │
│  │  │                 │  │              │  │              │ │    │
│  │  │ - Auth flow     │  │ - User auth  │  │ - Login page │ │    │
│  │  │ - Send temp     │  │ - Query BC   │  │ - Dashboard  │ │    │
│  │  │ - Auto register │  │ - Access ctl │  │ - Charts     │ │    │
│  │  └─────────────────┘  └──────────────┘  └──────────────┘ │    │
│  │      Docker × N         Docker × 1        Docker × 1     │    │
│  └────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘

User Access: http://localhost:3000
API Endpoint: http://localhost:8080
```

---

## ✨ Features

### 🔐 User Management
- **Self-Registration**: Users can register via web interface
- **Secure Login**: Password hashing, JWT tokens
- **Role-Based Access**: Users can only see sensors they registered or were granted access to
- **Access Control**: Grant/revoke access to specific sensors

### 🌡️ IoT Device Management
- **Self-Registration**: Devices auto-register on first run
- **Configurable Count**: Deploy 1-10 sensors (user choice)
- **Realistic Simulation**: Temperature follows sine wave (day/night cycle) + random noise
- **Random Intervals**: Send data every 10-30 seconds (configurable)
- **Session Management**: 5-minute sessions, auto re-authentication

### 📊 Web Dashboard
- **Real-Time Updates**: Auto-refresh every 5 seconds
- **Temperature Display**: Current reading + color-coded status
- **Historical Charts**: Line graphs showing temperature over time
- **Access Control**: Only shows sensors user has permission to view
- **Responsive Design**: Works on desktop, tablet, mobile

### 🔍 Blockchain Transparency
- **Immutable Audit Trail**: All events recorded on blockchain
- **Query Interface**: API to retrieve any historical data
- **Access Logs**: Who accessed which sensor and when
- **Authentication Tracking**: Every auth event with timestamp

---

## 📋 Prerequisites

### Required: Module 1 Must Be Running

This module **depends on** the base blockchain framework:

✅ Hyperledger Fabric network running (Module 1)
✅ AS, TGS, ISV chaincodes deployed (Module 1)
✅ Channel created and peers joined (Module 1)

**If you haven't set up Module 1:**
```bash
# Go to root directory
cd /home/user/blok_chain_authh

# Start base framework (Module 1)
make network-up
make channel-create
make deploy-cc

# Verify it's running
make verify
```

### Additional Requirements for This Module

- **Node.js**: 16+ (for simulators and web apps)
- **npm**: 8+ (comes with Node.js)
- **Docker**: 20.10+ (for containerization)

---

## 🚀 Quick Start

### Step 1: Navigate to Demo Module

```bash
cd /home/user/blok_chain_authh/iot-demo
```

### Step 2: Run Demo with Default Configuration (3 sensors)

```bash
./scripts/run-demo.sh
```

### Step 3: Run Demo with Custom Number of Sensors (1-10)

```bash
# Example: 5 sensors
./scripts/run-demo.sh 5

# Example: 10 sensors (maximum)
./scripts/run-demo.sh 10

# Example: 1 sensor (minimum)
./scripts/run-demo.sh 1
```

### Step 4: Access Web Interface

The script will automatically:
1. Deploy USER-ACL and IOT-DATA chaincodes
2. Start N IoT device simulators
3. Start web backend API (port 8080)
4. Start web frontend UI (port 3000)
5. Open browser to http://localhost:3000

### Step 5: Register and Login

**First Time Users:**
1. Click "Register" on login page
2. Enter username, password, email
3. Registration is recorded on blockchain
4. Login with your credentials

**Demo Accounts** (pre-configured):
```
Username: alice
Password: alice123
Access: Can see devices she registers

Username: bob
Password: bob123
Access: Can see devices he registers

Username: admin
Password: admin123
Access: Can see ALL devices (admin role)
```

### Step 6: Register IoT Devices

**Option A: Automatic (Recommended)**
- Devices auto-register on first run
- Each device gets unique ID: `sensor-001`, `sensor-002`, etc.
- Registered by the user who started the demo

**Option B: Manual Registration**
```bash
# Register device via CLI
docker exec -it iot-device-simulator-001 npm run register

# Or via API
curl -X POST http://localhost:8080/api/devices/register \
  -H "Authorization: Bearer <your-token>" \
  -d '{"deviceID": "sensor-custom-001", "name": "My Sensor"}'
```

### Step 7: Monitor in Real-Time

Dashboard shows:
- List of sensors you have access to
- Current temperature for each sensor
- Temperature chart (last 100 readings)
- Last update timestamp
- Device status (active/inactive)

---

## 🔗 Module Dependencies

### How Module 2 Uses Module 1:

```
IoT Device Simulator (Module 2)
    ↓ calls
AS Chaincode (Module 1) → Authenticate device
    ↓ returns TGT
TGS Chaincode (Module 1) → Request service ticket
    ↓ returns Service Ticket
ISV Chaincode (Module 1) → Validate access
    ↓ returns Session ID
IOT-DATA Chaincode (Module 2) → Store temperature
    ↓ verifies session via ISV
Blockchain records everything
```

**Key Point**: Module 2 chaincodes (USER-ACL, IOT-DATA) use cross-chaincode calls to Module 1 chaincodes (AS, TGS, ISV) for authentication and authorization.

---

## 📦 Components

### 1. USER-ACL Chaincode
**Location**: `chaincodes/user-acl-chaincode/`

**Purpose**: User management and access control

**Functions**:
```go
- RegisterUser(username, passwordHash, email, role)
- AuthenticateUser(username, password) → token
- RegisterDevice(deviceID, ownerID, deviceName)
- GrantAccess(ownerID, userID, deviceID)
- RevokeAccess(ownerID, userID, deviceID)
- GetUserPermissions(userID) → [deviceIDs]
- ValidateAccess(userID, deviceID) → bool
```

**Access Rules**:
- Users can register themselves
- Users can register devices (become owner)
- Owners can grant/revoke access to their devices
- Admins can see all devices
- Regular users only see devices they own or have been granted access to

[📖 Full Documentation](chaincodes/user-acl-chaincode/README.md)

---

### 2. IOT-DATA Chaincode
**Location**: `chaincodes/iot-data-chaincode/`

**Purpose**: Temperature data storage and retrieval

**Functions**:
```go
- StoreTemperature(deviceID, temperature, timestamp, sessionID)
  → Verifies session with ISV before storing

- GetDeviceReadings(deviceID, startTime, endTime)
  → Returns temperature readings for date range

- GetLatestReading(deviceID)
  → Returns most recent reading

- GetLatestReadings(limit)
  → Returns recent readings from all devices

- GetDeviceStats(deviceID)
  → Returns min, max, avg temperature
```

**Security**:
- All storage operations require valid session (checked via ISV)
- All retrieval operations check USER-ACL permissions
- Timestamps validated (must be within 5 minutes)
- Device must be registered in USER-ACL

[📖 Full Documentation](chaincodes/iot-data-chaincode/README.md)

---

### 3. IoT Device Simulator
**Location**: `simulator/iot-device/`

**Purpose**: Simulates temperature sensor sending data

**Features**:
- Configurable device ID
- Realistic temperature generation (18-30°C, sine wave + noise)
- Random intervals (10-30 seconds)
- Complete authentication flow
- Session management (5-minute sessions)
- Auto-registration on first run
- Graceful error handling and retry logic

**Configuration** (`config.json`):
```json
{
  "deviceID": "sensor-001",
  "ownerID": "alice",
  "temperature": {
    "baseTemp": 22,
    "amplitude": 5,
    "noiseLevel": 0.5,
    "updateInterval": { "min": 10, "max": 30 }
  },
  "session": {
    "duration": 300
  }
}
```

**How It Works**:
1. On startup: Register device if not exists
2. Every 10-30 seconds:
   - Check if session valid
   - If not: Authenticate → TGT → Service Ticket → Session
   - Store temperature reading
3. After 5 minutes: Terminate session and re-authenticate

[📖 Full Documentation](simulator/iot-device/README.md)

---

### 4. Web Backend API
**Location**: `simulator/web-backend/`

**Purpose**: REST API for web frontend

**Technology**: Node.js + Express + Fabric SDK

**Endpoints**:

#### Authentication
```
POST /api/auth/register
  Body: { username, password, email }
  Returns: { success, message }

POST /api/auth/login
  Body: { username, password }
  Returns: { token, user: { username, role } }

POST /api/auth/logout
  Headers: { Authorization: Bearer <token> }
  Returns: { success }
```

#### Device Management
```
GET /api/devices
  Headers: { Authorization: Bearer <token> }
  Returns: [{ deviceID, name, status, lastReading, ownerID }]

POST /api/devices/register
  Headers: { Authorization: Bearer <token> }
  Body: { deviceID, deviceName }
  Returns: { success, deviceID }

POST /api/devices/grant-access
  Headers: { Authorization: Bearer <token> }
  Body: { deviceID, targetUserID }
  Returns: { success }
```

#### Temperature Data
```
GET /api/readings/:deviceID
  Headers: { Authorization: Bearer <token> }
  Query: ?limit=100&startTime=...&endTime=...
  Returns: [{ timestamp, temperature, deviceID }]

GET /api/readings/:deviceID/latest
  Headers: { Authorization: Bearer <token> }
  Returns: { timestamp, temperature, deviceID }

GET /api/readings/:deviceID/stats
  Returns: { min, max, avg, count }
```

**Security**:
- JWT token validation middleware
- Permission checks via USER-ACL chaincode
- Rate limiting (100 req/min per user)
- CORS enabled for frontend origin only

[📖 Full Documentation](simulator/web-backend/README.md)

---

### 5. Web Frontend UI
**Location**: `simulator/web-frontend/`

**Purpose**: User interface for monitoring sensors

**Technology**: Vue.js 3 + Chart.js + Tailwind CSS

**Pages**:

#### Login/Register Page
- Username/password login
- Registration form with validation
- Error messages
- Remember me option

#### Dashboard
- **Header**: User info, logout button
- **Device List**: Cards for each accessible sensor
- **Device Card** (per sensor):
  - Device name and ID
  - Current temperature (large display)
  - Status indicator (color-coded)
  - Last update timestamp
  - Mini chart (sparkline)
- **Detailed View** (click device):
  - Full temperature chart (last 100 readings)
  - Statistics (min, max, avg)
  - Time range selector
  - Export data button

#### Device Management (for device owners)
- Register new devices
- View owned devices
- Grant access to other users
- Revoke access

**Features**:
- Auto-refresh every 5 seconds
- Real-time updates without page reload
- Responsive design (mobile-friendly)
- Temperature color coding:
  - Blue: < 20°C (cold)
  - Green: 20-25°C (normal)
  - Orange: 25-28°C (warm)
  - Red: > 28°C (hot)

[📖 Full Documentation](simulator/web-frontend/README.md)

---

## 📖 Usage Guide

### Scenario 1: Single User with Multiple Sensors

```bash
# Start demo with 5 sensors
./scripts/run-demo.sh 5

# Open browser, register as "alice"
# All 5 sensors automatically owned by alice
# Dashboard shows all 5 sensors with real-time data
```

### Scenario 2: Multiple Users with Shared Access

```bash
# Start demo
./scripts/run-demo.sh 3

# User 1 (alice) registers and logs in
# Owns sensor-001, sensor-002, sensor-003

# Alice grants access to bob for sensor-002
# Via dashboard: Device Management → sensor-002 → Grant Access → bob

# User 2 (bob) registers and logs in
# Sees only sensor-002 (granted by alice)
```

### Scenario 3: Admin Monitoring All Devices

```bash
# Start demo
./scripts/run-demo.sh 10

# Multiple users register devices
# Admin logs in with admin/admin123
# Sees ALL 10 devices regardless of ownership
```

### Scenario 4: Adding New Device After Demo Running

```bash
# Demo already running with 3 sensors
# Add 4th sensor manually:

docker-compose -f iot-demo/simulator/docker-compose-simulator.yaml up -d iot-device-004

# Device auto-registers and starts sending data
# Refresh dashboard to see new device
```

---

## ⚙️ Configuration

### Demo Configuration File

**Location**: `config/demo-config.yaml`

```yaml
# Number of IoT devices (1-10)
device_count: 3

# Device configuration
devices:
  prefix: "sensor"
  temperature:
    base: 22        # Base temperature (°C)
    amplitude: 5    # Variation range
    noise: 0.5      # Random noise level
  interval:
    min: 10         # Minimum seconds between readings
    max: 30         # Maximum seconds between readings
  session:
    duration: 300   # Session duration (5 minutes)

# Pre-configured users (optional)
users:
  - username: "alice"
    password: "alice123"
    email: "alice@example.com"
    role: "user"

  - username: "bob"
    password: "bob123"
    email: "bob@example.com"
    role: "user"

  - username: "admin"
    password: "admin123"
    email: "admin@example.com"
    role: "admin"

# Web configuration
web:
  backend_port: 8080
  frontend_port: 3000
  auto_refresh_interval: 5000  # milliseconds
  chart_data_points: 100

# Blockchain configuration (Module 1)
blockchain:
  channel: "authchannel"
  chaincodes:
    as: "as"
    tgs: "tgs"
    isv: "isv"
    user_acl: "user-acl"
    iot_data: "iot-data"
```

### Modifying Configuration

```bash
# Edit configuration
nano iot-demo/config/demo-config.yaml

# Change device count
device_count: 5  # Change from 3 to 5

# Restart demo
./scripts/run-demo.sh
```

---

## 📚 Documentation

### Quick Links

- **[Setup Guide](docs/SETUP.md)** - Detailed installation and configuration
- **[User Registration Guide](docs/USER_REGISTRATION.md)** - How to register users
- **[Device Registration Guide](docs/DEVICE_REGISTRATION.md)** - How to register devices
- **[API Reference](docs/API_REFERENCE.md)** - Complete API documentation
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

### Component Documentation

- **[USER-ACL Chaincode](chaincodes/user-acl-chaincode/README.md)**
- **[IOT-DATA Chaincode](chaincodes/iot-data-chaincode/README.md)**
- **[IoT Device Simulator](simulator/iot-device/README.md)**
- **[Web Backend](simulator/web-backend/README.md)**
- **[Web Frontend](simulator/web-frontend/README.md)**

---

## 🔧 Advanced Usage

### Customizing Temperature Generation

Edit `simulator/iot-device/temperature-generator.js`:

```javascript
// Sine wave for day/night cycle
function generateTemperature() {
  const now = Date.now();
  const hourOfDay = (now / 3600000) % 24;

  // Base follows sine wave (cooler at night, warmer during day)
  const baseTemp = 22 + 5 * Math.sin((hourOfDay - 6) * Math.PI / 12);

  // Add random noise
  const noise = (Math.random() - 0.5) * 1.0;

  return Math.round((baseTemp + noise) * 10) / 10; // Round to 1 decimal
}
```

### Adding New API Endpoints

Edit `simulator/web-backend/routes/readings.js`:

```javascript
// Example: Get average temperature for last hour
router.get('/average-hourly/:deviceID', async (req, res) => {
  const { deviceID } = req.params;
  const oneHourAgo = Date.now() - 3600000;

  // Query blockchain
  const readings = await fabricClient.query(
    'iot-data',
    'GetDeviceReadings',
    [deviceID, oneHourAgo.toString(), Date.now().toString()]
  );

  const avg = readings.reduce((sum, r) => sum + r.temperature, 0) / readings.length;
  res.json({ deviceID, average: avg, period: '1h' });
});
```

### Creating Custom Alerts

```javascript
// In iot-device simulator, add alert logic
if (temperature > 28) {
  console.warn(`🔥 HIGH TEMPERATURE ALERT: ${temperature}°C`);
  // Could send to monitoring system, webhook, etc.
}
```

---

## 🐛 Troubleshooting

### Demo Won't Start

```bash
# Check if Module 1 is running
docker ps | grep hyperledger

# If not running, start Module 1 first
cd /home/user/blok_chain_authh
make network-up
make channel-create
make deploy-cc
```

### Can't See Any Devices on Dashboard

**Cause**: Devices not registered or no access granted

**Solution**:
1. Check device logs: `docker logs iot-device-simulator-001`
2. Verify registration: `docker exec cli peer chaincode query -C authchannel -n user-acl -c '{"Args":["GetAllDevices"]}'`
3. Check user permissions: Login as admin to see all devices

### Temperature Data Not Updating

**Cause**: Device authentication failure or session expired

**Solution**:
1. Check device logs for errors
2. Verify AS/TGS/ISV chaincodes are running
3. Check blockchain network health: `make verify` (in Module 1)

### Web UI Shows "Permission Denied"

**Cause**: User doesn't have access to requested device

**Solution**:
1. Login as device owner
2. Grant access via Device Management page
3. Or login as admin to see all devices

---

## 📊 Monitoring

### View Blockchain Activity

```bash
# See all temperature readings
docker exec cli peer chaincode query \
  -C authchannel \
  -n iot-data \
  -c '{"Args":["GetLatestReadings","10"]}'

# See user permissions
docker exec cli peer chaincode query \
  -C authchannel \
  -n user-acl \
  -c '{"Args":["GetUserPermissions","alice"]}'

# See access logs (from Module 1 ISV)
docker exec cli peer chaincode query \
  -C authchannel \
  -n isv \
  -c '{"Args":["GetAccessLogs","sensor-001"]}'
```

### Prometheus Metrics (if Module 1 monitoring enabled)

```bash
# Open Grafana
open http://localhost:3001

# View metrics:
# - Number of active devices
# - Temperature readings per minute
# - Authentication events
# - Session creation/termination
```

---

## 🚀 Next Steps

1. **Explore the Code**: Read component READMEs to understand implementation
2. **Customize**: Modify temperature ranges, intervals, UI theme
3. **Extend**: Add new chaincode functions (e.g., alerts, analytics)
4. **Production**: See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for production setup

---

## 📄 License

Same as Module 1 - MIT License

---

**Module 2 built on top of Module 1** | **Vue.js Frontend** | **Real-Time Blockchain Monitoring**

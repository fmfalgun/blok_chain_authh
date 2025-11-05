#!/usr/bin/env node

/**
 * IoT Device Simulator - Temperature Sensor
 *
 * Simulates a temperature sensor that:
 * 1. Authenticates with blockchain (AS → TGS → ISV)
 * 2. Generates realistic temperature readings
 * 3. Sends data every 10-30 seconds
 * 4. Manages 5-minute sessions
 * 5. Auto-registers on first run
 */

const FabricClient = require('./fabric-client');
const TemperatureGenerator = require('./temperature-generator');
const fs = require('fs');
const path = require('path');

class IoTDeviceSimulator {
    constructor(config) {
        this.config = config;
        this.deviceID = config.deviceID;
        this.ownerID = config.ownerID;
        this.deviceName = config.deviceName || `Sensor ${config.deviceID}`;

        this.fabricClient = new FabricClient(config.blockchain);
        this.tempGenerator = new TemperatureGenerator(config.temperature);

        this.currentSession = null;
        this.sessionStartTime = null;
        this.sessionDuration = config.session.duration * 1000; // Convert to ms

        this.updateInterval = null;
        this.isRunning = false;

        console.log(`\n🌡️  IoT Device Simulator Initialized`);
        console.log(`📱 Device ID: ${this.deviceID}`);
        console.log(`👤 Owner: ${this.ownerID}`);
        console.log(`📊 Temp Range: ${config.temperature.baseTemp - config.temperature.amplitude}°C - ${config.temperature.baseTemp + config.temperature.amplitude}°C`);
        console.log(`⏱️  Update Interval: ${config.temperature.updateInterval.min}-${config.temperature.updateInterval.max}s`);
        console.log(`🔐 Session Duration: ${config.session.duration}s\n`);
    }

    /**
     * Start the device simulator
     */
    async start() {
        console.log(`🚀 Starting device simulator...`);

        try {
            // Initialize Fabric connection
            await this.fabricClient.connect();
            console.log(`✅ Connected to Hyperledger Fabric network`);

            // Check if device is registered, if not, register it
            await this.ensureDeviceRegistered();

            // Start the simulation loop
            this.isRunning = true;
            await this.simulationLoop();

        } catch (error) {
            console.error(`❌ Failed to start simulator: ${error.message}`);
            process.exit(1);
        }
    }

    /**
     * Ensure device is registered in USER-ACL chaincode
     */
    async ensureDeviceRegistered() {
        console.log(`🔍 Checking if device is registered...`);

        try {
            const device = await this.fabricClient.query(
                'user-acl',
                'GetDevice',
                [this.deviceID]
            );

            console.log(`✅ Device already registered`);
            return true;

        } catch (error) {
            // Device not found, register it
            console.log(`📝 Device not found. Registering...`);

            try {
                await this.fabricClient.invoke(
                    'user-acl',
                    'RegisterDevice',
                    [this.deviceID, this.deviceName, this.ownerID, 'temperature-sensor']
                );

                console.log(`✅ Device registered successfully`);
                return true;

            } catch (registerError) {
                console.error(`❌ Failed to register device: ${registerError.message}`);
                throw registerError;
            }
        }
    }

    /**
     * Main simulation loop
     */
    async simulationLoop() {
        console.log(`\n🔄 Starting simulation loop...\n`);

        while (this.isRunning) {
            try {
                // Check if we have a valid session
                if (!this.hasValidSession()) {
                    await this.authenticateAndCreateSession();
                }

                // Generate and send temperature reading
                await this.sendTemperatureReading();

                // Wait for random interval before next reading
                const waitTime = this.getRandomInterval();
                console.log(`⏳ Waiting ${waitTime}s until next reading...\n`);
                await this.sleep(waitTime * 1000);

            } catch (error) {
                console.error(`❌ Error in simulation loop: ${error.message}`);
                console.log(`🔄 Retrying in 10 seconds...`);
                await this.sleep(10000);
            }
        }
    }

    /**
     * Check if current session is still valid
     */
    hasValidSession() {
        if (!this.currentSession || !this.sessionStartTime) {
            return false;
        }

        const elapsed = Date.now() - this.sessionStartTime;
        const isValid = elapsed < this.sessionDuration;

        if (!isValid) {
            console.log(`⏰ Session expired (duration: ${Math.floor(elapsed / 1000)}s)`);
        }

        return isValid;
    }

    /**
     * Complete authentication flow: AS → TGS → ISV
     */
    async authenticateAndCreateSession() {
        console.log(`\n🔐 Starting authentication flow...`);

        try {
            // Step 1: Authenticate with AS (get TGT)
            console.log(`  1️⃣  Authenticating with AS...`);
            const timestamp = Math.floor(Date.now() / 1000);
            const nonce = this.generateNonce();
            const signature = this.generateSignature(this.deviceID, nonce, timestamp);

            const authRequest = {
                deviceID: this.deviceID,
                nonce: nonce,
                timestamp: timestamp,
                signature: signature
            };

            const authResponse = await this.fabricClient.invoke(
                'as',
                'Authenticate',
                [JSON.stringify(authRequest)]
            );

            const authResult = JSON.parse(authResponse);
            const tgtID = authResult.tgtID;
            const sessionKey = authResult.sessionKey;

            console.log(`  ✅ Received TGT: ${tgtID.substring(0, 20)}...`);

            // Step 2: Request service ticket from TGS
            console.log(`  2️⃣  Requesting service ticket from TGS...`);
            const ticketRequest = {
                deviceID: this.deviceID,
                tgtID: tgtID,
                serviceID: 'iot-data-service',
                timestamp: Math.floor(Date.now() / 1000),
                signature: this.generateSignature(this.deviceID, tgtID, 'iot-data-service')
            };

            const ticketResponse = await this.fabricClient.invoke(
                'tgs',
                'IssueServiceTicket',
                [JSON.stringify(ticketRequest)]
            );

            const ticket = JSON.parse(ticketResponse);
            const ticketID = ticket.ticketID;

            console.log(`  ✅ Received Service Ticket: ${ticketID.substring(0, 20)}...`);

            // Step 3: Validate access with ISV (create session)
            console.log(`  3️⃣  Creating session with ISV...`);
            const accessRequest = {
                deviceID: this.deviceID,
                serviceID: 'iot-data-service',
                ticketID: ticketID,
                action: 'write',
                timestamp: Math.floor(Date.now() / 1000),
                ipAddress: '192.168.1.100',
                userAgent: `IoT-Device-Simulator/${this.deviceID}`,
                signature: this.generateSignature(this.deviceID, ticketID)
            };

            const accessResponse = await this.fabricClient.invoke(
                'isv',
                'ValidateAccess',
                [JSON.stringify(accessRequest)]
            );

            const accessResult = JSON.parse(accessResponse);

            if (!accessResult.granted) {
                throw new Error(`Access denied: ${accessResult.message}`);
            }

            this.currentSession = {
                sessionID: accessResult.sessionID,
                ticketID: ticketID,
                expiresAt: accessResult.expiresAt
            };
            this.sessionStartTime = Date.now();

            console.log(`  ✅ Session created: ${this.currentSession.sessionID.substring(0, 20)}...`);
            console.log(`  ⏱️  Session valid for ${Math.floor((accessResult.expiresAt * 1000 - Date.now()) / 1000)}s\n`);

        } catch (error) {
            console.error(`  ❌ Authentication failed: ${error.message}`);
            throw error;
        }
    }

    /**
     * Generate and send temperature reading
     */
    async sendTemperatureReading() {
        try {
            // Generate temperature
            const temperature = this.tempGenerator.generate();
            const timestamp = Math.floor(Date.now() / 1000);

            // Send to blockchain
            console.log(`📊 Sending temperature: ${temperature.toFixed(1)}°C`);

            await this.fabricClient.invoke(
                'iot-data',
                'StoreTemperature',
                [
                    this.deviceID,
                    temperature.toString(),
                    timestamp.toString(),
                    this.currentSession.sessionID
                ]
            );

            console.log(`✅ Temperature stored on blockchain`);

            // Check for anomaly
            if (temperature > 28.0 || temperature < 18.0) {
                console.log(`⚠️  ANOMALY DETECTED: Temperature ${temperature.toFixed(1)}°C is outside normal range!`);
            }

        } catch (error) {
            console.error(`❌ Failed to send temperature: ${error.message}`);
            // Invalidate session on error (will re-authenticate next loop)
            this.currentSession = null;
            throw error;
        }
    }

    /**
     * Get random interval between readings
     */
    getRandomInterval() {
        const min = this.config.temperature.updateInterval.min;
        const max = this.config.temperature.updateInterval.max;
        return Math.floor(Math.random() * (max - min + 1)) + min;
    }

    /**
     * Generate nonce for authentication
     */
    generateNonce() {
        return Math.random().toString(36).substring(2, 15) +
               Math.random().toString(36).substring(2, 15);
    }

    /**
     * Generate signature (simplified - production should use actual crypto)
     */
    generateSignature(...args) {
        const crypto = require('crypto');
        const data = args.join('_');
        return crypto.createHash('sha256').update(data).digest('hex');
    }

    /**
     * Sleep helper
     */
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    /**
     * Stop the simulator
     */
    async stop() {
        console.log(`\n🛑 Stopping device simulator...`);
        this.isRunning = false;

        // Terminate session if active
        if (this.currentSession) {
            try {
                await this.fabricClient.invoke(
                    'isv',
                    'TerminateSession',
                    [this.currentSession.sessionID]
                );
                console.log(`✅ Session terminated`);
            } catch (error) {
                console.error(`⚠️  Failed to terminate session: ${error.message}`);
            }
        }

        await this.fabricClient.disconnect();
        console.log(`✅ Disconnected from network`);
        console.log(`👋 Device simulator stopped\n`);
    }
}

// Load configuration
function loadConfig() {
    const configPath = process.env.CONFIG_PATH || path.join(__dirname, 'config.json');

    if (!fs.existsSync(configPath)) {
        console.error(`❌ Configuration file not found: ${configPath}`);
        process.exit(1);
    }

    const configData = fs.readFileSync(configPath, 'utf8');
    return JSON.parse(configData);
}

// Main execution
if (require.main === module) {
    const config = loadConfig();
    const simulator = new IoTDeviceSimulator(config);

    // Handle graceful shutdown
    process.on('SIGINT', async () => {
        console.log(`\n\n⚠️  Received SIGINT signal`);
        await simulator.stop();
        process.exit(0);
    });

    process.on('SIGTERM', async () => {
        console.log(`\n\n⚠️  Received SIGTERM signal`);
        await simulator.stop();
        process.exit(0);
    });

    // Start simulator
    simulator.start().catch(error => {
        console.error(`\n❌ Fatal error: ${error.message}`);
        process.exit(1);
    });
}

module.exports = IoTDeviceSimulator;

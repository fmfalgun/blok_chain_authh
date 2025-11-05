#!/usr/bin/env node

/**
 * IoT Demo Backend API Server
 *
 * Provides REST API for:
 * - User authentication (login/register)
 * - Device management
 * - Temperature data retrieval
 * - Access control
 */

const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const morgan = require('morgan');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
require('dotenv').config();

const authRoutes = require('./routes/auth');
const deviceRoutes = require('./routes/devices');
const readingsRoutes = require('./routes/readings');
const FabricClient = require('./fabric-client');

const app = express();
const PORT = process.env.PORT || 8080;

// Middleware
app.use(helmet()); // Security headers
app.use(cors({
    origin: process.env.FRONTEND_URL || 'http://localhost:3000',
    credentials: true
}));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));
app.use(morgan('combined')); // Logging

// Rate limiting
const limiter = rateLimit({
    windowMs: 60 * 1000, // 1 minute
    max: 100, // 100 requests per minute
    message: 'Too many requests from this IP, please try again later.'
});
app.use('/api/', limiter);

// Initialize Fabric client
const fabricClient = new FabricClient({
    channelName: process.env.CHANNEL_NAME || 'authchannel',
    identity: process.env.FABRIC_IDENTITY || 'admin',
    mspId: process.env.MSP_ID || 'Org1MSP',
    caName: process.env.CA_NAME || 'ca.org1.example.com'
});

// Make fabricClient available to routes
app.locals.fabricClient = fabricClient;

// Health check endpoint
app.get('/health', (req, res) => {
    res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        uptime: process.uptime()
    });
});

// API routes
app.use('/api/auth', authRoutes);
app.use('/api/devices', deviceRoutes);
app.use('/api/readings', readingsRoutes);

// 404 handler
app.use((req, res) => {
    res.status(404).json({
        success: false,
        message: 'Endpoint not found'
    });
});

// Error handler
app.use((err, req, res, next) => {
    console.error('Error:', err);
    res.status(err.status || 500).json({
        success: false,
        message: err.message || 'Internal server error'
    });
});

// Start server
async function startServer() {
    try {
        // Connect to Fabric network
        console.log('🔗 Connecting to Hyperledger Fabric network...');
        await fabricClient.connect();
        console.log('✅ Connected to Fabric network\n');

        // Start Express server
        app.listen(PORT, () => {
            console.log(`\n🚀 IoT Demo Backend API Server`);
            console.log(`📡 Listening on port ${PORT}`);
            console.log(`🌐 CORS enabled for: ${process.env.FRONTEND_URL || 'http://localhost:3000'}`);
            console.log(`🔐 Fabric Channel: ${process.env.CHANNEL_NAME || 'authchannel'}`);
            console.log(`\n📚 API Documentation:`);
            console.log(`   POST   /api/auth/register`);
            console.log(`   POST   /api/auth/login`);
            console.log(`   POST   /api/auth/logout`);
            console.log(`   GET    /api/devices`);
            console.log(`   POST   /api/devices/register`);
            console.log(`   POST   /api/devices/grant-access`);
            console.log(`   POST   /api/devices/revoke-access`);
            console.log(`   GET    /api/readings/:deviceID`);
            console.log(`   GET    /api/readings/:deviceID/latest`);
            console.log(`   GET    /api/readings/:deviceID/stats`);
            console.log(`\n✨ Server ready!\n`);
        });

    } catch (error) {
        console.error('❌ Failed to start server:', error);
        process.exit(1);
    }
}

// Graceful shutdown
process.on('SIGINT', async () => {
    console.log('\n\n⚠️  Received SIGINT signal');
    console.log('🛑 Shutting down gracefully...');

    await fabricClient.disconnect();
    console.log('✅ Disconnected from Fabric network');

    process.exit(0);
});

process.on('SIGTERM', async () => {
    console.log('\n\n⚠️  Received SIGTERM signal');
    console.log('🛑 Shutting down gracefully...');

    await fabricClient.disconnect();
    console.log('✅ Disconnected from Fabric network');

    process.exit(0);
});

// Start the server
startServer();

module.exports = app;

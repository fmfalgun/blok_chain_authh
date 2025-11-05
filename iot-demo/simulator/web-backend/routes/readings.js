/**
 * Temperature Readings Routes
 *
 * Endpoints:
 * - GET /api/readings/:deviceID - Get readings for device (with time range)
 * - GET /api/readings/:deviceID/latest - Get latest reading
 * - GET /api/readings/:deviceID/stats - Get statistics
 */

const express = require('express');
const router = express.Router();
const { verifyToken } = require('./auth');

/**
 * Middleware to check if user has access to device
 */
async function checkDeviceAccess(req, res, next) {
    try {
        const fabricClient = req.app.locals.fabricClient;
        const userID = req.user.userID;
        const deviceID = req.params.deviceID;

        // Check access via USER-ACL chaincode
        const accessResponse = await fabricClient.query(
            'user-acl',
            'ValidateAccess',
            [userID, deviceID]
        );

        const accessResult = JSON.parse(accessResponse);

        if (!accessResult.hasAccess) {
            return res.status(403).json({
                success: false,
                message: 'Access denied to this device'
            });
        }

        // Store deviceID in request for route handlers
        req.deviceID = deviceID;
        next();

    } catch (error) {
        console.error('Access check error:', error);
        res.status(403).json({
            success: false,
            message: 'Access denied to this device'
        });
    }
}

/**
 * GET /api/readings/:deviceID
 * Get temperature readings for device
 */
router.get('/:deviceID', verifyToken, checkDeviceAccess, async (req, res) => {
    try {
        const deviceID = req.deviceID;
        const { limit, startTime, endTime } = req.query;

        const fabricClient = req.app.locals.fabricClient;

        // Calculate time range (default: last 24 hours)
        const now = Math.floor(Date.now() / 1000);
        const start = startTime ? parseInt(startTime) : now - 86400;
        const end = endTime ? parseInt(endTime) : now;

        // Get readings from IOT-DATA chaincode
        const response = await fabricClient.query(
            'iot-data',
            'GetDeviceReadings',
            [deviceID, start.toString(), end.toString()]
        );

        let readings = JSON.parse(response);

        // Apply limit if specified
        if (limit) {
            const limitNum = parseInt(limit);
            readings = readings.slice(-limitNum); // Get last N readings
        }

        res.json({
            success: true,
            deviceID: deviceID,
            readings: readings,
            count: readings.length,
            timeRange: {
                start: start,
                end: end
            }
        });

    } catch (error) {
        console.error('Get readings error:', error);
        res.status(500).json({
            success: false,
            message: 'Failed to retrieve readings'
        });
    }
});

/**
 * GET /api/readings/:deviceID/latest
 * Get latest temperature reading
 */
router.get('/:deviceID/latest', verifyToken, checkDeviceAccess, async (req, res) => {
    try {
        const deviceID = req.deviceID;
        const fabricClient = req.app.locals.fabricClient;

        // Get latest reading from IOT-DATA chaincode
        const response = await fabricClient.query(
            'iot-data',
            'GetLatestReading',
            [deviceID]
        );

        const reading = JSON.parse(response);

        res.json({
            success: true,
            deviceID: deviceID,
            reading: reading
        });

    } catch (error) {
        console.error('Get latest reading error:', error);

        if (error.message.includes('no readings found')) {
            return res.json({
                success: true,
                deviceID: req.deviceID,
                reading: null,
                message: 'No readings available yet'
            });
        }

        res.status(500).json({
            success: false,
            message: 'Failed to retrieve latest reading'
        });
    }
});

/**
 * GET /api/readings/:deviceID/stats
 * Get statistics for device
 */
router.get('/:deviceID/stats', verifyToken, checkDeviceAccess, async (req, res) => {
    try {
        const deviceID = req.deviceID;
        const fabricClient = req.app.locals.fabricClient;

        // Get statistics from IOT-DATA chaincode
        const response = await fabricClient.query(
            'iot-data',
            'GetDeviceStatistics',
            [deviceID]
        );

        const stats = JSON.parse(response);

        res.json({
            success: true,
            deviceID: deviceID,
            statistics: stats
        });

    } catch (error) {
        console.error('Get statistics error:', error);
        res.status(500).json({
            success: false,
            message: 'Failed to retrieve statistics'
        });
    }
});

module.exports = router;

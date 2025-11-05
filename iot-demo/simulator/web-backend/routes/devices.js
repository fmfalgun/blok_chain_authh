/**
 * Device Management Routes
 *
 * Endpoints:
 * - GET /api/devices - Get all accessible devices for user
 * - POST /api/devices/register - Register new device
 * - POST /api/devices/grant-access - Grant access to another user
 * - POST /api/devices/revoke-access - Revoke access from user
 */

const express = require('express');
const router = express.Router();
const { verifyToken } = require('./auth');

/**
 * GET /api/devices
 * Get all devices user has access to
 */
router.get('/', verifyToken, async (req, res) => {
    try {
        const fabricClient = req.app.locals.fabricClient;
        const userID = req.user.userID;

        // Get user's permissions
        const permissionsResponse = await fabricClient.query(
            'user-acl',
            'GetUserPermissions',
            [userID]
        );

        const permissions = JSON.parse(permissionsResponse);
        const deviceIDs = permissions.devices || [];

        // Get details for each device
        const devices = [];
        for (const deviceID of deviceIDs) {
            try {
                // Get device info from USER-ACL
                const deviceResponse = await fabricClient.query(
                    'user-acl',
                    'GetDevice',
                    [deviceID]
                );
                const device = JSON.parse(deviceResponse);

                // Get latest reading from IOT-DATA
                try {
                    const readingResponse = await fabricClient.query(
                        'iot-data',
                        'GetLatestReading',
                        [deviceID]
                    );
                    const latestReading = JSON.parse(readingResponse);
                    device.lastReading = latestReading;
                } catch (error) {
                    // No readings yet
                    device.lastReading = null;
                }

                devices.push(device);

            } catch (error) {
                console.error(`Error fetching device ${deviceID}:`, error.message);
            }
        }

        res.json({
            success: true,
            devices: devices,
            count: devices.length
        });

    } catch (error) {
        console.error('Get devices error:', error);
        res.status(500).json({
            success: false,
            message: 'Failed to retrieve devices'
        });
    }
});

/**
 * POST /api/devices/register
 * Register a new device
 */
router.post('/register', verifyToken, async (req, res) => {
    try {
        const { deviceID, deviceName, deviceType } = req.body;
        const ownerID = req.user.userID;

        // Validate input
        if (!deviceID || !deviceName) {
            return res.status(400).json({
                success: false,
                message: 'deviceID and deviceName are required'
            });
        }

        // Register device in USER-ACL chaincode
        const fabricClient = req.app.locals.fabricClient;
        await fabricClient.invoke(
            'user-acl',
            'RegisterDevice',
            [deviceID, deviceName, ownerID, deviceType || 'temperature-sensor']
        );

        res.json({
            success: true,
            message: 'Device registered successfully',
            device: {
                deviceID,
                deviceName,
                ownerID,
                deviceType: deviceType || 'temperature-sensor'
            }
        });

    } catch (error) {
        console.error('Register device error:', error);
        res.status(500).json({
            success: false,
            message: error.message || 'Failed to register device'
        });
    }
});

/**
 * POST /api/devices/grant-access
 * Grant access to another user
 */
router.post('/grant-access', verifyToken, async (req, res) => {
    try {
        const { deviceID, targetUsername, permissionType } = req.body;
        const ownerID = req.user.userID;

        // Validate input
        if (!deviceID || !targetUsername) {
            return res.status(400).json({
                success: false,
                message: 'deviceID and targetUsername are required'
            });
        }

        // First, get target user's ID by username
        // (In production, you'd have a GetUserByUsername function)
        // For now, we'll construct it as user_username
        const targetUserID = `user_${targetUsername}`;

        // Grant access
        const fabricClient = req.app.locals.fabricClient;
        await fabricClient.invoke(
            'user-acl',
            'GrantAccess',
            [ownerID, targetUserID, deviceID, permissionType || 'read']
        );

        res.json({
            success: true,
            message: `Access granted to ${targetUsername} for device ${deviceID}`
        });

    } catch (error) {
        console.error('Grant access error:', error);
        res.status(500).json({
            success: false,
            message: error.message || 'Failed to grant access'
        });
    }
});

/**
 * POST /api/devices/revoke-access
 * Revoke access from user
 */
router.post('/revoke-access', verifyToken, async (req, res) => {
    try {
        const { deviceID, targetUsername } = req.body;
        const ownerID = req.user.userID;

        // Validate input
        if (!deviceID || !targetUsername) {
            return res.status(400).json({
                success: false,
                message: 'deviceID and targetUsername are required'
            });
        }

        const targetUserID = `user_${targetUsername}`;

        // Revoke access
        const fabricClient = req.app.locals.fabricClient;
        await fabricClient.invoke(
            'user-acl',
            'RevokeAccess',
            [ownerID, targetUserID, deviceID]
        );

        res.json({
            success: true,
            message: `Access revoked from ${targetUsername} for device ${deviceID}`
        });

    } catch (error) {
        console.error('Revoke access error:', error);
        res.status(500).json({
            success: false,
            message: error.message || 'Failed to revoke access'
        });
    }
});

module.exports = router;

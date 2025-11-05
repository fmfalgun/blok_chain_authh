/**
 * Authentication Routes
 *
 * Endpoints:
 * - POST /api/auth/register - Register new user
 * - POST /api/auth/login - Login user
 * - POST /api/auth/logout - Logout user
 */

const express = require('express');
const router = express.Router();
const jwt = require('jsonwebtoken');

const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
const JWT_EXPIRY = '24h';

/**
 * POST /api/auth/register
 * Register a new user
 */
router.post('/register', async (req, res) => {
    try {
        const { username, password, email, role } = req.body;

        // Validate input
        if (!username || !password || !email) {
            return res.status(400).json({
                success: false,
                message: 'Username, password, and email are required'
            });
        }

        if (password.length < 6) {
            return res.status(400).json({
                success: false,
                message: 'Password must be at least 6 characters'
            });
        }

        // Call USER-ACL chaincode to register user
        const fabricClient = req.app.locals.fabricClient;
        const response = await fabricClient.invoke(
            'user-acl',
            'RegisterUser',
            [username, password, email, role || 'user']
        );

        const result = JSON.parse(response);

        if (!result.success) {
            return res.status(400).json({
                success: false,
                message: result.message
            });
        }

        // Generate JWT token
        const token = jwt.sign(
            {
                userID: result.userID,
                username: result.username,
                role: result.role
            },
            JWT_SECRET,
            { expiresIn: JWT_EXPIRY }
        );

        res.json({
            success: true,
            message: 'User registered successfully',
            token: token,
            user: {
                userID: result.userID,
                username: result.username,
                role: result.role
            }
        });

    } catch (error) {
        console.error('Register error:', error);
        res.status(500).json({
            success: false,
            message: error.message || 'Registration failed'
        });
    }
});

/**
 * POST /api/auth/login
 * Login user
 */
router.post('/login', async (req, res) => {
    try {
        const { username, password } = req.body;

        // Validate input
        if (!username || !password) {
            return res.status(400).json({
                success: false,
                message: 'Username and password are required'
            });
        }

        // Call USER-ACL chaincode to authenticate
        const fabricClient = req.app.locals.fabricClient;
        const response = await fabricClient.invoke(
            'user-acl',
            'AuthenticateUser',
            [username, password]
        );

        const result = JSON.parse(response);

        if (!result.success) {
            return res.status(401).json({
                success: false,
                message: 'Invalid username or password'
            });
        }

        // Generate JWT token
        const token = jwt.sign(
            {
                userID: result.userID,
                username: result.username,
                role: result.role
            },
            JWT_SECRET,
            { expiresIn: JWT_EXPIRY }
        );

        res.json({
            success: true,
            message: 'Login successful',
            token: token,
            user: {
                userID: result.userID,
                username: result.username,
                role: result.role
            }
        });

    } catch (error) {
        console.error('Login error:', error);
        res.status(401).json({
            success: false,
            message: 'Invalid username or password'
        });
    }
});

/**
 * POST /api/auth/logout
 * Logout user (client should delete token)
 */
router.post('/logout', (req, res) => {
    res.json({
        success: true,
        message: 'Logout successful'
    });
});

/**
 * Middleware to verify JWT token
 */
function verifyToken(req, res, next) {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

    if (!token) {
        return res.status(401).json({
            success: false,
            message: 'Access token required'
        });
    }

    try {
        const decoded = jwt.verify(token, JWT_SECRET);
        req.user = decoded;
        next();
    } catch (error) {
        return res.status(403).json({
            success: false,
            message: 'Invalid or expired token'
        });
    }
}

module.exports = router;
module.exports.verifyToken = verifyToken;

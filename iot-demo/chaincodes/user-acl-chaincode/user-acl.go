package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// UserACLChaincode provides user and access control management
type UserACLChaincode struct {
	contractapi.Contract
}

// User represents a registered user
type User struct {
	UserID       string   `json:"userID"`
	Username     string   `json:"username"`
	PasswordHash string   `json:"passwordHash"`
	Email        string   `json:"email"`
	Role         string   `json:"role"` // "user", "admin", "operator"
	CreatedAt    int64    `json:"createdAt"`
	LastLogin    int64    `json:"lastLogin"`
	OwnedDevices []string `json:"ownedDevices"` // DeviceIDs owned by this user
	Status       string   `json:"status"`       // "active", "suspended", "deleted"
}

// Device represents an IoT device
type Device struct {
	DeviceID    string `json:"deviceID"`
	DeviceName  string `json:"deviceName"`
	OwnerID     string `json:"ownerID"` // UserID of owner
	DeviceType  string `json:"deviceType"`
	RegisteredAt int64  `json:"registeredAt"`
	LastActive  int64   `json:"lastActive"`
	Status      string `json:"status"` // "active", "inactive", "decommissioned"
}

// AccessPermission represents a user's permission to access a device
type AccessPermission struct {
	PermissionID string `json:"permissionID"`
	UserID       string `json:"userID"`
	DeviceID     string `json:"deviceID"`
	GrantedBy    string `json:"grantedBy"`    // UserID who granted access
	GrantedAt    int64  `json:"grantedAt"`
	ExpiresAt    int64  `json:"expiresAt"`    // 0 means never expires
	PermissionType string `json:"permissionType"` // "read", "write", "admin"
	Status       string `json:"status"`        // "active", "revoked"
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success  bool   `json:"success"`
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"` // Simplified - in production use JWT
	Message  string `json:"message"`
}

// InitLedger initializes the chaincode
func (s *UserACLChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("Initializing USER-ACL Chaincode")

	// Create default admin user
	adminPasswordHash := hashPassword("admin123")
	admin := User{
		UserID:       "user_admin",
		Username:     "admin",
		PasswordHash: adminPasswordHash,
		Email:        "admin@example.com",
		Role:         "admin",
		CreatedAt:    getCurrentTimestamp(),
		LastLogin:    0,
		OwnedDevices: []string{},
		Status:       "active",
	}

	adminJSON, err := json.Marshal(admin)
	if err != nil {
		return fmt.Errorf("failed to marshal admin user: %v", err)
	}

	err = ctx.GetStub().PutState("USER_admin", adminJSON)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// Create index for username lookup
	usernameIndexKey, err := ctx.GetStub().CreateCompositeKey("USERNAME_INDEX", []string{admin.Username, admin.UserID})
	if err != nil {
		return fmt.Errorf("failed to create username index: %v", err)
	}
	err = ctx.GetStub().PutState(usernameIndexKey, []byte{0x00})
	if err != nil {
		return fmt.Errorf("failed to store username index: %v", err)
	}

	log.Println("USER-ACL Chaincode initialized with admin user")
	return nil
}

// RegisterUser registers a new user
func (s *UserACLChaincode) RegisterUser(ctx contractapi.TransactionContextInterface, username string, password string, email string, role string) (string, error) {
	// Validate inputs
	if len(username) < 3 || len(username) > 32 {
		return "", fmt.Errorf("username must be between 3 and 32 characters")
	}
	if len(password) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters")
	}
	if role != "user" && role != "admin" && role != "operator" {
		role = "user" // Default to user role
	}

	// Check if username already exists
	existingUserID, err := s.getUserIDByUsername(ctx, username)
	if err == nil && existingUserID != "" {
		return "", fmt.Errorf("username '%s' already exists", username)
	}

	// Generate unique user ID
	userID := fmt.Sprintf("user_%s_%d", username, getCurrentTimestamp())

	// Hash password
	passwordHash := hashPassword(password)

	// Create user
	user := User{
		UserID:       userID,
		Username:     username,
		PasswordHash: passwordHash,
		Email:        email,
		Role:         role,
		CreatedAt:    getCurrentTimestamp(),
		LastLogin:    0,
		OwnedDevices: []string{},
		Status:       "active",
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %v", err)
	}

	// Store user
	err = ctx.GetStub().PutState("USER_"+userID, userJSON)
	if err != nil {
		return "", fmt.Errorf("failed to store user: %v", err)
	}

	// Create username index for quick lookup
	usernameIndexKey, err := ctx.GetStub().CreateCompositeKey("USERNAME_INDEX", []string{username, userID})
	if err != nil {
		return "", fmt.Errorf("failed to create username index: %v", err)
	}
	err = ctx.GetStub().PutState(usernameIndexKey, []byte{0x00})
	if err != nil {
		return "", fmt.Errorf("failed to store username index: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("UserRegistered", []byte(userID))
	if err != nil {
		return "", fmt.Errorf("failed to emit event: %v", err)
	}

	log.Printf("User registered: %s (ID: %s, Role: %s)", username, userID, role)

	// Return auth response
	response := AuthResponse{
		Success:  true,
		UserID:   userID,
		Username: username,
		Role:     role,
		Token:    generateToken(userID),
		Message:  "User registered successfully",
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON), nil
}

// AuthenticateUser authenticates a user
func (s *UserACLChaincode) AuthenticateUser(ctx contractapi.TransactionContextInterface, username string, password string) (string, error) {
	// Find user by username
	userID, err := s.getUserIDByUsername(ctx, username)
	if err != nil || userID == "" {
		return "", fmt.Errorf("invalid username or password")
	}

	// Get user
	userJSON, err := ctx.GetStub().GetState("USER_" + userID)
	if err != nil {
		return "", fmt.Errorf("failed to read user: %v", err)
	}
	if userJSON == nil {
		return "", fmt.Errorf("invalid username or password")
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal user: %v", err)
	}

	// Check status
	if user.Status != "active" {
		return "", fmt.Errorf("user account is %s", user.Status)
	}

	// Verify password
	passwordHash := hashPassword(password)
	if passwordHash != user.PasswordHash {
		return "", fmt.Errorf("invalid username or password")
	}

	// Update last login
	user.LastLogin = getCurrentTimestamp()
	userJSON, _ = json.Marshal(user)
	ctx.GetStub().PutState("USER_"+userID, userJSON)

	// Emit event
	ctx.GetStub().SetEvent("UserLoggedIn", []byte(userID))

	log.Printf("User authenticated: %s (ID: %s)", username, userID)

	// Return auth response
	response := AuthResponse{
		Success:  true,
		UserID:   userID,
		Username: username,
		Role:     user.Role,
		Token:    generateToken(userID),
		Message:  "Authentication successful",
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON), nil
}

// RegisterDevice registers a new IoT device
func (s *UserACLChaincode) RegisterDevice(ctx contractapi.TransactionContextInterface, deviceID string, deviceName string, ownerID string, deviceType string) error {
	// Validate inputs
	if len(deviceID) < 3 || len(deviceID) > 64 {
		return fmt.Errorf("deviceID must be between 3 and 64 characters")
	}

	// Check if device already exists
	existing, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil {
		return fmt.Errorf("failed to read device: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("device %s already exists", deviceID)
	}

	// Verify owner exists
	ownerJSON, err := ctx.GetStub().GetState("USER_" + ownerID)
	if err != nil || ownerJSON == nil {
		return fmt.Errorf("owner %s not found", ownerID)
	}

	// Create device
	device := Device{
		DeviceID:     deviceID,
		DeviceName:   deviceName,
		OwnerID:      ownerID,
		DeviceType:   deviceType,
		RegisteredAt: getCurrentTimestamp(),
		LastActive:   getCurrentTimestamp(),
		Status:       "active",
	}

	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}

	// Store device
	err = ctx.GetStub().PutState("DEVICE_"+deviceID, deviceJSON)
	if err != nil {
		return fmt.Errorf("failed to store device: %v", err)
	}

	// Update owner's device list
	var owner User
	json.Unmarshal(ownerJSON, &owner)
	owner.OwnedDevices = append(owner.OwnedDevices, deviceID)
	ownerJSON, _ = json.Marshal(owner)
	ctx.GetStub().PutState("USER_"+ownerID, ownerJSON)

	// Emit event
	ctx.GetStub().SetEvent("DeviceRegistered", []byte(deviceID))

	log.Printf("Device registered: %s by owner %s", deviceID, ownerID)
	return nil
}

// GrantAccess grants a user access to a device
func (s *UserACLChaincode) GrantAccess(ctx contractapi.TransactionContextInterface, ownerID string, targetUserID string, deviceID string, permissionType string) error {
	// Verify device exists
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil || deviceJSON == nil {
		return fmt.Errorf("device %s not found", deviceID)
	}

	var device Device
	json.Unmarshal(deviceJSON, &device)

	// Verify caller is owner or admin
	if device.OwnerID != ownerID {
		// Check if caller is admin
		callerJSON, _ := ctx.GetStub().GetState("USER_" + ownerID)
		if callerJSON == nil {
			return fmt.Errorf("unauthorized: not device owner")
		}
		var caller User
		json.Unmarshal(callerJSON, &caller)
		if caller.Role != "admin" {
			return fmt.Errorf("unauthorized: not device owner or admin")
		}
	}

	// Verify target user exists
	targetUserJSON, err := ctx.GetStub().GetState("USER_" + targetUserID)
	if err != nil || targetUserJSON == nil {
		return fmt.Errorf("target user %s not found", targetUserID)
	}

	// Validate permission type
	if permissionType != "read" && permissionType != "write" && permissionType != "admin" {
		permissionType = "read" // Default to read
	}

	// Check if permission already exists
	permissionID := fmt.Sprintf("PERM_%s_%s", targetUserID, deviceID)
	existingPerm, _ := ctx.GetStub().GetState(permissionID)
	if existingPerm != nil {
		return fmt.Errorf("permission already exists for user %s on device %s", targetUserID, deviceID)
	}

	// Create permission
	permission := AccessPermission{
		PermissionID:   permissionID,
		UserID:         targetUserID,
		DeviceID:       deviceID,
		GrantedBy:      ownerID,
		GrantedAt:      getCurrentTimestamp(),
		ExpiresAt:      0, // Never expires
		PermissionType: permissionType,
		Status:         "active",
	}

	permJSON, err := json.Marshal(permission)
	if err != nil {
		return fmt.Errorf("failed to marshal permission: %v", err)
	}

	// Store permission
	err = ctx.GetStub().PutState(permissionID, permJSON)
	if err != nil {
		return fmt.Errorf("failed to store permission: %v", err)
	}

	// Emit event
	ctx.GetStub().SetEvent("AccessGranted", []byte(permissionID))

	log.Printf("Access granted: user %s can access device %s (%s)", targetUserID, deviceID, permissionType)
	return nil
}

// RevokeAccess revokes a user's access to a device
func (s *UserACLChaincode) RevokeAccess(ctx contractapi.TransactionContextInterface, ownerID string, targetUserID string, deviceID string) error {
	// Verify device exists and caller is owner/admin
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil || deviceJSON == nil {
		return fmt.Errorf("device %s not found", deviceID)
	}

	var device Device
	json.Unmarshal(deviceJSON, &device)

	if device.OwnerID != ownerID {
		// Check if caller is admin
		callerJSON, _ := ctx.GetStub().GetState("USER_" + ownerID)
		var caller User
		json.Unmarshal(callerJSON, &caller)
		if caller.Role != "admin" {
			return fmt.Errorf("unauthorized: not device owner or admin")
		}
	}

	// Get permission
	permissionID := fmt.Sprintf("PERM_%s_%s", targetUserID, deviceID)
	permJSON, err := ctx.GetStub().GetState(permissionID)
	if err != nil || permJSON == nil {
		return fmt.Errorf("permission not found")
	}

	var permission AccessPermission
	json.Unmarshal(permJSON, &permission)

	// Revoke permission
	permission.Status = "revoked"
	permJSON, _ = json.Marshal(permission)
	ctx.GetStub().PutState(permissionID, permJSON)

	// Emit event
	ctx.GetStub().SetEvent("AccessRevoked", []byte(permissionID))

	log.Printf("Access revoked: user %s can no longer access device %s", targetUserID, deviceID)
	return nil
}

// ValidateAccess checks if a user has access to a device
func (s *UserACLChaincode) ValidateAccess(ctx contractapi.TransactionContextInterface, userID string, deviceID string) (string, error) {
	// Get user
	userJSON, err := ctx.GetStub().GetState("USER_" + userID)
	if err != nil || userJSON == nil {
		return "", fmt.Errorf("user not found")
	}

	var user User
	json.Unmarshal(userJSON, &user)

	// Admins can access all devices
	if user.Role == "admin" {
		result := map[string]interface{}{
			"hasAccess":      true,
			"permissionType": "admin",
			"reason":         "Admin role",
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	// Check if user owns device
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil || deviceJSON == nil {
		return "", fmt.Errorf("device not found")
	}

	var device Device
	json.Unmarshal(deviceJSON, &device)

	if device.OwnerID == userID {
		result := map[string]interface{}{
			"hasAccess":      true,
			"permissionType": "owner",
			"reason":         "Device owner",
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	// Check explicit permission
	permissionID := fmt.Sprintf("PERM_%s_%s", userID, deviceID)
	permJSON, err := ctx.GetStub().GetState(permissionID)
	if err != nil || permJSON == nil {
		result := map[string]interface{}{
			"hasAccess": false,
			"reason":    "No permission granted",
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	var permission AccessPermission
	json.Unmarshal(permJSON, &permission)

	if permission.Status != "active" {
		result := map[string]interface{}{
			"hasAccess": false,
			"reason":    "Permission revoked",
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	// Check expiration
	if permission.ExpiresAt > 0 && getCurrentTimestamp() > permission.ExpiresAt {
		result := map[string]interface{}{
			"hasAccess": false,
			"reason":    "Permission expired",
		}
		resultJSON, _ := json.Marshal(result)
		return string(resultJSON), nil
	}

	result := map[string]interface{}{
		"hasAccess":      true,
		"permissionType": permission.PermissionType,
		"reason":         "Permission granted",
	}
	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// GetUserPermissions returns all devices a user has access to
func (s *UserACLChaincode) GetUserPermissions(ctx contractapi.TransactionContextInterface, userID string) (string, error) {
	// Get user
	userJSON, err := ctx.GetStub().GetState("USER_" + userID)
	if err != nil || userJSON == nil {
		return "", fmt.Errorf("user not found")
	}

	var user User
	json.Unmarshal(userJSON, &user)

	var devices []string

	// If admin, return all devices
	if user.Role == "admin" {
		allDevices, _ := s.getAllDevices(ctx)
		for _, device := range allDevices {
			devices = append(devices, device.DeviceID)
		}
	} else {
		// Add owned devices
		devices = append(devices, user.OwnedDevices...)

		// Add devices with explicit permissions
		startKey := fmt.Sprintf("PERM_%s_", userID)
		endKey := fmt.Sprintf("PERM_%s~", userID)
		resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
		if err == nil {
			defer resultsIterator.Close()

			for resultsIterator.HasNext() {
				queryResponse, err := resultsIterator.Next()
				if err != nil {
					continue
				}

				var permission AccessPermission
				json.Unmarshal(queryResponse.Value, &permission)

				if permission.Status == "active" {
					devices = append(devices, permission.DeviceID)
				}
			}
		}
	}

	result := map[string]interface{}{
		"userID":  userID,
		"role":    user.Role,
		"devices": devices,
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

// GetUser retrieves user information (without password hash)
func (s *UserACLChaincode) GetUser(ctx contractapi.TransactionContextInterface, userID string) (string, error) {
	userJSON, err := ctx.GetStub().GetState("USER_" + userID)
	if err != nil || userJSON == nil {
		return "", fmt.Errorf("user not found")
	}

	var user User
	json.Unmarshal(userJSON, &user)

	// Remove password hash before returning
	user.PasswordHash = ""

	safeUserJSON, _ := json.Marshal(user)
	return string(safeUserJSON), nil
}

// GetDevice retrieves device information
func (s *UserACLChaincode) GetDevice(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil || deviceJSON == nil {
		return "", fmt.Errorf("device not found")
	}

	return string(deviceJSON), nil
}

// GetAllDevices returns all registered devices (admin only or user's accessible devices)
func (s *UserACLChaincode) GetAllDevices(ctx contractapi.TransactionContextInterface) (string, error) {
	devices, err := s.getAllDevices(ctx)
	if err != nil {
		return "", err
	}

	devicesJSON, _ := json.Marshal(devices)
	return string(devicesJSON), nil
}

// Helper functions

func (s *UserACLChaincode) getUserIDByUsername(ctx contractapi.TransactionContextInterface, username string) (string, error) {
	// Query composite key index
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("USERNAME_INDEX", []string{username})
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	if resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}

		// Extract userID from composite key
		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if err != nil {
			return "", err
		}

		if len(compositeKeyParts) > 1 {
			return compositeKeyParts[1], nil
		}
	}

	return "", fmt.Errorf("username not found")
}

func (s *UserACLChaincode) getAllDevices(ctx contractapi.TransactionContextInterface) ([]Device, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("DEVICE_", "DEVICE_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var devices []Device
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			continue
		}

		var device Device
		json.Unmarshal(queryResponse.Value, &device)
		devices = append(devices, device)
	}

	return devices, nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func generateToken(userID string) string {
	// Simplified token generation - in production use JWT
	data := fmt.Sprintf("%s_%d", userID, getCurrentTimestamp())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func main() {
	chaincode, err := contractapi.NewChaincode(&UserACLChaincode{})
	if err != nil {
		log.Panicf("Error creating USER-ACL chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting USER-ACL chaincode: %v", err)
	}
}

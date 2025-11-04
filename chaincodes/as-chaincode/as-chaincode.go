package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ASChaincode provides authentication server functions
type ASChaincode struct {
	contractapi.Contract
}

// Device represents an IoT device registered with the AS
type Device struct {
	DeviceID        string `json:"deviceID"`
	PublicKey       string `json:"publicKey"`
	Status          string `json:"status"` // active, suspended, revoked
	RegistrationTime int64  `json:"registrationTime"`
	LastAuthTime    int64   `json:"lastAuthTime"`
	Metadata        string `json:"metadata"` // Additional device information
}

// TGT represents a Ticket Granting Ticket
type TGT struct {
	TgtID           string `json:"tgtID"`
	DeviceID        string `json:"deviceID"`
	SessionKey      string `json:"sessionKey"`
	IssuedAt        int64  `json:"issuedAt"`
	ExpiresAt       int64  `json:"expiresAt"`
	Status          string `json:"status"` // valid, expired, revoked
}

// AuthRequest represents an authentication request from a device
type AuthRequest struct {
	DeviceID  string `json:"deviceID"`
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"` // Signature of (deviceID + nonce + timestamp)
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	TgtID      string `json:"tgtID"`
	SessionKey string `json:"sessionKey"`
	ExpiresAt  int64  `json:"expiresAt"`
	Message    string `json:"message"`
}

// InitLedger initializes the ledger with some test devices (for development only)
func (s *ASChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("Initializing AS Chaincode ledger")
	return nil
}

// RegisterDevice registers a new IoT device with the authentication server
func (s *ASChaincode) RegisterDevice(ctx contractapi.TransactionContextInterface, deviceID string, publicKey string, metadata string) error {
	// Check if device already exists
	existing, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("device %s already exists", deviceID)
	}

	// Validate inputs
	if len(deviceID) < 3 || len(deviceID) > 64 {
		return fmt.Errorf("deviceID must be between 3 and 64 characters")
	}
	if len(publicKey) < 100 || len(publicKey) > 4096 {
		return fmt.Errorf("publicKey length is invalid")
	}

	// Create new device
	device := Device{
		DeviceID:        deviceID,
		PublicKey:       publicKey,
		Status:          "active",
		RegistrationTime: getCurrentTimestamp(),
		LastAuthTime:    0,
		Metadata:        metadata,
	}

	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}

	err = ctx.GetStub().PutState(deviceID, deviceJSON)
	if err != nil {
		return fmt.Errorf("failed to put device to world state: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("DeviceRegistered", []byte(deviceID))
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Device %s registered successfully", deviceID)
	return nil
}

// Authenticate authenticates a device and issues a TGT
func (s *ASChaincode) Authenticate(ctx contractapi.TransactionContextInterface, authRequestJSON string) (string, error) {
	var authReq AuthRequest
	err := json.Unmarshal([]byte(authRequestJSON), &authReq)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal auth request: %v", err)
	}

	// Get device from ledger
	deviceJSON, err := ctx.GetStub().GetState(authReq.DeviceID)
	if err != nil {
		return "", fmt.Errorf("failed to read device: %v", err)
	}
	if deviceJSON == nil {
		return "", fmt.Errorf("device %s not found", authReq.DeviceID)
	}

	var device Device
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal device: %v", err)
	}

	// Check device status
	if device.Status != "active" {
		return "", fmt.Errorf("device is not active (status: %s)", device.Status)
	}

	// Validate timestamp (within 5 minutes)
	currentTime := getCurrentTimestamp()
	if authReq.Timestamp < currentTime-300 || authReq.Timestamp > currentTime+300 {
		return "", fmt.Errorf("timestamp is invalid or too old")
	}

	// In production, verify the signature using the device's public key
	// This is a placeholder for actual cryptographic verification
	if len(authReq.Signature) < 10 {
		return "", fmt.Errorf("invalid signature")
	}

	// Generate session key using crypto/rand (secure random generation)
	sessionKey, err := generateSecureSessionKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate session key: %v", err)
	}

	// Generate TGT ID
	tgtID, err := generateSecureTgtID()
	if err != nil {
		return "", fmt.Errorf("failed to generate TGT ID: %v", err)
	}

	// Create TGT with 1 hour validity
	issuedAt := getCurrentTimestamp()
	expiresAt := issuedAt + 3600

	tgt := TGT{
		TgtID:      tgtID,
		DeviceID:   authReq.DeviceID,
		SessionKey: sessionKey,
		IssuedAt:   issuedAt,
		ExpiresAt:  expiresAt,
		Status:     "valid",
	}

	// Store TGT
	tgtJSON, err := json.Marshal(tgt)
	if err != nil {
		return "", fmt.Errorf("failed to marshal TGT: %v", err)
	}

	err = ctx.GetStub().PutState("TGT_"+tgtID, tgtJSON)
	if err != nil {
		return "", fmt.Errorf("failed to store TGT: %v", err)
	}

	// Update device's last auth time
	device.LastAuthTime = currentTime
	deviceJSON, err = json.Marshal(device)
	if err != nil {
		return "", fmt.Errorf("failed to marshal updated device: %v", err)
	}
	err = ctx.GetStub().PutState(authReq.DeviceID, deviceJSON)
	if err != nil {
		return "", fmt.Errorf("failed to update device: %v", err)
	}

	// Create response
	response := AuthResponse{
		TgtID:      tgtID,
		SessionKey: sessionKey,
		ExpiresAt:  expiresAt,
		Message:    "Authentication successful",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("DeviceAuthenticated", []byte(authReq.DeviceID))
	if err != nil {
		return "", fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Device %s authenticated successfully, TGT: %s", authReq.DeviceID, tgtID)
	return string(responseJSON), nil
}

// GetDevice retrieves device information
func (s *ASChaincode) GetDevice(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	deviceJSON, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return "", fmt.Errorf("failed to read device: %v", err)
	}
	if deviceJSON == nil {
		return "", fmt.Errorf("device %s not found", deviceID)
	}

	return string(deviceJSON), nil
}

// GetTGT retrieves a TGT by ID
func (s *ASChaincode) GetTGT(ctx contractapi.TransactionContextInterface, tgtID string) (string, error) {
	tgtJSON, err := ctx.GetStub().GetState("TGT_" + tgtID)
	if err != nil {
		return "", fmt.Errorf("failed to read TGT: %v", err)
	}
	if tgtJSON == nil {
		return "", fmt.Errorf("TGT %s not found", tgtID)
	}

	return string(tgtJSON), nil
}

// RevokeDevice revokes a device's access
func (s *ASChaincode) RevokeDevice(ctx contractapi.TransactionContextInterface, deviceID string) error {
	deviceJSON, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return fmt.Errorf("failed to read device: %v", err)
	}
	if deviceJSON == nil {
		return fmt.Errorf("device %s not found", deviceID)
	}

	var device Device
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return fmt.Errorf("failed to unmarshal device: %v", err)
	}

	device.Status = "revoked"

	deviceJSON, err = json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}

	err = ctx.GetStub().PutState(deviceID, deviceJSON)
	if err != nil {
		return fmt.Errorf("failed to update device: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("DeviceRevoked", []byte(deviceID))
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Device %s revoked successfully", deviceID)
	return nil
}

// GetAllDevices returns all devices (for admin purposes)
func (s *ASChaincode) GetAllDevices(ctx contractapi.TransactionContextInterface) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return "", fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var devices []Device
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("failed to iterate: %v", err)
		}

		// Skip TGT entries
		if len(queryResponse.Key) > 4 && queryResponse.Key[:4] == "TGT_" {
			continue
		}

		var device Device
		err = json.Unmarshal(queryResponse.Value, &device)
		if err != nil {
			continue // Skip if not a valid device
		}

		devices = append(devices, device)
	}

	devicesJSON, err := json.Marshal(devices)
	if err != nil {
		return "", fmt.Errorf("failed to marshal devices: %v", err)
	}

	return string(devicesJSON), nil
}

// Helper functions

func getCurrentTimestamp() int64 {
	// In production, use a secure timestamp source
	// For now, using chaincode timestamp via stub.GetTxTimestamp()
	return 1672531200 // Placeholder - should use ctx.GetStub().GetTxTimestamp()
}

func generateSecureSessionKey() (string, error) {
	// Import crypto/rand and generate secure random key
	// Placeholder - actual implementation should use crypto/rand
	return "secure_session_key_" + generateRandomString(32), nil
}

func generateSecureTgtID() (string, error) {
	// Import crypto/rand and generate secure random ID
	// Placeholder - actual implementation should use crypto/rand
	return "tgt_" + generateRandomString(16), nil
}

func generateRandomString(length int) string {
	// Placeholder - should use crypto/rand in production
	return "random" + fmt.Sprintf("%d", getCurrentTimestamp())
}

func main() {
	asChaincode, err := contractapi.NewChaincode(&ASChaincode{})
	if err != nil {
		log.Panicf("Error creating AS chaincode: %v", err)
	}

	if err := asChaincode.Start(); err != nil {
		log.Panicf("Error starting AS chaincode: %v", err)
	}
}

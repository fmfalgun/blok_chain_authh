package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ISVChaincode provides functions for IoT Service Validator operations
type ISVChaincode struct {
	contractapi.Contract
}

// ServiceTicket represents a ticket for accessing ISV services (received from TGS)
type ServiceTicket struct {
	ClientID   string    `json:"clientID"`
	SessionKey string    `json:"sessionKey"`  // KU,SS - session key for client-ISV communication
	Timestamp  time.Time `json:"timestamp"`
	Lifetime   int64     `json:"lifetime"`    // Lifetime in seconds
}

// IoTDevice represents an IoT device registered with the ISV
type IoTDevice struct {
	DeviceID      string    `json:"deviceID"`
	PublicKey     string    `json:"publicKey"`
	Status        string    `json:"status"`       // "active", "inactive", "busy"
	LastSeen      time.Time `json:"lastSeen"`
	RegisteredAt  time.Time `json:"registeredAt"`
	Capabilities  []string  `json:"capabilities"` // Device capabilities/services
}

// ServiceRequest represents a client's request to access an IoT device
type ServiceRequest struct {
	EncryptedServiceTicket string `json:"encryptedServiceTicket"` // Service ticket from TGS
	ClientID              string `json:"clientID"`
	DeviceID              string `json:"deviceID"`
	RequestType           string `json:"requestType"`
	EncryptedData         string `json:"encryptedData"` // Additional data encrypted with session key
}

// ServiceResponse represents ISV's response to a client's service request
type ServiceResponse struct {
	ClientID        string `json:"clientID"`
	DeviceID        string `json:"deviceID"`
	Status          string `json:"status"`          // "granted", "denied", "device_unavailable"
	SessionID       string `json:"sessionID"`       // Unique session identifier if granted
	EncryptedData   string `json:"encryptedData"`   // Response data encrypted with session key
}

// ClientDeviceSession represents an active session between a client and IoT device
type ClientDeviceSession struct {
	SessionID     string    `json:"sessionID"`
	ClientID      string    `json:"clientID"`
	DeviceID      string    `json:"deviceID"`
	SessionKey    string    `json:"sessionKey"`
	EstablishedAt time.Time `json:"establishedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
	Status        string    `json:"status"`        // "active", "terminated"
}

// PredefinedKeys holds the predefined keys for deterministic initialization
type PredefinedKeys struct {
	ISVPrivateKey string
	ISVPublicKey  string
}

// getDeterministicTimestamp gets a deterministic timestamp from the transaction context
func getDeterministicTimestamp(ctx contractapi.TransactionContextInterface) (time.Time, error) {
    // Get timestamp from transaction context - this will be identical across all peers
    txTimestamp, err := ctx.GetStub().GetTxTimestamp()
    if err != nil {
        return time.Time{}, fmt.Errorf("failed to get transaction timestamp: %v", err)
    }
    
    // Convert to Go time.Time
    return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)), nil
}

// Initialize sets up the chaincode state
// This function is called when the chaincode is instantiated
func (s *ISVChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	// Check if already initialized to make this idempotent
	existingKey, err := ctx.GetStub().GetState("ISV_INITIALIZED")
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %v", err)
	}
	
	if existingKey != nil {
		// Already initialized, skip to maintain consistency
		return nil
	}
	
	// Use predefined keys instead of generating them dynamically
	keys := getPredefinedKeys()
	
	// Store the ISV private key
	err = ctx.GetStub().PutState("ISV_PRIVATE_KEY", []byte(keys.ISVPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to store ISV private key: %v", err)
	}
	
	// Store the ISV public key
	err = ctx.GetStub().PutState("ISV_PUBLIC_KEY", []byte(keys.ISVPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store ISV public key: %v", err)
	}
	
	// Mark as initialized
	err = ctx.GetStub().PutState("ISV_INITIALIZED", []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark ISV as initialized: %v", err)
	}
	
	return nil
}

// getPredefinedKeys returns the predefined cryptographic keys for deterministic initialization
func getPredefinedKeys() PredefinedKeys {
	// These keys are hardcoded for consistent initialization across all peers
	// In a production system, these could be loaded from secure configuration
	return PredefinedKeys{
		ISVPrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEApqAtGdmCJr3GYzs6fSQiN1PO3GFiDtEAJyWbxRpKJRPv6/GG
BLSqr5QQjDw7Vy1RwFXW7Z+j0/8C8xOBtu5JUPoNBRJ5DMRyHGlGqxQgLjEySt8s
ObaJVq9WyHoNTLCD3lsmExxhhHM+ccc8dSZSpX9qXAoHYvGZ0SJpGPBd7OXUQgzI
UlJZRKP9Qz+d472xVMzpCrFJpPGkKcL1WoCPGSgS3cx8NUb2xZnUHD1mmIyVwaDFm
5RU4aBHrj/jx/tR9Dy0MKJC61/HAZEdU8zZc3kD/7PbsU0RXDzNzG8i8UtXSJYjgw
BQhVlPn0/aQeiI7fk+Jf8E5zGtpKGI9L+RCQIDAQABAoIBAQCDXY4cG9Yf0sms7SV
SrES0F+abE1nYqCzE4/N9QZlrWDGkSvQj2Hj0iQwJxHKP5XSjBZLJw3ULqU8JwZN
L5JgbDhDNs0vCamT8nSEhP56/0PSJQfbXN8xB9tp8qGbIsdW5s/G2cK0qROJdT9C
e13Wd0c0jGxYqbbjIJDZygvUzFZXQVY6eymwXIxpWKl40ZkZtXIFMwIosP9/UitN
yBBJwgPK0iRxnBgydD1qIQYZbBL6IGUii73iLhLZvj1SNSGMdz0ni/A/dTNu878S
mlWlCkTOlFgJDmxb8d2JXkYxkQBAdRJk5FhFliW5qj5aprIbMqQzLcxLp/+n1bqR
c7vqd+NBAoGBAOQUYCZZ/yhNAooOQiBBfxj7SgI0PWsjNndwkJLCZtRCqm9Qm49p
oJOX1WsQDlc2QY1KG75+ms4Cq+EBxwY+lxEGaQSiA9BbHtWfHtSaXHayR9lRHNzL
FT+zdJJ+RkCdmSfL9upAo8/EPVn9CJV5wYXZZlXJaS/59lnqpxJSxI4JAoGBALuR
ufD62zl83TUJmW3gwBbQYKTxFkLGxGa0yZ5fLNDBFXfk4k/1xKxEX4MwbBLhDaQG
lxhLDK0jzgmFKP+VI8h5HwOgdBj03181+uEPGDHCQNqXu0XBsGHdztjIiqXM8OCR
J4ZYvyUjB5m0VzGoKQO66FMIjWVp7TqOfnwt19pRAoGAbkmX4iKJPzdH4wCo1rwd
N8DQZRQb3Blahm7zFdWF4a0IWjazZ/l+J7fieXwFEa9VvORJk8vgWwqHSyqLIT0h
y/kGcIhXMvqiBXPEYmA7GqvN1cjL8HnFXF5tL2FBLW1BO7nRU3B9VvlZ0W61+1CU
EZkdGSGjmWztZQ/8qfRCZmkCgYAqafBGZFnwB9EvWr4d4ZU3MFCpsa1tAroKimNz
e4TnZXDIjupKkIGMNRIJveT4IiYIoLKOXJ+Wjak28Ft7TZ45ldGS5QQEKSjxwIUD
6NtTGzwYL9FnYLxHFZ6PUWrgPNFNp4gpqrLQHZnRy9aCiGVXcRSKz1W8dSChUEsT
T/HfAQKBgQC+IFG/l3qPltDDxPo09QsH6LFpXCxLr5lyOwuFdZMMkmYSEHXcG/Z6
8cXP3kAmQCgAQbB2+T4CBJCceFC4LA6GOKrOg9IHPB8jrwmgpqAvt6OCJJRFJqgS
R0uXRj5xjUyNY4h9hnTB8Y0z23YaqnEa4/vQYHrI01YKldzfxPatvQ==
-----END RSA PRIVATE KEY-----`,
		ISVPublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApqAtGdmCJr3GYzs6fSQi
N1PO3GFiDtEAJyWbxRpKJRPv6/GGBLSqr5QQjDw7Vy1RwFXW7Z+j0/8C8xOBtu5J
UPoNBRJ5DMRyHGlGqxQgLjEySt8sObaJVq9WyHoNTLCD3lsmExxhhHM+ccc8dSZS
pX9qXAoHYvGZ0SJpGPBd7OXUQgzIUlJZRKP9Qz+d472xVMzpCrFJpPGkKcL1WoCP
GSgS3cx8NUb2xZnUHD1mmIyVwaDFm5RU4aBHrj/jx/tR9Dy0MKJC61/HAZEdU8zZ
c3kD/7PbsU0RXDzNzG8i8UtXSJYjgwBQhVlPn0/aQeiI7fk+Jf8E5zGtpKGI9L+R
CQIDAQAB
-----END PUBLIC KEY-----`,
	}
}

// ==================== Helper Functions ====================

// getPrivateKey retrieves the ISV's private key from the chaincode state
func (s *ISVChaincode) getPrivateKey(ctx contractapi.TransactionContextInterface) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := ctx.GetStub().GetState("ISV_PRIVATE_KEY")
	if err != nil {
		return nil, err
	}
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("ISV private key not found")
	}
	
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	
	return privateKey, nil
}

// getDevicePublicKey retrieves a device's public key from the chaincode state
func (s *ISVChaincode) getDevicePublicKey(ctx contractapi.TransactionContextInterface, deviceID string) (*rsa.PublicKey, error) {
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil {
		return nil, err
	}
	if deviceJSON == nil {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}
	
	var device IoTDevice
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return nil, err
	}
	
	block, _ := pem.Decode([]byte(device.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing device public key")
	}
	
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	return publicKey, nil
}

// ==================== Core ISV Operations ====================

// RegisterIoTDevice registers a new IoT device with the ISV
// This implements the "Register IoT devices" operation
func (s *ISVChaincode) RegisterIoTDevice(ctx contractapi.TransactionContextInterface, deviceID string, devicePublicKeyPEM string, capabilities []string) error {
	// Check if device already exists
	existingDeviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existingDeviceJSON != nil {
		return fmt.Errorf("device %s already exists", deviceID)
	}
	
	// Verify the provided public key is valid
	block, _ := pem.Decode([]byte(devicePublicKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block containing public key")
	}
	
	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %v", err)
	}
	
	// Use deterministic timestamp
	registrationTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get registration timestamp: %v", err)
	}
	
	// Create and store the IoT device record
	device := IoTDevice{
		DeviceID:      deviceID,
		PublicKey:     devicePublicKeyPEM,
		Status:        "active",
		LastSeen:      registrationTime,
		RegisteredAt:  registrationTime,
		Capabilities:  capabilities,
	}
	
	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}
	
	// Store device data in the world state
	err = ctx.GetStub().PutState("DEVICE_"+deviceID, deviceJSON)
	if err != nil {
		return err
	}
	
	// Record this registration on the blockchain with deterministic ID
	registrationEvent := struct {
		DeviceID      string    `json:"deviceID"`
		Timestamp     time.Time `json:"timestamp"`
		Capabilities  []string  `json:"capabilities"`
	}{
		DeviceID:      deviceID,
		Timestamp:     registrationTime,
		Capabilities:  capabilities,
	}
	
	registrationEventJSON, err := json.Marshal(registrationEvent)
	if err != nil {
		return err
	}
	
	// Create a deterministic registration ID
	registrationID := "DEVICE_REG_" + deviceID + "_" + strconv.FormatInt(registrationTime.Unix(), 10)
	return ctx.GetStub().PutState(registrationID, registrationEventJSON)
}

// UpdateDeviceStatus updates the availability status of an IoT device
// This is part of the "Check availability of IoT devices" operation
func (s *ISVChaincode) UpdateDeviceStatus(ctx contractapi.TransactionContextInterface, deviceID string, status string, signature string) error {
	// Retrieve the device record
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil {
		return fmt.Errorf("failed to read device data: %v", err)
	}
	if deviceJSON == nil {
		return fmt.Errorf("device %s does not exist", deviceID)
	}
	
	var device IoTDevice
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return err
	}
	
	// In a real implementation, we would verify the signature here
	// The signature would be created by the device using its private key
	// And we would verify it using the device's public key to ensure authenticity
	// For simplicity, we'll skip this verification in this example
	
	// Update the device status
	updateTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get update timestamp: %v", err)
	}
	
	device.Status = status
	device.LastSeen = updateTime
	
	updatedDeviceJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}
	
	// Store the updated device record
	err = ctx.GetStub().PutState("DEVICE_"+deviceID, updatedDeviceJSON)
	if err != nil {
		return err
	}
	
	// Record this status update on the blockchain with deterministic ID
	statusUpdateEvent := struct {
		DeviceID      string    `json:"deviceID"`
		Status        string    `json:"status"`
		Timestamp     time.Time `json:"timestamp"`
	}{
		DeviceID:      deviceID,
		Status:        status,
		Timestamp:     updateTime,
	}
	
	statusUpdateEventJSON, err := json.Marshal(statusUpdateEvent)
	if err != nil {
		return err
	}
	
	// Create a deterministic status update ID
	statusUpdateID := "DEVICE_STATUS_" + deviceID + "_" + strconv.FormatInt(updateTime.Unix(), 10)
	return ctx.GetStub().PutState(statusUpdateID, statusUpdateEventJSON)
}

// CheckDeviceAvailability checks if an IoT device is available for connection
// This implements the "Check availability of IoT devices" operation
func (s *ISVChaincode) CheckDeviceAvailability(ctx contractapi.TransactionContextInterface, deviceID string) (bool, error) {
	// Retrieve the device record
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + deviceID)
	if err != nil {
		return false, fmt.Errorf("failed to read device data: %v", err)
	}
	if deviceJSON == nil {
		return false, fmt.Errorf("device %s does not exist", deviceID)
	}
	
	var device IoTDevice
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return false, err
	}
	
	// Check if the device is active and not busy
	if device.Status == "active" {
		return true, nil
	}
	
	return false, nil
}

// ValidateServiceTicket validates a service ticket from TGS
// This implements the "Check for record & validity of Org2 registration" operation
// and Step 5: Client Requests Service from ISV from the paper
func (s *ISVChaincode) ValidateServiceTicket(ctx contractapi.TransactionContextInterface, encryptedServiceTicket string) (*ServiceTicket, error) {
	// Decode the base64 encoded encrypted service ticket
	serviceTicketBytes, err := base64.StdEncoding.DecodeString(encryptedServiceTicket)
	if err != nil {
		return nil, fmt.Errorf("invalid service ticket format: %v", err)
	}
	
	// Get the ISV private key
	privateKey, err := s.getPrivateKey(ctx)
	if err != nil {
		return nil, err
	}
	
	// Decrypt the service ticket using ISV's private key
	// This implements: M = TSS^dISV = (M^eISV)^dISV mod nISV from the paper
	decryptedServiceTicketBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, serviceTicketBytes)
	if err != nil {
		return nil, fmt.Errorf("service ticket decryption failed: %v", err)
	}
	
	// Parse the decrypted service ticket
	var serviceTicket ServiceTicket
	err = json.Unmarshal(decryptedServiceTicketBytes, &serviceTicket)
	if err != nil {
		return nil, fmt.Errorf("invalid service ticket structure: %v", err)
	}
	
	// Validate the service ticket timestamp and lifetime
	currentTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current timestamp: %v", err)
	}
	
	if currentTime.After(serviceTicket.Timestamp.Add(time.Duration(serviceTicket.Lifetime) * time.Second)) {
		return nil, fmt.Errorf("service ticket has expired")
	}
	
	// Store the session key for later use with deterministic ID
	sessionKeyID := "SESSION_KEY_" + serviceTicket.ClientID + "_" + strconv.FormatInt(serviceTicket.Timestamp.Unix(), 10)
	err = ctx.GetStub().PutState(sessionKeyID, []byte(serviceTicket.SessionKey))
	if err != nil {
		return nil, err
	}
	
	return &serviceTicket, nil
}

// ProcessServiceRequest processes a client's request to access an IoT device
// This implements the "Endorse & validate registration" operation
// and part of Step 6: Service Exchange Between IoT (ISV) and Client from the paper
func (s *ISVChaincode) ProcessServiceRequest(ctx contractapi.TransactionContextInterface, requestJSON string) (*ServiceResponse, error) {
	var request ServiceRequest
	err := json.Unmarshal([]byte(requestJSON), &request)
	if err != nil {
		return nil, fmt.Errorf("invalid request format: %v", err)
	}
	
	// Step 1: Validate the service ticket
	serviceTicket, err := s.ValidateServiceTicket(ctx, request.EncryptedServiceTicket)
	if err != nil {
		return nil, err
	}
	
	// Verify that the client ID in the request matches the one in the service ticket
	if request.ClientID != serviceTicket.ClientID {
		return nil, fmt.Errorf("client ID mismatch")
	}
	
	// Step 2: Check device availability
	available, err := s.CheckDeviceAvailability(ctx, request.DeviceID)
	if err != nil {
		return nil, err
	}
	if !available {
		return &ServiceResponse{
			ClientID: request.ClientID,
			DeviceID: request.DeviceID,
			Status:   "device_unavailable",
		}, nil
	}
	
	// Step 3: Create a session between the client and the device with deterministic approach
	currentTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current timestamp: %v", err)
	}
	
	sessionID := "SESSION_" + request.ClientID + "_" + request.DeviceID + "_" + strconv.FormatInt(currentTime.Unix(), 10)
	
	expiryTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get expiry timestamp: %v", err)
	}
	
	session := ClientDeviceSession{
		SessionID:     sessionID,
		ClientID:      request.ClientID,
		DeviceID:      request.DeviceID,
		SessionKey:    serviceTicket.SessionKey,
		EstablishedAt: currentTime,
		ExpiresAt:     expiryTime.Add(time.Hour), // 1 hour session
		Status:        "active",
	}
	
	// Store the session record
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	
	err = ctx.GetStub().PutState(sessionID, sessionJSON)
	if err != nil {
		return nil, err
	}
	
	// Update device status to "busy"
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + request.DeviceID)
	if err != nil {
		return nil, err
	}
	
	var device IoTDevice
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return nil, err
	}
	
	device.Status = "busy"
	updatedDeviceJSON, err := json.Marshal(device)
	if err != nil {
		return nil, err
	}
	
	err = ctx.GetStub().PutState("DEVICE_"+request.DeviceID, updatedDeviceJSON)
	if err != nil {
		return nil, err
	}
	
	// Prepare and encrypt response data for the client
	// In a real implementation, this would be encrypted with the session key
	// For this example, we'll use a deterministic approach
	responseData := fmt.Sprintf("Connection established with device %s at %s", request.DeviceID, currentTime.Format(time.RFC3339))
	responseHash := sha256.Sum256([]byte(responseData))
	encryptedResponseData := base64.StdEncoding.EncodeToString(responseHash[:])
	
	// Create the response
	response := ServiceResponse{
		ClientID:      request.ClientID,
		DeviceID:      request.DeviceID,
		Status:        "granted",
		SessionID:     sessionID,
		EncryptedData: encryptedResponseData,
	}
	
	// Record this service grant on the blockchain
	recordTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get record timestamp: %v", err)
	}
	
	serviceGrantEvent := struct {
		ClientID      string    `json:"clientID"`
		DeviceID      string    `json:"deviceID"`
		SessionID     string    `json:"sessionID"`
		Timestamp     time.Time `json:"timestamp"`
	}{
		ClientID:      request.ClientID,
		DeviceID:      request.DeviceID,
		SessionID:     sessionID,
		Timestamp:     recordTime,
	}
	
	serviceGrantEventJSON, err := json.Marshal(serviceGrantEvent)
	if err != nil {
		return nil, err
	}
	
	// Store the service grant record with deterministic ID
	serviceGrantID := "SERVICE_GRANT_" + request.ClientID + "_" + request.DeviceID + "_" + strconv.FormatInt(recordTime.Unix(), 10)
	err = ctx.GetStub().PutState(serviceGrantID, serviceGrantEventJSON)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

// HandleDeviceResponse processes a device's response to a client's request
// This implements the Step 6.2: ISV Sends the Service Response Back to the Client from the paper
func (s *ISVChaincode) HandleDeviceResponse(ctx contractapi.TransactionContextInterface, sessionID string, deviceResponse string) error {
	// Retrieve the session record
	sessionJSON, err := ctx.GetStub().GetState(sessionID)
	if err != nil {
		return fmt.Errorf("failed to read session data: %v", err)
	}
	if sessionJSON == nil {
		return fmt.Errorf("session %s does not exist", sessionID)
	}
	
	var session ClientDeviceSession
	err = json.Unmarshal(sessionJSON, &session)
	if err != nil {
		return err
	}
	
	// Verify that the session is active
	if session.Status != "active" {
		return fmt.Errorf("session is not active")
	}
	
	// Store the device response for the client to retrieve with deterministic approach
	currentTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current timestamp: %v", err)
	}
	
	responseRecord := struct {
		SessionID      string    `json:"sessionID"`
		DeviceResponse string    `json:"deviceResponse"`
		Timestamp      time.Time `json:"timestamp"`
	}{
		SessionID:      sessionID,
		DeviceResponse: deviceResponse,
		Timestamp:      currentTime,
	}
	
	responseRecordJSON, err := json.Marshal(responseRecord)
	if err != nil {
		return err
	}
	
	// Store the response record with deterministic ID
	responseID := "RESPONSE_" + sessionID + "_" + strconv.FormatInt(currentTime.Unix(), 10)
	return ctx.GetStub().PutState(responseID, responseRecordJSON)
}

// CloseSession terminates a session between a client and an IoT device
func (s *ISVChaincode) CloseSession(ctx contractapi.TransactionContextInterface, sessionID string) error {
	// Retrieve the session record
	sessionJSON, err := ctx.GetStub().GetState(sessionID)
	if err != nil {
		return fmt.Errorf("failed to read session data: %v", err)
	}
	if sessionJSON == nil {
		return fmt.Errorf("session %s does not exist", sessionID)
	}
	
	var session ClientDeviceSession
	err = json.Unmarshal(sessionJSON, &session)
	if err != nil {
		return err
	}
	
	// Update the session status
	currentTime, err := getDeterministicTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current timestamp: %v", err)
	}
	
	session.Status = "terminated"
	
	updatedSessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}
	
	// Store the updated session record
	err = ctx.GetStub().PutState(sessionID, updatedSessionJSON)
	if err != nil {
		return err
	}
	
	// Update device status back to "active"
	deviceJSON, err := ctx.GetStub().GetState("DEVICE_" + session.DeviceID)
	if err != nil {
		return err
	}
	
	var device IoTDevice
	err = json.Unmarshal(deviceJSON, &device)
	if err != nil {
		return err
	}
	
	device.Status = "active"
	device.LastSeen = currentTime
	updatedDeviceJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}
	
	return ctx.GetStub().PutState("DEVICE_"+session.DeviceID, updatedDeviceJSON)
}

// GetAllIoTDevices retrieves all registered IoT devices
func (s *ISVChaincode) GetAllIoTDevices(ctx contractapi.TransactionContextInterface) ([]*IoTDevice, error) {
	// Get all IoT devices from the world state
	resultsIterator, err := ctx.GetStub().GetStateByRange("DEVICE_", "DEVICE_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	
	var devices []*IoTDevice
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		
		var device IoTDevice
		err = json.Unmarshal(queryResponse.Value, &device)
		if err != nil {
			return nil, err
		}
		
		devices = append(devices, &device)
	}
	
	return devices, nil
}

// GetActiveSessionsByClient retrieves all active sessions for a specific client
func (s *ISVChaincode) GetActiveSessionsByClient(ctx contractapi.TransactionContextInterface, clientID string) ([]*ClientDeviceSession, error) {
	// Get all sessions from the world state
	resultsIterator, err := ctx.GetStub().GetStateByRange("SESSION_", "SESSION_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	
	var sessions []*ClientDeviceSession
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		
		var session ClientDeviceSession
		err = json.Unmarshal(queryResponse.Value, &session)
		if err != nil {
			return nil, err
		}
		
		// Filter for active sessions belonging to the specified client
		if session.ClientID == clientID && session.Status == "active" {
			sessions = append(sessions, &session)
		}
	}
	
	return sessions, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&ISVChaincode{})
	if err != nil {
		fmt.Printf("Error creating ISV chaincode: %s", err.Error())
		return
	}
	
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting ISV chaincode: %s", err.Error())
	}
}

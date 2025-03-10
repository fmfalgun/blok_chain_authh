package main

import (
	"crypto/rand"
	"crypto/rsa"
	//"crypto/sha256"
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

// Initialize sets up the chaincode state
// This function is called when the chaincode is instantiated
func (s *ISVChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	// Initialize the ISV server's own RSA key pair
	err := s.generateAndStoreISVKeyPair(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize ISV key pair: %v", err)
	}
	
	return nil
}

// ==================== Helper Functions ====================

// generateAndStoreISVKeyPair creates and stores the ISV's RSA key pair
// This implements the RSA key generation as described in the paper section 3.2
func (s *ISVChaincode) generateAndStoreISVKeyPair(ctx contractapi.TransactionContextInterface) error {
	// Generate a new RSA key pair with 2048 bits
	// In RSA key generation, this creates:
	// 1. Two large prime numbers p and q
	// 2. Computes modulus n = p × q
	// 3. Calculates Euler's totient φ(n) = (p−1)×(q−1)
	// 4. Chooses public exponent e (usually 65537)
	// 5. Computes private exponent d so that d × e ≡ 1 (mod φ(n))
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	
	// Encode the private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	
	// Encode the public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	
	// Store the keys in the chaincode state
	err = ctx.GetStub().PutState("ISV_PRIVATE_KEY", privateKeyPEM)
	if err != nil {
		return err
	}
	
	// The public key is also stored on the blockchain as described in the paper
	// This allows for transparent verification by all participants
	err = ctx.GetStub().PutState("ISV_PUBLIC_KEY", publicKeyPEM)
	if err != nil {
		return err
	}
	
	return nil
}

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

// getPublicKey retrieves a device's public key from the chaincode state
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
	
	// Create and store the IoT device record
	device := IoTDevice{
		DeviceID:      deviceID,
		PublicKey:     devicePublicKeyPEM,
		Status:        "active",
		LastSeen:      time.Now(),
		RegisteredAt:  time.Now(),
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
	
	// Record this registration on the blockchain
	registrationEvent := struct {
		DeviceID      string    `json:"deviceID"`
		Timestamp     time.Time `json:"timestamp"`
		Capabilities  []string  `json:"capabilities"`
	}{
		DeviceID:      deviceID,
		Timestamp:     time.Now(),
		Capabilities:  capabilities,
	}
	
	registrationEventJSON, err := json.Marshal(registrationEvent)
	if err != nil {
		return err
	}
	
	// Create a unique registration ID
	registrationID := "DEVICE_REG_" + deviceID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
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
	device.Status = status
	device.LastSeen = time.Now()
	
	updatedDeviceJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}
	
	// Store the updated device record
	err = ctx.GetStub().PutState("DEVICE_"+deviceID, updatedDeviceJSON)
	if err != nil {
		return err
	}
	
	// Record this status update on the blockchain
	statusUpdateEvent := struct {
		DeviceID      string    `json:"deviceID"`
		Status        string    `json:"status"`
		Timestamp     time.Time `json:"timestamp"`
	}{
		DeviceID:      deviceID,
		Status:        status,
		Timestamp:     time.Now(),
	}
	
	statusUpdateEventJSON, err := json.Marshal(statusUpdateEvent)
	if err != nil {
		return err
	}
	
	// Create a unique status update ID
	statusUpdateID := "DEVICE_STATUS_" + deviceID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
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
	if time.Now().After(serviceTicket.Timestamp.Add(time.Duration(serviceTicket.Lifetime) * time.Second)) {
		return nil, fmt.Errorf("service ticket has expired")
	}
	
	// Store the session key for later use
	err = ctx.GetStub().PutState("SESSION_KEY_"+serviceTicket.ClientID, []byte(serviceTicket.SessionKey))
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
	
	// Step 3: Create a session between the client and the device
	sessionID := "SESSION_" + request.ClientID + "_" + request.DeviceID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	
	session := ClientDeviceSession{
		SessionID:     sessionID,
		ClientID:      request.ClientID,
		DeviceID:      request.DeviceID,
		SessionKey:    serviceTicket.SessionKey,
		EstablishedAt: time.Now(),
		ExpiresAt:     time.Now().Add(time.Hour), // 1 hour session
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
	// For this example, we'll use a placeholder
	responseData := fmt.Sprintf("Connection established with device %s", request.DeviceID)
	encryptedResponseData := base64.StdEncoding.EncodeToString([]byte(responseData))
	
	// Create the response
	response := ServiceResponse{
		ClientID:      request.ClientID,
		DeviceID:      request.DeviceID,
		Status:        "granted",
		SessionID:     sessionID,
		EncryptedData: encryptedResponseData,
	}
	
	// Record this service grant on the blockchain
	serviceGrantEvent := struct {
		ClientID      string    `json:"clientID"`
		DeviceID      string    `json:"deviceID"`
		SessionID     string    `json:"sessionID"`
		Timestamp     time.Time `json:"timestamp"`
	}{
		ClientID:      request.ClientID,
		DeviceID:      request.DeviceID,
		SessionID:     sessionID,
		Timestamp:     time.Now(),
	}
	
	serviceGrantEventJSON, err := json.Marshal(serviceGrantEvent)
	if err != nil {
		return nil, err
	}
	
	// Store the service grant record
	serviceGrantID := "SERVICE_GRANT_" + request.ClientID + "_" + request.DeviceID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
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
	
	// Store the device response for the client to retrieve
	// In a real implementation, this would be encrypted with the session key
	responseRecord := struct {
		SessionID      string    `json:"sessionID"`
		DeviceResponse string    `json:"deviceResponse"`
		Timestamp      time.Time `json:"timestamp"`
	}{
		SessionID:      sessionID,
		DeviceResponse: deviceResponse,
		Timestamp:      time.Now(),
	}
	
	responseRecordJSON, err := json.Marshal(responseRecord)
	if err != nil {
		return err
	}
	
	// Store the response record
	responseID := "RESPONSE_" + sessionID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
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

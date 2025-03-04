#!/bin/bash

# Exit on first error
set -e

echo "===== Preparing for Chaincode Deployment ====="

# Set environment variables for peer commands
export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_ORG1_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export PEER0_ORG2_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export PEER0_ORG3_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt

# Define variables
CHANNEL_NAME="chaichis-channel"
CC_SRC_PATH="/opt/gopath/src/github.com/chaincode/iot-auth/go"
CC_NAME="iot-auth"
CC_VERSION="1.0"
CC_SEQUENCE="1"
CC_INIT_FCN="InitLedger"

# Create chaincode directory if it doesn't exist
echo "Creating chaincode directory structure..."
mkdir -p $CC_SRC_PATH

# Write the chaincode file
echo "Writing chaincode file..."
cat > $CC_SRC_PATH/iot-auth.go << 'EOF'
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// IoTAuthContract provides functions for managing IoT authentication
type IoTAuthContract struct {
	contractapi.Contract
}

// ClientRegistration represents a client registration record
type ClientRegistration struct {
	ClientID      string    `json:"clientId"`
	PublicKey     string    `json:"publicKey"`
	Organization  string    `json:"organization"`
	Status        string    `json:"status"` // Pending, Validated by AS, Validated by TGS, Validated by ISV
	Timestamp     time.Time `json:"timestamp"`
	ExpiryTime    time.Time `json:"expiryTime"`
	Nonce         string    `json:"nonce,omitempty"`
	SessionKey    string    `json:"sessionKey,omitempty"`
	TGT           string    `json:"tgt,omitempty"`
	ServiceTicket string    `json:"serviceTicket,omitempty"`
}

// IoTDevice represents an IoT device registered in the system
type IoTDevice struct {
	DeviceID     string    `json:"deviceId"`
	PublicKey    string    `json:"publicKey"`
	Status       string    `json:"status"` // Active, Inactive
	Organization string    `json:"organization"`
	LastUpdate   time.Time `json:"lastUpdate"`
	DeviceType   string    `json:"deviceType"`
}

// Transaction represents a transaction record for audit purposes
type Transaction struct {
	TxID        string    `json:"txId"`
	ClientID    string    `json:"clientId"`
	Type        string    `json:"type"` // Registration, TGT Request, Service Ticket Request, etc.
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
}

// InitLedger initializes the ledger with sample data
func (s *IoTAuthContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Nothing to initialize - will be populated through use
	return nil
}

// ============================================================
// Org1 (AS - Authentication Server) Functions
// ============================================================

// RegisterClient registers a new client in the system
func (s *IoTAuthContract) RegisterClient(ctx contractapi.TransactionContextInterface, clientID string, publicKey string, organization string) error {
	// Check if client already exists
	exists, err := s.ClientExists(ctx, clientID)
	if err != nil {
		return fmt.Errorf("failed to check if client exists: %v", err)
	}
	if exists {
		return fmt.Errorf("client already exists: %s", clientID)
	}

	// Create a new client registration
	registration := ClientRegistration{
		ClientID:     clientID,
		PublicKey:    publicKey,
		Organization: organization,
		Status:       "Pending",
		Timestamp:    time.Now(),
		ExpiryTime:   time.Now().Add(24 * time.Hour), // Default validity of 24 hours
	}

	// Store client registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to put client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "Registration", clientID, "Success", "Client registered successfully")

	return nil
}

// ValidateClient validates a client registration (AS function)
func (s *IoTAuthContract) ValidateClient(ctx contractapi.TransactionContextInterface, clientID string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Update status to "Validated by AS"
	registration.Status = "Validated by AS"
	
	// Generate a nonce
	nonce := generateNonce()
	registration.Nonce = nonce
	
	// Update the registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to update client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "ClientValidation", clientID, "Success", "Client validated by AS")

	return nil
}

// IssueTicketGrantingTicket issues a TGT to a validated client
func (s *IoTAuthContract) IssueTicketGrantingTicket(ctx contractapi.TransactionContextInterface, clientID string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Check if client is validated by AS
	if registration.Status != "Validated by AS" {
		return fmt.Errorf("client is not validated by AS: %s", clientID)
	}

	// Generate a session key
	sessionKey := generateSessionKey()
	registration.SessionKey = sessionKey

	// Generate a TGT (in a real system, this would be encrypted)
	tgt := generateTGT(clientID, registration.PublicKey, sessionKey)
	registration.TGT = tgt

	// Update the registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to update client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "TGTIssue", clientID, "Success", "TGT issued to client")

	return nil
}

// GetAllClients queries all clients
func (s *IoTAuthContract) GetAllClients(ctx contractapi.TransactionContextInterface) ([]*ClientRegistration, error) {
	// Get all clients from the ledger
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %v", err)
	}
	defer resultsIterator.Close()

	var clients []*ClientRegistration
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next client: %v", err)
		}

		// Skip if this is not a client entry (we need to add more sophisticated filtering in real system)
		if queryResponse.Key[:7] == "DEVICE_" || queryResponse.Key[:3] == "TX_" {
			continue
		}

		var client ClientRegistration
		err = json.Unmarshal(queryResponse.Value, &client)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal client: %v", err)
		}
		
		clients = append(clients, &client)
	}

	return clients, nil
}

// AllocatePeerTasks allocates tasks to peers for processing client requests
func (s *IoTAuthContract) AllocatePeerTasks(ctx contractapi.TransactionContextInterface, peerID string, taskType string, clientIDs string) error {
	// In a real implementation, this would distribute tasks among peers for load balancing
	// For simplicity, we'll just record the allocation
	taskID := "TASK_" + peerID + "_" + time.Now().Format("20060102150405")
	
	taskJSON, err := json.Marshal(map[string]string{
		"peerID": peerID,
		"taskType": taskType,
		"clientIDs": clientIDs,
		"status": "Allocated",
		"timestamp": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal task allocation: %v", err)
	}

	err = ctx.GetStub().PutState(taskID, taskJSON)
	if err != nil {
		return fmt.Errorf("failed to put task allocation on ledger: %v", err)
	}

	return nil
}

// ============================================================
// Org2 (TGS - Ticket Granting Server) Functions
// ============================================================

// VerifyTicketGrantingTicket verifies a TGT submitted by a client
func (s *IoTAuthContract) VerifyTicketGrantingTicket(ctx contractapi.TransactionContextInterface, clientID string, tgt string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Check if the TGT matches
	if registration.TGT != tgt {
		return fmt.Errorf("invalid TGT for client: %s", clientID)
	}

	// Check if TGT is expired
	if time.Now().After(registration.ExpiryTime) {
		return fmt.Errorf("TGT has expired for client: %s", clientID)
	}

	// Update status to "Validated by TGS"
	registration.Status = "Validated by TGS"
	
	// Update the registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to update client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "TGTVerification", clientID, "Success", "TGT verified by TGS")

	return nil
}

// IssueServiceTicket issues a service ticket to a client with a valid TGT
func (s *IoTAuthContract) IssueServiceTicket(ctx contractapi.TransactionContextInterface, clientID string, serviceID string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Check if client is validated by TGS
	if registration.Status != "Validated by TGS" {
		return fmt.Errorf("client is not validated by TGS: %s", clientID)
	}

	// Generate a service ticket (in a real system, this would be encrypted)
	serviceTicket := generateServiceTicket(clientID, serviceID, registration.SessionKey)
	registration.ServiceTicket = serviceTicket

	// Update the registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to update client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "ServiceTicketIssue", clientID, "Success", "Service ticket issued to client for service "+serviceID)

	return nil
}

// ForwardToISV forwards a validated registration to the ISV
func (s *IoTAuthContract) ForwardToISV(ctx contractapi.TransactionContextInterface, clientID string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Check if client has a service ticket
	if registration.ServiceTicket == "" {
		return fmt.Errorf("client does not have a service ticket: %s", clientID)
	}

	// Tag the registration as forwarded to ISV
	// In a real system, you might add more data here about which ISV, etc.
	registration.Status = "Forwarded to ISV"
	
	// Update the registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to update client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "ForwardToISV", clientID, "Success", "Registration forwarded to ISV")

	return nil
}

// ============================================================
// Org3 (ISV - IoT Service Validator) Functions
// ============================================================

// RegisterIoTDevice registers a new IoT device in the system
func (s *IoTAuthContract) RegisterIoTDevice(ctx contractapi.TransactionContextInterface, deviceID string, publicKey string, organization string, deviceType string) error {
	// Create device key
	deviceKey := "DEVICE_" + deviceID

	// Check if device already exists
	deviceBytes, err := ctx.GetStub().GetState(deviceKey)
	if err != nil {
		return fmt.Errorf("failed to get device: %v", err)
	}
	if deviceBytes != nil {
		return fmt.Errorf("device already exists: %s", deviceID)
	}

	// Create a new device
	device := IoTDevice{
		DeviceID:     deviceID,
		PublicKey:    publicKey,
		Status:       "Active",
		Organization: organization,
		LastUpdate:   time.Now(),
		DeviceType:   deviceType,
	}

	// Store device on the ledger
	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %v", err)
	}

	err = ctx.GetStub().PutState(deviceKey, deviceJSON)
	if err != nil {
		return fmt.Errorf("failed to put device on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "DeviceRegistration", deviceID, "Success", "IoT device registered successfully")

	return nil
}

// CheckDeviceAvailability checks if an IoT device is available
func (s *IoTAuthContract) CheckDeviceAvailability(ctx contractapi.TransactionContextInterface, deviceID string) (bool, error) {
	// Create device key
	deviceKey := "DEVICE_" + deviceID

	// Get device from ledger
	deviceBytes, err := ctx.GetStub().GetState(deviceKey)
	if err != nil {
		return false, fmt.Errorf("failed to get device: %v", err)
	}
	if deviceBytes == nil {
		return false, fmt.Errorf("device does not exist: %s", deviceID)
	}

	// Unmarshal device
	var device IoTDevice
	err = json.Unmarshal(deviceBytes, &device)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal device: %v", err)
	}

	// Check if device is active
	return device.Status == "Active", nil
}

// VerifyServiceTicket verifies a service ticket submitted by a client
func (s *IoTAuthContract) VerifyServiceTicket(ctx contractapi.TransactionContextInterface, clientID string, serviceTicket string, deviceID string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Check if the service ticket matches
	if registration.ServiceTicket != serviceTicket {
		return fmt.Errorf("invalid service ticket for client: %s", clientID)
	}

	// Check if service ticket is for the requested device
	// In a real system, the service ticket would contain the service/device ID
	// For now, we'll skip this check for simplicity

	// Check if service ticket is expired
	if time.Now().After(registration.ExpiryTime) {
		return fmt.Errorf("service ticket has expired for client: %s", clientID)
	}

	// Check if device is available
	available, err := s.CheckDeviceAvailability(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to check device availability: %v", err)
	}
	if !available {
		return fmt.Errorf("device is not available: %s", deviceID)
	}

	// Update status to "Validated by ISV"
	registration.Status = "Validated by ISV"
	
	// Update the registration on the ledger
	regJSON, err := json.Marshal(registration)
	if err != nil {
		return fmt.Errorf("failed to marshal client registration: %v", err)
	}

	err = ctx.GetStub().PutState(clientID, regJSON)
	if err != nil {
		return fmt.Errorf("failed to update client registration on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "ServiceTicketVerification", clientID, "Success", "Service ticket verified by ISV for device "+deviceID)

	return nil
}

// GrantDeviceAccess grants a client access to an IoT device
func (s *IoTAuthContract) GrantDeviceAccess(ctx contractapi.TransactionContextInterface, clientID string, deviceID string) error {
	// Get client registration from ledger
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client registration: %v", err)
	}
	if regBytes == nil {
		return fmt.Errorf("client does not exist: %s", clientID)
	}

	// Unmarshal client registration
	var registration ClientRegistration
	err = json.Unmarshal(regBytes, &registration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal client registration: %v", err)
	}

	// Check if client is validated by ISV
	if registration.Status != "Validated by ISV" {
		return fmt.Errorf("client is not validated by ISV: %s", clientID)
	}

	// Create an access grant
	accessKey := "ACCESS_" + clientID + "_" + deviceID
	
	accessJSON, err := json.Marshal(map[string]string{
		"clientID": clientID,
		"deviceID": deviceID,
		"status": "Granted",
		"timestamp": time.Now().Format(time.RFC3339),
		"expiryTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339), // Access expires in 1 hour
	})
	if err != nil {
		return fmt.Errorf("failed to marshal access grant: %v", err)
	}

	err = ctx.GetStub().PutState(accessKey, accessJSON)
	if err != nil {
		return fmt.Errorf("failed to put access grant on ledger: %v", err)
	}

	// Record the transaction
	s.recordTransaction(ctx, "AccessGrant", clientID, "Success", "Access granted to device "+deviceID)

	return nil
}

// ============================================================
// Helper Functions
// ============================================================

// ClientExists checks if a client exists
func (s *IoTAuthContract) ClientExists(ctx contractapi.TransactionContextInterface, clientID string) (bool, error) {
	regBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return false, fmt.Errorf("failed to read client registration: %v", err)
	}
	return regBytes != nil, nil
}

// recordTransaction records a transaction on the ledger
func (s *IoTAuthContract) recordTransaction(ctx contractapi.TransactionContextInterface, txType string, clientID string, status string, description string) error {
	// Create a unique transaction ID
	txID := "TX_" + ctx.GetStub().GetTxID()
	
	// Create a transaction record
	transaction := Transaction{
		TxID:        txID,
		ClientID:    clientID,
		Type:        txType,
		Timestamp:   time.Now(),
		Status:      status,
		Description: description,
	}
	
	// Store transaction on the ledger
	txJSON, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}
	
	err = ctx.GetStub().PutState(txID, txJSON)
	if err != nil {
		return fmt.Errorf("failed to put transaction on ledger: %v", err)
	}
	
	return nil
}

// Helper functions to generate values
func generateNonce() string {
	hash := sha256.New()
	hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	return hex.EncodeToString(hash.Sum(nil))
}

func generateSessionKey() string {
	hash := sha256.New()
	hash.Write([]byte("session_" + strconv.FormatInt(time.Now().UnixNano(), 10)))
	return hex.EncodeToString(hash.Sum(nil))
}

func generateTGT(clientID string, publicKey string, sessionKey string) string {
	hash := sha256.New()
	hash.Write([]byte("tgt_" + clientID + "_" + publicKey + "_" + sessionKey + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)))
	return hex.EncodeToString(hash.Sum(nil))
}

func generateServiceTicket(clientID string, serviceID string, sessionKey string) string {
	hash := sha256.New()
	hash.Write([]byte("ticket_" + clientID + "_" + serviceID + "_" + sessionKey + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)))
	return hex.EncodeToString(hash.Sum(nil))
}

func main() {
	chaincode, err := contractapi.NewChaincode(&IoTAuthContract{})
	if err != nil {
		fmt.Printf("Error creating IoT authentication chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting IoT authentication chaincode: %v\n", err)
	}
}
EOF

# Write the go.mod file
echo "Writing go.mod file..."
cat > $CC_SRC_PATH/go.mod << 'EOF'
module github.com/falgunmarothia/iot-auth

go 1.16

require (
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20220131132609-1476cf1d3206
	github.com/hyperledger/fabric-contract-api-go v1.1.1
)
EOF

# Initialize the Go module and download dependencies
echo "Initializing Go module and downloading dependencies..."
cd $CC_SRC_PATH
go mod tidy
go mod download
cd -

# Function to set environment variables for each organization
setOrg1Env() {
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG1_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
}

setOrg2Env() {
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG2_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
}

setOrg3Env() {
  export CORE_PEER_LOCALMSPID="Org3MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ORG3_CA
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
}

# Package the chaincode
echo "Packaging chaincode..."
peer lifecycle chaincode package ${CC_NAME}.tar.gz --path ${CC_SRC_PATH} --lang golang --label ${CC_NAME}_${CC_VERSION}

# Install chaincode on all peers
echo "Installing chaincode on peer0.org1.example.com..."
setOrg1Env
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer1.org1.example.com..."
export CORE_PEER_ADDRESS=peer1.org1.example.com:8051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer2.org1.example.com..."
export CORE_PEER_ADDRESS=peer2.org1.example.com:11051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer0.org2.example.com..."
setOrg2Env
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer1.org2.example.com..."
export CORE_PEER_ADDRESS=peer1.org2.example.com:10051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer2.org2.example.com..."
export CORE_PEER_ADDRESS=peer2.org2.example.com:12051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer0.org3.example.com..."
setOrg3Env
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer1.org3.example.com..."
export CORE_PEER_ADDRESS=peer1.org3.example.com:14051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

echo "Installing chaincode on peer2.org3.example.com..."
export CORE_PEER_ADDRESS=peer2.org3.example.com:15051
peer lifecycle chaincode install ${CC_NAME}.tar.gz

# Get package ID for the installed chaincode
echo "Getting package ID..."
setOrg1Env
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep "${CC_NAME}_${CC_VERSION}" | awk '{print $3}' | cut -d ',' -f 1)
echo "Package ID is ${PACKAGE_ID}"

# Approve chaincode definition for Org1
echo "Approving chaincode definition for Org1..."
setOrg1Env
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --tls --cafile $ORDERER_CA

# Approve chaincode definition for Org2
echo "Approving chaincode definition for Org2..."
setOrg2Env
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --tls --cafile $ORDERER_CA

# Approve chaincode definition for Org3
echo "Approving chaincode definition for Org3..."
setOrg3Env
export CORE_PEER_ADDRESS=peer0.org3.example.com:13051
peer lifecycle chaincode approveformyorg -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --package-id ${PACKAGE_ID} --sequence ${CC_SEQUENCE} --tls --cafile $ORDERER_CA

# Check commit readiness
echo "Checking commit readiness..."
peer lifecycle chaincode checkcommitreadiness --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --sequence ${CC_SEQUENCE} --tls --cafile $ORDERER_CA --output json

# Commit chaincode definition
echo "Committing chaincode definition..."
peer lifecycle chaincode commit -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --channelID $CHANNEL_NAME --name ${CC_NAME} --version ${CC_VERSION} --sequence ${CC_SEQUENCE} --tls --cafile $ORDERER_CA --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA

# Query committed chaincode
echo "Querying committed chaincode..."
peer lifecycle chaincode querycommitted --channelID $CHANNEL_NAME --name ${CC_NAME} --cafile $ORDERER_CA

# Initialize the chaincode
echo "Initializing chaincode..."
peer chaincode invoke -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA -C $CHANNEL_NAME -n ${CC_NAME} --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA --peerAddresses peer0.org3.example.com:13051 --tlsRootCertFiles $PEER0_ORG3_CA -c '{"function":"InitLedger","Args":[]}'

echo "===== Chaincode deployment completed ====="

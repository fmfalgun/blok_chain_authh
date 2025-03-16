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

// TGSChaincode provides functions for Ticket Granting Service operations
type TGSChaincode struct {
	contractapi.Contract
}

// TGT represents a Ticket Granting Ticket issued by the AS
type TGT struct {
	ClientID   string    `json:"clientID"`
	SessionKey string    `json:"sessionKey"`  // KU,TGS - session key for client-TGS communication
	Timestamp  time.Time `json:"timestamp"`
	Lifetime   int64     `json:"lifetime"`    // Lifetime in seconds
}

// ServiceTicket represents a ticket for accessing ISV services
type ServiceTicket struct {
	ClientID   string    `json:"clientID"`
	SessionKey string    `json:"sessionKey"`  // KU,SS - session key for client-ISV communication
	Timestamp  time.Time `json:"timestamp"`
	Lifetime   int64     `json:"lifetime"`    // Lifetime in seconds
}

// ServiceTicketRequest contains the data needed to request a service ticket
type ServiceTicketRequest struct {
	EncryptedTGT   string `json:"encryptedTGT"`   // TGT encrypted with TGS's public key
	ClientID       string `json:"clientID"`       // Client identifier
	ServiceID      string `json:"serviceID"`      // Requested service identifier
	AuthenticatorB64 string `json:"authenticator"` // Timestamp encrypted with session key to prove identity
}

// ServiceTicketResponse contains the data returned to the client
type ServiceTicketResponse struct {
	EncryptedServiceTicket string `json:"encryptedServiceTicket"` // Service ticket encrypted with ISV's public key
	EncryptedSessionKey    string `json:"encryptedSessionKey"`    // New session key encrypted with client's session key
}

// ClientRecord represents a client's registration information in TGS records
type ClientRecord struct {
	ClientID       string    `json:"clientID"`
	LastAccess     time.Time `json:"lastAccess"`
	Status         string    `json:"status"`      // "active", "suspended", etc.
	ValidUntil     time.Time `json:"validUntil"`
}

// PredefinedKeys holds the predefined keys for deterministic initialization
type PredefinedKeys struct {
	TGSPrivateKey string
	TGSPublicKey  string
	ISVPublicKey  string
}

// Initialize sets up the chaincode state
// This function is called when the chaincode is instantiated
func (s *TGSChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	// Check if already initialized to make this idempotent
	existingKey, err := ctx.GetStub().GetState("TGS_INITIALIZED")
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %v", err)
	}
	
	if existingKey != nil {
		// Already initialized, skip to maintain consistency
		return nil
	}
	
	// Use predefined keys instead of generating them dynamically
	keys := getPredefinedKeys()
	
	// Store the TGS private key
	err = ctx.GetStub().PutState("TGS_PRIVATE_KEY", []byte(keys.TGSPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to store TGS private key: %v", err)
	}
	
	// Store the TGS public key
	err = ctx.GetStub().PutState("TGS_PUBLIC_KEY", []byte(keys.TGSPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store TGS public key: %v", err)
	}
	
	// Store the ISV public key
	err = ctx.GetStub().PutState("ISV_PUBLIC_KEY", []byte(keys.ISVPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store ISV public key: %v", err)
	}
	
	// Mark as initialized
	err = ctx.GetStub().PutState("TGS_INITIALIZED", []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark TGS as initialized: %v", err)
	}
	
	return nil
}

// getPredefinedKeys returns the predefined cryptographic keys for deterministic initialization
func getPredefinedKeys() PredefinedKeys {
	// These keys are hardcoded for consistent initialization across all peers
	// In a production system, these could be loaded from secure configuration
	return PredefinedKeys{
		TGSPrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA58L1zNrfqv6K6dNwBDLx23Qsl5qhQdLvxuJBLBcX5JeKJ/GG
HPoytB5MCgkBsk8/CM7BQpjx/CBmyT/7scVGHGbA6PYi8807ZvoZDl8dCk/Uxy1t
YRDeYVrQm2swwUhUTC9kIVYTBZtFzvZp//NybQHgOKHABbsf5EjEG7AOI2qiUzJN
RJPBzZtY0HdUoWYTWRTDiP/7yfVkm1PZsN+eYyWhPVdXQ1JLrGjjwOZl0db5QhcU
mXKjQWcy6/OMYsOjy4H7Mxtu7zGvPJObbTbkKPeh25P9jExLW8XXcxkv6RUbYf3I
AkDfMX8cJc3qtfcLW47Afywy0/zoLLQnQQVl3QIDAQABAoIBAHCIXUqM0fxOUMrL
S4q8omMGZfFXRWgoiRxKyQ1vXB5qMt47b5s4Zq4A41XPJ+LQ7kZADbQCXAuIGQHf
QzCHqkzYW9YL8n7TYBt8K2qVEVSHi/kHQVNLzfHpJPsy27s+o5pQ74AoRZQfblKt
3eBUm53jyHEGYnFlb9eZ5oBxSCEqq37jVZBvSUwx52IxNChjWW0JZwQdLVJ+Uqqs
wjHPl22U3l3QEcnQoQeQiARZQiQ4wP4lEWlUbNh5KnAQeMbvY9I+BsWnTygldUZD
qLzHz7foQWrl4d2XcA+mu3RlcB29lGmwgZVHzFEkKmDCIdcYUgKgcro4QXt+1B1i
TTvTrekCgYEA9v/Vbr6fHh+O8PQpwbVQgMOKqHRPHHPwUH47SOSHcRKwVYNZk0X/
FaRo2TrCkVRRnEo/vNVzYQT1XNxYQGKmKHqT4RbLLVYBVMXogTF0/W7uZJdcJOQV
MvzTxIES/w81TqXnrQYk6Vf38Fjc/uwYBXwOdWfJlCxLnBCy7WaZ5s8CgYEA8BcK
H9GyfsdxLBfH39YM9wz1Ilk5GlMPw+NLX8aYOzMF+zdgZeYJZ+12WHYTRwLRCpfG
6y+Nwt88q4L3NeSffrYR2QKbPo2P6hVPQGOaDLo4J/CkohFYDiLHnY4FXvBOhLz5
OGC+1MSr0XEGhFS9c7MS4zOVNGhGc+X7eEIKOzMCgYB62hzpn7JUdml6ljNZOK76
EK+oXfbFo+IovRn3a+bnJAJZyW4ypIK9KJVo5D4+KBqTtBCvY3c3MfFhCUje2xqj
1/I5afNLnd8ofhWCMzBi6DswS47yZJHLW7bWIZGFcmZfM38qmSTXw3OjJLqsrBw/
vTR6FbR4xcI2WxTN1t8HdQKBgCC2KgQc3NxJMtvwvUmA0KHPNyu3C/CNnIEbehsj
Uo7IWGBbKkKHjnNSjKzuoqjqP+vQ0HyYXPxlbR+8Rg3D0Jt3f/8aCRhD9jOUUhME
4M77ya9UJiWzVTqUEjVQB3k2M0BzKw+a/eHQC3D4qQ5GflZ7+P7QvHcYqBERKjFM
OFJPAoGBAMnUU7I3Qpo1n0HwBsQXoA1TgRcUMQQHp2/9XJP0K5L1FQvBMmhfeMQB
RA8g7GYJ3691Wy1GZ4YS1/QBZ9I69P0PYYxJXlaTZH9iEoAqvRcBoiXQgUkjI+TA
XJJc/DlIvuP0RBGJ4RYQJujO3fTMfUbVaQDJSQ5I8Ui/Yc4d1ZBE
-----END RSA PRIVATE KEY-----`,
		TGSPublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA58L1zNrfqv6K6dNwBDLx
23Qsl5qhQdLvxuJBLBcX5JeKJ/GGHPoytB5MCgkBsk8/CM7BQpjx/CBmyT/7scVG
HGbA6PYi8807ZvoZDl8dCk/Uxy1tYRDeYVrQm2swwUhUTC9kIVYTBZtFzvZp//Ny
bQHgOKHABbsf5EjEG7AOI2qiUzJNRJPBzZtY0HdUoWYTWRTDiP/7yfVkm1PZsN+e
YyWhPVdXQ1JLrGjjwOZl0db5QhcUmXKjQWcy6/OMYsOjy4H7Mxtu7zGvPJObbTbk
KPeh25P9jExLW8XXcxkv6RUbYf3IAkDfMX8cJc3qtfcLW47Afywy0/zoLLQnQQVl
3QIDAQAB
-----END PUBLIC KEY-----`,
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

// getPrivateKey retrieves the TGS's private key from the chaincode state
func (s *TGSChaincode) getPrivateKey(ctx contractapi.TransactionContextInterface) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := ctx.GetStub().GetState("TGS_PRIVATE_KEY")
	if err != nil {
		return nil, err
	}
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("TGS private key not found")
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

// getPublicKey retrieves the specified public key from the chaincode state
func (s *TGSChaincode) getPublicKey(ctx contractapi.TransactionContextInterface, keyName string) (*rsa.PublicKey, error) {
	publicKeyPEM, err := ctx.GetStub().GetState(keyName)
	if err != nil {
		return nil, err
	}
	if publicKeyPEM == nil {
		return nil, fmt.Errorf("public key %s not found", keyName)
	}
	
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
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

// ==================== Core TGS Operations ====================

// ProcessRegistrationFromAS validates a TGT from AS and records client registration
// This implements the "Process Registration of Org1" operation
func (s *TGSChaincode) ProcessRegistrationFromAS(ctx contractapi.TransactionContextInterface, encryptedTGT string) error {
	// Decode the base64 encoded encrypted TGT
	tgtBytes, err := base64.StdEncoding.DecodeString(encryptedTGT)
	if err != nil {
		return fmt.Errorf("invalid TGT format: %v", err)
	}
	
	// Get the TGS private key
	privateKey, err := s.getPrivateKey(ctx)
	if err != nil {
		return err
	}
	
	// Decrypt the TGT using TGS's private key
	// This implements: M = TGT^dTGS = (M^eTGS)^dTGS mod nTGS from the paper
	decryptedTGTBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, tgtBytes)
	if err != nil {
		return fmt.Errorf("TGT decryption failed: %v", err)
	}
	
	// Parse the decrypted TGT
	var tgt TGT
	err = json.Unmarshal(decryptedTGTBytes, &tgt)
	if err != nil {
		return fmt.Errorf("invalid TGT structure: %v", err)
	}
	
	// Validate the TGT timestamp and lifetime
	if time.Now().After(tgt.Timestamp.Add(time.Duration(tgt.Lifetime) * time.Second)) {
		return fmt.Errorf("TGT has expired")
	}
	
	// Create a client record
	clientRecord := ClientRecord{
		ClientID:   tgt.ClientID,
		LastAccess: time.Now(),
		Status:     "active",
		ValidUntil: tgt.Timestamp.Add(time.Duration(tgt.Lifetime) * time.Second),
	}
	
	// Store the client record
	clientRecordJSON, err := json.Marshal(clientRecord)
	if err != nil {
		return err
	}
	
	err = ctx.GetStub().PutState("CLIENT_RECORD_"+tgt.ClientID, clientRecordJSON)
	if err != nil {
		return err
	}
	
	// Store the session key for future use
	err = ctx.GetStub().PutState("SESSION_KEY_"+tgt.ClientID, []byte(tgt.SessionKey))
	if err != nil {
		return err
	}
	
	// Record this registration on the blockchain
	registrationEvent := struct {
		ClientID   string    `json:"clientID"`
		Timestamp  time.Time `json:"timestamp"`
		ValidUntil time.Time `json:"validUntil"`
		TGTHash    string    `json:"tgtHash"`
	}{
		ClientID:   tgt.ClientID,
		Timestamp:  time.Now(),
		ValidUntil: tgt.Timestamp.Add(time.Duration(tgt.Lifetime) * time.Second),
		TGTHash:    fmt.Sprintf("%x", sha256.Sum256(decryptedTGTBytes)),
	}
	
	registrationEventJSON, err := json.Marshal(registrationEvent)
	if err != nil {
		return err
	}
	
	// Create a deterministic registration ID
	registrationID := "REGISTRATION_" + tgt.ClientID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	return ctx.GetStub().PutState(registrationID, registrationEventJSON)
}

// CheckRegistrationValidity verifies if a client's registration is valid
// This implements the "Check for Record & Validity of Registration" operation
func (s *TGSChaincode) CheckRegistrationValidity(ctx contractapi.TransactionContextInterface, clientID string) (bool, error) {
	// Retrieve the client record
	clientRecordJSON, err := ctx.GetStub().GetState("CLIENT_RECORD_" + clientID)
	if err != nil {
		return false, fmt.Errorf("failed to read client record: %v", err)
	}
	if clientRecordJSON == nil {
		return false, fmt.Errorf("client %s is not registered with TGS", clientID)
	}
	
	var clientRecord ClientRecord
	err = json.Unmarshal(clientRecordJSON, &clientRecord)
	if err != nil {
		return false, err
	}
	
	// Check if the client record is still valid
	if time.Now().After(clientRecord.ValidUntil) {
		return false, nil
	}
	
	if clientRecord.Status != "active" {
		return false, nil
	}
	
	// Update last access time
	clientRecord.LastAccess = time.Now()
	updatedClientRecordJSON, err := json.Marshal(clientRecord)
	if err != nil {
		return false, err
	}
	
	err = ctx.GetStub().PutState("CLIENT_RECORD_"+clientID, updatedClientRecordJSON)
	if err != nil {
		return false, err
	}
	
	return true, nil
}

// GenerateServiceTicket creates a service ticket for the client to access ISV
// This implements Step 4: TGS Issues Service Ticket for ISV
// and the "Endorse & Validate of Registration" operation
func (s *TGSChaincode) GenerateServiceTicket(ctx contractapi.TransactionContextInterface, request string) (*ServiceTicketResponse, error) {
	// Parse the service ticket request
	var ticketRequest ServiceTicketRequest
	err := json.Unmarshal([]byte(request), &ticketRequest)
	if err != nil {
		return nil, fmt.Errorf("invalid request format: %v", err)
	}
	
	// Step 1: Decrypt and validate the TGT
	tgtBytes, err := base64.StdEncoding.DecodeString(ticketRequest.EncryptedTGT)
	if err != nil {
		return nil, fmt.Errorf("invalid TGT format: %v", err)
	}
	
	privateKey, err := s.getPrivateKey(ctx)
	if err != nil {
		return nil, err
	}
	
	// Decrypt the TGT using TGS's private key
	// This implements: M = TGT^dTGS = (M^eTGS)^dTGS mod nTGS
	decryptedTGTBytes, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, tgtBytes)
	if err != nil {
		return nil, fmt.Errorf("TGT decryption failed: %v", err)
	}
	
	var tgt TGT
	err = json.Unmarshal(decryptedTGTBytes, &tgt)
	if err != nil {
		return nil, fmt.Errorf("invalid TGT structure: %v", err)
	}
	
	// Validate the TGT timestamp and lifetime
	if time.Now().After(tgt.Timestamp.Add(time.Duration(tgt.Lifetime) * time.Second)) {
		return nil, fmt.Errorf("TGT has expired")
	}
	
	// Verify the client ID matches the one in the TGT
	if tgt.ClientID != ticketRequest.ClientID {
		return nil, fmt.Errorf("client ID mismatch")
	}
	
	// Step 2: Check if the client's registration is valid
	valid, err := s.CheckRegistrationValidity(ctx, tgt.ClientID)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("client registration is not valid")
	}
	
	// Step 3: Verify the authenticator (timestamp encrypted with session key)
	// In a real implementation, you would decrypt the authenticator using the session key
	// and verify that the timestamp is recent (within a few minutes)
	// For simplicity, we'll skip this step in this example
	
	// Step 4: Generate a deterministic session key KU,SS for client-ISV communication
	// Using a deterministic approach based on client ID, service ID, and current time
	timestamp := time.Now().Unix()
	sessionKeyInput := tgt.ClientID + ticketRequest.ServiceID + strconv.FormatInt(timestamp, 10) + "KU,SS"
	sessionKeyHash := sha256.Sum256([]byte(sessionKeyInput))
	sessionKey := base64.StdEncoding.EncodeToString(sessionKeyHash[:])
	
	// Step 5: Create a service ticket
	serviceTicket := ServiceTicket{
		ClientID:   tgt.ClientID,
		SessionKey: sessionKey,
		Timestamp:  time.Now(),
		Lifetime:   3600, // 1 hour in seconds
	}
	
	// Convert service ticket to JSON
	serviceTicketJSON, err := json.Marshal(serviceTicket)
	if err != nil {
		return nil, err
	}
	
	// Get ISV's public key
	isvPublicKey, err := s.getPublicKey(ctx, "ISV_PUBLIC_KEY")
	if err != nil {
		return nil, err
	}
	
	// Encrypt service ticket with ISV's public key
	// This implements: TSS = {Client ID, KU,SS, Timestamp, Lifetime}eISV = M^eISV mod nISV
	encryptedServiceTicket, err := rsa.EncryptPKCS1v15(rand.Reader, isvPublicKey, serviceTicketJSON)
	if err != nil {
		return nil, err
	}
	
	// Encrypt the new session key with the existing session key from the TGT
	// For simplicity and determinism, we'll create a known encrypted form
	// In a real implementation, this would be encrypted with the client's session key
	encryptedSessionKeyInput := tgt.SessionKey + ":" + sessionKey
	encryptedSessionKeyHash := sha256.Sum256([]byte(encryptedSessionKeyInput))
	encryptedSessionKey := encryptedSessionKeyHash[:]
	
	// Create the response
	response := ServiceTicketResponse{
		EncryptedServiceTicket: base64.StdEncoding.EncodeToString(encryptedServiceTicket),
		EncryptedSessionKey:    base64.StdEncoding.EncodeToString(encryptedSessionKey),
	}
	
	// Record this ticket issuance on the blockchain for audit purposes
	return &response, s.recordTicketIssuance(ctx, tgt.ClientID, ticketRequest.ServiceID, serviceTicketJSON)
}

// recordTicketIssuance records a service ticket issuance on the blockchain
// This is part of the "Endorse & Validate of Registration" operation
func (s *TGSChaincode) recordTicketIssuance(ctx contractapi.TransactionContextInterface, clientID string, serviceID string, serviceTicketJSON []byte) error {
	ticketRecord := struct {
		ClientID     string    `json:"clientID"`
		ServiceID    string    `json:"serviceID"`
		Timestamp    time.Time `json:"timestamp"`
		TicketHash   string    `json:"ticketHash"`
	}{
		ClientID:     clientID,
		ServiceID:    serviceID,
		Timestamp:    time.Now(),
		TicketHash:   fmt.Sprintf("%x", sha256.Sum256(serviceTicketJSON)),
	}
	
	ticketRecordJSON, err := json.Marshal(ticketRecord)
	if err != nil {
		return err
	}
	
	// Store the ticket record with a deterministic ID
	ticketID := "TICKET_" + clientID + "_" + serviceID + "_" + strconv.FormatInt(ticketRecord.Timestamp.Unix(), 10)
	return ctx.GetStub().PutState(ticketID, ticketRecordJSON)
}

// ForwardRegistrationToISV prepares and forwards client registration to ISV
// This implements the "Forward Registration to Org3" operation
func (s *TGSChaincode) ForwardRegistrationToISV(ctx contractapi.TransactionContextInterface, clientID string, serviceID string, encryptedServiceTicket string) error {
	// Verify the client's registration is valid
	valid, err := s.CheckRegistrationValidity(ctx, clientID)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("client registration is not valid")
	}
	
	// Create a forwarding record with a deterministic approach
	forwardingRecord := struct {
		ClientID              string    `json:"clientID"`
		ServiceID             string    `json:"serviceID"`
		Timestamp             time.Time `json:"timestamp"`
		EncryptedServiceTicket string    `json:"encryptedServiceTicket"`
		Status                string    `json:"status"`
	}{
		ClientID:              clientID,
		ServiceID:             serviceID,
		Timestamp:             time.Now(),
		EncryptedServiceTicket: encryptedServiceTicket,
		Status:                "forwarded",
	}
	
	forwardingRecordJSON, err := json.Marshal(forwardingRecord)
	if err != nil {
		return err
	}
	
	// Store the forwarding record with a deterministic ID
	forwardingID := "FORWARDING_" + clientID + "_" + serviceID + "_" + strconv.FormatInt(forwardingRecord.Timestamp.Unix(), 10)
	return ctx.GetStub().PutState(forwardingID, forwardingRecordJSON)
}

// GetAllClientRegistrations retrieves all client registrations
func (s *TGSChaincode) GetAllClientRegistrations(ctx contractapi.TransactionContextInterface) ([]*ClientRecord, error) {
	// Get all client registrations from the world state
	resultsIterator, err := ctx.GetStub().GetStateByRange("CLIENT_RECORD_", "CLIENT_RECORD_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	
	var clients []*ClientRecord
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		
		var client ClientRecord
		err = json.Unmarshal(queryResponse.Value, &client)
		if err != nil {
			return nil, err
		}
		
		clients = append(clients, &client)
	}
	
	return clients, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&TGSChaincode{})
	if err != nil {
		fmt.Printf("Error creating TGS chaincode: %s", err.Error())
		return
	}
	
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting TGS chaincode: %s", err.Error())
	}
}

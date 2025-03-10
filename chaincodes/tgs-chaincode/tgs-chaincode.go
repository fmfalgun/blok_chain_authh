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

// Initialize sets up the chaincode state
// This function is called when the chaincode is instantiated
func (s *TGSChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	// Initialize the TGS server's own RSA key pair
	err := s.generateAndStoreTGSKeyPair(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize TGS key pair: %v", err)
	}
	
	// Register the ISV public key (in a real system, this would be fetched from the ISV)
	// For demonstration, we'll generate it here
	err = s.generateAndStoreISVPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize ISV public key: %v", err)
	}
	
	return nil
}

// ==================== Helper Functions ====================

// generateAndStoreTGSKeyPair creates and stores the TGS's RSA key pair
// This implements the RSA key generation as described in the paper section 3.2
func (s *TGSChaincode) generateAndStoreTGSKeyPair(ctx contractapi.TransactionContextInterface) error {
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
	err = ctx.GetStub().PutState("TGS_PRIVATE_KEY", privateKeyPEM)
	if err != nil {
		return err
	}
	
	// The public key is also stored on the blockchain as described in the paper
	// This allows for transparent verification by all participants
	err = ctx.GetStub().PutState("TGS_PUBLIC_KEY", publicKeyPEM)
	if err != nil {
		return err
	}
	
	return nil
}

// generateAndStoreISVPublicKey creates and stores a sample ISV public key
// In a real system, this would be obtained from the ISV's blockchain record
func (s *TGSChaincode) generateAndStoreISVPublicKey(ctx contractapi.TransactionContextInterface) error {
	// This is a placeholder - in a real system, this would be fetched
	// from the ISV's blockchain registration
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	
	// Store the ISV public key
	err = ctx.GetStub().PutState("ISV_PUBLIC_KEY", publicKeyPEM)
	if err != nil {
		return err
	}
	
	return nil
}

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
	
	// Create a unique registration ID
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
	
	// Step 4: Generate a new session key KU,SS for client-ISV communication
	sessionKeyBytes := make([]byte, 32)
	_, err = rand.Read(sessionKeyBytes)
	if err != nil {
		return nil, err
	}
	sessionKey := base64.StdEncoding.EncodeToString(sessionKeyBytes)
	
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
	// In a real implementation, you would use the session key KU,TGS for encryption
	// For simplicity, we'll just encrypt it with the client's public key
	// Get client's public key (in a real system this would be stored or retrieved differently)
	clientPublicKey, err := s.getPublicKey(ctx, "CLIENT_PK_"+tgt.ClientID)
	if err != nil {
		// If we can't find the client's public key, we'll use a simpler approach
		// Just store the session key and return a reference
		err = ctx.GetStub().PutState("NEW_SESSION_KEY_"+tgt.ClientID, []byte(sessionKey))
		if err != nil {
			return nil, err
		}
		encryptedSessionKey := []byte("KEY_REF_" + tgt.ClientID)
		
		// Create the response
		response := ServiceTicketResponse{
			EncryptedServiceTicket: base64.StdEncoding.EncodeToString(encryptedServiceTicket),
			EncryptedSessionKey:    base64.StdEncoding.EncodeToString(encryptedSessionKey),
		}
		
		// Record this ticket issuance on the blockchain for audit purposes
		return &response, s.recordTicketIssuance(ctx, tgt.ClientID, ticketRequest.ServiceID, serviceTicketJSON)
	}
	
	// If we have the client's public key, encrypt the session key with it
	// This implements: {KU,SS}eU = KU,SS^eU mod nU
	encryptedSessionKey, err := rsa.EncryptPKCS1v15(rand.Reader, clientPublicKey, []byte(sessionKey))
	if err != nil {
		return nil, err
	}
	
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
	
	// Store the ticket record in the world state
	ticketID := "TICKET_" + clientID + "_" + serviceID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
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
	
	// Create a forwarding record
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
	
	// Store the forwarding record
	forwardingID := "FORWARDING_" + clientID + "_" + serviceID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	return ctx.GetStub().PutState(forwardingID, forwardingRecordJSON)
	
	// In a real system, this function would also communicate with the ISV chaincode
	// to notify it of the forwarded registration, possibly through events or direct chaincode-to-chaincode calls
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

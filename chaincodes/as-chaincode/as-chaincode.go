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
	//"math/big"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ASChaincode provides functions for Authentication Server operations
type ASChaincode struct {
	contractapi.Contract
}

// ClientIdentity represents a client's registration information
type ClientIdentity struct {
	ID              string    `json:"id"`
	PublicKey       string    `json:"publicKey"`
	RegistrationTime time.Time `json:"registrationTime"`
	Valid           bool      `json:"valid"`
	Nonce           string    `json:"nonce,omitempty"`  // Used during authentication process
}

// TGT represents a Ticket Granting Ticket
type TGT struct {
	ClientID   string    `json:"clientID"`
	SessionKey string    `json:"sessionKey"`  // KU,TGS - session key for client-TGS communication
	Timestamp  time.Time `json:"timestamp"`
	Lifetime   int64     `json:"lifetime"`    // Lifetime in seconds
}

// ResponseToClient contains the TGT and the encrypted session key for the client
type ResponseToClient struct {
	EncryptedTGT          string `json:"encryptedTGT"`          // TGT encrypted with TGS's public key
	EncryptedSessionKey   string `json:"encryptedSessionKey"`   // Session key encrypted with client's public key
}

// NonceChallenge represents a challenge sent to the client for authentication
type NonceChallenge struct {
	Nonce          string `json:"nonce"`
	ExpirationTime int64  `json:"expirationTime"` // Unix timestamp
}

// Initialize sets up the chaincode state
// This function is called when the chaincode is instantiated
func (s *ASChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	// Initialize the AS server's own RSA key pair
	err := s.generateAndStoreASKeyPair(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize AS key pair: %v", err)
	}
	
	// Register the TGS public key (in a real system, this would be fetched from the TGS)
	// For demonstration, we'll generate it here
	err = s.generateAndStoreTGSPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize TGS public key: %v", err)
	}
	
	return nil
}

// ==================== Helper Functions ====================

// generateAndStoreASKeyPair creates and stores the AS's RSA key pair
// This implements the RSA key generation as described in the paper section 3.2
func (s *ASChaincode) generateAndStoreASKeyPair(ctx contractapi.TransactionContextInterface) error {
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
	err = ctx.GetStub().PutState("AS_PRIVATE_KEY", privateKeyPEM)
	if err != nil {
		return err
	}
	
	// The public key is also stored on the blockchain as described in the paper
	// This allows for transparent verification by all participants
	err = ctx.GetStub().PutState("AS_PUBLIC_KEY", publicKeyPEM)
	if err != nil {
		return err
	}
	
	return nil
}

// generateAndStoreTGSPublicKey creates and stores a sample TGS public key
// In a real system, this would be obtained from the TGS's blockchain record
func (s *ASChaincode) generateAndStoreTGSPublicKey(ctx contractapi.TransactionContextInterface) error {
	// This is a placeholder - in a real system, this would be fetched
	// from the TGS's blockchain registration
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
	
	// Store the TGS public key
	err = ctx.GetStub().PutState("TGS_PUBLIC_KEY", publicKeyPEM)
	if err != nil {
		return err
	}
	
	return nil
}

// getPrivateKey retrieves the AS's private key from the chaincode state
func (s *ASChaincode) getPrivateKey(ctx contractapi.TransactionContextInterface) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := ctx.GetStub().GetState("AS_PRIVATE_KEY")
	if err != nil {
		return nil, err
	}
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("AS private key not found")
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
func (s *ASChaincode) getPublicKey(ctx contractapi.TransactionContextInterface, keyName string) (*rsa.PublicKey, error) {
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

// ==================== Core AS Operations ====================

// RegisterClient registers a new client with the AS
// This performs the initial client registration before authentication
func (s *ASChaincode) RegisterClient(ctx contractapi.TransactionContextInterface, clientID string, clientPublicKeyPEM string) error {
	// Check if client already exists
	existingClientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existingClientJSON != nil {
		return fmt.Errorf("client %s already exists", clientID)
	}
	
	// Verify the provided public key is valid
	block, _ := pem.Decode([]byte(clientPublicKeyPEM))
	if block == nil {
		return fmt.Errorf("failed to decode PEM block containing public key")
	}
	
	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("invalid public key: %v", err)
	}
	
	// Create and store the client record
	client := ClientIdentity{
		ID:              clientID,
		PublicKey:       clientPublicKeyPEM,
		RegistrationTime: time.Now(),
		Valid:           true,
	}
	
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return err
	}
	
	// Store client data in the world state
	err = ctx.GetStub().PutState("CLIENT_"+clientID, clientJSON)
	if err != nil {
		return err
	}
	
	// Store the client's public key separately for easy access
	err = ctx.GetStub().PutState("CLIENT_PK_"+clientID, []byte(clientPublicKeyPEM))
	if err != nil {
		return err
	}
	
	return nil
}

// CheckClientValidity verifies if a client is valid
// This checks the client's registration status
func (s *ASChaincode) CheckClientValidity(ctx contractapi.TransactionContextInterface, clientID string) (bool, error) {
	clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
	if err != nil {
		return false, fmt.Errorf("failed to read client data: %v", err)
	}
	if clientJSON == nil {
		return false, fmt.Errorf("client %s does not exist", clientID)
	}
	
	var client ClientIdentity
	err = json.Unmarshal(clientJSON, &client)
	if err != nil {
		return false, err
	}
	
	// Check if the client is valid and not expired
	// In a real implementation, you might check against revocation lists
	// or apply additional validation rules
	return client.Valid, nil
}

// InitiateAuthentication generates a nonce challenge for client authentication
// This is the first step in the authentication process as described in the paper
// Step 1: Client Requests Authentication from AS
func (s *ASChaincode) InitiateAuthentication(ctx contractapi.TransactionContextInterface, clientID string) (*NonceChallenge, error) {
	// Check if client exists and is valid
	valid, err := s.CheckClientValidity(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid client")
	}
	
	// Generate a random nonce (NU in the paper)
	// This will be used for the challenge-response authentication
	nonceBytes := make([]byte, 32)
	_, err = rand.Read(nonceBytes)
	if err != nil {
		return nil, err
	}
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)
	
	// Set expiration time for the nonce (e.g., 5 minutes from now)
	expirationTime := time.Now().Add(5 * time.Minute).Unix()
	
	// Create the challenge
	challenge := NonceChallenge{
		Nonce:          nonce,
		ExpirationTime: expirationTime,
	}
	
	// Store the nonce with the client record for later verification
	clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
	if err != nil {
		return nil, err
	}
	
	var client ClientIdentity
	err = json.Unmarshal(clientJSON, &client)
	if err != nil {
		return nil, err
	}
	
	client.Nonce = nonce
	updatedClientJSON, err := json.Marshal(client)
	if err != nil {
		return nil, err
	}
	
	err = ctx.GetStub().PutState("CLIENT_"+clientID, updatedClientJSON)
	if err != nil {
		return nil, err
	}
	
	return &challenge, nil
}

// VerifyClientIdentity verifies a client's response to the nonce challenge
// This implements the client authentication verification from the paper
// Step 3: AS decrypts the nonce using its private key to verify client identity
func (s *ASChaincode) VerifyClientIdentity(ctx contractapi.TransactionContextInterface, clientID string, encryptedNonce string) (bool, error) {
	// Retrieve the client record to get the expected nonce
	clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
	if err != nil {
		return false, err
	}
	if clientJSON == nil {
		return false, fmt.Errorf("client %s does not exist", clientID)
	}
	
	var client ClientIdentity
	err = json.Unmarshal(clientJSON, &client)
	if err != nil {
		return false, err
	}
	
	if client.Nonce == "" {
		return false, fmt.Errorf("no authentication challenge found for client")
	}
	
	// Get the AS private key to decrypt the client's response
	privateKey, err := s.getPrivateKey(ctx)
	if err != nil {
		return false, err
	}
	
	// Decode the base64 encoded encrypted nonce
	encryptedNonceBytes, err := base64.StdEncoding.DecodeString(encryptedNonce)
	if err != nil {
		return false, err
	}
	
	// Decrypt the nonce using AS's private key
	// This implements: NU = CU^dAS = (NU^eAS)^dAS mod nAS from the paper
	decryptedNonce, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedNonceBytes)
	if err != nil {
		return false, fmt.Errorf("decryption failed: %v", err)
	}
	
	// Convert decrypted nonce to base64 for comparison
	decryptedNonceB64 := base64.StdEncoding.EncodeToString(decryptedNonce)
	
	// Compare the decrypted nonce with the expected nonce
	// This verifies that the client correctly encrypted the nonce with AS's public key
	if decryptedNonceB64 != client.Nonce {
		return false, nil
	}
	
	// Clear the nonce from the client record as it's been used
	client.Nonce = ""
	updatedClientJSON, err := json.Marshal(client)
	if err != nil {
		return false, err
	}
	
	err = ctx.GetStub().PutState("CLIENT_"+clientID, updatedClientJSON)
	if err != nil {
		return false, err
	}
	
	return true, nil
}

// GenerateTGT generates a Ticket Granting Ticket (TGT) for a client
// This implements Step 2: AS Issues TGT Encrypted with TGS's Public Key
func (s *ASChaincode) GenerateTGT(ctx contractapi.TransactionContextInterface, clientID string) (*ResponseToClient, error) {
	// Verify that client exists and is valid
	valid, err := s.CheckClientValidity(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid client")
	}
	
	// Generate a session key KU,TGS for client-TGS communication
	sessionKeyBytes := make([]byte, 32)
	_, err = rand.Read(sessionKeyBytes)
	if err != nil {
		return nil, err
	}
	sessionKey := base64.StdEncoding.EncodeToString(sessionKeyBytes)
	
	// Create the TGT
	tgt := TGT{
		ClientID:   clientID,
		SessionKey: sessionKey,
		Timestamp:  time.Now(),
		Lifetime:   3600, // 1 hour in seconds
	}
	
	// Convert TGT to JSON
	tgtJSON, err := json.Marshal(tgt)
	if err != nil {
		return nil, err
	}
	
	// Get TGS's public key
	tgsPublicKey, err := s.getPublicKey(ctx, "TGS_PUBLIC_KEY")
	if err != nil {
		return nil, err
	}
	
	// Encrypt TGT with TGS's public key
	// This implements: TGT = {Client ID, KU,TGS, Timestamp, Lifetime}eTGS = M^eTGS mod nTGS
	encryptedTGT, err := rsa.EncryptPKCS1v15(rand.Reader, tgsPublicKey, tgtJSON)
	if err != nil {
		return nil, err
	}
	
	// Get client's public key
	clientPublicKeyPEM, err := ctx.GetStub().GetState("CLIENT_PK_" + clientID)
	if err != nil {
		return nil, err
	}
	if clientPublicKeyPEM == nil {
		return nil, fmt.Errorf("client public key not found")
	}
	
	block, _ := pem.Decode(clientPublicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing client public key")
	}
	
	clientPublicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	
	clientPublicKey, ok := clientPublicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	// Encrypt the session key with client's public key
	// This implements: {KU,TGS}eU = KU,TGS^eU mod nU
	encryptedSessionKey, err := rsa.EncryptPKCS1v15(rand.Reader, clientPublicKey, []byte(sessionKey))
	if err != nil {
		return nil, err
	}
	
	// Create the response for the client
	response := ResponseToClient{
		EncryptedTGT:        base64.StdEncoding.EncodeToString(encryptedTGT),
		EncryptedSessionKey: base64.StdEncoding.EncodeToString(encryptedSessionKey),
	}
	
	// Record this TGT issuance on the ledger for audit purposes
	tgtRecord := struct {
		ClientID  string    `json:"clientID"`
		Timestamp time.Time `json:"timestamp"`
		TGTHash   string    `json:"tgtHash"`
	}{
		ClientID:  clientID,
		Timestamp: time.Now(),
		TGTHash:   fmt.Sprintf("%x", sha256.Sum256(tgtJSON)),
	}
	
	tgtRecordJSON, err := json.Marshal(tgtRecord)
	if err != nil {
		return nil, err
	}
	
	// Store the TGT record in the world state
	tgtID := "TGT_" + clientID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	err = ctx.GetStub().PutState(tgtID, tgtRecordJSON)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

// GetAllClientRegistrations retrieves all client registrations
// This implements the operation to get all registrations from clients
func (s *ASChaincode) GetAllClientRegistrations(ctx contractapi.TransactionContextInterface) ([]*ClientIdentity, error) {
	// Get all client registrations from the world state
	resultsIterator, err := ctx.GetStub().GetStateByRange("CLIENT_", "CLIENT_~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	
	var clients []*ClientIdentity
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		
		var client ClientIdentity
		err = json.Unmarshal(queryResponse.Value, &client)
		if err != nil {
			return nil, err
		}
		
		// Remove sensitive nonce data if present
		client.Nonce = ""
		clients = append(clients, &client)
	}
	
	return clients, nil
}

// AllocatePeerTask assigns a task to a specific peer
// This implements task allocation for efficient processing
func (s *ASChaincode) AllocatePeerTask(ctx contractapi.TransactionContextInterface, peerID string, taskType string, clientID string) error {
	// Create a task record
	task := struct {
		PeerID      string    `json:"peerID"`
		TaskType    string    `json:"taskType"`
		ClientID    string    `json:"clientID"`
		AssignedAt  time.Time `json:"assignedAt"`
		Status      string    `json:"status"`
	}{
		PeerID:      peerID,
		TaskType:    taskType,
		ClientID:    clientID,
		AssignedAt:  time.Now(),
		Status:      "assigned",
	}
	
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}
	
	// Store the task in the world state
	taskID := "TASK_" + peerID + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	return ctx.GetStub().PutState(taskID, taskJSON)
}

// ReserveAndValidateRegistration finalizes a client registration
// This is used for reserving and validating client registrations
func (s *ASChaincode) ReserveAndValidateRegistration(ctx contractapi.TransactionContextInterface, clientID string) error {
	// Retrieve the client record
	clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
	if err != nil {
		return err
	}
	if clientJSON == nil {
		return fmt.Errorf("client %s does not exist", clientID)
	}
	
	var client ClientIdentity
	err = json.Unmarshal(clientJSON, &client)
	if err != nil {
		return err
	}
	
	// Mark the client as valid (this would include more validation in a real system)
	client.Valid = true
	
	// Update the client record
	updatedClientJSON, err := json.Marshal(client)
	if err != nil {
		return err
	}
	
	return ctx.GetStub().PutState("CLIENT_"+clientID, updatedClientJSON)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&ASChaincode{})
	if err != nil {
		fmt.Printf("Error creating AS chaincode: %s", err.Error())
		return
	}
	
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting AS chaincode: %s", err.Error())
	}
}

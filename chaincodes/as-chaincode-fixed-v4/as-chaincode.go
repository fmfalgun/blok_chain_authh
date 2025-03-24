package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"
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
	// Nonce field removed - now stored separately
}

// AuthChallenge represents an authentication challenge for a client
type AuthChallenge struct {
	ClientID       string    `json:"clientID"`
	Nonce          string    `json:"nonce"`
	ExpirationTime int64     `json:"expirationTime"`
	CreatedAt      time.Time `json:"createdAt"`
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

// PredefinedKeys holds the predefined keys for deterministic initialization
type PredefinedKeys struct {
	ASPrivateKey string
	ASPublicKey  string
	TGSPublicKey string
}

// Helper function for string truncation in logs
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
func (s *ASChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	// Check if already initialized to make this idempotent
	existingKey, err := ctx.GetStub().GetState("AS_INITIALIZED")
	if err != nil {
		return fmt.Errorf("failed to check initialization status: %v", err)
	}
	
	if existingKey != nil {
		// Already initialized, skip to maintain consistency
		fmt.Println("AS chaincode already initialized, skipping initialization")
		return nil
	}
	
	// Use predefined keys instead of generating them dynamically
	// This ensures all peers have the same keys
	keys := getPredefinedKeys()
	
	// Log the keys being used (truncated for security)
	fmt.Printf("AS private key (first 50 chars): %s...\n", 
		keys.ASPrivateKey[:min(50, len(keys.ASPrivateKey))])
	fmt.Printf("TGS public key (first 50 chars): %s...\n", 
		keys.TGSPublicKey[:min(50, len(keys.TGSPublicKey))])
	
	// Store the AS private key
	err = ctx.GetStub().PutState("AS_PRIVATE_KEY", []byte(keys.ASPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to store AS private key: %v", err)
	}
	
	// Store the AS public key
	err = ctx.GetStub().PutState("AS_PUBLIC_KEY", []byte(keys.ASPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store AS public key: %v", err)
	}
	
	// Store the TGS public key
	err = ctx.GetStub().PutState("TGS_PUBLIC_KEY", []byte(keys.TGSPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store TGS public key: %v", err)
	}
	
	// Mark as initialized
	err = ctx.GetStub().PutState("AS_INITIALIZED", []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark AS as initialized: %v", err)
	}
	
	// Verify key storage
	verifyKey, err := ctx.GetStub().GetState("AS_PRIVATE_KEY")
	if err != nil {
		return fmt.Errorf("failed to verify key storage: %v", err)
	}
	if verifyKey == nil {
		return fmt.Errorf("verification failed: AS private key not stored correctly")
	}
	
	fmt.Println("AS chaincode successfully initialized")
	return nil
}

// getPredefinedKeys returns the predefined cryptographic keys for deterministic initialization
func getPredefinedKeys() PredefinedKeys {
	// These keys are hardcoded for consistent initialization across all peers
	// In a production system, these could be loaded from secure configuration
	return PredefinedKeys{
		ASPrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAtOL3THYTwCk35h9/BYpX/5pQGH4jK5nyO55oI8PqBMx6GHfn
P0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4YTNNmV/IQKqX1eUtRJSP/gZR5/wQ0
6H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKigP48qIMYQYGXOSbL7IfvOWXV+TZ6o
9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMwlnUdZ/Yx4hOABHX6+0yQJxECU2OW
ve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnYFpV35FTy98dDYL7N6+y6whldMMQ6
80dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1TXQIDAQAB
-----END RSA PRIVATE KEY-----`,
		ASPublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtOL3THYTwCk35h9/BYpX
/5pQGH4jK5nyO55oI8PqBMx6GHfnP0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4Y
TNNmV/IQKqX1eUtRJSP/gZR5/wQ06H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKig
P48qIMYQYGXOSbL7IfvOWXV+TZ6o9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMw
lnUdZ/Yx4hOABHX6+0yQJxECU2OWve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnY
FpV35FTy98dDYL7N6+y6whldMMQ680dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1T
XQIDAQAB
-----END PUBLIC KEY-----`,
		TGSPublicKey: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA58L1zNrfqv6K6dNwBDLx
23Qsl5qhQdLvxuJBLBcX5JeKJ/GGHPoytB5MCgkBsk8/CM7BQpjx/CBmyT/7scVG
HGbA6PYi8807ZvoZDl8dCk/Uxy1tYRDeYVrQm2swwUhUTC9kIVYTBZtFzvZp//Ny
bQHgOKHABbsf5EjEG7AOI2qiUzJNRJPBzZtY0HdUoWYTWRTDiP/7yfVkm1PZsN+e
YyWhPVdXQ1JLrGjjwOZl0db5QhcUmXKjQWcy6/OMYsOjy4H7Mxtu7zGvPJObbTbk
KPeh25P9jExLW8XXcxkv6RUbYf3IAkDfMX8cJc3qtfcLW47Afywy0/zoLLQnQQVl
3QIDAQAB
-----END PUBLIC KEY-----`,
	}
}

// ==================== Helper Functions ====================

// getPrivateKey retrieves the AS's private key from the chaincode state
func (s *ASChaincode) getPrivateKey(ctx contractapi.TransactionContextInterface) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := ctx.GetStub().GetState("AS_PRIVATE_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to get AS private key: %v", err)
	}
	if privateKeyPEM == nil {
		return nil, fmt.Errorf("AS private key not found")
	}
	
	// Add debug logging
	fmt.Printf("Retrieved private key PEM (first 50 chars): %s...\n", 
		string(privateKeyPEM)[:min(50, len(string(privateKeyPEM)))])
	
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	
	// Ensure we're using the right parse function for the key format
	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try alternative parsing in case the key is in a different format
		parsedKey, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key (both PKCS1 and PKCS8): %v, %v", err, err2)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parsed key is not an RSA private key")
		}
	}
	
	return privateKey, nil
}

// getPublicKey retrieves the specified public key from the chaincode state
func (s *ASChaincode) getPublicKey(ctx contractapi.TransactionContextInterface, keyName string) (*rsa.PublicKey, error) {
	publicKeyPEM, err := ctx.GetStub().GetState(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key %s: %v", keyName, err)
	}
	if publicKeyPEM == nil {
		return nil, fmt.Errorf("public key %s not found", keyName)
	}
	
	// Add debug logging
	fmt.Printf("Retrieved %s (first 50 chars): %s...\n", 
		keyName, string(publicKeyPEM)[:min(50, len(string(publicKeyPEM)))])
	
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}
	
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}
	
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	return publicKey, nil
}

// getClientPublicKey retrieves a client's public key from the chaincode state
func (s *ASChaincode) getClientPublicKey(ctx contractapi.TransactionContextInterface, clientID string) (*rsa.PublicKey, error) {
	clientPublicKeyPEM, err := ctx.GetStub().GetState("CLIENT_PK_" + clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client public key: %v", err)
	}
	if clientPublicKeyPEM == nil {
		return nil, fmt.Errorf("client public key not found")
	}
	
	// Add debug logging
	fmt.Printf("Retrieved client public key (first 50 chars): %s...\n", 
		string(clientPublicKeyPEM)[:min(50, len(string(clientPublicKeyPEM)))])
	
	block, _ := pem.Decode(clientPublicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing client public key")
	}
	
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client public key: %v", err)
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
	fmt.Printf("Registering client: %s\n", clientID)
	fmt.Printf("Client public key (first 50 chars): %s...\n", 
		clientPublicKeyPEM[:min(50, len(clientPublicKeyPEM))])
	
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
	
	// Get transaction timestamp from the blockchain
	txTimestamp, err := getDeterministicTimestamp(ctx)
	if err != nil {
    	return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	
	// Create and store the client record
	client := ClientIdentity{
	    ID:              clientID,
	    PublicKey:       clientPublicKeyPEM,
	    RegistrationTime: txTimestamp,
	    Valid:           true,
	}
	
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return fmt.Errorf("failed to marshal client data: %v", err)
	}
	
	// Store client data in the world state
	err = ctx.GetStub().PutState("CLIENT_"+clientID, clientJSON)
	if err != nil {
		return fmt.Errorf("failed to store client data: %v", err)
	}
	
	// Store the client's public key separately for easy access
	err = ctx.GetStub().PutState("CLIENT_PK_"+clientID, []byte(clientPublicKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to store client public key: %v", err)
	}
	
	fmt.Printf("Successfully registered client: %s\n", clientID)
	return nil
}

// CheckClientValidity verifies if a client is valid
// This checks the client's registration status
func (s *ASChaincode) CheckClientValidity(ctx contractapi.TransactionContextInterface, clientID string) (bool, error) {
    fmt.Printf("Checking validity for client: %s\n", clientID)
    
    // Get the client record using the exact key format used when storing it
    clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
    if err != nil {
        return false, fmt.Errorf("failed to read client data: %v", err)
    }
    if clientJSON == nil {
        return false, fmt.Errorf("client %s does not exist", clientID)
    }
    
    // Debug: Log the client data
    fmt.Printf("Client data for %s: %s\n", clientID, string(clientJSON))
    
    var client ClientIdentity
    err = json.Unmarshal(clientJSON, &client)
    if err != nil {
        return false, fmt.Errorf("error unmarshaling client data: %v", err)
    }
    
    // Extra check to ensure ID field matches the requested client ID
    if client.ID != clientID {
        // If there's a mismatch, update the ID field to match
        client.ID = clientID
        // Optionally update the client record to fix the mismatch
        updatedClientJSON, err := json.Marshal(client)
        if err != nil {
            return false, fmt.Errorf("error marshaling updated client: %v", err)
        }
        err = ctx.GetStub().PutState("CLIENT_"+clientID, updatedClientJSON)
        if err != nil {
            return false, fmt.Errorf("error updating client record: %v", err)
        }
        
        fmt.Printf("Fixed client ID mismatch for %s\n", clientID)
    }
    
    // Check if the client is valid
    fmt.Printf("Client %s validity check result: %t\n", clientID, client.Valid)
    return client.Valid, nil
}

// InitiateAuthentication generates a nonce challenge for client authentication
// This is the first step in the authentication process as described in the paper
// Step 1: Client Requests Authentication from AS
func (s *ASChaincode) InitiateAuthentication(ctx contractapi.TransactionContextInterface, clientID string) (*NonceChallenge, error) {
	fmt.Printf("Initiating authentication for client: %s\n", clientID)
	
	// Check if client exists and is valid
	valid, err := s.CheckClientValidity(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("error checking client validity: %v", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid client")
	}
	
	// Get deterministic timestamp
    timestamp, err := getDeterministicTimestamp(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get timestamp: %v", err)
    }
    
    // Generate a deterministic nonce based on clientID and current timestamp
    nonceInput := clientID + strconv.FormatInt(timestamp.Unix(), 10)
    nonceHash := sha256.Sum256([]byte(nonceInput))
    nonce := base64.StdEncoding.EncodeToString(nonceHash[:])
    
    fmt.Printf("Generated nonce for client %s: %s\n", clientID, nonce)
    
    // Set expiration time for the nonce (e.g., 5 minutes from now)
    expirationTime := timestamp.Unix() + 300 // 5 minutes
    
    // Create the challenge response for the client
    challenge := NonceChallenge{
        Nonce:          nonce,
        ExpirationTime: expirationTime,
    }
    
    // Create and store the auth challenge in the world state
    authChallenge := AuthChallenge{
        ClientID:       clientID,
        Nonce:          nonce,
        ExpirationTime: expirationTime,
        CreatedAt:      timestamp,
    }
    
    // Convert to JSON
    authChallengeJSON, err := json.Marshal(authChallenge)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal auth challenge: %v", err)
    }
    
    // Store in world state with a deterministic key
    // This allows all peers to access the same challenge
    authChallengeKey := fmt.Sprintf("AUTH_CHALLENGE_%s", clientID)
    err = ctx.GetStub().PutState(authChallengeKey, authChallengeJSON)
    if err != nil {
        return nil, fmt.Errorf("failed to store auth challenge: %v", err)
    }
    
    fmt.Printf("Authentication challenge created for client %s\n", clientID)
    return &challenge, nil
}

// VerifyClientIdentity verifies a client's response to the nonce challenge using RSA encryption
// This implements the client authentication verification from the paper
// Step 3: AS decrypts the nonce using its private key to verify client identity
func (s *ASChaincode) VerifyClientIdentity(ctx contractapi.TransactionContextInterface, clientID string, encryptedNonce string) (bool, error) {
	fmt.Printf("Verifying client identity for: %s\n", clientID)
	
	// Retrieve the client record to confirm existence
    clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
    if err != nil {
        return false, fmt.Errorf("failed to read client data: %v", err)
    }
    if clientJSON == nil {
        return false, fmt.Errorf("client %s does not exist", clientID)
    }
    
    // Retrieve the auth challenge from world state
    authChallengeKey := fmt.Sprintf("AUTH_CHALLENGE_%s", clientID)
    authChallengeJSON, err := ctx.GetStub().GetState(authChallengeKey)
    if err != nil {
        return false, fmt.Errorf("failed to retrieve auth challenge: %v", err)
    }
    if authChallengeJSON == nil {
        return false, fmt.Errorf("no authentication challenge found for client")
    }
    
    // Parse the auth challenge
    var authChallenge AuthChallenge
    err = json.Unmarshal(authChallengeJSON, &authChallenge)
    if err != nil {
        return false, fmt.Errorf("failed to unmarshal auth challenge: %v", err)
    }
    
    // Check if the challenge has expired
    timestamp, err := getDeterministicTimestamp(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to get timestamp: %v", err)
    }
    
    if timestamp.Unix() > authChallenge.ExpirationTime {
        // Delete the expired challenge
        err = ctx.GetStub().DelState(authChallengeKey)
        if err != nil {
            return false, fmt.Errorf("failed to delete expired challenge: %v", err)
        }
        return false, fmt.Errorf("authentication challenge has expired")
    }
    
    // Get the AS private key to decrypt the client's response
    privateKey, err := s.getPrivateKey(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to get AS private key: %v", err)
    }
    
    // Decode the base64 encoded encrypted nonce
    encryptedNonceBytes, err := base64.StdEncoding.DecodeString(encryptedNonce)
    if err != nil {
        return false, fmt.Errorf("invalid encrypted nonce format: %v", err)
    }
    
    // Use a recovery mechanism for decryption
    var decryptedNonce []byte
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic during nonce decryption: %v", r)
        }
    }()
    
    // Decrypt the nonce using AS's private key
    decryptedNonce, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedNonceBytes)
    if err != nil {
        return false, fmt.Errorf("decryption failed: %v", err)
    }
    
    // Convert decrypted nonce to base64 for comparison
    decryptedNonceB64 := base64.StdEncoding.EncodeToString(decryptedNonce)
    
    fmt.Printf("Decrypted nonce: %s, Expected: %s\n", 
        decryptedNonceB64, authChallenge.Nonce)
    
    // Compare the decrypted nonce with the expected nonce
    if decryptedNonceB64 != authChallenge.Nonce {
        return false, nil
    }
    
    // Delete the used challenge from the world state
    err = ctx.GetStub().DelState(authChallengeKey)
    if err != nil {
        return false, fmt.Errorf("failed to delete used challenge: %v", err)
    }
    
    fmt.Printf("Client %s identity verified successfully\n", clientID)
    return true, nil
}

// VerifyClientIdentityWithSignature verifies a client's identity using signature-based verification
// This is a more compatible alternative to VerifyClientIdentity for cross-platform use
func (s *ASChaincode) VerifyClientIdentityWithSignature(ctx contractapi.TransactionContextInterface, clientID string, signedNonceBase64 string) (bool, error) {
    fmt.Printf("Verifying client %s identity using signature\n", clientID)
    
    // Retrieve the client record to confirm existence
    clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
    if err != nil {
        return false, fmt.Errorf("failed to read client data: %v", err)
    }
    if clientJSON == nil {
        return false, fmt.Errorf("client %s does not exist", clientID)
    }
    
    // Retrieve the auth challenge from world state
    authChallengeKey := fmt.Sprintf("AUTH_CHALLENGE_%s", clientID)
    authChallengeJSON, err := ctx.GetStub().GetState(authChallengeKey)
    if err != nil {
        return false, fmt.Errorf("failed to retrieve auth challenge: %v", err)
    }
    if authChallengeJSON == nil {
        return false, fmt.Errorf("no authentication challenge found for client")
    }
    
    // Parse the auth challenge
    var authChallenge AuthChallenge
    err = json.Unmarshal(authChallengeJSON, &authChallenge)
    if err != nil {
        return false, fmt.Errorf("failed to unmarshal auth challenge: %v", err)
    }
    
    // Check if the challenge has expired
    timestamp, err := getDeterministicTimestamp(ctx)
    if err != nil {
        return false, fmt.Errorf("failed to get timestamp: %v", err)
    }
    
    if timestamp.Unix() > authChallenge.ExpirationTime {
        // Delete the expired challenge
        err = ctx.GetStub().DelState(authChallengeKey)
        if err != nil {
            return false, fmt.Errorf("failed to delete expired challenge: %v", err)
        }
        return false, fmt.Errorf("authentication challenge has expired")
    }
    
    // Get client's public key
    clientPublicKey, err := s.getClientPublicKey(ctx, clientID)
    if err != nil {
        return false, fmt.Errorf("failed to get client public key: %v", err)
    }
    
    // Decode the base64 encoded signature
    signatureBytes, err := base64.StdEncoding.DecodeString(signedNonceBase64)
    if err != nil {
        return false, fmt.Errorf("invalid signature format: %v", err)
    }
    
    // Decode the nonce from base64
    nonceBytes, err := base64.StdEncoding.DecodeString(authChallenge.Nonce)
    if err != nil {
        return false, fmt.Errorf("invalid nonce format: %v", err)
    }
    
    // Create a hash of the nonce to verify against the signature
    hashed := sha256.Sum256(nonceBytes)
    
    // Use a recovery mechanism
    var verifyErr error
    defer func() {
        if r := recover(); r != nil {
            verifyErr = fmt.Errorf("panic during signature verification: %v", r)
        }
    }()
    
    // Verify the signature
    verifyErr = rsa.VerifyPKCS1v15(clientPublicKey, crypto.SHA256, hashed[:], signatureBytes)
    if verifyErr != nil {
        return false, fmt.Errorf("signature verification failed: %v", verifyErr)
    }
    
    // Signature is valid, delete the used challenge
    err = ctx.GetStub().DelState(authChallengeKey)
    if err != nil {
        return false, fmt.Errorf("failed to delete used challenge: %v", err)
    }
    fmt.Printf("Client %s identity verified successfully using signature\n", clientID)
    return true, nil
}

// GenerateTGT generates a Ticket Granting Ticket (TGT) for a client
// This implements Step 2: AS Issues TGT Encrypted with TGS's Public Key
func (s *ASChaincode) GenerateTGT(ctx contractapi.TransactionContextInterface, clientID string) (*ResponseToClient, error) {
    fmt.Printf("Generating TGT for client: %s\n", clientID)
    
    // Verify that client exists and is valid
    valid, err := s.CheckClientValidity(ctx, clientID)
    if err != nil {
        return nil, fmt.Errorf("failed to check client validity: %v", err)
    }
    if !valid {
        return nil, fmt.Errorf("invalid client")
    }
    
    // Get deterministic timestamp
    timestamp, err := getDeterministicTimestamp(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get timestamp: %v", err)
    }
    
    // Generate a deterministic session key based on clientID and timestamp
    // This ensures that if multiple organizations attempt to generate the same TGT,
    // they will produce identical results
    sessionKeyInput := clientID + strconv.FormatInt(timestamp.Unix(), 10) + "KU,TGS"
    sessionKeyHash := sha256.Sum256([]byte(sessionKeyInput))
    sessionKey := base64.StdEncoding.EncodeToString(sessionKeyHash[:])
    
    // Log session key generation (only in development)
    fmt.Printf("Generated session key for client %s: %s\n", clientID, sessionKey)
    
    // Create the TGT
    tgt := TGT{
        ClientID:   clientID,
        SessionKey: sessionKey,
        Timestamp:  timestamp,
        Lifetime:   3600, // 1 hour in seconds
    }
    
    // Convert TGT to JSON
    tgtJSON, err := json.Marshal(tgt)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal TGT: %v", err)
    }
    
    fmt.Printf("TGT JSON for client %s: %s\n", clientID, string(tgtJSON))
    
    // Get TGS's public key
    tgsPublicKey, err := s.getPublicKey(ctx, "TGS_PUBLIC_KEY")
    if err != nil {
        return nil, fmt.Errorf("failed to get TGS public key: %v", err)
    }
    
    // Encrypt TGT with TGS's public key
    // This implements: TGT = {Client ID, KU,TGS, Timestamp, Lifetime}eTGS = M^eTGS mod nTGS
    encryptedTGT, err := rsa.EncryptPKCS1v15(rand.Reader, tgsPublicKey, tgtJSON)
    if err != nil {
        return nil, fmt.Errorf("TGT encryption failed: %v", err)
    }
    
    // Encode the encrypted TGT as base64
    encryptedTGTBase64 := base64.StdEncoding.EncodeToString(encryptedTGT)
    fmt.Printf("Encrypted TGT for client %s (first 50 chars): %s...\n", 
               clientID, encryptedTGTBase64[:min(50, len(encryptedTGTBase64))])
    
    // Get client's public key
    clientPublicKey, err := s.getClientPublicKey(ctx, clientID)
    if err != nil {
        return nil, fmt.Errorf("failed to get client public key: %v", err)
    }
    
    // Encrypt the session key with client's public key
    // This implements: {KU,TGS}eU = KU,TGS^eU mod nU
    encryptedSessionKey, err := rsa.EncryptPKCS1v15(rand.Reader, clientPublicKey, []byte(sessionKey))
    if err != nil {
        return nil, fmt.Errorf("session key encryption failed: %v", err)
    }
    
    // Create the response for the client
    response := ResponseToClient{
        EncryptedTGT:        encryptedTGTBase64,
        EncryptedSessionKey: base64.StdEncoding.EncodeToString(encryptedSessionKey),
    }
    
    // Record this TGT issuance on the ledger for audit purposes
    tgtRecord := struct {
        ClientID  string    `json:"clientID"`
        Timestamp time.Time `json:"timestamp"`
        TGTHash   string    `json:"tgtHash"`
    }{
        ClientID:  clientID,
        Timestamp: timestamp,
        TGTHash:   fmt.Sprintf("%x", sha256.Sum256(tgtJSON)),
    }
    
    tgtRecordJSON, err := json.Marshal(tgtRecord)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal TGT record: %v", err)
    }
    
    // Store the TGT record in the world state with deterministic ID
    tgtID := "TGT_" + clientID + "_" + strconv.FormatInt(tgt.Timestamp.Unix(), 10)
    err = ctx.GetStub().PutState(tgtID, tgtRecordJSON)
    if err != nil {
        return nil, fmt.Errorf("failed to store TGT record: %v", err)
    }
    
    fmt.Printf("Generated TGT for client %s successfully\n", clientID)
    return &response, nil
}

// GetAllClientRegistrations retrieves all client registrations
// This implements the operation to get all registrations from clients
func (s *ASChaincode) GetAllClientRegistrations(ctx contractapi.TransactionContextInterface) ([]*ClientIdentity, error) {
    fmt.Println("Getting all client registrations")
    
    // Get all client registrations from the world state
    resultsIterator, err := ctx.GetStub().GetStateByRange("CLIENT_", "CLIENT_~")
    if err != nil {
        return nil, fmt.Errorf("failed to get client records: %v", err)
    }
    defer resultsIterator.Close()
    
    var clients []*ClientIdentity
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate client records: %v", err)
        }
        
        // Skip keys that don't match client records (e.g., CLIENT_PK_ keys)
        if strings.HasPrefix(queryResponse.Key, "CLIENT_PK_") {
            continue
        }
        
        // Extract client ID from the key (remove the "CLIENT_" prefix)
        clientID := queryResponse.Key[7:] // Skip the "CLIENT_" prefix
        
        var client ClientIdentity
        err = json.Unmarshal(queryResponse.Value, &client)
        if err != nil {
            fmt.Printf("Error unmarshaling client %s: %v\n", clientID, err)
            continue // Skip this record but continue processing others
        }
        
        // Ensure the ID field matches the key used to store it
        if client.ID != clientID {
            client.ID = clientID
        }
        
        clients = append(clients, &client)
    }
    
    fmt.Printf("Found %d client registrations\n", len(clients))
    return clients, nil
}

// AllocatePeerTask assigns a task to a specific peer
// This implements task allocation for efficient processing
func (s *ASChaincode) AllocatePeerTask(ctx contractapi.TransactionContextInterface, peerID string, taskType string, clientID string) error {
    fmt.Printf("Allocating %s task for client %s to peer %s\n", taskType, clientID, peerID)
    
    // Get deterministic timestamp
    timestamp, err := getDeterministicTimestamp(ctx)
    if err != nil {
        return fmt.Errorf("failed to get timestamp: %v", err)
    }
    
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
        AssignedAt:  timestamp,
        Status:      "assigned",
    }
    
    taskJSON, err := json.Marshal(task)
    if err != nil {
        return fmt.Errorf("failed to marshal task data: %v", err)
    }
    
    // Store the task in the world state with deterministic ID
    taskID := "TASK_" + peerID + "_" + clientID + "_" + taskType
    err = ctx.GetStub().PutState(taskID, taskJSON)
    if err != nil {
        return fmt.Errorf("failed to store task data: %v", err)
    }
    
    fmt.Printf("Task allocated successfully: %s\n", taskID)
    return nil
}

// ReserveAndValidateRegistration finalizes a client registration
// This is used for reserving and validating client registrations
func (s *ASChaincode) ReserveAndValidateRegistration(ctx contractapi.TransactionContextInterface, clientID string) error {
    fmt.Printf("Reserving and validating registration for client: %s\n", clientID)
    
    // Retrieve the client record
    clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
    if err != nil {
        return fmt.Errorf("failed to read client data: %v", err)
    }
    if clientJSON == nil {
        return fmt.Errorf("client %s does not exist", clientID)
    }
    
    var client ClientIdentity
    err = json.Unmarshal(clientJSON, &client)
    if err != nil {
        return fmt.Errorf("failed to unmarshal client data: %v", err)
    }
    
    // Mark the client as valid (this would include more validation in a real system)
    client.Valid = true
    
    // Update the client record
    updatedClientJSON, err := json.Marshal(client)
    if err != nil {
        return fmt.Errorf("failed to marshal updated client data: %v", err)
    }
    
    err = ctx.GetStub().PutState("CLIENT_"+clientID, updatedClientJSON)
    if err != nil {
        return fmt.Errorf("failed to store updated client data: %v", err)
    }
    
    fmt.Printf("Client %s registration reserved and validated successfully\n", clientID)
    return nil
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

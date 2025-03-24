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
		return nil
	}
	
	// Use predefined keys instead of generating them dynamically
	// This ensures all peers have the same keys
	keys := getPredefinedKeys()
	
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
80dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1TXQIDAQABAoIBAC7HQRRQBHFmBMwj
nNMEZgQLYnwoE5mTUExBLPfkMKoMfSlnqJHGbXvL7r7h4LtW01+HzxutxjXJc/6n
ObmLCLHwTDVxHYmUALjhpBPtuDcyGUYdCwNKnXzkGpG4DfE0rE9y93VnfX9JLs05
7aZdEDK5QoGNUqdW9nOI2lyHUZiQu7zUZbYQakxd+7zbjdO1NZHaUrh+s1co5QhB
KnPWjTuKZQHIf5H5EPUEJQTIexb5+csP8SjXJ0M5kDRR7C2u5xCYm5YgVrvMKeK+
AKmJLbQsYl2+9X9HIEiX4GaXO5+hHjTXZWQiAy4zdFcH45jGqUqRCbXr9Il4KnDr
rIphgAECgYEA7TA0a5h3WAUpYBLFBnUnNPpxhxBySN9rZ/xRvdJ6jLOnjueUgP+Z
aP57kYWs5JSZG0gLOCL1ilJ1D4jXx6UO7oeNZ9EBUmIILNUPxnk2/+lhVYj4/4FS
2QFd0j/oj1C+WKjk8JK9GjFQOzBYyzqPT3EGU8/c5T9U1C2N5B/2RG0CgYEAwuLC
o0lKSJnfJjG2n7iCvEDQYBKEzl8jYXKXnmKcTCeQRBDjA8K+1iyYs9FnYgj+BIXQ
mDsXCVXjzYWGAOzDc3pGpHKEPLRl1O7IZWjl6xiTvG/v1gE+OtJe3qZLzOkUQaFh
+OhEzWnOfJ2Jajz4+5A+xHzOkF9Rx1gQPODg4AECgYEAxlPuIR0WxwOWLK3n1pLz
nEmJGGhTbBV/0AKinK+7vDbFCchUGQ13nPQrcZSKIuODhvJLJ9NXbqSnzK4ZiMj+
wkr5v8xFBzHSoG11IHVxOjQCYWvgnkFhUiGZLDJMYrWKu9DBodcCrcQYKVtu2AAz
ahJYHRh9WuQpC0bqABX+NvUCgYEAm02Bfiwza+cL827L3c3+Uz5Z9LVJgH8pQK4P
pJRLWLcQdRMuvCckK9YkwIju4FVAKMfb3QpHdKPDG9175RCxxD6Z0LtFSX0SQClA
3KwMUW9X1vXb27B6IceVmzbJd+iGXTU6o1d32wM6HHZtg6xPVg7VQA/pbbOa9L4J
GqoBAIECgYANGUrwlkjqGjrYXJ4jnBaykVvvCZW6n7mBVJLUm/cUQGWzDsJeMROA
qnPhOdmA1YRO9yzrk2kGSH9tIhoJyPKRsKwNAUa+CRVxJYBy/3+OuELQ8ZwBGZnO
ZrZpLIx6XR+0VYcKGhFPVnm2SyORJ+YxmlSEbxHV54ymuhZbBNRJ9Q==
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

// getClientPublicKey retrieves a client's public key from the chaincode state
func (s *ASChaincode) getClientPublicKey(ctx contractapi.TransactionContextInterface, clientID string) (*rsa.PublicKey, error) {
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
	
	// Get deterministic timestamp
    timestamp, err := getDeterministicTimestamp(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get timestamp: %v", err)
    }
    
    // Generate a deterministic nonce based on clientID and current timestamp
    nonceInput := clientID + strconv.FormatInt(timestamp.Unix(), 10)
    nonceHash := sha256.Sum256([]byte(nonceInput))
    nonce := base64.StdEncoding.EncodeToString(nonceHash[:])
    
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
        return nil, err
    }
    
    // Store in world state with a deterministic key
    // This allows all peers to access the same challenge
    authChallengeKey := fmt.Sprintf("AUTH_CHALLENGE_%s", clientID)
    err = ctx.GetStub().PutState(authChallengeKey, authChallengeJSON)
    if err != nil {
        return nil, err
    }
    
    return &challenge, nil
}

// VerifyClientIdentity verifies a client's response to the nonce challenge using RSA encryption
// This implements the client authentication verification from the paper
// Step 3: AS decrypts the nonce using its private key to verify client identity
func (s *ASChaincode) VerifyClientIdentity(ctx contractapi.TransactionContextInterface, clientID string, encryptedNonce string) (bool, error) {
	// Retrieve the client record to confirm existence
    clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
    if err != nil {
        return false, err
    }
    if clientJSON == nil {
        return false, fmt.Errorf("client %s does not exist", clientID)
    }
    
    // Retrieve the auth challenge from world state
    authChallengeKey := fmt.Sprintf("AUTH_CHALLENGE_%s", clientID)
    authChallengeJSON, err := ctx.GetStub().GetState(authChallengeKey)
    if err != nil {
        return false, err
    }
    if authChallengeJSON == nil {
        return false, fmt.Errorf("no authentication challenge found for client")
    }
    
    // Parse the auth challenge
    var authChallenge AuthChallenge
    err = json.Unmarshal(authChallengeJSON, &authChallenge)
    if err != nil {
        return false, err
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
            return false, err
        }
        return false, fmt.Errorf("authentication challenge has expired")
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
    decryptedNonce, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedNonceBytes)
    if err != nil {
        return false, fmt.Errorf("decryption failed: %v", err)
    }
    
    // Convert decrypted nonce to base64 for comparison
    decryptedNonceB64 := base64.StdEncoding.EncodeToString(decryptedNonce)
    
    // Compare the decrypted nonce with the expected nonce
    if decryptedNonceB64 != authChallenge.Nonce {
        return false, nil
    }
    
    // Delete the used challenge from the world state
    err = ctx.GetStub().DelState(authChallengeKey)
    if err != nil {
        return false, err
    }
    
    return true, nil
}

// VerifyClientIdentityWithSignature verifies a client's identity using signature-based verification
// This is a more compatible alternative to VerifyClientIdentity for cross-platform use
func (s *ASChaincode) VerifyClientIdentityWithSignature(ctx contractapi.TransactionContextInterface, clientID string, signedNonceBase64 string) (bool, error) {
    // Retrieve the client record to confirm existence
    clientJSON, err := ctx.GetStub().GetState("CLIENT_" + clientID)
    if err != nil {
        return false, err
    }
    if clientJSON == nil {
        return false, fmt.Errorf("client %s does not exist", clientID)
    }
    
    // Retrieve the auth challenge from world state
    authChallengeKey := fmt.Sprintf("AUTH_CHALLENGE_%s", clientID)
    authChallengeJSON, err := ctx.GetStub().GetState(authChallengeKey)
    if err != nil {
        return false, err
    }
    if authChallengeJSON == nil {
        return false, fmt.Errorf("no authentication challenge found for client")
    }
    
    // Parse the auth challenge
    var authChallenge AuthChallenge
    err = json.Unmarshal(authChallengeJSON, &authChallenge)
    if err != nil {
        return false, err
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
            return false, err
        }
        return false, fmt.Errorf("authentication challenge has expired")
    }
    
    // Get client's public key
    clientPublicKey, err := s.getClientPublicKey(ctx, clientID)
    if err != nil {
        return false, err
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
    
    // Verify the signature
    err = rsa.VerifyPKCS1v15(clientPublicKey, crypto.SHA256, hashed[:], signatureBytes)
    if err != nil {
        return false, fmt.Errorf("signature verification failed: %v", err)
    }
    
    // Signature is valid, delete the used challenge
    err = ctx.GetStub().DelState(authChallengeKey)
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
	clientPublicKey, err := s.getClientPublicKey(ctx, clientID)
	if err != nil {
		return nil, err
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
		Timestamp: timestamp,
		TGTHash:   fmt.Sprintf("%x", sha256.Sum256(tgtJSON)),
	}
	
	tgtRecordJSON, err := json.Marshal(tgtRecord)
	if err != nil {
		return nil, err
	}
	
	// Store the TGT record in the world state with deterministic ID
	tgtID := "TGT_" + clientID + "_" + strconv.FormatInt(tgt.Timestamp.Unix(), 10)
	err = ctx.GetStub().PutState(tgtID, tgtRecordJSON)
	if err != nil {
		return nil, err
	}
	
	return &response, nil

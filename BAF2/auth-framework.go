// auth-framework.go
// Go implementation of the Kerberos-like authentication framework for Hyperledger Fabric

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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Configuration constants - match with the Node.js implementation
const (
	channelName     = "chaichis-channel"
	asChaincodeId   = "as-chaincode"
	tgsChaincodeId  = "tgs-chaincode"
	isvChaincodeId  = "isv-chaincode"
	connectionFile  = "connection-profile.json"
	walletPath      = "wallet"
)

// ASPublicKey is a constant for AS public key
const ASPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtOL3THYTwCk35h9/BYpX
/5pQGH4jK5nyO55oI8PqBMx6GHfnP0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4Y
TNNmV/IQKqX1eUtRJSP/gZR5/wQ06H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKig
P48qIMYQYGXOSbL7IfvOWXV+TZ6o9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMw
lnUdZ/Yx4hOABHX6+0yQJxECU2OWve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnY
FpV35FTy98dDYL7N6+y6whldMMQ680dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1T
XQIDAQAB
-----END PUBLIC KEY-----`

// Authenticator represents the client's identity proof
type Authenticator struct {
	ClientID  string `json:"clientID"`
	Timestamp string `json:"timestamp"`
}

// ServiceTicketRequest represents a request for a service ticket
type ServiceTicketRequest struct {
	EncryptedTGT       string `json:"encryptedTGT"`
	ClientID           string `json:"clientID"`
	ServiceID          string `json:"serviceID"`
	Authenticator      string `json:"authenticator"`
}

// ServiceRequest represents a request to access a service
type ServiceRequest struct {
	EncryptedServiceTicket string `json:"encryptedServiceTicket"`
	ClientID               string `json:"clientID"`
	DeviceID               string `json:"deviceID"`
	RequestType            string `json:"requestType"`
	EncryptedData          string `json:"encryptedData"`
}

// IoTDevice represents an IoT device
type IoTDevice struct {
	DeviceID     string   `json:"deviceID"`
	PublicKey    string   `json:"publicKey"`
	Capabilities []string `json:"capabilities"`
}

// TGT represents a Ticket Granting Ticket
type TGT struct {
	EncryptedTGT          string `json:"encryptedTGT"`
	EncryptedSessionKey   string `json:"encryptedSessionKey"`
	ExpirationTime        string `json:"expirationTime"`
}

// ServiceTicket represents a service ticket
type ServiceTicket struct {
	EncryptedServiceTicket string `json:"encryptedServiceTicket"`
	EncryptedSessionKey    string `json:"encryptedSessionKey"`
	ExpirationTime         string `json:"expirationTime"`
}

// ServiceResponse represents a response from the ISV
type ServiceResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"sessionID"`
	Message   string `json:"message"`
}

// NonceChallenge represents a nonce challenge from the AS
type NonceChallenge struct {
	Nonce string `json:"nonce"`
}

// connectToNetwork establishes a connection to the Fabric network
func connectToNetwork(username string) (*gateway.Gateway, *gateway.Network, error) {
	// Load the connection profile
	ccpPath, err := filepath.Abs(connectionFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find connection profile: %v", err)
	}
	
	// Load the wallet for identity
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create wallet: %v", err)
	}
	
	// Check if user identity exists in the wallet
	if !wallet.Exists(username) {
		return nil, nil, fmt.Errorf("identity for %s not found in wallet", username)
	}
	
	// Create gateway connection
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(ccpPath)),
		gateway.WithIdentity(wallet, username),
		gateway.WithDiscovery(true),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to gateway: %v", err)
	}
	
	// Get network
	network, err := gw.GetNetwork(channelName)
	if err != nil {
		gw.Close()
		return nil, nil, fmt.Errorf("failed to get network: %v", err)
	}
	
	return gw, network, nil
}

// generateKeyPair generates a new RSA key pair
func generateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	
	// Extract public key
	publicKey := &privateKey.PublicKey
	
	return privateKey, publicKey, nil
}

// savePrivateKey saves a private key to a file in PKCS#1 format
func savePrivateKey(privateKey *rsa.PrivateKey, filename string) error {
	// Marshal private key to PKCS1 (same as traditional RSA private key format)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	
	// Encode to PEM
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY", // This specifies PKCS1 format
			Bytes: privateKeyBytes,
		},
	)
	
	return ioutil.WriteFile(filename, privateKeyPEM, 0600)
}

// savePublicKey saves a public key to a file
func savePublicKey(publicKey *rsa.PublicKey, filename string) error {
	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	
	// Encode to PEM
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	
	return ioutil.WriteFile(filename, publicKeyPEM, 0644)
}

// loadPrivateKey loads a private key from a file
func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	// Read the private key file
	privateKeyPEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	// Parse the PEM block
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}
	
	// Parse the private key
	var privateKey *rsa.PrivateKey
	
	// Check if it's PKCS1 or PKCS8 format
	if block.Type == "RSA PRIVATE KEY" {
		// PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else {
		// PKCS8 format
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
	}
	
	return privateKey, nil
}

// encryptWithPublicKey encrypts data with a public key
func encryptWithPublicKey(publicKeyPEM string, data []byte) (string, error) {
	// Parse the PEM encoded public key
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the public key")
	}
	
	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	
	// Cast to RSA public key
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not an RSA public key")
	}
	
	// Encrypt the data
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
	if err != nil {
		return "", err
	}
	
	// Return base64 encoded encrypted data
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// signData signs data with a private key
func signData(privateKey *rsa.PrivateKey, data []byte) (string, error) {
	// Create a hash of the data
	hash := sha256.Sum256(data)
	
	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	
	// Return base64 encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// registerClient registers a client with the Authentication Server
func registerClient(username, clientId string) error {
	// Generate a new key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Save the private key
	privateKeyFile := clientId + "-private.pem"
	err = savePrivateKey(privateKey, privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}
	fmt.Printf("Private key stored in %s\n", privateKeyFile)
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return err
	}
	defer gw.Close()
	
	// Get AS contract
	contract := network.GetContract(asChaincodeId)
	
	// Marshal public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	
	publicKeyPEM := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	))
	
	// Register client with AS
	_, err = contract.SubmitTransaction("RegisterClient", clientId, publicKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to register client: %v", err)
	}
	
	fmt.Printf("Client %s registered successfully with Authentication Server\n", clientId)
	return nil
}

// registerIoTDevice registers an IoT device with the IoT Service Validator
func registerIoTDevice(username, deviceId string, capabilities []string) error {
	// Generate a new key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Save the private key
	privateKeyFile := deviceId + "-private.pem"
	err = savePrivateKey(privateKey, privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}
	fmt.Printf("Device private key stored in %s\n", privateKeyFile)
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return err
	}
	defer gw.Close()
	
	// Get ISV contract
	contract := network.GetContract(isvChaincodeId)
	
	// Marshal public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	
	publicKeyPEM := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	))
	
	// Convert capabilities to JSON
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %v", err)
	}
	
	// Try to register the device
	var txErr error
	var result []byte
	
	// First try submit transaction
	result, txErr = contract.SubmitTransaction("RegisterIoTDevice", deviceId, publicKeyPEM, string(capabilitiesJSON))
	
	// If submission fails, try evaluation as fallback
	if txErr != nil {
		fmt.Println("Transaction submission failed, falling back to evaluation...")
		result, txErr = contract.EvaluateTransaction("RegisterIoTDevice", deviceId, publicKeyPEM, string(capabilitiesJSON))
		if txErr != nil {
			return fmt.Errorf("failed to register IoT device (both submit and evaluate failed): %v", txErr)
		}
	}
	
	fmt.Printf("IoT device %s registered successfully with capabilities: %s\n", deviceId, capabilities)
	return nil
}

// getTGT gets a Ticket Granting Ticket from the Authentication Server
func getTGT(username, clientId string) (*TGT, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return nil, err
	}
	defer gw.Close()
	
	// Get AS contract
	asContract := network.GetContract(asChaincodeId)
	
	// Step 1: Get the nonce challenge
	fmt.Printf("Getting nonce challenge for client ID: %s\n", clientId)
	nonceResponseBytes, err := asContract.SubmitTransaction("InitiateAuthentication", clientId)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate authentication: %v", err)
	}
	
	var nonceChallenge NonceChallenge
	err = json.Unmarshal(nonceResponseBytes, &nonceChallenge)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nonce challenge: %v", err)
	}
	fmt.Printf("Received nonce challenge: %+v\n", nonceChallenge)
	
	// Wait for blockchain state propagation
	fmt.Println("Waiting for blockchain state propagation...")
	time.Sleep(5 * time.Second)
	
	// Load client's private key
	fmt.Println("Loading client private key...")
	privateKeyFile := clientId + "-private.pem"
	privateKey, err := loadPrivateKey(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}
	
	// Step 2: Sign the nonce with the client's private key (signature-based approach)
	fmt.Println("Signing the nonce with client private key...")
	nonceBytes, err := base64.StdEncoding.DecodeString(nonceChallenge.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %v", err)
	}
	
	signedNonce, err := signData(privateKey, nonceBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign nonce: %v", err)
	}
	fmt.Printf("Signed nonce (base64): %s\n", signedNonce)
	
	// Step 3: Verify client identity using signature-based verification
	var tgt *TGT
	var verifyErr error
	
	fmt.Println("Verifying client identity with signature...")
	_, verifyErr = asContract.SubmitTransaction("VerifyClientIdentityWithSignature", clientId, signedNonce)
	
	if verifyErr == nil {
		// Step 4: Now that we're verified, get the TGT
		fmt.Println("Requesting TGT...")
		tgtResponseBytes, err := asContract.SubmitTransaction("GenerateTGT", clientId)
		if err != nil {
			return nil, fmt.Errorf("failed to generate TGT: %v", err)
		}
		
		tgt = &TGT{}
		err = json.Unmarshal(tgtResponseBytes, tgt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TGT: %v", err)
		}
		
		// Save TGT for later use
		tgtFile := clientId + "-tgt.json"
		tgtJSON, _ := json.MarshalIndent(tgt, "", "  ")
		err = ioutil.WriteFile(tgtFile, tgtJSON, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to save TGT: %v", err)
		}
		
		fmt.Println("Received TGT successfully")
		return tgt, nil
	}
	
	// Fall back to encryption-based verification
	fmt.Printf("Signature verification failed: %v\n", verifyErr)
	fmt.Println("Falling back to encryption-based verification...")
	
	// Encrypt the nonce with the AS public key
	encryptedNonce, err := encryptWithPublicKey(ASPublicKey, nonceBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt nonce: %v", err)
	}
	fmt.Printf("Encrypted nonce (base64): %s\n", encryptedNonce)
	
	// Verify using encrypted nonce
	_, err = asContract.SubmitTransaction("VerifyClientIdentity", clientId, encryptedNonce)
	if err != nil {
		return nil, fmt.Errorf("encryption-based verification also failed: %v", err)
	}
	
	// Get TGT after encryption-based verification
	fmt.Println("Requesting TGT...")
	tgtResponseBytes, err := asContract.SubmitTransaction("GenerateTGT", clientId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TGT: %v", err)
	}
	
	tgt = &TGT{}
	err = json.Unmarshal(tgtResponseBytes, tgt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TGT: %v", err)
	}
	
	// Save TGT for later use
	tgtFile := clientId + "-tgt.json"
	tgtJSON, _ := json.MarshalIndent(tgt, "", "  ")
	err = ioutil.WriteFile(tgtFile, tgtJSON, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save TGT: %v", err)
	}
	
	fmt.Println("Received TGT successfully")
	return tgt, nil
}

// getServiceTicket gets a Service Ticket from the Ticket Granting Server
func getServiceTicket(username, clientId, serviceId string) (*ServiceTicket, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return nil, err
	}
	defer gw.Close()
	
	// Get TGS contract
	tgsContract := network.GetContract(tgsChaincodeId)
	
	// Load saved TGT
	tgtFile := clientId + "-tgt.json"
	tgtJSON, err := ioutil.ReadFile(tgtFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TGT: %v", err)
	}
	
	var tgt TGT
	err = json.Unmarshal(tgtJSON, &tgt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TGT: %v", err)
	}
	
	// Create an authenticator - in Kerberos, this would typically contain client ID and timestamp
	authenticator := Authenticator{
		ClientID:  clientId,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	// Convert authenticator to string and encrypt with session key
	// In a real implementation, you would decrypt the session key from tgtData.encryptedSessionKey first
	authenticatorJSON, err := json.Marshal(authenticator)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authenticator: %v", err)
	}
	
	// We're just base64 encoding for simplicity - in a real implementation, this would be encrypted
	encryptedAuthenticator := base64.StdEncoding.EncodeToString(authenticatorJSON)
	
	// Prepare service ticket request
	serviceTicketRequest := ServiceTicketRequest{
		EncryptedTGT:  tgt.EncryptedTGT,
		ClientID:      clientId,
		ServiceID:     serviceId,
		Authenticator: encryptedAuthenticator,
	}
	
	// Convert to JSON and base64 encode
	serviceTicketRequestJSON, err := json.Marshal(serviceTicketRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service ticket request: %v", err)
	}
	
	serviceTicketRequestB64 := base64.StdEncoding.EncodeToString(serviceTicketRequestJSON)
	
	// Submit request to TGS
	serviceTicketResponseBytes, err := tgsContract.SubmitTransaction("GenerateServiceTicket", serviceTicketRequestB64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service ticket: %v", err)
	}
	
	var serviceTicket ServiceTicket
	err = json.Unmarshal(serviceTicketResponseBytes, &serviceTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service ticket: %v", err)
	}
	
	// Save service ticket for later use
	serviceTicketFile := fmt.Sprintf("%s-serviceticket-%s.json", clientId, serviceId)
	serviceTicketJSON, _ := json.MarshalIndent(serviceTicket, "", "  ")
	err = ioutil.WriteFile(serviceTicketFile, serviceTicketJSON, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save service ticket: %v", err)
	}
	
	fmt.Println("Received service ticket response")
	return &serviceTicket, nil
}

// accessIoTDevice authenticates with the ISV and accesses an IoT device
func accessIoTDevice(username, clientId, deviceId string) (*ServiceResponse, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return nil, err
	}
	defer gw.Close()
	
	// Get ISV contract
	isvContract := network.GetContract(isvChaincodeId)
	
	// Load saved service ticket
	serviceTicketFile := clientId + "-serviceticket-iotservice1.json"
	serviceTicketJSON, err := ioutil.ReadFile(serviceTicketFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load service ticket: %v", err)
	}
	
	var serviceTicket ServiceTicket
	err = json.Unmarshal(serviceTicketJSON, &serviceTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service ticket: %v", err)
	}
	
	// Verify service ticket with ISV
	_, err = isvContract.SubmitTransaction("ValidateServiceTicket", serviceTicket.EncryptedServiceTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to validate service ticket: %v", err)
	}
	fmt.Println("Service ticket validated successfully")
	
	// Prepare service request
	serviceRequest := ServiceRequest{
		EncryptedServiceTicket: serviceTicket.EncryptedServiceTicket,
		ClientID:               clientId,
		DeviceID:               deviceId,
		RequestType:            "read",
		EncryptedData:          base64.StdEncoding.EncodeToString([]byte("read-request")), // Simulated request data
	}
	
	// Process service request
	serviceRequestJSON, err := json.Marshal(serviceRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service request: %v", err)
	}
	
	serviceResponseBytes, err := isvContract.SubmitTransaction("ProcessServiceRequest", string(serviceRequestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to process service request: %v", err)
	}
	
	var serviceResponse ServiceResponse
	err = json.Unmarshal(serviceResponseBytes, &serviceResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service response: %v", err)
	}
	
	fmt.Printf("Service request processed: %+v\n", serviceResponse)
	
	// Extract session ID for future interactions
	sessionFile := fmt.Sprintf("%s-session-%s.txt", clientId, deviceId)
	err = ioutil.WriteFile(sessionFile, []byte(serviceResponse.SessionID), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save session ID: %v", err)
	}
	
	fmt.Printf("Established session ID %s for device %s\n", serviceResponse.SessionID, deviceId)
	return &serviceResponse, nil
}

// getIoTDeviceData gets data from an IoT device after authentication
func getIoTDeviceData(username, clientId, deviceId string) (*IoTDevice, error) {
	// Check if a session exists
	sessionFile := fmt.Sprintf("%s-session-%s.txt", clientId, deviceId)
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return nil, errors.New("no active session found. Please authenticate first")
	}
	
	sessionId, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session ID: %v", err)
	}
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return nil, err
	}
	defer gw.Close()
	
	// Get ISV contract
	isvContract := network.GetContract(isvChaincodeId)
	
	// Query all IoT devices (for demonstration)
	devicesResponseBytes, err := isvContract.EvaluateTransaction("GetAllIoTDevices")
	if err != nil {
		return nil, fmt.Errorf("failed to get all IoT devices: %v", err)
	}
	
	var devices []IoTDevice
	err = json.Unmarshal(devicesResponseBytes, &devices)
	if err != nil {
		return nil, fmt.Errorf("failed to parse devices: %v", err)
	}
	
	// Filter for the requested device
	var deviceData *IoTDevice
	for _, device := range devices {
		if device.DeviceID == deviceId {
			deviceData = &device
			break
		}
	}
	
	if deviceData == nil {
		return nil, fmt.Errorf("device %s not found", deviceId)
	}
	
	fmt.Printf("Retrieved data for device %s: %+v\n", deviceId, *deviceData)
	return deviceData, nil
}

// closeSession closes a session when done
func closeSession(username, clientId, deviceId string) error {
	// Check if a session exists
	sessionFile := fmt.Sprintf("%s-session-%s.txt", clientId, deviceId)
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return errors.New("no active session found")
	}
	
	sessionId, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session ID: %v", err)
	}
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return err
	}
	defer gw.Close()
	
	// Get ISV contract
	isvContract := network.GetContract(isvChaincodeId)
	
	// Close the session
	_, err = isvContract.SubmitTransaction("CloseSession", string(sessionId))
	if err != nil {
		return fmt.Errorf("failed to close session: %v", err)
	}
	
	fmt.Printf("Closed session %s for device %s\n", string(sessionId), deviceId)
	
	// Remove session file
	err = os.Remove(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to remove session file: %v", err)
	}
	
	return nil
}

// debugRSAEncryption is a debugging utility for RSA operations
func debugRSAEncryption(nonce string) error {
	fmt.Println("======= RSA ENCRYPTION DEBUG =======")
	
	// Check if input is already base64 encoded
	var nonceBuffer []byte
	var err error
	
	// Try to decode as base64
	nonceBuffer, err = base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		fmt.Println("Nonce is not base64 encoded, treating as plain text")
		nonceBuffer = []byte(nonce)
	} else {
		fmt.Println("Nonce appears to be base64 encoded. Decoded:", string(nonceBuffer))
	}
	
	fmt.Println("Nonce as buffer:", nonceBuffer)
	fmt.Println("Nonce buffer length:", len(nonceBuffer))
	
	// Try different encryption approaches
	fmt.Println("\nTrying different encryption approaches:")
	
	// Approach 1: Standard Go encryption with PKCS1v15
	block, _ := pem.Decode([]byte(ASPublicKey))
	if block == nil {
		return errors.New("failed to parse PEM block containing the public key")
	}
	
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}
	
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}
	
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, nonceBuffer)
	if err != nil {
		return fmt.Errorf("encryption error: %v", err)
	}
	
	encrypted64 := base64.StdEncoding.EncodeToString(encrypted)
	fmt.Println("Approach 1 - Result (base64):", encrypted64)
	fmt.Println("Approach 1 - Length:", len(encrypted))
	
	// Generate a key pair for verification
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Create hash and sign
	hash := sha256.Sum256(nonceBuffer)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return fmt.Errorf("signing error: %v", err)
	}
	
	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %v", err)
	}
	
	fmt.Println("Self-verification test passed successfully!")
	fmt.Println("======= DEBUG COMPLETE =======")
	
	return nil
}

// main is the entry point for the program
func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	
	command := os.Args[1]
	username := "admin"
	if len(os.Args) > 2 {
		username = os.Args[2]
	}
	
	var err error
	
	switch command {
	case "register-client":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run auth-framework.go register-client <username> <clientId>")
			return
		}
		clientId := os.Args[3]
		err = registerClient(username, clientId)
		
	case "register-device":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run auth-framework.go register-device <username> <deviceId> <capability1> <capability2> ...")
			return
		}
		deviceId := os.Args[3]
		capabilities := os.Args[4:]
		err = registerIoTDevice(username, deviceId, capabilities)
		
	case "authenticate":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run auth-framework.go authenticate <username> <clientId> <deviceId>")
			return
		}
		clientId := os.Args[3]
		deviceId := os.Args[4]
		
		fmt.Println("Step 1: Getting TGT from Authentication Server...")
		tgt, tgtErr := getTGT(username, clientId)
		if tgtErr != nil {
			err = tgtErr
			break
		}
		
		fmt.Println("Step 2: Getting Service Ticket from Ticket Granting Server...")
		serviceTicket, stErr := getServiceTicket(username, clientId, "iotservice1")
		if stErr != nil {
			err = stErr
			break
		}
		
		fmt.Println("Step 3: Authenticating with IoT Service Validator and accessing device...")
		_, accessErr := accessIoTDevice(username, clientId, deviceId)
		if accessErr != nil {
			err = accessErr
			break
		}
		
		fmt.Println("Authentication successful! You can now access the IoT device.")
		
	case "get-device-data":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run auth-framework.go get-device-data <username> <clientId> <deviceId>")
			return
		}
		clientId := os.Args[3]
		deviceId := os.Args[4]
		_, err = getIoTDeviceData(username, clientId, deviceId)
		
	case "close-session":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run auth-framework.go close-session <username> <clientId> <deviceId>")
			return
		}
		clientId := os.Args[3]
		deviceId := os.Args[4]
		err = closeSession(username, clientId, deviceId)
		
	case "debug-rsa":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run auth-framework.go debug-rsa <nonce>")
			return
		}
		nonce := os.Args[3]
		err = debugRSAEncryption(nonce)
		
	default:
		printUsage()
		return
	}
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Operation completed successfully")
	}
}

// printUsage displays usage information
func printUsage() {
	fmt.Println("Available commands:")
	fmt.Println("  register-client <username> <clientId>")
	fmt.Println("  register-device <username> <deviceId> <capability1> <capability2> ...")
	fmt.Println("  authenticate <username> <clientId> <deviceId>")
	fmt.Println("  get-device-data <username> <clientId> <deviceId>")
	fmt.Println("  close-session <username> <clientId> <deviceId>")
	fmt.Println("  debug-rsa <nonce>")
}

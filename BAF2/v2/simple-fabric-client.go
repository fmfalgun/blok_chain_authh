// simple-fabric-client.go
// A simplified Go client for interacting with Hyperledger Fabric using the keys 
// generated by the standalone authentication framework

package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	keysDir         = "keys"
)

// getKeyPath returns the path to a key file
func getKeyPath(id, keyType string) string {
	if keyType == "private" {
		return filepath.Join(keysDir, fmt.Sprintf("%s-private.pem", id))
	}
	return filepath.Join(keysDir, fmt.Sprintf("%s-public.pem", id))
}

// loadPublicKeyPEM loads and returns a public key in PEM format
func loadPublicKeyPEM(id string) (string, error) {
	// Read the public key file
	publicKeyPath := getKeyPath(id, "public")
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read public key file: %v", err)
	}
	
	// Verify it's a proper PEM format
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the public key")
	}
	
	return string(publicKeyBytes), nil
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

// registerClient registers a client with the Authentication Server
func registerClient(username, clientId string) error {
	// Get the public key from the keys directory
	publicKeyPEM, err := loadPublicKeyPEM(clientId)
	if err != nil {
		return fmt.Errorf("failed to load public key: %v", err)
	}
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get AS contract
	contract := network.GetContract(asChaincodeId)
	
	// Register client with AS
	_, err = contract.SubmitTransaction("RegisterClient", clientId, publicKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to register client: %v", err)
	}
	
	fmt.Printf("Client %s registered successfully with Authentication Server\n", clientId)
	return nil
}

// registerIoTDevice registers an IoT device with capabilities
func registerIoTDevice(username, deviceId string, capabilities []string) error {
	// Get the public key from the keys directory
	publicKeyPEM, err := loadPublicKeyPEM(deviceId)
	if err != nil {
		return fmt.Errorf("failed to load public key: %v", err)
	}
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get ISV contract
	contract := network.GetContract(isvChaincodeId)
	
	// Convert capabilities to JSON
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %v", err)
	}
	
	// Try to register the device
	var registerErr error
	
	// First try submit transaction
	_, registerErr = contract.SubmitTransaction("RegisterIoTDevice", deviceId, publicKeyPEM, string(capabilitiesJSON))
	
	// If submission fails, try evaluation
	if registerErr != nil {
		fmt.Println("Transaction submission failed, falling back to evaluation...")
		_, registerErr = contract.EvaluateTransaction("RegisterIoTDevice", deviceId, publicKeyPEM, string(capabilitiesJSON))
	}
	
	if registerErr != nil {
		return fmt.Errorf("failed to register IoT device: %v", registerErr)
	}
	
	fmt.Printf("IoT device %s registered successfully with capabilities: %s\n", deviceId, strings.Join(capabilities, ", "))
	return nil
}

// getNonceChallenge gets a nonce challenge from the AS
func getNonceChallenge(username, clientId string) (string, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return "", fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get AS contract
	contract := network.GetContract(asChaincodeId)
	
	// Get the nonce challenge
	nonceResponseBytes, err := contract.SubmitTransaction("InitiateAuthentication", clientId)
	if err != nil {
		return "", fmt.Errorf("failed to initiate authentication: %v", err)
	}
	
	var nonceResponse map[string]string
	err = json.Unmarshal(nonceResponseBytes, &nonceResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse nonce challenge: %v", err)
	}
	
	nonce, ok := nonceResponse["nonce"]
	if !ok {
		return "", fmt.Errorf("nonce not found in response")
	}
	
	return nonce, nil
}

// verifyClientIdentity verifies a client identity using a signed nonce
func verifyClientIdentity(username, clientId, signedNonce string) error {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get AS contract
	contract := network.GetContract(asChaincodeId)
	
	// Verify client identity
	_, err = contract.SubmitTransaction("VerifyClientIdentityWithSignature", clientId, signedNonce)
	if err != nil {
		return fmt.Errorf("failed to verify client identity: %v", err)
	}
	
	return nil
}

// generateTGT generates a Ticket Granting Ticket
func generateTGT(username, clientId string) (string, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return "", fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get AS contract
	contract := network.GetContract(asChaincodeId)
	
	// Generate TGT
	tgtResponseBytes, err := contract.SubmitTransaction("GenerateTGT", clientId)
	if err != nil {
		return "", fmt.Errorf("failed to generate TGT: %v", err)
	}
	
	// Save TGT to file
	tgtFile := fmt.Sprintf("%s-tgt.json", clientId)
	err = ioutil.WriteFile(tgtFile, tgtResponseBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save TGT: %v", err)
	}
	
	return tgtFile, nil
}

// generateServiceTicket generates a service ticket
func generateServiceTicket(username, clientId, serviceId, tgtFile string) (string, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return "", fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get TGS contract
	contract := network.GetContract(tgsChaincodeId)
	
	// Load TGT
	tgtJSON, err := ioutil.ReadFile(tgtFile)
	if err != nil {
		return "", fmt.Errorf("failed to load TGT: %v", err)
	}
	
	var tgt map[string]string
	err = json.Unmarshal(tgtJSON, &tgt)
	if err != nil {
		return "", fmt.Errorf("failed to parse TGT: %v", err)
	}
	
	// Create authenticator
	authenticator := map[string]string{
		"clientID":  clientId,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	}
	
	authenticatorJSON, err := json.Marshal(authenticator)
	if err != nil {
		return "", fmt.Errorf("failed to marshal authenticator: %v", err)
	}
	
	// Base64 encode for simplicity
	encryptedAuthenticator := base64.StdEncoding.EncodeToString(authenticatorJSON)
	
	// Prepare service ticket request
	serviceTicketRequest := map[string]string{
		"encryptedTGT":  tgt["encryptedTGT"],
		"clientID":      clientId,
		"serviceID":     serviceId,
		"authenticator": encryptedAuthenticator,
	}
	
	serviceTicketRequestJSON, err := json.Marshal(serviceTicketRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal service ticket request: %v", err)
	}
	
	serviceTicketRequestB64 := base64.StdEncoding.EncodeToString(serviceTicketRequestJSON)
	
	// Submit request to TGS
	serviceTicketResponseBytes, err := contract.SubmitTransaction("GenerateServiceTicket", serviceTicketRequestB64)
	if err != nil {
		return "", fmt.Errorf("failed to generate service ticket: %v", err)
	}
	
	// Save service ticket
	serviceTicketFile := fmt.Sprintf("%s-serviceticket-%s.json", clientId, serviceId)
	err = ioutil.WriteFile(serviceTicketFile, serviceTicketResponseBytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save service ticket: %v", err)
	}
	
	return serviceTicketFile, nil
}

// validateServiceTicket validates a service ticket with ISV
func validateServiceTicket(username, serviceTicketFile string) error {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get ISV contract
	contract := network.GetContract(isvChaincodeId)
	
	// Load service ticket
	serviceTicketJSON, err := ioutil.ReadFile(serviceTicketFile)
	if err != nil {
		return fmt.Errorf("failed to load service ticket: %v", err)
	}
	
	var serviceTicket map[string]string
	err = json.Unmarshal(serviceTicketJSON, &serviceTicket)
	if err != nil {
		return fmt.Errorf("failed to parse service ticket: %v", err)
	}
	
	// Validate service ticket
	_, err = contract.SubmitTransaction("ValidateServiceTicket", serviceTicket["encryptedServiceTicket"])
	if err != nil {
		return fmt.Errorf("failed to validate service ticket: %v", err)
	}
	
	return nil
}

// processServiceRequest processes a service request
func processServiceRequest(username, clientId, deviceId, serviceTicketFile string) (string, error) {
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return "", fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get ISV contract
	contract := network.GetContract(isvChaincodeId)
	
	// Load service ticket
	serviceTicketJSON, err := ioutil.ReadFile(serviceTicketFile)
	if err != nil {
		return "", fmt.Errorf("failed to load service ticket: %v", err)
	}
	
	var serviceTicket map[string]string
	err = json.Unmarshal(serviceTicketJSON, &serviceTicket)
	if err != nil {
		return "", fmt.Errorf("failed to parse service ticket: %v", err)
	}
	
	// Prepare service request
	serviceRequest := map[string]string{
		"encryptedServiceTicket": serviceTicket["encryptedServiceTicket"],
		"clientID":               clientId,
		"deviceID":               deviceId,
		"requestType":            "read",
		"encryptedData":          base64.StdEncoding.EncodeToString([]byte("read-request")),
	}
	
	serviceRequestJSON, err := json.Marshal(serviceRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal service request: %v", err)
	}
	
	// Process service request
	serviceResponseBytes, err := contract.SubmitTransaction("ProcessServiceRequest", string(serviceRequestJSON))
	if err != nil {
		return "", fmt.Errorf("failed to process service request: %v", err)
	}
	
	var serviceResponse map[string]string
	err = json.Unmarshal(serviceResponseBytes, &serviceResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse service response: %v", err)
	}
	
	// Save session ID
	sessionFile := fmt.Sprintf("%s-session-%s.txt", clientId, deviceId)
	err = ioutil.WriteFile(sessionFile, []byte(serviceResponse["sessionID"]), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save session ID: %v", err)
	}
	
	return sessionFile, nil
}

// getIoTDeviceData gets data for an IoT device
func getIoTDeviceData(username, clientId, deviceId string) error {
	// Check if session exists
	sessionFile := fmt.Sprintf("%s-session-%s.txt", clientId, deviceId)
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("no active session found. Please authenticate first")
	}
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get ISV contract
	contract := network.GetContract(isvChaincodeId)
	
	// Query all IoT devices
	devicesResponseBytes, err := contract.EvaluateTransaction("GetAllIoTDevices")
	if err != nil {
		return fmt.Errorf("failed to get all IoT devices: %v", err)
	}
	
	var devices []map[string]interface{}
	err = json.Unmarshal(devicesResponseBytes, &devices)
	if err != nil {
		return fmt.Errorf("failed to parse devices: %v", err)
	}
	
	// Find the requested device
	found := false
	for _, device := range devices {
		if device["deviceID"] == deviceId {
			fmt.Printf("Device data for %s:\n", deviceId)
			fmt.Printf("  Device ID: %s\n", device["deviceID"])
			
			// Print capabilities
			capabilities, ok := device["capabilities"].([]interface{})
			if ok {
				fmt.Printf("  Capabilities: ")
				for i, cap := range capabilities {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Print(cap)
				}
				fmt.Println()
			}
			
			// Print public key (shortened for display)
			publicKey, ok := device["publicKey"].(string)
			if ok && len(publicKey) > 50 {
				fmt.Printf("  Public Key: %s...\n", publicKey[:50])
			} else {
				fmt.Printf("  Public Key: %s\n", publicKey)
			}
			
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("device %s not found", deviceId)
	}
	
	return nil
}

// closeSession closes an active session
func closeSession(username, clientId, deviceId string) error {
	// Check if session exists
	sessionFile := fmt.Sprintf("%s-session-%s.txt", clientId, deviceId)
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("no active session found")
	}
	
	// Read session ID
	sessionID, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session ID: %v", err)
	}
	
	// Connect to the network
	gw, network, err := connectToNetwork(username)
	if err != nil {
		return fmt.Errorf("failed to connect to network: %v", err)
	}
	defer gw.Close()
	
	// Get ISV contract
	contract := network.GetContract(isvChaincodeId)
	
	// Close the session
	_, err = contract.SubmitTransaction("CloseSession", string(sessionID))
	if err != nil {
		return fmt.Errorf("failed to close session: %v", err)
	}
	
	// Remove session file
	err = os.Remove(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to remove session file: %v", err)
	}
	
	fmt.Printf("Closed session for device %s\n", deviceId)
	return nil
}

// authenticate performs the full authentication flow
func authenticate(username, clientId, deviceId string) error {
	// Step 1: Get nonce challenge from AS
	fmt.Println("Step 1: Getting nonce challenge from Authentication Server...")
	nonce, err := getNonceChallenge(username, clientId)
	if err != nil {
		return fmt.Errorf("failed to get nonce challenge: %v", err)
	}
	
	// Step 2: Use the standalone framework to sign the nonce
	fmt.Println("Step 2: Signing the nonce with client's private key...")
	cmd := exec.Command("go", "run", "standalone-auth-framework.go", "simulate-auth", clientId, nonce)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to sign nonce: %v\n%s", err, output)
	}
	
	// Extract signed nonce from output (this is a bit hacky, might need adjustment)
	outputStr := string(output)
	signedNonceIdx := strings.Index(outputStr, "Signed nonce (base64): ")
	if signedNonceIdx == -1 {
		return fmt.Errorf("could not find signed nonce in output")
	}
	
	signedNonceLine := outputStr[signedNonceIdx:]
	signedNonceLine = signedNonceLine[:strings.Index(signedNonceLine, "\n")]
	signedNonce := strings.TrimPrefix(signedNonceLine, "Signed nonce (base64): ")
	
	// Step 3: Verify client identity
	fmt.Println("Step 3: Verifying client identity with Authentication Server...")
	err = verifyClientIdentity(username, clientId, signedNonce)
	if err != nil {
		return fmt.Errorf("failed to verify client identity: %v", err)
	}
	
	// Step 4: Generate TGT
	fmt.Println("Step 4: Getting Ticket Granting Ticket (TGT)...")
	tgtFile, err := generateTGT(username, clientId)
	if err != nil {
		return fmt.Errorf("failed to generate TGT: %v", err)
	}
	
	// Step 5: Generate Service Ticket
	fmt.Println("Step 5: Getting Service Ticket from Ticket Granting Server...")
	serviceTicketFile, err := generateServiceTicket(username, clientId, "iotservice1", tgtFile)
	if err != nil {
		return fmt.Errorf("failed to generate service ticket: %v", err)
	}
	
	// Step 6: Validate Service Ticket
	fmt.Println("Step 6: Validating Service Ticket with IoT Service Validator...")
	err = validateServiceTicket(username, serviceTicketFile)
	if err != nil {
		return fmt.Errorf("failed to validate service ticket: %v", err)
	}
	
	// Step 7: Process Service Request
	fmt.Println("Step 7: Processing service request to access IoT device...")
	sessionFile, err := processServiceRequest(username, clientId, deviceId, serviceTicketFile)
	if err != nil {
		return fmt.Errorf("failed to process service request: %v", err)
	}
	
	fmt.Printf("Authentication successful! Session established and saved to %s\n", sessionFile)
	return nil
}

func showUsage() {
	fmt.Println("Simple Fabric Client for Authentication Framework")
	fmt.Println("Usage: go run simple-fabric-client.go COMMAND [OPTIONS]")
	fmt.Println("Commands:")
	fmt.Println("  register-client <username> <clientId>               - Register client with AS")
	fmt.Println("  register-device <username> <deviceId> <capabilities> - Register IoT device with ISV")
	fmt.Println("  authenticate <username> <clientId> <deviceId>       - Authenticate client for device access")
	fmt.Println("  get-device-data <username> <clientId> <deviceId>    - Get device data after authentication")
	fmt.Println("  close-session <username> <clientId> <deviceId>      - Close an active session")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run simple-fabric-client.go register-client admin client1")
	fmt.Println("  go run simple-fabric-client.go register-device admin device1 temperature humidity")
	fmt.Println("  go run simple-fabric-client.go authenticate admin client1 device1")
}

// Initialize checks if the wallet exists and has the required identities
func initializeWallet(username string) error {
	fmt.Printf("Checking for identity %s in wallet...\n", username)

	// Check if wallet exists
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		fmt.Printf("Wallet directory not found, creating: %s\n", walletPath)
		if err := os.MkdirAll(walletPath, 0755); err != nil {
			return fmt.Errorf("failed to create wallet directory: %v", err)
		}
	}

	// Check if identity exists
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %v", err)
	}

	if wallet.Exists(username) {
		fmt.Printf("Identity %s already exists in wallet\n", username)
		return nil
	}

	fmt.Printf("Identity %s not found in wallet\n", username)
	fmt.Println("Attempting to locate and import credentials...")

	// Common paths to check for credentials
	mspPaths := []string{
		"./certs/msp",
		"./certs/org1/msp",
		filepath.Join(os.Getenv("HOME"), ".fabric-ca-client"),
	}

	// Check common certificate locations
	certificatePaths := []string{
		fmt.Sprintf("./certs/org1/%s.crt", username),
		fmt.Sprintf("./certs/%s.crt", username),
		"./certs/org1/admin.crt",
		"./certs/admin.crt",
	}

	// Check common key locations
	keyPaths := []string{
		fmt.Sprintf("./certs/org1/%s.key", username),
		fmt.Sprintf("./certs/%s.key", username),
		"./certs/org1/admin.key",
		"./certs/admin.key",
	}

	// Try to find MSP with signcerts and keystore directories
	for _, mspPath := range mspPaths {
		if !fileExists(mspPath) {
			continue
		}

		fmt.Printf("Checking MSP directory: %s\n", mspPath)

		// Look for signcerts
		signcertsPath := filepath.Join(mspPath, "signcerts")
		if !fileExists(signcertsPath) {
			continue
		}

		// Look for keystore
		keystorePath := filepath.Join(mspPath, "keystore")
		if !fileExists(keystorePath) {
			continue
		}

		// Get certificate files
		certFiles, err := filepath.Glob(filepath.Join(signcertsPath, "*.pem"))
		if err != nil || len(certFiles) == 0 {
			// Try .crt extension
			certFiles, err = filepath.Glob(filepath.Join(signcertsPath, "*.crt"))
			if err != nil || len(certFiles) == 0 {
				continue
			}
		}

		// Get key files
		keyFiles, err := filepath.Glob(filepath.Join(keystorePath, "*_sk"))
		if err != nil || len(keyFiles) == 0 {
			// Try .key extension
			keyFiles, err = filepath.Glob(filepath.Join(keystorePath, "*.key"))
			if err != nil || len(keyFiles) == 0 {
				continue
			}
		}

		if len(certFiles) > 0 && len(keyFiles) > 0 {
			cert, err := ioutil.ReadFile(certFiles[0])
			if err != nil {
				fmt.Printf("Error reading certificate: %v\n", err)
				continue
			}

			key, err := ioutil.ReadFile(keyFiles[0])
			if err != nil {
				fmt.Printf("Error reading key: %v\n", err)
				continue
			}

			// Create identity with Org1MSP as default
			identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

			// Add to wallet
			err = wallet.Put(username, identity)
			if err != nil {
				fmt.Printf("Error adding identity to wallet: %v\n", err)
				continue
			}

			fmt.Printf("Successfully imported %s identity from %s\n", username, mspPath)
			return nil
		}
	}

	// Try certificate and key files directly
	for i, certPath := range certificatePaths {
		if !fileExists(certPath) {
			continue
		}

		// Try to find a matching key
		keyPath := ""
		if i < len(keyPaths) {
			keyPath = keyPaths[i]
			if !fileExists(keyPath) {
				continue
			}
		} else {
			// If we don't have a matching key path, try all key paths
			for _, kp := range keyPaths {
				if fileExists(kp) {
					keyPath = kp
					break
				}
			}
		}

		if keyPath == "" {
			continue
		}

		fmt.Printf("Found certificate: %s\n", certPath)
		fmt.Printf("Found key: %s\n", keyPath)

		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			fmt.Printf("Error reading certificate: %v\n", err)
			continue
		}

		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			fmt.Printf("Error reading key: %v\n", err)
			continue
		}

		// Create identity with Org1MSP as default
		identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

		// Add to wallet
		err = wallet.Put(username, identity)
		if err != nil {
			fmt.Printf("Error adding identity to wallet: %v\n", err)
			continue
		}

		fmt.Printf("Successfully imported %s identity from %s and %s\n", username, certPath, keyPath)
		return nil
	}

	// If we reach here, let the user manually specify
	fmt.Println("\nCould not automatically locate credentials.")
	fmt.Println("Please provide the path to the certificate and key files for the identity:")

	fmt.Print("Certificate path: ")
	var certPath string
	fmt.Scanln(&certPath)

	fmt.Print("Key path: ")
	var keyPath string
	fmt.Scanln(&keyPath)

	fmt.Print("MSP ID (default: Org1MSP): ")
	var mspID string
	fmt.Scanln(&mspID)
	if mspID == "" {
		mspID = "Org1MSP"
	}

	if certPath == "" || keyPath == "" {
		return fmt.Errorf("certificate and key paths are required")
	}

	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %v", err)
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key: %v", err)
	}

	// Create identity
	identity := gateway.NewX509Identity(mspID, string(cert), string(key))

	// Add to wallet
	err = wallet.Put(username, identity)
	if err != nil {
		return fmt.Errorf("failed to add identity to wallet: %v", err)
	}

	fmt.Printf("Successfully imported %s identity\n", username)
	return nil
}

// Helper function to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func main() {
	// Initialize wallet before any operations
	if len(os.Args) > 2 {
		err := initializeWallet(os.Args[2]) // Use the username from command line
		if err != nil {
			fmt.Printf("Error initializing wallet: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Default to admin if no username provided
		err := initializeWallet("admin")
		if err != nil {
			fmt.Printf("Error initializing wallet: %v\n", err)
			os.Exit(1)
		}
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		showUsage()
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "register-client":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run simple-fabric-client.go register-client <username> <clientId>")
			return
		}
		username := os.Args[2]
		clientId := os.Args[3]
		err := registerClient(username, clientId)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
	case "register-device":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run simple-fabric-client.go register-device <username> <deviceId> <capability1> <capability2> ...")
			return
		}
		username := os.Args[2]
		deviceId := os.Args[3]
		capabilities := os.Args[4:]
		err := registerIoTDevice(username, deviceId, capabilities)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
	case "authenticate":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run simple-fabric-client.go authenticate <username> <clientId> <deviceId>")
			return
		}
		username := os.Args[2]
		clientId := os.Args[3]
		deviceId := os.Args[4]
		err := authenticate(username, clientId, deviceId)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
	case "get-device-data":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run simple-fabric-client.go get-device-data <username> <clientId> <deviceId>")
			return
		}
		username := os.Args[2]
		clientId := os.Args[3]
		deviceId := os.Args[4]
		err := getIoTDeviceData(username, clientId, deviceId)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
	case "close-session":
		if len(os.Args) < 5 {
			fmt.Println("Usage: go run simple-fabric-client.go close-session <username> <clientId> <deviceId>")
			return
		}
		username := os.Args[2]
		clientId := os.Args[3]
		deviceId := os.Args[4]
		err := closeSession(username, clientId, deviceId)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
	
	fmt.Println("Operation completed successfully")
}

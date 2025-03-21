package auth

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/chaichis-network/v3/internal/crypto"
	"github.com/chaichis-network/v3/internal/fabric"
	"github.com/chaichis-network/v3/pkg/logger"
	"github.com/pkg/errors"
)

var log = logger.Default()

// ClientManager manages client authentication operations
type ClientManager struct {
	fabricClient *fabric.Client
	asContract   *fabric.AuthServerContract
	tgsContract  *fabric.TicketGrantingContract
	identity     string
}

// NewClientManager creates a new client manager
func NewClientManager(fabricClient *fabric.Client, identity string) (*ClientManager, error) {
	// Ensure client is connected
	if err := fabricClient.Connect(identity); err != nil {
		return nil, errors.Wrap(err, "failed to connect to Fabric network")
	}
	
	// Get contracts
	asContract, err := fabric.NewAuthServerContract(fabricClient)
	if err != nil {
		return nil, err
	}
	
	tgsContract, err := fabric.NewTicketGrantingContract(fabricClient)
	if err != nil {
		return nil, err
	}
	
	return &ClientManager{
		fabricClient: fabricClient,
		asContract:   asContract,
		tgsContract:  tgsContract,
		identity:     identity,
	}, nil
}

// RegisterClient registers a new client with the Authentication Server
func (cm *ClientManager) RegisterClient(clientID string) error {
	// Generate or load client keys
	_, _, err := crypto.LoadOrGenerateKeys(clientID)
	if err != nil {
		return errors.Wrap(err, "failed to load or generate client keys")
	}
	
	// Get client's public key PEM
	publicKeyPEM, err := crypto.GetPublicKeyPEM(clientID)
	if err != nil {
		return errors.Wrap(err, "failed to get client's public key PEM")
	}
	
	// Register client with AS
	if err := cm.asContract.RegisterClient(clientID, publicKeyPEM); err != nil {
		return errors.Wrap(err, "failed to register client with Authentication Server")
	}
	
	log.Infof("Client %s registered successfully with Authentication Server", clientID)
	return nil
}

// Authenticate performs the full authentication flow for a client
func (cm *ClientManager) Authenticate(clientID, deviceID string) error {
	log.Infof("Starting authentication flow for client %s to access device %s", clientID, deviceID)
	
	// Step 1: Get nonce challenge from AS
	log.Info("Step 1: Getting nonce challenge from Authentication Server...")
	nonce, err := cm.asContract.GetNonceChallenge(clientID)
	if err != nil {
		return errors.Wrap(err, "failed to get nonce challenge")
	}
	
	// Step 2: Sign the nonce
	log.Info("Step 2: Signing nonce with client's private key...")
	signedNonce, err := crypto.SignNonce(clientID, nonce)
	if err != nil {
		return errors.Wrap(err, "failed to sign nonce")
	}
	
	// Step 3: Verify client identity
	log.Info("Step 3: Verifying client identity with Authentication Server...")
	if err := cm.asContract.VerifyClientIdentity(clientID, signedNonce); err != nil {
		return errors.Wrap(err, "failed to verify client identity")
	}
	
	// Step 4: Generate TGT
	log.Info("Step 4: Getting Ticket Granting Ticket (TGT)...")
	tgt, err := cm.asContract.GenerateTGT(clientID)
	if err != nil {
		return errors.Wrap(err, "failed to generate TGT")
	}
	
	// Save TGT to file
	tgtFile := clientID + "-tgt.json"
	tgtJSON, err := json.Marshal(tgt)
	if err != nil {
		return errors.Wrap(err, "failed to marshal TGT")
	}
	if err := ioutil.WriteFile(tgtFile, tgtJSON, 0600); err != nil {
		return errors.Wrap(err, "failed to save TGT to file")
	}
	
	// Step 5: Generate Service Ticket
	log.Info("Step 5: Getting Service Ticket from TGS...")
	serviceID := "iotservice1" // Default service ID
	
	// Create authenticator (timestamp encrypted with session key)
	// In a real implementation, this would be properly encrypted
	// For now, we'll use a simpler approach
	authenticator := Authenticator{
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
	}
	authenticatorJSON, err := json.Marshal(authenticator)
	if err != nil {
		return errors.Wrap(err, "failed to marshal authenticator")
	}
	
	authenticatorB64 := base64.StdEncoding.EncodeToString(authenticatorJSON)
	
	// Create service ticket request
	serviceTicketRequest := ServiceTicketRequest{
		EncryptedTGT:  tgt["encryptedTGT"],
		ClientID:      clientID,
		ServiceID:     serviceID,
		Authenticator: authenticatorB64,
	}
	
	// Convert request to map for contract
	requestMap := map[string]string{
		"encryptedTGT":  serviceTicketRequest.EncryptedTGT,
		"clientID":      serviceTicketRequest.ClientID,
		"serviceID":     serviceTicketRequest.ServiceID,
		"authenticator": serviceTicketRequest.Authenticator,
	}
	
	// Get service ticket
	serviceTicket, err := cm.tgsContract.GenerateServiceTicket(requestMap)
	if err != nil {
		return errors.Wrap(err, "failed to generate service ticket")
	}
	
	// Save service ticket to file
	serviceTicketFile := clientID + "-serviceticket-" + deviceID + ".json"
	serviceTicketJSON, err := json.Marshal(serviceTicket)
	if err != nil {
		return errors.Wrap(err, "failed to marshal service ticket")
	}
	if err := ioutil.WriteFile(serviceTicketFile, serviceTicketJSON, 0600); err != nil {
		return errors.Wrap(err, "failed to save service ticket to file")
	}
	
	log.Infof("Authentication successful! Service ticket saved to %s", serviceTicketFile)
	return nil
}

// GetTGT retrieves a saved TGT for a client
func (cm *ClientManager) GetTGT(clientID string) (map[string]string, error) {
	tgtFile := clientID + "-tgt.json"
	
	// Check if TGT file exists
	if _, err := os.Stat(tgtFile); os.IsNotExist(err) {
		return nil, errors.New("TGT not found, please authenticate first")
	}
	
	// Read TGT file
	tgtJSON, err := ioutil.ReadFile(tgtFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read TGT file")
	}
	
	// Parse TGT
	var tgt map[string]string
	if err := json.Unmarshal(tgtJSON, &tgt); err != nil {
		return nil, errors.Wrap(err, "failed to parse TGT")
	}
	
	return tgt, nil
}

// GetServiceTicket retrieves a saved service ticket for a client and device
func (cm *ClientManager) GetServiceTicket(clientID, deviceID string) (map[string]string, error) {
	serviceTicketFile := clientID + "-serviceticket-" + deviceID + ".json"
	
	// Check if service ticket file exists
	if _, err := os.Stat(serviceTicketFile); os.IsNotExist(err) {
		return nil, errors.New("service ticket not found, please authenticate first")
	}
	
	// Read service ticket file
	serviceTicketJSON, err := ioutil.ReadFile(serviceTicketFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read service ticket file")
	}
	
	// Parse service ticket
	var serviceTicket map[string]string
	if err := json.Unmarshal(serviceTicketJSON, &serviceTicket); err != nil {
		return nil, errors.Wrap(err, "failed to parse service ticket")
	}
	
	return serviceTicket, nil
}

// Close closes the connection to the Fabric network
func (cm *ClientManager) Close() {
	cm.fabricClient.Close()
}

package fabric

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

// Initialize logging
var logger = logging.NewLogger("fabricClient")

// FabricClient manages connections to the Fabric network
type FabricClient struct {
	wallet          *gateway.Wallet
	gateway         *gateway.Gateway
	network         *gateway.Network
	connectionProfile string
	org             string
	user            string
	channel         string
}

// NewFabricClient creates a new Fabric client
func NewFabricClient(connectionProfile, org, user, channel string) (*FabricClient, error) {
	logger.Infof("Creating new Fabric client with profile: %s, org: %s, user: %s, channel: %s", 
		connectionProfile, org, user, channel)
	
	// Validate file existence first
	if _, err := os.Stat(connectionProfile); os.IsNotExist(err) {
		return nil, fmt.Errorf("connection profile not found: %s", connectionProfile)
	}
	
	// Read the file to verify it's valid JSON
	content, err := ioutil.ReadFile(connectionProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read connection profile: %s", err)
	}
	logger.Infof("Successfully read connection profile (%d bytes)", len(content))
	
	return &FabricClient{
		connectionProfile: connectionProfile,
		org:              org,
		user:             user,
		channel:          channel,
	}, nil
}

// Connect establishes a connection to the Fabric network
func (fc *FabricClient) Connect() error {
	var err error
	
	// Create a new file system wallet for managing identities
	walletPath := filepath.Join("wallet", fc.org)
	logger.Infof("Creating wallet at: %s", walletPath)
	
	if fc.wallet, err = gateway.NewFileSystemWallet(walletPath); err != nil {
		return fmt.Errorf("failed to create wallet: %s", err)
	}
	
	// Check if the user identity exists in the wallet
	if !fc.identityExists() {
		return fmt.Errorf("identity '%s' not found in wallet", fc.user)
	}
	logger.Infof("Successfully found identity '%s' in wallet", fc.user)
	
	// Create the gateway connection
	logger.Infof("Connecting to gateway with connection profile: %s", fc.connectionProfile)
	
	// Get absolute path to connection profile
	absPath, err := filepath.Abs(fc.connectionProfile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for connection profile: %s", err)
	}
	logger.Infof("Using absolute path for connection profile: %s", absPath)
	
	// Print all environment variables for debugging
	logger.Infof("Current environment variables:")
	for _, env := range os.Environ() {
		logger.Infof("  %s", env)
	}
	
	// Configure connection options
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(absPath)),
		gateway.WithIdentity(fc.wallet, fc.user),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %s", err)
	}
	fc.gateway = gw
	logger.Infof("Successfully connected to gateway")
	
	// Get network
	if fc.network, err = fc.gateway.GetNetwork(fc.channel); err != nil {
		fc.gateway.Close()
		return fmt.Errorf("failed to get network: %s", err)
	}
	logger.Infof("Successfully connected to channel: %s", fc.channel)
	
	return nil
}

// Close closes the connection to the Fabric network
func (fc *FabricClient) Close() {
	if fc.gateway != nil {
		fc.gateway.Close()
		logger.Infof("Gateway connection closed")
	}
}

// identityExists checks if the user identity exists in the wallet
func (fc *FabricClient) identityExists() bool {
	exists, err := fc.wallet.Exists(fc.user)
	if err != nil {
		logger.Errorf("Failed to check if identity exists: %s", err)
		return false
	}
	return exists
}

// GetContract returns a contract for the specified chaincode
func (fc *FabricClient) GetContract(chaincodeName string) (*gateway.Contract, error) {
	if fc.network == nil {
		return nil, fmt.Errorf("not connected to network")
	}
	
	logger.Infof("Getting contract for chaincode: %s", chaincodeName)
	contract := fc.network.GetContract(chaincodeName)
	return contract, nil
}

// ExecuteTransaction executes a transaction on the specified chaincode
func (fc *FabricClient) ExecuteTransaction(chaincodeName, function string, args ...string) ([]byte, error) {
	contract, err := fc.GetContract(chaincodeName)
	if err != nil {
		return nil, err
	}
	
	logger.Infof("Executing transaction: chaincode=%s, function=%s, args=%v", 
		chaincodeName, function, args)
	
	// Convert string args to []byte args
	byteArgs := make([][]byte, len(args))
	for i, arg := range args {
		byteArgs[i] = []byte(arg)
	}
	
	// Execute the transaction
	result, err := contract.SubmitTransaction(function, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %s", err)
	}
	
	logger.Infof("Transaction executed successfully")
	return result, nil
}

// QueryTransaction evaluates a transaction on the specified chaincode
func (fc *FabricClient) QueryTransaction(chaincodeName, function string, args ...string) ([]byte, error) {
	contract, err := fc.GetContract(chaincodeName)
	if err != nil {
		return nil, err
	}
	
	logger.Infof("Querying transaction: chaincode=%s, function=%s, args=%v", 
		chaincodeName, function, args)
	
	// Execute the query
	result, err := contract.EvaluateTransaction(function, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %s", err)
	}
	
	logger.Infof("Query executed successfully")
	return result, nil
}

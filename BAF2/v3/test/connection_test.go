package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	// Configure logging
	logging.SetLevel("", logging.INFO)
	
	// Parse command line arguments
	connectionProfile := flag.String("profile", "config/connection-profile-simple.json", "Path to connection profile")
	org := flag.String("org", "Org1", "Organization")
	user := flag.String("user", "admin", "User")
	channel := flag.String("channel", "mychannel", "Channel")
	chaincode := flag.String("chaincode", "as_chaincode_1.1", "Chaincode name")
	flag.Parse()
	
	fmt.Printf("Testing connection to Fabric network:\n")
	fmt.Printf("  Connection Profile: %s\n", *connectionProfile)
	fmt.Printf("  Organization: %s\n", *org)
	fmt.Printf("  User: %s\n", *user)
	fmt.Printf("  Channel: %s\n", *channel)
	fmt.Printf("  Chaincode: %s\n", *chaincode)
	
	// Check if the connection profile exists
	if _, err := os.Stat(*connectionProfile); os.IsNotExist(err) {
		fmt.Printf("Error: Connection profile not found: %s\n", *connectionProfile)
		os.Exit(1)
	}
	
	// Create a new file system wallet for managing identities
	walletPath := filepath.Join("wallet", *org)
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}
	
	// Check if the user identity exists in the wallet
	exists, err := wallet.Exists(*user)
	if err != nil {
		fmt.Printf("Failed to check if identity exists: %s\n", err)
		os.Exit(1)
	}
	if !exists {
		fmt.Printf("Identity '%s' not found in wallet at %s\n", *user, walletPath)
		fmt.Printf("Contents of wallet directory:\n")
		files, err := os.ReadDir(walletPath)
		if err != nil {
			fmt.Printf("  Error reading wallet directory: %s\n", err)
		} else {
			for _, file := range files {
				fmt.Printf("  %s\n", file.Name())
			}
		}
		os.Exit(1)
	}
	fmt.Printf("Successfully found identity '%s' in wallet\n", *user)
	
	// Get absolute path to connection profile
	absPath, err := filepath.Abs(*connectionProfile)
	if err != nil {
		fmt.Printf("Failed to get absolute path for connection profile: %s\n", err)
		os.Exit(1)
	}
	
	// Configure connection options with error handling
	fmt.Printf("Connecting to gateway with connection profile: %s\n", absPath)
	
	// Try with multiple connection options
	var gw *gateway.Gateway
	
	// Try first with FromFile
	fmt.Printf("Trying connection with config.FromFile...\n")
	gw, err = gateway.Connect(
		gateway.WithConfig(config.FromFile(absPath)),
		gateway.WithIdentity(wallet, *user),
	)
	if err != nil {
		fmt.Printf("Failed to connect with config.FromFile: %s\n", err)
		
		// Try with FromFile and allowing insecure connections
		fmt.Printf("Trying connection with additional options...\n")
		gw, err = gateway.Connect(
			gateway.WithConfig(config.FromFile(absPath)),
			gateway.WithIdentity(wallet, *user),
			gateway.WithTimeout(60),
		)
		if err != nil {
			fmt.Printf("Failed to connect with additional options: %s\n", err)
			os.Exit(1)
		}
	}
	defer gw.Close()
	
	fmt.Printf("Successfully connected to gateway\n")
	
	// Get network
	network, err := gw.GetNetwork(*channel)
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully connected to channel: %s\n", *channel)
	
	// Get contract
	contract := network.GetContract(*chaincode)
	fmt.Printf("Successfully got contract for chaincode: %s\n", *chaincode)
	
	// Try a simple query to ensure everything is working
	fmt.Printf("Testing query on chaincode...\n")
	result, err := contract.EvaluateTransaction("ping")
	if err != nil {
		fmt.Printf("Failed to query chaincode: %s\n", err)
		// Try an alternative function if "ping" doesn't exist
		fmt.Printf("Trying alternative query...\n")
		result, err = contract.EvaluateTransaction("getInfo")
		if err != nil {
			fmt.Printf("Failed to query with alternative function: %s\n", err)
			os.Exit(1)
		}
	}
	
	fmt.Printf("Successfully queried chaincode, result: %s\n", string(result))
	fmt.Printf("Connection test completed successfully!\n")
}

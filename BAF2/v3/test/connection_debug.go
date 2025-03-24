package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	// Print working directory
	dir, _ := os.Getwd()
	fmt.Println("Working directory:", dir)
	
	// Path to connection profile
	ccpPath := "config/connection-profile.json"
	fmt.Println("Connection profile path:", filepath.Join(dir, ccpPath))
	
	// Check if the file exists
	if _, err := os.Stat(ccpPath); os.IsNotExist(err) {
		log.Fatalf("Connection profile not found at %s", ccpPath)
	}
	
	// Read the connection profile
	ccpBytes, err := ioutil.ReadFile(ccpPath)
	if err != nil {
		log.Fatalf("Failed to read connection profile: %v", err)
	}
	fmt.Printf("Connection profile size: %d bytes\n", len(ccpBytes))
	
	// Create a wallet for identity management
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	
	// Check if admin identity exists
	if !wallet.Exists("admin") {
		log.Fatalf("Admin identity not found in wallet")
	}
	fmt.Println("Admin identity found in wallet")
	
	// Try to connect to gateway
	fmt.Println("Attempting to connect to Fabric gateway...")
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(ccpPath)),
		gateway.WithIdentity(wallet, "admin"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	
	defer gw.Close()
	fmt.Println("Successfully connected to Fabric gateway!")
	
	// Try to get the network
	network, err := gw.GetNetwork("chaichis-channel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}
	fmt.Println("Successfully connected to network 'chaichis-channel'")
	
	// Try to get the AS contract
	contract := network.GetContract("as_chaincode_1.1")
	fmt.Printf("Successfully got contract 'as_chaincode_1.1'\n")
	
	// Try a simple query
	fmt.Println("Attempting to query contract...")
	response, err := contract.EvaluateTransaction("GetAllClientRegistrations")
	if err != nil {
		log.Fatalf("Failed to query contract: %v", err)
	}
	
	fmt.Printf("Query response: %s\n", string(response))
	fmt.Println("Test completed successfully!")
}

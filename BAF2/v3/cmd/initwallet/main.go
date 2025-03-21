package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const walletPath = "wallet"

func main() {
	fmt.Println("Wallet Initialization Tool")
	fmt.Println("=========================")
	
	// Create wallet directory if it doesn't exist
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		os.MkdirAll(walletPath, 0755)
		fmt.Printf("Created wallet directory: %s\n", walletPath)
	}
	
	// Create wallet
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		fmt.Printf("Failed to create wallet: %v\n", err)
		os.Exit(1)
	}
	
	// Get certificate and key paths
	var certPath, keyPath, mspID string
	
	fmt.Println("Please provide the certificate path (PEM format):")
	fmt.Scanln(&certPath)
	
	fmt.Println("Please provide the key path (PEM format):")
	fmt.Scanln(&keyPath)
	
	fmt.Println("Please provide the MSP ID (e.g., Org1MSP):")
	fmt.Scanln(&mspID)
	
	if certPath == "" || keyPath == "" {
		fmt.Println("Certificate and key paths are required")
		os.Exit(1)
	}
	
	if mspID == "" {
		mspID = "Org1MSP"
		fmt.Printf("Using default MSP ID: %s\n", mspID)
	}
	
	// Read certificate file
	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		fmt.Printf("Failed to read certificate file: %v\n", err)
		os.Exit(1)
	}
	
	// Read key file
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		fmt.Printf("Failed to read key file: %v\n", err)
		os.Exit(1)
	}
	
	// Create identity
	identity := gateway.NewX509Identity(mspID, string(cert), string(key))
	
	// Add to wallet
	err = wallet.Put("admin", identity)
	if err != nil {
		fmt.Printf("Failed to put identity into wallet: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Successfully imported admin identity")
}

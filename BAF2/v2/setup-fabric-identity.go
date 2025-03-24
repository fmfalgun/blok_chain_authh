// setup-fabric-identity.go
// A utility to set up identities for the Fabric client by checking for existing credentials

package main

import (
	//"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

const (
	walletPath     = "wallet"
	certsDirectory = "certs"  // Directory where your Fabric certificates are stored - adjust as needed
	mspDirectories = "./certs/msp,~/.fabric-ca-client,/etc/hyperledger/fabric/msp"  // Common MSP directories
)

func setupWallet() error {
	fmt.Println("Setting up Fabric wallet...")
	
	// Create wallet directory if it doesn't exist
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		if err := os.MkdirAll(walletPath, 0755); err != nil {
			return fmt.Errorf("failed to create wallet directory: %v", err)
		}
		fmt.Println("Created wallet directory")
	}
	
	// Check if wallet already has identities
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %v", err)
	}

	// Check for admin identity
	if wallet.Exists("admin") {
		fmt.Println("Admin identity already exists in wallet")
		return nil
	}
	
	// Look for credentials in common MSP directories
	mspPaths := strings.Split(mspDirectories, ",")
	for _, mspPath := range mspPaths {
		mspPath = strings.TrimSpace(mspPath)
		// Expand home directory if needed
		if strings.HasPrefix(mspPath, "~/") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Warning: Could not expand home directory: %v\n", err)
				continue
			}
			mspPath = filepath.Join(homeDir, mspPath[2:])
		}
		
		fmt.Printf("Checking MSP directory: %s\n", mspPath)
		
		// Check if the directory exists
		if _, err := os.Stat(mspPath); os.IsNotExist(err) {
			fmt.Printf("Directory does not exist: %s\n", mspPath)
			continue
		}
		
		// Try to find credentials - this would need customization based on your setup
		err = findAndImportCredentials(wallet, mspPath)
		if err != nil {
			fmt.Printf("Warning: Could not import credentials from %s: %v\n", mspPath, err)
			continue
		}
		
		// Check if admin was imported
		if wallet.Exists("admin") {
			fmt.Println("Successfully imported admin identity")
			return nil
		}
	}
	
	// Look in the certs directory (if exists)
	if _, err := os.Stat(certsDirectory); !os.IsNotExist(err) {
		fmt.Printf("Checking certs directory: %s\n", certsDirectory)
		
		// Look for org1 admin credentials
		org1Dir := filepath.Join(certsDirectory, "org1")
		if _, err := os.Stat(org1Dir); !os.IsNotExist(err) {
			adminKeyPath := filepath.Join(org1Dir, "admin.key")
			adminCertPath := filepath.Join(org1Dir, "admin.crt")
			
			if fileExists(adminKeyPath) && fileExists(adminCertPath) {
				fmt.Println("Found admin key and certificate in certs directory")
				
				// Read key and certificate
				privateKey, err := ioutil.ReadFile(adminKeyPath)
				if err != nil {
					return fmt.Errorf("failed to read admin key: %v", err)
				}
				
				certificate, err := ioutil.ReadFile(adminCertPath)
				if err != nil {
					return fmt.Errorf("failed to read admin certificate: %v", err)
				}
				
				// Create identity
				identity := gateway.NewX509Identity("Org1MSP", string(certificate), string(privateKey))
				
				// Add to wallet
				err = wallet.Put("admin", identity)
				if err != nil {
					return fmt.Errorf("failed to put admin identity into wallet: %v", err)
				}
				
				fmt.Println("Successfully imported admin identity")
				return nil
			}
		}
	}
	
	// If we reach here, we couldn't find/import admin credentials
	fmt.Println("\nCould not find or import admin credentials from standard locations.")
	fmt.Println("Please provide the path to your admin certificate and private key:")
	fmt.Println("1. Certificate path (PEM format):")
	var certPath string
	fmt.Scanln(&certPath)
	
	fmt.Println("2. Private key path (PEM format):")
	var keyPath string
	fmt.Scanln(&keyPath)
	
	fmt.Println("3. MSP ID (e.g., Org1MSP):")
	var mspID string
	fmt.Scanln(&mspID)
	
	if certPath == "" || keyPath == "" || mspID == "" {
		return fmt.Errorf("certificate path, key path, and MSP ID are required")
	}
	
	// Read key and certificate
	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}
	
	certificate, err := ioutil.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %v", err)
	}
	
	// Create identity
	identity := gateway.NewX509Identity(mspID, string(certificate), string(privateKey))
	
	// Add to wallet
	err = wallet.Put("admin", identity)
	if err != nil {
		return fmt.Errorf("failed to put admin identity into wallet: %v", err)
	}
	
	fmt.Println("Successfully imported admin identity")
	return nil
}

func findAndImportCredentials(wallet *gateway.Wallet, mspPath string) error {
	// This is a simplified version that looks for typical Fabric MSP structure
	// You may need to customize this for your specific setup
	
	// Look for signcerts and keystore directories
	signcertsPath := filepath.Join(mspPath, "signcerts")
	keystorePath := filepath.Join(mspPath, "keystore")
	
	// Check if directories exist
	if _, err := os.Stat(signcertsPath); os.IsNotExist(err) {
		return fmt.Errorf("signcerts directory not found: %s", signcertsPath)
	}
	
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return fmt.Errorf("keystore directory not found: %s", keystorePath)
	}
	
	// Get the first certificate file in signcerts
	certFiles, err := ioutil.ReadDir(signcertsPath)
	if err != nil {
		return fmt.Errorf("failed to read signcerts directory: %v", err)
	}
	
	if len(certFiles) == 0 {
		return fmt.Errorf("no certificate files found in signcerts directory")
	}
	
	certFile := filepath.Join(signcertsPath, certFiles[0].Name())
	
	// Get the first key file in keystore
	keyFiles, err := ioutil.ReadDir(keystorePath)
	if err != nil {
		return fmt.Errorf("failed to read keystore directory: %v", err)
	}
	
	if len(keyFiles) == 0 {
		return fmt.Errorf("no key files found in keystore directory")
	}
	
	keyFile := filepath.Join(keystorePath, keyFiles[0].Name())
	
	// Read certificate and key
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %v", err)
	}
	
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key: %v", err)
	}
	
	// Determine MSP ID from directory structure or config file
	mspID := determineMspID(mspPath)
	
	// Create identity
	identity := gateway.NewX509Identity(mspID, string(cert), string(key))
	
	// Put in wallet
	err = wallet.Put("admin", identity)
	if err != nil {
		return fmt.Errorf("failed to put identity into wallet: %v", err)
	}
	
	return nil
}

func determineMspID(mspPath string) string {
	// Try to read msp-id from config.yaml if it exists
	configPath := filepath.Join(mspPath, "config.yaml")
	if fileExists(configPath) {
		configBytes, err := ioutil.ReadFile(configPath)
		if err == nil {
			configLines := strings.Split(string(configBytes), "\n")
			for _, line := range configLines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "mspid:") || strings.HasPrefix(line, "MSPID:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}
	
	// Default to Org1MSP if we couldn't determine from config
	return "Org1MSP"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func main() {
	err := setupWallet()
	if err != nil {
		fmt.Printf("Error setting up wallet: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Wallet setup completed successfully")
}

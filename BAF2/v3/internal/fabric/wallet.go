package fabric

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//"strings"

	"github.com/chaichis-network/v3/pkg/logger"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
)

const (
	// WalletPath is the default path for the identity wallet
	WalletPath = "wallet"
)

var log = logger.Default()

// Wallet represents an identity wallet for Fabric
type Wallet struct {
	path   string
	wallet *gateway.Wallet
}

// NewWallet creates a new wallet instance
func NewWallet(path string) (*Wallet, error) {
	if path == "" {
		path = WalletPath
	}
	
	// Create wallet directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, errors.Wrap(err, "failed to create wallet directory")
		}
	}
	
	// Create new file system wallet
	wallet, err := gateway.NewFileSystemWallet(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wallet")
	}
	
	return &Wallet{
		path:   path,
		wallet: wallet,
	}, nil
}

// DefaultWallet returns a wallet at the default location
func DefaultWallet() (*Wallet, error) {
	return NewWallet(WalletPath)
}

// Exists checks if an identity exists in the wallet
func (w *Wallet) Exists(label string) bool {
	return w.wallet.Exists(label)
}

// Put adds an identity to the wallet
func (w *Wallet) Put(label string, identity *gateway.X509Identity) error {
	return w.wallet.Put(label, identity)
}

// Get retrieves an identity from the wallet
func (w *Wallet) Get(label string) (*gateway.X509Identity, error) {
	id, err := w.wallet.Get(label)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get identity from wallet")
	}
	
	x509Identity, ok := id.(*gateway.X509Identity)
	if !ok {
		return nil, errors.New("identity is not an X509 identity")
	}
	
	return x509Identity, nil
}

// Remove removes an identity from the wallet
func (w *Wallet) Remove(label string) error {
	return w.wallet.Remove(label)
}

// List returns all identities in the wallet
func (w *Wallet) List() ([]string, error) {
	return w.wallet.List()
}

// ImportIdentity imports an identity from certificate and key files
func (w *Wallet) ImportIdentity(label, mspID, certPath, keyPath string) error {
	// Read certificate file
	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return errors.Wrap(err, "failed to read certificate file")
	}
	
	// Read key file
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return errors.Wrap(err, "failed to read key file")
	}
	
	// Create identity
	identity := gateway.NewX509Identity(mspID, string(cert), string(key))
	
	// Add to wallet
	return w.wallet.Put(label, identity)
}

// SearchAndImport searches common locations for certificates and imports them
func (w *Wallet) SearchAndImport(username string, mspID string) error {
	log.Infof("Searching for certificates for %s", username)
	
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

	// First try to find MSP structure
	for _, mspPath := range mspPaths {
		if !fileExists(mspPath) {
			continue
		}

		log.Infof("Checking MSP directory: %s", mspPath)

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

		// Find certificate files
		certFiles, err := filepath.Glob(filepath.Join(signcertsPath, "*.pem"))
		if err != nil || len(certFiles) == 0 {
			// Try .crt extension
			certFiles, err = filepath.Glob(filepath.Join(signcertsPath, "*.crt"))
			if err != nil || len(certFiles) == 0 {
				continue
			}
		}

		// Find key files
		keyFiles, err := filepath.Glob(filepath.Join(keystorePath, "*_sk"))
		if err != nil || len(keyFiles) == 0 {
			// Try .key extension
			keyFiles, err = filepath.Glob(filepath.Join(keystorePath, "*.key"))
			if err != nil || len(keyFiles) == 0 {
				continue
			}
		}

		if len(certFiles) > 0 && len(keyFiles) > 0 {
			log.Infof("Found certificate: %s", certFiles[0])
			log.Infof("Found key: %s", keyFiles[0])
			
			err := w.ImportIdentity(username, mspID, certFiles[0], keyFiles[0])
			if err != nil {
				log.Warnf("Failed to import identity: %v", err)
				continue
			}
			
			log.Infof("Successfully imported identity '%s' from MSP directory", username)
			return nil
		}
	}

	// Try individual certificate and key files
	for i, certPath := range certificatePaths {
		if !fileExists(certPath) {
			continue
		}

		// Find a matching key
		keyPath := ""
		if i < len(keyPaths) {
			keyPath = keyPaths[i]
			if !fileExists(keyPath) {
				continue
			}
		} else {
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

		log.Infof("Found certificate: %s", certPath)
		log.Infof("Found key: %s", keyPath)
		
		err := w.ImportIdentity(username, mspID, certPath, keyPath)
		if err != nil {
			log.Warnf("Failed to import identity: %v", err)
			continue
		}
		
		log.Infof("Successfully imported identity '%s'", username)
		return nil
	}

	return errors.New("could not find valid certificate and key files")
}

// PromptAndImport prompts the user for certificate and key paths and imports them
func (w *Wallet) PromptAndImport(username string) error {
	log.Info("Please provide the certificate and key paths for the identity:")
	
	var certPath, keyPath, mspID string
	
	log.Info("Certificate path (PEM format):")
	fmt.Scanln(&certPath)
	
	log.Info("Key path (PEM format):")
	fmt.Scanln(&keyPath)
	
	log.Info("MSP ID (e.g., Org1MSP):")
	fmt.Scanln(&mspID)
	
	if certPath == "" || keyPath == "" {
		return errors.New("certificate and key paths are required")
	}
	
	if mspID == "" {
		mspID = "Org1MSP"
		log.Infof("Using default MSP ID: %s", mspID)
	}
	
	return w.ImportIdentity(username, mspID, certPath, keyPath)
}

// EnsureIdentity ensures that an identity exists in the wallet, either by finding it
// or by importing it from common locations
func (w *Wallet) EnsureIdentity(username string) error {
	log.Infof("Ensuring identity for %s exists in wallet", username)
	
	// Check if identity already exists
	if w.Exists(username) {
		log.Infof("Identity %s already exists in wallet", username)
		return nil
	}
	
	log.Infof("Identity %s not found in wallet", username)
	
	// Try to find and import identity
	err := w.SearchAndImport(username, "Org1MSP")
	if err != nil {
		log.Warnf("Automatic import failed: %v", err)
		log.Info("Attempting manual import...")
		return w.PromptAndImport(username)
	}
	
	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

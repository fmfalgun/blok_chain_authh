package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	// KeyDir is the default directory for storing keys
	KeyDir = "keys"
	
	// DefaultKeySize is the default RSA key size in bits
	DefaultKeySize = 2048
)

// GenerateKeyPair generates a new RSA key pair
func GenerateKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate RSA key pair")
	}
	
	return privateKey, &privateKey.PublicKey, nil
}

// SavePrivateKey saves a private key to a file in PKCS#1 format
func SavePrivateKey(privateKey *rsa.PrivateKey, id string) (string, error) {
	// Ensure key directory exists
	if err := os.MkdirAll(KeyDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create key directory")
	}
	
	// Create path for private key
	keyPath := filepath.Join(KeyDir, fmt.Sprintf("%s-private.pem", id))
	
	// Marshal private key to PKCS1
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	
	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	
	// Create file with restricted permissions
	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", errors.Wrap(err, "failed to create private key file")
	}
	defer file.Close()
	
	// Write PEM block to file
	if err := pem.Encode(file, pemBlock); err != nil {
		return "", errors.Wrap(err, "failed to write private key to file")
	}
	
	return keyPath, nil
}

// SavePublicKey saves a public key to a file
func SavePublicKey(publicKey *rsa.PublicKey, id string) (string, error) {
	// Ensure key directory exists
	if err := os.MkdirAll(KeyDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create key directory")
	}
	
	// Create path for public key
	keyPath := filepath.Join(KeyDir, fmt.Sprintf("%s-public.pem", id))
	
	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal public key")
	}
	
	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	
	// Create file
	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", errors.Wrap(err, "failed to create public key file")
	}
	defer file.Close()
	
	// Write PEM block to file
	if err := pem.Encode(file, pemBlock); err != nil {
		return "", errors.Wrap(err, "failed to write public key to file")
	}
	
	return keyPath, nil
}

// LoadPrivateKey loads a private key from a file
func LoadPrivateKey(id string) (*rsa.PrivateKey, error) {
	// Get private key path
	keyPath := filepath.Join(KeyDir, fmt.Sprintf("%s-private.pem", id))
	
	// Read private key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key file")
	}
	
	// Decode PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	
	// Parse private key based on PEM block type
	var privateKey *rsa.PrivateKey
	
	if block.Type == "RSA PRIVATE KEY" {
		// PKCS#1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse PKCS1 private key")
		}
	} else if block.Type == "PRIVATE KEY" {
		// PKCS#8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse PKCS8 private key")
		}
		
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
	} else {
		return nil, errors.New("unsupported private key format")
	}
	
	return privateKey, nil
}

// LoadPublicKey loads a public key from a file
func LoadPublicKey(id string) (*rsa.PublicKey, error) {
	// Get public key path
	keyPath := filepath.Join(KeyDir, fmt.Sprintf("%s-public.pem", id))
	
	// Read public key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read public key file")
	}
	
	return ParsePublicKeyPEM(keyData)
}

// ParsePublicKeyPEM parses a public key from PEM data
func ParsePublicKeyPEM(pemData []byte) (*rsa.PublicKey, error) {
	// Decode PEM block
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	
	// Parse public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key")
	}
	
	// Ensure it's an RSA public key
	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	
	return rsaKey, nil
}

// LoadOrGenerateKeys loads existing keys for an entity or generates new ones if they don't exist
func LoadOrGenerateKeys(id string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Check if private key exists
	privateKeyPath := filepath.Join(KeyDir, fmt.Sprintf("%s-private.pem", id))
	if _, err := os.Stat(privateKeyPath); err == nil {
		// Load existing private key
		privateKey, err := LoadPrivateKey(id)
		if err != nil {
			return nil, nil, err
		}
		
		return privateKey, &privateKey.PublicKey, nil
	}
	
	// Generate new key pair
	privateKey, publicKey, err := GenerateKeyPair(DefaultKeySize)
	if err != nil {
		return nil, nil, err
	}
	
	// Save keys
	_, err = SavePrivateKey(privateKey, id)
	if err != nil {
		return nil, nil, err
	}
	
	_, err = SavePublicKey(publicKey, id)
	if err != nil {
		return nil, nil, err
	}
	
	return privateKey, publicKey, nil
}

// GetPublicKeyPEM returns the PEM-encoded public key for an entity
func GetPublicKeyPEM(id string) (string, error) {
	keyPath := filepath.Join(KeyDir, fmt.Sprintf("%s-public.pem", id))
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read public key file")
	}
	
	return string(keyData), nil
}

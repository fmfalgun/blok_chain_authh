// test-keys.go
// A utility for testing RSA key generation, signing, and encryption in Go

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	// Define key paths
	privateKeyFile = "test-private.pem"
	publicKeyFile  = "test-public.pem"
)

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

// saveKeys saves the generated keys to files
func saveKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) error {
	// Save private key in PKCS1 format (Go's default for RSA private keys)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY", // This specifies PKCS1 format
			Bytes: privateKeyBytes,
		},
	)
	err := ioutil.WriteFile(privateKeyFile, privateKeyPEM, 0600)
	if err != nil {
		return err
	}
	
	// Save public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	err = ioutil.WriteFile(publicKeyFile, publicKeyPEM, 0644)
	if err != nil {
		return err
	}
	
	return nil
}

// testSigning tests the signing and verification process
func testSigning(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) error {
	// Test data
	testData := []byte("test data for signing")
	
	// Create hash of the data
	hash := sha256.Sum256(testData)
	
	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign data: %v", err)
	}
	
	fmt.Printf("Signature created successfully (base64): %s\n", base64.StdEncoding.EncodeToString(signature))
	fmt.Printf("Signature length: %d bytes\n", len(signature))
	
	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}
	
	fmt.Println("Signature verification succeeded")
	return nil
}

// testEncryption tests the encryption and decryption process
func testEncryption(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) error {
	// Test data
	testData := []byte("test data for encryption")
	
	// Encrypt with public key
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, testData)
	if err != nil {
		return fmt.Errorf("encryption failed: %v", err)
	}
	
	fmt.Printf("Encrypted data (base64): %s\n", base64.StdEncoding.EncodeToString(ciphertext))
	fmt.Printf("Encrypted data length: %d bytes\n", len(ciphertext))
	
	// Decrypt with private key
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		return fmt.Errorf("decryption failed: %v", err)
	}
	
	fmt.Printf("Decrypted data: %s\n", string(plaintext))
	return nil
}

// loadKeys loads the keys from files
func loadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Load private key
	privateKeyPEM, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key file: %v", err)
	}
	
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	
	var privateKey *rsa.PrivateKey
	
	// Check the key type by examining the PEM block type
	switch block.Type {
	case "RSA PRIVATE KEY":
		// PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		// PKCS8 format
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse PKCS8 private key: %v", err)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, fmt.Errorf("not an RSA private key")
		}
	default:
		return nil, nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}
	
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	
	// Load public key
	publicKeyPEM, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key file: %v", err)
	}
	
	block, _ = pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}
	
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public key: %v", err)
	}
	
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("not an RSA public key")
	}
	
	return privateKey, publicKey, nil
}

// simpleSignAndVerify demonstrates the simple signing and verification process
func simpleSignAndVerify() error {
	// Generate key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Test data - this could be a challenge nonce
	testData := []byte("challenge-nonce-123456")
	
	// Create hash of the data
	hash := sha256.Sum256(testData)
	
	// Sign the hash with private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign data: %v", err)
	}
	
	fmt.Printf("Original data: %s\n", string(testData))
	fmt.Printf("SHA256 hash (base64): %s\n", base64.StdEncoding.EncodeToString(hash[:]))
	fmt.Printf("Signature (base64): %s\n", base64.StdEncoding.EncodeToString(signature))
	
	// Verify the signature with public key
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}
	
	fmt.Println("Signature verification succeeded")
	return nil
}

// main function
func main() {
	fmt.Println("=== RSA Key Testing Utility ===")
	
	if len(os.Args) > 1 && os.Args[1] == "simple" {
		fmt.Println("\nRunning simple sign and verify test...")
		err := simpleSignAndVerify()
		if err != nil {
			fmt.Printf("Simple test failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Simple test completed successfully")
		return
	}
	
	// Generate new key pair
	fmt.Println("\nGenerating new RSA key pair...")
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		fmt.Printf("Failed to generate key pair: %v\n", err)
		os.Exit(1)
	}
	
	// Save keys to files
	fmt.Println("Saving keys to files...")
	err = saveKeys(privateKey, publicKey)
	if err != nil {
		fmt.Printf("Failed to save keys: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Private key saved to: %s\n", privateKeyFile)
	fmt.Printf("Public key saved to: %s\n", publicKeyFile)
	
	// Test signing
	fmt.Println("\nTesting signing and verification...")
	err = testSigning(privateKey, publicKey)
	if err != nil {
		fmt.Printf("Signing test failed: %v\n", err)
		os.Exit(1)
	}
	
	// Test encryption
	fmt.Println("\nTesting encryption and decryption...")
	err = testEncryption(privateKey, publicKey)
	if err != nil {
		fmt.Printf("Encryption test failed: %v\n", err)
		os.Exit(1)
	}
	
	// Load keys from files and test again
	fmt.Println("\nLoading keys from files and testing again...")
	loadedPrivateKey, loadedPublicKey, err := loadKeys()
	if err != nil {
		fmt.Printf("Failed to load keys: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Testing signing with loaded keys...")
	err = testSigning(loadedPrivateKey, loadedPublicKey)
	if err != nil {
		fmt.Printf("Signing test with loaded keys failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("\nAll tests completed successfully!")
}

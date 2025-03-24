// standalone-auth-framework.go
// Simplified Go implementation for testing RSA compatibility with Hyperledger Fabric chaincodes

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
	"path/filepath"
)

// ASPublicKey is a constant for AS public key
const ASPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtOL3THYTwCk35h9/BYpX
/5pQGH4jK5nyO55oI8PqBMx6GHfnP0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4Y
TNNmV/IQKqX1eUtRJSP/gZR5/wQ06H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKig
P48qIMYQYGXOSbL7IfvOWXV+TZ6o9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMw
lnUdZ/Yx4hOABHX6+0yQJxECU2OWve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnY
FpV35FTy98dDYL7N6+y6whldMMQ680dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1T
XQIDAQAB
-----END PUBLIC KEY-----`

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

// savePrivateKey saves a private key to a file in PKCS#1 format
func savePrivateKey(privateKey *rsa.PrivateKey, filename string) error {
	// Create directories if they don't exist
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	
	// Marshal private key to PKCS1 (same as traditional RSA private key format)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	
	// Encode to PEM
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY", // This specifies PKCS1 format
			Bytes: privateKeyBytes,
		},
	)
	
	return ioutil.WriteFile(filename, privateKeyPEM, 0600)
}

// savePublicKey saves a public key to a file
func savePublicKey(publicKey *rsa.PublicKey, filename string) error {
	// Create directories if they don't exist
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	
	// Marshal public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	
	// Encode to PEM
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	
	return ioutil.WriteFile(filename, publicKeyPEM, 0644)
}

// loadPrivateKey loads a private key from a file
func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	// Read the private key file
	privateKeyPEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	// Parse the PEM block
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	
	// Parse the private key
	var privateKey *rsa.PrivateKey
	
	// Check if it's PKCS1 or PKCS8 format
	if block.Type == "RSA PRIVATE KEY" {
		// PKCS1 format
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %v", err)
		}
	} else {
		// PKCS8 format
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %v", err)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}
	
	return privateKey, nil
}

// loadPublicKey loads a public key from a PEM string
func loadPublicKey(pemString string) (*rsa.PublicKey, error) {
	// Parse the PEM block
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}
	
	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}
	
	// Cast to RSA public key
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	
	return rsaPub, nil
}

// encryptWithPublicKey encrypts data with a public key
func encryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) (string, error) {
	// Encrypt the data
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return "", fmt.Errorf("encryption error: %v", err)
	}
	
	// Return base64 encoded encrypted data
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// signData signs data with a private key
func signData(privateKey *rsa.PrivateKey, data []byte) (string, error) {
	// Create a hash of the data
	hash := sha256.Sum256(data)
	
	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("signing error: %v", err)
	}
	
	// Return base64 encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifySignature verifies a signature using a public key
func verifySignature(publicKey *rsa.PublicKey, data []byte, signatureBase64 string) error {
	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %v", err)
	}
	
	// Create a hash of the data
	hash := sha256.Sum256(data)
	
	// Verify the signature
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
}

// generateClientKeys generates and saves client keys
func generateClientKeys(clientId string) error {
	// Generate a new key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Save the private key
	privateKeyFile := fmt.Sprintf("keys/%s-private.pem", clientId)
	err = savePrivateKey(privateKey, privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}
	
	// Save the public key
	publicKeyFile := fmt.Sprintf("keys/%s-public.pem", clientId)
	err = savePublicKey(publicKey, publicKeyFile)
	if err != nil {
		return fmt.Errorf("failed to save public key: %v", err)
	}
	
	fmt.Printf("Generated keys for client %s:\n", clientId)
	fmt.Printf("Private key: %s\n", privateKeyFile)
	fmt.Printf("Public key: %s\n", publicKeyFile)
	
	return nil
}

// simulateAuthentication simulates the Kerberos-like authentication flow
func simulateAuthentication(clientId string, nonce string) error {
	fmt.Println("\n=== Starting Authentication Simulation ===")
	
	// Step 1: Load client's private key
	fmt.Printf("Loading private key for client %s...\n", clientId)
	privateKeyFile := fmt.Sprintf("keys/%s-private.pem", clientId)
	privateKey, err := loadPrivateKey(privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load private key: %v", err)
	}
	
	// Step 2: Load AS public key
	fmt.Println("Loading AS public key...")
	asPublicKey, err := loadPublicKey(ASPublicKey)
	if err != nil {
		return fmt.Errorf("failed to load AS public key: %v", err)
	}
	
	// Step 3: Sign the nonce with the client's private key
	fmt.Printf("Signing nonce '%s' with client's private key...\n", nonce)
	nonceBytes := []byte(nonce)
	signedNonce, err := signData(privateKey, nonceBytes)
	if err != nil {
		return fmt.Errorf("failed to sign nonce: %v", err)
	}
	fmt.Printf("Signed nonce (base64): %s\n", signedNonce)
	
	// Step 4: Verify the signature using client's public key (this would be done by the AS)
	fmt.Println("Verifying signature (simulating AS)...")
	publicKeyFile := fmt.Sprintf("keys/%s-public.pem", clientId)
	publicKeyPEM, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read public key: %v", err)
	}
	
	clientPublicKey, err := loadPublicKey(string(publicKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to load client's public key: %v", err)
	}
	
	err = verifySignature(clientPublicKey, nonceBytes, signedNonce)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}
	fmt.Println("Signature verified successfully!")
	
	// Step 5: Encrypt the nonce with AS public key (alternative approach)
	fmt.Println("\nTrying alternative encryption approach...")
	encryptedNonce, err := encryptWithPublicKey(asPublicKey, nonceBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt nonce: %v", err)
	}
	fmt.Printf("Encrypted nonce (base64): %s\n", encryptedNonce)
	
	// Step 6: Debug - Try encrypt/decrypt with own keys for verification
	fmt.Println("\nVerifying encrypt/decrypt with client's own keys...")
	testMessage := []byte("test-message-123")
	encryptedTest, err := encryptWithPublicKey(clientPublicKey, testMessage)
	if err != nil {
		return fmt.Errorf("failed to encrypt test message: %v", err)
	}
	fmt.Printf("Encrypted test message: %s\n", encryptedTest)
	
	// Step 7: Sign test message and verify
	signedTest, err := signData(privateKey, testMessage)
	if err != nil {
		return fmt.Errorf("failed to sign test message: %v", err)
	}
	
	err = verifySignature(clientPublicKey, testMessage, signedTest)
	if err != nil {
		return fmt.Errorf("test signature verification failed: %v", err)
	}
	fmt.Println("Test signature verification successful!")
	
	// Step 8: Export hex values for Go/Node.js comparison
	fmt.Println("\nExporting values for comparison:")
	fmt.Printf("Original nonce (hex): %x\n", nonceBytes)
	signatureBytes, _ := base64.StdEncoding.DecodeString(signedNonce)
	fmt.Printf("Signature (hex): %x\n", signatureBytes)
	
	fmt.Println("\n=== Authentication Simulation Completed ===")
	
	return nil
}

// debugRSAEncryption is a debugging utility for RSA operations
func debugRSAEncryption(nonce string) error {
	fmt.Println("======= RSA ENCRYPTION DEBUG =======")
	
	// Check if input is already base64 encoded
	var nonceBuffer []byte
	var err error
	
	// Try to decode as base64
	nonceBuffer, err = base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		fmt.Println("Nonce is not base64 encoded, treating as plain text")
		nonceBuffer = []byte(nonce)
	} else {
		fmt.Println("Nonce appears to be base64 encoded. Decoded:", string(nonceBuffer))
	}
	
	fmt.Println("Nonce as buffer:", nonceBuffer)
	fmt.Println("Nonce buffer length:", len(nonceBuffer))
	
	// Try different encryption approaches
	fmt.Println("\nTrying different encryption approaches:")
	
	// Approach 1: Standard Go encryption with PKCS1v15
	block, _ := pem.Decode([]byte(ASPublicKey))
	if block == nil {
		return fmt.Errorf("failed to parse PEM block containing the public key")
	}
	
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}
	
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}
	
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, nonceBuffer)
	if err != nil {
		return fmt.Errorf("encryption error: %v", err)
	}
	
	encrypted64 := base64.StdEncoding.EncodeToString(encrypted)
	fmt.Println("Approach 1 - Result (base64):", encrypted64)
	fmt.Println("Approach 1 - Length:", len(encrypted))
	
	// Generate a key pair for verification
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Create hash and sign
	hash := sha256.Sum256(nonceBuffer)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return fmt.Errorf("signing error: %v", err)
	}
	
	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %v", err)
	}
	
	fmt.Println("Self-verification test passed successfully!")
	fmt.Println("======= DEBUG COMPLETE =======")
	
	return nil
}

// testKeysURI Tests if client key URI/paths would match between Nodejs and Go clients
func testKeysURI(clientId string) error {
	fmt.Println("======= TESTING KEY PATHS =======")
	
	// File paths that would be used
	nodejsPrivatePath := fmt.Sprintf("%s-private.pem", clientId)
	nodejsPublicPath := "auto-generated (not saved separately)"
	
	goPrivatePath := fmt.Sprintf("keys/%s-private.pem", clientId)
	goPublicPath := fmt.Sprintf("keys/%s-public.pem", clientId)
	
	// Check if nodejs key exists
	if _, err := os.Stat(nodejsPrivatePath); !os.IsNotExist(err) {
		fmt.Printf("Node.js private key found: %s\n", nodejsPrivatePath)
		
		// Compare with Go private key if it exists
		if _, err := os.Stat(goPrivatePath); !os.IsNotExist(err) {
			fmt.Println("Comparing Node.js and Go private keys...")
			nodejsKey, err := loadPrivateKey(nodejsPrivatePath)
			if err != nil {
				return fmt.Errorf("failed to load Node.js private key: %v", err)
			}
			
			goKey, err := loadPrivateKey(goPrivatePath)
			if err != nil {
				return fmt.Errorf("failed to load Go private key: %v", err)
			}
			
			// Compare private key modulus
			if nodejsKey.N.Cmp(goKey.N) == 0 {
				fmt.Println("Key modulus matches! Keys are compatible.")
			} else {
				fmt.Println("WARNING: Key modulus does not match. Keys are different.")
			}
		}
	} else {
		fmt.Printf("Node.js private key not found: %s\n", nodejsPrivatePath)
	}
	
	fmt.Println("\nKey paths comparison:")
	fmt.Println("Node.js client:")
	fmt.Printf("  Private key: %s\n", nodejsPrivatePath)
	fmt.Printf("  Public key: %s\n", nodejsPublicPath)
	
	fmt.Println("Go client:")
	fmt.Printf("  Private key: %s\n", goPrivatePath)
	fmt.Printf("  Public key: %s\n", goPublicPath)
	
	fmt.Println("\nTo make Node.js and Go clients use the same keys:")
	fmt.Println("1. First generate keys with Go client")
	fmt.Println("2. Copy the private key to the Node.js working directory:")
	fmt.Printf("   cp %s %s\n", goPrivatePath, nodejsPrivatePath)
	
	fmt.Println("======= KEY PATHS TEST COMPLETE =======")
	return nil
}

// printUsage displays usage information
func printUsage() {
	fmt.Println("===== Standalone Authentication Framework for Testing =====")
	fmt.Println("Available commands:")
	fmt.Println("  generate-keys <clientId>  - Generate new RSA keys for a client")
	fmt.Println("  simulate-auth <clientId> <nonce> - Simulate authentication with a nonce")
	fmt.Println("  debug-rsa <nonce>         - Debug RSA encryption/signing with a nonce")
	fmt.Println("  test-keys-uri <clientId>  - Test key file paths between Node.js and Go")
	fmt.Println("\nExamples:")
	fmt.Println("  go run standalone-auth-framework.go generate-keys client1")
	fmt.Println("  go run standalone-auth-framework.go simulate-auth client1 test-nonce-123")
	fmt.Println("  go run standalone-auth-framework.go debug-rsa test-nonce-123")
	fmt.Println("  go run standalone-auth-framework.go test-keys-uri client1")
}

// main is the entry point for the program
func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	
	command := os.Args[1]
	var err error
	
	switch command {
	case "generate-keys":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run standalone-auth-framework.go generate-keys <clientId>")
			return
		}
		clientId := os.Args[2]
		err = generateClientKeys(clientId)
		
	case "simulate-auth":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run standalone-auth-framework.go simulate-auth <clientId> <nonce>")
			return
		}
		clientId := os.Args[2]
		nonce := os.Args[3]
		err = simulateAuthentication(clientId, nonce)
		
	case "debug-rsa":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run standalone-auth-framework.go debug-rsa <nonce>")
			return
		}
		nonce := os.Args[2]
		err = debugRSAEncryption(nonce)
		
	case "test-keys-uri":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run standalone-auth-framework.go test-keys-uri <clientId>")
			return
		}
		clientId := os.Args[2]
		err = testKeysURI(clientId)
		
	default:
		printUsage()
		return
	}
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

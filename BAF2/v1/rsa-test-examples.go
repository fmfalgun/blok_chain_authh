// rsa-test-examples.go
// Examples of RSA operations in Go that are compatible with Node.js

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

// PKCS1 vs PKCS8 private key formats
func demoKeyFormats() {
	fmt.Println("\n=== Key Format Comparison ===")
	
	// Generate a new key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return
	}
	
	// PKCS1 Format (Traditional RSA format)
	pkcs1Bytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pkcs1PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY", // Note this type identifier
		Bytes: pkcs1Bytes,
	})
	
	// PKCS8 Format (Modern format that can hold different key types)
	pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	pkcs8PEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY", // Note the different type identifier
		Bytes: pkcs8Bytes,
	})
	
	fmt.Println("PKCS#1 Format (Go's default for RSA, compatible with the Go chaincode):")
	fmt.Println(string(pkcs1PEM))
	
	fmt.Println("PKCS#8 Format (Node.js default, can cause compatibility issues):")
	fmt.Println(string(pkcs8PEM))
	
	// Save the keys for reference
	ioutil.WriteFile("demo-pkcs1.pem", pkcs1PEM, 0600)
	ioutil.WriteFile("demo-pkcs8.pem", pkcs8PEM, 0600)
	fmt.Println("Key samples saved to demo-pkcs1.pem and demo-pkcs8.pem")
}

// Example of signing a nonce in Go (compatible with Go chaincode)
func demoSigningForGo() {
	fmt.Println("\n=== Signing Example for Go Compatibility ===")
	
	// Load or generate a private key in PKCS1 format
	var privateKey *rsa.PrivateKey
	
	if _, err := os.Stat("demo-pkcs1.pem"); os.IsNotExist(err) {
		fmt.Println("Generating new key...")
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			fmt.Printf("Error generating key: %v\n", err)
			return
		}
	} else {
		fmt.Println("Loading existing key...")
		keyData, err := ioutil.ReadFile("demo-pkcs1.pem")
		if err != nil {
			fmt.Printf("Error reading key file: %v\n", err)
			return
		}
		
		block, _ := pem.Decode(keyData)
		if block == nil {
			fmt.Println("Failed to decode PEM block")
			return
		}
		
		if block.Type == "RSA PRIVATE KEY" {
			privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		} else {
			parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				fmt.Printf("Error parsing private key: %v\n", err)
				return
			}
			var ok bool
			privateKey, ok = parsedKey.(*rsa.PrivateKey)
			if !ok {
				fmt.Println("Not an RSA private key")
				return
			}
		}
	}
	
	// Sample nonce to sign (this would come from the chaincode)
	nonce := []byte("test-nonce-123456")
	fmt.Printf("Original nonce: %s\n", string(nonce))
	fmt.Printf("Nonce in hex: %s\n", hex.EncodeToString(nonce))
	
	// Hash the nonce (SHA-256)
	hash := sha256.Sum256(nonce)
	fmt.Printf("SHA-256 hash of nonce (hex): %s\n", hex.EncodeToString(hash[:]))
	
	// Sign the hash with the private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		fmt.Printf("Error signing hash: %v\n", err)
		return
	}
	
	// Output the signature in both hex and base64 formats
	fmt.Printf("Signature (hex): %s\n", hex.EncodeToString(signature))
	fmt.Printf("Signature (base64): %s\n", base64.StdEncoding.EncodeToString(signature))
	
	// Verify the signature (this is what the Go chaincode would do)
	publicKey := &privateKey.PublicKey
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		fmt.Printf("Signature verification failed: %v\n", err)
	} else {
		fmt.Println("Signature verified successfully!")
	}
	
	// Export the public key for the chaincode
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Printf("Error marshaling public key: %v\n", err)
		return
	}
	
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	
	fmt.Println("\nPublic key for chaincode:")
	fmt.Println(string(publicKeyPEM))
}

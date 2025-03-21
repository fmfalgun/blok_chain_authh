package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"

	"github.com/pkg/errors"
)

// SignData signs data with the given private key
func SignData(privateKey *rsa.PrivateKey, data []byte) (string, error) {
	// Create SHA-256 hash of data
	hash := sha256.Sum256(data)
	
	// Sign the hash with the private key
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", errors.Wrap(err, "failed to sign data")
	}
	
	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies a signature using the given public key
func VerifySignature(publicKey *rsa.PublicKey, data []byte, signatureBase64 string) error {
	// Decode base64 signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return errors.Wrap(err, "failed to decode base64 signature")
	}
	
	// Create SHA-256 hash of data
	hash := sha256.Sum256(data)
	
	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return errors.Wrap(err, "signature verification failed")
	}
	
	return nil
}

// EncryptWithPublicKey encrypts data with a public key
func EncryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) (string, error) {
	// Encrypt data with the public key
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt data with public key")
	}
	
	// Return base64-encoded encrypted data
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptWithPrivateKey decrypts data with a private key
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, encryptedBase64 string) ([]byte, error) {
	// Decode base64 encrypted data
	encrypted, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64 encrypted data")
	}
	
	// Decrypt data with the private key
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encrypted)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt data with private key")
	}
	
	return decrypted, nil
}

// SignNonce signs a nonce using an entity's private key
func SignNonce(id string, nonce string) (string, error) {
	// Load private key
	privateKey, err := LoadPrivateKey(id)
	if err != nil {
		return "", errors.Wrap(err, "failed to load private key")
	}
	
	// Sign nonce
	return SignData(privateKey, []byte(nonce))
}

// VerifyNonceSignature verifies a signed nonce
func VerifyNonceSignature(id string, nonce string, signature string) error {
	// Load public key
	publicKey, err := LoadPublicKey(id)
	if err != nil {
		return errors.Wrap(err, "failed to load public key")
	}
	
	// Verify signature
	return VerifySignature(publicKey, []byte(nonce), signature)
}

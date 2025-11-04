package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateSecureRandomBytes generates cryptographically secure random bytes
func GenerateSecureRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %v", err)
	}
	return bytes, nil
}

// GenerateSecureNonce generates a secure random nonce (256 bits)
func GenerateSecureNonce() (string, error) {
	nonce, err := GenerateSecureRandomBytes(32) // 256 bits
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(nonce), nil
}

// GenerateSessionKey generates a secure random session key (256 bits)
func GenerateSessionKey() (string, error) {
	key, err := GenerateSecureRandomBytes(32) // 256 bits
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// GenerateTicketID generates a unique ticket identifier
func GenerateTicketID() (string, error) {
	id, err := GenerateSecureRandomBytes(16)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}

// HashData creates a SHA256 hash of the input data
func HashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetCurrentTimestamp returns the current Unix timestamp
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// IsExpired checks if a timestamp has expired relative to the current time
func IsExpired(timestamp int64, validitySeconds int64) bool {
	currentTime := GetCurrentTimestamp()
	return currentTime > (timestamp + validitySeconds)
}

// ValidateTimestamp checks if a timestamp is within a valid range
func ValidateTimestamp(timestamp int64, maxAgeSeconds int64) error {
	currentTime := GetCurrentTimestamp()
	age := currentTime - timestamp

	if age < 0 {
		return fmt.Errorf("timestamp is in the future")
	}

	if age > maxAgeSeconds {
		return fmt.Errorf("timestamp is too old (age: %d seconds, max: %d seconds)", age, maxAgeSeconds)
	}

	return nil
}

// GenerateDeviceID generates a unique device identifier
func GenerateDeviceID() (string, error) {
	id, err := GenerateSecureRandomBytes(16)
	if err != nil {
		return "", err
	}
	return "device_" + hex.EncodeToString(id), nil
}

// EncodeToHex encodes bytes to hexadecimal string
func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeFromHex decodes hexadecimal string to bytes
func DecodeFromHex(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// EncodeToBase64 encodes bytes to base64 string
func EncodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeFromBase64 decodes base64 string to bytes
func DecodeFromBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

package common

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MaxIDLength         = 64
	MinIDLength         = 3
	MaxPEMLength        = 4096
	MinPEMLength        = 100
	MaxMetadataLength   = 1024
	MaxSignatureLength  = 4096
	MinSignatureLength  = 10
	MaxNonceLength      = 256
	MinNonceLength      = 16
	MaxIPAddressLength  = 45  // IPv6 max length
	MaxUserAgentLength  = 256
	MaxDescriptionLength = 512
)

var (
	// ValidIDPattern allows alphanumeric characters, underscores, and hyphens
	ValidIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// ValidActionPattern for ISV actions
	ValidActionPattern = regexp.MustCompile(`^(read|write|execute|delete)$`)

	// ValidStatusPattern for status fields
	ValidStatusPattern = regexp.MustCompile(`^(active|inactive|suspended|revoked|valid|expired|used|terminated)$`)

	// ValidIPv4Pattern for IPv4 addresses
	ValidIPv4Pattern = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)

	// ValidIPv6Pattern for IPv6 addresses (simplified)
	ValidIPv6Pattern = regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateDeviceID validates a device ID
func ValidateDeviceID(deviceID string) error {
	if len(deviceID) < MinIDLength {
		return &ValidationError{
			Field:   "deviceID",
			Message: fmt.Sprintf("length must be at least %d characters", MinIDLength),
		}
	}

	if len(deviceID) > MaxIDLength {
		return &ValidationError{
			Field:   "deviceID",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxIDLength),
		}
	}

	if !ValidIDPattern.MatchString(deviceID) {
		return &ValidationError{
			Field:   "deviceID",
			Message: "must contain only alphanumeric characters, underscores, and hyphens",
		}
	}

	return nil
}

// ValidateServiceID validates a service ID
func ValidateServiceID(serviceID string) error {
	if len(serviceID) < MinIDLength {
		return &ValidationError{
			Field:   "serviceID",
			Message: fmt.Sprintf("length must be at least %d characters", MinIDLength),
		}
	}

	if len(serviceID) > MaxIDLength {
		return &ValidationError{
			Field:   "serviceID",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxIDLength),
		}
	}

	if !ValidIDPattern.MatchString(serviceID) {
		return &ValidationError{
			Field:   "serviceID",
			Message: "must contain only alphanumeric characters, underscores, and hyphens",
		}
	}

	return nil
}

// ValidatePublicKey validates a PEM-encoded public key
func ValidatePublicKey(publicKey string) error {
	if len(publicKey) < MinPEMLength {
		return &ValidationError{
			Field:   "publicKey",
			Message: fmt.Sprintf("length must be at least %d characters", MinPEMLength),
		}
	}

	if len(publicKey) > MaxPEMLength {
		return &ValidationError{
			Field:   "publicKey",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxPEMLength),
		}
	}

	if !strings.Contains(publicKey, "BEGIN PUBLIC KEY") || !strings.Contains(publicKey, "END PUBLIC KEY") {
		return &ValidationError{
			Field:   "publicKey",
			Message: "must be a valid PEM-encoded public key",
		}
	}

	return nil
}

// ValidateSignature validates a signature
func ValidateSignature(signature string) error {
	if len(signature) < MinSignatureLength {
		return &ValidationError{
			Field:   "signature",
			Message: fmt.Sprintf("length must be at least %d characters", MinSignatureLength),
		}
	}

	if len(signature) > MaxSignatureLength {
		return &ValidationError{
			Field:   "signature",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxSignatureLength),
		}
	}

	return nil
}

// ValidateNonce validates a nonce
func ValidateNonce(nonce string) error {
	if len(nonce) < MinNonceLength {
		return &ValidationError{
			Field:   "nonce",
			Message: fmt.Sprintf("length must be at least %d characters", MinNonceLength),
		}
	}

	if len(nonce) > MaxNonceLength {
		return &ValidationError{
			Field:   "nonce",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxNonceLength),
		}
	}

	return nil
}

// ValidateMetadata validates metadata fields
func ValidateMetadata(metadata string) error {
	if len(metadata) > MaxMetadataLength {
		return &ValidationError{
			Field:   "metadata",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxMetadataLength),
		}
	}

	return nil
}

// ValidateAction validates an action string
func ValidateAction(action string) error {
	if !ValidActionPattern.MatchString(action) {
		return &ValidationError{
			Field:   "action",
			Message: "must be one of: read, write, execute, delete",
		}
	}

	return nil
}

// ValidateStatus validates a status string
func ValidateStatus(status string) error {
	if !ValidStatusPattern.MatchString(status) {
		return &ValidationError{
			Field:   "status",
			Message: "invalid status value",
		}
	}

	return nil
}

// ValidateIPAddress validates an IP address
func ValidateIPAddress(ipAddress string) error {
	if len(ipAddress) == 0 {
		return nil // IP address is optional in some contexts
	}

	if len(ipAddress) > MaxIPAddressLength {
		return &ValidationError{
			Field:   "ipAddress",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxIPAddressLength),
		}
	}

	if !ValidIPv4Pattern.MatchString(ipAddress) && !ValidIPv6Pattern.MatchString(ipAddress) {
		return &ValidationError{
			Field:   "ipAddress",
			Message: "must be a valid IPv4 or IPv6 address",
		}
	}

	return nil
}

// ValidateUserAgent validates a user agent string
func ValidateUserAgent(userAgent string) error {
	if len(userAgent) > MaxUserAgentLength {
		return &ValidationError{
			Field:   "userAgent",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxUserAgentLength),
		}
	}

	return nil
}

// ValidateDescription validates a description string
func ValidateDescription(description string) error {
	if len(description) > MaxDescriptionLength {
		return &ValidationError{
			Field:   "description",
			Message: fmt.Sprintf("length must not exceed %d characters", MaxDescriptionLength),
		}
	}

	return nil
}

// ValidateTimestampRange validates that a timestamp is within acceptable bounds
func ValidateTimestampRange(timestamp int64, maxAgeSeconds int64) error {
	currentTime := GetCurrentTimestamp()
	age := currentTime - timestamp

	if age < -300 { // Allow 5 minutes clock skew into the future
		return &ValidationError{
			Field:   "timestamp",
			Message: "timestamp is too far in the future",
		}
	}

	if age > maxAgeSeconds {
		return &ValidationError{
			Field:   "timestamp",
			Message: fmt.Sprintf("timestamp is too old (max age: %d seconds)", maxAgeSeconds),
		}
	}

	return nil
}

// SanitizeInput removes potentially dangerous characters from input
func SanitizeInput(input string) string {
	// Remove control characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)

	return strings.TrimSpace(sanitized)
}

// ValidateJSONField validates that a field is valid JSON
func ValidateJSONField(fieldName string, jsonStr string) error {
	if len(jsonStr) == 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: "cannot be empty",
		}
	}

	if len(jsonStr) > 10240 { // 10KB max for JSON fields
		return &ValidationError{
			Field:   fieldName,
			Message: "JSON data too large (max 10KB)",
		}
	}

	return nil
}

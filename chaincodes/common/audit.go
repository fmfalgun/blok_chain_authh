package common

import (
	"encoding/json"
	"fmt"
	"time"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Authentication events
	EventDeviceRegistered    AuditEventType = "DEVICE_REGISTERED"
	EventDeviceAuthenticated AuditEventType = "DEVICE_AUTHENTICATED"
	EventDeviceRevoked       AuditEventType = "DEVICE_REVOKED"
	EventAuthenticationFailed AuditEventType = "AUTHENTICATION_FAILED"

	// Service ticket events
	EventServiceTicketIssued  AuditEventType = "SERVICE_TICKET_ISSUED"
	EventServiceTicketRevoked AuditEventType = "SERVICE_TICKET_REVOKED"
	EventServiceTicketExpired AuditEventType = "SERVICE_TICKET_EXPIRED"

	// Access events
	EventAccessGranted AuditEventType = "ACCESS_GRANTED"
	EventAccessDenied  AuditEventType = "ACCESS_DENIED"
	EventSessionCreated AuditEventType = "SESSION_CREATED"
	EventSessionTerminated AuditEventType = "SESSION_TERMINATED"

	// Security events
	EventRateLimitExceeded AuditEventType = "RATE_LIMIT_EXCEEDED"
	EventValidationFailed  AuditEventType = "VALIDATION_FAILED"
	EventSignatureInvalid  AuditEventType = "SIGNATURE_INVALID"
	EventTimestampInvalid  AuditEventType = "TIMESTAMP_INVALID"

	// Administrative events
	EventServiceRegistered AuditEventType = "SERVICE_REGISTERED"
	EventServiceDeactivated AuditEventType = "SERVICE_DEACTIVATED"
	EventConfigurationChanged AuditEventType = "CONFIGURATION_CHANGED"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	SeverityInfo     AuditSeverity = "INFO"
	SeverityWarning  AuditSeverity = "WARNING"
	SeverityError    AuditSeverity = "ERROR"
	SeverityCritical AuditSeverity = "CRITICAL"
)

// AuditEvent represents a comprehensive audit log entry
type AuditEvent struct {
	// Event identification
	EventID   string         `json:"eventID"`
	EventType AuditEventType `json:"eventType"`
	Timestamp int64          `json:"timestamp"`

	// Severity and status
	Severity AuditSeverity `json:"severity"`
	Status   string        `json:"status"` // success, failure, denied

	// Actor information
	DeviceID string `json:"deviceID,omitempty"`
	ActorID  string `json:"actorID,omitempty"` // Could be admin, system, etc.

	// Resource information
	ResourceType string `json:"resourceType,omitempty"` // device, service, ticket, session
	ResourceID   string `json:"resourceID,omitempty"`

	// Action details
	Action      string `json:"action,omitempty"`
	Description string `json:"description"`

	// Context information
	IPAddress     string `json:"ipAddress,omitempty"`
	UserAgent     string `json:"userAgent,omitempty"`
	ChaincodeName string `json:"chaincodeName,omitempty"`

	// Additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// Error information (if applicable)
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	chaincodeName string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(chaincodeName string) *AuditLogger {
	return &AuditLogger{
		chaincodeName: chaincodeName,
	}
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(eventType AuditEventType, severity AuditSeverity, status string, description string) *AuditEvent {
	eventID, _ := GenerateSecureRandomBytes(16)

	event := &AuditEvent{
		EventID:       EncodeToHex(eventID),
		EventType:     eventType,
		Timestamp:     GetCurrentTimestamp(),
		Severity:      severity,
		Status:        status,
		Description:   description,
		ChaincodeName: al.chaincodeName,
		Metadata:      make(map[string]string),
	}

	return event
}

// LogDeviceRegistration logs a device registration event
func (al *AuditLogger) LogDeviceRegistration(deviceID string, status string) *AuditEvent {
	event := al.LogEvent(EventDeviceRegistered, SeverityInfo, status, "Device registered")
	event.DeviceID = deviceID
	event.ResourceType = "device"
	event.ResourceID = deviceID
	event.Action = "register"
	return event
}

// LogAuthentication logs an authentication event
func (al *AuditLogger) LogAuthentication(deviceID string, success bool, reason string) *AuditEvent {
	var eventType AuditEventType
	var severity AuditSeverity
	var status string

	if success {
		eventType = EventDeviceAuthenticated
		severity = SeverityInfo
		status = "success"
	} else {
		eventType = EventAuthenticationFailed
		severity = SeverityWarning
		status = "failure"
	}

	event := al.LogEvent(eventType, severity, status, reason)
	event.DeviceID = deviceID
	event.ResourceType = "device"
	event.ResourceID = deviceID
	event.Action = "authenticate"
	return event
}

// LogServiceTicketIssuance logs a service ticket issuance event
func (al *AuditLogger) LogServiceTicketIssuance(deviceID string, serviceID string, ticketID string, status string) *AuditEvent {
	event := al.LogEvent(EventServiceTicketIssued, SeverityInfo, status, "Service ticket issued")
	event.DeviceID = deviceID
	event.ResourceType = "ticket"
	event.ResourceID = ticketID
	event.Action = "issue"
	event.Metadata["serviceID"] = serviceID
	return event
}

// LogAccessAttempt logs an access attempt event
func (al *AuditLogger) LogAccessAttempt(deviceID string, serviceID string, action string, granted bool, reason string, ipAddress string) *AuditEvent {
	var eventType AuditEventType
	var severity AuditSeverity
	var status string

	if granted {
		eventType = EventAccessGranted
		severity = SeverityInfo
		status = "success"
	} else {
		eventType = EventAccessDenied
		severity = SeverityWarning
		status = "denied"
	}

	event := al.LogEvent(eventType, severity, status, reason)
	event.DeviceID = deviceID
	event.ResourceType = "service"
	event.ResourceID = serviceID
	event.Action = action
	event.IPAddress = ipAddress
	return event
}

// LogSecurityEvent logs a security-related event
func (al *AuditLogger) LogSecurityEvent(eventType AuditEventType, deviceID string, reason string, severity AuditSeverity) *AuditEvent {
	event := al.LogEvent(eventType, severity, "failure", reason)
	event.DeviceID = deviceID
	return event
}

// LogRateLimitExceeded logs a rate limit exceeded event
func (al *AuditLogger) LogRateLimitExceeded(deviceID string, ipAddress string) *AuditEvent {
	event := al.LogEvent(EventRateLimitExceeded, SeverityWarning, "failure", "Rate limit exceeded")
	event.DeviceID = deviceID
	event.IPAddress = ipAddress
	return event
}

// LogValidationError logs a validation error
func (al *AuditLogger) LogValidationError(deviceID string, field string, errorMessage string) *AuditEvent {
	description := fmt.Sprintf("Validation failed for field: %s", field)
	event := al.LogEvent(EventValidationFailed, SeverityWarning, "failure", description)
	event.DeviceID = deviceID
	event.ErrorMessage = errorMessage
	event.Metadata["field"] = field
	return event
}

// SetIPAddress sets the IP address for an event
func (event *AuditEvent) SetIPAddress(ipAddress string) *AuditEvent {
	event.IPAddress = ipAddress
	return event
}

// SetUserAgent sets the user agent for an event
func (event *AuditEvent) SetUserAgent(userAgent string) *AuditEvent {
	event.UserAgent = userAgent
	return event
}

// AddMetadata adds metadata to an event
func (event *AuditEvent) AddMetadata(key string, value string) *AuditEvent {
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Metadata[key] = value
	return event
}

// SetError sets error information for an event
func (event *AuditEvent) SetError(code string, message string) *AuditEvent {
	event.ErrorCode = code
	event.ErrorMessage = message
	return event
}

// ToJSON converts the audit event to JSON
func (event *AuditEvent) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("failed to marshal audit event: %v", err)
	}
	return string(jsonBytes), nil
}

// FormatForDisplay formats the event for human-readable display
func (event *AuditEvent) FormatForDisplay() string {
	timestamp := time.Unix(event.Timestamp, 0).Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] [%s] %s - Device: %s, Resource: %s/%s",
		timestamp,
		event.Severity,
		event.EventType,
		event.Description,
		event.DeviceID,
		event.ResourceType,
		event.ResourceID,
	)
}

// IsSecurityCritical determines if an event is security-critical
func (event *AuditEvent) IsSecurityCritical() bool {
	return event.Severity == SeverityCritical ||
		event.EventType == EventRateLimitExceeded ||
		event.EventType == EventSignatureInvalid ||
		event.EventType == EventAuthenticationFailed
}

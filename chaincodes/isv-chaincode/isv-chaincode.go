package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ISVChaincode provides IoT service validator functions
type ISVChaincode struct {
	contractapi.Contract
}

// AccessLog represents an access log entry
type AccessLog struct {
	LogID       string `json:"logID"`
	DeviceID    string `json:"deviceID"`
	ServiceID   string `json:"serviceID"`
	TicketID    string `json:"ticketID"`
	Timestamp   int64  `json:"timestamp"`
	Action      string `json:"action"` // read, write, execute
	Status      string `json:"status"` // success, failure, denied
	IPAddress   string `json:"ipAddress"`
	UserAgent   string `json:"userAgent"`
	Description string `json:"description"`
}

// DeviceSession represents an active device session
type DeviceSession struct {
	SessionID  string `json:"sessionID"`
	DeviceID   string `json:"deviceID"`
	ServiceID  string `json:"serviceID"`
	StartTime  int64  `json:"startTime"`
	LastActive int64  `json:"lastActive"`
	Status     string `json:"status"` // active, expired, terminated
}

// AccessRequest represents a request to access a service
type AccessRequest struct {
	DeviceID   string `json:"deviceID"`
	ServiceID  string `json:"serviceID"`
	TicketID   string `json:"ticketID"`
	Action     string `json:"action"`
	Timestamp  int64  `json:"timestamp"`
	IPAddress  string `json:"ipAddress"`
	UserAgent  string `json:"userAgent"`
	Signature  string `json:"signature"`
}

// AccessResponse represents the response to an access request
type AccessResponse struct {
	Granted    bool   `json:"granted"`
	SessionID  string `json:"sessionID"`
	Message    string `json:"message"`
	ExpiresAt  int64  `json:"expiresAt"`
}

// InitLedger initializes the ledger
func (s *ISVChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("Initializing ISV Chaincode ledger")
	return nil
}

// ValidateAccess validates a device's access request to a service
func (s *ISVChaincode) ValidateAccess(ctx contractapi.TransactionContextInterface, accessRequestJSON string) (string, error) {
	var accessReq AccessRequest
	err := json.Unmarshal([]byte(accessRequestJSON), &accessReq)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal access request: %v", err)
	}

	// Validate timestamp (within 5 minutes)
	currentTime := getCurrentTimestamp()
	if accessReq.Timestamp < currentTime-300 || accessReq.Timestamp > currentTime+300 {
		logAccess(ctx, accessReq.DeviceID, accessReq.ServiceID, accessReq.TicketID, accessReq.Action, "failure", accessReq.IPAddress, accessReq.UserAgent, "Invalid timestamp")
		return createAccessResponse(false, "", "Invalid or expired timestamp", 0)
	}

	// Verify signature (in production)
	if len(accessReq.Signature) < 10 {
		logAccess(ctx, accessReq.DeviceID, accessReq.ServiceID, accessReq.TicketID, accessReq.Action, "failure", accessReq.IPAddress, accessReq.UserAgent, "Invalid signature")
		return createAccessResponse(false, "", "Invalid signature", 0)
	}

	// Validate action
	validActions := []string{"read", "write", "execute"}
	isValidAction := false
	for _, validAction := range validActions {
		if accessReq.Action == validAction {
			isValidAction = true
			break
		}
	}
	if !isValidAction {
		logAccess(ctx, accessReq.DeviceID, accessReq.ServiceID, accessReq.TicketID, accessReq.Action, "denied", accessReq.IPAddress, accessReq.UserAgent, "Invalid action")
		return createAccessResponse(false, "", "Invalid action", 0)
	}

	// In production, validate ticket by cross-chaincode invocation to TGS
	// For now, basic validation
	if len(accessReq.TicketID) < 5 {
		logAccess(ctx, accessReq.DeviceID, accessReq.ServiceID, accessReq.TicketID, accessReq.Action, "denied", accessReq.IPAddress, accessReq.UserAgent, "Invalid ticket")
		return createAccessResponse(false, "", "Invalid ticket", 0)
	}

	// Check for existing active session
	sessionID, err := findActiveSession(ctx, accessReq.DeviceID, accessReq.ServiceID)
	if err == nil && sessionID != "" {
		// Update existing session
		err = updateSession(ctx, sessionID)
		if err != nil {
			return "", fmt.Errorf("failed to update session: %v", err)
		}

		logAccess(ctx, accessReq.DeviceID, accessReq.ServiceID, accessReq.TicketID, accessReq.Action, "success", accessReq.IPAddress, accessReq.UserAgent, "Using existing session")
		return createAccessResponse(true, sessionID, "Access granted (existing session)", currentTime+1800)
	}

	// Create new session
	newSessionID, err := generateSecureSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %v", err)
	}

	session := DeviceSession{
		SessionID:  newSessionID,
		DeviceID:   accessReq.DeviceID,
		ServiceID:  accessReq.ServiceID,
		StartTime:  currentTime,
		LastActive: currentTime,
		Status:     "active",
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %v", err)
	}

	err = ctx.GetStub().PutState("SESSION_"+newSessionID, sessionJSON)
	if err != nil {
		return "", fmt.Errorf("failed to store session: %v", err)
	}

	// Log successful access
	logAccess(ctx, accessReq.DeviceID, accessReq.ServiceID, accessReq.TicketID, accessReq.Action, "success", accessReq.IPAddress, accessReq.UserAgent, "New session created")

	// Emit event
	err = ctx.GetStub().SetEvent("AccessGranted", []byte(newSessionID))
	if err != nil {
		return "", fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Access granted to device %s for service %s (session: %s)", accessReq.DeviceID, accessReq.ServiceID, newSessionID)
	return createAccessResponse(true, newSessionID, "Access granted", currentTime+1800)
}

// TerminateSession terminates an active session
func (s *ISVChaincode) TerminateSession(ctx contractapi.TransactionContextInterface, sessionID string) error {
	sessionJSON, err := ctx.GetStub().GetState("SESSION_" + sessionID)
	if err != nil {
		return fmt.Errorf("failed to read session: %v", err)
	}
	if sessionJSON == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	var session DeviceSession
	err = json.Unmarshal(sessionJSON, &session)
	if err != nil {
		return fmt.Errorf("failed to unmarshal session: %v", err)
	}

	session.Status = "terminated"

	sessionJSON, err = json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}

	err = ctx.GetStub().PutState("SESSION_"+sessionID, sessionJSON)
	if err != nil {
		return fmt.Errorf("failed to update session: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("SessionTerminated", []byte(sessionID))
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Session %s terminated", sessionID)
	return nil
}

// GetAccessLogs retrieves access logs for a device
func (s *ISVChaincode) GetAccessLogs(ctx contractapi.TransactionContextInterface, deviceID string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("LOG_", "LOG_~")
	if err != nil {
		return "", fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var logs []AccessLog
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("failed to iterate: %v", err)
		}

		var log AccessLog
		err = json.Unmarshal(queryResponse.Value, &log)
		if err != nil {
			continue
		}

		if log.DeviceID == deviceID {
			logs = append(logs, log)
		}
	}

	logsJSON, err := json.Marshal(logs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal logs: %v", err)
	}

	return string(logsJSON), nil
}

// GetSession retrieves session information
func (s *ISVChaincode) GetSession(ctx contractapi.TransactionContextInterface, sessionID string) (string, error) {
	sessionJSON, err := ctx.GetStub().GetState("SESSION_" + sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to read session: %v", err)
	}
	if sessionJSON == nil {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	return string(sessionJSON), nil
}

// GetActiveSessions returns all active sessions
func (s *ISVChaincode) GetActiveSessions(ctx contractapi.TransactionContextInterface) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("SESSION_", "SESSION_~")
	if err != nil {
		return "", fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var sessions []DeviceSession
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("failed to iterate: %v", err)
		}

		var session DeviceSession
		err = json.Unmarshal(queryResponse.Value, &session)
		if err != nil {
			continue
		}

		if session.Status == "active" {
			sessions = append(sessions, session)
		}
	}

	sessionsJSON, err := json.Marshal(sessions)
	if err != nil {
		return "", fmt.Errorf("failed to marshal sessions: %v", err)
	}

	return string(sessionsJSON), nil
}

// Helper functions

func logAccess(ctx contractapi.TransactionContextInterface, deviceID, serviceID, ticketID, action, status, ipAddress, userAgent, description string) error {
	logID, err := generateSecureLogID()
	if err != nil {
		return err
	}

	accessLog := AccessLog{
		LogID:       logID,
		DeviceID:    deviceID,
		ServiceID:   serviceID,
		TicketID:    ticketID,
		Timestamp:   getCurrentTimestamp(),
		Action:      action,
		Status:      status,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Description: description,
	}

	logJSON, err := json.Marshal(accessLog)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState("LOG_"+logID, logJSON)
}

func findActiveSession(ctx contractapi.TransactionContextInterface, deviceID, serviceID string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("SESSION_", "SESSION_~")
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	currentTime := getCurrentTimestamp()
	sessionTimeout := int64(1800) // 30 minutes

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			continue
		}

		var session DeviceSession
		err = json.Unmarshal(queryResponse.Value, &session)
		if err != nil {
			continue
		}

		if session.DeviceID == deviceID && session.ServiceID == serviceID && session.Status == "active" {
			// Check if session hasn't timed out
			if currentTime-session.LastActive < sessionTimeout {
				return session.SessionID, nil
			}
		}
	}

	return "", fmt.Errorf("no active session found")
}

func updateSession(ctx contractapi.TransactionContextInterface, sessionID string) error {
	sessionJSON, err := ctx.GetStub().GetState("SESSION_" + sessionID)
	if err != nil {
		return err
	}

	var session DeviceSession
	err = json.Unmarshal(sessionJSON, &session)
	if err != nil {
		return err
	}

	session.LastActive = getCurrentTimestamp()

	sessionJSON, err = json.Marshal(session)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState("SESSION_"+sessionID, sessionJSON)
}

func createAccessResponse(granted bool, sessionID, message string, expiresAt int64) (string, error) {
	response := AccessResponse{
		Granted:   granted,
		SessionID: sessionID,
		Message:   message,
		ExpiresAt: expiresAt,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", err
	}

	return string(responseJSON), nil
}

func getCurrentTimestamp() int64 {
	return 1672531200 // Placeholder
}

func generateSecureSessionID() (string, error) {
	return "session_" + generateRandomString(16), nil
}

func generateSecureLogID() (string, error) {
	return "log_" + generateRandomString(16), nil
}

func generateRandomString(length int) string {
	return "random" + fmt.Sprintf("%d", getCurrentTimestamp())
}

func main() {
	isvChaincode, err := contractapi.NewChaincode(&ISVChaincode{})
	if err != nil {
		log.Panicf("Error creating ISV chaincode: %v", err)
	}

	if err := isvChaincode.Start(); err != nil {
		log.Panicf("Error starting ISV chaincode: %v", err)
	}
}

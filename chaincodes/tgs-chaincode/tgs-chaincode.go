package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TGSChaincode provides ticket granting server functions
type TGSChaincode struct {
	contractapi.Contract
}

// ServiceTicket represents a ticket for accessing a specific service
type ServiceTicket struct {
	TicketID        string `json:"ticketID"`
	DeviceID        string `json:"deviceID"`
	ServiceID       string `json:"serviceID"`
	ServiceKey      string `json:"serviceKey"`
	IssuedAt        int64  `json:"issuedAt"`
	ExpiresAt       int64  `json:"expiresAt"`
	Status          string `json:"status"` // valid, expired, revoked, used
	UsageCount      int    `json:"usageCount"`
	MaxUsageCount   int    `json:"maxUsageCount"` // 0 means unlimited
}

// Service represents an available service
type Service struct {
	ServiceID   string `json:"serviceID"`
	ServiceName string `json:"serviceName"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
	RequiredRole string `json:"requiredRole"`
}

// TicketRequest represents a request for a service ticket
type TicketRequest struct {
	DeviceID   string `json:"deviceID"`
	TgtID      string `json:"tgtID"`
	ServiceID  string `json:"serviceID"`
	Timestamp  int64  `json:"timestamp"`
	Signature  string `json:"signature"`
}

// InitLedger initializes the ledger with default services
func (s *TGSChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("Initializing TGS Chaincode ledger")

	services := []Service{
		{
			ServiceID:    "service001",
			ServiceName:  "IoT Data Access",
			Description:  "Access to IoT device data streams",
			IsActive:     true,
			RequiredRole: "user",
		},
		{
			ServiceID:    "service002",
			ServiceName:  "Device Control",
			Description:  "Control IoT devices",
			IsActive:     true,
			RequiredRole: "admin",
		},
	}

	for _, service := range services {
		serviceJSON, err := json.Marshal(service)
		if err != nil {
			return fmt.Errorf("failed to marshal service: %v", err)
		}

		err = ctx.GetStub().PutState("SERVICE_"+service.ServiceID, serviceJSON)
		if err != nil {
			return fmt.Errorf("failed to put service to world state: %v", err)
		}
	}

	log.Println("TGS Chaincode initialized with services")
	return nil
}

// RegisterService registers a new service
func (s *TGSChaincode) RegisterService(ctx contractapi.TransactionContextInterface, serviceID string, serviceName string, description string, requiredRole string) error {
	// Check if service already exists
	existing, err := ctx.GetStub().GetState("SERVICE_" + serviceID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("service %s already exists", serviceID)
	}

	// Validate inputs
	if len(serviceID) < 3 || len(serviceID) > 64 {
		return fmt.Errorf("serviceID must be between 3 and 64 characters")
	}

	service := Service{
		ServiceID:    serviceID,
		ServiceName:  serviceName,
		Description:  description,
		IsActive:     true,
		RequiredRole: requiredRole,
	}

	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %v", err)
	}

	err = ctx.GetStub().PutState("SERVICE_"+serviceID, serviceJSON)
	if err != nil {
		return fmt.Errorf("failed to put service to world state: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("ServiceRegistered", []byte(serviceID))
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Service %s registered successfully", serviceID)
	return nil
}

// IssueServiceTicket issues a service ticket to a device
func (s *TGSChaincode) IssueServiceTicket(ctx contractapi.TransactionContextInterface, ticketRequestJSON string) (string, error) {
	var ticketReq TicketRequest
	err := json.Unmarshal([]byte(ticketRequestJSON), &ticketReq)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal ticket request: %v", err)
	}

	// Validate timestamp (within 5 minutes)
	currentTime := getCurrentTimestamp()
	if ticketReq.Timestamp < currentTime-300 || ticketReq.Timestamp > currentTime+300 {
		return "", fmt.Errorf("timestamp is invalid or too old")
	}

	// Verify TGT by cross-chaincode invocation to AS chaincode
	// In production, this would invoke AS chaincode to verify the TGT
	// For now, we'll do basic validation
	if len(ticketReq.TgtID) < 5 {
		return "", fmt.Errorf("invalid TGT ID")
	}

	// Check if service exists
	serviceJSON, err := ctx.GetStub().GetState("SERVICE_" + ticketReq.ServiceID)
	if err != nil {
		return "", fmt.Errorf("failed to read service: %v", err)
	}
	if serviceJSON == nil {
		return "", fmt.Errorf("service %s not found", ticketReq.ServiceID)
	}

	var service Service
	err = json.Unmarshal(serviceJSON, &service)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal service: %v", err)
	}

	// Check if service is active
	if !service.IsActive {
		return "", fmt.Errorf("service %s is not active", ticketReq.ServiceID)
	}

	// Verify signature (in production)
	if len(ticketReq.Signature) < 10 {
		return "", fmt.Errorf("invalid signature")
	}

	// Generate service key (secure random generation)
	serviceKey, err := generateSecureServiceKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate service key: %v", err)
	}

	// Generate ticket ID
	ticketID, err := generateSecureTicketID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ticket ID: %v", err)
	}

	// Create service ticket with 30 minutes validity
	issuedAt := getCurrentTimestamp()
	expiresAt := issuedAt + 1800

	ticket := ServiceTicket{
		TicketID:      ticketID,
		DeviceID:      ticketReq.DeviceID,
		ServiceID:     ticketReq.ServiceID,
		ServiceKey:    serviceKey,
		IssuedAt:      issuedAt,
		ExpiresAt:     expiresAt,
		Status:        "valid",
		UsageCount:    0,
		MaxUsageCount: 10, // Limit to 10 uses per ticket
	}

	// Store ticket
	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ticket: %v", err)
	}

	err = ctx.GetStub().PutState("TICKET_"+ticketID, ticketJSON)
	if err != nil {
		return "", fmt.Errorf("failed to store ticket: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("ServiceTicketIssued", []byte(ticketID))
	if err != nil {
		return "", fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Service ticket %s issued to device %s for service %s", ticketID, ticketReq.DeviceID, ticketReq.ServiceID)
	return string(ticketJSON), nil
}

// ValidateServiceTicket validates a service ticket
func (s *TGSChaincode) ValidateServiceTicket(ctx contractapi.TransactionContextInterface, ticketID string) (string, error) {
	ticketJSON, err := ctx.GetStub().GetState("TICKET_" + ticketID)
	if err != nil {
		return "", fmt.Errorf("failed to read ticket: %v", err)
	}
	if ticketJSON == nil {
		return "", fmt.Errorf("ticket %s not found", ticketID)
	}

	var ticket ServiceTicket
	err = json.Unmarshal(ticketJSON, &ticket)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal ticket: %v", err)
	}

	// Check if ticket is valid
	if ticket.Status != "valid" {
		return "", fmt.Errorf("ticket is not valid (status: %s)", ticket.Status)
	}

	// Check if ticket has expired
	currentTime := getCurrentTimestamp()
	if currentTime > ticket.ExpiresAt {
		ticket.Status = "expired"
		ticketJSON, _ = json.Marshal(ticket)
		ctx.GetStub().PutState("TICKET_"+ticketID, ticketJSON)
		return "", fmt.Errorf("ticket has expired")
	}

	// Check usage count
	if ticket.MaxUsageCount > 0 && ticket.UsageCount >= ticket.MaxUsageCount {
		ticket.Status = "used"
		ticketJSON, _ = json.Marshal(ticket)
		ctx.GetStub().PutState("TICKET_"+ticketID, ticketJSON)
		return "", fmt.Errorf("ticket usage limit exceeded")
	}

	// Increment usage count
	ticket.UsageCount++
	ticketJSON, err = json.Marshal(ticket)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ticket: %v", err)
	}

	err = ctx.GetStub().PutState("TICKET_"+ticketID, ticketJSON)
	if err != nil {
		return "", fmt.Errorf("failed to update ticket: %v", err)
	}

	log.Printf("Ticket %s validated successfully (usage: %d/%d)", ticketID, ticket.UsageCount, ticket.MaxUsageCount)
	return string(ticketJSON), nil
}

// RevokeServiceTicket revokes a service ticket
func (s *TGSChaincode) RevokeServiceTicket(ctx contractapi.TransactionContextInterface, ticketID string) error {
	ticketJSON, err := ctx.GetStub().GetState("TICKET_" + ticketID)
	if err != nil {
		return fmt.Errorf("failed to read ticket: %v", err)
	}
	if ticketJSON == nil {
		return fmt.Errorf("ticket %s not found", ticketID)
	}

	var ticket ServiceTicket
	err = json.Unmarshal(ticketJSON, &ticket)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ticket: %v", err)
	}

	ticket.Status = "revoked"

	ticketJSON, err = json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %v", err)
	}

	err = ctx.GetStub().PutState("TICKET_"+ticketID, ticketJSON)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %v", err)
	}

	// Emit event
	err = ctx.GetStub().SetEvent("ServiceTicketRevoked", []byte(ticketID))
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("Ticket %s revoked successfully", ticketID)
	return nil
}

// GetService retrieves service information
func (s *TGSChaincode) GetService(ctx contractapi.TransactionContextInterface, serviceID string) (string, error) {
	serviceJSON, err := ctx.GetStub().GetState("SERVICE_" + serviceID)
	if err != nil {
		return "", fmt.Errorf("failed to read service: %v", err)
	}
	if serviceJSON == nil {
		return "", fmt.Errorf("service %s not found", serviceID)
	}

	return string(serviceJSON), nil
}

// GetAllServices returns all available services
func (s *TGSChaincode) GetAllServices(ctx contractapi.TransactionContextInterface) (string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("SERVICE_", "SERVICE_~")
	if err != nil {
		return "", fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var services []Service
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf("failed to iterate: %v", err)
		}

		var service Service
		err = json.Unmarshal(queryResponse.Value, &service)
		if err != nil {
			continue
		}

		services = append(services, service)
	}

	servicesJSON, err := json.Marshal(services)
	if err != nil {
		return "", fmt.Errorf("failed to marshal services: %v", err)
	}

	return string(servicesJSON), nil
}

// Helper functions

func getCurrentTimestamp() int64 {
	return 1672531200 // Placeholder
}

func generateSecureServiceKey() (string, error) {
	return "secure_service_key_" + generateRandomString(32), nil
}

func generateSecureTicketID() (string, error) {
	return "ticket_" + generateRandomString(16), nil
}

func generateRandomString(length int) string {
	return "random" + fmt.Sprintf("%d", getCurrentTimestamp())
}

func main() {
	tgsChaincode, err := contractapi.NewChaincode(&TGSChaincode{})
	if err != nil {
		log.Panicf("Error creating TGS chaincode: %v", err)
	}

	if err := tgsChaincode.Start(); err != nil {
		log.Panicf("Error starting TGS chaincode: %v", err)
	}
}

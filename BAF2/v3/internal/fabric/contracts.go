package fabric

import (
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/pkg/errors"
)

const (
	// Contract IDs
	ASContractID  = "as-chaincode_1.1"
	TGSContractID = "tgs-chaincode_2.0"
	ISVContractID = "isv-chaincode_2.0"
)

// ContractManager manages interactions with the Fabric contracts
type ContractManager struct {
	client *Client
}

// NewContractManager creates a new contract manager
func NewContractManager(client *Client) *ContractManager {
	return &ContractManager{
		client: client,
	}
}

// GetASContract returns the Authentication Server contract
func (cm *ContractManager) GetASContract() (*gateway.Contract, error) {
	return cm.client.GetContract(ASContractID)
}

// GetTGSContract returns the Ticket Granting Server contract
func (cm *ContractManager) GetTGSContract() (*gateway.Contract, error) {
	return cm.client.GetContract(TGSContractID)
}

// GetISVContract returns the IoT Service Validator contract
func (cm *ContractManager) GetISVContract() (*gateway.Contract, error) {
	return cm.client.GetContract(ISVContractID)
}

// AuthServerContract provides operations for the Authentication Server chaincode
type AuthServerContract struct {
	contract *gateway.Contract
}

// NewAuthServerContract creates a new Auth Server contract handler
func NewAuthServerContract(client *Client) (*AuthServerContract, error) {
	contract, err := client.GetContract(ASContractID)
	if err != nil {
		return nil, err
	}
	
	return &AuthServerContract{
		contract: contract,
	}, nil
}

// RegisterClient registers a client with the Authentication Server
func (as *AuthServerContract) RegisterClient(clientID, clientPublicKeyPEM string) error {
	_, err := as.contract.SubmitTransaction("RegisterClient", clientID, clientPublicKeyPEM)
	if err != nil {
		return errors.Wrap(err, "failed to register client with AS")
	}
	
	return nil
}

// GetNonceChallenge gets a nonce challenge for client authentication
func (as *AuthServerContract) GetNonceChallenge(clientID string) (string, error) {
	responseBytes, err := as.contract.SubmitTransaction("InitiateAuthentication", clientID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get nonce challenge from AS")
	}
	
	var response struct {
		Nonce          string `json:"nonce"`
		ExpirationTime int64  `json:"expirationTime"`
	}
	
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return "", errors.Wrap(err, "failed to parse nonce response")
	}
	
	return response.Nonce, nil
}

// VerifyClientIdentity verifies a client's identity using a signed nonce
func (as *AuthServerContract) VerifyClientIdentity(clientID, signedNonce string) error {
	_, err := as.contract.SubmitTransaction("VerifyClientIdentityWithSignature", clientID, signedNonce)
	if err != nil {
		return errors.Wrap(err, "failed to verify client identity with AS")
	}
	
	return nil
}

// GenerateTGT generates a Ticket Granting Ticket for a client
func (as *AuthServerContract) GenerateTGT(clientID string) (map[string]string, error) {
	responseBytes, err := as.contract.SubmitTransaction("GenerateTGT", clientID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate TGT from AS")
	}
	
	var response map[string]string
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to parse TGT response")
	}
	
	return response, nil
}

// TicketGrantingContract provides operations for the Ticket Granting Server chaincode
type TicketGrantingContract struct {
	contract *gateway.Contract
}

// NewTicketGrantingContract creates a new Ticket Granting contract handler
func NewTicketGrantingContract(client *Client) (*TicketGrantingContract, error) {
	contract, err := client.GetContract(TGSContractID)
	if err != nil {
		return nil, err
	}
	
	return &TicketGrantingContract{
		contract: contract,
	}, nil
}

// GenerateServiceTicket generates a service ticket for a client
func (tgs *TicketGrantingContract) GenerateServiceTicket(request map[string]string) (map[string]string, error) {
	// Convert request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal service ticket request")
	}
	
	responseBytes, err := tgs.contract.SubmitTransaction("GenerateServiceTicket", string(requestJSON))
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate service ticket from TGS")
	}
	
	var response map[string]string
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to parse service ticket response")
	}
	
	return response, nil
}

// ISVContract provides operations for the IoT Service Validator chaincode
type ISVContract struct {
	contract *gateway.Contract
}

// NewISVContract creates a new ISV contract handler
func NewISVContract(client *Client) (*ISVContract, error) {
	contract, err := client.GetContract(ISVContractID)
	if err != nil {
		return nil, err
	}
	
	return &ISVContract{
		contract: contract,
	}, nil
}

// RegisterIoTDevice registers an IoT device with the ISV
func (isv *ISVContract) RegisterIoTDevice(deviceID, devicePublicKeyPEM string, capabilities []string) error {
	// Convert capabilities to JSON
	capabilitiesJSON, err := json.Marshal(capabilities)
	if err != nil {
		return errors.Wrap(err, "failed to marshal capabilities")
	}
	
	_, err = isv.contract.SubmitTransaction("RegisterIoTDevice", deviceID, devicePublicKeyPEM, string(capabilitiesJSON))
	if err != nil {
		return errors.Wrap(err, "failed to register IoT device with ISV")
	}
	
	return nil
}

// ValidateServiceTicket validates a service ticket with the ISV
func (isv *ISVContract) ValidateServiceTicket(encryptedServiceTicket string) error {
	_, err := isv.contract.SubmitTransaction("ValidateServiceTicket", encryptedServiceTicket)
	if err != nil {
		return errors.Wrap(err, "failed to validate service ticket with ISV")
	}
	
	return nil
}

// ProcessServiceRequest processes a service request for an IoT device
func (isv *ISVContract) ProcessServiceRequest(request map[string]string) (map[string]string, error) {
	// Convert request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal service request")
	}
	
	responseBytes, err := isv.contract.SubmitTransaction("ProcessServiceRequest", string(requestJSON))
	if err != nil {
		return nil, errors.Wrap(err, "failed to process service request with ISV")
	}
	
	var response map[string]string
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, errors.Wrap(err, "failed to parse service response")
	}
	
	return response, nil
}

// CloseSession closes an active session with an IoT device
func (isv *ISVContract) CloseSession(sessionID string) error {
	_, err := isv.contract.SubmitTransaction("CloseSession", sessionID)
	if err != nil {
		return errors.Wrap(err, "failed to close session with ISV")
	}
	
	return nil
}

// GetAllIoTDevices retrieves all registered IoT devices
func (isv *ISVContract) GetAllIoTDevices() ([]map[string]interface{}, error) {
	responseBytes, err := isv.contract.EvaluateTransaction("GetAllIoTDevices")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get IoT devices from ISV")
	}
	
	var devices []map[string]interface{}
	if err := json.Unmarshal(responseBytes, &devices); err != nil {
		return nil, errors.Wrap(err, "failed to parse IoT devices response")
	}
	
	return devices, nil
}

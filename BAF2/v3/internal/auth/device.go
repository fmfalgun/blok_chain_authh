package auth

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/chaichis-network/v3/internal/crypto"
	"github.com/chaichis-network/v3/internal/fabric"
	"github.com/pkg/errors"
)

// DeviceManager manages IoT device operations
type DeviceManager struct {
	fabricClient *fabric.Client
	isvContract  *fabric.ISVContract
	identity     string
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(fabricClient *fabric.Client, identity string) (*DeviceManager, error) {
	// Ensure client is connected
	if err := fabricClient.Connect(identity); err != nil {
		return nil, errors.Wrap(err, "failed to connect to Fabric network")
	}
	
	// Get ISV contract
	isvContract, err := fabric.NewISVContract(fabricClient)
	if err != nil {
		return nil, err
	}
	
	return &DeviceManager{
		fabricClient: fabricClient,
		isvContract:  isvContract,
		identity:     identity,
	}, nil
}

// RegisterDevice registers a new IoT device with the ISV
func (dm *DeviceManager) RegisterDevice(deviceID string, capabilities []string) error {
	// Generate or load device keys
	_, _, err := crypto.LoadOrGenerateKeys(deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to load or generate device keys")
	}
	
	// Get device's public key PEM
	publicKeyPEM, err := crypto.GetPublicKeyPEM(deviceID)
	if err != nil {
		return errors.Wrap(err, "failed to get device's public key PEM")
	}
	
	// Register device with ISV
	if err := dm.isvContract.RegisterIoTDevice(deviceID, publicKeyPEM, capabilities); err != nil {
		return errors.Wrap(err, "failed to register device with ISV")
	}
	
	log.Infof("Device %s registered successfully with capabilities: %v", deviceID, capabilities)
	return nil
}

// GetDeviceData gets information about a device
func (dm *DeviceManager) GetDeviceData(deviceID string) (*IoTDevice, error) {
	// Get all devices
	devices, err := dm.isvContract.GetAllIoTDevices()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get IoT devices")
	}
	
	// Find the requested device
	for _, device := range devices {
		if device["deviceID"] == deviceID {
			// Convert to IoTDevice struct
			// Extract capabilities from interface{} slice
			capabilitiesIface, ok := device["capabilities"].([]interface{})
			capabilities := make([]string, 0)
			if ok {
				for _, cap := range capabilitiesIface {
					if capStr, ok := cap.(string); ok {
						capabilities = append(capabilities, capStr)
					}
				}
			}
			
			// Create IoTDevice
			iotDevice := &IoTDevice{
				DeviceID:     deviceID,
				Status:       device["status"].(string),
				Capabilities: capabilities,
			}
			
			// Optional fields
			if lastSeen, ok := device["lastSeen"].(string); ok {
				iotDevice.LastSeen = lastSeen
			}
			
			if registeredAt, ok := device["registeredAt"].(string); ok {
				iotDevice.RegisteredAt = registeredAt
			}
			
			return iotDevice, nil
		}
	}
	
	return nil, errors.Errorf("device %s not found", deviceID)
}

// AccessDevice requests access to an IoT device
func (dm *DeviceManager) AccessDevice(clientID, deviceID string) (*Session, error) {
	// Get service ticket
	serviceTicket, err := (&ClientManager{
		fabricClient: dm.fabricClient,
		identity:     dm.identity,
	}).GetServiceTicket(clientID, deviceID)
	
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service ticket")
	}
	
	// Create service request
	serviceRequest := ServiceRequest{
		EncryptedServiceTicket: serviceTicket["encryptedServiceTicket"],
		ClientID:               clientID,
		DeviceID:               deviceID,
		RequestType:            "read",
		EncryptedData:          base64.StdEncoding.EncodeToString([]byte("read-request")),
	}
	
	// Convert to map for contract
	requestMap := map[string]string{
		"encryptedServiceTicket": serviceRequest.EncryptedServiceTicket,
		"clientID":               serviceRequest.ClientID,
		"deviceID":               serviceRequest.DeviceID,
		"requestType":            serviceRequest.RequestType,
		"encryptedData":          serviceRequest.EncryptedData,
	}
	
	// Process service request
	response, err := dm.isvContract.ProcessServiceRequest(requestMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to process service request")
	}
	
	// Check status
	if response["status"] != "granted" {
		return nil, errors.Errorf("access denied: %s", response["status"])
	}
	
	// Create session
	session := &Session{
		SessionID: response["sessionID"],
		ClientID:  clientID,
		DeviceID:  deviceID,
		Status:    "active",
	}
	
	// Save session to file
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal session")
	}
	
	sessionFile := clientID + "-session-" + deviceID + ".json"
	if err := ioutil.WriteFile(sessionFile, sessionJSON, 0600); err != nil {
		return nil, errors.Wrap(err, "failed to save session to file")
	}
	
	log.Infof("Access granted to device %s, session ID: %s", deviceID, session.SessionID)
	return session, nil
}

// CloseSession closes an active session with a device
func (dm *DeviceManager) CloseSession(clientID, deviceID string) error {
	// Read session file
	sessionFile := clientID + "-session-" + deviceID + ".json"
	sessionJSON, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		return errors.Wrap(err, "failed to read session file")
	}
	
	// Parse session
	var session Session
	if err := json.Unmarshal(sessionJSON, &session); err != nil {
		return errors.Wrap(err, "failed to parse session")
	}
	
	// Close session
	if err := dm.isvContract.CloseSession(session.SessionID); err != nil {
		return errors.Wrap(err, "failed to close session")
	}
	
	// Remove session file
	if err := os.Remove(sessionFile); err != nil {
		log.Warnf("Failed to remove session file: %v", err)
	}
	
	log.Infof("Session with device %s closed", deviceID)
	return nil
}

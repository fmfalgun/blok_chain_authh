package auth

// NonceChallenge represents a nonce challenge from the Authentication Server
type NonceChallenge struct {
	Nonce          string `json:"nonce"`
	ExpirationTime int64  `json:"expirationTime"`
}

// TGT represents a Ticket Granting Ticket
type TGT struct {
	EncryptedTGT        string `json:"encryptedTGT"`
	EncryptedSessionKey string `json:"encryptedSessionKey"`
}

// ServiceTicket represents a service ticket for accessing a service
type ServiceTicket struct {
	EncryptedServiceTicket string `json:"encryptedServiceTicket"`
	EncryptedSessionKey    string `json:"encryptedSessionKey"`
}

// ServiceTicketRequest represents a request for a service ticket
type ServiceTicketRequest struct {
	EncryptedTGT   string `json:"encryptedTGT"`
	ClientID       string `json:"clientID"`
	ServiceID      string `json:"serviceID"`
	Authenticator  string `json:"authenticator"`
}

// ServiceRequest represents a request to access a service
type ServiceRequest struct {
	EncryptedServiceTicket string `json:"encryptedServiceTicket"`
	ClientID               string `json:"clientID"`
	DeviceID               string `json:"deviceID"`
	RequestType            string `json:"requestType"`
	EncryptedData          string `json:"encryptedData"`
}

// ServiceResponse represents a response to a service request
type ServiceResponse struct {
	ClientID      string `json:"clientID"`
	DeviceID      string `json:"deviceID"`
	Status        string `json:"status"`
	SessionID     string `json:"sessionID"`
	EncryptedData string `json:"encryptedData"`
}

// IoTDevice represents an IoT device registered with the ISV
type IoTDevice struct {
	DeviceID      string   `json:"deviceID"`
	Status        string   `json:"status"`
	LastSeen      string   `json:"lastSeen"`
	RegisteredAt  string   `json:"registeredAt"`
	Capabilities  []string `json:"capabilities"`
}

// Authenticator represents a timestamp encrypted with the session key
// Used to prove client identity to TGS
type Authenticator struct {
	ClientID  string `json:"clientID"`
	Timestamp int64  `json:"timestamp"`
}

// Session represents an active session between a client and a device
type Session struct {
	SessionID     string `json:"sessionID"`
	ClientID      string `json:"clientID"`
	DeviceID      string `json:"deviceID"`
	EstablishedAt string `json:"establishedAt"`
	ExpiresAt     string `json:"expiresAt"`
	Status        string `json:"status"`
}

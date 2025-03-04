package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// IoTAuth implements a simple chaincode
type IoTAuth struct {
}

// Init initializes chaincode
func (t *IoTAuth) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction
func (t *IoTAuth) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	
	if function == "registerClient" {
		return t.registerClient(stub, args)
	} else if function == "getClient" {
		return t.getClient(stub, args)
	}
	
	return shim.Error("Invalid function name")
}

// registerClient - register a new client
func (t *IoTAuth) registerClient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect arguments. Expecting client ID, public key, and organization")
	}
	
	clientID := args[0]
	publicKey := args[1]
	organization := args[2]
	
	client := struct {
		ID           string `json:"id"`
		PublicKey    string `json:"publicKey"`
		Organization string `json:"organization"`
	}{
		ID:           clientID,
		PublicKey:    publicKey,
		Organization: organization,
	}
	
	clientAsBytes, _ := json.Marshal(client)
	err := stub.PutState(clientID, clientAsBytes)
	if err != nil {
		return shim.Error("Failed to register client: " + err.Error())
	}
	
	return shim.Success(nil)
}

// getClient - get client by ID
func (t *IoTAuth) getClient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a client ID")
	}
	
	clientID := args[0]
	clientAsBytes, err := stub.GetState(clientID)
	
	if err != nil {
		return shim.Error("Failed to get client: " + err.Error())
	}
	
	if clientAsBytes == nil {
		return shim.Error("Client not found: " + clientID)
	}
	
	return shim.Success(clientAsBytes)
}

func main() {
	err := shim.Start(new(IoTAuth))
	if err != nil {
		fmt.Printf("Error starting IoT Auth chaincode: %s", err)
	}
}

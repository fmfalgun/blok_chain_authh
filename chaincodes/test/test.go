package main

import (
    "fmt"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type TestChaincode struct {
    contractapi.Contract
}

func (t *TestChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
    return nil
}

func (t *TestChaincode) Set(ctx contractapi.TransactionContextInterface, key string, value string) error {
    return ctx.GetStub().PutState(key, []byte(value))
}

func (t *TestChaincode) Get(ctx contractapi.TransactionContextInterface, key string) (string, error) {
    valueBytes, err := ctx.GetStub().GetState(key)
    if err != nil {
        return "", fmt.Errorf("failed to read from world state: %v", err)
    }
    if valueBytes == nil {
        return "", fmt.Errorf("key %s does not exist", key)
    }
    return string(valueBytes), nil
}

func main() {
    chaincode, err := contractapi.NewChaincode(&TestChaincode{})
    if err != nil {
        fmt.Printf("Error creating test chaincode: %s", err.Error())
        return
    }
    if err := chaincode.Start(); err != nil {
        fmt.Printf("Error starting test chaincode: %s", err.Error())
    }
}

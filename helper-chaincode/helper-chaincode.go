package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// HelperChaincode provides functions to store keys for other chaincodes
type HelperChaincode struct {
	contractapi.Contract
}

// StoreASKeys stores the AS keys in the blockchain state
func (s *HelperChaincode) StoreASKeys(ctx contractapi.TransactionContextInterface) error {
	// AS private key
	asPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAtOL3THYTwCk35h9/BYpX/5pQGH4jK5nyO55oI8PqBMx6GHfn
P0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4YTNNmV/IQKqX1eUtRJSP/gZR5/wQ0
6H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKigP48qIMYQYGXOSbL7IfvOWXV+TZ6o
9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMwlnUdZ/Yx4hOABHX6+0yQJxECU2OW
ve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnYFpV35FTy98dDYL7N6+y6whldMMQ6
80dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1TXQIDAQAB
-----END RSA PRIVATE KEY-----`

	// AS public key
	asPublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtOL3THYTwCk35h9/BYpX
/5pQGH4jK5nyO55oI8PqBMx6GHfnP0oG7+OgJQfNBsaPFoIzZuW7kRlv4x4jyG4Y
TNNmV/IQKqX1eUtRJSP/gZR5/wQ06H5722hLpzS8RCJQYnkGUcuEJA8xyBa8GKig
P48qIMYQYGXOSbL7IfvOWXV+TZ6o9mo/KcO88davW4IQ8LRHMIcODTY3iyDgLvMw
lnUdZ/Yx4hOABHX6+0yQJxECU2OWve3PaMAJCzqdKI4fDi4RZHwDpxP7+jrUYvnY
FpV35FTy98dDYL7N6+y6whldMMQ680dNMGqO2XyH5H3pY+H7y0K0em2OBCUmhB1T
XQIDAQAB
-----END PUBLIC KEY-----`
	
	// Store the keys
	err := ctx.GetStub().PutState("AS_PRIVATE_KEY", []byte(asPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to store AS private key: %s", err)
	}
	
	err = ctx.GetStub().PutState("AS_PUBLIC_KEY", []byte(asPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store AS public key: %s", err)
	}

	// Mark AS as initialized
	err = ctx.GetStub().PutState("AS_INITIALIZED", []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark AS as initialized: %s", err)
	}
	
	return nil
}

// StoreTGSKeys stores the TGS keys in the blockchain state
func (s *HelperChaincode) StoreTGSKeys(ctx contractapi.TransactionContextInterface) error {
	// TGS private key
	tgsPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA58L1zNrfqv6K6dNwBDLx23Qsl5qhQdLvxuJBLBcX5JeKJ/GG
HPoytB5MCgkBsk8/CM7BQpjx/CBmyT/7scVGHGbA6PYi8807ZvoZDl8dCk/Uxy1t
YRDeYVrQm2swwUhUTC9kIVYTBZtFzvZp//NybQHgOKHABbsf5EjEG7AOI2qiUzJN
RJPBzZtY0HdUoWYTWRTDiP/7yfVkm1PZsN+eYyWhPVdXQ1JLrGjjwOZl0db5QhcU
mXKjQWcy6/OMYsOjy4H7Mxtu7zGvPJObbTbkKPeh25P9jExLW8XXcxkv6RUbYf3I
AkDfMX8cJc3qtfcLW47Afywy0/zoLLQnQQVl3QIDAQABAoIBAHCIXUqM0fxOUMrL
S4q8omMGZfFXRWgoiRxKyQ1vXB5qMt47b5s4Zq4A41XPJ+LQ7kZADbQCXAuIGQHf
QzCHqkzYW9YL8n7TYBt8K2qVEVSHi/kHQVNLzfHpJPsy27s+o5pQ74AoRZQfblKt
3eBUm53jyHEGYnFlb9eZ5oBxSCEqq37jVZBvSUwx52IxNChjWW0JZwQdLVJ+Uqqs
wjHPl22U3l3QEcnQoQeQiARZQiQ4wP4lEWlUbNh5KnAQeMbvY9I+BsWnTygldUZD
qLzHz7foQWrl4d2XcA+mu3RlcB29lGmwgZVHzFEkKmDCIdcYUgKgcro4QXt+1B1i
TTvTrekCgYEA9v/Vbr6fHh+O8PQpwbVQgMOKqHRPHHPwUH47SOSHcRKwVYNZk0X/
FaRo2TrCkVRRnEo/vNVzYQT1XNxYQGKmKHqT4RbLLVYBVMXogTF0/W7uZJdcJOQV
MvzTxIES/w81TqXnrQYk6Vf38Fjc/uwYBXwOdWfJlCxLnBCy7WaZ5s8CgYEA8BcK
H9GyfsdxLBfH39YM9wz1Ilk5GlMPw+NLX8aYOzMF+zdgZeYJZ+12WHYTRwLRCpfG
6y+Nwt88q4L3NeSffrYR2QKbPo2P6hVPQGOaDLo4J/CkohFYDiLHnY4FXvBOhLz5
OGC+1MSr0XEGhFS9c7MS4zOVNGhGc+X7eEIKOzMCgYB62hzpn7JUdml6ljNZOK76
EK+oXfbFo+IovRn3a+bnJAJZyW4ypIK9KJVo5D4+KBqTtBCvY3c3MfFhCUje2xqj
1/I5afNLnd8ofhWCMzBi6DswS47yZJHLW7bWIZGFcmZfM38qmSTXw3OjJLqsrBw/
vTR6FbR4xcI2WxTN1t8HdQKBgCC2KgQc3NxJMtvwvUmA0KHPNyu3C/CNnIEbehsj
Uo7IWGBbKkKHjnNSjKzuoqjqP+vQ0HyYXPxlbR+8Rg3D0Jt3f/8aCRhD9jOUUhME
4M77ya9UJiWzVTqUEjVQB3k2M0BzKw+a/eHQC3D4qQ5GflZ7+P7QvHcYqBERKjFM
OFJPAoGBAMnUU7I3Qpo1n0HwBsQXoA1TgRcUMQQHp2/9XJP0K5L1FQvBMmhfeMQB
RA8g7GYJ3691Wy1GZ4YS1/QBZ9I69P0PYYxJXlaTZH9iEoAqvRcBoiXQgUkjI+TA
XJJc/DlIvuP0RBGJ4RYQJujO3fTMfUbVaQDJSQ5I8Ui/Yc4d1ZBE
-----END RSA PRIVATE KEY-----`

	// TGS public key
	tgsPublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA58L1zNrfqv6K6dNwBDLx
23Qsl5qhQdLvxuJBLBcX5JeKJ/GGHPoytB5MCgkBsk8/CM7BQpjx/CBmyT/7scVG
HGbA6PYi8807ZvoZDl8dCk/Uxy1tYRDeYVrQm2swwUhUTC9kIVYTBZtFzvZp//Ny
bQHgOKHABbsf5EjEG7AOI2qiUzJNRJPBzZtY0HdUoWYTWRTDiP/7yfVkm1PZsN+e
YyWhPVdXQ1JLrGjjwOZl0db5QhcUmXKjQWcy6/OMYsOjy4H7Mxtu7zGvPJObbTbk
KPeh25P9jExLW8XXcxkv6RUbYf3IAkDfMX8cJc3qtfcLW47Afywy0/zoLLQnQQVl
3QIDAQAB
-----END PUBLIC KEY-----`

	// ISV public key
	isvPublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApqAtGdmCJr3GYzs6fSQi
N1PO3GFiDtEAJyWbxRpKJRPv6/GGBLSqr5QQjDw7Vy1RwFXW7Z+j0/8C8xOBtu5J
UPoNBRJ5DMRyHGlGqxQgLjEySt8sObaJVq9WyHoNTLCD3lsmExxhhHM+ccc8dSZS
pX9qXAoHYvGZ0SJpGPBd7OXUQgzIUlJZRKP9Qz+d472xVMzpCrFJpPGkKcL1WoCP
GSgS3cx8NUb2xZnUHD1mmIyVwaDFm5RU4aBHrj/jx/tR9Dy0MKJC61/HAZEdU8zZ
c3kD/7PbsU0RXDzNzG8i8UtXSJYjgwBQhVlPn0/aQeiI7fk+Jf8E5zGtpKGI9L+R
CQIDAQAB
-----END PUBLIC KEY-----`
	
	// Store the keys
	err := ctx.GetStub().PutState("TGS_PRIVATE_KEY", []byte(tgsPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to store TGS private key: %s", err)
	}
	
	err = ctx.GetStub().PutState("TGS_PUBLIC_KEY", []byte(tgsPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store TGS public key: %s", err)
	}
	
	err = ctx.GetStub().PutState("ISV_PUBLIC_KEY", []byte(isvPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store ISV public key: %s", err)
	}

	// Mark TGS as initialized
	err = ctx.GetStub().PutState("TGS_INITIALIZED", []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark TGS as initialized: %s", err)
	}
	
	return nil
}

// StoreISVKeys stores the ISV keys in the blockchain state
func (s *HelperChaincode) StoreISVKeys(ctx contractapi.TransactionContextInterface) error {
	// ISV private key
	isvPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEApqAtGdmCJr3GYzs6fSQiN1PO3GFiDtEAJyWbxRpKJRPv6/GG
BLSqr5QQjDw7Vy1RwFXW7Z+j0/8C8xOBtu5JUPoNBRJ5DMRyHGlGqxQgLjEySt8s
ObaJVq9WyHoNTLCD3lsmExxhhHM+ccc8dSZSpX9qXAoHYvGZ0SJpGPBd7OXUQgzI
UlJZRKP9Qz+d472xVMzpCrFJpPGkKcL1WoCPGSgS3cx8NUb2xZnUHD1mmIyVwaDFm
5RU4aBHrj/jx/tR9Dy0MKJC61/HAZEdU8zZc3kD/7PbsU0RXDzNzG8i8UtXSJYjgw
BQhVlPn0/aQeiI7fk+Jf8E5zGtpKGI9L+RCQIDAQABAoIBAQCDXY4cG9Yf0sms7SV
SrES0F+abE1nYqCzE4/N9QZlrWDGkSvQj2Hj0iQwJxHKP5XSjBZLJw3ULqU8JwZN
L5JgbDhDNs0vCamT8nSEhP56/0PSJQfbXN8xB9tp8qGbIsdW5s/G2cK0qROJdT9C
e13Wd0c0jGxYqbbjIJDZygvUzFZXQVY6eymwXIxpWKl40ZkZtXIFMwIosP9/UitN
yBBJwgPK0iRxnBgydD1qIQYZbBL6IGUii73iLhLZvj1SNSGMdz0ni/A/dTNu878S
mlWlCkTOlFgJDmxb8d2JXkYxkQBAdRJk5FhFliW5qj5aprIbMqQzLcxLp/+n1bqR
c7vqd+NBAoGBAOQUYCZZ/yhNAooOQiBBfxj7SgI0PWsjNndwkJLCZtRCqm9Qm49p
oJOX1WsQDlc2QY1KG75+ms4Cq+EBxwY+lxEGaQSiA9BbHtWfHtSaXHayR9lRHNzL
FT+zdJJ+RkCdmSfL9upAo8/EPVn9CJV5wYXZZlXJaS/59lnqpxJSxI4JAoGBALuR
ufD62zl83TUJmW3gwBbQYKTxFkLGxGa0yZ5fLNDBFXfk4k/1xKxEX4MwbBLhDaQG
lxhLDK0jzgmFKP+VI8h5HwOgdBj03181+uEPGDHCQNqXu0XBsGHdztjIiqXM8OCR
J4ZYvyUjB5m0VzGoKQO66FMIjWVp7TqOfnwt19pRAoGAbkmX4iKJPzdH4wCo1rwd
N8DQZRQb3Blahm7zFdWF4a0IWjazZ/l+J7fieXwFEa9VvORJk8vgWwqHSyqLIT0h
y/kGcIhXMvqiBXPEYmA7GqvN1cjL8HnFXF5tL2FBLW1BO7nRU3B9VvlZ0W61+1CU
EZkdGSGjmWztZQ/8qfRCZmkCgYAqafBGZFnwB9EvWr4d4ZU3MFCpsa1tAroKimNz
e4TnZXDIjupKkIGMNRIJveT4IiYIoLKOXJ+Wjak28Ft7TZ45ldGS5QQEKSjxwIUD
6NtTGzwYL9FnYLxHFZ6PUWrgPNFNp4gpqrLQHZnRy9aCiGVXcRSKz1W8dSChUEsT
T/HfAQKBgQC+IFG/l3qPltDDxPo09QsH6LFpXCxLr5lyOwuFdZMMkmYSEHXcG/Z6
8cXP3kAmQCgAQbB2+T4CBJCceFC4LA6GOKrOg9IHPB8jrwmgpqAvt6OCJJRFJqgS
R0uXRj5xjUyNY4h9hnTB8Y0z23YaqnEa4/vQYHrI01YKldzfxPatvQ==
-----END RSA PRIVATE KEY-----`

	// ISV public key
	isvPublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApqAtGdmCJr3GYzs6fSQi
N1PO3GFiDtEAJyWbxRpKJRPv6/GGBLSqr5QQjDw7Vy1RwFXW7Z+j0/8C8xOBtu5J
UPoNBRJ5DMRyHGlGqxQgLjEySt8sObaJVq9WyHoNTLCD3lsmExxhhHM+ccc8dSZS
pX9qXAoHYvGZ0SJpGPBd7OXUQgzIUlJZRKP9Qz+d472xVMzpCrFJpPGkKcL1WoCP
GSgS3cx8NUb2xZnUHD1mmIyVwaDFm5RU4aBHrj/jx/tR9Dy0MKJC61/HAZEdU8zZ
c3kD/7PbsU0RXDzNzG8i8UtXSJYjgwBQhVlPn0/aQeiI7fk+Jf8E5zGtpKGI9L+R
CQIDAQAB
-----END PUBLIC KEY-----`
	
	// Store the keys
	err := ctx.GetStub().PutState("ISV_PRIVATE_KEY", []byte(isvPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to store ISV private key: %s", err)
	}
	
	err = ctx.GetStub().PutState("ISV_PUBLIC_KEY", []byte(isvPublicKey))
	if err != nil {
		return fmt.Errorf("failed to store ISV public key: %s", err)
	}

	// Mark ISV as initialized
	err = ctx.GetStub().PutState("ISV_INITIALIZED", []byte("true"))
	if err != nil {
		return fmt.Errorf("failed to mark ISV as initialized: %s", err)
	}
	
	return nil
}

// Initialize initializes the chaincode
func (s *HelperChaincode) Initialize(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Helper chaincode initialized")
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&HelperChaincode{})
	if err != nil {
		fmt.Printf("Error creating helper chaincode: %s", err.Error())
		return
	}
	
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting helper chaincode: %s", err.Error())
	}
} 
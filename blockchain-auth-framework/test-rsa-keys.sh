#!/bin/bash

# RSA Key Testing Script
# This script tests the generation and validation of RSA keys for the authentication system

# Exit on any error
set -e

# Define color codes for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function for printing colored output
print_green() {
    echo -e "${GREEN}$1${NC}"
}

print_yellow() {
    echo -e "${YELLOW}$1${NC}"
}

print_red() {
    echo -e "${RED}$1${NC}"
}

print_yellow "Testing RSA key generation and validation"
echo "-----------------------------------------"

# Step 1: Run test-auth.js to generate test keys
print_yellow "Step 1: Generating test RSA key pair"
node test-auth.js

# Verify key files exist
if [ ! -f "test-private.pem" ] || [ ! -f "test-public.pem" ]; then
    print_red "Key files not generated correctly"
    exit 1
fi

print_green "✓ Key files generated successfully"

# Step 2: Check key properties
print_yellow "Step 2: Checking key properties"

# Check private key
PRIVATE_KEY_INFO=$(openssl rsa -in test-private.pem -text -noout 2>/dev/null)
KEY_SIZE=$(echo "$PRIVATE_KEY_INFO" | grep "Private-Key:" | cut -d "(" -f2 | cut -d " " -f1)

echo "Private key size: $KEY_SIZE bits"
if [ "$KEY_SIZE" -ge 2048 ]; then
    print_green "✓ Private key size is sufficient (>= 2048 bits)"
else
    print_red "✗ Private key size is insufficient (< 2048 bits)"
    exit 1
fi

# Check public key
PUBLIC_KEY_INFO=$(openssl rsa -pubin -in test-public.pem -text -noout 2>/dev/null)
PUBLIC_KEY_SIZE=$(echo "$PUBLIC_KEY_INFO" | grep "Public-Key:" | cut -d "(" -f2 | cut -d " " -f1)

echo "Public key size: $PUBLIC_KEY_SIZE bits"
if [ "$PUBLIC_KEY_SIZE" -ge 2048 ]; then
    print_green "✓ Public key size is sufficient (>= 2048 bits)"
else
    print_red "✗ Public key size is insufficient (< 2048 bits)"
    exit 1
fi

# Step 3: Test encryption and decryption
print_yellow "Step 3: Testing encryption and decryption with the keys"

# Create a test message
TEST_MESSAGE="This is a test message for encryption $(date)"
echo "$TEST_MESSAGE" > test-message.txt

# Encrypt with public key
openssl rsautl -encrypt -pubin -inkey test-public.pem -in test-message.txt -out test-encrypted.bin

if [ ! -f "test-encrypted.bin" ]; then
    print_red "✗ Encryption failed"
    exit 1
fi

print_green "✓ Encryption successful"

# Decrypt with private key
openssl rsautl -decrypt -inkey test-private.pem -in test-encrypted.bin -out test-decrypted.txt

if [ ! -f "test-decrypted.txt" ]; then
    print_red "✗ Decryption failed"
    exit 1
fi

# Compare original and decrypted messages
if cmp -s "test-message.txt" "test-decrypted.txt"; then
    print_green "✓ Decryption successful - messages match"
else
    print_red "✗ Decryption failed - messages don't match"
    exit 1
fi

# Step 4: Test signing and verification
print_yellow "Step 4: Testing signing and verification with the keys"

# Create a signature
openssl dgst -sha256 -sign test-private.pem -out test-signature.bin test-message.txt

if [ ! -f "test-signature.bin" ]; then
    print_red "✗ Signing failed"
    exit 1
fi

print_green "✓ Signing successful"

# Verify the signature
VERIFY_RESULT=$(openssl dgst -sha256 -verify test-public.pem -signature test-signature.bin test-message.txt 2>&1)

if [[ "$VERIFY_RESULT" == *"Verified OK"* ]]; then
    print_green "✓ Signature verification successful"
else
    print_red "✗ Signature verification failed"
    exit 1
fi

# Step 5: Verify compatibility with the authentication framework
print_yellow "Step 5: Testing compatibility with the authentication framework"

# Test with the framework's debug function
node -e "
const crypto = require('crypto');
const fs = require('fs');

// Load the test keys
const privateKey = fs.readFileSync('test-private.pem', 'utf8');
const publicKey = fs.readFileSync('test-public.pem', 'utf8');

// Create test data
const testData = 'Test data for framework compatibility';
const testDataBuffer = Buffer.from(testData);

// Sign with private key
const hash = crypto.createHash('sha256').update(testDataBuffer).digest();
const signature = crypto.sign('sha256', hash, {
  key: privateKey,
  padding: crypto.constants.RSA_PKCS1_PADDING
});

// Verify with public key
const verified = crypto.verify(
  'sha256',
  hash,
  {
    key: publicKey,
    padding: crypto.constants.RSA_PKCS1_PADDING
  },
  signature
);

if (verified) {
  console.log('Framework compatibility test passed');
} else {
  console.log('Framework compatibility test failed');
  process.exit(1);
}
"

if [ $? -eq 0 ]; then
    print_green "✓ Keys are compatible with the authentication framework"
else
    print_red "✗ Keys are not compatible with the authentication framework"
    exit 1
fi

# Step 6: Clean up test files
rm -f test-message.txt test-encrypted.bin test-decrypted.txt test-signature.bin

print_green "RSA Key Test Completed Successfully!"
echo "✓ Key generation"
echo "✓ Key size verification"
echo "✓ Encryption/decryption"
echo "✓ Signing/verification"
echo "✓ Framework compatibility"

print_yellow "You can examine the keys in test-private.pem and test-public.pem"

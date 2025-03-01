#!/bin/bash

# Create symbolic links for CA private keys
for org in org1 org2 org3; do
  KEY_DIR="./organizations/peerOrganizations/${org}.example.com/ca"
  # Find the actual private key file (should start with numbers and end with _sk)
  PRIVATE_KEY=$(find ${KEY_DIR} -type f -name "*_sk")
  if [ -n "$PRIVATE_KEY" ]; then
    # Create symbolic link named 'priv_sk' pointing to the actual key file
    ln -sf $(basename "$PRIVATE_KEY") ${KEY_DIR}/priv_sk
    echo "Created symlink for ${org} CA key"
  fi
done

# For orderer CA
KEY_DIR="./organizations/ordererOrganizations/example.com/ca"
PRIVATE_KEY=$(find ${KEY_DIR} -type f -name "*_sk")
if [ -n "$PRIVATE_KEY" ]; then
  ln -sf $(basename "$PRIVATE_KEY") ${KEY_DIR}/priv_sk
  echo "Created symlink for orderer CA key"
fi

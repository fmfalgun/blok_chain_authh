#!/bin/bash

# Create a backup of the original file
cp ../config/connection-profile.json ../config/connection-profile.json.bak

# Fix the JSON directly with sed
# 1. Replace multiline certificates with escaped newlines
sed -i ':a;N;$!ba;s/-----BEGIN CERTIFICATE-----\n/-----BEGIN CERTIFICATE-----\\n/g' ../config/connection-profile.json
sed -i ':a;N;$!ba;s/\n-----END CERTIFICATE-----/\\n-----END CERTIFICATE-----/g' ../config/connection-profile.json
sed -i ':a;N;$!ba;s/CERTIFICATE-----\\n\([A-Za-z0-9+\/=]\+\)\n/CERTIFICATE-----\\n\1\\n/g' ../config/connection-profile.json

# 2. Fix organization name
sed -i 's/"organization": "Org1"/"organization": "Org1MSP"/g' ../config/connection-profile.json

# 3. Add discovery settings
sed -i 's/"endorser": "300"/"endorser": "300",\n                    "discover": {\n                        "enabled": true,\n                        "asLocalhost": true\n                    }/g' ../config/connection-profile.json

echo "Connection profile has been fixed."

#!/usr/bin/env python3

import json
import os

# Create a backup of the original file
os.system('cp ../config/connection-profile.json ../config/connection-profile.json.bak')

# Read the connection profile
with open('../config/connection-profile.json', 'r') as f:
    content = f.read()

# Load as JSON
connection_profile = json.loads(content)

# Fix organization name
connection_profile['client']['organization'] = 'Org1MSP'

# Add discovery settings
connection_profile['client']['connection']['timeout']['discover'] = {
    'enabled': True,
    'asLocalhost': True
}

# Fix certificates - replace newlines with \\n in all pem fields
def fix_certificates(obj):
    if isinstance(obj, dict):
        for key, value in obj.items():
            if key == 'pem' and isinstance(value, str) and '-----BEGIN CERTIFICATE-----' in value:
                # Convert multiline certificate to single line
                lines = value.strip().split('\n')
                obj[key] = '\\n'.join(lines)
            elif isinstance(value, (dict, list)):
                fix_certificates(value)
    elif isinstance(obj, list):
        for item in obj:
            fix_certificates(item)

fix_certificates(connection_profile)

# Write the fixed connection profile
with open('config/connection-profile.json', 'w') as f:
    json.dump(connection_profile, f, indent=4)

print("Connection profile has been fixed.")

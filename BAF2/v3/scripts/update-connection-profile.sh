#!/bin/bash
# Save as scripts/update-connection-profile.sh

# Define paths
CONNECTION_PROFILE="config/connection-profile.json"
TEMP_PROFILE="config/connection-profile.temp.json"
BACKUP_PROFILE="config/connection-profile.backup.json"

# Make a backup
cp "$CONNECTION_PROFILE" "$BACKUP_PROFILE"
echo "Created backup at $BACKUP_PROFILE"

# Create a temporary file
cp "$CONNECTION_PROFILE" "$TEMP_PROFILE"

# Find all certificate paths in the connection profile
CERT_PATHS=$(grep -A 2 "tlsCACerts" "$CONNECTION_PROFILE" | grep "path" | awk -F'"' '{print $4}')

# Process each path and replace with certificate content
for cert_path in $CERT_PATHS; do
    if [ -f "$cert_path" ]; then
        echo "Processing certificate: $cert_path"
        
        # Read certificate and escape newlines for JSON
        CERT_CONTENT=$(cat "$cert_path" | sed 's/$/\\n/' | tr -d '\n')
        
        # Create search and replace patterns
        SEARCH_PATTERN="\"path\": \"$cert_path\""
        REPLACE_PATTERN="\"pem\": \"$CERT_CONTENT\""
        
        # Replace in the temporary file
        sed -i "s|$SEARCH_PATTERN|$REPLACE_PATTERN|g" "$TEMP_PROFILE"
    else
        echo "Warning: Certificate file not found: $cert_path"
    fi
done

# Move the temporary file to the original
mv "$TEMP_PROFILE" "$CONNECTION_PROFILE"
echo "Updated connection profile with inline certificates"

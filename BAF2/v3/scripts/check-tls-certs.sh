#!/bin/bash
# Save as check-tls-certs.sh

CONNECTION_PROFILE="config/connection-profile.json"

# Extract certificate paths
CERT_PATHS=$(grep -A 2 "path" $CONNECTION_PROFILE | grep "path" | awk '{print $2}' | tr -d '",')

echo "Checking certificate paths:"
for path in $CERT_PATHS; do
    if [ -f "$path" ]; then
        echo "✅ $path exists"
        echo "   Certificate details:"
        openssl x509 -in "$path" -text -noout | grep "Subject:" -A 1
        echo ""
    else
        echo "❌ $path does not exist"
    fi
done

#!/bin/bash

# Script to add a new named authorized network to a Cloud SQL instance
# using the REST API to preserve network names

set -e

# Configuration
PROJECT_ID="${1:-$(gcloud config get-value project)}"
INSTANCE_ID="${2:-}"
API_URL="https://sqladmin.googleapis.com/sql/v1beta4/projects/${PROJECT_ID}/instances/${INSTANCE_ID}"

# Function to display usage
usage() {
    echo "Usage: $0 --name 'Network Name' --ip 'IP/CIDR'"
    echo "Example: $0 --name 'John Doe' --ip '192.168.1.100/32'"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --name)
            NETWORK_NAME="$2"
            shift 2
            ;;
        --ip)
            NETWORK_IP="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required arguments
if [ -z "$NETWORK_NAME" ] || [ -z "$NETWORK_IP" ]; then
    echo "Error: Both --name and --ip are required"
    usage
fi

# Validate IP/CIDR format (basic validation)
if [[ ! "$NETWORK_IP" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}(/[0-9]{1,2})?$ ]]; then
    echo "Error: Invalid IP/CIDR format: $NETWORK_IP"
    echo "Expected format: x.x.x.x/xx (e.g., 192.168.1.100/32)"
    exit 1
fi

# Add /32 if no CIDR specified
if [[ ! "$NETWORK_IP" =~ / ]]; then
    NETWORK_IP="${NETWORK_IP}/32"
fi

echo "Adding network: $NETWORK_NAME ($NETWORK_IP)"
echo "Project: $PROJECT_ID"
echo "Instance: $INSTANCE_ID"
echo

# Step 1: Get access token
echo "Step 1: Getting access token..."
ACCESS_TOKEN=$(gcloud auth print-access-token)

# Step 2: Get current instance configuration
echo "Step 2: Getting current instance configuration..."
curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    "$API_URL" > /tmp/current_config.json

# Step 3: Check if network already exists
echo "Step 3: Checking if network already exists..."
if jq -e --arg ip "$NETWORK_IP" '.settings.ipConfiguration.authorizedNetworks[] | select(.value == $ip)' /tmp/current_config.json > /dev/null; then
    echo "Network $NETWORK_IP already exists in authorized networks!"
    existing_name=$(jq -r --arg ip "$NETWORK_IP" '.settings.ipConfiguration.authorizedNetworks[] | select(.value == $ip) | .name' /tmp/current_config.json)
    echo "Existing name: $existing_name"
    echo "Requested name: $NETWORK_NAME"
    
    if [ "$existing_name" != "$NETWORK_NAME" ]; then
        echo "Warning: Network exists with different name. Updating name..."
    else
        echo "Network already exists with the same name. Nothing to do."
        exit 0
    fi
fi

# Step 4: Extract current authorized networks and add new one
echo "Step 4: Preparing updated authorized networks..."
jq --arg name "$NETWORK_NAME" --arg ip "$NETWORK_IP" '
    .settings.ipConfiguration.authorizedNetworks += [
        {
            "kind": "sql#aclEntry",
            "name": $name,
            "value": $ip
        }
    ] | .settings.ipConfiguration.authorizedNetworks |= unique_by(.value)
' /tmp/current_config.json > /tmp/updated_config.json

# Step 5: Create the patch request body (only ipConfiguration)
echo "Step 5: Creating patch request..."
jq '{
    settings: {
        ipConfiguration: .settings.ipConfiguration
    }
}' /tmp/updated_config.json > /tmp/patch_request.json

echo "New authorized networks configuration:"
jq -r '.settings.ipConfiguration.authorizedNetworks[] | "\(.name): \(.value)"' /tmp/patch_request.json

# Step 6: Apply the patch using REST API
echo
echo "Step 6: Applying changes to Cloud SQL instance..."
read -p "Do you want to proceed? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled."
    exit 0
fi

# Patch the instance using REST API
echo "Updating instance..."
response=$(curl -s -X PATCH \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d @/tmp/patch_request.json \
    "$API_URL")

# Check if the operation was successful
operation_name=$(echo "$response" | jq -r '.name // empty')
if [ -z "$operation_name" ]; then
    echo "❌ Error updating instance:"
    echo "$response" | jq .
    exit 1
fi

echo "✅ Operation started: $operation_name"
echo "⏳ Waiting for operation to complete..."

# Wait for operation to complete
operation_url="https://sqladmin.googleapis.com/sql/v1beta4/projects/${PROJECT_ID}/operations/${operation_name}"
while true; do
    op_status=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$operation_url" | jq -r '.status')
    if [ "$op_status" = "DONE" ]; then
        echo "✅ Operation completed successfully!"
        break
    elif [ "$op_status" = "RUNNING" ] || [ "$op_status" = "PENDING" ]; then
        echo "⏳ Status: $op_status"
        sleep 5
    else
        echo "❌ Operation failed with status: $op_status"
        exit 1
    fi
done

echo
echo "✅ Successfully added network: $NETWORK_NAME ($NETWORK_IP)"

# Verify the change
echo "Verifying authorized networks..."
curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    "$API_URL" | jq -r '.settings.ipConfiguration.authorizedNetworks[] | "\(.name): \(.value)"'

# Clean up temp files
rm -f /tmp/current_config.json /tmp/updated_config.json /tmp/patch_request.json

echo "Done!"
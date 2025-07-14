#!/bin/bash

# Script to add a new named authorized network to a Cloud SQL instance
# while preserving existing networks and their names

set -e

# Configuration - get from environment or use defaults
PROJECT_ID="${GCP_PROJECT_ID:-$(gcloud config get-value project)}"
INSTANCE_ID="${GCP_INSTANCE_ID:-}"

# Function to display usage
usage() {
    echo "Usage: $0 --name 'Network Name' --ip 'IP/CIDR' [--project PROJECT_ID] [--instance INSTANCE_ID]"
    echo "Example: $0 --name 'John Doe' --ip '192.168.1.100/32'"
    echo "         $0 --name 'Office' --ip '10.0.0.0/24' --project my-project --instance my-instance"
    echo ""
    echo "Environment variables:"
    echo "  GCP_PROJECT_ID   - Default project ID (current: ${PROJECT_ID})"
    echo "  GCP_INSTANCE_ID  - Default instance ID (current: ${INSTANCE_ID})"
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
        --project)
            PROJECT_ID="$2"
            shift 2
            ;;
        --instance)
            INSTANCE_ID="$2"
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

if [ -z "$PROJECT_ID" ] || [ -z "$INSTANCE_ID" ]; then
    echo "Error: Project ID and Instance ID are required"
    echo "Set them via --project/--instance flags or GCP_PROJECT_ID/GCP_INSTANCE_ID environment variables"
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

# Step 1: Get current instance configuration
echo "Step 1: Getting current instance configuration..."
gcloud sql instances describe "$INSTANCE_ID" --project="$PROJECT_ID" --format="json" > /tmp/current_config.json

# Step 2: Check if network already exists
echo "Step 2: Checking if network already exists..."
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

# Step 3: Extract current authorized networks and add new one
echo "Step 3: Preparing updated authorized networks..."
jq --arg name "$NETWORK_NAME" --arg ip "$NETWORK_IP" '
    .settings.ipConfiguration.authorizedNetworks += [
        {
            "kind": "sql#aclEntry",
            "name": $name,
            "value": $ip
        }
    ] | .settings.ipConfiguration.authorizedNetworks |= unique_by(.value)
' /tmp/current_config.json > /tmp/updated_networks.json

# Step 4: Create the patch request body
echo "Step 4: Creating patch request..."
jq '{
    settings: {
        ipConfiguration: {
            authorizedNetworks: .settings.ipConfiguration.authorizedNetworks
        }
    }
}' /tmp/updated_networks.json > /tmp/patch_request.json

echo "New authorized networks configuration:"
jq '.settings.ipConfiguration.authorizedNetworks[] | "\(.name): \(.value)"' /tmp/patch_request.json

# Step 5: Apply the patch using gcloud
echo
echo "Step 5: Applying changes to Cloud SQL instance..."
read -p "Do you want to proceed? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled."
    exit 0
fi

# Use gcloud to patch the instance using the REST API
echo "Updating instance..."
gcloud sql instances patch "$INSTANCE_ID" \
    --project="$PROJECT_ID" \
    --quiet \
    --format="json" \
    --authorized-networks="$(jq -r '.settings.ipConfiguration.authorizedNetworks[].value' /tmp/patch_request.json | tr '\n' ',' | sed 's/,$//')"

echo
echo "✅ Successfully added network: $NETWORK_NAME ($NETWORK_IP)"
echo "⚠️  Note: Network names are preserved in the backend but may not show in simple gcloud queries"

# Clean up temp files
rm -f /tmp/current_config.json /tmp/updated_networks.json /tmp/patch_request.json

echo "Done!"
#!/bin/bash

# Migration script from sql-network-manager to piam-anc

echo "🔄 Migrating from sql-network-manager to piam-anc..."

# Check if old tool is installed via Homebrew
if brew list exclamationlabs/piam/sql-network-manager &>/dev/null; then
    echo "📦 Found sql-network-manager installed via Homebrew"
    echo "🗑️  Uninstalling old version..."
    brew uninstall exclamationlabs/piam/sql-network-manager
else
    echo "✅ sql-network-manager not found in Homebrew"
fi

# Check if old binary exists in common locations
OLD_LOCATIONS=(
    "/usr/local/bin/sql-network-manager"
    "$HOME/.local/bin/sql-network-manager"
    "$HOME/bin/sql-network-manager"
)

for location in "${OLD_LOCATIONS[@]}"; do
    if [ -f "$location" ]; then
        echo "🗑️  Removing old binary at $location"
        rm -f "$location"
    fi
done

# Install new version
echo "📦 Installing piam-anc..."
brew tap exclamationlabs/piam
brew install exclamationlabs/piam/piam-anc

# Verify installation
if command -v piam-anc &>/dev/null; then
    echo "✅ Successfully installed piam-anc!"
    echo ""
    piam-anc --version
    echo ""
    echo "🎉 Migration complete! You can now use 'piam-anc' command."
    echo ""
    echo "What's new:"
    echo "  • Supports both SQL instances and GKE clusters"
    echo "  • Press 'c' to open resources in Google Cloud Console"
    echo "  • Auto-fills your username and public IP"
    echo "  • Better search and visual improvements"
    echo ""
    echo "Run 'piam-anc' to get started!"
else
    echo "❌ Installation failed. Please try manually:"
    echo "  brew install exclamationlabs/piam/piam-anc"
fi
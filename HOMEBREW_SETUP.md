# Homebrew Setup Instructions

## For the Tap Maintainer (You)

### 1. Update Your Homebrew Tap Repository

Clone or navigate to your existing tap:
```bash
git clone https://github.com/ExclamationLabs/homebrew-piam.git
cd homebrew-piam
```

### 2. Add the New Formula

Copy the `piam-anc.rb` formula file to the Formula directory:
```bash
cp /path/to/piam-anc/piam-anc.rb Formula/piam-anc.rb
```

### 3. Optional: Deprecate the Old Formula

If you want to keep sql-network-manager during transition, update it:
```ruby
# In Formula/sql-network-manager.rb, add at the top:
class SqlNetworkManager < Formula
  desc "DEPRECATED: Please use piam-anc instead"
  deprecate! date: "2024-01-14", because: "renamed to piam-anc"
  # ... rest of formula
end
```

### 4. Commit and Push

```bash
git add Formula/piam-anc.rb
git commit -m "Add piam-anc formula v1.0.0"
git push origin main
```

## For Your Teammates (The Surprise! 🎉)

Once you've pushed the formula, they can install with:

```bash
# If they haven't tapped before
brew tap exclamationlabs/piam

# Install the app
brew install exclamationlabs/piam/piam-anc
```

Or in one line:
```bash
brew install exclamationlabs/piam/piam-anc
```

### Upgrading from sql-network-manager

If they had the old tool installed:
```bash
brew uninstall exclamationlabs/piam/sql-network-manager
brew install exclamationlabs/piam/piam-anc
```

## Testing the Formula Locally

Before pushing, you can test locally:
```bash
brew install --build-from-source ./piam-anc.rb
```

## Quick Verification

After installation, they can verify with:
```bash
piam-anc --version
# Should show: piam-anc version 1.0.0
```

## The Surprise Commands 🎁

Send this to your team:
```
Hey team! The new network management tool is ready for testing.

Installation is now super simple:
  brew install exclamationlabs/piam/piam-anc

Then just run:
  piam-anc

It now supports both SQL instances AND GKE clusters, with a bunch of improvements!
Press 'c' to open resources in the Cloud Console, and it auto-fills your IP address.

Let me know what you think!
```
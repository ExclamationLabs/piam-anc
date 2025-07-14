# GitHub Repository Setup

## Step 1: Create the Repository

1. Go to https://github.com/organizations/ExclamationLabs/repositories/new
2. Repository name: `piam-anc`
3. Description: "Beautiful TUI for managing Google Cloud SQL and GKE authorized networks"
4. Make it Public (or Private if preferred)
5. **DO NOT** initialize with README, .gitignore, or license (we already have these)
6. Click "Create repository"

## Step 2: Push the Code

After creating the empty repository, run these commands:

```bash
# We already have the remote set up, so just push
git push -u origin main
```

## Step 3: Create the Release

1. Go to https://github.com/ExclamationLabs/piam-anc/releases/new
2. Click "Choose a tag" and type `v1.0.0`
3. Release title: `PIAM Admin Network Configurator v1.0.0`
4. Copy the contents of `RELEASE_NOTES_v1.0.0.md` into the description
5. Upload these files from the `dist/` directory:
   - `piam-anc-darwin-amd64.tar.gz`
   - `piam-anc-darwin-arm64.tar.gz`
   - `piam-anc-linux-amd64.tar.gz`
   - `piam-anc-linux-arm64.tar.gz`
   - `checksums.txt`
6. Check "Set as the latest release"
7. Click "Publish release"

## Step 4: Update Homebrew Tap

1. Clone your Homebrew tap:
   ```bash
   git clone git@github.com:ExclamationLabs/homebrew-piam.git
   cd homebrew-piam
   ```

2. Copy the formula:
   ```bash
   cp ../piam-anc/piam-anc.rb Formula/piam-anc.rb
   ```

3. Commit and push:
   ```bash
   git add Formula/piam-anc.rb
   git commit -m "Add piam-anc formula v1.0.0

   PIAM Admin Network Configurator - successor to sql-network-manager
   - Supports both Cloud SQL and GKE clusters
   - Google Cloud Console integration
   - Auto-populates user info
   - Improved UI and performance"
   
   git push origin main
   ```

## Step 5: Test Installation

```bash
# Test the Homebrew installation
brew update
brew install exclamationlabs/piam/piam-anc

# Verify it works
piam-anc --version
```

## Step 6: Notify the Team

Send this message:
```
🎉 New Network Management Tool Ready!

I've just released PIAM Admin Network Configurator (piam-anc) v1.0.0!
This is the successor to sql-network-manager with major improvements.

Installation:
  brew install exclamationlabs/piam/piam-anc

What's New:
  ✨ Supports both SQL instances AND GKE clusters
  🌐 Press 'c' to open any resource in Google Cloud Console
  🤖 Auto-fills your username and public IP
  🎨 Visual improvements and better search
  ⏱️ Shows progress during operations

Just run 'piam-anc' after installing. Let me know what you think!

If you have the old tool installed, you can run:
  brew uninstall exclamationlabs/piam/sql-network-manager
```
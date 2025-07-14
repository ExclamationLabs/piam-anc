# Release Notes - PIAM Admin Network Configurator v1.0.0

## 🎉 First Major Release!

We're excited to announce the release of PIAM Admin Network Configurator (piam-anc) v1.0.0! This release represents a complete rewrite and rebranding of the former sql-network-manager tool, now with support for both Cloud SQL instances and GKE clusters.

## ✨ Highlights

### 🔐 Unified Network Management
Manage authorized networks for both Cloud SQL instances and GKE clusters from a single, beautiful terminal interface.

### 🌐 Google Cloud Console Integration
Press 'c' to instantly open any resource in the Google Cloud Console for advanced management tasks.

### 🤖 Smart Auto-Population
The tool now automatically fills in your username and public IP address (with /32 CIDR) when adding networks, saving time and reducing errors.

### 📊 Visual Resource Differentiation
SQL instances and GKE clusters have distinct background colors, making it easy to identify resource types at a glance.

### ⏱️ Real-Time Progress Tracking
See exactly how long network operations are taking with a live countdown timer. The tool now clearly indicates that GCP operations may take up to 60 seconds.

## 🚀 Key Features

- **Multi-Resource Support**: Manage both Cloud SQL and GKE resources
- **Parallel Discovery**: Lightning-fast resource discovery across ALL your projects
- **Smart Access Detection**: Clear warnings for resources that cannot accept external networks
- **Beautiful UI**: Catppuccin Mocha themed interface
- **Add-Only Design**: Focused on adding networks (removal must be done via Console)
- **Improved Search**: More accurate fuzzy search with better result ranking

## 🔧 Technical Improvements

- Removed 100-project discovery limit
- Increased parallel discovery to 20 concurrent requests
- Reduced operation timeout from 5 minutes to 30 seconds
- Fixed hanging "Adding network..." operations
- Improved form focus handling to prevent accidental input

## 📦 Installation

### Homebrew (Coming Soon)
```bash
brew install piam-anc
```

### From Source
```bash
git clone https://github.com/ExclamationLabs/piam-anc.git
cd piam-anc
go build -o piam-anc
sudo mv piam-anc /usr/local/bin/
```

### Pre-built Binaries
Download from the [releases page](https://github.com/ExclamationLabs/piam-anc/releases/tag/v1.0.0).

## 🙏 Acknowledgments

Thanks to all the users who provided feedback on the original sql-network-manager tool. Your input has been invaluable in shaping this release.

## 📝 Full Changelog

See [CHANGELOG.md](https://github.com/ExclamationLabs/piam-anc/blob/main/CHANGELOG.md) for a detailed list of all changes.

---

**Note**: This tool requires appropriate Google Cloud permissions:
- `cloudsql.instances.list/get/update`
- `container.clusters.list/get/update`
- `resourcemanager.projects.list`
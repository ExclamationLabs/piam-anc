# 🔐 PIAM Admin Network Configurator (piam-anc)

A beautiful TUI for managing Google Cloud SQL and GKE authorized networks across all your projects.

## ✨ Features

- 🎨 **Beautiful TUI** with Catppuccin Mocha theme
- 🔍 **Parallel Multi-Resource Discovery** - Lightning-fast discovery of SQL instances and GKE clusters across ALL your projects
- 📋 **Unified Interface** - Manage both Cloud SQL and GKE authorized networks in one place
- 🗄️ **SQL Instance Support** - View and manage authorized networks for Cloud SQL instances
- ☸️ **GKE Cluster Support** - Manage master authorized networks for Kubernetes clusters
- 🔒 **Smart Access Detection** - Shows which resources can accept external networks
- 🌐 **Google Cloud Console Integration** - Open resources directly in the console (press 'c')
- 🤖 **Auto-Population** - Automatically fills your username and public IP when adding networks
- 👥 **View All Networks** - See everyone's authorized networks in a beautiful table
- ➕ **Add Networks Only** - Users can only add their own networks (no removal)
- 🔒 **Preserves Names** - Unlike gcloud CLI, maintains human-readable network names
- ⏱️ **Progress Tracking** - Live timer shows operation progress (GCP may take up to 60s)
- ⚡ **Real-time Updates** - Instant feedback and smooth loading states
- 🎯 **Smart Validation** - Validates IP/CIDR format with helpful error messages

## 🆕 What's New in v1.0.0

- **Google Cloud Console Integration** - Press 'c' to manage resources in the web console
- **Auto-Population** - Your username and public IP are automatically filled
- **Visual Resource Types** - SQL and GKE resources have distinct colors
- **Progress Timer** - See how long operations are taking
- **Better Search** - More accurate fuzzy search results
- See [CHANGELOG.md](CHANGELOG.md) for full details

## 🚀 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ExclamationLabs/piam-anc.git
cd piam-anc

# Build the binary
go build -o piam-anc

# Install to /usr/local/bin
sudo mv piam-anc /usr/local/bin/
```

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/ExclamationLabs/piam-anc/releases).

## 🎬 Demo

![PIAM ANC Demo](demo.gif)

To create your own demo recording:
```bash
# Install VHS
brew install vhs

# Record the demo
vhs demo.tape
```

## 🎯 Usage

### Prerequisites

1. **Google Cloud Authentication:**
   ```bash
   gcloud auth application-default login
   ```

2. **Required Permissions:**
   - `cloudsql.instances.list`
   - `cloudsql.instances.get`  
   - `cloudsql.instances.update`
   - `container.clusters.list`
   - `container.clusters.get`
   - `container.clusters.update`
   - `resourcemanager.projects.list`

### Running the Application

```bash
piam-anc
```

### Navigation

- **↑/↓** - Navigate through lists
- **Enter** - Select resource
- **/** - Search resources (fuzzy search by name, project, region)
- **a** - Add authorized network (when available)
- **c** - Open resource in Google Cloud Console
- **r** - Refresh resource list
- **Esc** - Go back
- **q** - Quit
- **?** - Show help

### Resource Icons

- 🗄️ **SQL Database** - Cloud SQL instance
- ☸️ **GKE Cluster** - Kubernetes cluster
- 🔒 **Locked** - Resource cannot accept external networks

## 🔍 Resource Detection

The tool automatically detects:

### For SQL Instances:
- **Public IP Status** - Only instances with public IPs can have authorized networks
- **Private-only Instances** - Shows warning that external networks cannot be added
- **Connection Names** - Full instance connection string

### For GKE Clusters:
- **Private Cluster Status** - Detects if cluster has private nodes
- **Master Authorized Networks** - Shows existing authorized networks
- **Public/Private Endpoints** - Displays both endpoints when available

## 📋 Network Restrictions

### SQL Instances
- ❌ **Private IP only** - Cannot add authorized networks
- ✅ **Public IP enabled** - Can add authorized networks

### GKE Clusters
- ✅ **Always supported** - Master authorized networks can always be configured
- ⚠️ **Private clusters** - May require VPN or jumphost for actual access

## 🏗️ Architecture

```
piam-anc/
├── main.go           # Application entry point
├── models.go         # Data models and API interactions
├── tui.go           # Terminal UI implementation
├── theme.go         # Catppuccin Mocha theme
└── build.sh         # Cross-platform build script
```

## 🔧 Configuration

The app automatically discovers all resources across your accessible projects. No manual configuration needed!

## 🚨 Problem Solved

Managing network access for cloud resources is painful:

### Cloud SQL Issues:
- `gcloud sql instances patch` **replaces** the entire authorized networks list
- No way to incrementally add networks
- Loses human-readable network names
- No visibility into private vs public instances

### GKE Issues:
- Complex API for updating master authorized networks
- No unified interface with SQL instances
- Hard to see which clusters need jumphost access

Our solution:
- ✅ Unified interface for both SQL and GKE
- ✅ Preserves existing network names
- ✅ Adds networks incrementally
- ✅ Shows access restrictions clearly
- ✅ Beautiful, user-friendly interface
- ✅ Works across ALL your projects in parallel

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
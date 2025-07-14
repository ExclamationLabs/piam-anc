# Changelog

All notable changes to PIAM Admin Network Configurator (piam-anc) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-01-14

### Added
- **Multi-Resource Support**: Manage both Cloud SQL instances and GKE clusters in one unified interface
- **Google Cloud Console Integration**: Press 'c' to open any resource directly in the Google Cloud Console
- **Auto-Population**: Automatically fills in your username and public IP address (with /32) when adding networks
- **Visual Resource Differentiation**: SQL instances and GKE clusters have different background colors for easy identification
- **Network Restrictions Display**: Clear warnings for resources that cannot accept external networks (private-only instances)
- **Progress Timer**: Shows elapsed time when adding networks with clear indication that GCP operations may take up to 60 seconds
- **Smart Form Focus**: Prevents accidental typing when opening the add network form
- **Improved Fuzzy Search**: More accurate search results with less aggressive matching
- **Enhanced Error Handling**: Better error messages and recovery from failed operations

### Changed
- **Project Rename**: Changed from `sql-network-manager` to `piam-anc` (PIAM Admin Network Configurator)
- **Resource Discovery**: Now discovers ALL resources across ALL accessible projects (removed 100-project limit)
- **UI Improvements**: Better alignment, consistent spacing, and clearer visual hierarchy
- **Performance**: Increased parallel discovery to 20 concurrent requests for faster loading
- **Timeout Optimization**: Reduced network operation timeout from 5 minutes to 30 seconds

### Removed
- **Command Line Search**: Removed broken command-line search feature (use '/' within the app instead)

### Fixed
- **Hanging Operations**: Fixed issue where "Adding network..." would hang indefinitely
- **Missing Resources**: Fixed issue where some resources weren't discovered due to project limits
- **Form Input Issues**: Fixed 'a' key being typed into name field when opening add network form
- **State Management**: Improved state transitions and error recovery

### Security
- Automatic retrieval of public IP from secure ipinfo.io endpoint
- No storage of credentials or sensitive information
- Read-only access to Google Cloud Console (modifications require explicit user action)

## [0.1.0] - 2024-01-10

### Added
- Initial release as `sql-network-manager`
- Basic Cloud SQL instance network management
- TUI interface with Catppuccin Mocha theme
- Fuzzy search functionality
- Network addition capability
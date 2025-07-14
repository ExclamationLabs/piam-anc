package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printHelp()
			return
		case "-v", "--version", "version":
			printVersion()
			return
		default:
			fmt.Printf("Unknown argument: %s\n", os.Args[1])
			fmt.Println("Use --help for usage information.")
			os.Exit(1)
		}
	}

	// Create the program
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func printVersion() {
	fmt.Println("piam-anc version 1.0.0")
	fmt.Println("PIAM Admin Network Configurator")
	fmt.Println("Built with ❤️ using Charm Bracelet Bubble Tea")
}

func printHelp() {
	fmt.Println(`
🔐 PIAM Admin Network Configurator (piam-anc)

A beautiful TUI for managing Cloud SQL and GKE authorized networks across all your Google Cloud projects.

USAGE:
  piam-anc [FLAGS]

FLAGS:
  -h, --help     Show this help message
  -v, --version  Show version information

EXAMPLES:
  piam-anc                    # Launch the application
  piam-anc --help            # Show this help
  piam-anc --version         # Show version information

FEATURES:
  🗄️  SQL Instance Management - View and manage authorized networks
  ☸️  GKE Cluster Support - Manage master authorized networks  
  🔍 Multi-Project Discovery - Finds ALL resources across your projects
  🔒 Smart Access Detection - Shows which resources accept external networks
  🎨 Beautiful Interface - Catppuccin Mocha themed TUI
  ⚡ Fast Parallel Discovery - Lightning-fast resource scanning

NAVIGATION:
  ↑/↓     Navigate lists
  Enter   Select resource
  /       Search/filter resources
  a       Add authorized network (when available)
  c       Open resource in Google Cloud Console
  r       Refresh resource list
  Esc     Go back
  ?       Show help
  q       Quit

RESOURCE INDICATORS:
  🗄️      SQL Database instance
  ☸️      GKE Kubernetes cluster  
  🔒      Resource cannot accept external networks (private)

REQUIREMENTS:
  • Google Cloud SDK (gcloud) authenticated
  • Required permissions:
    - cloudsql.instances.list/get/update
    - container.clusters.list/get/update
    - resourcemanager.projects.list

AUTHENTICATION:
  Before using, ensure you're authenticated:
    gcloud auth application-default login

For more information, visit: https://github.com/ExclamationLabs/piam-anc
`)
}
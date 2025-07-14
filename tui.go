package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application states
type sessionState int

const (
	stateLoading sessionState = iota
	stateResourceSelection
	stateNetworkView
	stateAddNetwork
	stateError
)

// Model represents the main TUI model
type Model struct {
	state           sessionState
	width           int
	height          int
	spinner         spinner.Model
	resourceList    list.Model
	networkManager  *NetworkManager
	selectedResource CloudResource
	resources       []CloudResource
	
	// Add network form
	nameInput    textinput.Model
	ipInput      textinput.Model
	addFormFocus int
	
	// Status and errors
	message     string
	isError     bool
	isLoading   bool
	isSubmitting bool
	submitStartTime time.Time
	
	// Navigation
	activeTab   int
	showHelp    bool
	
}

// List item for resources
type resourceItem struct {
	resource CloudResource
}

func (r resourceItem) FilterValue() string {
	// More focused search - prioritize name and project, but make region and type less prominent
	return fmt.Sprintf("%s %s %s_%s", 
		r.resource.GetName(), 
		r.resource.GetProject(), 
		r.resource.GetRegion(),
		r.resource.GetType())
}

func (r resourceItem) Title() string {
	title := fmt.Sprintf("%s %s (%s)", getResourceIcon(r.resource), r.resource.GetName(), r.resource.GetProject())
	if !r.resource.CanAddNetwork() {
		title += " 🔒"
	}
	return title
}

func (r resourceItem) Description() string {
	// Format with consistent width for alignment
	region := fmt.Sprintf("%-12s", r.resource.GetRegion())
	resourceType := fmt.Sprintf("%-3s", r.resource.GetType())
	
	var networkCount int
	var desc string
	
	// Get network count and apply resource-specific styling
	switch res := r.resource.(type) {
	case SQLInstance:
		networkCount = len(res.AuthorizedNetworks)
		desc = fmt.Sprintf("%s • %s • %d networks", region, resourceType, networkCount)
		// Apply SQL-specific background styling
		desc = lipgloss.NewStyle().
			Background(lipgloss.Color(CatppuccinMocha.Surface0)).
			Foreground(lipgloss.Color(CatppuccinMocha.Blue)).
			Padding(0, 1).
			Render(desc)
	case GKECluster:
		networkCount = len(res.MasterAuthorizedNetworks)
		desc = fmt.Sprintf("%s • %s • %d networks", region, resourceType, networkCount)
		// Apply GKE-specific background styling
		desc = lipgloss.NewStyle().
			Background(lipgloss.Color(CatppuccinMocha.Surface1)).
			Foreground(lipgloss.Color(CatppuccinMocha.Mauve)).
			Padding(0, 1).
			Render(desc)
	default:
		desc = fmt.Sprintf("%s • %s", region, resourceType)
	}
	
	// Add restrictions if any
	if restrictions := r.resource.GetNetworkRestrictions(); restrictions != "" {
		desc += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(CatppuccinMocha.Yellow)).Render("⚠️  " + restrictions)
	}
	
	return desc
}

// getPublicIP fetches the user's public IP address
func getPublicIP() string {
	resp, err := http.Get("https://ipinfo.io/ip")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	ip := strings.TrimSpace(string(body))
	if ip != "" {
		return ip + "/32"
	}
	return ""
}

// getUserName gets the current user's name
func getUserName() string {
	// Try environment variable first
	if username := os.Getenv("USER"); username != "" {
		return username
	}
	
	// Fallback to user package
	if currentUser, err := user.Current(); err == nil {
		if currentUser.Username != "" {
			return currentUser.Username
		}
		if currentUser.Name != "" {
			return currentUser.Name
		}
	}
	
	return ""
}

// getConsoleURL generates the Google Cloud Console URL for a resource
func getConsoleURL(resource CloudResource) string {
	switch r := resource.(type) {
	case SQLInstance:
		return fmt.Sprintf("https://console.cloud.google.com/sql/instances/%s/edit?project=%s", 
			r.Name, r.Project)
	case GKECluster:
		return fmt.Sprintf("https://console.cloud.google.com/kubernetes/clusters/details/%s/%s?project=%s", 
			r.Location, r.Name, r.Project)
	default:
		return ""
	}
}

func getResourceIcon(resource CloudResource) string {
	switch resource.GetType() {
	case ResourceTypeSQL:
		return "🗄️"
	case ResourceTypeGKE:
		return "☸️"
	default:
		return "📦"
	}
}

// Messages
type resourcesLoadedMsg struct {
	resources []CloudResource
}

type resourceSelectedMsg struct {
	resource CloudResource
}

type networkAddedMsg struct {
	success bool
	message string
}

type errorMsg struct {
	err error
}

type tickMsg time.Time


// Initialize model
func initialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle
	
	// Create name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter your name..."
	nameInput.CharLimit = 50
	nameInput.Width = 30
	nameInput.Focus()
	
	// Create IP input
	ipInput := textinput.New()
	ipInput.Placeholder = "192.168.1.100/32"
	ipInput.CharLimit = 20
	ipInput.Width = 20
	
	// Create resource list
	resourceList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	resourceList.Title = "Select Cloud Resource (type to search)"
	resourceList.SetShowStatusBar(false)
	resourceList.SetFilteringEnabled(true)
	
	return Model{
		state:        stateLoading,
		spinner:      s,
		resourceList: resourceList,
		nameInput:    nameInput,
		ipInput:      ipInput,
	}
}

// Init command
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadResources,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resourceList.SetWidth(msg.Width - 4)
		m.resourceList.SetHeight(msg.Height - 10)
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			if m.state != stateLoading && m.state != stateError {
				m.showHelp = !m.showHelp
			}
		case "esc":
			if m.showHelp {
				m.showHelp = false
			} else if m.state == stateNetworkView || m.state == stateAddNetwork {
				m.state = stateResourceSelection
				m.message = ""
			}
		case "a":
			if m.state == stateNetworkView {
				if m.selectedResource.CanAddNetwork() {
					m.state = stateAddNetwork
					m.addFormFocus = 0
					
					// Auto-populate with user's name and public IP
					m.nameInput.Reset()
					m.ipInput.Reset()
					
					if username := getUserName(); username != "" {
						m.nameInput.SetValue(username)
					}
					
					if publicIP := getPublicIP(); publicIP != "" {
						m.ipInput.SetValue(publicIP)
					}
					
					// Don't auto-focus to prevent 'a' from being typed in field
					m.nameInput.Blur()
					m.ipInput.Blur()
					m.addFormFocus = -1 // No field focused initially
				} else {
					m.message = "Cannot add networks to this resource: " + m.selectedResource.GetNetworkRestrictions()
					m.isError = true
				}
			}
		case "c":
			if m.state == stateNetworkView {
				// Open console URL
				return m, openConsoleURL(m.selectedResource)
			}
		case "r":
			if m.state == stateResourceSelection || m.state == stateNetworkView {
				m.state = stateLoading
				return m, loadResources
			}
		case "enter":
			if m.state == stateResourceSelection {
				if i, ok := m.resourceList.SelectedItem().(resourceItem); ok {
					return m, selectResource(i.resource)
				}
			} else if m.state == stateAddNetwork {
				if m.addFormFocus == -1 {
					// First enter focuses name field
					m.addFormFocus = 0
					m.nameInput.Focus()
					m.ipInput.Blur()
				} else if m.addFormFocus == 0 {
					m.addFormFocus = 1
					m.nameInput.Blur()
					m.ipInput.Focus()
				} else if m.addFormFocus == 1 {
					// Show loading state
					m.message = "Adding network... (0s) - GCP may take up to 60 seconds"
					m.isError = false
					m.isSubmitting = true
					m.submitStartTime = time.Now()
					return m, tea.Batch(
						submitAddNetwork(m.selectedResource, m.nameInput.Value(), m.ipInput.Value()),
						tickCmd(),
					)
				}
			}
		case "tab":
			if m.state == stateAddNetwork {
				if m.addFormFocus == -1 {
					// First tab focuses name field
					m.addFormFocus = 0
					m.nameInput.Focus()
					m.ipInput.Blur()
				} else if m.addFormFocus == 0 {
					m.addFormFocus = 1
					m.nameInput.Blur()
					m.ipInput.Focus()
				} else {
					m.addFormFocus = 0
					m.ipInput.Blur()
					m.nameInput.Focus()
				}
			}
		}
		
	case resourcesLoadedMsg:
		m.isLoading = false
		m.resources = msg.resources
		
		items := make([]list.Item, len(msg.resources))
		for i, resource := range msg.resources {
			items[i] = resourceItem{resource: resource}
		}
		
		m.resourceList.SetItems(items)
		m.state = stateResourceSelection
		
	case resourceSelectedMsg:
		m.selectedResource = msg.resource
		m.state = stateNetworkView
		m.message = ""
		
	case networkAddedMsg:
		// Clear submitting state
		m.isSubmitting = false
		
		if msg.success {
			// Show success message and go back to network view
			m.message = msg.message
			m.isError = false
			m.state = stateNetworkView
			// Refresh the selected resource to show updated networks
			return m, selectResource(m.selectedResource)
		} else {
			// Show error message but stay in add form
			m.message = msg.message
			m.isError = true
		}
		
	case errorMsg:
		m.message = msg.err.Error()
		m.isError = true
		m.isSubmitting = false // Clear submitting state on error
		// Don't change state to error if we're in add network form
		if m.state != stateAddNetwork {
			m.state = stateError
		}
		
	case tickMsg:
		if m.isSubmitting && m.state == stateAddNetwork {
			// Update the message with elapsed time
			elapsed := time.Since(m.submitStartTime).Seconds()
			m.message = fmt.Sprintf("Adding network... (%.0fs) - GCP may take up to 60 seconds", elapsed)
			// Continue ticking
			return m, tickCmd()
		}
		
	}
	
	// Update components
	switch m.state {
	case stateLoading:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		
	case stateResourceSelection:
		m.resourceList, cmd = m.resourceList.Update(msg)
		cmds = append(cmds, cmd)
		
	case stateAddNetwork:
		// Don't update inputs when submitting
		if !m.isSubmitting {
			if m.addFormFocus == 0 {
				m.nameInput, cmd = m.nameInput.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.addFormFocus == 1 {
				m.ipInput, cmd = m.ipInput.Update(msg)
				cmds = append(cmds, cmd)
			}
			// Don't update inputs when addFormFocus == -1 (no field focused)
		}
	}
	
	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelpView()
	}
	
	var content string
	
	switch m.state {
	case stateLoading:
		content = m.renderLoadingView()
	case stateResourceSelection:
		content = m.renderResourceSelectionView()
	case stateNetworkView:
		content = m.renderNetworkView()
	case stateAddNetwork:
		content = m.renderAddNetworkView()
	case stateError:
		content = m.renderErrorView()
	}
	
	return BaseStyle.Width(m.width).Height(m.height).Render(content)
}

func (m Model) renderLoadingView() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			m.spinner.View(),
			"Discovering all cloud resources across your projects...",
			lipgloss.NewStyle().Foreground(lipgloss.Color(CatppuccinMocha.Overlay0)).Render("This may take a moment for many projects"),
		),
	)
}

func (m Model) renderResourceSelectionView() string {
	title := RenderTitle("🔐 PIAM Admin Network Configurator")
	subtitle := RenderSubtitle(fmt.Sprintf("Found %d resources across your projects", len(m.resources)))
	
	help := RenderHelp([]string{
		"↑/↓ Navigate • Enter Select • / Search",
		"r Refresh • q Quit • ? Help",
	})
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		m.resourceList.View(),
	)
	
	if m.message != "" {
		messageStyle := MessageStyle
		if m.isError {
			messageStyle = ErrorMessageStyle
		}
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			messageStyle.Render(m.message),
		)
		m.message = ""
		m.isError = false
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)
}

func (m Model) renderNetworkView() string {
	var networks []AuthorizedNetwork
	resourceType := m.selectedResource.GetType()
	
	// Get networks based on resource type
	switch r := m.selectedResource.(type) {
	case SQLInstance:
		networks = r.AuthorizedNetworks
	case GKECluster:
		networks = r.MasterAuthorizedNetworks
	}
	
	title := RenderTitle(fmt.Sprintf("%s %s", getResourceIcon(m.selectedResource), m.selectedResource.GetDisplayName()))
	
	var subtitle string
	if resourceType == ResourceTypeSQL {
		subtitle = "SQL Instance Authorized Networks"
	} else {
		subtitle = "GKE Cluster Master Authorized Networks"
	}
	subtitle = RenderSubtitle(subtitle)
	
	// Show console link
	consoleLink := lipgloss.NewStyle().
		Foreground(lipgloss.Color(CatppuccinMocha.Blue)).
		Underline(true).
		Render("🌐 Open in Google Cloud Console (press 'c')")
	
	// Show restrictions if any
	restrictions := ""
	if r := m.selectedResource.GetNetworkRestrictions(); r != "" {
		restrictions = WarningStyle.Render("⚠️  " + r)
	}
	
	// Create table
	table := RenderNetworkTable(networks)
	
	// Help text
	helpItems := []string{
		"↑/↓ Navigate",
		"c Console",
		"Esc Back",
		"r Refresh",
		"q Quit",
	}
	
	if m.selectedResource.CanAddNetwork() {
		helpItems = append([]string{"a Add network"}, helpItems...)
	}
	
	help := RenderHelp(helpItems)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		consoleLink,
	)
	
	if restrictions != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			restrictions,
		)
	}
	
	content = lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		"",
		table,
	)
	
	if m.message != "" {
		messageStyle := MessageStyle
		if m.isError {
			messageStyle = ErrorMessageStyle
		}
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			messageStyle.Render(m.message),
		)
		m.message = ""
		m.isError = false
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)
}

func (m Model) renderAddNetworkView() string {
	title := RenderTitle("➕ Add Authorized Network")
	subtitle := RenderSubtitle(fmt.Sprintf("Adding to: %s", m.selectedResource.GetDisplayName()))
	
	// Name input
	nameLabel := "Network Name:"
	nameField := m.nameInput.View()
	if m.addFormFocus == 0 {
		nameField = ActiveInputStyle.Render(nameField)
	} else {
		nameField = InputStyle.Render(nameField)
	}
	
	// IP input
	ipLabel := "IP Address/CIDR:"
	ipField := m.ipInput.View()
	if m.addFormFocus == 1 {
		ipField = ActiveInputStyle.Render(ipField)
	} else {
		ipField = InputStyle.Render(ipField)
	}
	
	form := lipgloss.JoinVertical(
		lipgloss.Left,
		LabelStyle.Render(nameLabel),
		nameField,
		"",
		LabelStyle.Render(ipLabel),
		ipField,
		"",
		SubtleTextStyle.Render("Example: 192.168.1.100/32 or 10.0.0.0/24"),
	)
	
	formBox := FormBoxStyle.Render(form)
	
	var helpItems []string
	if m.isSubmitting {
		helpItems = []string{
			"Please wait...",
		}
	} else if m.addFormFocus == -1 {
		helpItems = []string{
			"Tab/Enter Focus first field",
			"Esc Cancel • q Quit",
		}
	} else {
		helpItems = []string{
			"Tab Navigate fields • Enter Submit",
			"Esc Cancel • q Quit",
		}
	}
	help := RenderHelp(helpItems)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		"",
		formBox,
	)
	
	if m.message != "" {
		messageStyle := MessageStyle
		if m.isError {
			messageStyle = ErrorMessageStyle
		}
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			messageStyle.Render(m.message),
		)
		m.message = ""
		m.isError = false
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)
}

func (m Model) renderErrorView() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ErrorBoxStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				ErrorTitleStyle.Render("⚠️  Error"),
				"",
				m.message,
				"",
				SubtleTextStyle.Render("Press q to quit or Esc to go back"),
			),
		),
	)
}

func (m Model) renderHelpView() string {
	helpText := `
🔐 PIAM Admin Network Configurator

A beautiful TUI for managing Cloud SQL and GKE authorized networks.

NAVIGATION
  ↑/↓ or j/k     Navigate through lists
  Enter          Select resource or submit form
  Esc            Go back / Cancel
  Tab            Switch between form fields
  /              Search resources
  q or Ctrl+C    Quit

ACTIONS
  a              Add authorized network (when available)
  r              Refresh resource list
  ?              Toggle this help

RESOURCE ICONS
  🗄️             SQL Database Instance
  ☸️             GKE Kubernetes Cluster
  🔒             Resource cannot accept external networks

NETWORK RESTRICTIONS
  • Private SQL instances cannot have authorized networks
  • GKE clusters always support master authorized networks
  • Some resources may require VPN or jumphost access

Press any key to return...`
	
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		HelpBoxStyle.Width(80).Render(helpText),
	)
}

// Commands
func loadResources() tea.Msg {
	ctx := context.Background()
	nm, err := NewNetworkManager(ctx)
	if err != nil {
		return errorMsg{err}
	}
	
	resources, err := nm.ListAllResources()
	if err != nil {
		return errorMsg{err}
	}
	
	return resourcesLoadedMsg{resources}
}

func selectResource(resource CloudResource) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		nm, err := NewNetworkManager(ctx)
		if err != nil {
			return errorMsg{err}
		}
		
		// Get fresh details
		updated, err := nm.GetResourceDetails(resource)
		if err != nil {
			return errorMsg{err}
		}
		
		return resourceSelectedMsg{resource: updated}
	}
}

func submitAddNetwork(resource CloudResource, name, ip string) tea.Cmd {
	return func() tea.Msg {
		// Validate inputs
		if strings.TrimSpace(name) == "" {
			return networkAddedMsg{
				success: false,
				message: "Network name is required",
			}
		}
		
		if strings.TrimSpace(ip) == "" {
			return networkAddedMsg{
				success: false,
				message: "IP address is required",
			}
		}
		
		ctx := context.Background()
		nm, err := NewNetworkManager(ctx)
		if err != nil {
			return errorMsg{err}
		}
		
		err = nm.AddNetworkToResource(resource, name, ip)
		if err != nil {
			return networkAddedMsg{
				success: false,
				message: fmt.Sprintf("Failed to add network: %v", err),
			}
		}
		
		return networkAddedMsg{
			success: true,
			message: fmt.Sprintf("Successfully added network %s", name),
		}
	}
}

// Helper function to render network table
func RenderNetworkTable(networks []AuthorizedNetwork) string {
	if len(networks) == 0 {
		return EmptyStateStyle.Render("No authorized networks configured")
	}
	
	// Create table header
	header := TableHeaderStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			TableCellStyle.Width(30).Render("Name"),
			TableCellStyle.Width(20).Render("IP/CIDR"),
		),
	)
	
	// Create table rows
	rows := make([]string, len(networks))
	for i, network := range networks {
		name := network.Name
		if name == "" {
			name = network.DisplayName
		}
		if name == "" {
			name = "(unnamed)"
		}
		
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			TableCellStyle.Width(30).Render(name),
			TableCellStyle.Width(20).Render(network.Value),
		)
		
		if i%2 == 0 {
			rows[i] = TableRowEvenStyle.Render(row)
		} else {
			rows[i] = TableRowOddStyle.Render(row)
		}
	}
	
	table := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		strings.Join(rows, "\n"),
	)
	
	return TableStyle.Render(table)
}

// openConsoleURL opens the Google Cloud Console for the resource
func openConsoleURL(resource CloudResource) tea.Cmd {
	return func() tea.Msg {
		url := getConsoleURL(resource)
		if url == "" {
			return networkAddedMsg{
				success: false,
				message: "Unable to generate console URL for this resource",
			}
		}
		
		// Try to open the URL using the default system browser
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			return networkAddedMsg{
				success: false,
				message: "Unsupported platform for opening URLs",
			}
		}
		
		err := cmd.Start()
		if err != nil {
			return networkAddedMsg{
				success: false,
				message: fmt.Sprintf("Failed to open console: %v", err),
			}
		}
		
		return networkAddedMsg{
			success: true,
			message: "Opened Google Cloud Console in browser",
		}
	}
}

// tickCmd creates a command that ticks every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
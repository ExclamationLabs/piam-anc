package main

import (
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha color palette
var (
	CatppuccinMocha = struct {
		Rosewater string
		Flamingo  string
		Pink      string
		Mauve     string
		Red       string
		Maroon    string
		Peach     string
		Yellow    string
		Green     string
		Teal      string
		Sky       string
		Sapphire  string
		Blue      string
		Lavender  string
		Text      string
		Subtext1  string
		Subtext0  string
		Overlay2  string
		Overlay1  string
		Overlay0  string
		Surface2  string
		Surface1  string
		Surface0  string
		Base      string
		Mantle    string
		Crust     string
	}{
		Rosewater: "#f5e0dc",
		Flamingo:  "#f2cdcd",
		Pink:      "#f5c2e7",
		Mauve:     "#cba6f7",
		Red:       "#f38ba8",
		Maroon:    "#eba0ac",
		Peach:     "#fab387",
		Yellow:    "#f9e2af",
		Green:     "#a6e3a1",
		Teal:      "#94e2d5",
		Sky:       "#89dceb",
		Sapphire:  "#74c7ec",
		Blue:      "#89b4fa",
		Lavender:  "#b4befe",
		Text:      "#cdd6f4",
		Subtext1:  "#bac2de",
		Subtext0:  "#a6adc8",
		Overlay2:  "#9399b2",
		Overlay1:  "#7f849c",
		Overlay0:  "#6c7086",
		Surface2:  "#585b70",
		Surface1:  "#45475a",
		Surface0:  "#313244",
		Base:      "#1e1e2e",
		Mantle:    "#181825",
		Crust:     "#11111b",
	}
)

// Theme styles
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(CatppuccinMocha.Base)).
			Foreground(lipgloss.Color(CatppuccinMocha.Text))

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Lavender)).
			Background(lipgloss.Color(CatppuccinMocha.Surface0)).
			Padding(0, 1).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Subtext1)).
			Italic(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CatppuccinMocha.Surface2))

	ActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(CatppuccinMocha.Mauve))

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Base)).
			Background(lipgloss.Color(CatppuccinMocha.Blue)).
			Padding(0, 2).
			Margin(0, 1)

	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Base)).
				Background(lipgloss.Color(CatppuccinMocha.Mauve)).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Text)).
			Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Base)).
				Background(lipgloss.Color(CatppuccinMocha.Mauve)).
				Padding(0, 2).
				Bold(true)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Green)).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Red)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Yellow)).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Blue)).
			Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Text)).
			Background(lipgloss.Color(CatppuccinMocha.Surface0)).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CatppuccinMocha.Surface2))

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Text)).
				Background(lipgloss.Color(CatppuccinMocha.Surface0)).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(CatppuccinMocha.Mauve))

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Lavender)).
				Background(lipgloss.Color(CatppuccinMocha.Surface1)).
				Bold(true).
				Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Text)).
			Padding(0, 1)

	TableAltRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Text)).
				Background(lipgloss.Color(CatppuccinMocha.Surface0)).
				Padding(0, 1)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Subtext0)).
			Italic(true)

	KeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Mauve)).
			Bold(true)

	// Loading styles
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Blue))

	// Network type styles
	NetworkNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Green)).
				Bold(true)

	NetworkIPStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Blue))

	NetworkMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Subtext0)).
				Italic(true)

	// Message styles  
	MessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Green)).
			Bold(true)

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Red)).
				Bold(true)

	// Additional missing styles
	EmptyStateStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Subtext0)).
				Italic(true).
				Align(lipgloss.Center)

	TableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CatppuccinMocha.Surface2))

	TableCellStyle = lipgloss.NewStyle().
				Padding(0, 1)

	TableRowEvenStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(CatppuccinMocha.Surface0))

	TableRowOddStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(CatppuccinMocha.Base))

	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CatppuccinMocha.Blue)).
			Padding(1, 2)

	// Form styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(CatppuccinMocha.Lavender)).
			Bold(true)

	ActiveInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(CatppuccinMocha.Mauve)).
				Padding(0, 1)

	FormBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CatppuccinMocha.Surface2)).
			Padding(1, 2)

	SubtleTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Subtext0)).
				Italic(true)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(CatppuccinMocha.Red)).
			Padding(1, 2)

	ErrorTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(CatppuccinMocha.Red)).
				Bold(true)
)

// Helper functions for common styling patterns
func RenderTitle(title string) string {
	return TitleStyle.Render(title)
}

func RenderSubtitle(subtitle string) string {
	return SubtitleStyle.Render(subtitle)
}

func RenderSuccess(text string) string {
	return SuccessStyle.Render("✓ " + text)
}

func RenderError(text string) string {
	return ErrorStyle.Render("✗ " + text)
}

func RenderWarning(text string) string {
	return WarningStyle.Render("⚠ " + text)
}

func RenderInfo(text string) string {
	return InfoStyle.Render("ℹ " + text)
}

func RenderHelp(items []string) string {
	if len(items) == 0 {
		return ""
	}
	
	// Join items with " • " separator
	text := strings.Join(items, " • ")
	return HelpStyle.Render(text)
}

func RenderKey(key string) string {
	return KeyStyle.Render(key)
}

func RenderNetworkName(name string) string {
	return NetworkNameStyle.Render(name)
}

func RenderNetworkIP(ip string) string {
	return NetworkIPStyle.Render(ip)
}

func RenderNetworkMeta(meta string) string {
	return NetworkMetaStyle.Render(meta)
}
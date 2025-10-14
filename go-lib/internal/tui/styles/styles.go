package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	ProtonBlue   = lipgloss.Color("#6D4AFF")
	ProtonPurple = lipgloss.Color("#8A2BE2")
	
	// Status colors
	SuccessGreen = lipgloss.Color("#00C851")
	WarningYellow = lipgloss.Color("#FFB900")
	ErrorRed     = lipgloss.Color("#FF4444")
	
	// Neutral colors
	TextPrimary   = lipgloss.Color("#FFFFFF")
	TextSecondary = lipgloss.Color("#B0B0B0")
	TextMuted     = lipgloss.Color("#808080")
	Background    = lipgloss.Color("#1A1A1A")
	Surface       = lipgloss.Color("#2D2D2D")
	Border        = lipgloss.Color("#404040")
)

// Base styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Background(Background).
		Foreground(TextPrimary)
	
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonBlue).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(Border).
		Padding(0, 1, 1, 1).
		MarginBottom(1)
	
	FooterStyle = lipgloss.NewStyle().
		Foreground(TextMuted).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(Border).
		Padding(1, 1, 0, 1).
		MarginTop(1)
	
	// Content styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(TextPrimary).
		MarginBottom(1)
	
	SubtitleStyle = lipgloss.NewStyle().
		Foreground(TextSecondary).
		MarginBottom(1)
	
	// Interactive elements
	ButtonStyle = lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(ProtonBlue).
		Padding(0, 2).
		MarginRight(1).
		Bold(true)
	
	ButtonActiveStyle = ButtonStyle.Copy().
		Background(ProtonPurple).
		Underline(true)
	
	ButtonDisabledStyle = ButtonStyle.Copy().
		Background(Surface).
		Foreground(TextMuted)
	
	// Input styles
	InputStyle = lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(Surface).
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(Border)
	
	InputFocusedStyle = InputStyle.Copy().
		BorderForeground(ProtonBlue)
	
	// Status styles
	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessGreen).
		Bold(true)
	
	WarningStyle = lipgloss.NewStyle().
		Foreground(WarningYellow).
		Bold(true)
	
	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorRed).
		Bold(true)
	
	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
		Background(Surface).
		Foreground(ProtonBlue)
	
	ProgressTextStyle = lipgloss.NewStyle().
		Foreground(TextSecondary)
)

// Helper functions for dynamic styling
func WithWidth(style lipgloss.Style, width int) lipgloss.Style {
	return style.Copy().Width(width)
}

func WithHeight(style lipgloss.Style, height int) lipgloss.Style {
	return style.Copy().Height(height)
}

func CenterHorizontal(style lipgloss.Style) lipgloss.Style {
	return style.Copy().Align(lipgloss.Center)
}

// Responsive width calculation
func ResponsiveWidth(terminalWidth int, percentage float64) int {
	width := int(float64(terminalWidth) * percentage)
	if width < 40 {
		return 40 // Minimum width
	}
	if width > 120 {
		return 120 // Maximum width
	}
	return width
}
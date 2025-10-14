package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal"
	"github.com/ProtonMail/export-tool/internal/tui/styles"
)

// WelcomeModel represents the welcome screen
type WelcomeModel struct {
	width          int
	height         int
	selectedOption int
	options        []string
	versionChecked bool
	hasNewVersion  bool
}

// NewWelcomeModel creates a new welcome model
func NewWelcomeModel() WelcomeModel {
	return WelcomeModel{
		options: []string{
			"Start Export/Import",
			"Exit",
		},
		selectedOption: 0,
	}
}

// Init implements tea.Model
func (m WelcomeModel) Init() tea.Cmd {
	return tea.Batch(
		checkVersionCmd(),
		tea.EnterAltScreen,
	)
}

// Update implements tea.Model
func (m WelcomeModel) Update(msg tea.Msg) (WelcomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.selectedOption < len(m.options)-1 {
				m.selectedOption++
			}
		case "enter", " ":
			return m.handleSelection()
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case VersionCheckMsg:
		m.versionChecked = true
		m.hasNewVersion = msg.HasNewVersion
	}

	return m, nil
}

// View implements tea.Model
func (m WelcomeModel) View() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Version status
	versionStatus := m.renderVersionStatus()
	sections = append(sections, versionStatus)

	// Menu options
	menu := m.renderMenu()
	sections = append(sections, menu)

	// Footer with help
	footer := m.renderFooter()
	sections = append(sections, footer)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	
	// Center the content
	return lipgloss.Place(
		m.width-4, m.height-4,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

// SetSize updates the model dimensions
func (m *WelcomeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// renderHeader creates the application header
func (m WelcomeModel) renderHeader() string {
	title := styles.TitleStyle.Render("Proton Mail Export Tool")
	version := styles.SubtitleStyle.Render(fmt.Sprintf("Version %s", internal.ETVersionString))
	copyright := styles.SubtitleStyle.Render("© Proton AG, Switzerland")
	license := styles.SubtitleStyle.Render("Licensed under GNU General Public License v3")
	
	headerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		version,
		"",
		copyright,
		license,
	)
	
	return styles.HeaderStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(headerContent)
}

// renderVersionStatus shows version check results
func (m WelcomeModel) renderVersionStatus() string {
	if !m.versionChecked {
		return styles.ProgressTextStyle.Render("Checking for updates...")
	}
	
	if m.hasNewVersion {
		message := "🔄 A new version is available!"
		url := "Visit: https://proton.me/support/proton-mail-export-tool"
		return lipgloss.JoinVertical(
			lipgloss.Left,
			styles.WarningStyle.Render(message),
			styles.TextSecondary.Render(url),
		)
	}
	
	return styles.SuccessStyle.Render("✓ Your version is up to date")
}

// renderMenu creates the main menu
func (m WelcomeModel) renderMenu() string {
	var menuItems []string
	
	for i, option := range m.options {
		var style lipgloss.Style
		prefix := "  "
		
		if i == m.selectedOption {
			style = styles.ButtonActiveStyle
			prefix = "▶ "
		} else {
			style = styles.ButtonStyle
		}
		
		menuItems = append(menuItems, prefix+style.Render(option))
	}
	
	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)
	
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(1, 2).
		MarginTop(2).
		MarginBottom(2).
		Render(menu)
}

// renderFooter creates the help footer
func (m WelcomeModel) renderFooter() string {
	help := []string{
		"↑/↓ or j/k: Navigate",
		"Enter/Space: Select",
		"q/Ctrl+C: Quit",
	}
	
	helpText := strings.Join(help, " • ")
	
	return styles.FooterStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(helpText)
}

// handleSelection processes menu selection
func (m WelcomeModel) handleSelection() (WelcomeModel, tea.Cmd) {
	switch m.selectedOption {
	case 0: // Start Export/Import
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: ScreenAuth,
				Data:   make(map[string]interface{}),
			}
		}
	case 1: // Exit
		return m, tea.Quit
	}
	
	return m, nil
}

// Messages

// VersionCheckMsg is sent when version check completes
type VersionCheckMsg struct {
	HasNewVersion bool
}

// checkVersionCmd performs version checking
func checkVersionCmd() tea.Cmd {
	return func() tea.Msg {
		hasNewVersion := internal.HasNewVersion()
		return VersionCheckMsg{HasNewVersion: hasNewVersion}
	}
}
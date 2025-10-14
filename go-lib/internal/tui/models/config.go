package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal/tui/styles"
)

// ConfigModel handles configuration for the selected operation
type ConfigModel struct {
	width     int
	height    int
	operation string
	
	// Directory selection
	currentPath   string
	inputPath     string
	cursorPos     int
	focused       bool
	
	// UI state
	errorMsg      string
	pathValid     bool
	
	// Configuration options
	selectedOption int
	options        []string
}

// NewConfigModel creates a new configuration model
func NewConfigModel() ConfigModel {
	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, "ProtonMailExport")
	
	return ConfigModel{
		currentPath:    defaultPath,
		inputPath:      defaultPath,
		focused:        true,
		pathValid:      true,
		selectedOption: 0,
		options:        []string{"Directory Path", "Start Operation"},
	}
}

// Init implements tea.Model
func (m ConfigModel) Init() tea.Cmd {
	return tea.Batch(
		tea.SetCursorMode(tea.CursorBlink),
		validatePathCmd(m.inputPath, m.operation),
	)
}

// Update implements tea.Model
func (m ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: ScreenOperation, Data: make(map[string]interface{})}
			}
		case "tab", "up", "down":
			// Toggle between path input and start button
			if m.selectedOption == 0 {
				m.selectedOption = 1
				m.focused = false
			} else {
				m.selectedOption = 0
				m.focused = true
			}
		case "enter":
			return m.handleEnter()
		default:
			if m.selectedOption == 0 && m.focused {
				return m.handlePathInput(msg)
			}
		}
		
	case PathValidationMsg:
		m.pathValid = msg.Valid
		m.errorMsg = msg.Error
		
	case PathCreatedMsg:
		m.pathValid = true
		m.errorMsg = ""
		m.currentPath = m.inputPath
	}

	return m, nil
}

// View implements tea.Model
func (m ConfigModel) View() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Configuration content
	configContent := m.renderConfig()
	sections = append(sections, configContent)

	// Error message
	if m.errorMsg != "" {
		errorContent := styles.ErrorStyle.Render(fmt.Sprintf("❌ %s", m.errorMsg))
		sections = append(sections, errorContent)
	}

	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.Place(
		m.width-4, m.height-4,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

// SetSize updates the model dimensions
func (m *ConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetOperation sets the operation type
func (m *ConfigModel) SetOperation(operation string) {
	m.operation = operation
	// Update cursor position to end of current path
	m.cursorPos = len(m.inputPath)
}

// renderHeader creates the configuration header
func (m ConfigModel) renderHeader() string {
	var title string
	var subtitle string
	
	switch m.operation {
	case "backup":
		title = "Backup Configuration"
		subtitle = "Configure your email export settings"
	case "restore":
		title = "Restore Configuration"
		subtitle = "Configure your email import settings"
	default:
		title = "Configuration"
		subtitle = "Configure operation settings"
	}

	headerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render(title),
		styles.SubtitleStyle.Render(subtitle),
	)

	return styles.HeaderStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(headerContent)
}

// renderConfig creates the configuration form
func (m ConfigModel) renderConfig() string {
	var sections []string

	// Directory path section
	pathSection := m.renderPathSection()
	sections = append(sections, pathSection)

	// Start button section
	buttonSection := m.renderButtonSection()
	sections = append(sections, buttonSection)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderPathSection creates the directory path input section
func (m ConfigModel) renderPathSection() string {
	var label string
	var description string
	
	switch m.operation {
	case "backup":
		label = "Export Directory:"
		description = "Choose where to save your exported emails"
	case "restore":
		label = "Import Directory:"
		description = "Choose the directory containing exported emails"
	default:
		label = "Directory:"
		description = "Choose the directory path"
	}

	// Render input field
	displayValue := m.inputPath
	if m.focused && m.selectedOption == 0 {
		if m.cursorPos >= len(displayValue) {
			displayValue += "│"
		} else {
			displayValue = displayValue[:m.cursorPos] + "│" + displayValue[m.cursorPos:]
		}
	}

	inputStyle := styles.InputStyle
	if m.focused && m.selectedOption == 0 {
		inputStyle = styles.InputFocusedStyle
	}
	if !m.pathValid {
		inputStyle = inputStyle.BorderForeground(styles.ErrorRed)
	}

	input := inputStyle.
		Width(styles.ResponsiveWidth(m.width, 0.7)).
		Render(displayValue)

	// Status indicator
	var statusIcon string
	if m.pathValid {
		statusIcon = styles.SuccessStyle.Render("✓")
	} else {
		statusIcon = styles.ErrorStyle.Render("✗")
	}

	pathContent := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.SubtitleStyle.Render(label),
		lipgloss.JoinHorizontal(lipgloss.Left, input, " ", statusIcon),
		"",
		styles.TextMuted.Render(description),
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(1, 2).
		MarginBottom(2).
		Render(pathContent)
}

// renderButtonSection creates the start operation button
func (m ConfigModel) renderButtonSection() string {
	var buttonText string
	switch m.operation {
	case "backup":
		buttonText = "Start Backup"
	case "restore":
		buttonText = "Start Restore"
	default:
		buttonText = "Start Operation"
	}

	buttonStyle := styles.ButtonStyle
	if m.selectedOption == 1 {
		buttonStyle = styles.ButtonActiveStyle
	}
	if !m.pathValid {
		buttonStyle = styles.ButtonDisabledStyle
	}

	button := buttonStyle.
		Width(styles.ResponsiveWidth(m.width, 0.3)).
		Align(lipgloss.Center).
		Render(buttonText)

	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(button)
}

// renderFooter creates the help footer
func (m ConfigModel) renderFooter() string {
	help := []string{
		"Tab: Switch fields",
		"Enter: Confirm",
		"Esc: Back",
		"Ctrl+C: Quit",
	}

	helpText := strings.Join(help, " • ")

	return styles.FooterStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(helpText)
}

// handlePathInput handles keyboard input for path editing
func (m ConfigModel) handlePathInput(msg tea.KeyMsg) (ConfigModel, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		if m.cursorPos > 0 {
			m.inputPath = m.inputPath[:m.cursorPos-1] + m.inputPath[m.cursorPos:]
			m.cursorPos--
			return m, validatePathCmd(m.inputPath, m.operation)
		}
	case "left":
		if m.cursorPos > 0 {
			m.cursorPos--
		}
	case "right":
		if m.cursorPos < len(m.inputPath) {
			m.cursorPos++
		}
	case "home":
		m.cursorPos = 0
	case "end":
		m.cursorPos = len(m.inputPath)
	default:
		// Handle character input
		if len(msg.String()) == 1 {
			char := msg.String()
			m.inputPath = m.inputPath[:m.cursorPos] + char + m.inputPath[m.cursorPos:]
			m.cursorPos++
			return m, validatePathCmd(m.inputPath, m.operation)
		}
	}

	return m, nil
}

// handleEnter processes enter key based on current selection
func (m ConfigModel) handleEnter() (ConfigModel, tea.Cmd) {
	if m.selectedOption == 0 {
		// Path input - try to create directory if it doesn't exist
		if !m.pathValid {
			return m, createDirectoryCmd(m.inputPath)
		}
	} else if m.selectedOption == 1 {
		// Start button - proceed to operation
		if m.pathValid {
			return m, func() tea.Msg {
				return NavigateMsg{
					Screen: ScreenProgress,
					Data: map[string]interface{}{
						"path": m.inputPath,
					},
				}
			}
		}
	}

	return m, nil
}

// Messages for configuration

// PathValidationMsg is sent when path validation completes
type PathValidationMsg struct {
	Valid bool
	Error string
}

// PathCreatedMsg is sent when directory creation succeeds
type PathCreatedMsg struct {
	Path string
}

// Commands for configuration operations

func validatePathCmd(path, operation string) tea.Cmd {
	return func() tea.Msg {
		if path == "" {
			return PathValidationMsg{Valid: false, Error: "Path cannot be empty"}
		}

		// Clean the path
		cleanPath := filepath.Clean(path)

		if operation == "backup" {
			// For backup, check if we can create the directory
			if err := os.MkdirAll(cleanPath, 0755); err != nil {
				return PathValidationMsg{Valid: false, Error: fmt.Sprintf("Cannot create directory: %v", err)}
			}
			return PathValidationMsg{Valid: true, Error: ""}
		} else if operation == "restore" {
			// For restore, check if directory exists and contains files
			if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
				return PathValidationMsg{Valid: false, Error: "Directory does not exist"}
			}
			// Could add more validation here to check for .eml files
			return PathValidationMsg{Valid: true, Error: ""}
		}

		return PathValidationMsg{Valid: true, Error: ""}
	}
}

func createDirectoryCmd(path string) tea.Cmd {
	return func() tea.Msg {
		cleanPath := filepath.Clean(path)
		if err := os.MkdirAll(cleanPath, 0755); err != nil {
			return PathValidationMsg{Valid: false, Error: fmt.Sprintf("Failed to create directory: %v", err)}
		}
		return PathCreatedMsg{Path: cleanPath}
	}
}
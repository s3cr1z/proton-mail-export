package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal/tui/styles"
)

// CompleteModel handles the completion screen
type CompleteModel struct {
	width          int
	height         int
	success        bool
	errorMsg       string
	selectedOption int
	options        []string
}

// NewCompleteModel creates a new completion model
func NewCompleteModel() CompleteModel {
	return CompleteModel{
		options: []string{
			"Start New Operation",
			"Exit Application",
		},
		selectedOption: 0,
	}
}

// Init implements tea.Model
func (m CompleteModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m CompleteModel) Update(msg tea.Msg) (CompleteModel, tea.Cmd) {
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
	}

	return m, nil
}

// View implements tea.Model
func (m CompleteModel) View() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Status message
	statusContent := m.renderStatus()
	sections = append(sections, statusContent)

	// Options
	optionsContent := m.renderOptions()
	sections = append(sections, optionsContent)

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
func (m *CompleteModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetResult sets the operation result
func (m *CompleteModel) SetResult(success bool, errorMsg string) {
	m.success = success
	m.errorMsg = errorMsg
}

// renderHeader creates the completion header
func (m CompleteModel) renderHeader() string {
	var title string
	if m.success {
		title = "Operation Complete"
	} else {
		title = "Operation Failed"
	}

	subtitle := "What would you like to do next?"

	headerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render(title),
		styles.SubtitleStyle.Render(subtitle),
	)

	return styles.HeaderStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(headerContent)
}

// renderStatus creates the status message
func (m CompleteModel) renderStatus() string {
	var statusContent string
	
	if m.success {
		statusContent = lipgloss.JoinVertical(
			lipgloss.Center,
			styles.SuccessStyle.Render("✓ Success!"),
			"",
			styles.TextSecondary.Render("Your operation completed successfully."),
		)
	} else {
		errorText := "✗ Operation Failed"
		if m.errorMsg != "" {
			errorText = "✗ " + m.errorMsg
		}
		
		statusContent = lipgloss.JoinVertical(
			lipgloss.Center,
			styles.ErrorStyle.Render(errorText),
			"",
			styles.TextSecondary.Render("Please check the logs for more details."),
		)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(func() lipgloss.Color {
			if m.success {
				return styles.SuccessGreen
			}
			return styles.ErrorRed
		}()).
		Padding(2, 4).
		MarginTop(2).
		MarginBottom(2).
		Render(statusContent)
}

// renderOptions creates the action options
func (m CompleteModel) renderOptions() string {
	var optionItems []string

	for i, option := range m.options {
		var style lipgloss.Style
		prefix := "  "

		if i == m.selectedOption {
			style = styles.ButtonActiveStyle
			prefix = "▶ "
		} else {
			style = styles.ButtonStyle
		}

		optionItems = append(optionItems, prefix+style.Render(option))
	}

	options := lipgloss.JoinVertical(lipgloss.Left, optionItems...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(1, 2).
		MarginBottom(2).
		Render(options)
}

// renderFooter creates the help footer
func (m CompleteModel) renderFooter() string {
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

// handleSelection processes option selection
func (m CompleteModel) handleSelection() (CompleteModel, tea.Cmd) {
	switch m.selectedOption {
	case 0: // Start New Operation
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: ScreenWelcome,
				Data:   make(map[string]interface{}),
			}
		}
	case 1: // Exit Application
		return m, tea.Quit
	}

	return m, nil
}
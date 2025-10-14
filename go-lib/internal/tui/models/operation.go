package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal/tui/styles"
)

// OperationModel handles operation selection (backup/restore)
type OperationModel struct {
	width          int
	height         int
	selectedOption int
	operations     []Operation
}

// Operation represents an available operation
type Operation struct {
	Name        string
	Description string
	Key         string
}

// NewOperationModel creates a new operation selection model
func NewOperationModel() OperationModel {
	return OperationModel{
		operations: []Operation{
			{
				Name:        "Backup Emails",
				Description: "Export your emails to local storage as EML files",
				Key:         "backup",
			},
			{
				Name:        "Restore Emails",
				Description: "Import previously exported emails back to your account",
				Key:         "restore",
			},
		},
		selectedOption: 0,
	}
}

// Init implements tea.Model
func (m OperationModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m OperationModel) Update(msg tea.Msg) (OperationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.selectedOption < len(m.operations)-1 {
				m.selectedOption++
			}
		case "enter", " ":
			return m.handleSelection()
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: ScreenAuth, Data: make(map[string]interface{})}
			}
		}
	}

	return m, nil
}

// View implements tea.Model
func (m OperationModel) View() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Operation options
	operationsContent := m.renderOperations()
	sections = append(sections, operationsContent)

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
func (m *OperationModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// renderHeader creates the operation selection header
func (m OperationModel) renderHeader() string {
	title := styles.TitleStyle.Render("Select Operation")
	subtitle := styles.SubtitleStyle.Render("Choose what you would like to do")

	headerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
	)

	return styles.HeaderStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(headerContent)
}

// renderOperations creates the operation selection menu
func (m OperationModel) renderOperations() string {
	var operationItems []string

	for i, operation := range m.operations {
		var itemStyle lipgloss.Style
		var prefix string

		if i == m.selectedOption {
			itemStyle = styles.ButtonActiveStyle.Copy().
				Width(styles.ResponsiveWidth(m.width, 0.6)).
				Padding(1, 2).
				MarginBottom(1)
			prefix = "▶ "
		} else {
			itemStyle = styles.ButtonStyle.Copy().
				Background(styles.Surface).
				Width(styles.ResponsiveWidth(m.width, 0.6)).
				Padding(1, 2).
				MarginBottom(1)
			prefix = "  "
		}

		// Create operation card content
		nameStyle := lipgloss.NewStyle().Bold(true)
		if i == m.selectedOption {
			nameStyle = nameStyle.Foreground(styles.TextPrimary)
		} else {
			nameStyle = nameStyle.Foreground(styles.TextSecondary)
		}

		descStyle := lipgloss.NewStyle()
		if i == m.selectedOption {
			descStyle = descStyle.Foreground(styles.TextSecondary)
		} else {
			descStyle = descStyle.Foreground(styles.TextMuted)
		}

		cardContent := lipgloss.JoinVertical(
			lipgloss.Left,
			nameStyle.Render(operation.Name),
			descStyle.Render(operation.Description),
		)

		operationCard := itemStyle.Render(prefix + cardContent)
		operationItems = append(operationItems, operationCard)
	}

	return lipgloss.JoinVertical(lipgloss.Left, operationItems...)
}

// renderFooter creates the help footer
func (m OperationModel) renderFooter() string {
	help := []string{
		"↑/↓ or j/k: Navigate",
		"Enter/Space: Select",
		"Esc: Back",
		"Ctrl+C: Quit",
	}

	helpText := strings.Join(help, " • ")

	return styles.FooterStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(helpText)
}

// handleSelection processes operation selection
func (m OperationModel) handleSelection() (OperationModel, tea.Cmd) {
	selectedOperation := m.operations[m.selectedOption]

	return m, func() tea.Msg {
		return NavigateMsg{
			Screen: ScreenConfig,
			Data: map[string]interface{}{
				"operation": selectedOperation.Key,
			},
		}
	}
}
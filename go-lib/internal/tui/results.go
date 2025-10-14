package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResultsModel struct {
	app    *App
	width  int
	height int
	
	// Results data
	success       bool
	totalEmails   int
	processedEmails int
	failedEmails  int
	outputPath    string
	exportFormat  string
	errorMessage  string
	duration      string
}

func NewResultsModel(app *App) *ResultsModel {
	return &ResultsModel{
		app: app,
		// These would be populated from the actual export results
		success:         true,
		totalEmails:     100,
		processedEmails: 98,
		failedEmails:    2,
		outputPath:      app.exportConfig.OutputPath,
		exportFormat:    app.exportConfig.Format,
		duration:        "2m 34s",
	}
}

func (m *ResultsModel) Init() tea.Cmd {
	return nil
}

func (m *ResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return m, m.app.SwitchToScreen(MainMenuScreen)
		case "o":
			// Open output folder (would need platform-specific implementation)
			return m, nil
		case "r":
			// Start new export
			return m, m.app.SwitchToScreen(ExportScreen)
		case "q", "esc":
			return m, m.app.SwitchToScreen(MainMenuScreen)
		}
	}
	
	return m, nil
}

func (m *ResultsModel) View() string {
	var content strings.Builder
	
	// Title with status
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	var title string
	var titleColor lipgloss.Color
	
	if m.success {
		title = "✅ Export Completed Successfully!"
		titleColor = ProtonGreen
	} else {
		title = "❌ Export Failed"
		titleColor = ProtonRed
	}
	
	titleStyle = titleStyle.Foreground(titleColor)
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")
	
	// Summary section
	summarySection := m.renderSummary()
	content.WriteString(summarySection)
	content.WriteString("\n\n")
	
	// Statistics section
	statsSection := m.renderStatistics()
	content.WriteString(statsSection)
	content.WriteString("\n\n")
	
	// Actions section
	actionsSection := m.renderActions()
	content.WriteString(actionsSection)
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Padding(2).
		Width(m.width).
		Height(m.height - 4) // Account for header/footer
	
	return containerStyle.Render(content.String())
}

func (m *ResultsModel) renderSummary() string {
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonBlue).
		Padding(1).
		Margin(0, 2)
	
	var summaryContent strings.Builder
	summaryContent.WriteString("📊 Export Summary\n\n")
	
	summaryContent.WriteString(fmt.Sprintf("• Operation: %s\n", strings.Title(m.app.exportConfig.Operation)))
	summaryContent.WriteString(fmt.Sprintf("• Format: %s\n", strings.ToUpper(m.exportFormat)))
	summaryContent.WriteString(fmt.Sprintf("• Output: %s\n", m.outputPath))
	summaryContent.WriteString(fmt.Sprintf("• Duration: %s\n", m.duration))
	
	if m.app.exportConfig.Encryption {
		summaryContent.WriteString("• Encryption: Enabled\n")
	}
	
	return summaryStyle.Render(summaryContent.String())
}

func (m *ResultsModel) renderStatistics() string {
	var statsColor lipgloss.Color
	if m.success {
		statsColor = ProtonGreen
	} else {
		statsColor = ProtonRed
	}
	
	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statsColor).
		Padding(1).
		Margin(0, 2)
	
	var statsContent strings.Builder
	statsContent.WriteString("📈 Processing Statistics\n\n")
	
	statsContent.WriteString(fmt.Sprintf("• Total emails: %d\n", m.totalEmails))
	statsContent.WriteString(fmt.Sprintf("• Successfully processed: %d\n", m.processedEmails))
	
	if m.failedEmails > 0 {
		statsContent.WriteString(fmt.Sprintf("• Failed: %d\n", m.failedEmails))
	}
	
	// Success rate
	successRate := 0.0
	if m.totalEmails > 0 {
		successRate = (float64(m.processedEmails) / float64(m.totalEmails)) * 100
	}
	statsContent.WriteString(fmt.Sprintf("• Success rate: %.1f%%\n", successRate))
	
	// Error details if any
	if m.errorMessage != "" {
		statsContent.WriteString(fmt.Sprintf("\n❌ Error: %s\n", m.errorMessage))
	}
	
	return statsStyle.Render(statsContent.String())
}

func (m *ResultsModel) renderActions() string {
	actionsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonPurple).
		Padding(1).
		Margin(0, 2)
	
	var actionsContent strings.Builder
	actionsContent.WriteString("🎯 Available Actions\n\n")
	
	actionsContent.WriteString("• Enter/Space: Return to Main Menu\n")
	actionsContent.WriteString("• R: Start New Export\n")
	actionsContent.WriteString("• O: Open Output Folder\n")
	actionsContent.WriteString("• Q/Esc: Back to Main Menu\n")
	
	return actionsStyle.Render(actionsContent.String())
}
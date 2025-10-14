package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Completion Model
type CompletionModel struct {
	success     bool
	message     string
	operation   string
	path        string
	metrics     ProgressMetrics
	width       int
	height      int
	showDetails bool
}

func NewCompletionModel() CompletionModel {
	return CompletionModel{}
}

func (m CompletionModel) Init() tea.Cmd {
	return nil
}

func (m CompletionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter", " ":
			return m, tea.Quit
		case "d":
			m.showDetails = !m.showDetails
		case "r":
			// Restart - go back to welcome screen
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: WelcomeScreen}
			}
		}

	case CompletionMsg:
		m.success = msg.Success
		m.message = msg.Message
	}

	return m, nil
}

func (m CompletionModel) View() string {
	var content strings.Builder

	// Title with appropriate styling
	if m.success {
		content.WriteString(successStyle.Render("✅ Operation Completed Successfully!"))
	} else {
		content.WriteString(errorStyle.Render("❌ Operation Failed"))
	}
	content.WriteString("\n\n")

	// Message
	if m.message != "" {
		content.WriteString(m.message)
		content.WriteString("\n\n")
	}

	// Operation summary
	if m.operation != "" {
		content.WriteString(fmt.Sprintf("Operation: %s\n", strings.Title(m.operation)))
	}
	if m.path != "" {
		content.WriteString(fmt.Sprintf("Path: %s\n", m.path))
	}

	// Statistics summary
	if m.metrics.ProcessedItems > 0 {
		content.WriteString("\n")
		content.WriteString(m.renderSummary())
	}

	// Detailed statistics (if requested)
	if m.showDetails && m.metrics.ProcessedItems > 0 {
		content.WriteString("\n\n")
		content.WriteString(subtitleStyle.Render("Detailed Statistics:"))
		content.WriteString("\n")
		content.WriteString(m.renderDetailedStats())
	}

	// Instructions
	content.WriteString("\n\n")
	instructions := []string{
		"Press ENTER or SPACE to exit",
		"Press R to start another operation",
	}
	if m.metrics.ProcessedItems > 0 {
		if m.showDetails {
			instructions = append(instructions, "Press D to hide details")
		} else {
			instructions = append(instructions, "Press D to show details")
		}
	}
	instructions = append(instructions, "Press Q or Ctrl+C to quit")

	content.WriteString(blurredStyle.Render(strings.Join(instructions, " • ")))

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2).
		Align(lipgloss.Center).
		Render(content.String())
}

func (m CompletionModel) renderSummary() string {
	var stats []string

	stats = append(stats, fmt.Sprintf("Total items: %d", m.metrics.ProcessedItems))
	
	if m.metrics.FailedItems > 0 {
		stats = append(stats, fmt.Sprintf("Failed items: %d", m.metrics.FailedItems))
	}

	if !m.metrics.StartTime.IsZero() {
		elapsed := m.metrics.LastUpdateTime.Sub(m.metrics.StartTime)
		if elapsed == 0 {
			elapsed = time.Since(m.metrics.StartTime)
		}
		stats = append(stats, fmt.Sprintf("Total time: %s", formatDuration(elapsed)))
	}

	if m.metrics.ProcessedBytes > 0 {
		stats = append(stats, fmt.Sprintf("Data processed: %s", formatBytes(m.metrics.ProcessedBytes)))
	}

	return strings.Join(stats, " • ")
}

func (m CompletionModel) renderDetailedStats() string {
	var content strings.Builder

	// Time statistics
	if !m.metrics.StartTime.IsZero() {
		content.WriteString("⏱️  Time Statistics:\n")
		
		elapsed := m.metrics.LastUpdateTime.Sub(m.metrics.StartTime)
		if elapsed == 0 {
			elapsed = time.Since(m.metrics.StartTime)
		}
		
		content.WriteString(fmt.Sprintf("   Start time: %s\n", m.metrics.StartTime.Format("15:04:05")))
		content.WriteString(fmt.Sprintf("   End time: %s\n", m.metrics.LastUpdateTime.Format("15:04:05")))
		content.WriteString(fmt.Sprintf("   Total duration: %s\n", formatDuration(elapsed)))
		
		if m.metrics.ProcessedItems > 0 && elapsed > 0 {
			rate := float64(m.metrics.ProcessedItems) / elapsed.Seconds()
			content.WriteString(fmt.Sprintf("   Average rate: %.1f items/second\n", rate))
		}
		content.WriteString("\n")
	}

	// Item statistics
	content.WriteString("📊 Item Statistics:\n")
	content.WriteString(fmt.Sprintf("   Total processed: %d\n", m.metrics.ProcessedItems))
	
	if m.metrics.TotalItems > 0 {
		successRate := float64(m.metrics.ProcessedItems-m.metrics.FailedItems) / float64(m.metrics.ProcessedItems) * 100
		content.WriteString(fmt.Sprintf("   Success rate: %.1f%%\n", successRate))
	}
	
	if m.metrics.FailedItems > 0 {
		content.WriteString(fmt.Sprintf("   Failed items: %d\n", m.metrics.FailedItems))
	}
	content.WriteString("\n")

	// Data statistics
	if m.metrics.ProcessedBytes > 0 {
		content.WriteString("💾 Data Statistics:\n")
		content.WriteString(fmt.Sprintf("   Total data: %s\n", formatBytes(m.metrics.ProcessedBytes)))
		
		if !m.metrics.StartTime.IsZero() {
			elapsed := m.metrics.LastUpdateTime.Sub(m.metrics.StartTime)
			if elapsed == 0 {
				elapsed = time.Since(m.metrics.StartTime)
			}
			if elapsed > 0 {
				throughput := float64(m.metrics.ProcessedBytes) / elapsed.Seconds()
				content.WriteString(fmt.Sprintf("   Average throughput: %s/second\n", formatBytes(uint64(throughput))))
			}
		}
		
		if m.metrics.TotalBytes > 0 {
			percentage := float64(m.metrics.ProcessedBytes) / float64(m.metrics.TotalBytes) * 100
			content.WriteString(fmt.Sprintf("   Completion: %.1f%% of expected data\n", percentage))
		}
	}

	return content.String()
}

// Error Model
type ErrorModel struct {
	error       error
	operation   string
	path        string
	width       int
	height      int
	canRetry    bool
	showDetails bool
}

func NewErrorModel() ErrorModel {
	return ErrorModel{
		canRetry: true,
	}
}

func (m ErrorModel) Init() tea.Cmd {
	return nil
}

func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.canRetry {
				// Go back to operation selection
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: OperationScreen}
				}
			}
			return m, tea.Quit
		case "r":
			if m.canRetry {
				// Restart from welcome
				return m, func() tea.Msg {
					return ScreenChangeMsg{Screen: WelcomeScreen}
				}
			}
		case "d":
			m.showDetails = !m.showDetails
		}

	case ErrorMsg:
		m.error = msg.Error
	}

	return m, nil
}

func (m ErrorModel) View() string {
	var content strings.Builder

	// Title
	content.WriteString(errorStyle.Render("❌ Operation Failed"))
	content.WriteString("\n\n")

	// Error message
	if m.error != nil {
		content.WriteString(fmt.Sprintf("Error: %s\n\n", m.error.Error()))
	}

	// Context information
	if m.operation != "" {
		content.WriteString(fmt.Sprintf("Operation: %s\n", strings.Title(m.operation)))
	}
	if m.path != "" {
		content.WriteString(fmt.Sprintf("Path: %s\n", m.path))
	}

	// Detailed error information (if requested)
	if m.showDetails && m.error != nil {
		content.WriteString("\n")
		content.WriteString(subtitleStyle.Render("Error Details:"))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Type: %T\n", m.error))
		content.WriteString(fmt.Sprintf("Message: %s\n", m.error.Error()))
	}

	// Instructions
	content.WriteString("\n\n")
	instructions := []string{}
	
	if m.canRetry {
		instructions = append(instructions, "Press ESC to try again")
		instructions = append(instructions, "Press R to restart")
	}
	
	if m.error != nil {
		if m.showDetails {
			instructions = append(instructions, "Press D to hide details")
		} else {
			instructions = append(instructions, "Press D to show details")
		}
	}
	
	instructions = append(instructions, "Press Q or Ctrl+C to quit")

	content.WriteString(blurredStyle.Render(strings.Join(instructions, " • ")))

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2).
		Align(lipgloss.Center).
		Render(content.String())
}
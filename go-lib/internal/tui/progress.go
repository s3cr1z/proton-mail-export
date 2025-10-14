package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProgressModel struct {
	app              *App
	width            int
	height           int
	
	// Progress tracking
	progress         progress.Model
	spinner          spinner.Model
	
	// State
	isRunning        bool
	isPaused         bool
	isCompleted      bool
	isCancelled      bool
	hasError         bool
	errorMessage     string
	
	// Metrics
	totalEmails      int
	processedEmails  int
	failedEmails     int
	currentEmail     string
	currentStage     string
	startTime        time.Time
	
	// Performance metrics
	emailsPerSecond  float64
	estimatedTimeRemaining time.Duration
	
	// Export context
	ctx              context.Context
	cancelFunc       context.CancelFunc
}

type ProgressUpdateMsg struct {
	TotalEmails     int
	ProcessedEmails int
	FailedEmails    int
	CurrentEmail    string
	CurrentStage    string
}

type ProgressCompleteMsg struct {
	Success bool
	Error   string
}

type ProgressTickMsg time.Time

func NewProgressModel(app *App) *ProgressModel {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 60
	
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ProtonPurple)
	
	return &ProgressModel{
		app:      app,
		progress: p,
		spinner:  s,
	}
}

func (m *ProgressModel) Init() tea.Cmd {
	m.startTime = time.Now()
	m.isRunning = true
	
	// Create cancellable context
	m.ctx, m.cancelFunc = context.WithCancel(context.Background())
	
	return tea.Batch(
		m.spinner.Tick,
		m.tickCmd(),
		m.startExportProcess(),
	)
}

func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.isRunning && !m.isCancelled {
				m.isCancelled = true
				if m.cancelFunc != nil {
					m.cancelFunc()
				}
				return m, nil
			}
			return m, tea.Quit
			
		case "p", " ":
			if m.isRunning && !m.isCompleted {
				m.isPaused = !m.isPaused
				// In a real implementation, this would pause/resume the export
			}
			
		case "enter":
			if m.isCompleted || m.hasError {
				return m, m.app.SwitchToScreen(ResultsScreen)
			}
		}
		
	case ProgressUpdateMsg:
		m.totalEmails = msg.TotalEmails
		m.processedEmails = msg.ProcessedEmails
		m.failedEmails = msg.FailedEmails
		m.currentEmail = msg.CurrentEmail
		m.currentStage = msg.CurrentStage
		
		// Update progress bar
		if m.totalEmails > 0 {
			percent := float64(m.processedEmails) / float64(m.totalEmails)
			cmds = append(cmds, m.progress.SetPercent(percent))
		}
		
		// Calculate performance metrics
		m.calculateMetrics()
		
	case ProgressCompleteMsg:
		m.isRunning = false
		m.isCompleted = true
		if !msg.Success {
			m.hasError = true
			m.errorMessage = msg.Error
		}
		
	case ProgressTickMsg:
		if m.isRunning && !m.isPaused {
			cmds = append(cmds, m.tickCmd())
		}
		
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m *ProgressModel) View() string {
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	var title string
	if m.app.exportConfig.Operation == "restore" {
		title = "📥 Importing Emails"
	} else {
		title = "📤 Exporting Emails"
	}
	
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")
	
	// Status section
	statusSection := m.renderStatus()
	content.WriteString(statusSection)
	content.WriteString("\n\n")
	
	// Progress section
	progressSection := m.renderProgress()
	content.WriteString(progressSection)
	content.WriteString("\n\n")
	
	// Statistics section
	statsSection := m.renderStatistics()
	content.WriteString(statsSection)
	content.WriteString("\n\n")
	
	// Current activity section
	activitySection := m.renderCurrentActivity()
	content.WriteString(activitySection)
	content.WriteString("\n\n")
	
	// Controls section
	controlsSection := m.renderControls()
	content.WriteString(controlsSection)
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Padding(2).
		Width(m.width).
		Height(m.height - 4) // Account for header/footer
	
	return containerStyle.Render(content.String())
}

func (m *ProgressModel) renderStatus() string {
	var statusText string
	var statusColor lipgloss.Color
	
	if m.isCancelled {
		statusText = "❌ Cancelled"
		statusColor = ProtonRed
	} else if m.hasError {
		statusText = "❌ Error: " + m.errorMessage
		statusColor = ProtonRed
	} else if m.isCompleted {
		statusText = "✅ Completed Successfully"
		statusColor = ProtonGreen
	} else if m.isPaused {
		statusText = "⏸️ Paused"
		statusColor = ProtonBlue
	} else if m.isRunning {
		statusText = m.spinner.View() + " Processing..."
		statusColor = ProtonPurple
	} else {
		statusText = "⏳ Initializing..."
		statusColor = ProtonGray
	}
	
	statusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(statusColor).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Padding(1)
	
	return statusStyle.Render(statusText)
}

func (m *ProgressModel) renderProgress() string {
	var progressContent strings.Builder
	
	// Progress bar
	progressContent.WriteString(m.progress.View())
	progressContent.WriteString("\n\n")
	
	// Progress percentage and counts
	var percent float64
	if m.totalEmails > 0 {
		percent = (float64(m.processedEmails) / float64(m.totalEmails)) * 100
	}
	
	progressText := fmt.Sprintf("%.1f%% (%d/%d emails)", percent, m.processedEmails, m.totalEmails)
	if m.failedEmails > 0 {
		progressText += fmt.Sprintf(" • %d failed", m.failedEmails)
	}
	
	progressStyle := lipgloss.NewStyle().
		Foreground(ProtonBlue).
		Bold(true).
		Align(lipgloss.Center)
	
	progressContent.WriteString(progressStyle.Render(progressText))
	
	return progressContent.String()
}

func (m *ProgressModel) renderStatistics() string {
	elapsed := time.Since(m.startTime)
	
	var statsContent strings.Builder
	statsContent.WriteString("📊 Statistics:\n\n")
	
	// Time information
	statsContent.WriteString(fmt.Sprintf("• Elapsed time: %s\n", m.formatDuration(elapsed)))
	
	if m.estimatedTimeRemaining > 0 && m.isRunning && !m.isPaused {
		statsContent.WriteString(fmt.Sprintf("• Estimated remaining: %s\n", m.formatDuration(m.estimatedTimeRemaining)))
	}
	
	// Performance information
	if m.emailsPerSecond > 0 {
		statsContent.WriteString(fmt.Sprintf("• Processing speed: %.1f emails/sec\n", m.emailsPerSecond))
	}
	
	// Export information
	if m.app.exportConfig != nil {
		statsContent.WriteString(fmt.Sprintf("• Format: %s\n", strings.ToUpper(m.app.exportConfig.Format)))
		statsContent.WriteString(fmt.Sprintf("• Destination: %s\n", m.app.exportConfig.OutputPath))
		
		if m.app.exportConfig.Encryption {
			statsContent.WriteString("• Encryption: Enabled\n")
		}
	}
	
	statsStyle := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkBorder).
		Padding(1)
	
	return statsStyle.Render(statsContent.String())
}

func (m *ProgressModel) renderCurrentActivity() string {
	if !m.isRunning || m.isCompleted {
		return ""
	}
	
	var activityContent strings.Builder
	
	if m.currentStage != "" {
		activityContent.WriteString(fmt.Sprintf("Stage: %s\n", m.currentStage))
	}
	
	if m.currentEmail != "" {
		// Truncate long email subjects
		email := m.currentEmail
		if len(email) > 60 {
			email = email[:57] + "..."
		}
		activityContent.WriteString(fmt.Sprintf("Processing: %s", email))
	}
	
	if activityContent.Len() == 0 {
		return ""
	}
	
	activityStyle := lipgloss.NewStyle().
		Foreground(ProtonPink).
		Italic(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonPink).
		Padding(1)
	
	return activityStyle.Render(activityContent.String())
}

func (m *ProgressModel) renderControls() string {
	var controls []string
	
	if m.isCompleted || m.hasError {
		controls = append(controls, "Enter: View Results")
	} else if m.isRunning {
		if m.isPaused {
			controls = append(controls, "P/Space: Resume")
		} else {
			controls = append(controls, "P/Space: Pause")
		}
		controls = append(controls, "Q/Ctrl+C: Cancel")
	}
	
	if len(controls) == 0 {
		return ""
	}
	
	controlsText := strings.Join(controls, " • ")
	
	controlsStyle := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Italic(true).
		Align(lipgloss.Center)
	
	return controlsStyle.Render(controlsText)
}

func (m *ProgressModel) calculateMetrics() {
	elapsed := time.Since(m.startTime)
	if elapsed.Seconds() > 0 && m.processedEmails > 0 {
		m.emailsPerSecond = float64(m.processedEmails) / elapsed.Seconds()
		
		if m.emailsPerSecond > 0 && m.totalEmails > m.processedEmails {
			remainingEmails := m.totalEmails - m.processedEmails
			remainingSeconds := float64(remainingEmails) / m.emailsPerSecond
			m.estimatedTimeRemaining = time.Duration(remainingSeconds) * time.Second
		}
	}
}

func (m *ProgressModel) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func (m *ProgressModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return ProgressTickMsg(t)
	})
}

func (m *ProgressModel) startExportProcess() tea.Cmd {
	return func() tea.Msg {
		// This would start the actual export process
		// For now, simulate with mock progress updates
		go m.simulateExport()
		return nil
	}
}

func (m *ProgressModel) simulateExport() {
	// Mock export simulation for demonstration
	totalEmails := 100
	
	for i := 0; i <= totalEmails; i++ {
		if m.ctx.Err() != nil {
			// Export was cancelled
			return
		}
		
		// Simulate processing time
		time.Sleep(100 * time.Millisecond)
		
		// Send progress update
		// In a real implementation, this would be sent from the export goroutine
		currentEmail := fmt.Sprintf("Email %d: Sample subject line", i)
		stage := "Downloading"
		if i > totalEmails/2 {
			stage = "Converting"
		}
		if i > totalEmails*3/4 {
			stage = "Writing"
		}
		
		// This would need to be sent through a proper channel in a real implementation
		// For now, this is just a simulation
	}
	
	// Export completed
	// In a real implementation, this would be sent through a proper channel
}
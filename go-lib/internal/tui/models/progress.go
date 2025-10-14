package models

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal/mail"
	"github.com/ProtonMail/export-tool/internal/session"
	"github.com/ProtonMail/export-tool/internal/tui/styles"
)

// ProgressModel handles progress display during operations
type ProgressModel struct {
	width     int
	height    int
	operation string
	path      string
	session   *session.Session
	
	// Progress tracking
	totalMessages     uint64
	processedMessages uint64
	currentMessage    string
	
	// Status
	running   bool
	completed bool
	success   bool
	errorMsg  string
	
	// UI state
	startTime     time.Time
	lastUpdate    time.Time
	progressWidth int
	
	// Animation
	spinner       int
	spinnerChars  []string
}

// NewProgressModel creates a new progress model
func NewProgressModel() ProgressModel {
	return ProgressModel{
		spinnerChars: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		progressWidth: 50,
	}
}

// Init implements tea.Model
func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.startOperation(),
		m.tickCmd(),
	)
}

// Update implements tea.Model
func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if !m.running {
				return m, tea.Quit
			}
			// Could implement cancellation here
		case "enter":
			if m.completed {
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: ScreenComplete,
						Data: map[string]interface{}{
							"success": m.success,
							"error":   m.errorMsg,
						},
					}
				}
			}
		}
		
	case TickMsg:
		if m.running {
			m.spinner = (m.spinner + 1) % len(m.spinnerChars)
			return m, m.tickCmd()
		}
		
	case ProgressUpdateMsg:
		m.totalMessages = msg.Total
		m.processedMessages = msg.Processed
		m.currentMessage = msg.CurrentMessage
		m.lastUpdate = time.Now()
		
	case OperationCompleteMsg:
		m.running = false
		m.completed = true
		m.success = msg.Success
		m.errorMsg = msg.Error
		
		// Auto-navigate to completion screen after a brief delay
		return m, tea.Tick(time.Second*2, func(time.Time) tea.Msg {
			return NavigateMsg{
				Screen: ScreenComplete,
				Data: map[string]interface{}{
					"success": m.success,
					"error":   m.errorMsg,
				},
			}
		})
	}

	return m, nil
}

// View implements tea.Model
func (m ProgressModel) View() string {
	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Progress content
	progressContent := m.renderProgress()
	sections = append(sections, progressContent)

	// Status
	statusContent := m.renderStatus()
	sections = append(sections, statusContent)

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
func (m *ProgressModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.progressWidth = styles.ResponsiveWidth(width, 0.6)
	if m.progressWidth > 60 {
		m.progressWidth = 60
	}
}

// SetConfig sets the operation configuration
func (m *ProgressModel) SetConfig(operation, path string, session *session.Session) {
	m.operation = operation
	m.path = path
	m.session = session
	m.startTime = time.Now()
	m.running = true
}

// renderHeader creates the progress header
func (m ProgressModel) renderHeader() string {
	var title string
	switch m.operation {
	case "backup":
		title = "Exporting Emails"
	case "restore":
		title = "Importing Emails"
	default:
		title = "Processing"
	}

	subtitle := fmt.Sprintf("Path: %s", m.path)

	headerContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render(title),
		styles.SubtitleStyle.Render(subtitle),
	)

	return styles.HeaderStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(headerContent)
}

// renderProgress creates the progress visualization
func (m ProgressModel) renderProgress() string {
	var sections []string

	// Progress bar
	progressBar := m.renderProgressBar()
	sections = append(sections, progressBar)

	// Statistics
	stats := m.renderStats()
	sections = append(sections, stats)

	// Current activity
	if m.currentMessage != "" {
		activity := styles.TextSecondary.Render(fmt.Sprintf("Processing: %s", m.currentMessage))
		sections = append(sections, activity)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderProgressBar creates the visual progress bar
func (m ProgressModel) renderProgressBar() string {
	var percentage float64
	if m.totalMessages > 0 {
		percentage = float64(m.processedMessages) / float64(m.totalMessages)
	}

	// Calculate filled width
	filledWidth := int(float64(m.progressWidth) * percentage)
	if filledWidth > m.progressWidth {
		filledWidth = m.progressWidth
	}

	// Create progress bar
	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("░", m.progressWidth-filledWidth)
	
	progressBar := styles.ProgressBarStyle.Render(filled + empty)
	
	// Add percentage text
	percentText := fmt.Sprintf("%.1f%%", percentage*100)
	
	// Add spinner if running
	var spinner string
	if m.running {
		spinner = m.spinnerChars[m.spinner] + " "
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		progressBar,
		styles.ProgressTextStyle.Render(fmt.Sprintf("%s%s (%d/%d)", spinner, percentText, m.processedMessages, m.totalMessages)),
	)
}

// renderStats creates the statistics display
func (m ProgressModel) renderStats() string {
	elapsed := time.Since(m.startTime)
	
	var eta string
	if m.totalMessages > 0 && m.processedMessages > 0 && m.running {
		rate := float64(m.processedMessages) / elapsed.Seconds()
		remaining := m.totalMessages - m.processedMessages
		etaSeconds := float64(remaining) / rate
		eta = fmt.Sprintf("ETA: %s", time.Duration(etaSeconds*float64(time.Second)).Round(time.Second))
	} else if m.completed {
		eta = fmt.Sprintf("Completed in: %s", elapsed.Round(time.Second))
	}

	stats := []string{
		fmt.Sprintf("Elapsed: %s", elapsed.Round(time.Second)),
	}
	
	if eta != "" {
		stats = append(stats, eta)
	}

	return styles.TextMuted.Render(strings.Join(stats, " • "))
}

// renderStatus creates the status display
func (m ProgressModel) renderStatus() string {
	if m.completed {
		if m.success {
			return styles.SuccessStyle.Render("✓ Operation completed successfully!")
		} else {
			errorText := "✗ Operation failed"
			if m.errorMsg != "" {
				errorText += ": " + m.errorMsg
			}
			return styles.ErrorStyle.Render(errorText)
		}
	}

	if m.running {
		return styles.ProgressTextStyle.Render("⏳ Operation in progress...")
	}

	return ""
}

// renderFooter creates the help footer
func (m ProgressModel) renderFooter() string {
	var help []string
	
	if m.completed {
		help = []string{
			"Enter: Continue",
			"Ctrl+C: Quit",
		}
	} else {
		help = []string{
			"Ctrl+C: Cancel",
		}
	}

	helpText := strings.Join(help, " • ")

	return styles.FooterStyle.
		Width(styles.ResponsiveWidth(m.width, 0.8)).
		Render(helpText)
}

// startOperation begins the backup or restore operation
func (m ProgressModel) startOperation() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		reporter := &TUIReporter{
			updateChan: make(chan ProgressUpdateMsg, 100),
		}

		// Start the operation in a goroutine
		go func() {
			var err error
			
			if m.operation == "backup" {
				exportTask := mail.NewExportTask(ctx, m.path, m.session)
				err = exportTask.Run(ctx, reporter)
			} else if m.operation == "restore" {
				restoreTask, taskErr := mail.NewRestoreTask(ctx, m.path, m.session)
				if taskErr != nil {
					err = taskErr
				} else {
					err = restoreTask.Run(reporter)
				}
			}

			// Send completion message
			success := err == nil
			errorMsg := ""
			if err != nil {
				errorMsg = err.Error()
			}

			reporter.updateChan <- ProgressUpdateMsg{
				Type: "complete",
				Complete: OperationCompleteMsg{
					Success: success,
					Error:   errorMsg,
				},
			}
		}()

		return ProgressUpdateMsg{Type: "start"}
	}
}

// tickCmd creates a tick command for animation
func (m ProgressModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Messages for progress tracking

// TickMsg is sent for animation updates
type TickMsg time.Time

// ProgressUpdateMsg is sent when progress updates
type ProgressUpdateMsg struct {
	Type           string
	Total          uint64
	Processed      uint64
	CurrentMessage string
	Complete       OperationCompleteMsg
}

// OperationCompleteMsg is sent when operation completes
type OperationCompleteMsg struct {
	Success bool
	Error   string
}

// TUIReporter implements the reporter interface for TUI progress updates
type TUIReporter struct {
	totalMessages   atomic.Uint64
	currentMessages atomic.Uint64
	updateChan      chan ProgressUpdateMsg
}

// SetMessageTotal implements the reporter interface
func (r *TUIReporter) SetMessageTotal(total uint64) {
	r.totalMessages.Store(total)
	// Send update via channel (would need proper channel handling)
}

// SetMessageProcessed implements the reporter interface
func (r *TUIReporter) SetMessageProcessed(processed uint64) {
	r.currentMessages.Store(processed)
	// Send update via channel (would need proper channel handling)
}

// OnProgress implements the reporter interface
func (r *TUIReporter) OnProgress(delta int) {
	current := r.currentMessages.Add(uint64(delta))
	total := r.totalMessages.Load()
	
	// Send update via channel (would need proper channel handling)
	select {
	case r.updateChan <- ProgressUpdateMsg{
		Type:      "progress",
		Total:     total,
		Processed: current,
	}:
	default:
		// Channel full, skip update
	}
}
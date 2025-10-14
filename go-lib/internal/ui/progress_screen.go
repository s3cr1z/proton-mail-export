package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// Progress Screen Model for TUI integration
type ProgressScreenModel struct {
	operation     string
	path          string
	progress      progress.Model
	metrics       ProgressMetrics
	width         int
	height        int
	finished      bool
	error         error
	cancelled     bool
	startTime     time.Time
	style         string
}

func NewProgressScreenModel() ProgressScreenModel {
	p := progress.New(
		progress.WithScaledGradient(protonPurple, protonPink),
		progress.WithWidth(60),
		progress.WithoutPercentage(),
	)

	return ProgressScreenModel{
		progress:  p,
		style:     "enhanced",
		startTime: time.Now(),
	}
}

func (m ProgressScreenModel) Init() tea.Cmd {
	return tea.Batch(
		m.progress.Init(),
		m.startOperation(),
	)
}

func (m ProgressScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, func() tea.Msg {
				return ErrorMsg{fmt.Errorf("operation cancelled by user")}
			}
		}

	case PathSelectedMsg:
		m.path = msg.Path
		return m, m.startOperation()

	case OperationSelectedMsg:
		m.operation = msg.Operation

	case ProgressUpdateMsg:
		m.metrics.ProcessedItems++
		m.metrics.LastUpdateTime = time.Now()
		m.metrics.CurrentOperation = msg.Status
		
		// Update progress bar
		percent := msg.Progress / 100.0
		cmd := m.progress.SetPercent(percent)
		cmds = append(cmds, cmd)

		// Check if complete
		if msg.Progress >= 100.0 {
			m.finished = true
			return m, func() tea.Msg {
				return CompletionMsg{
					Success: true,
					Message: fmt.Sprintf("%s completed successfully", strings.Title(m.operation)),
				}
			}
		}

	case ErrorMsg:
		m.error = msg.Error
		return m, func() tea.Msg {
			return msg
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ProgressScreenModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Title
	title := fmt.Sprintf("%s in Progress", strings.Title(m.operation))
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Path information
	content.WriteString(fmt.Sprintf("Path: %s\n\n", m.path))

	// Current operation
	if m.metrics.CurrentOperation != "" {
		content.WriteString(focusedStyle.Render(m.metrics.CurrentOperation))
		content.WriteString("\n\n")
	}

	// Progress bar
	content.WriteString(m.progress.View())
	content.WriteString("\n\n")

	// Statistics
	if m.style == "enhanced" {
		content.WriteString(m.renderEnhancedStats())
	} else {
		content.WriteString(m.renderSimpleStats())
	}

	// Current item
	if m.metrics.CurrentItem != "" {
		content.WriteString("\n\n")
		content.WriteString(blurredStyle.Render(fmt.Sprintf("Processing: %s", m.metrics.CurrentItem)))
	}

	// Instructions
	content.WriteString("\n\n")
	content.WriteString(blurredStyle.Render("Press Ctrl+C to cancel"))

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2).
		Render(content.String())
}

func (m ProgressScreenModel) renderEnhancedStats() string {
	elapsed := time.Since(m.startTime)
	
	var stats []string
	
	if m.metrics.TotalItems > 0 {
		percent := (float64(m.metrics.ProcessedItems) / float64(m.metrics.TotalItems)) * 100.0
		stats = append(stats, fmt.Sprintf("Items: %d/%d (%.1f%%)", 
			m.metrics.ProcessedItems, m.metrics.TotalItems, percent))
	}

	if m.metrics.FailedItems > 0 {
		stats = append(stats, fmt.Sprintf("Failed: %d", m.metrics.FailedItems))
	}

	stats = append(stats, fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)))

	if m.metrics.ProcessedItems > 0 && m.metrics.TotalItems > 0 {
		remaining := m.estimateTimeRemaining(elapsed)
		stats = append(stats, fmt.Sprintf("Remaining: %s", formatDuration(remaining)))
	}

	if m.metrics.ProcessedBytes > 0 {
		speed := float64(m.metrics.ProcessedBytes) / elapsed.Seconds()
		stats = append(stats, fmt.Sprintf("Speed: %s/s", formatBytes(uint64(speed))))
	}

	if m.metrics.TotalBytes > 0 {
		percent := (float64(m.metrics.ProcessedBytes) / float64(m.metrics.TotalBytes)) * 100.0
		stats = append(stats, fmt.Sprintf("Data: %s/%s (%.1f%%)",
			formatBytes(m.metrics.ProcessedBytes), formatBytes(m.metrics.TotalBytes), percent))
	}

	// Format in a grid
	var content strings.Builder
	for i, stat := range stats {
		if i > 0 && i%2 == 0 {
			content.WriteString("\n")
		} else if i > 0 {
			content.WriteString("  │  ")
		}
		content.WriteString(stat)
	}

	return content.String()
}

func (m ProgressScreenModel) renderSimpleStats() string {
	elapsed := time.Since(m.startTime)
	
	if m.metrics.TotalItems > 0 {
		percent := (float64(m.metrics.ProcessedItems) / float64(m.metrics.TotalItems)) * 100.0
		return fmt.Sprintf("Progress: %.1f%% (%d/%d) • Elapsed: %s",
			percent, m.metrics.ProcessedItems, m.metrics.TotalItems, formatDuration(elapsed))
	}

	return fmt.Sprintf("Elapsed: %s", formatDuration(elapsed))
}

func (m ProgressScreenModel) estimateTimeRemaining(elapsed time.Duration) time.Duration {
	if m.metrics.ProcessedItems == 0 || m.metrics.TotalItems == 0 {
		return 0
	}

	itemsPerSecond := float64(m.metrics.ProcessedItems) / elapsed.Seconds()
	remainingItems := m.metrics.TotalItems - m.metrics.ProcessedItems
	
	if itemsPerSecond <= 0 {
		return 0
	}

	return time.Duration(float64(remainingItems)/itemsPerSecond) * time.Second
}

func (m ProgressScreenModel) startOperation() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// This would integrate with the actual backup/restore logic
		// For now, simulate progress updates
		return m.simulateProgress()
	})
}

func (m ProgressScreenModel) simulateProgress() tea.Msg {
	// This is a placeholder - in real implementation, this would
	// start the actual backup/restore operation and send progress updates
	
	// Initialize metrics
	m.metrics.TotalItems = 100
	m.metrics.StartTime = time.Now()
	
	// Start progress simulation
	go func() {
		for i := 0; i <= 100; i++ {
			if m.cancelled {
				break
			}
			
			time.Sleep(100 * time.Millisecond)
			
			// Send progress update (this would be done by the actual operation)
			// In real implementation, the operation would send these messages
		}
	}()
	
	return ProgressUpdateMsg{
		Progress: 0,
		Status:   fmt.Sprintf("Starting %s operation...", m.operation),
	}
}
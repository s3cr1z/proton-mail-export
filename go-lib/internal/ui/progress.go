package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

const protonPink = "#6d4dfb"   // Pink OCD for accents
const protonPurple = "#4226a2" // Galactic Purple for primary text and buttons
const protonLilac = "#bca4fc"  // Winterspring Lilac for secondary elements

var darkModeStyle = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a1a")).Foreground(lipgloss.Color("#ffffff")) // Dark background with white text as base

type ProgressMetrics struct {
	TotalItems       uint64    `json:"totalItems"`
	ProcessedItems   uint64    `json:"processedItems"`
	FailedItems      uint64    `json:"failedItems"`
	TotalBytes       uint64    `json:"totalBytes"`
	ProcessedBytes   uint64    `json:"processedBytes"`
	StartTime        time.Time `json:"startTime"`
	LastUpdateTime   time.Time `json:"lastUpdateTime"`
	CurrentOperation string    `json:"currentOperation"`
	CurrentItem      string    `json:"currentItem"`
}

func (m ProgressMetrics) GetProgressPercent() float64 {
	if m.TotalItems == 0 {
		return 0.0
	}
	return (float64(m.ProcessedItems) / float64(m.TotalItems)) * 100.0
}

func (m ProgressMetrics) GetElapsedTime() time.Duration {
	return time.Since(m.StartTime)
}

func (m ProgressMetrics) GetEstimatedTimeRemaining() time.Duration {
	if m.ProcessedItems == 0 {
		return 0
	}

	elapsed := m.GetElapsedTime()
	itemsPerSecond := float64(m.ProcessedItems) / elapsed.Seconds()
	remainingItems := m.TotalItems - m.ProcessedItems

	return time.Duration(float64(remainingItems)/itemsPerSecond) * time.Second
}

func (m ProgressMetrics) GetBytesPerSecond() float64 {
	elapsed := m.GetElapsedTime()
	if elapsed.Seconds() == 0 {
		return 0.0
	}
	return float64(m.ProcessedBytes) / elapsed.Seconds()
}

type Model struct {
	progress  progress.Model
	metrics   ProgressMetrics
	finished  bool
	error     string
	cancelled bool
	theme     string
}

type UpdateMsg struct {
	Metrics ProgressMetrics
}

type FinishMsg struct{}
type ErrorMsg struct{ Error string }
type CancelMsg struct{}

func NewModel() Model {
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 60

	return Model{
		progress: p,
		theme:    "default",
	}
}

func (m Model) setTheme(theme string) {
	m.theme = theme
	if theme == "dark-proton" {
		// Apply Proton dark mode styles
		darkModeStyle = darkModeStyle.Copy().Foreground(lipgloss.Color(protonPurple)) // Primary text color
	} else {
		// Reset or use default
		darkModeStyle = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a1a")).Foreground(lipgloss.Color("#ffffff"))
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	case UpdateMsg:
		m.metrics = msg.Metrics
		progressPercent := m.metrics.GetProgressPercent() / 100.0
		return m, m.progress.SetPercent(progressPercent)
	case FinishMsg:
		m.finished = true
		return m, tea.Quit
	case ErrorMsg:
		m.error = msg.Error
		return m, tea.Quit
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.error != "" {
		return m.renderError()
	}

	if m.finished {
		return m.renderFinished()
	}

	return m.renderProgress()
}

func (m Model) renderProgress() string {
	var b strings.Builder

	// Updated title with Proton accent color
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(protonPink)).Render("📧 Proton Mail Export")
	b.WriteString(darkModeStyle.Render(titleStyle) + "\n\n")

	// Current operation with primary color
	if m.metrics.CurrentOperation != "" {
		operationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(protonPurple)).Render(fmt.Sprintf("Operation: %s", m.metrics.CurrentOperation))
		b.WriteString(darkModeStyle.Render(operationStyle) + "\n")
	}

	// Progress bar with accent color
	progressBarStyle := m.progress.Copy().BarStyle(lipgloss.NewStyle().Foreground(lipgloss.Color(protonPink)))
	b.WriteString(progressBarStyle.View() + "\n")

	// Statistics with secondary color
	stats := m.renderStats()
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(protonLilac))
	b.WriteString(darkModeStyle.Render(statsStyle.Render(stats)) + "\n")

	// Current item with faint style
	if m.metrics.CurrentItem != "" {
		currentItemStyle := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Processing: %s", m.metrics.CurrentItem))
		b.WriteString(darkModeStyle.Render(currentItemStyle) + "\n")
	}

	// Controls with faint style
	controlsStyle := lipgloss.NewStyle().Faint(true).Render("Press 'q' or Ctrl+C to cancel")
	b.WriteString("\n" + darkModeStyle.Render(controlsStyle))

	return b.String()
}

func (m Model) renderStats() string {
	percent := m.metrics.GetProgressPercent()
	elapsed := m.metrics.GetElapsedTime()
	remaining := m.metrics.GetEstimatedTimeRemaining()
	bytesPerSec := m.metrics.GetBytesPerSecond()

	stats := []string{
		fmt.Sprintf("Progress: %.1f%% (%d/%d)", percent, m.metrics.ProcessedItems, m.metrics.TotalItems),
		fmt.Sprintf("Failed: %d", m.metrics.FailedItems),
		fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)),
		fmt.Sprintf("Remaining: %s", formatDuration(remaining)),
		fmt.Sprintf("Speed: %s/s", formatBytes(uint64(bytesPerSec))),
	}

	if m.metrics.TotalBytes > 0 {
		bytesPercent := (float64(m.metrics.ProcessedBytes) / float64(m.metrics.TotalBytes)) * 100.0
		stats = append(stats, fmt.Sprintf("Data: %.1f%% (%s/%s)",
			bytesPercent, formatBytes(m.metrics.ProcessedBytes), formatBytes(m.metrics.TotalBytes)))
	}

	var b strings.Builder
	for i, stat := range stats {
		if i > 0 && i%2 == 0 {
			b.WriteString("\n")
		} else if i > 0 {
			b.WriteString("  |  ")
		}
		b.WriteString(stat)
	}

	return b.String()
}

func (m Model) renderFinished() string {
	var b strings.Builder

	// Success message with accent color
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(protonPink)).Render("✅ Export Completed Successfully!")
	b.WriteString(darkModeStyle.Render(successStyle) + "\n\n")

	// Final stats with primary color
	finalStats := []string{
		fmt.Sprintf("Total items processed: %d", m.metrics.ProcessedItems),
		fmt.Sprintf("Failed items: %d", m.metrics.FailedItems),
		fmt.Sprintf("Total time: %s", formatDuration(m.metrics.GetElapsedTime())),
		fmt.Sprintf("Total data: %s", formatBytes(m.metrics.ProcessedBytes)),
	}
	for _, stat := range finalStats {
		statStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(protonPurple))
		b.WriteString(darkModeStyle.Render(statStyle.Render("  "+stat)) + "\n")
	}

	return b.String()
}

func (m Model) renderError() string {
	errorStyleBase := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")) // Red for errors, but use Proton accent for consistency if desired
	errorText := fmt.Sprintf("❌ Error: %s", m.error)
	return darkModeStyle.Render(errorStyleBase.Render(errorText))
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Global state for C interface
var currentProgram *tea.Program
var currentModel *Model

//export bubbletea_create
func bubbletea_create() uintptr {
	model := NewModel()
	currentModel = &model
	return uintptr(0) // Return handle (we use globals for simplicity)
}

//export bubbletea_start
func bubbletea_start() {
	if currentModel != nil {
		currentProgram = tea.NewProgram(*currentModel)
		go currentProgram.Run()
	}
}

//export bubbletea_update
func bubbletea_update(jsonMetrics *C.char) {
	if currentProgram == nil {
		return
	}

	metricsStr := C.GoString(jsonMetrics)
	var metrics ProgressMetrics
	if err := json.Unmarshal([]byte(metricsStr), &metrics); err != nil {
		return
	}

	currentProgram.Send(UpdateMsg{Metrics: metrics})
}

//export bubbletea_finish
func bubbletea_finish() {
	if currentProgram != nil {
		currentProgram.Send(FinishMsg{})
	}
}

//export bubbletea_should_cancel
func bubbletea_should_cancel() C.int {
	if currentModel != nil && currentModel.cancelled {
		return 1
	}
	return 0
}

//export bubbletea_destroy
func bubbletea_destroy() {
	if currentProgram != nil {
		currentProgram.Quit()
		currentProgram = nil
	}
	currentModel = nil
}

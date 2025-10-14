package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type FilterModel struct {
	app           *App
	form          *huh.Form
	width         int
	height        int
	activePreview bool
	previewCount  int
	
	// Form fields
	dateStart     string
	dateEnd       string
	senders       string
	recipients    string
	domains       string
	labels        string
	folders       string
	hasAttachments string
	minSizeStr    string
	maxSizeStr    string
	searchQuery   string
}

type FilterPreviewMsg struct {
	Count int
	Error string
}

func NewFilterModel(app *App) *FilterModel {
	m := &FilterModel{
		app: app,
	}
	
	// Initialize with current filter config
	if app.filterConfig != nil {
		m.dateStart = app.filterConfig.DateStart
		m.dateEnd = app.filterConfig.DateEnd
		m.senders = strings.Join(app.filterConfig.Senders, ", ")
		m.recipients = strings.Join(app.filterConfig.Recipients, ", ")
		m.domains = strings.Join(app.filterConfig.Domains, ", ")
		m.labels = strings.Join(app.filterConfig.Labels, ", ")
		m.folders = strings.Join(app.filterConfig.Folders, ", ")
		m.searchQuery = app.filterConfig.SearchQuery
		
		if app.filterConfig.HasAttachments != nil {
			if *app.filterConfig.HasAttachments {
				m.hasAttachments = "yes"
			} else {
				m.hasAttachments = "no"
			}
		} else {
			m.hasAttachments = "any"
		}
		
		if app.filterConfig.MinSize != nil {
			m.minSizeStr = fmt.Sprintf("%d", *app.filterConfig.MinSize)
		}
		if app.filterConfig.MaxSize != nil {
			m.maxSizeStr = fmt.Sprintf("%d", *app.filterConfig.MaxSize)
		}
	}
	
	m.initForm()
	return m
}

func (m *FilterModel) initForm() {
	m.form = huh.NewForm(
		// Date Range Group
		huh.NewGroup(
			huh.NewInput().
				Title("Start Date").
				Description("Format: YYYY-MM-DD (leave empty for no limit)").
				Placeholder("2023-01-01").
				Value(&m.dateStart).
				Validate(m.validateDate),
			
			huh.NewInput().
				Title("End Date").
				Description("Format: YYYY-MM-DD (leave empty for no limit)").
				Placeholder("2023-12-31").
				Value(&m.dateEnd).
				Validate(m.validateDate),
		).Title("📅 Date Range").
			Description("Filter emails by date range"),
		
		// Sender/Recipient Group
		huh.NewGroup(
			huh.NewInput().
				Title("Senders").
				Description("Comma-separated email addresses or patterns").
				Placeholder("user@example.com, *@company.com").
				Value(&m.senders),
			
			huh.NewInput().
				Title("Recipients").
				Description("Comma-separated email addresses or patterns").
				Placeholder("recipient@example.com, *@domain.com").
				Value(&m.recipients),
			
			huh.NewInput().
				Title("Domains").
				Description("Comma-separated domain names").
				Placeholder("proton.me, gmail.com").
				Value(&m.domains),
		).Title("👥 People & Domains").
			Description("Filter by sender, recipient, or domain"),
		
		// Organization Group
		huh.NewGroup(
			huh.NewInput().
				Title("Labels").
				Description("Comma-separated label names").
				Placeholder("Important, Work, Personal").
				Value(&m.labels),
			
			huh.NewInput().
				Title("Folders").
				Description("Comma-separated folder names").
				Placeholder("Inbox, Sent, Archive").
				Value(&m.folders),
		).Title("🏷️ Organization").
			Description("Filter by labels and folders"),
		
		// Content & Size Group
		huh.NewGroup(
			huh.NewInput().
				Title("Search Query").
				Description("Search in subject and body text").
				Placeholder("meeting agenda").
				Value(&m.searchQuery),
			
			huh.NewSelect[string]().
				Title("Has Attachments").
				Options(
					huh.NewOption("Any", "any"),
					huh.NewOption("Yes", "yes"),
					huh.NewOption("No", "no"),
				).
				Value(&m.hasAttachments),
			
			huh.NewInput().
				Title("Minimum Size (KB)").
				Description("Minimum email size in kilobytes").
				Placeholder("100").
				Value(&m.minSizeStr).
				Validate(m.validateSize),
			
			huh.NewInput().
				Title("Maximum Size (KB)").
				Description("Maximum email size in kilobytes").
				Placeholder("10000").
				Value(&m.maxSizeStr).
				Validate(m.validateSize),
		).Title("📄 Content & Size").
			Description("Filter by content and size"),
		
	).WithTheme(huh.ThemeDracula())
}

func (m *FilterModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *FilterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, m.app.SwitchToScreen(MainMenuScreen)
		case "ctrl+p":
			// Toggle preview
			m.activePreview = !m.activePreview
			if m.activePreview {
				return m, m.generatePreview()
			}
		case "ctrl+r":
			// Reset filters
			return m, m.resetFilters()
		case "ctrl+s":
			// Save and apply filters
			return m, m.saveFilters()
		}
		
	case FilterPreviewMsg:
		m.previewCount = msg.Count
		if msg.Error != "" {
			// Handle preview error
		}
	}
	
	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		// Auto-generate preview when form changes
		if m.activePreview {
			cmds = append(cmds, m.generatePreview())
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *FilterModel) View() string {
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	content.WriteString(titleStyle.Render("🔍 Advanced Email Filters"))
	content.WriteString("\n")
	
	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Italic(true).
		Align(lipgloss.Center).
		Margin(0, 0, 1, 0)
	
	instructions := "Configure filters to narrow down your email export • Ctrl+P: Preview • Ctrl+R: Reset • Ctrl+S: Save"
	content.WriteString(instructionStyle.Render(instructions))
	content.WriteString("\n")
	
	// Form
	formView := m.form.View()
	content.WriteString(formView)
	
	// Preview section
	if m.activePreview {
		previewSection := m.renderPreview()
		content.WriteString("\n")
		content.WriteString(previewSection)
	}
	
	// Action buttons
	actionSection := m.renderActions()
	content.WriteString("\n")
	content.WriteString(actionSection)
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height - 4) // Account for header/footer
	
	return containerStyle.Render(content.String())
}

func (m *FilterModel) renderPreview() string {
	previewStyle := lipgloss.NewStyle().
		Foreground(ProtonBlue).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonBlue).
		Padding(1).
		Margin(1, 0)
	
	var previewText string
	if m.previewCount >= 0 {
		previewText = fmt.Sprintf("📊 Preview: %d emails match these filters", m.previewCount)
	} else {
		previewText = "📊 Preview: Calculating..."
	}
	
	return previewStyle.Render(previewText)
}

func (m *FilterModel) renderActions() string {
	actionStyle := lipgloss.NewStyle().
		Foreground(ProtonGreen).
		Bold(true).
		Align(lipgloss.Center)
	
	actions := "Enter: Apply Filters • Esc: Back to Main Menu • Ctrl+R: Reset All"
	return actionStyle.Render(actions)
}

func (m *FilterModel) validateDate(s string) error {
	if s == "" {
		return nil // Empty is allowed
	}
	
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	
	return nil
}

func (m *FilterModel) validateSize(s string) error {
	if s == "" {
		return nil // Empty is allowed
	}
	
	size, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid size, must be a number")
	}
	
	if size < 0 {
		return fmt.Errorf("size cannot be negative")
	}
	
	return nil
}

func (m *FilterModel) generatePreview() tea.Cmd {
	return func() tea.Msg {
		// This would call the backend to get a count of matching emails
		// For now, return a mock count
		config := m.buildFilterConfig()
		count := m.mockPreviewCount(config)
		
		return FilterPreviewMsg{
			Count: count,
		}
	}
}

func (m *FilterModel) mockPreviewCount(config *FilterConfig) int {
	// Mock implementation - in real app this would query the backend
	baseCount := 1000
	
	// Reduce count based on filters
	if config.DateStart != "" || config.DateEnd != "" {
		baseCount = baseCount * 60 / 100 // 60% of emails
	}
	
	if len(config.Senders) > 0 {
		baseCount = baseCount * 30 / 100 // 30% of emails
	}
	
	if config.SearchQuery != "" {
		baseCount = baseCount * 20 / 100 // 20% of emails
	}
	
	if config.HasAttachments != nil {
		baseCount = baseCount * 40 / 100 // 40% of emails
	}
	
	return baseCount
}

func (m *FilterModel) resetFilters() tea.Cmd {
	return func() tea.Msg {
		// Reset all form fields
		m.dateStart = ""
		m.dateEnd = ""
		m.senders = ""
		m.recipients = ""
		m.domains = ""
		m.labels = ""
		m.folders = ""
		m.hasAttachments = "any"
		m.minSizeStr = ""
		m.maxSizeStr = ""
		m.searchQuery = ""
		
		// Reinitialize form
		m.initForm()
		
		return nil
	}
}

func (m *FilterModel) saveFilters() tea.Cmd {
	return func() tea.Msg {
		// Build and save filter config
		config := m.buildFilterConfig()
		m.app.filterConfig = config
		
		// Switch to export screen
		return ScreenChangeMsg{Screen: ExportScreen}
	}
}

func (m *FilterModel) buildFilterConfig() *FilterConfig {
	config := &FilterConfig{
		DateStart:   m.dateStart,
		DateEnd:     m.dateEnd,
		SearchQuery: m.searchQuery,
	}
	
	// Parse comma-separated lists
	if m.senders != "" {
		config.Senders = m.parseCommaSeparated(m.senders)
	}
	if m.recipients != "" {
		config.Recipients = m.parseCommaSeparated(m.recipients)
	}
	if m.domains != "" {
		config.Domains = m.parseCommaSeparated(m.domains)
	}
	if m.labels != "" {
		config.Labels = m.parseCommaSeparated(m.labels)
	}
	if m.folders != "" {
		config.Folders = m.parseCommaSeparated(m.folders)
	}
	
	// Parse attachments filter
	if m.hasAttachments == "yes" {
		hasAttachments := true
		config.HasAttachments = &hasAttachments
	} else if m.hasAttachments == "no" {
		hasAttachments := false
		config.HasAttachments = &hasAttachments
	}
	
	// Parse size filters
	if m.minSizeStr != "" {
		if minSize, err := strconv.ParseInt(m.minSizeStr, 10, 64); err == nil {
			config.MinSize = &minSize
		}
	}
	if m.maxSizeStr != "" {
		if maxSize, err := strconv.ParseInt(m.maxSizeStr, 10, 64); err == nil {
			config.MaxSize = &maxSize
		}
	}
	
	return config
}

func (m *FilterModel) parseCommaSeparated(input string) []string {
	var result []string
	parts := strings.Split(input, ",")
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}
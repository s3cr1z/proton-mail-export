package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type ExportModel struct {
	app           *App
	form          *huh.Form
	width         int
	height        int
	
	// Form fields
	operation     string
	format        string
	outputPath    string
	encryption    bool
	encryptionPwd string
	incremental   bool
}

func NewExportModel(app *App) *ExportModel {
	m := &ExportModel{
		app: app,
	}
	
	// Initialize with current export config
	if app.exportConfig != nil {
		m.operation = app.exportConfig.Operation
		m.format = app.exportConfig.Format
		m.outputPath = app.exportConfig.OutputPath
		m.encryption = app.exportConfig.Encryption
		m.incremental = app.exportConfig.Incremental
	}
	
	// Set defaults if empty
	if m.operation == "" {
		m.operation = "backup"
	}
	if m.format == "" {
		m.format = "eml"
	}
	if m.outputPath == "" {
		m.outputPath = m.getDefaultOutputPath()
	}
	
	m.initForm()
	return m
}

func (m *ExportModel) initForm() {
	m.form = huh.NewForm(
		// Operation Group
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Operation").
				Description("Choose what you want to do").
				Options(
					huh.NewOption("Export/Backup Emails", "backup"),
					huh.NewOption("Import/Restore Emails", "restore"),
				).
				Value(&m.operation),
		).Title("🎯 Operation").
			Description("Select the operation to perform"),
		
		// Format Group (only for backup)
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Export Format").
				Description("Choose the output format for your emails").
				Options(
					huh.NewOption("EML (Standard Email Format)", "eml"),
					huh.NewOption("PDF (Portable Document Format)", "pdf"),
					huh.NewOption("MBOX (Mailbox Format)", "mbox"),
				).
				Value(&m.format),
		).Title("📄 Export Format").
			Description("Select how you want your emails saved").
			WithHideFunc(func() bool { return m.operation != "backup" }),
		
		// Path Group
		huh.NewGroup(
			huh.NewInput().
				Title("Output Path").
				Description("Where to save the exported emails").
				Placeholder("/path/to/export/folder").
				Value(&m.outputPath).
				Validate(m.validatePath),
		).Title("📁 Destination").
			Description("Choose where to save your export"),
		
		// Security Group
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Encryption").
				Description("Encrypt the exported data with a password").
				Value(&m.encryption),
			
			huh.NewInput().
				Title("Encryption Password").
				Description("Password to encrypt the export (leave empty to disable)").
				EchoMode(huh.EchoModePassword).
				Value(&m.encryptionPwd).
				WithHideFunc(func() bool { return !m.encryption }),
		).Title("🔒 Security").
			Description("Protect your exported data"),
		
		// Options Group
		huh.NewGroup(
			huh.NewConfirm().
				Title("Incremental Backup").
				Description("Only export emails that have changed since last backup").
				Value(&m.incremental).
				WithHideFunc(func() bool { return m.operation != "backup" }),
		).Title("⚡ Options").
			Description("Additional export options"),
		
	).WithTheme(huh.ThemeDracula())
}

func (m *ExportModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *ExportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, m.app.SwitchToScreen(MainMenuScreen)
		case "f":
			// Quick access to filters
			return m, m.app.SwitchToScreen(FilterScreen)
		}
	}
	
	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		// Check if form is complete
		if m.form.State == huh.StateCompleted {
			return m, m.startExport()
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *ExportModel) View() string {
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	var title string
	if m.operation == "restore" {
		title = "📥 Import/Restore Configuration"
	} else {
		title = "📤 Export/Backup Configuration"
	}
	
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n")
	
	// Current filter summary
	filterSummary := m.renderFilterSummary()
	content.WriteString(filterSummary)
	content.WriteString("\n")
	
	// Form
	formView := m.form.View()
	content.WriteString(formView)
	
	// Export preview
	exportPreview := m.renderExportPreview()
	content.WriteString("\n")
	content.WriteString(exportPreview)
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height - 4) // Account for header/footer
	
	return containerStyle.Render(content.String())
}

func (m *ExportModel) renderFilterSummary() string {
	filterStyle := lipgloss.NewStyle().
		Foreground(ProtonBlue).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonBlue).
		Padding(1).
		Margin(0, 0, 1, 0)
	
	var filterText strings.Builder
	filterText.WriteString("🔍 Active Filters: ")
	
	if m.app.filterConfig == nil || m.isFilterEmpty() {
		filterText.WriteString("None (all emails will be processed)")
	} else {
		filters := m.getActiveFilters()
		if len(filters) > 0 {
			filterText.WriteString(strings.Join(filters, ", "))
		} else {
			filterText.WriteString("None")
		}
	}
	
	filterText.WriteString("\n\nPress 'F' to configure filters")
	
	return filterStyle.Render(filterText.String())
}

func (m *ExportModel) renderExportPreview() string {
	previewStyle := lipgloss.NewStyle().
		Foreground(ProtonGreen).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonGreen).
		Padding(1)
	
	var previewText strings.Builder
	previewText.WriteString("📋 Export Summary:\n\n")
	
	previewText.WriteString(fmt.Sprintf("• Operation: %s\n", strings.Title(m.operation)))
	if m.operation == "backup" {
		previewText.WriteString(fmt.Sprintf("• Format: %s\n", strings.ToUpper(m.format)))
	}
	previewText.WriteString(fmt.Sprintf("• Destination: %s\n", m.outputPath))
	
	if m.encryption {
		previewText.WriteString("• Encryption: Enabled\n")
	} else {
		previewText.WriteString("• Encryption: Disabled\n")
	}
	
	if m.operation == "backup" && m.incremental {
		previewText.WriteString("• Mode: Incremental backup\n")
	}
	
	// Estimated count (mock)
	estimatedCount := m.getEstimatedEmailCount()
	previewText.WriteString(fmt.Sprintf("• Estimated emails: %d\n", estimatedCount))
	
	return previewStyle.Render(previewText.String())
}

func (m *ExportModel) validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	
	// Check if parent directory exists or can be created
	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		// Try to create parent directory
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("cannot create directory: %v", err)
		}
	}
	
	return nil
}

func (m *ExportModel) getDefaultOutputPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./proton-mail-export"
	}
	
	return filepath.Join(homeDir, "Documents", "ProtonMailExport")
}

func (m *ExportModel) isFilterEmpty() bool {
	if m.app.filterConfig == nil {
		return true
	}
	
	config := m.app.filterConfig
	return config.DateStart == "" &&
		config.DateEnd == "" &&
		len(config.Senders) == 0 &&
		len(config.Recipients) == 0 &&
		len(config.Domains) == 0 &&
		len(config.Labels) == 0 &&
		len(config.Folders) == 0 &&
		config.HasAttachments == nil &&
		config.MinSize == nil &&
		config.MaxSize == nil &&
		config.SearchQuery == ""
}

func (m *ExportModel) getActiveFilters() []string {
	if m.app.filterConfig == nil {
		return nil
	}
	
	var filters []string
	config := m.app.filterConfig
	
	if config.DateStart != "" || config.DateEnd != "" {
		filters = append(filters, "Date range")
	}
	if len(config.Senders) > 0 {
		filters = append(filters, fmt.Sprintf("Senders (%d)", len(config.Senders)))
	}
	if len(config.Recipients) > 0 {
		filters = append(filters, fmt.Sprintf("Recipients (%d)", len(config.Recipients)))
	}
	if len(config.Domains) > 0 {
		filters = append(filters, fmt.Sprintf("Domains (%d)", len(config.Domains)))
	}
	if len(config.Labels) > 0 {
		filters = append(filters, fmt.Sprintf("Labels (%d)", len(config.Labels)))
	}
	if len(config.Folders) > 0 {
		filters = append(filters, fmt.Sprintf("Folders (%d)", len(config.Folders)))
	}
	if config.HasAttachments != nil {
		if *config.HasAttachments {
			filters = append(filters, "Has attachments")
		} else {
			filters = append(filters, "No attachments")
		}
	}
	if config.MinSize != nil || config.MaxSize != nil {
		filters = append(filters, "Size limits")
	}
	if config.SearchQuery != "" {
		filters = append(filters, "Search query")
	}
	
	return filters
}

func (m *ExportModel) getEstimatedEmailCount() int {
	// Mock implementation - in real app this would query the backend
	baseCount := 1000
	
	if m.app.filterConfig != nil && !m.isFilterEmpty() {
		// Apply filter reduction (mock calculation)
		baseCount = baseCount * 30 / 100 // Assume filters reduce to 30%
	}
	
	return baseCount
}

func (m *ExportModel) startExport() tea.Cmd {
	return func() tea.Msg {
		// Save export configuration
		m.app.exportConfig.Operation = m.operation
		m.app.exportConfig.Format = m.format
		m.app.exportConfig.OutputPath = m.outputPath
		m.app.exportConfig.Encryption = m.encryption
		m.app.exportConfig.Incremental = m.incremental
		
		// Switch to progress screen
		return ScreenChangeMsg{Screen: ProgressScreen}
	}
}
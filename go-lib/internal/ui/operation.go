package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Operation types
const (
	OperationBackup  = "backup"
	OperationRestore = "restore"
)

// Operation Model
type OperationModel struct {
	selected    string
	form        *huh.Form
	width       int
	height      int
	showAdvanced bool
	options     OperationOptions
}

type OperationOptions struct {
	Operation        string
	Format          string
	Encrypt         bool
	EncryptPassword string
	Incremental     bool
	DateStart       string
	DateEnd         string
	Senders         []string
	HasAttachments  *bool
	ProgressStyle   string
}

func NewOperationModel() OperationModel {
	return OperationModel{
		options: OperationOptions{
			Format:        "eml",
			ProgressStyle: "enhanced",
		},
	}
}

func (m OperationModel) Init() tea.Cmd {
	return m.createForm()
}

func (m OperationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: LoginScreen}
			}
		case "tab":
			m.showAdvanced = !m.showAdvanced
			return m, m.createForm()
		}

	case huh.FormCompleteMsg:
		return m, func() tea.Msg {
			return OperationSelectedMsg{Operation: m.selected}
		}

	case OperationSelectedMsg:
		m.selected = msg.Operation
		m.options.Operation = msg.Operation
	}

	// Update form
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		return m, cmd
	}

	return m, nil
}

func (m OperationModel) View() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Select Operation"))
	content.WriteString("\n\n")

	// Description
	content.WriteString("Choose what you want to do:\n\n")

	// Show form
	if m.form != nil {
		content.WriteString(m.form.View())
	}

	// Show advanced options toggle
	content.WriteString("\n\n")
	if m.showAdvanced {
		content.WriteString(focusedStyle.Render("Advanced Options (TAB to hide)"))
	} else {
		content.WriteString(blurredStyle.Render("Press TAB for advanced options"))
	}

	// Instructions
	content.WriteString("\n\n")
	content.WriteString(blurredStyle.Render("Press ESC to go back • Ctrl+C to quit"))

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2).
		Render(content.String())
}

func (m OperationModel) createForm() tea.Cmd {
	var groups []*huh.Group

	// Basic operation selection
	basicGroup := huh.NewGroup(
		huh.NewSelect[string]().
			Key("operation").
			Title("Operation").
			Options(
				huh.NewOption("Backup emails to local storage", OperationBackup),
				huh.NewOption("Restore emails from backup", OperationRestore),
			).
			Value(&m.selected),
	)
	groups = append(groups, basicGroup)

	// Advanced options
	if m.showAdvanced {
		// Format options
		formatGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Key("format").
				Title("Export Format").
				Options(
					huh.NewOption("EML (Email Message Format)", "eml"),
					huh.NewOption("PDF (Portable Document Format)", "pdf"),
				).
				Value(&m.options.Format),
		)
		groups = append(groups, formatGroup)

		// Encryption options
		encryptionGroup := huh.NewGroup(
			huh.NewConfirm().
				Key("encrypt").
				Title("Enable Encryption").
				Description("Encrypt backup files for additional security").
				Value(&m.options.Encrypt),
		)
		groups = append(groups, encryptionGroup)

		if m.options.Encrypt {
			passwordGroup := huh.NewGroup(
				huh.NewInput().
					Key("encrypt_password").
					Title("Encryption Password").
					EchoMode(huh.EchoModePassword).
					Value(&m.options.EncryptPassword).
					Validate(func(s string) error {
						if len(s) < 8 {
							return fmt.Errorf("password must be at least 8 characters")
						}
						return nil
					}),
			)
			groups = append(groups, passwordGroup)
		}

		// Backup-specific options
		if m.selected == OperationBackup {
			backupGroup := huh.NewGroup(
				huh.NewConfirm().
					Key("incremental").
					Title("Incremental Backup").
					Description("Only backup new/changed emails since last backup").
					Value(&m.options.Incremental),
			)
			groups = append(groups, backupGroup)

			// Date filtering
			dateGroup := huh.NewGroup(
				huh.NewInput().
					Key("date_start").
					Title("Start Date (optional)").
					Placeholder("YYYY-MM-DD").
					Value(&m.options.DateStart).
					Validate(validateDate),
				huh.NewInput().
					Key("date_end").
					Title("End Date (optional)").
					Placeholder("YYYY-MM-DD").
					Value(&m.options.DateEnd).
					Validate(validateDate),
			)
			groups = append(groups, dateGroup)

			// Attachment filtering
			attachmentGroup := huh.NewGroup(
				huh.NewSelect[string]().
					Key("has_attachments").
					Title("Filter by Attachments").
					Options(
						huh.NewOption("All emails", "all"),
						huh.NewOption("Only emails with attachments", "with"),
						huh.NewOption("Only emails without attachments", "without"),
					).
					Value(func() *string {
						if m.options.HasAttachments == nil {
							s := "all"
							return &s
						}
						if *m.options.HasAttachments {
							s := "with"
							return &s
						}
						s := "without"
						return &s
					}()),
			)
			groups = append(groups, attachmentGroup)
		}

		// Progress style
		progressGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Key("progress_style").
				Title("Progress Display").
				Options(
					huh.NewOption("Enhanced progress with details", "enhanced"),
					huh.NewOption("Simple progress bar", "simple"),
				).
				Value(&m.options.ProgressStyle),
		)
		groups = append(groups, progressGroup)
	}

	m.form = huh.NewForm(groups...).WithTheme(huh.ThemeCharm())
	return m.form.Init()
}

func validateDate(s string) error {
	if s == "" {
		return nil // Optional field
	}
	// Simple date validation - in real implementation, use time.Parse
	if len(s) != 10 {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
	}
	return nil
}

// Get the configured options
func (m OperationModel) GetOptions() OperationOptions {
	return m.options
}
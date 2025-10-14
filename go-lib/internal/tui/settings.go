package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type SettingsModel struct {
	app    *App
	form   *huh.Form
	width  int
	height int
	
	// Form fields
	theme           string
	autoSave        bool
	defaultFormat   string
	defaultPath     string
	progressStyle   string
	notifications   bool
}

func NewSettingsModel(app *App) *SettingsModel {
	m := &SettingsModel{
		app: app,
	}
	
	// Initialize with current settings
	if app.userSettings != nil {
		m.theme = app.userSettings.Theme
		m.autoSave = app.userSettings.AutoSave
		m.defaultFormat = app.userSettings.DefaultFormat
		m.defaultPath = app.userSettings.DefaultPath
		m.progressStyle = app.userSettings.ProgressStyle
		m.notifications = app.userSettings.Notifications
	}
	
	m.initForm()
	return m
}

func (m *SettingsModel) initForm() {
	m.form = huh.NewForm(
		// Appearance Group
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Theme").
				Description("Choose the application theme").
				Options(
					huh.NewOption("Dark (Proton)", "dark"),
					huh.NewOption("Light", "light"),
					huh.NewOption("Auto (System)", "auto"),
				).
				Value(&m.theme),
			
			huh.NewSelect[string]().
				Title("Progress Style").
				Description("Choose how progress is displayed").
				Options(
					huh.NewOption("Enhanced (Detailed)", "enhanced"),
					huh.NewOption("Simple (Minimal)", "simple"),
					huh.NewOption("Compact", "compact"),
				).
				Value(&m.progressStyle),
		).Title("🎨 Appearance").
			Description("Customize the look and feel"),
		
		// Defaults Group
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Default Export Format").
				Description("Default format for new exports").
				Options(
					huh.NewOption("EML (Email Format)", "eml"),
					huh.NewOption("PDF (Document)", "pdf"),
					huh.NewOption("MBOX (Mailbox)", "mbox"),
				).
				Value(&m.defaultFormat),
			
			huh.NewInput().
				Title("Default Export Path").
				Description("Default location for exports").
				Placeholder("~/Documents/ProtonMailExport").
				Value(&m.defaultPath),
		).Title("📁 Defaults").
			Description("Set default values for exports"),
		
		// Behavior Group
		huh.NewGroup(
			huh.NewConfirm().
				Title("Auto-save Settings").
				Description("Automatically save settings when changed").
				Value(&m.autoSave),
			
			huh.NewConfirm().
				Title("Enable Notifications").
				Description("Show system notifications for completed exports").
				Value(&m.notifications),
		).Title("⚙️ Behavior").
			Description("Configure application behavior"),
		
	).WithTheme(huh.ThemeDracula())
}

func (m *SettingsModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, m.app.SwitchToScreen(MainMenuScreen)
		case "ctrl+s":
			return m, m.saveSettings()
		}
	}
	
	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		// Auto-save if enabled and form is complete
		if m.form.State == huh.StateCompleted {
			return m, m.saveSettings()
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *SettingsModel) View() string {
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	content.WriteString(titleStyle.Render("⚙️ Settings & Preferences"))
	content.WriteString("\n")
	
	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Italic(true).
		Align(lipgloss.Center).
		Margin(0, 0, 1, 0)
	
	instructions := "Configure your preferences • Ctrl+S: Save • Esc: Back to Main Menu"
	content.WriteString(instructionStyle.Render(instructions))
	content.WriteString("\n")
	
	// Form
	formView := m.form.View()
	content.WriteString(formView)
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height - 4) // Account for header/footer
	
	return containerStyle.Render(content.String())
}

func (m *SettingsModel) saveSettings() tea.Cmd {
	return func() tea.Msg {
		// Update app settings
		m.app.userSettings.Theme = m.theme
		m.app.userSettings.AutoSave = m.autoSave
		m.app.userSettings.DefaultFormat = m.defaultFormat
		m.app.userSettings.DefaultPath = m.defaultPath
		m.app.userSettings.ProgressStyle = m.progressStyle
		m.app.userSettings.Notifications = m.notifications
		
		// In a real implementation, this would save to a config file
		// For now, just return to main menu
		return ScreenChangeMsg{Screen: MainMenuScreen}
	}
}
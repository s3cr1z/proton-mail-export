package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	app    *App
	list   list.Model
	width  int
	height int
}

type menuItem struct {
	title       string
	description string
	action      string
	icon        string
}

func (i menuItem) Title() string       { return i.icon + " " + i.title }
func (i menuItem) Description() string { return i.description }
func (i menuItem) FilterValue() string { return i.title }

func NewMainMenuModel(app *App) *MainMenuModel {
	items := []list.Item{
		menuItem{
			title:       "Export Emails",
			description: "Export your emails to various formats (EML, PDF, MBOX)",
			action:      "export",
			icon:        "📤",
		},
		menuItem{
			title:       "Import/Restore",
			description: "Import emails from a previous backup",
			action:      "import",
			icon:        "📥",
		},
		menuItem{
			title:       "Advanced Filters",
			description: "Configure advanced filtering options",
			action:      "filters",
			icon:        "🔍",
		},
		menuItem{
			title:       "Export History",
			description: "View and manage previous exports",
			action:      "history",
			icon:        "📋",
		},
		menuItem{
			title:       "Settings",
			description: "Configure application preferences",
			action:      "settings",
			icon:        "⚙️",
		},
		menuItem{
			title:       "Help & Support",
			description: "Get help and view documentation",
			action:      "help",
			icon:        "❓",
		},
		menuItem{
			title:       "Logout",
			description: "Sign out and return to login screen",
			action:      "logout",
			icon:        "🚪",
		},
	}

	// Create list with custom styling
	l := list.New(items, newMenuDelegate(), 0, 0)
	l.Title = "Main Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Padding(0, 0, 1, 0)

	return &MainMenuModel{
		app:  app,
		list: l,
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if selectedItem, ok := m.list.SelectedItem().(menuItem); ok {
				return m.handleMenuAction(selectedItem.action)
			}
		case "s":
			return m, m.app.SwitchToScreen(SettingsScreen)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MainMenuModel) View() string {
	// User info section
	userInfo := m.renderUserInfo()
	
	// Main menu list
	menuView := m.list.View()
	
	// Quick stats section
	quickStats := m.renderQuickStats()
	
	// Combine sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		userInfo,
		menuView,
		quickStats,
	)
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height - 4) // Account for header/footer
	
	return containerStyle.Render(content)
}

func (m *MainMenuModel) renderUserInfo() string {
	var email string
	if m.app.session != nil && m.app.session.GetUser() != nil {
		email = m.app.session.GetUser().Email
	} else {
		email = "Not logged in"
	}
	
	userStyle := lipgloss.NewStyle().
		Foreground(ProtonBlue).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ProtonBlue).
		Padding(1).
		Margin(0, 0, 1, 0)
	
	userText := fmt.Sprintf("👤 Logged in as: %s", email)
	return userStyle.Render(userText)
}

func (m *MainMenuModel) renderQuickStats() string {
	// This would show recent activity, storage usage, etc.
	statsStyle := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DarkBorder).
		Padding(1).
		Margin(1, 0, 0, 0)
	
	statsText := `📊 Quick Stats:
• Last export: Never
• Total exports: 0
• Storage used: 0 MB`
	
	return statsStyle.Render(statsText)
}

func (m *MainMenuModel) handleMenuAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "export":
		return m, m.app.SwitchToScreen(ExportScreen)
	case "import":
		// Set import mode and go to export screen
		m.app.exportConfig.Operation = "restore"
		return m, m.app.SwitchToScreen(ExportScreen)
	case "filters":
		return m, m.app.SwitchToScreen(FilterScreen)
	case "history":
		// TODO: Implement history screen
		return m, nil
	case "settings":
		return m, m.app.SwitchToScreen(SettingsScreen)
	case "help":
		// TODO: Implement help screen
		return m, nil
	case "logout":
		// Clear session and return to login
		m.app.session = nil
		return m, m.app.SwitchToScreen(LoginScreen)
	default:
		return m, nil
	}
}

// Custom list delegate for menu items
func newMenuDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPink).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ProtonPink).
		Padding(0, 0, 0, 1)
	
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(ProtonPurple).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(ProtonPink).
		Padding(0, 0, 0, 1)
	
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(DarkFg).
		Padding(0, 0, 0, 1)
	
	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(ProtonGray).
		Padding(0, 0, 0, 1)
	
	return d
}
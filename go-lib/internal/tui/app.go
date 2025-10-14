package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal"
	"github.com/ProtonMail/export-tool/internal/session"
)

// Proton brand colors
var (
	ProtonPurple = lipgloss.Color("#6d4aff")
	ProtonPink   = lipgloss.Color("#ff6ec7") 
	ProtonBlue   = lipgloss.Color("#1c7cd6")
	ProtonGreen  = lipgloss.Color("#1db584")
	ProtonRed    = lipgloss.Color("#dc3545")
	ProtonGray   = lipgloss.Color("#6c757d")
	
	// Dark theme colors
	DarkBg       = lipgloss.Color("#1a1a1a")
	DarkFg       = lipgloss.Color("#ffffff")
	DarkBorder   = lipgloss.Color("#404040")
)

// Screen types
type Screen int

const (
	WelcomeScreen Screen = iota
	LoginScreen
	MainMenuScreen
	FilterScreen
	ExportScreen
	ProgressScreen
	ResultsScreen
	SettingsScreen
)

// Main TUI application model
type App struct {
	ctx           context.Context
	session       *session.Session
	currentScreen Screen
	screens       map[Screen]tea.Model
	width         int
	height        int
	err           error
	quitting      bool
	
	// Shared state
	exportConfig  *ExportConfig
	filterConfig  *FilterConfig
	userSettings  *UserSettings
}

// Export configuration
type ExportConfig struct {
	Operation    string // "backup" or "restore"
	Format       string // "eml", "pdf", "mbox"
	OutputPath   string
	Encryption   bool
	Incremental  bool
}

// Filter configuration
type FilterConfig struct {
	DateStart    string
	DateEnd      string
	Senders      []string
	Recipients   []string
	Domains      []string
	Labels       []string
	Folders      []string
	HasAttachments *bool
	MinSize      *int64
	MaxSize      *int64
	SearchQuery  string
}

// User settings
type UserSettings struct {
	Theme           string
	AutoSave        bool
	DefaultFormat   string
	DefaultPath     string
	ProgressStyle   string
	Notifications   bool
}

// Messages
type ScreenChangeMsg struct {
	Screen Screen
}

type ErrorMsg struct {
	Err error
}

type SessionReadyMsg struct {
	Session *session.Session
}

// Initialize the TUI application
func NewApp(ctx context.Context) *App {
	app := &App{
		ctx:           ctx,
		currentScreen: WelcomeScreen,
		screens:       make(map[Screen]tea.Model),
		exportConfig:  &ExportConfig{Format: "eml"},
		filterConfig:  &FilterConfig{},
		userSettings:  &UserSettings{
			Theme:         "dark",
			AutoSave:      true,
			DefaultFormat: "eml",
			ProgressStyle: "enhanced",
			Notifications: true,
		},
	}
	
	// Initialize all screens
	app.initScreens()
	
	return app
}

func (a *App) initScreens() {
	a.screens[WelcomeScreen] = NewWelcomeModel(a)
	a.screens[LoginScreen] = NewLoginModel(a)
	a.screens[MainMenuScreen] = NewMainMenuModel(a)
	a.screens[FilterScreen] = NewFilterModel(a)
	a.screens[ExportScreen] = NewExportModel(a)
	a.screens[ProgressScreen] = NewProgressModel(a)
	a.screens[ResultsScreen] = NewResultsModel(a)
	a.screens[SettingsScreen] = NewSettingsModel(a)
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		a.screens[a.currentScreen].Init(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update all screens with new size
		for screen, model := range a.screens {
			newModel, newCmd := model.Update(msg)
			a.screens[screen] = newModel
			if newCmd != nil {
				cmds = append(cmds, newCmd)
			}
		}
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			a.quitting = true
			return a, tea.Quit
		case "f1":
			// Global help
			return a, nil
		}
		
	case ScreenChangeMsg:
		a.currentScreen = msg.Screen
		return a, a.screens[a.currentScreen].Init()
		
	case ErrorMsg:
		a.err = msg.Err
		// Could switch to error screen or show error overlay
		
	case SessionReadyMsg:
		a.session = msg.Session
	}
	
	// Update current screen
	currentModel, currentCmd := a.screens[a.currentScreen].Update(msg)
	a.screens[a.currentScreen] = currentModel
	if currentCmd != nil {
		cmds = append(cmds, currentCmd)
	}
	
	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.quitting {
		return ""
	}
	
	// Main layout with header and footer
	header := a.renderHeader()
	content := a.screens[a.currentScreen].View()
	footer := a.renderFooter()
	
	// Apply dark theme styling
	style := lipgloss.NewStyle().
		Background(DarkBg).
		Foreground(DarkFg).
		Width(a.width).
		Height(a.height)
	
	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
	
	return style.Render(layout)
}

func (a *App) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Background(DarkBg).
		Padding(0, 1).
		Render("📧 Proton Mail Export Tool")
	
	version := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Render(fmt.Sprintf("v%s", internal.ETVersionString))
	
	// Screen indicator
	screenName := a.getScreenName()
	screen := lipgloss.NewStyle().
		Foreground(ProtonPink).
		Render(screenName)
	
	headerStyle := lipgloss.NewStyle().
		Background(DarkBg).
		Foreground(DarkFg).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(DarkBorder).
		Padding(0, 1).
		Width(a.width)
	
	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(4).Render(""),
		screen,
		lipgloss.NewStyle().Width(a.width-lipgloss.Width(title)-lipgloss.Width(screen)-lipgloss.Width(version)-8).Render(""),
		version,
	)
	
	return headerStyle.Render(headerContent)
}

func (a *App) renderFooter() string {
	var shortcuts []string
	
	switch a.currentScreen {
	case WelcomeScreen:
		shortcuts = []string{"enter: continue", "q: quit"}
	case LoginScreen:
		shortcuts = []string{"tab: next field", "enter: login", "esc: back", "q: quit"}
	case MainMenuScreen:
		shortcuts = []string{"↑/↓: navigate", "enter: select", "s: settings", "q: quit"}
	case FilterScreen:
		shortcuts = []string{"tab: next field", "enter: apply", "r: reset", "esc: back"}
	case ExportScreen:
		shortcuts = []string{"tab: next field", "enter: start export", "esc: back"}
	case ProgressScreen:
		shortcuts = []string{"q: cancel", "p: pause/resume"}
	default:
		shortcuts = []string{"F1: help", "q: quit"}
	}
	
	shortcutText := ""
	for i, shortcut := range shortcuts {
		if i > 0 {
			shortcutText += " • "
		}
		shortcutText += shortcut
	}
	
	footerStyle := lipgloss.NewStyle().
		Background(DarkBg).
		Foreground(ProtonGray).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(DarkBorder).
		Padding(0, 1).
		Width(a.width)
	
	return footerStyle.Render(shortcutText)
}

func (a *App) getScreenName() string {
	switch a.currentScreen {
	case WelcomeScreen:
		return "Welcome"
	case LoginScreen:
		return "Login"
	case MainMenuScreen:
		return "Main Menu"
	case FilterScreen:
		return "Filters"
	case ExportScreen:
		return "Export"
	case ProgressScreen:
		return "Progress"
	case ResultsScreen:
		return "Results"
	case SettingsScreen:
		return "Settings"
	default:
		return "Unknown"
	}
}

// Helper methods for screen navigation
func (a *App) SwitchToScreen(screen Screen) tea.Cmd {
	return func() tea.Msg {
		return ScreenChangeMsg{Screen: screen}
	}
}

func (a *App) ShowError(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{Err: err}
	}
}

// Run the TUI application
func Run(ctx context.Context) error {
	// Setup logging
	logDir := filepath.Join(os.TempDir(), "proton-mail-export", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	app := NewApp(ctx)
	program := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	_, err := program.Run()
	return err
}
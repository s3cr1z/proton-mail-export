package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"fmt"
	"strings"
)

// Note: This file may require workspace configuration. Run 'go mod tidy' or add a go.work file to resolve import errors.

// Proton color constants
const protonPink = "#6d4dfb"   // Pink OCD for accents
const protonPurple = "#4226a2" // Galactic Purple for primary text and buttons
const protonLilac = "#bca4fc"  // Winterspring Lilac for secondary elements

// Base dark mode style
var darkModeBase = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a1a")).Foreground(lipgloss.Color("#ffffff")) // Dark background with white text

// TUI State Enum
type ScreenType int

const (
	LoginScreen ScreenType = iota
	OperationScreen
	ProgressScreen
	FilterScreen
	PluginScreen
)

type Model struct {
	currentScreen ScreenType
	screens       map[ScreenType]tea.Model
	metrics       ProgressMetrics
	error         string
	cancelled     bool
}

func InitialModel() Model {
	m := Model{
		currentScreen: LoginScreen,
		screens:       make(map[ScreenType]tea.Model),
	}
	
	// Initialize sub-models
	m.screens[LoginScreen] = NewLoginModel()
	m.screens[OperationScreen] = NewOperationModel()
	m.screens[ProgressScreen] = NewProgressModel()
	m.screens[FilterScreen] = NewFilterModel()
	m.screens[PluginScreen] = NewPluginModel()
	
	return m
}

func (m Model) Init() tea.Cmd {
	// Delegate init to the current screen model
	return m.screens[m.currentScreen].Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmdList []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "tab": // Example: Switch screens or navigate
			// Handle navigation logic here, e.g., cycle through screens
			m.currentScreen = (m.currentScreen + 1) % 5 // Cycle through screens for demo
			return m, nil
		}
	case tea.WindowSizeMsg:
		// Propagate size change to all screens if needed
		for key, screen := range m.screens {
			var c tea.Cmd
			m.screens[key], c = screen.Update(msg)
			cmdList = append(cmdList, c)
		}
		return m, tea.Batch(cmdList...)
	}
	
	// Delegate update to the current screen
	var newModel tea.Model
	newModel, cmd = m.screens[m.currentScreen].Update(msg)
	m.screens[m.currentScreen] = newModel
	
	return m, cmd
}

func (m Model) View() string {
	// Render the current screen with dark mode styling
	view := m.screens[m.currentScreen].View()
	return darkModeBase.Render(view)
}

// Enhanced placeholder models with basic implementations and Proton branding

type LoginModel struct {
    username string
    password string
    focused  bool // For focusing on input fields
}

func (m LoginModel) Init() tea.Cmd { return nil }
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.focused = !m.focused // Toggle focus between fields
            return m, nil
        case "enter":
            if m.focused {
                // Simulate password entry or login action
                return m, nil
            } else {
                // Simulate username entry
                return m, nil
            }
        }
    }
    return m, nil
}
func (m LoginModel) View() string {
    doc := strings.Builder{}
    if m.focused {
        doc.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPink)).Render("Focused on Password"))
    } else {
        doc.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple)).Render("Focused on Username"))
    }
    return darkModeBase.Render(fmt.Sprintf("Login Screen\nUsername: %s\nPassword: ********", m.username))
}

type OperationModel struct {
    selected int // 0 for backup, 1 for restore, etc.
}

func (m OperationModel) Init() tea.Cmd { return nil }
func (m OperationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up":
            if m.selected > 0 {
                m.selected--
            }
            return m, nil
        case "down":
            m.selected++ // Assume bounds checking
            return m, nil
        case "enter":
            // Handle selection (e.g., switch to progress screen)
            return m, nil
        }
    }
    return m, nil
}
func (m OperationModel) View() string {
    operations := []string{"Backup", "Restore"}
    var b strings.Builder
    for i, op := range operations {
        if i == m.selected {
            b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPink)).Render(fmt.Sprintf("-> %s", op)) + "\n")
        } else {
            b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple)).Render(fmt.Sprintf("   %s", op)) + "\n")
        }
    }
    return darkModeBase.Render(b.String())
}

type ProgressModel struct {
    metrics ProgressMetrics
}
func (m ProgressModel) Init() tea.Cmd { return nil }
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m ProgressModel) View() string {
    // Use existing progress rendering with Proton colors
    return darkModeBase.Copy().Foreground(lipgloss.Color(protonLilac)).Render(fmt.Sprintf("Progress: %.1f%%", m.metrics.GetProgressPercent()))
}

type FilterModel struct {
    filters map[string]string // Key-value for filters, e.g., "dateStart": "2023-01-01"
}

func (m FilterModel) Init() tea.Cmd { return nil }
func (m FilterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m FilterModel) View() string {
    return darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple)).Render("Filter Configuration Screen\nAdd filters here...")
}

type PluginModel struct {
    plugins []string // List of available plugins
}

func (m PluginModel) Init() tea.Cmd { return nil }
func (m PluginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m PluginModel) View() string {
    return darkModeBase.Copy().Foreground(lipgloss.Color(protonPink)).Render("Plugin Management Screen\nSelect exporter...")
}

// Factory functions
func NewLoginModel() tea.Model { return LoginModel{} }
func NewOperationModel() tea.Model { return OperationModel{} }
func NewProgressModel() tea.Model { return ProgressModel{} }
func NewFilterModel() tea.Model { return FilterModel{} }
func NewPluginModel() tea.Model { return PluginModel{} }

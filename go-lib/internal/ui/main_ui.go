package ui

import (
        "context"
        "fmt"
        "os"
        "path/filepath"
        "strings"
        "time"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/huh"
        "github.com/charmbracelet/lipgloss"
        "github.com/ProtonMail/export-tool/internal/session"
)

// Proton color constants
const (
        protonPink   = "#6d4dfb"   // Pink OCD for accents
        protonPurple = "#4226a2"   // Galactic Purple for primary text and buttons
        protonLilac  = "#bca4fc"   // Winterspring Lilac for secondary elements
        protonGray   = "#706d6b"   // Gray for secondary text
        protonWhite  = "#ffffff"   // White for primary text
        protonDark   = "#1a1a1a"   // Dark background
)

// Styles
var (
        baseStyle = lipgloss.NewStyle().
                        Background(lipgloss.Color(protonDark)).
                        Foreground(lipgloss.Color(protonWhite))

        titleStyle = lipgloss.NewStyle().
                        Foreground(lipgloss.Color(protonPink)).
                        Bold(true).
                        Margin(1, 0)

        subtitleStyle = lipgloss.NewStyle().
                        Foreground(lipgloss.Color(protonLilac)).
                        Margin(0, 0, 1, 0)

        errorStyle = lipgloss.NewStyle().
                        Foreground(lipgloss.Color("#ff6b6b")).
                        Bold(true)

        successStyle = lipgloss.NewStyle().
                        Foreground(lipgloss.Color("#51cf66")).
                        Bold(true)

        focusedStyle = lipgloss.NewStyle().
                        Foreground(lipgloss.Color(protonPink)).
                        Bold(true)

        blurredStyle = lipgloss.NewStyle().
                        Foreground(lipgloss.Color(protonGray))
)

// Screen types
type ScreenType int

const (
        WelcomeScreen ScreenType = iota
        LoginScreen
        OperationScreen
        PathScreen
        ProgressScreen
        CompletionScreen
        ErrorScreen
)

// Messages for communication between screens
type ScreenChangeMsg struct {
        Screen ScreenType
        Data   interface{}
}

type LoginCompleteMsg struct {
        Session *session.Session
}

type OperationSelectedMsg struct {
        Operation string
}

type PathSelectedMsg struct {
        Path string
}

type ProgressUpdateMsg struct {
        Progress float64
        Status   string
}

type ErrorMsg struct {
        Error error
}

type CompletionMsg struct {
        Success bool
        Message string
}

// Main TUI Model
type Model struct {
        currentScreen ScreenType
        screens       map[ScreenType]tea.Model
        session       *session.Session
        operation     string
        path          string
        width         int
        height        int
        error         error
        cancelled     bool
}

func NewModel() Model {
        m := Model{
                currentScreen: WelcomeScreen,
                screens:       make(map[ScreenType]tea.Model),
        }

        // Initialize all screen models
        m.screens[WelcomeScreen] = NewWelcomeModel()
        m.screens[LoginScreen] = NewLoginModel()
        m.screens[OperationScreen] = NewOperationModel()
        m.screens[PathScreen] = NewPathModel()
        m.screens[ProgressScreen] = NewProgressScreenModel()
        m.screens[CompletionScreen] = NewCompletionModel()
        m.screens[ErrorScreen] = NewErrorModel()

        return m
}

func (m Model) Init() tea.Cmd {
        return m.screens[m.currentScreen].Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        var cmd tea.Cmd
        var cmds []tea.Cmd

        switch msg := msg.(type) {
        case tea.KeyMsg:
                switch msg.String() {
                case "ctrl+c":
                        m.cancelled = true
                        return m, tea.Quit
                }

        case tea.WindowSizeMsg:
                m.width = msg.Width
                m.height = msg.Height
                // Propagate size to all screens
                for screenType, screen := range m.screens {
                        var c tea.Cmd
                        m.screens[screenType], c = screen.Update(msg)
                        if c != nil {
                                cmds = append(cmds, c)
                        }
                }

        case ScreenChangeMsg:
                m.currentScreen = msg.Screen
                if msg.Data != nil {
                        // Pass data to the new screen
                        var c tea.Cmd
                        m.screens[m.currentScreen], c = m.screens[m.currentScreen].Update(msg)
                        if c != nil {
                                cmds = append(cmds, c)
                        }
                }
                return m, tea.Batch(cmds...)

        case LoginCompleteMsg:
                m.session = msg.Session
                m.currentScreen = OperationScreen
                return m, nil

        case OperationSelectedMsg:
                m.operation = msg.Operation
                m.currentScreen = PathScreen
                // Pass operation to path screen
                var c tea.Cmd
                m.screens[PathScreen], c = m.screens[PathScreen].Update(msg)
                if c != nil {
                        cmds = append(cmds, c)
                }
                return m, tea.Batch(cmds...)

        case PathSelectedMsg:
                m.path = msg.Path
                m.currentScreen = ProgressScreen
                // Start the operation
                var c tea.Cmd
                m.screens[ProgressScreen], c = m.screens[ProgressScreen].Update(msg)
                if c != nil {
                        cmds = append(cmds, c)
                }
                return m, tea.Batch(cmds...)

        case ErrorMsg:
                m.error = msg.Error
                m.currentScreen = ErrorScreen
                var c tea.Cmd
                m.screens[ErrorScreen], c = m.screens[ErrorScreen].Update(msg)
                if c != nil {
                        cmds = append(cmds, c)
                }
                return m, tea.Batch(cmds...)

        case CompletionMsg:
                m.currentScreen = CompletionScreen
                var c tea.Cmd
                m.screens[CompletionScreen], c = m.screens[CompletionScreen].Update(msg)
                if c != nil {
                        cmds = append(cmds, c)
                }
                return m, tea.Batch(cmds...)
        }

        // Update current screen
        var newModel tea.Model
        newModel, cmd = m.screens[m.currentScreen].Update(msg)
        m.screens[m.currentScreen] = newModel
        if cmd != nil {
                cmds = append(cmds, cmd)
        }

        return m, tea.Batch(cmds...)
}

func (m Model) View() string {
        if m.width == 0 || m.height == 0 {
                return "Loading..."
        }

        content := m.screens[m.currentScreen].View()
        
        // Add header with Proton branding
        header := titleStyle.Render("Proton Mail Export Tool") + "\n" +
                subtitleStyle.Render("Secure email backup and restore")

        // Combine header and content
        view := lipgloss.JoinVertical(lipgloss.Left, header, content)

        // Apply base styling and center
        return baseStyle.
                Width(m.width).
                Height(m.height).
                Align(lipgloss.Center, lipgloss.Center).
                Render(view)
}

func (m Model) IsCancelled() bool {
        return m.cancelled
}

// Welcome Screen
type WelcomeModel struct {
        ready bool
}

func NewWelcomeModel() WelcomeModel {
        return WelcomeModel{}
}

func (m WelcomeModel) Init() tea.Cmd {
        return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
                return ScreenChangeMsg{Screen: LoginScreen}
        })
}

func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.KeyMsg:
                if msg.String() == "enter" || msg.String() == " " {
                        return m, func() tea.Msg {
                                return ScreenChangeMsg{Screen: LoginScreen}
                        }
                }
        }
        return m, nil
}

func (m WelcomeModel) View() string {
        welcome := `
Welcome to Proton Mail Export Tool

This tool allows you to:
• Backup your Proton Mail data
• Restore from previous backups
• Export in multiple formats

Press ENTER to continue or wait 2 seconds...
`
        return lipgloss.NewStyle().
                Margin(2, 0).
                Padding(2).
                Border(lipgloss.RoundedBorder()).
                BorderForeground(lipgloss.Color(protonPink)).
                Render(welcome)
}
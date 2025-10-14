package models

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal"
	"github.com/ProtonMail/export-tool/internal/session"
	"github.com/ProtonMail/export-tool/internal/tui/styles"
)

// Screen represents different application screens
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenAuth
	ScreenOperation
	ScreenConfig
	ScreenProgress
	ScreenComplete
)

// AppModel is the main application model that coordinates between screens
type AppModel struct {
	ctx           context.Context
	currentScreen Screen
	width         int
	height        int
	session       *session.Session
	
	// Screen models
	welcomeModel    WelcomeModel
	authModel       AuthModel
	operationModel  OperationModel
	configModel     ConfigModel
	progressModel   ProgressModel
	completeModel   CompleteModel
	
	// Shared state
	selectedOperation string
	exportPath        string
	err               error
}

// NewAppModel creates a new application model
func NewAppModel(ctx context.Context) AppModel {
	return AppModel{
		ctx:           ctx,
		currentScreen: ScreenWelcome,
		welcomeModel:  NewWelcomeModel(),
		authModel:     NewAuthModel(),
		operationModel: NewOperationModel(),
		configModel:   NewConfigModel(),
		progressModel: NewProgressModel(),
		completeModel: NewCompleteModel(),
	}
}

// Init implements tea.Model
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.welcomeModel.Init(),
	)
}

// Update implements tea.Model
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update all screen models with new dimensions
		m.welcomeModel.SetSize(msg.Width, msg.Height)
		m.authModel.SetSize(msg.Width, msg.Height)
		m.operationModel.SetSize(msg.Width, msg.Height)
		m.configModel.SetSize(msg.Width, msg.Height)
		m.progressModel.SetSize(msg.Width, msg.Height)
		m.completeModel.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentScreen == ScreenWelcome {
				return m, tea.Quit
			}
		case "esc":
			// Navigate back to previous screen
			return m.navigateBack()
		}

	case NavigateMsg:
		return m.handleNavigation(msg)
		
	case ErrorMsg:
		m.err = msg.Error
		// Stay on current screen and display error
		
	case SessionCreatedMsg:
		m.session = msg.Session
		m.currentScreen = ScreenOperation
		return m, m.operationModel.Init()
	}

	// Update current screen model
	switch m.currentScreen {
	case ScreenWelcome:
		m.welcomeModel, cmd = m.welcomeModel.Update(msg)
	case ScreenAuth:
		m.authModel, cmd = m.authModel.Update(msg)
	case ScreenOperation:
		m.operationModel, cmd = m.operationModel.Update(msg)
	case ScreenConfig:
		m.configModel, cmd = m.configModel.Update(msg)
	case ScreenProgress:
		m.progressModel, cmd = m.progressModel.Update(msg)
	case ScreenComplete:
		m.completeModel, cmd = m.completeModel.Update(msg)
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var content string
	
	// Render current screen
	switch m.currentScreen {
	case ScreenWelcome:
		content = m.welcomeModel.View()
	case ScreenAuth:
		content = m.authModel.View()
	case ScreenOperation:
		content = m.operationModel.View()
	case ScreenConfig:
		content = m.configModel.View()
	case ScreenProgress:
		content = m.progressModel.View()
	case ScreenComplete:
		content = m.completeModel.View()
	}

	// Add error display if present
	if m.err != nil {
		errorMsg := styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", errorMsg)
	}

	// Wrap in app container
	return styles.AppStyle.
		Width(m.width - 4).
		Height(m.height - 2).
		Render(content)
}

// handleNavigation processes navigation messages
func (m AppModel) handleNavigation(msg NavigateMsg) (AppModel, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.Screen {
	case ScreenAuth:
		m.currentScreen = ScreenAuth
		cmd = m.authModel.Init()
	case ScreenOperation:
		m.currentScreen = ScreenOperation
		cmd = m.operationModel.Init()
	case ScreenConfig:
		m.selectedOperation = msg.Data["operation"].(string)
		m.configModel.SetOperation(m.selectedOperation)
		m.currentScreen = ScreenConfig
		cmd = m.configModel.Init()
	case ScreenProgress:
		m.exportPath = msg.Data["path"].(string)
		m.progressModel.SetConfig(m.selectedOperation, m.exportPath, m.session)
		m.currentScreen = ScreenProgress
		cmd = m.progressModel.Init()
	case ScreenComplete:
		m.currentScreen = ScreenComplete
		cmd = m.completeModel.Init()
	}
	
	// Clear any previous errors when navigating
	m.err = nil
	
	return m, cmd
}

// navigateBack handles back navigation
func (m AppModel) navigateBack() (AppModel, tea.Cmd) {
	switch m.currentScreen {
	case ScreenAuth:
		m.currentScreen = ScreenWelcome
		return m, m.welcomeModel.Init()
	case ScreenOperation:
		m.currentScreen = ScreenAuth
		return m, m.authModel.Init()
	case ScreenConfig:
		m.currentScreen = ScreenOperation
		return m, m.operationModel.Init()
	case ScreenProgress:
		// Don't allow back navigation during progress
		return m, nil
	case ScreenComplete:
		m.currentScreen = ScreenWelcome
		// Reset state for new session
		m.session = nil
		m.selectedOperation = ""
		m.exportPath = ""
		return m, m.welcomeModel.Init()
	default:
		return m, nil
	}
}

// Messages for communication between models

// NavigateMsg is sent to navigate to a different screen
type NavigateMsg struct {
	Screen Screen
	Data   map[string]interface{}
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Error error
}

// SessionCreatedMsg is sent when authentication is successful
type SessionCreatedMsg struct {
	Session *session.Session
}

// OperationSelectedMsg is sent when an operation is selected
type OperationSelectedMsg struct {
	Operation string
}

// ConfigCompleteMsg is sent when configuration is complete
type ConfigCompleteMsg struct {
	Path string
}

// ProgressCompleteMsg is sent when an operation completes
type ProgressCompleteMsg struct {
	Success bool
	Message string
}
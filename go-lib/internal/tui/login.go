package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal/apiclient"
	"github.com/ProtonMail/export-tool/internal/session"
)

type LoginModel struct {
	app           *App
	form          *huh.Form
	width         int
	height        int
	loginState    session.LoginState
	session       *session.Session
	loading       bool
	error         string
	
	// Form fields
	username      string
	password      string
	totpCode      string
	mboxPassword  string
	hvCompleted   bool
}

type LoginStartMsg struct{}
type LoginSuccessMsg struct {
	Session *session.Session
}
type LoginErrorMsg struct {
	Error string
}
type LoginStateChangeMsg struct {
	State session.LoginState
}

func NewLoginModel(app *App) *LoginModel {
	m := &LoginModel{
		app:        app,
		loginState: session.LoginStateLoggedOut,
	}
	
	m.initForm()
	return m
}

func (m *LoginModel) initForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Email Address").
				Placeholder("your.email@proton.me").
				Value(&m.username).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("email address is required")
					}
					if !strings.Contains(s, "@") {
						return fmt.Errorf("please enter a valid email address")
					}
					return nil
				}),
			
			huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&m.password).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("password is required")
					}
					return nil
				}),
		).Title("Login to Proton Mail").
			Description("Enter your Proton Mail credentials to continue"),
	).WithTheme(huh.ThemeDracula())
}

func (m *LoginModel) initTOTPForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Two-Factor Authentication Code").
				Placeholder("123456").
				Value(&m.totpCode).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("2FA code is required")
					}
					if len(s) != 6 {
						return fmt.Errorf("2FA code must be 6 digits")
					}
					return nil
				}),
		).Title("Two-Factor Authentication").
			Description("Enter the 6-digit code from your authenticator app"),
	).WithTheme(huh.ThemeDracula())
}

func (m *LoginModel) initMailboxPasswordForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Mailbox Password").
				EchoMode(huh.EchoModePassword).
				Value(&m.mboxPassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("mailbox password is required")
					}
					return nil
				}),
		).Title("Mailbox Password Required").
			Description("Enter your mailbox password (Two-Password Mode)"),
	).WithTheme(huh.ThemeDracula())
}

func (m *LoginModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if !m.loading {
				return m, m.app.SwitchToScreen(WelcomeScreen)
			}
		case "ctrl+c":
			return m, tea.Quit
		}
		
	case LoginStartMsg:
		m.loading = true
		m.error = ""
		return m, m.performLogin()
		
	case LoginSuccessMsg:
		m.loading = false
		m.session = msg.Session
		return m, tea.Batch(
			m.app.SwitchToScreen(MainMenuScreen),
			func() tea.Msg {
				return SessionReadyMsg{Session: msg.Session}
			},
		)
		
	case LoginErrorMsg:
		m.loading = false
		m.error = msg.Error
		
	case LoginStateChangeMsg:
		m.loginState = msg.State
		switch msg.State {
		case session.LoginStateAwaitingTOTP:
			m.initTOTPForm()
			return m, m.form.Init()
		case session.LoginStateAwaitingMailboxPassword:
			m.initMailboxPasswordForm()
			return m, m.form.Init()
		case session.LoginStateAwaitingHV:
			// Handle human verification
			return m, m.handleHumanVerification()
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
		if m.form.State == huh.StateCompleted && !m.loading {
			cmds = append(cmds, func() tea.Msg {
				return LoginStartMsg{}
			})
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *LoginModel) View() string {
	var content strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ProtonPurple).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	content.WriteString(titleStyle.Render("🔐 Secure Login"))
	content.WriteString("\n\n")
	
	// Show current state
	if m.loading {
		spinner := lipgloss.NewStyle().
			Foreground(ProtonBlue).
			Render("⏳ Authenticating...")
		content.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Render(spinner))
		content.WriteString("\n\n")
	}
	
	// Show error if any
	if m.error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(ProtonRed).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ProtonRed).
			Padding(1).
			Margin(1, 0)
		
		content.WriteString(errorStyle.Render("❌ " + m.error))
		content.WriteString("\n\n")
	}
	
	// Show human verification message if needed
	if m.loginState == session.LoginStateAwaitingHV {
		hvStyle := lipgloss.NewStyle().
			Foreground(ProtonBlue).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ProtonBlue).
			Padding(1).
			Margin(1, 0)
		
		hvText := `🤖 Human Verification Required

Please complete the human verification challenge in your web browser.
The verification URL has been opened automatically.

Press ENTER when you have completed the verification.`
		
		content.WriteString(hvStyle.Render(hvText))
		content.WriteString("\n\n")
		
		// Show continue button
		continueStyle := lipgloss.NewStyle().
			Foreground(ProtonGreen).
			Bold(true).
			Align(lipgloss.Center)
		
		content.WriteString(continueStyle.Render("Press ENTER to continue after completing verification"))
		
	} else {
		// Show form
		formView := m.form.View()
		content.WriteString(formView)
	}
	
	// Instructions
	if !m.loading && m.loginState != session.LoginStateAwaitingHV {
		instructionStyle := lipgloss.NewStyle().
			Foreground(ProtonGray).
			Italic(true).
			Align(lipgloss.Center).
			Margin(2, 0)
		
		var instructions string
		switch m.loginState {
		case session.LoginStateLoggedOut:
			instructions = "Fill in your credentials and press TAB to navigate between fields"
		case session.LoginStateAwaitingTOTP:
			instructions = "Enter the 6-digit code from your authenticator app"
		case session.LoginStateAwaitingMailboxPassword:
			instructions = "Enter your mailbox password for Two-Password Mode"
		}
		
		content.WriteString(instructionStyle.Render(instructions))
	}
	
	// Container styling
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 4). // Account for header/footer
		Align(lipgloss.Center, lipgloss.Center).
		Padding(2)
	
	return containerStyle.Render(content.String())
}

func (m *LoginModel) performLogin() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		
		// Create session if not exists
		if m.session == nil {
			sessionCb := &SessionCallback{}
			builder, err := apiclient.NewProtonAPIClientBuilder(
				m.app.getAPIURL(),
				nil, // panic handler
				sessionCb,
			)
			if err != nil {
				return LoginErrorMsg{Error: fmt.Sprintf("Failed to create API client: %v", err)}
			}
			
			clientBuilder := apiclient.NewAutoRetryClientBuilder(
				builder,
				&apiclient.SleepRetryStrategyBuilder{},
			)
			
			m.session = session.NewSession(clientBuilder, sessionCb, nil, nil, false)
		}
		
		// Perform login based on current state
		switch m.loginState {
		case session.LoginStateLoggedOut:
			if err := m.session.Login(ctx, m.username, m.password); err != nil {
				return LoginErrorMsg{Error: fmt.Sprintf("Login failed: %v", err)}
			}
			
		case session.LoginStateAwaitingTOTP:
			if err := m.session.SubmitTOTP(ctx, m.totpCode); err != nil {
				return LoginErrorMsg{Error: fmt.Sprintf("2FA verification failed: %v", err)}
			}
			
		case session.LoginStateAwaitingMailboxPassword:
			validator := apiclient.NewProtonMailboxPasswordValidator(
				m.session.GetUser(),
				m.session.GetUserSalts(),
			)
			if err := m.session.SubmitMailboxPassword(validator, m.mboxPassword); err != nil {
				return LoginErrorMsg{Error: fmt.Sprintf("Mailbox password verification failed: %v", err)}
			}
		}
		
		// Check new login state
		newState := m.session.LoginState()
		if newState != m.loginState {
			return LoginStateChangeMsg{State: newState}
		}
		
		if newState == session.LoginStateLoggedIn {
			return LoginSuccessMsg{Session: m.session}
		}
		
		return nil
	}
}

func (m *LoginModel) handleHumanVerification() tea.Cmd {
	return func() tea.Msg {
		// Get HV URL and open it
		url, err := m.session.GetHVSolveURL()
		if err != nil {
			return LoginErrorMsg{Error: fmt.Sprintf("Failed to get verification URL: %v", err)}
		}
		
		// Try to open URL in browser (platform-specific)
		// For now, just return the URL for manual opening
		fmt.Printf("\nPlease open this URL in your browser:\n%s\n\n", url)
		
		return nil
	}
}

// Session callback implementation
type SessionCallback struct{}

func (s *SessionCallback) OnNetworkLost() {
	// Handle network loss
}

func (s *SessionCallback) OnNetworkRestored() {
	// Handle network restoration
}

// Helper method to get API URL
func (a *App) getAPIURL() string {
	// This should be moved to a config or environment variable
	return "https://mail-api.proton.me"
}
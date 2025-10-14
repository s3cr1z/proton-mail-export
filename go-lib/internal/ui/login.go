package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal/session"
)

// Login states
type LoginState int

const (
	LoginStateUsername LoginState = iota
	LoginStatePassword
	LoginStateTOTP
	LoginStateMailboxPassword
	LoginStateHV
	LoginStateProcessing
	LoginStateComplete
	LoginStateError
)

// Login Model
type LoginModel struct {
	state           LoginState
	username        string
	password        string
	totp            string
	mailboxPassword string
	hvURL           string
	error           string
	form            *huh.Form
	session         *session.Session
	attempts        int
	maxAttempts     int
	width           int
	height          int
}

func NewLoginModel() LoginModel {
	return LoginModel{
		state:       LoginStateUsername,
		maxAttempts: 3,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return m.createForm()
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.state != LoginStateUsername {
				// Go back to previous state or username
				m.state = LoginStateUsername
				m.error = ""
				return m, m.createForm()
			}
		}

	case huh.FormCompleteMsg:
		return m, m.processLogin()
	}

	// Update form if it exists
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		return m, cmd
	}

	return m, nil
}

func (m LoginModel) View() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Login to Proton Mail"))
	content.WriteString("\n\n")

	// Show current state
	switch m.state {
	case LoginStateUsername, LoginStatePassword:
		content.WriteString("Enter your Proton Mail credentials:\n\n")
	case LoginStateTOTP:
		content.WriteString("Two-Factor Authentication required:\n\n")
	case LoginStateMailboxPassword:
		content.WriteString("Mailbox password required:\n\n")
	case LoginStateHV:
		content.WriteString("Human Verification required:\n\n")
		content.WriteString(fmt.Sprintf("Please open this URL in your browser:\n%s\n\n", m.hvURL))
		content.WriteString("Press ENTER when completed...\n")
	case LoginStateProcessing:
		content.WriteString("Logging in...\n")
	case LoginStateError:
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.error)))
		content.WriteString("\n\nPress ESC to try again or Ctrl+C to quit\n")
	}

	// Show form if available
	if m.form != nil {
		content.WriteString(m.form.View())
	}

	// Show attempts counter
	if m.attempts > 0 {
		content.WriteString(fmt.Sprintf("\nAttempts: %d/%d", m.attempts, m.maxAttempts))
	}

	// Instructions
	content.WriteString("\n\n")
	content.WriteString(blurredStyle.Render("Press ESC to go back • Ctrl+C to quit"))

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2).
		Render(content.String())
}

func (m LoginModel) createForm() tea.Cmd {
	var inputs []huh.Field

	switch m.state {
	case LoginStateUsername, LoginStatePassword:
		inputs = []huh.Field{
			huh.NewInput().
				Key("username").
				Title("Username").
				Value(&m.username).
				Placeholder("your-email@proton.me"),
			huh.NewInput().
				Key("password").
				Title("Password").
				Value(&m.password).
				EchoMode(huh.EchoModePassword).
				Placeholder("Enter your password"),
		}

	case LoginStateTOTP:
		inputs = []huh.Field{
			huh.NewInput().
				Key("totp").
				Title("2FA Code").
				Value(&m.totp).
				Placeholder("Enter 6-digit code"),
		}

	case LoginStateMailboxPassword:
		inputs = []huh.Field{
			huh.NewInput().
				Key("mailbox_password").
				Title("Mailbox Password").
				Value(&m.mailboxPassword).
				EchoMode(huh.EchoModePassword).
				Placeholder("Enter mailbox password"),
		}

	case LoginStateHV:
		inputs = []huh.Field{
			huh.NewConfirm().
				Key("hv_complete").
				Title("Human Verification Complete?").
				Affirmative("Yes").
				Negative("No"),
		}

	default:
		return nil
	}

	m.form = huh.NewForm(
		huh.NewGroup(inputs...),
	).WithTheme(huh.ThemeCharm())

	return m.form.Init()
}

func (m LoginModel) processLogin() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// This would integrate with the actual session login logic
		// For now, simulate the login process
		
		switch m.state {
		case LoginStateUsername, LoginStatePassword:
			if m.username == "" || m.password == "" {
				m.attempts++
				if m.attempts >= m.maxAttempts {
					return ErrorMsg{fmt.Errorf("maximum login attempts exceeded")}
				}
				m.error = "Username and password are required"
				m.state = LoginStateError
				return nil
			}

			// Simulate login attempt
			// In real implementation, this would call session.Login()
			// and handle the different login states
			
			// For demo, simulate successful login
			return LoginCompleteMsg{Session: m.session}

		case LoginStateTOTP:
			if m.totp == "" {
				m.error = "TOTP code is required"
				m.state = LoginStateError
				return nil
			}
			// Process TOTP
			return LoginCompleteMsg{Session: m.session}

		case LoginStateMailboxPassword:
			if m.mailboxPassword == "" {
				m.error = "Mailbox password is required"
				m.state = LoginStateError
				return nil
			}
			// Process mailbox password
			return LoginCompleteMsg{Session: m.session}

		case LoginStateHV:
			// Process HV completion
			return LoginCompleteMsg{Session: m.session}
		}

		return nil
	})
}

// Integration function to connect with existing session logic
func (m *LoginModel) LoginWithSession(ctx context.Context, sess *session.Session) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		m.session = sess
		
		for {
			switch sess.LoginState() {
			case session.LoginStateLoggedOut:
				if m.username == "" || m.password == "" {
					m.state = LoginStateUsername
					return nil
				}
				
				if err := sess.Login(ctx, m.username, m.password); err != nil {
					m.attempts++
					if m.attempts >= m.maxAttempts {
						return ErrorMsg{fmt.Errorf("maximum login attempts exceeded")}
					}
					m.error = err.Error()
					m.state = LoginStateError
					return nil
				}

			case session.LoginStateAwaitingTOTP:
				if m.totp == "" {
					m.state = LoginStateTOTP
					return nil
				}
				
				if err := sess.SubmitTOTP(ctx, m.totp); err != nil {
					m.error = err.Error()
					m.state = LoginStateError
					return nil
				}

			case session.LoginStateAwaitingMailboxPassword:
				if m.mailboxPassword == "" {
					m.state = LoginStateMailboxPassword
					return nil
				}
				
				// This would need the proper validator implementation
				// if err := sess.SubmitMailboxPassword(validator, m.mailboxPassword); err != nil {
				// 	m.error = err.Error()
				// 	m.state = LoginStateError
				// 	return nil
				// }

			case session.LoginStateAwaitingHV:
				url, err := sess.GetHVSolveURL()
				if err != nil {
					return ErrorMsg{err}
				}
				m.hvURL = url
				m.state = LoginStateHV
				return nil

			case session.LoginStateLoggedIn:
				return LoginCompleteMsg{Session: sess}

			default:
				return ErrorMsg{fmt.Errorf("unknown login state: %v", sess.LoginState())}
			}
		}
	})
}
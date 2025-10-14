package models

import (
        "context"
        "fmt"
        "os"
        "strings"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/lipgloss"
        "github.com/ProtonMail/export-tool/internal/apiclient"
        "github.com/ProtonMail/export-tool/internal/session"
        "github.com/ProtonMail/export-tool/internal/tui/styles"
        "github.com/ProtonMail/gluon/async"
)

// AuthStep represents the current authentication step
type AuthStep int

const (
        StepUsername AuthStep = iota
        StepPassword
        StepTOTP
        StepMailboxPassword
        StepHumanVerification
        StepComplete
)

// AuthModel handles the authentication flow
type AuthModel struct {
        width    int
        height   int
        step     AuthStep
        session  *session.Session
        
        // Input fields
        username        string
        password        string
        totp           string
        mailboxPassword string
        
        // UI state
        focused      bool
        inputValue   string
        cursorPos    int
        showPassword bool
        
        // Status
        loading      bool
        errorMsg     string
        hvURL        string
        
        // Context
        ctx context.Context
}

// NewAuthModel creates a new authentication model
func NewAuthModel() AuthModel {
        return AuthModel{
                step:    StepUsername,
                focused: true,
                ctx:     context.Background(),
        }
}

// Init implements tea.Model
func (m AuthModel) Init() tea.Cmd {
        return tea.Batch(
                createSessionCmd(),
                tea.SetCursorMode(tea.CursorBlink),
        )
}

// Update implements tea.Model
func (m AuthModel) Update(msg tea.Msg) (AuthModel, tea.Cmd) {
        var cmd tea.Cmd
        
        switch msg := msg.(type) {
        case tea.KeyMsg:
                if m.loading {
                        return m, nil // Ignore input while loading
                }
                
                switch msg.String() {
                case "ctrl+c":
                        return m, tea.Quit
                case "esc":
                        // Navigate back or clear current input
                        if m.inputValue == "" {
                                return m, func() tea.Msg {
                                        return NavigateMsg{Screen: ScreenWelcome, Data: make(map[string]interface{})}
                                }
                        }
                        m.inputValue = ""
                        m.cursorPos = 0
                case "enter":
                        return m.handleSubmit()
                case "backspace":
                        if m.cursorPos > 0 {
                                m.inputValue = m.inputValue[:m.cursorPos-1] + m.inputValue[m.cursorPos:]
                                m.cursorPos--
                        }
                case "left":
                        if m.cursorPos > 0 {
                                m.cursorPos--
                        }
                case "right":
                        if m.cursorPos < len(m.inputValue) {
                                m.cursorPos++
                        }
                case "home":
                        m.cursorPos = 0
                case "end":
                        m.cursorPos = len(m.inputValue)
                default:
                        // Handle character input
                        if len(msg.String()) == 1 {
                                char := msg.String()
                                m.inputValue = m.inputValue[:m.cursorPos] + char + m.inputValue[m.cursorPos:]
                                m.cursorPos++
                        }
                }
                
        case SessionCreatedInternalMsg:
                m.session = msg.Session
                m.step = StepUsername
                m.loading = false
                
        case AuthProgressMsg:
                m.loading = false
                m.errorMsg = ""
                
                switch msg.State {
                case session.LoginStateAwaitingTOTP:
                        m.step = StepTOTP
                        m.inputValue = ""
                        m.cursorPos = 0
                case session.LoginStateAwaitingMailboxPassword:
                        m.step = StepMailboxPassword
                        m.inputValue = ""
                        m.cursorPos = 0
                case session.LoginStateAwaitingHV:
                        m.step = StepHumanVerification
                        m.hvURL = msg.HVURL
                case session.LoginStateLoggedIn:
                        m.step = StepComplete
                        return m, func() tea.Msg {
                                return SessionCreatedMsg{Session: m.session}
                        }
                }
                
        case AuthErrorMsg:
                m.loading = false
                m.errorMsg = msg.Error.Error()
                // Clear sensitive input on error
                if m.step == StepPassword || m.step == StepMailboxPassword {
                        m.inputValue = ""
                        m.cursorPos = 0
                }
        }
        
        return m, cmd
}

// View implements tea.Model
func (m AuthModel) View() string {
        var sections []string
        
        // Header
        header := m.renderHeader()
        sections = append(sections, header)
        
        // Current step content
        stepContent := m.renderCurrentStep()
        sections = append(sections, stepContent)
        
        // Error message
        if m.errorMsg != "" {
                errorContent := styles.ErrorStyle.Render(fmt.Sprintf("❌ %s", m.errorMsg))
                sections = append(sections, errorContent)
        }
        
        // Loading indicator
        if m.loading {
                loadingContent := styles.ProgressTextStyle.Render("⏳ Processing...")
                sections = append(sections, loadingContent)
        }
        
        // Footer
        footer := m.renderFooter()
        sections = append(sections, footer)
        
        content := lipgloss.JoinVertical(lipgloss.Left, sections...)
        
        return lipgloss.Place(
                m.width-4, m.height-4,
                lipgloss.Center, lipgloss.Center,
                content,
        )
}

// SetSize updates the model dimensions
func (m *AuthModel) SetSize(width, height int) {
        m.width = width
        m.height = height
}

// renderHeader creates the authentication header
func (m AuthModel) renderHeader() string {
        title := styles.TitleStyle.Render("Authentication")
        subtitle := styles.SubtitleStyle.Render("Please provide your Proton account credentials")
        
        headerContent := lipgloss.JoinVertical(
                lipgloss.Center,
                title,
                subtitle,
        )
        
        return styles.HeaderStyle.
                Width(styles.ResponsiveWidth(m.width, 0.8)).
                Render(headerContent)
}

// renderCurrentStep renders the current authentication step
func (m AuthModel) renderCurrentStep() string {
        switch m.step {
        case StepUsername:
                return m.renderUsernameStep()
        case StepPassword:
                return m.renderPasswordStep()
        case StepTOTP:
                return m.renderTOTPStep()
        case StepMailboxPassword:
                return m.renderMailboxPasswordStep()
        case StepHumanVerification:
                return m.renderHVStep()
        default:
                return ""
        }
}

// renderUsernameStep renders the username input step
func (m AuthModel) renderUsernameStep() string {
        label := "Username or Email:"
        input := m.renderInput(m.inputValue, false)
        
        return lipgloss.JoinVertical(
                lipgloss.Left,
                styles.SubtitleStyle.Render(label),
                input,
                "",
                styles.TextMuted.Render("Enter your Proton account username or email address"),
        )
}

// renderPasswordStep renders the password input step
func (m AuthModel) renderPasswordStep() string {
        label := "Password:"
        input := m.renderInput(m.inputValue, true)
        
        return lipgloss.JoinVertical(
                lipgloss.Left,
                styles.SubtitleStyle.Render(label),
                input,
                "",
                styles.TextMuted.Render("Enter your account password"),
        )
}

// renderTOTPStep renders the TOTP input step
func (m AuthModel) renderTOTPStep() string {
        label := "Two-Factor Authentication Code:"
        input := m.renderInput(m.inputValue, false)
        
        return lipgloss.JoinVertical(
                lipgloss.Left,
                styles.SubtitleStyle.Render(label),
                input,
                "",
                styles.TextMuted.Render("Enter the 6-digit code from your authenticator app"),
        )
}

// renderMailboxPasswordStep renders the mailbox password input step
func (m AuthModel) renderMailboxPasswordStep() string {
        label := "Mailbox Password:"
        input := m.renderInput(m.inputValue, true)
        
        return lipgloss.JoinVertical(
                lipgloss.Left,
                styles.SubtitleStyle.Render(label),
                input,
                "",
                styles.TextMuted.Render("Enter your mailbox password (may be different from login password)"),
        )
}

// renderHVStep renders the human verification step
func (m AuthModel) renderHVStep() string {
        title := styles.SubtitleStyle.Render("Human Verification Required")
        instruction := "Please complete the verification challenge in your web browser:"
        url := styles.SuccessStyle.Render(m.hvURL)
        prompt := "Press Enter when you have completed the challenge"
        
        return lipgloss.JoinVertical(
                lipgloss.Left,
                title,
                "",
                instruction,
                url,
                "",
                styles.TextMuted.Render(prompt),
        )
}

// renderInput renders an input field with cursor
func (m AuthModel) renderInput(value string, isPassword bool) string {
        displayValue := value
        if isPassword && len(value) > 0 {
                displayValue = strings.Repeat("•", len(value))
        }
        
        // Add cursor
        if m.focused {
                if m.cursorPos >= len(displayValue) {
                        displayValue += "│"
                } else {
                        displayValue = displayValue[:m.cursorPos] + "│" + displayValue[m.cursorPos:]
                }
        }
        
        style := styles.InputStyle
        if m.focused {
                style = styles.InputFocusedStyle
        }
        
        return style.
                Width(styles.ResponsiveWidth(m.width, 0.6)).
                Render(displayValue)
}

// renderFooter creates the help footer
func (m AuthModel) renderFooter() string {
        var help []string
        
        if m.step == StepHumanVerification {
                help = []string{
                        "Enter: Continue after completing verification",
                        "Esc: Back",
                }
        } else {
                help = []string{
                        "Enter: Submit",
                        "Esc: Back/Clear",
                        "Ctrl+C: Quit",
                }
        }
        
        helpText := strings.Join(help, " • ")
        
        return styles.FooterStyle.
                Width(styles.ResponsiveWidth(m.width, 0.8)).
                Render(helpText)
}

// handleSubmit processes form submission
func (m AuthModel) handleSubmit() (AuthModel, tea.Cmd) {
        if m.session == nil {
                return m, nil
        }
        
        m.loading = true
        m.errorMsg = ""
        
        switch m.step {
        case StepUsername:
                m.username = strings.TrimSpace(m.inputValue)
                if m.username == "" {
                        m.loading = false
                        m.errorMsg = "Username cannot be empty"
                        return m, nil
                }
                m.step = StepPassword
                m.inputValue = ""
                m.cursorPos = 0
                m.loading = false
                return m, nil
                
        case StepPassword:
                m.password = m.inputValue
                if m.password == "" {
                        m.loading = false
                        m.errorMsg = "Password cannot be empty"
                        return m, nil
                }
                return m, loginCmd(m.session, m.username, m.password)
                
        case StepTOTP:
                m.totp = strings.TrimSpace(m.inputValue)
                if m.totp == "" {
                        m.loading = false
                        m.errorMsg = "TOTP code cannot be empty"
                        return m, nil
                }
                return m, submitTOTPCmd(m.session, m.totp)
                
        case StepMailboxPassword:
                m.mailboxPassword = m.inputValue
                if m.mailboxPassword == "" {
                        m.loading = false
                        m.errorMsg = "Mailbox password cannot be empty"
                        return m, nil
                }
                return m, submitMailboxPasswordCmd(m.session, m.mailboxPassword)
                
        case StepHumanVerification:
                return m, markHVSolvedCmd(m.session)
        }
        
        return m, nil
}

// Messages for authentication flow

// SessionCreatedInternalMsg is sent when session is created
type SessionCreatedInternalMsg struct {
        Session *session.Session
}

// AuthProgressMsg is sent during authentication progress
type AuthProgressMsg struct {
        State session.LoginState
        HVURL string
}

// AuthErrorMsg is sent when authentication fails
type AuthErrorMsg struct {
        Error error
}

// Commands for authentication operations

func createSessionCmd() tea.Cmd {
        return func() tea.Msg {
                // Create session similar to app.go newSession function
                panicHandler := func(v interface{}) {} // Simplified for TUI
                sessionCb := TUICallback{}
                
                builder, err := apiclient.NewProtonAPIClientBuilder(
                        getAPIURL(), 
                        async.PanicHandler(panicHandler), 
                        sessionCb,
                )
                if err != nil {
                        return AuthErrorMsg{Error: err}
                }
                
                clientBuilder := apiclient.NewAutoRetryClientBuilder(
                        builder,
                        &apiclient.SleepRetryStrategyBuilder{},
                )
                
                sess := session.NewSession(clientBuilder, sessionCb, async.PanicHandler(panicHandler), nil, false)
                return SessionCreatedInternalMsg{Session: sess}
        }
}

func loginCmd(sess *session.Session, username, password string) tea.Cmd {
        return func() tea.Msg {
                ctx := context.Background()
                if err := sess.Login(ctx, username, password); err != nil {
                        return AuthErrorMsg{Error: err}
                }
                
                state := sess.LoginState()
                return AuthProgressMsg{State: state}
        }
}

func submitTOTPCmd(sess *session.Session, totp string) tea.Cmd {
        return func() tea.Msg {
                ctx := context.Background()
                if err := sess.SubmitTOTP(ctx, totp); err != nil {
                        return AuthErrorMsg{Error: err}
                }
                
                state := sess.LoginState()
                return AuthProgressMsg{State: state}
        }
}

func submitMailboxPasswordCmd(sess *session.Session, password string) tea.Cmd {
        return func() tea.Msg {
                validator := apiclient.NewProtonMailboxPasswordValidator(sess.GetUser(), sess.GetUserSalts())
                if err := sess.SubmitMailboxPassword(validator, password); err != nil {
                        return AuthErrorMsg{Error: err}
                }
                
                state := sess.LoginState()
                return AuthProgressMsg{State: state}
        }
}

func markHVSolvedCmd(sess *session.Session) tea.Cmd {
        return func() tea.Msg {
                ctx := context.Background()
                if err := sess.MarkHVSolved(ctx); err != nil {
                        return AuthErrorMsg{Error: err}
                }
                
                state := sess.LoginState()
                return AuthProgressMsg{State: state}
        }
}

// TUICallback implements session callbacks for TUI
type TUICallback struct{}

func (t TUICallback) OnNetworkRestored() {
        // Could send a message to update UI
}

func (t TUICallback) OnNetworkLost() {
        // Could send a message to update UI
}

// Helper function to get API URL (from app.go)
func getAPIURL() string {
        url := os.Getenv("ET_API_URL")
        if len(url) == 0 {
                // Use default API URL - this should match the CMake generated constant
                url = "https://mail-api.proton.me"
        }
        return url
}
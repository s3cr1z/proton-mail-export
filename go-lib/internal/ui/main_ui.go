package ui

import (
        tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/lipgloss"
        "fmt"
        "strings"
        "time"
        
        "github.com/ProtonMail/export-tool/internal/mail"
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
        filter       *mail.ExportFilter
        currentField int // 0: date start, 1: date end, 2: labels, 3: contacts, 4: domains
        inputMode    bool
        inputBuffer  string
        labels       []string // Available labels for selection
        selectedLabels map[string]bool
        contacts     []string // Contact input list
        domains      []string // Domain input list
        errorMsg     string
}

func (m FilterModel) Init() tea.Cmd { return nil }

func (m FilterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.KeyMsg:
                if m.inputMode {
                        return m.handleInputMode(msg)
                }
                return m.handleNavigationMode(msg)
        }
        return m, nil
}

func (m FilterModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "enter":
                m.inputMode = false
                m.applyInput()
                m.inputBuffer = ""
                return m, nil
        case "esc":
                m.inputMode = false
                m.inputBuffer = ""
                return m, nil
        case "backspace":
                if len(m.inputBuffer) > 0 {
                        m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
                }
                return m, nil
        default:
                if len(msg.String()) == 1 {
                        m.inputBuffer += msg.String()
                }
                return m, nil
        }
}

func (m FilterModel) handleNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
        switch msg.String() {
        case "up":
                if m.currentField > 0 {
                        m.currentField--
                }
                return m, nil
        case "down":
                if m.currentField < 4 {
                        m.currentField++
                }
                return m, nil
        case "enter":
                m.inputMode = true
                return m, nil
        case "c":
                // Clear current filter section
                m.clearCurrentFilter()
                return m, nil
        case "r":
                // Reset all filters
                m.filter = &mail.ExportFilter{}
                return m, nil
        }
        return m, nil
}

func (m *FilterModel) applyInput() {
        if m.filter == nil {
                m.filter = &mail.ExportFilter{}
        }
        
        switch m.currentField {
        case 0: // Start date
                if date, err := time.Parse("2006-01-02", m.inputBuffer); err == nil {
                        if m.filter.DateRange == nil {
                                m.filter.DateRange = &mail.DateRangeFilter{}
                        }
                        m.filter.DateRange.StartDate = &date
                        m.errorMsg = ""
                } else {
                        m.errorMsg = "Invalid date format. Use YYYY-MM-DD"
                }
        case 1: // End date
                if date, err := time.Parse("2006-01-02", m.inputBuffer); err == nil {
                        if m.filter.DateRange == nil {
                                m.filter.DateRange = &mail.DateRangeFilter{}
                        }
                        m.filter.DateRange.EndDate = &date
                        m.errorMsg = ""
                } else {
                        m.errorMsg = "Invalid date format. Use YYYY-MM-DD"
                }
        case 2: // Labels
                if m.inputBuffer != "" {
                        if m.filter.Labels == nil {
                                m.filter.Labels = &mail.LabelFilter{}
                        }
                        m.filter.Labels.IncludeLabels = append(m.filter.Labels.IncludeLabels, m.inputBuffer)
                }
        case 3: // Contacts
                if m.inputBuffer != "" {
                        if m.filter.Contacts == nil {
                                m.filter.Contacts = &mail.ContactFilter{}
                        }
                        m.filter.Contacts.IncludeSenders = append(m.filter.Contacts.IncludeSenders, m.inputBuffer)
                }
        case 4: // Domains
                if m.inputBuffer != "" {
                        if m.filter.Domains == nil {
                                m.filter.Domains = &mail.DomainFilter{}
                        }
                        m.filter.Domains.IncludeDomains = append(m.filter.Domains.IncludeDomains, m.inputBuffer)
                }
        }
}

func (m *FilterModel) clearCurrentFilter() {
        if m.filter == nil {
                return
        }
        
        switch m.currentField {
        case 0, 1: // Date fields
                m.filter.DateRange = nil
        case 2: // Labels
                m.filter.Labels = nil
        case 3: // Contacts
                m.filter.Contacts = nil
        case 4: // Domains
                m.filter.Domains = nil
        }
}

func (m FilterModel) View() string {
        var b strings.Builder
        
        b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple)).Render("📧 Email Export Filters\n\n"))
        
        if m.errorMsg != "" {
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#ff6b6b")).Render("❌ " + m.errorMsg + "\n\n"))
        }
        
        // Date Range Section
        b.WriteString(m.renderDateSection())
        b.WriteString("\n")
        
        // Labels Section
        b.WriteString(m.renderLabelsSection())
        b.WriteString("\n")
        
        // Contacts Section
        b.WriteString(m.renderContactsSection())
        b.WriteString("\n")
        
        // Domains Section
        b.WriteString(m.renderDomainsSection())
        b.WriteString("\n")
        
        // Instructions
        if m.inputMode {
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonLilac)).Render("Enter value (ESC to cancel, ENTER to confirm): "))
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPink)).Render(m.inputBuffer + "█"))
        } else {
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonLilac)).Render("↑/↓: Navigate  ENTER: Edit  C: Clear section  R: Reset all  TAB: Next screen"))
        }
        
        return darkModeBase.Render(b.String())
}

func (m FilterModel) renderDateSection() string {
        var b strings.Builder
        
        prefix := "  "
        if m.currentField == 0 || m.currentField == 1 {
                prefix = "→ "
        }
        
        b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple)).Render(prefix + "📅 Date Range:\n"))
        
        startDate := "Not set"
        endDate := "Not set"
        
        if m.filter != nil && m.filter.DateRange != nil {
                if m.filter.DateRange.StartDate != nil {
                        startDate = m.filter.DateRange.StartDate.Format("2006-01-02")
                }
                if m.filter.DateRange.EndDate != nil {
                        endDate = m.filter.DateRange.EndDate.Format("2006-01-02")
                }
        }
        
        startStyle := darkModeBase.Copy().Foreground(lipgloss.Color("#888888"))
        endStyle := darkModeBase.Copy().Foreground(lipgloss.Color("#888888"))
        
        if m.currentField == 0 {
                startStyle = startStyle.Foreground(lipgloss.Color(protonPink))
        }
        if m.currentField == 1 {
                endStyle = endStyle.Foreground(lipgloss.Color(protonPink))
        }
        
        b.WriteString(startStyle.Render("    Start: " + startDate + "\n"))
        b.WriteString(endStyle.Render("    End:   " + endDate + "\n"))
        
        return b.String()
}

func (m FilterModel) renderLabelsSection() string {
        var b strings.Builder
        
        prefix := "  "
        if m.currentField == 2 {
                prefix = "→ "
        }
        
        style := darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple))
        if m.currentField == 2 {
                style = style.Foreground(lipgloss.Color(protonPink))
        }
        
        b.WriteString(style.Render(prefix + "🏷️  Labels/Folders:\n"))
        
        if m.filter != nil && m.filter.Labels != nil && len(m.filter.Labels.IncludeLabels) > 0 {
                for _, label := range m.filter.Labels.IncludeLabels {
                        b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#888888")).Render("    • " + label + "\n"))
                }
        } else {
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#888888")).Render("    No labels selected\n"))
        }
        
        return b.String()
}

func (m FilterModel) renderContactsSection() string {
        var b strings.Builder
        
        prefix := "  "
        if m.currentField == 3 {
                prefix = "→ "
        }
        
        style := darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple))
        if m.currentField == 3 {
                style = style.Foreground(lipgloss.Color(protonPink))
        }
        
        b.WriteString(style.Render(prefix + "👤 Contacts (Senders):\n"))
        
        if m.filter != nil && m.filter.Contacts != nil && len(m.filter.Contacts.IncludeSenders) > 0 {
                for _, contact := range m.filter.Contacts.IncludeSenders {
                        b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#888888")).Render("    • " + contact + "\n"))
                }
        } else {
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#888888")).Render("    No contacts selected\n"))
        }
        
        return b.String()
}

func (m FilterModel) renderDomainsSection() string {
        var b strings.Builder
        
        prefix := "  "
        if m.currentField == 4 {
                prefix = "→ "
        }
        
        style := darkModeBase.Copy().Foreground(lipgloss.Color(protonPurple))
        if m.currentField == 4 {
                style = style.Foreground(lipgloss.Color(protonPink))
        }
        
        b.WriteString(style.Render(prefix + "🌐 Domains:\n"))
        
        if m.filter != nil && m.filter.Domains != nil && len(m.filter.Domains.IncludeDomains) > 0 {
                for _, domain := range m.filter.Domains.IncludeDomains {
                        b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#888888")).Render("    • " + domain + "\n"))
                }
        } else {
                b.WriteString(darkModeBase.Copy().Foreground(lipgloss.Color("#888888")).Render("    No domains selected\n"))
        }
        
        return b.String()
}

// GetFilter returns the current filter configuration
func (m FilterModel) GetFilter() *mail.ExportFilter {
        return m.filter
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
func NewFilterModel() tea.Model { 
        return FilterModel{
                filter: &mail.ExportFilter{},
                selectedLabels: make(map[string]bool),
        } 
}
func NewPluginModel() tea.Model { return PluginModel{} }

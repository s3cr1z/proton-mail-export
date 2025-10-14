package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ProtonMail/export-tool/internal"
)

type WelcomeModel struct {
	app    *App
	width  int
	height int
}

func NewWelcomeModel(app *App) *WelcomeModel {
	return &WelcomeModel{
		app: app,
	}
}

func (m *WelcomeModel) Init() tea.Cmd {
	return nil
}

func (m *WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			return m, m.app.SwitchToScreen(LoginScreen)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	
	return m, nil
}

func (m *WelcomeModel) View() string {
	// ASCII art logo
	logo := `
    ██████╗ ██████╗  ██████╗ ████████╗ ██████╗ ███╗   ██╗
    ██╔══██╗██╔══██╗██╔═══██╗╚══██╔══╝██╔═══██╗████╗  ██║
    ██████╔╝██████╔╝██║   ██║   ██║   ██║   ██║██╔██╗ ██║
    ██╔═══╝ ██╔══██╗██║   ██║   ██║   ██║   ██║██║╚██╗██║
    ██║     ██║  ██║╚██████╔╝   ██║   ╚██████╔╝██║ ╚████║
    ╚═╝     ╚═╝  ╚═╝ ╚═════╝    ╚═╝    ╚═════╝ ╚═╝  ╚═══╝
    
    ███╗   ███╗ █████╗ ██╗██╗         ███████╗██╗  ██╗██████╗  ██████╗ ██████╗ ████████╗
    ████╗ ████║██╔══██╗██║██║         ██╔════╝╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝
    ██╔████╔██║███████║██║██║         █████╗   ╚███╔╝ ██████╔╝██║   ██║██████╔╝   ██║   
    ██║╚██╔╝██║██╔══██║██║██║         ██╔══╝   ██╔██╗ ██╔═══╝ ██║   ██║██╔══██╗   ██║   
    ██║ ╚═╝ ██║██║  ██║██║███████╗    ███████╗██╔╝ ██╗██║     ╚██████╔╝██║  ██║   ██║   
    ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝    ╚══════╝╚═╝  ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝   
`

	// Style the logo
	logoStyle := lipgloss.NewStyle().
		Foreground(ProtonPurple).
		Bold(true).
		Align(lipgloss.Center)

	// Welcome message
	welcomeText := fmt.Sprintf(`
Welcome to Proton Mail Export Tool v%s

A modern, interactive tool for exporting and managing your Proton Mail data.

✨ Features:
  • Interactive TUI interface with modern design
  • Advanced email filtering and search
  • Multiple export formats (EML, PDF, MBOX)
  • Encrypted backups with password protection
  • Incremental backup support
  • Real-time progress tracking
  • Cross-platform compatibility

🔒 Privacy First:
  • End-to-end encrypted communication
  • Local processing - your data stays private
  • No telemetry without consent
  • Open source and auditable

Press ENTER to continue or Q to quit
`, internal.ETVersionString)

	welcomeStyle := lipgloss.NewStyle().
		Foreground(DarkFg).
		Padding(1, 2).
		Align(lipgloss.Center)

	// Copyright and license info
	footerText := `
© 2024 Proton AG, Switzerland
Licensed under GNU General Public License v3
Get support at https://proton.me/support/proton-mail-export-tool
`

	footerStyle := lipgloss.NewStyle().
		Foreground(ProtonGray).
		Align(lipgloss.Center).
		Italic(true)

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		logoStyle.Render(logo),
		welcomeStyle.Render(welcomeText),
		footerStyle.Render(footerText),
	)

	// Center the content vertically and horizontally
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 4). // Account for header/footer
		Align(lipgloss.Center, lipgloss.Center)

	return containerStyle.Render(content)
}

// Helper function to center text
func centerText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var centeredLines []string
	
	for _, line := range lines {
		if len(line) >= width {
			centeredLines = append(centeredLines, line)
			continue
		}
		
		padding := (width - len(line)) / 2
		centeredLine := strings.Repeat(" ", padding) + line
		centeredLines = append(centeredLines, centeredLine)
	}
	
	return strings.Join(centeredLines, "\n")
}
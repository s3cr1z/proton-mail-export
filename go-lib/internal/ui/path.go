package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Path Model
type PathModel struct {
	operation     string
	selectedPath  string
	useDefault    bool
	form          *huh.Form
	width         int
	height        int
	defaultPath   string
	userEmail     string
	error         string
	spaceInfo     SpaceInfo
}

type SpaceInfo struct {
	Available uint64
	Required  uint64
	Path      string
}

func NewPathModel() PathModel {
	return PathModel{}
}

func (m PathModel) Init() tea.Cmd {
	return m.createForm()
}

func (m PathModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg {
				return ScreenChangeMsg{Screen: OperationScreen}
			}
		}

	case OperationSelectedMsg:
		m.operation = msg.Operation
		m.defaultPath = m.getDefaultPath()
		return m, m.createForm()

	case huh.FormCompleteMsg:
		return m, m.validateAndProceed()
	}

	// Update form
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		return m, cmd
	}

	return m, nil
}

func (m PathModel) View() string {
	var content strings.Builder

	// Title
	title := "Select Path"
	if m.operation == OperationBackup {
		title = "Select Backup Destination"
	} else if m.operation == OperationRestore {
		title = "Select Backup Source"
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")

	// Description
	if m.operation == OperationBackup {
		content.WriteString("Choose where to save your email backup:\n\n")
		if m.defaultPath != "" {
			content.WriteString(fmt.Sprintf("Default location: %s\n\n", m.defaultPath))
		}
	} else {
		content.WriteString("Choose the backup folder to restore from:\n\n")
	}

	// Show error if any
	if m.error != "" {
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.error)))
		content.WriteString("\n\n")
	}

	// Show space information
	if m.spaceInfo.Path != "" {
		content.WriteString(m.renderSpaceInfo())
		content.WriteString("\n\n")
	}

	// Show form
	if m.form != nil {
		content.WriteString(m.form.View())
	}

	// Instructions
	content.WriteString("\n\n")
	content.WriteString(blurredStyle.Render("Press ESC to go back • Ctrl+C to quit"))

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(2).
		Render(content.String())
}

func (m PathModel) createForm() tea.Cmd {
	var inputs []huh.Field

	if m.operation == OperationBackup && m.defaultPath != "" {
		// For backup, offer default path option
		inputs = append(inputs,
			huh.NewConfirm().
				Key("use_default").
				Title(fmt.Sprintf("Use default path: %s", m.defaultPath)).
				Value(&m.useDefault),
		)
	}

	// Path input (always show for restore, or if not using default for backup)
	inputs = append(inputs,
		huh.NewInput().
			Key("path").
			Title(func() string {
				if m.operation == OperationBackup {
					return "Custom backup path"
				}
				return "Backup folder path"
			}()).
			Value(&m.selectedPath).
			Placeholder(m.getExamplePath()).
			Validate(m.validatePath),
	)

	m.form = huh.NewForm(
		huh.NewGroup(inputs...),
	).WithTheme(huh.ThemeCharm())

	return m.form.Init()
}

func (m PathModel) validatePath(path string) error {
	if path == "" && !m.useDefault {
		return fmt.Errorf("path cannot be empty")
	}

	if path == "" {
		return nil // Using default
	}

	// Expand path
	expandedPath := m.expandPath(path)

	// Check if path exists for restore operation
	if m.operation == OperationRestore {
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist")
		}

		// Check if it's a directory
		if info, err := os.Stat(expandedPath); err == nil {
			if !info.IsDir() {
				return fmt.Errorf("path is not a directory")
			}
		}
	}

	// For backup, check if we can create the directory
	if m.operation == OperationBackup {
		if err := os.MkdirAll(expandedPath, 0755); err != nil {
			return fmt.Errorf("cannot create directory: %v", err)
		}
	}

	return nil
}

func (m PathModel) validateAndProceed() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		var finalPath string

		if m.useDefault && m.defaultPath != "" {
			finalPath = m.defaultPath
		} else {
			finalPath = m.expandPath(m.selectedPath)
		}

		// Final validation
		if err := m.validatePath(finalPath); err != nil {
			m.error = err.Error()
			return nil
		}

		// Check disk space for backup operations
		if m.operation == OperationBackup {
			spaceInfo, err := m.checkDiskSpace(finalPath)
			if err != nil {
				m.error = fmt.Sprintf("Failed to check disk space: %v", err)
				return nil
			}

			m.spaceInfo = spaceInfo
			if spaceInfo.Required > spaceInfo.Available {
				// Show warning but allow user to proceed
				m.error = fmt.Sprintf("Warning: Operation requires %s but only %s available",
					formatBytes(spaceInfo.Required), formatBytes(spaceInfo.Available))
				return nil
			}
		}

		return PathSelectedMsg{Path: finalPath}
	})
}

func (m PathModel) getDefaultPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Use Downloads/proton-mail-export-cli
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Downloads", "proton-mail-export-cli")
	case "windows":
		// Windows: Use Documents folder
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Documents", "proton-mail-export-cli")
	default:
		// Linux/Unix: Use current directory or home
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "proton-mail-export-cli")
	}
}

func (m PathModel) getExamplePath() string {
	switch runtime.GOOS {
	case "darwin":
		return "~/Documents/my-backup"
	case "windows":
		return "%USERPROFILE%\\Documents\\my-backup"
	default:
		return "~/Documents/my-backup"
	}
}

func (m PathModel) expandPath(path string) string {
	// Handle ~ expansion
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}

	// Handle Windows environment variables
	if runtime.GOOS == "windows" {
		path = os.ExpandEnv(path)
	}

	// Convert to absolute path
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
	}

	return path
}

func (m PathModel) checkDiskSpace(path string) (SpaceInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you'd use syscalls to get actual disk space
	info := SpaceInfo{
		Path:      path,
		Available: 10 * 1024 * 1024 * 1024, // 10GB placeholder
		Required:  1 * 1024 * 1024 * 1024,  // 1GB placeholder
	}

	return info, nil
}

func (m PathModel) renderSpaceInfo() string {
	var content strings.Builder

	content.WriteString("Disk Space Information:\n")
	content.WriteString(fmt.Sprintf("Available: %s\n", formatBytes(m.spaceInfo.Available)))
	content.WriteString(fmt.Sprintf("Required:  %s\n", formatBytes(m.spaceInfo.Required)))

	if m.spaceInfo.Required > m.spaceInfo.Available {
		content.WriteString(errorStyle.Render("⚠ Insufficient disk space"))
	} else {
		content.WriteString(successStyle.Render("✓ Sufficient disk space"))
	}

	return content.String()
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
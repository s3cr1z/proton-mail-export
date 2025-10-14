package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ProtonMail/export-tool/internal"
	"github.com/ProtonMail/export-tool/internal/sentry"
	"github.com/ProtonMail/export-tool/internal/tui/models"
	"github.com/sirupsen/logrus"
)

// Run starts the TUI application
func Run() error {
	// Initialize application similar to CLI version
	folder, err := getDefaultOperationFolder()
	if err != nil {
		return fmt.Errorf("failed to get default folder: %w", err)
	}

	if err := initApp(filepath.Join(folder, "logs")); err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer closeApp()

	// Create the main application model
	ctx := context.Background()
	app := models.NewAppModel(ctx)

	// Create the Bubble Tea program
	program := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// getDefaultOperationFolder returns the default folder for operations
// This is copied from the CLI app.go but should be shared
func getDefaultOperationFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "ProtonMailExport"), nil
}

// initApp initializes the application
// This is a simplified version of the CLI initApp function
func initApp(logPath string) error {
	if err := sentry.InitSentry(); err != nil {
		return err
	}

	if err := os.MkdirAll(logPath, 0o700); err != nil {
		return err
	}

	logFile := filepath.Join(logPath, internal.NewLogFileName())
	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		ForceQuote:       true,
		FullTimestamp:    true,
		QuoteEmptyFields: true,
		TimestampFormat:  "2006-01-02 15:04:05.000",
	})

	internal.LogPrelude()
	
	// Store file handle for cleanup (simplified)
	appState.logFile = file

	return nil
}

// closeApp cleans up application resources
func closeApp() {
	if appState.logFile != nil {
		logrus.SetOutput(os.Stdout)
		if err := appState.logFile.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close log file")
		}
		appState.logFile = nil
	}
}

// Simple global state for cleanup
var appState struct {
	logFile *os.File
}
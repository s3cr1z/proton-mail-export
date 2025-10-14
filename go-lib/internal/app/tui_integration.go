package app

import (
        "context"
        "fmt"
        "os"
        "strings"
        "time"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/ProtonMail/export-tool/internal/apiclient"
        "github.com/ProtonMail/export-tool/internal/mail"
        "github.com/ProtonMail/export-tool/internal/sentry"
        "github.com/ProtonMail/export-tool/internal/session"
        "github.com/ProtonMail/export-tool/internal/ui"
        "github.com/urfave/cli/v2"
        "golang.org/x/term"
)

// TUI mode detection and integration
func shouldUseTUI(ctx *cli.Context) bool {
        // Check if explicitly disabled
        if ctx.Bool("no-tui") || os.Getenv("ET_NO_TUI") != "" {
                return false
        }

        // Check if running in a terminal
        if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
                return false
        }

        // Check if any CLI arguments are provided (prefer CLI mode for automation)
        if ctx.String("username") != "" || ctx.String("password") != "" || 
           ctx.String("operation") != "" || ctx.String("dir") != "" {
                return false
        }

        // Check terminal capabilities
        termType := os.Getenv("TERM")
        if termType == "" || termType == "dumb" {
                return false
        }

        return true
}

// Run TUI mode
func runTUI(ctx *cli.Context) error {
        // Initialize TUI model
        model := ui.NewModel()
        
        // Create program with alt screen
        program := tea.NewProgram(
                model,
                tea.WithAltScreen(),
                tea.WithMouseCellMotion(),
        )

        // Run the TUI
        finalModel, err := program.Run()
        if err != nil {
                return fmt.Errorf("TUI error: %v", err)
        }

        // Check if user cancelled
        if m, ok := finalModel.(ui.Model); ok {
                if m.IsCancelled() {
                        return fmt.Errorf("operation cancelled by user")
                }
        }

        return nil
}

// Enhanced CLI mode with better UX
func runEnhancedCLI(ctx *cli.Context) error {
        printHeader()
        checkForNewVersion()

        fmt.Printf("\nSession log: %v\n\n", state.logPath)

        session, err := newSession(sentry.NewPanicHandler(func() {}))
        if err != nil {
                return err
        }

        // Enhanced login with better prompts
        if err = enhancedLogin(ctx, session); err != nil {
                return err
        }

        // Enhanced operation selection
        operation, err := getEnhancedOperation(ctx)
        if err != nil {
                return err
        }

        // Enhanced path selection
        dir, err := getEnhancedTargetFolder(ctx, operation, session.GetUser().Email)
        if err != nil {
                return err
        }

        // Run operation with enhanced progress
        if operation == operationBackup {
                return runEnhancedBackup(ctx.Context, dir, session)
        }

        if operation == operationRestore {
                return runEnhancedRestore(ctx.Context, dir, session)
        }

        return nil
}

// Enhanced login with better UX
func enhancedLogin(ctx *cli.Context, s *session.Session) error {
        fmt.Println("🔐 Login to Proton Mail")
        fmt.Println(strings.Repeat("─", 50))

        creds := newCredentialsFromCLI(ctx)
        var err error

        for {
                switch s.LoginState() {
                case session.LoginStateLoggedOut:
                        if len(creds.username) == 0 {
                                fmt.Print("\n📧 Username: ")
                                if creds.username, err = readLine(""); err != nil {
                                        return err
                                }
                        } else {
                                fmt.Printf("\n📧 Username: %s\n", creds.username)
                        }

                        if len(creds.password) == 0 {
                                fmt.Print("🔑 Password: ")
                                if passwordBytes, err := readPassword(""); err != nil {
                                        return err
                                } else {
                                        creds.password = string(passwordBytes)
                                }
                        }

                        fmt.Print("🔄 Logging in...")
                        if err := s.Login(ctx.Context, creds.username, creds.password); err != nil {
                                fmt.Printf(" ❌ Failed\n")
                                printError(err)
                                if err := creds.nextAttempt(); err != nil {
                                        return err
                                }
                        } else {
                                fmt.Printf(" ✅ Success\n")
                        }

                case session.LoginStateAwaitingTOTP:
                        if len(creds.totp) == 0 {
                                fmt.Print("\n🔢 Enter 2FA code: ")
                                if creds.totp, err = readLine(""); err != nil {
                                        return err
                                }
                        }

                        fmt.Print("🔄 Verifying 2FA...")
                        if err := s.SubmitTOTP(ctx.Context, creds.totp); err != nil {
                                fmt.Printf(" ❌ Failed\n")
                                printError(err)
                                if err := creds.nextAttempt(); err != nil {
                                        return err
                                }
                        } else {
                                fmt.Printf(" ✅ Success\n")
                        }

                case session.LoginStateAwaitingMailboxPassword:
                        if len(creds.mboxPassword) == 0 {
                                fmt.Print("\n🔐 Mailbox password: ")
                                if passwordBytes, err := readPassword(""); err != nil {
                                        return err
                                } else {
                                        creds.mboxPassword = string(passwordBytes)
                                }
                        }

                        fmt.Print("🔄 Verifying mailbox password...")
                        if err := s.SubmitMailboxPassword(
                                apiclient.NewProtonMailboxPasswordValidator(s.GetUser(), s.GetUserSalts()),
                                creds.mboxPassword,
                        ); err != nil {
                                fmt.Printf(" ❌ Failed\n")
                                printError(err)
                                if err := creds.nextAttempt(); err != nil {
                                        return err
                                }
                        } else {
                                fmt.Printf(" ✅ Success\n")
                        }

                case session.LoginStateAwaitingHV:
                        url, err := s.GetHVSolveURL()
                        if err != nil {
                                return err
                        }

                        fmt.Printf("\n🤖 Human Verification Required\n")
                        fmt.Printf("Please open this URL in your browser:\n")
                        fmt.Printf("🔗 %s\n\n", url)
                        fmt.Print("Press ENTER when completed...")
                        waitForReturn()

                        if err := s.MarkHVSolved(ctx.Context); err != nil {
                                return err
                        }

                case session.LoginStateLoggedIn:
                        fmt.Printf("\n✅ Successfully logged in as %s\n", s.GetUser().Email)
                        return nil

                default:
                        return fmt.Errorf("unknown login state: %v", s.LoginState())
                }
        }
}

// Enhanced operation selection
func getEnhancedOperation(ctx *cli.Context) (string, error) {
        if operation := ctx.String("operation"); operation != "" {
                return operation, nil
        }

        if envOp := os.Getenv("ET_OPERATION"); envOp != "" {
                return envOp, nil
        }

        fmt.Println("\n📋 Select Operation")
        fmt.Println(strings.Repeat("─", 50))
        fmt.Println("1. 📤 Backup - Export emails to local storage")
        fmt.Println("2. 📥 Restore - Import emails from backup")
        fmt.Print("\nEnter your choice (1-2 or b/r): ")

        choice, err := readLine("")
        if err != nil {
                return "", err
        }

        choice = strings.ToLower(strings.TrimSpace(choice))
        switch choice {
        case "1", "b", "backup":
                return operationBackup, nil
        case "2", "r", "restore":
                return operationRestore, nil
        default:
                return "", fmt.Errorf("invalid choice: %s", choice)
        }
}

// Enhanced target folder selection
func getEnhancedTargetFolder(ctx *cli.Context, operation, email string) (string, error) {
        if dir := ctx.String("dir"); dir != "" {
                return dir, nil
        }

        if envDir := os.Getenv("ET_DIR"); envDir != "" {
                return envDir, nil
        }

        fmt.Printf("\n📁 Select %s Path\n", strings.Title(operation))
        fmt.Println(strings.Repeat("─", 50))

        if operation == operationBackup {
                defaultPath, _ := getDefaultOperationFolder()
                fmt.Printf("Default path: %s\n", defaultPath)
                fmt.Print("Use default path? (Y/n): ")

                if choice, err := readLine(""); err != nil {
                        return "", err
                } else if strings.ToLower(strings.TrimSpace(choice)) != "n" {
                        return defaultPath, nil
                }
        }

        fmt.Printf("Enter %s path: ", operation)
        return readLine("")
}

// Enhanced backup with progress
func runEnhancedBackup(ctx context.Context, exportPath string, session *session.Session) error {
        fmt.Printf("\n📤 Starting Backup\n")
        fmt.Println(strings.Repeat("─", 50))
        fmt.Printf("📍 Destination: %s\n\n", exportPath)

        exportTask := mail.NewExportTask(ctx, exportPath, session)
        
        // Use enhanced CLI reporter
        reporter := newEnhancedCliReporter()
        
        err := exportTask.Run(ctx, reporter)
        if err == nil {
                fmt.Println("\n✅ Backup completed successfully!")
        } else {
                fmt.Printf("\n❌ Backup failed: %v\n", err)
        }

        return err
}

// Enhanced restore with progress
func runEnhancedRestore(ctx context.Context, backupPath string, session *session.Session) error {
        fmt.Printf("\n📥 Starting Restore\n")
        fmt.Println(strings.Repeat("─", 50))
        fmt.Printf("📍 Source: %s\n\n", backupPath)

        restoreTask, err := mail.NewRestoreTask(ctx, backupPath, session)
        if err != nil {
                return err
        }

        // Use enhanced CLI reporter
        reporter := newEnhancedCliReporter()
        
        err = restoreTask.Run(reporter)
        if err == nil {
                fmt.Println("\n✅ Restore completed successfully!")
                printRestoreTaskSummary(restoreTask)
        } else {
                fmt.Printf("\n❌ Restore failed: %v\n", err)
        }

        return err
}

// Enhanced CLI reporter with better progress display
func newEnhancedCliReporter() *EnhancedCliReporter {
        return &EnhancedCliReporter{
                lastUpdate: time.Now(),
        }
}

type EnhancedCliReporter struct {
        lastUpdate time.Time
        lastCount  int
}

func (r *EnhancedCliReporter) ReportProgress(current, total int, message string) {
        now := time.Now()
        
        // Throttle updates to avoid spam
        if now.Sub(r.lastUpdate) < 100*time.Millisecond && current != total {
                return
        }
        
        r.lastUpdate = now
        
        // Calculate progress
        percent := float64(current) / float64(total) * 100
        
        // Calculate rate
        rate := 0.0
        if current > r.lastCount {
                elapsed := now.Sub(r.lastUpdate).Seconds()
                if elapsed > 0 {
                        rate = float64(current-r.lastCount) / elapsed
                }
        }
        r.lastCount = current
        
        // Create progress bar
        barWidth := 40
        filled := int(float64(barWidth) * percent / 100)
        bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
        
        // Print progress line
        fmt.Printf("\r🔄 [%s] %.1f%% (%d/%d) %.1f items/s - %s", 
                bar, percent, current, total, rate, message)
        
        if current == total {
                fmt.Println() // New line when complete
        }
}
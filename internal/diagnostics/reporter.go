package diagnostics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/pkg/hyerrors"
)

// POSSIBLE REFACTOR

// Reporter handles error logging and crash reporting
type Reporter struct {
	logDir   string
	crashDir string
	version  string
}

// CrashReport contains all information about a crash
type CrashReport struct {
	Timestamp  time.Time          `json:"timestamp"`
	AppVersion string             `json:"app_version"`
	OS         string             `json:"os"`
	Arch       string             `json:"arch"`
	Error      *hyerrors.AppError `json:"error"`
	SystemInfo SystemInfo         `json:"system_info"`
	RecentLogs []string           `json:"recent_logs,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	NumCPU       int    `json:"num_cpu"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
}

// NewReporter creates a new diagnostics reporter
func NewReporter(appVersion string) (*Reporter, error) {
	appDir := env.GetDefaultAppDir()

	logDir := filepath.Join(appDir, "logs")
	crashDir := filepath.Join(appDir, "crashes")

	// Create directories
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	if err := os.MkdirAll(crashDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create crash directory: %w", err)
	}

	reporter := &Reporter{
		logDir:   logDir,
		crashDir: crashDir,
		version:  appVersion,
	}

	// Register this reporter as the global error handler
	hyerrors.SetHandler(reporter)

	// Clean up old crash reports on startup
	go reporter.ClearOldCrashReports()

	return reporter, nil
}

// HandleError implements hyerrors.ErrorHandler interface
// This is called automatically whenever hyerrors.NewAppError() is called
func (r *Reporter) HandleError(err *hyerrors.AppError) {
	// Log to file
	r.logError(err)

	// Save crash report for critical errors
	if err.IsCritical() {
		_ = r.saveCrashReport(err)
	}
}

// logError writes errors to a log file
func (r *Reporter) logError(err *hyerrors.AppError) {
	logFile := filepath.Join(r.logDir, "errors.log")
	f, fileErr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", fileErr)
		return
	}
	defer f.Close()

	logEntry := fmt.Sprintf(
		"[%s] [%s] %s\nTechnical: %s\nStack:\n%s\n---\n",
		err.Timestamp.Format("2006-01-02 15:04:05"),
		err.Type,
		err.Message,
		err.Technical,
		err.Stack,
	)

	if _, err := f.WriteString(logEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log file: %v\n", err)
	}
}

// saveCrashReport saves a crash report to disk
func (r *Reporter) saveCrashReport(err *hyerrors.AppError) error {
	report := CrashReport{
		Timestamp:  time.Now(),
		AppVersion: r.version,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Error:      err,
		SystemInfo: SystemInfo{
			NumCPU:       runtime.NumCPU(),
			GOOS:         runtime.GOOS,
			GOARCH:       runtime.GOARCH,
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
		},
	}

	// Try to read recent logs
	logFile := filepath.Join(r.logDir, "errors.log")
	if logData, readErr := os.ReadFile(logFile); readErr == nil {
		// Get last 5000 characters
		lines := string(logData)
		if len(lines) > 5000 {
			lines = lines[len(lines)-5000:]
		}
		report.RecentLogs = []string{lines}
	}

	// Marshal to JSON
	data, marshalErr := json.MarshalIndent(report, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal crash report: %w", marshalErr)
	}

	// Save to file
	filename := fmt.Sprintf("crash_%s.json", time.Now().Format("2006-01-02_15-04-05"))
	crashFile := filepath.Join(r.crashDir, filename)

	return os.WriteFile(crashFile, data, 0644)
}

// ClearOldCrashReports removes crash reports older than 30 days
func (r *Reporter) ClearOldCrashReports() error {
	entries, err := os.ReadDir(r.crashDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(thirtyDaysAgo) {
			filePath := filepath.Join(r.crashDir, entry.Name())
			_ = os.Remove(filePath)
		}
	}

	return nil
}

// Cleanup performs cleanup operations
func (r *Reporter) Cleanup() {
	// Any cleanup needed before shutdown
}

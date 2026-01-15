package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
)

// CrashReport contains all information about a crash
type CrashReport struct {
	Timestamp  time.Time  `json:"timestamp"`
	AppVersion string     `json:"app_version"`
	OS         string     `json:"os"`
	Arch       string     `json:"arch"`
	Error      *AppError  `json:"error"`
	SystemInfo SystemInfo `json:"system_info"`
	RecentLogs []string   `json:"recent_logs,omitempty"`
}

// SystemInfo contains system information
type SystemInfo struct {
	NumCPU       int    `json:"num_cpu"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
}

// SaveCrashReport saves a crash report to disk
func SaveCrashReport(err *AppError) error {
	crashDir := filepath.Join(env.GetDefaultAppDir(), "crashes")
	if mkdirErr := os.MkdirAll(crashDir, 0755); mkdirErr != nil {
		return mkdirErr
	}

	report := CrashReport{
		Timestamp:  time.Now(),
		AppVersion: AppVersion,
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
	logFile := filepath.Join(env.GetDefaultAppDir(), "logs", "errors.log")
	if logData, readErr := os.ReadFile(logFile); readErr == nil {
		// Get last 50 lines or so
		lines := string(logData)
		if len(lines) > 5000 {
			lines = lines[len(lines)-5000:]
		}
		report.RecentLogs = []string{lines}
	}

	// Marshal to JSON
	data, marshalErr := json.MarshalIndent(report, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}

	// Save to file
	filename := fmt.Sprintf("crash_%s.json", time.Now().Format("2006-01-02_15-04-05"))
	crashFile := filepath.Join(crashDir, filename)

	return os.WriteFile(crashFile, data, 0644)
}

// GetCrashReports returns all crash reports
func (a *App) GetCrashReports() ([]CrashReport, error) {
	crashDir := filepath.Join(env.GetDefaultAppDir(), "crashes")

	entries, err := os.ReadDir(crashDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []CrashReport{}, nil
		}
		return nil, err
	}

	var reports []CrashReport
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(crashDir, entry.Name()))
		if err != nil {
			continue
		}

		var report CrashReport
		if err := json.Unmarshal(data, &report); err != nil {
			continue
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// ClearOldCrashReports removes crash reports older than 30 days
func ClearOldCrashReports() error {
	crashDir := filepath.Join(env.GetDefaultAppDir(), "crashes")

	entries, err := os.ReadDir(crashDir)
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
			os.Remove(filepath.Join(crashDir, entry.Name()))
		}
	}

	return nil
}

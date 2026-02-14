package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"HyLauncher/pkg/hyerrors"
)

type Reporter struct {
	rootDir    string
	appVersion string
	mu         sync.Mutex
}

type CrashReport struct {
	ID         string          `json:"id"`
	Timestamp  time.Time       `json:"timestamp"`
	AppVersion string          `json:"app_version"`
	Error      *hyerrors.Error `json:"error"`
	System     SystemInfo      `json:"system"`
	Logs       []LogEntry      `json:"recent_logs,omitempty"`
}

type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
}

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Severity  hyerrors.Severity `json:"severity"`
	Category  hyerrors.Category `json:"category"`
	Message   string            `json:"message"`
	Details   string            `json:"details,omitempty"`
}

func NewCrashReporter(rootDir, appVersion string) (*Reporter, error) {
	r := &Reporter{
		rootDir:    rootDir,
		appVersion: appVersion,
	}

	if err := r.ensureDirs(); err != nil {
		return nil, err
	}

	if err := ensureErrorsFile(rootDir); err != nil {
		return nil, err
	}

	hyerrors.RegisterHandlerFunc(r.handleError)

	go r.cleanupOldReports(30 * 24 * time.Hour)

	return r, nil
}

func (r *Reporter) ensureDirs() error {
	dirs := []string{
		r.logsDir(),
		r.crashesDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	return nil
}

func (r *Reporter) logsDir() string {
	return filepath.Join(r.rootDir, "logs")
}

func (r *Reporter) crashesDir() string {
	return filepath.Join(r.rootDir, "crashes")
}

func (r *Reporter) handleError(err *hyerrors.Error) {
	r.logError(err)

	if err.IsCritical() {
		r.saveCrashReport(err)
	}
}

func (r *Reporter) logError(err *hyerrors.Error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ensureErrorsFile(r.rootDir); err != nil {
		// Silently ignore - errors.log is optional
	}

	logPath := filepath.Join(r.logsDir(), "errors.log")
	f, fileErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		fmt.Fprintf(os.Stderr, "failed to open log: %v\n", fileErr)
		return
	}
	defer f.Close()

	entry := fmt.Sprintf("[%s] [%s] [%s] %s\n",
		err.Timestamp.Format("2006-01-02 15:04:05"),
		severityString(err.Severity),
		err.Category,
		err.Error(),
	)

	if err.Details != "" {
		entry += fmt.Sprintf("  Details: %s\n", err.Details)
	}

	if len(err.Stack) > 0 {
		entry += "  Stack:\n"
		for _, frame := range err.Stack {
			entry += fmt.Sprintf("    %s:%d %s\n", frame.File, frame.Line, frame.Function)
		}
	}

	entry += "---\n"

	f.WriteString(entry)
}

func (r *Reporter) saveCrashReport(err *hyerrors.Error) {
	report := CrashReport{
		ID:         err.ID,
		Timestamp:  time.Now(),
		AppVersion: r.appVersion,
		Error:      err,
		System: SystemInfo{
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
			NumCPU:       runtime.NumCPU(),
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Logs: r.readRecentLogs(),
	}

	data, marshalErr := json.MarshalIndent(report, "", "  ")
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "marshal crash report: %v\n", marshalErr)
		return
	}

	filename := fmt.Sprintf("crash_%s_%s.json",
		time.Now().Format("2006-01-02_15-04-05"),
		err.ID,
	)
	crashPath := filepath.Join(r.crashesDir(), filename)

	if writeErr := os.WriteFile(crashPath, data, 0644); writeErr != nil {
		fmt.Fprintf(os.Stderr, "write crash report: %v\n", writeErr)
	}
}

func (r *Reporter) readRecentLogs() []LogEntry {
	logPath := filepath.Join(r.logsDir(), "errors.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil
	}

	if len(data) > 10000 {
		data = data[len(data)-10000:]
	}

	return nil
}

func (r *Reporter) cleanupOldReports(maxAge time.Duration) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	r.performCleanup(maxAge)

	for range ticker.C {
		r.performCleanup(maxAge)
	}
}

func (r *Reporter) performCleanup(maxAge time.Duration) {
	entries, err := os.ReadDir(r.crashesDir())
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(r.crashesDir(), entry.Name()))
		}
	}
}

func (r *Reporter) GetCrashReports() ([]CrashReport, error) {
	entries, err := os.ReadDir(r.crashesDir())
	if err != nil {
		if os.IsNotExist(err) {
			return []CrashReport{}, nil
		}
		return nil, err
	}

	reports := make([]CrashReport, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(r.crashesDir(), entry.Name()))
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

func (r *Reporter) GetLogs() (string, error) {
	logPath := filepath.Join(r.logsDir(), "errors.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func severityString(s hyerrors.Severity) string {
	switch s {
	case hyerrors.SeverityInfo:
		return "INFO"
	case hyerrors.SeverityWarning:
		return "WARN"
	case hyerrors.SeverityError:
		return "ERROR"
	case hyerrors.SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func ensureErrorsFile(appDir string) error {
	path := filepath.Join(appDir, "logs", "errors.log")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if _, err := os.Create(path); err != nil {
			return err
		}
	}

	return nil
}

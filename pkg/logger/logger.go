package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"HyLauncher/pkg/sysinfo"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	mu        sync.Mutex
	file      *os.File
	level     Level
	console   bool
	sessionID string
}

var defaultLogger *Logger

func Init(logDir string, level Level, console bool) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("launcher_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	sessionID := generateSessionID()
	defaultLogger = &Logger{
		file:      file,
		level:     level,
		console:   console,
		sessionID: sessionID,
	}

	defaultLogger.log(INFO, "=== LAUNCHER STARTED ===")
	defaultLogger.log(INFO, "Session", "id", sessionID)

	logSystemInfo()

	defaultLogger.log(INFO, "Log file", "path", logFile)

	return nil
}

func generateSessionID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, 4)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

func logSystemInfo() {
	info := sysinfo.GetSystemInfo()

	Info("OS",
		"name", info.OS.Name,
		"version", info.OS.Version,
		"kernel", info.OS.Kernel,
		"arch", info.OS.Arch,
		"go_version", info.OS.GoVersion,
	)

	Info("CPU",
		"model", info.CPU.Model,
		"cores", info.CPU.Cores,
		"threads", info.CPU.Threads,
	)

	Info("Memory",
		"total", info.Memory.Total,
	)

	for i, gpu := range info.GPU {
		Info("GPU",
			"index", i,
			"model", gpu.Model,
			"vendor", gpu.Vendor,
		)
	}

	for i, display := range info.Displays {
		Info("Display",
			"index", i,
			"resolution", display.Resolution,
		)
	}

	Info("Environment",
		"home", os.Getenv("HOME"),
		"user", os.Getenv("USER"),
		"xdg_session_type", os.Getenv("XDG_SESSION_TYPE"),
		"xdg_current_desktop", os.Getenv("XDG_CURRENT_DESKTOP"),
	)
}

func (l *Logger) log(level Level, msg string, keysAndValues ...any) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	fields := ""
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		if i > 0 {
			fields += " "
		}
		key := keysAndValues[i]
		val := keysAndValues[i+1]
		fields += fmt.Sprintf("%s=%v", key, val)
	}

	line := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level.String(), l.sessionID, msg)
	if fields != "" {
		line += " | " + fields
	}
	line += "\n"

	l.file.WriteString(line)

	if l.console {
		fmt.Print(line)
	}
}

func Debug(msg string, keysAndValues ...any) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, msg, keysAndValues...)
	}
}

func Info(msg string, keysAndValues ...any) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, msg, keysAndValues...)
	}
}

func Warn(msg string, keysAndValues ...any) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, msg, keysAndValues...)
	}
}

func Error(msg string, keysAndValues ...any) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, msg, keysAndValues...)
	}
}

func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

func SessionID() string {
	if defaultLogger != nil {
		return defaultLogger.sessionID
	}
	return ""
}

func CleanupOldLogs(logDir string, maxAge time.Duration) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
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
			os.Remove(filepath.Join(logDir, entry.Name()))
		}
	}
	return nil
}

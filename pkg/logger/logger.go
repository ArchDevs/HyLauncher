package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
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

	defaultLogger = &Logger{
		file:      file,
		level:     level,
		console:   console,
		sessionID: generateSessionID(),
	}

	defaultLogger.log(INFO, "Logger initialized", "session", defaultLogger.sessionID, "file", logFile)
	return nil
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (l *Logger) log(level Level, msg string, keysAndValues ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Build structured fields
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

func Debug(msg string, keysAndValues ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, msg, keysAndValues...)
	}
}

func Info(msg string, keysAndValues ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, msg, keysAndValues...)
	}
}

func Warn(msg string, keysAndValues ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, msg, keysAndValues...)
	}
}

func Error(msg string, keysAndValues ...interface{}) {
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

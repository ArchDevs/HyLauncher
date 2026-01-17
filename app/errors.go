package app

// TODO FULL REFACTOR + MERGE /internal/models/errors

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
)

// ErrorType categorizes errors for better handling
type ErrorType string

const (
	ErrorTypeNetwork    ErrorType = "NETWORK"
	ErrorTypeFileSystem ErrorType = "FILESYSTEM"
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeGame       ErrorType = "GAME"
	ErrorTypeJava       ErrorType = "JAVA"
	ErrorTypeButler     ErrorType = "BUTLER"
	ErrorTypeUnknown    ErrorType = "UNKNOWN"
)

// AppError represents a structured error with context
type AppError struct {
	Type      ErrorType `json:"type"`
	Message   string    `json:"message"`
	Technical string    `json:"technical"`
	Timestamp time.Time `json:"timestamp"`
	Stack     string    `json:"stack,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(errType ErrorType, userMessage string, err error) *AppError {
	technical := ""
	if err != nil {
		technical = err.Error()
	}

	// Capture stack trace
	stack := captureStack(3)

	appErr := &AppError{
		Type:      errType,
		Message:   userMessage,
		Technical: technical,
		Timestamp: time.Now(),
		Stack:     stack,
	}

	// Log error
	logError(appErr)

	return appErr
}

// captureStack captures the call stack
func captureStack(skip int) string {
	stack := ""
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		stack += fmt.Sprintf("%s:%d %s\n", filepath.Base(file), line, fn.Name())
	}
	return stack
}

// logError writes errors to a log file
func logError(err *AppError) {
	logDir := filepath.Join(env.GetDefaultAppDir(), "logs")
	_ = os.MkdirAll(logDir, 0755)

	logFile := filepath.Join(logDir, "errors.log")
	f, fileErr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		fmt.Println("Failed to open log file:", fileErr)
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

	f.WriteString(logEntry)

	// Save crash report for critical errors
	if err.Type == ErrorTypeGame || err.Type == ErrorTypeFileSystem {
		_ = SaveCrashReport(err)
	}
}

// WrapError wraps an error with user-friendly message
func WrapError(errType ErrorType, userMessage string, err error) error {
	if err == nil {
		return nil
	}
	return NewAppError(errType, userMessage, err)
}

// Common error constructors
func NetworkError(operation string, err error) error {
	return NewAppError(
		ErrorTypeNetwork,
		fmt.Sprintf("Network error during %s. Please check your internet connection.", operation),
		err,
	)
}

func FileSystemError(operation string, err error) error {
	return NewAppError(
		ErrorTypeFileSystem,
		fmt.Sprintf("File system error during %s. Please check disk space and permissions.", operation),
		err,
	)
}

func ValidationError(message string) error {
	return NewAppError(
		ErrorTypeValidation,
		message,
		nil,
	)
}

func GameError(message string, err error) error {
	return NewAppError(
		ErrorTypeGame,
		message,
		err,
	)
}

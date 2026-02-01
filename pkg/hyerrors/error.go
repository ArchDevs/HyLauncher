package hyerrors

import (
	"fmt"
	"runtime"
	"time"
)

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

type Category string

const (
	CategoryInternal   Category = "internal"
	CategoryGame       Category = "game"
	CategoryJava       Category = "java"
	CategoryNetwork    Category = "network"
	CategoryValidation Category = "validation"
	CategoryFileSystem Category = "filesystem"
	CategoryConfig     Category = "config"
	CategoryUpdate     Category = "update"
	CategoryUnknown    Category = "unknown"
)

type Error struct {
	ID        string    `json:"id"`
	Category  Category  `json:"category"`
	Severity  Severity  `json:"severity"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Cause     error     `json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Stack     []Frame   `json:"stack,omitempty"`
	Context   Context   `json:"context,omitempty"`
}

type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Context map[string]any

func New(category Category, severity Severity, message string) *Error {
	return &Error{
		ID:        generateID(),
		Category:  category,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Stack:     captureStack(2),
		Context:   make(Context),
	}
}

func Wrap(err error, category Category, message string) *Error {
	if err == nil {
		return nil
	}

	e := New(category, SeverityError, message)
	e.Cause = err
	e.Details = err.Error()
	return e
}

func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func (e *Error) WithContext(key string, value any) *Error {
	e.Context[key] = value
	return e
}

func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

func (e *Error) IsCritical() bool {
	return e.Severity == SeverityCritical
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func captureStack(skip int) []Frame {
	frames := make([]Frame, 0, 10)
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		frames = append(frames, Frame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}
	return frames
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

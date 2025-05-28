package logger

import (
	"context"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel string

const (
	// DEBUG level for detailed information
	DEBUG LogLevel = "DEBUG"
	// INFO level for general information
	INFO LogLevel = "INFO"
	// WARN level for warning information
	WARN LogLevel = "WARN"
	// ERROR level for error information
	ERROR LogLevel = "ERROR"
	// FATAL level for fatal error information
	FATAL LogLevel = "FATAL"
)

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	ServiceName string                 `json:"service_name"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      int                    `json:"user_id,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       error                  `json:"-"`
}

// Logger defines the interface for logging
type Logger interface {
	Debug(ctx context.Context, message string, metadata ...map[string]interface{})
	Info(ctx context.Context, message string, metadata ...map[string]interface{})
	Warn(ctx context.Context, message string, metadata ...map[string]interface{})
	Error(ctx context.Context, message string, err error, metadata ...map[string]interface{})
	Fatal(ctx context.Context, message string, err error, metadata ...map[string]interface{})
}

// CompositeLogger combines multiple loggers
type CompositeLogger struct {
	loggers []Logger
}

// NewCompositeLogger creates a new composite logger
func NewCompositeLogger(loggers ...Logger) *CompositeLogger {
	return &CompositeLogger{
		loggers: loggers,
	}
}

// Debug logs a debug message to all loggers
func (l *CompositeLogger) Debug(ctx context.Context, message string, metadata ...map[string]interface{}) {
	for _, logger := range l.loggers {
		logger.Debug(ctx, message, metadata...)
	}
}

// Info logs an info message to all loggers
func (l *CompositeLogger) Info(ctx context.Context, message string, metadata ...map[string]interface{}) {
	for _, logger := range l.loggers {
		logger.Info(ctx, message, metadata...)
	}
}

// Warn logs a warning message to all loggers
func (l *CompositeLogger) Warn(ctx context.Context, message string, metadata ...map[string]interface{}) {
	for _, logger := range l.loggers {
		logger.Warn(ctx, message, metadata...)
	}
}

// Error logs an error message to all loggers
func (l *CompositeLogger) Error(ctx context.Context, message string, err error, metadata ...map[string]interface{}) {
	for _, logger := range l.loggers {
		logger.Error(ctx, message, err, metadata...)
	}
}

// Fatal logs a fatal message to all loggers
func (l *CompositeLogger) Fatal(ctx context.Context, message string, err error, metadata ...map[string]interface{}) {
	for _, logger := range l.loggers {
		logger.Fatal(ctx, message, err, metadata...)
	}
}

// GetRequestIDFromContext extracts the request ID from the context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	
	if requestID, ok := ctx.Value("requestID").(string); ok {
		return requestID
	}
	
	return ""
}

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(ctx context.Context) int {
	if ctx == nil {
		return 0
	}
	
	if userID, ok := ctx.Value("userID").(int); ok {
		return userID
	}
	
	return 0
}
package logger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// DBLogger implements the Logger interface for database logging
type DBLogger struct {
	db          *sql.DB
	serviceName string
	tableName   string
}

// NewDBLogger creates a new database logger
func NewDBLogger(db *sql.DB, serviceName, tableName string) *DBLogger {
	return &DBLogger{
		db:          db,
		serviceName: serviceName,
		tableName:   tableName,
	}
}

// ensureLogTable ensures that the log table exists
func (l *DBLogger) ensureLogTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			service_name TEXT NOT NULL,
			request_id TEXT,
			user_id INTEGER,
			action TEXT,
			resource TEXT,
			resource_id TEXT,
			metadata JSONB,
			error_message TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`, l.tableName)

	_, err := l.db.ExecContext(ctx, query)
	return err
}

// createLogEntry creates a log entry with common fields
func (l *DBLogger) createLogEntry(ctx context.Context, level LogLevel, message string, err error, metadata ...map[string]interface{}) LogEntry {
	entry := LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		ServiceName: l.serviceName,
		RequestID:   GetRequestIDFromContext(ctx),
		UserID:      GetUserIDFromContext(ctx),
		Error:       err,
	}

	// Merge metadata
	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
	}

	// Extract action, resource, and resourceID from metadata if available
	if entry.Metadata != nil {
		if action, ok := entry.Metadata["action"].(string); ok {
			entry.Action = action
		}
		if resource, ok := entry.Metadata["resource"].(string); ok {
			entry.Resource = resource
		}
		if resourceID, ok := entry.Metadata["resource_id"].(string); ok {
			entry.ResourceID = resourceID
		}
	}

	return entry
}

// Debug logs a debug message to the database
func (l *DBLogger) Debug(ctx context.Context, message string, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, DEBUG, message, nil, metadata...)
	l.logToDB(ctx, entry)
}

// Info logs an info message to the database
func (l *DBLogger) Info(ctx context.Context, message string, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, INFO, message, nil, metadata...)
	l.logToDB(ctx, entry)
}

// Warn logs a warning message to the database
func (l *DBLogger) Warn(ctx context.Context, message string, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, WARN, message, nil, metadata...)
	l.logToDB(ctx, entry)
}

// Error logs an error message to the database
func (l *DBLogger) Error(ctx context.Context, message string, err error, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, ERROR, message, err, metadata...)
	l.logToDB(ctx, entry)
}

// Fatal logs a fatal message to the database
func (l *DBLogger) Fatal(ctx context.Context, message string, err error, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, FATAL, message, err, metadata...)
	l.logToDB(ctx, entry)
}

// logToDB sends the log entry to the database
func (l *DBLogger) logToDB(ctx context.Context, entry LogEntry) {
	// Ensure log table exists
	if err := l.ensureLogTable(ctx); err != nil {
		fmt.Printf("Error ensuring log table: %v\n", err)
		return
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	var err error
	if entry.Metadata != nil {
		metadataJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			fmt.Printf("Error marshaling metadata: %v\n", err)
			return
		}
	}

	// Get error message if error is not nil
	var errorMessage string
	if entry.Error != nil {
		errorMessage = entry.Error.Error()
	}

	// Insert log entry into database
	query := fmt.Sprintf(`
		INSERT INTO %s (
			timestamp, level, message, service_name, request_id, user_id,
			action, resource, resource_id, metadata, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`, l.tableName)

	_, err = l.db.ExecContext(ctx, query,
		entry.Timestamp,
		string(entry.Level),
		entry.Message,
		entry.ServiceName,
		entry.RequestID,
		entry.UserID,
		entry.Action,
		entry.Resource,
		entry.ResourceID,
		metadataJSON,
		errorMessage,
	)

	if err != nil {
		fmt.Printf("Error inserting log entry: %v\n", err)
	}
}
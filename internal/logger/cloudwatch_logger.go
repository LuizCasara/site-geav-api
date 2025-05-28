package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// CloudWatchLogger implements the Logger interface for AWS CloudWatch
type CloudWatchLogger struct {
	client      *cloudwatch.Client
	serviceName string
	namespace   string
}

// NewCloudWatchLogger creates a new CloudWatch logger
func NewCloudWatchLogger(client *cloudwatch.Client, serviceName, namespace string) *CloudWatchLogger {
	return &CloudWatchLogger{
		client:      client,
		serviceName: serviceName,
		namespace:   namespace,
	}
}

// createLogEntry creates a log entry with common fields
func (l *CloudWatchLogger) createLogEntry(ctx context.Context, level LogLevel, message string, err error, metadata ...map[string]interface{}) LogEntry {
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

// putMetricData sends a metric to CloudWatch
func (l *CloudWatchLogger) putMetricData(ctx context.Context, entry LogEntry) error {
	// Create metric name based on log level and resource
	metricName := fmt.Sprintf("%s_%s", entry.Level, entry.Resource)
	if metricName == "" {
		metricName = string(entry.Level)
	}

	// Create dimensions
	dimensions := []types.Dimension{
		{
			Name:  aws.String("ServiceName"),
			Value: aws.String(l.serviceName),
		},
	}

	if entry.Resource != "" {
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String("Resource"),
			Value: aws.String(entry.Resource),
		})
	}

	if entry.Action != "" {
		dimensions = append(dimensions, types.Dimension{
			Name:  aws.String("Action"),
			Value: aws.String(entry.Action),
		})
	}

	// Create metric data
	_, err := l.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
		Namespace: aws.String(l.namespace),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Dimensions: dimensions,
				Timestamp:  aws.Time(entry.Timestamp),
				Value:      aws.Float64(1.0),
				Unit:       types.StandardUnitCount,
			},
		},
	})

	return err
}

// Debug logs a debug message to CloudWatch
func (l *CloudWatchLogger) Debug(ctx context.Context, message string, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, DEBUG, message, nil, metadata...)
	l.logToCloudWatch(ctx, entry)
}

// Info logs an info message to CloudWatch
func (l *CloudWatchLogger) Info(ctx context.Context, message string, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, INFO, message, nil, metadata...)
	l.logToCloudWatch(ctx, entry)
}

// Warn logs a warning message to CloudWatch
func (l *CloudWatchLogger) Warn(ctx context.Context, message string, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, WARN, message, nil, metadata...)
	l.logToCloudWatch(ctx, entry)
}

// Error logs an error message to CloudWatch
func (l *CloudWatchLogger) Error(ctx context.Context, message string, err error, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, ERROR, message, err, metadata...)
	l.logToCloudWatch(ctx, entry)
}

// Fatal logs a fatal message to CloudWatch
func (l *CloudWatchLogger) Fatal(ctx context.Context, message string, err error, metadata ...map[string]interface{}) {
	entry := l.createLogEntry(ctx, FATAL, message, err, metadata...)
	l.logToCloudWatch(ctx, entry)
}

// logToCloudWatch sends the log entry to CloudWatch
func (l *CloudWatchLogger) logToCloudWatch(ctx context.Context, entry LogEntry) {
	// Convert entry to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Error marshaling log entry: %v\n", err)
		return
	}

	// Log to CloudWatch
	if err := l.putMetricData(ctx, entry); err != nil {
		fmt.Printf("Error sending metric to CloudWatch: %v\n", err)
	}

	// Print to console for local development
	fmt.Println(string(jsonData))
}
package monitoring

import (
	"fmt"
	"io"
	"log"
	"time"

	"suppercommand/internal/config"
)

// Logger interface defines structured logging methods
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
}

// Field represents a structured log field
type Field struct {
	Key   string
	Value interface{}
}

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// BasicLogger implements the Logger interface
type BasicLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(config config.MonitoringConfig) Logger {
	// For clean UI, discard all logging output during normal operation
	logger := &BasicLogger{
		level:  LogLevelError,                          // Only show errors if needed
		logger: log.New(io.Discard, "", log.LstdFlags), // Discard all output for clean UI
	}

	return logger
}

// Debug logs a debug message
func (l *BasicLogger) Debug(msg string, fields ...Field) {
	if l.level <= LogLevelDebug {
		l.logWithFields("DEBUG", msg, fields...)
	}
}

// Info logs an info message
func (l *BasicLogger) Info(msg string, fields ...Field) {
	if l.level <= LogLevelInfo {
		l.logWithFields("INFO", msg, fields...)
	}
}

// Warn logs a warning message
func (l *BasicLogger) Warn(msg string, fields ...Field) {
	if l.level <= LogLevelWarn {
		l.logWithFields("WARN", msg, fields...)
	}
}

// Error logs an error message
func (l *BasicLogger) Error(msg string, err error, fields ...Field) {
	if l.level <= LogLevelError {
		allFields := append(fields, Field{Key: "error", Value: err.Error()})
		l.logWithFields("ERROR", msg, allFields...)
	}
}

// logWithFields logs a message with structured fields
func (l *BasicLogger) logWithFields(level, msg string, fields ...Field) {
	timestamp := time.Now().Format(time.RFC3339)
	logMsg := fmt.Sprintf("[%s] %s %s", timestamp, level, msg)

	if len(fields) > 0 {
		logMsg += " |"
		for _, field := range fields {
			logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	l.logger.Println(logMsg)
}

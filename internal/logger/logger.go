package logger

import (
	"database/sql"
	"fmt"
	"time"
)

// Logger writes structured log entries to the database and stdout.
type Logger struct {
	DB *sql.DB
}

// NewLogger creates a new Logger.
func NewLogger(db *sql.DB) *Logger {
	return &Logger{DB: db}
}

// Log writes a log entry to the database.
func (l *Logger) Log(level, source, instanceID, message, details string) {
	if l.DB == nil {
		fmt.Printf("[%s] [%s] %s\n", level, source, message)
		return
	}
	_, err := l.DB.Exec(`
		INSERT INTO logs (level, source, instance_id, message, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, level, source, instanceID, message, details, time.Now())
	if err != nil {
		fmt.Printf("ERROR writing log: %v\n", err)
	}
	// Also print to stdout for container logs
	fmt.Printf("[%s] [%s] %s\n", level, source, message)
}

// Convenience methods
func (l *Logger) Info(source, message string) {
	l.Log("info", source, "", message, "")
}

func (l *Logger) Infof(source, instanceID, format string, args ...interface{}) {
	l.Log("info", source, instanceID, fmt.Sprintf(format, args...), "")
}

func (l *Logger) Warn(source, message string) {
	l.Log("warn", source, "", message, "")
}

func (l *Logger) Warnf(source, instanceID, format string, args ...interface{}) {
	l.Log("warn", source, instanceID, fmt.Sprintf(format, args...), "")
}

func (l *Logger) Error(source, message string) {
	l.Log("error", source, "", message, "")
}

func (l *Logger) Errorf(source, instanceID, format string, args ...interface{}) {
	l.Log("error", source, instanceID, fmt.Sprintf(format, args...), "")
}

func (l *Logger) Debug(source, message string) {
	l.Log("debug", source, "", message, "")
}

func (l *Logger) Debugf(source, instanceID, format string, args ...interface{}) {
	l.Log("debug", source, instanceID, fmt.Sprintf(format, args...), "")
}

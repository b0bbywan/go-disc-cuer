package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity level of log messages
type LogLevel int

const (
	// DEBUG level for detailed diagnostic information
	DEBUG LogLevel = iota
	// INFO level for general informational messages
	INFO
	// WARN level for warning messages that don't prevent operation
	WARN
	// ERROR level for error conditions that may affect functionality
	ERROR
)

var (
	// currentLevel is the minimum log level that will be output
	currentLevel = INFO
	// logger is the underlying standard library logger
	stdLogger = log.New(os.Stderr, "", log.LstdFlags)
)

// levelNames maps LogLevel to their string representations
var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// ParseLogLevel converts a string to a LogLevel
//
// Parameters:
//   - level: The log level as a string (case-insensitive)
//
// Returns:
//   - LogLevel: The parsed log level, defaults to INFO if unrecognized
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// SetLevel configures the minimum log level to output
//
// Parameters:
//   - level: The minimum LogLevel to display
func SetLevel(level LogLevel) {
	currentLevel = level
}

// GetLevel returns the current log level
//
// Returns:
//   - LogLevel: The current minimum log level
func GetLevel() LogLevel {
	return currentLevel
}

// logMessage outputs a log message if its level meets the threshold
//
// Parameters:
//   - level: The severity level of the message
//   - format: Printf-style format string
//   - args: Arguments for the format string
func logMessage(level LogLevel, format string, args ...interface{}) {
	if level >= currentLevel {
		prefix := fmt.Sprintf("[%s] ", levelNames[level])
		message := fmt.Sprintf(format, args...)
		stdLogger.Print(prefix + message)
	}
}

// Debug logs a debug-level message
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for the format string
func Debug(format string, args ...interface{}) {
	logMessage(DEBUG, format, args...)
}

// Debugf is an alias for Debug to match common logging conventions
func Debugf(format string, args ...interface{}) {
	Debug(format, args...)
}

// Info logs an info-level message
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for the format string
func Info(format string, args ...interface{}) {
	logMessage(INFO, format, args...)
}

// Infof is an alias for Info to match common logging conventions
func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

// Warn logs a warning-level message
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for the format string
func Warn(format string, args ...interface{}) {
	logMessage(WARN, format, args...)
}

// Warnf is an alias for Warn to match common logging conventions
func Warnf(format string, args ...interface{}) {
	Warn(format, args...)
}

// Error logs an error-level message
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for the format string
func Error(format string, args ...interface{}) {
	logMessage(ERROR, format, args...)
}

// Errorf is an alias for Error to match common logging conventions
func Errorf(format string, args ...interface{}) {
	Error(format, args...)
}

// Fatal logs an error-level message and exits the program
//
// Parameters:
//   - format: Printf-style format string
//   - args: Arguments for the format string
func Fatal(format string, args ...interface{}) {
	logMessage(ERROR, format, args...)
	os.Exit(1)
}

// Fatalf is an alias for Fatal to match common logging conventions
func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

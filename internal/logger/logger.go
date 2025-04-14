package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the logger
func Init(logLevel string, logFile string) error {
	log = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	// Set output
	if logFile != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// Open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}

	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	return nil
}

// Get returns the logger instance
func Get() *logrus.Logger {
	if log == nil {
		// Initialize with default values if not initialized
		Init("info", "")
	}
	return log
}

// WithField adds a field to the log entry
func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

// WithFields adds multiple fields to the log entry
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Get().Debug(args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Get().Info(args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Get().Warn(args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Get().Error(args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

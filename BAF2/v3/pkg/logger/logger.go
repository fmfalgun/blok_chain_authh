package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is a wrapper around logrus.Logger
type Logger struct {
	*logrus.Logger
}

// New creates a new Logger instance
func New(level string) *Logger {
	log := logrus.New()
	
	// Parse log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)
	
	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
	
	log.SetOutput(os.Stdout)
	
	return &Logger{log}
}

// Default returns a new logger with default settings (INFO level)
func Default() *Logger {
	return New("info")
}

// WithField adds a field to the log entry
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields adds multiple fields to the log entry
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError adds an error field to the log entry
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}
